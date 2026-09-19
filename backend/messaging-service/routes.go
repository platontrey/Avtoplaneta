package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	dbErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_db_errors_total",
			Help: "Total number of database errors",
		},
		[]string{"operation", "service"},
	)
	businessOperationsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_business_operations_total",
			Help: "Total number of business operations",
		},
		[]string{"operation", "service", "status"},
	)
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "avtoplaneta_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)
	httpRequestsErrorsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "avtoplaneta_http_requests_errors_total",
			Help: "Total number of HTTP request errors",
		},
		[]string{"method", "endpoint", "status"},
	)
)

func RecordDBError(operation, service string) {
	dbErrorsTotal.WithLabelValues(operation, service).Inc()
}
func RecordBusinessOperation(operation, service, status string) {
	businessOperationsTotal.WithLabelValues(operation, service, status).Inc()
}

func init() {
	// Register metrics with Prometheus (avoid registering multiple times if init runs twice in tests)
	_ = prometheus.Register(dbErrorsTotal)
	_ = prometheus.Register(businessOperationsTotal)
	_ = prometheus.Register(httpRequestsTotal)
	_ = prometheus.Register(httpRequestDuration)
	_ = prometheus.Register(httpRequestsErrorsTotal)
}

func setupRoutes(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	r.Use(metricsMiddleware())

	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	api := r.Group("/api/messaging")
	{
		// WebSocket
		api.GET("/ws", ServeWebSocket)

		// Conversations
		api.GET("/conversations", getConversations)
		api.POST("/conversations", createConversation)
		api.GET("/conversations/:id", getConversation)
		api.PUT("/conversations/:id", updateConversation)
		api.DELETE("/conversations/:id", deleteConversation)
		api.DELETE("/conversations/:id/participants/:userId", removeParticipant)

		// Messages
		api.GET("/conversations/:id/messages", getMessages)
		api.POST("/conversations/:id/messages", sendMessage)
		api.POST("/conversations/:id/messages/voice", sendVoiceMessage)
		api.DELETE("/messages/:id", deleteMessage)
		api.PUT("/messages/:id/read", markMessageRead)

		// Reactions
		api.POST("/messages/:id/reactions", addReaction)
		api.DELETE("/messages/:id/reactions/:reactionId", removeReaction)

		// Notifications
		api.GET("/notifications", getNotifications)
		api.PUT("/notifications/:id/read", markNotificationRead)
		api.PUT("/notifications/read-all", markAllNotificationsRead)

		// User status
		api.GET("/users/status", getUserStatuses)
		api.PUT("/users/status", updateUserStatus)

		// Users
		api.GET("/users", getUsers)

		// Search
		api.GET("/search", searchMessages)

		// Drom
		api.GET("/drom/dialogs", getDromDialogs)
		api.GET("/drom/messages", getDromMessages)
		api.POST("/drom/messages", sendDromMessage)
	}
}

func metricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()
		status := strconv.Itoa(c.Writer.Status())
		method := c.Request.Method
		endpoint := c.FullPath()
		if endpoint == "" {
			endpoint = "unknown"
		}

		httpRequestsTotal.WithLabelValues(method, endpoint, status).Inc()
		httpRequestDuration.WithLabelValues(method, endpoint).Observe(duration)

		if c.Writer.Status() >= 400 {
			httpRequestsErrorsTotal.WithLabelValues(method, endpoint, status).Inc()
		}
	}
}

// Conversations
func getConversations(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversations, err := svc.GetConversations(c.Request.Context(), userIDInt)
	if err != nil {
		RecordDBError("fetch_conversations", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	c.JSON(http.StatusOK, conversations)
}

func createConversation(c *gin.Context) {
	ctx := c.Request.Context()
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Participants []int64 `json:"participants"`
		Title        string  `json:"title,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if len(req.Participants) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one participant required"})
		return
	}

	creatorInParticipants := false
	for _, p := range req.Participants {
		if p == userIDInt {
			creatorInParticipants = true
			break
		}
	}
	if !creatorInParticipants {
		req.Participants = append(req.Participants, userIDInt)
	}

	conv, err := svc.CreateConversation(ctx, req.Title, req.Participants)
	if err != nil {
		RecordDBError("create_conversation", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	RecordBusinessOperation("create_conversation", "messaging-service", "success")
	c.JSON(http.StatusCreated, conv)
}

func getConversation(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.ParseInt(conversationID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	conv, err := svc.GetConversation(c.Request.Context(), conversationIDInt, userIDInt)
	if err != nil {
		if err.Error() == "access denied" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		} else {
			c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		}
		return
	}

	c.JSON(http.StatusOK, conv)
}

func updateConversation(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.ParseInt(conversationID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// First verify exists and user has access
	conv, err := svc.GetConversation(c.Request.Context(), conversationIDInt, userIDInt)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	var req struct {
		Title        string  `json:"title"`
		Participants []int64 `json:"participants"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Participants == nil {
		req.Participants = conv.Participants
	}

	updated, err := svc.UpdateConversation(c.Request.Context(), conversationIDInt, req.Title, req.Participants)
	if err != nil {
		RecordDBError("update_conversation", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update conversation"})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func deleteConversation(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.ParseInt(conversationID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Verify access
	_, err = svc.GetConversation(c.Request.Context(), conversationIDInt, userIDInt)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := svc.DeleteConversation(c.Request.Context(), conversationIDInt); err != nil {
		RecordDBError("delete_conversation", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted"})
}

func removeParticipant(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.ParseInt(conversationID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	removeUserID := c.Param("userId")
	removeUserIDInt, err := strconv.ParseInt(removeUserID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid participant ID"})
		return
	}

	conv, err := svc.GetConversation(c.Request.Context(), conversationIDInt, userIDInt)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check if user to remove is in participants
	newParticipants := []int64{}
	found := false
	for _, p := range conv.Participants {
		if p == removeUserIDInt {
			found = true
		} else {
			newParticipants = append(newParticipants, p)
		}
	}

	if !found {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not a participant"})
		return
	}

	if len(newParticipants) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove last participant"})
		return
	}

	_, err = svc.UpdateConversation(c.Request.Context(), conversationIDInt, conv.Title, newParticipants)
	if err != nil {
		RecordDBError("remove_participant", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove participant"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Participant removed"})
}

// Messages
func getMessages(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.ParseInt(conversationID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Verify access
	_, err = svc.GetConversation(c.Request.Context(), conversationIDInt, userIDInt)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	messages, err := svc.GetMessages(c.Request.Context(), conversationIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

func sendMessage(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.ParseInt(conversationID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req struct {
		Content     string   `json:"content"`
		Attachments []string `json:"attachments,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message content is required"})
		return
	}

	msg, err := svc.SendMessage(c.Request.Context(), conversationIDInt, userIDInt, req.Content, req.Attachments)
	if err != nil {
		RecordDBError("create_message", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Broadcast message to WebSocket connections
	go func() {
		conv, err := svc.GetConversation(context.Background(), conversationIDInt, userIDInt)
		if err == nil && conv != nil {
			globalHub.broadcast <- BroadcastMessage{
				Participants: conv.Participants,
				Payload: map[string]interface{}{
					"event": "new_message",
					"data":  msg,
				},
			}
		}
	}()

	RecordBusinessOperation("send_message", "messaging-service", "success")
	c.JSON(http.StatusCreated, msg)
}

func sendVoiceMessage(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.ParseInt(conversationID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Verify access
	_, err = svc.GetConversation(c.Request.Context(), conversationIDInt, userIDInt)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	file, err := c.FormFile("voice")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Voice file is required"})
		return
	}

	if err := os.MkdirAll("./uploads", 0755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
		return
	}

	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".wav"
	}
	filename := fmt.Sprintf("voice_%d_%d%s", conversationIDInt, time.Now().Unix(), ext)
	filePath := filepath.Join("./uploads", filename)

	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save voice file"})
		return
	}

	voiceURL := fmt.Sprintf("%s/uploads/%s", os.Getenv("API_BASE_URL"), filename)

	msg, err := svc.SendVoiceMessage(c.Request.Context(), conversationIDInt, userIDInt, voiceURL)
	if err != nil {
		RecordDBError("create_voice_message", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	RecordBusinessOperation("send_voice_message", "messaging-service", "success")
	c.JSON(http.StatusCreated, msg)
}

func deleteMessage(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	messageID := c.Param("id")
	messageIDInt, err := strconv.ParseInt(messageID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := svc.DeleteMessage(c.Request.Context(), messageIDInt, userIDInt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}

func markMessageRead(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	messageID := c.Param("id")
	messageIDInt, err := strconv.ParseInt(messageID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	if err := svc.MarkMessageRead(c.Request.Context(), messageIDInt, userIDInt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark message as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
}

// Reactions
func addReaction(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	messageID := c.Param("id")
	messageIDInt, err := strconv.ParseInt(messageID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var req struct {
		Emoji string `json:"emoji"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	reaction, err := svc.AddReaction(c.Request.Context(), messageIDInt, userIDInt, req.Emoji)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add reaction"})
		return
	}

	c.JSON(http.StatusCreated, reaction)
}

func removeReaction(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	reactionID := c.Param("reactionId")
	reactionIDInt, err := strconv.ParseInt(reactionID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reaction ID"})
		return
	}

	if err := svc.RemoveReaction(c.Request.Context(), reactionIDInt, userIDInt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Reaction removed"})
}

// Notifications
func getNotifications(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	notifications, err := svc.GetNotifications(c.Request.Context(), userIDInt)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func markNotificationRead(c *gin.Context) {
	notificationID := c.Param("id")
	notificationIDInt, err := strconv.ParseInt(notificationID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	if err := svc.MarkNotificationRead(c.Request.Context(), notificationIDInt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark notification as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Notification marked as read"})
}

func markAllNotificationsRead(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := svc.MarkAllNotificationsRead(c.Request.Context(), userIDInt); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// UserStatus
func getUserStatuses(c *gin.Context) {
	statuses, err := svc.GetUserStatuses(c.Request.Context(), nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch user statuses"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"statuses": statuses})
}

func updateUserStatus(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		IsOnline bool `json:"is_online"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if err := svc.UpdateUserStatus(c.Request.Context(), userIDInt, req.IsOnline); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Status updated"})
}

// Search
func searchMessages(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)
	offsetStr := c.DefaultQuery("offset", "0")
	offset, _ := strconv.Atoi(offsetStr)

	messages, err := svc.SearchMessages(c.Request.Context(), userIDInt, query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// Drom integration
func getDromDialogs(c *gin.Context) {
	briefs, err := svc.GetDromBriefs(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, briefs)
}

func getDromMessages(c *gin.Context) {
	dialogID := c.Query("dialog_id")
	if dialogID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dialog_id parameter is required"})
		return
	}
	messages, err := svc.GetDromMessages(c.Request.Context(), dialogID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, messages)
}

func sendDromMessage(c *gin.Context) {
	var req struct {
		DialogID string `json:"dialog_id"`
		Content  string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if req.DialogID == "" || req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dialog_id and content required"})
		return
	}

	msg, err := svc.SendDromMessage(c.Request.Context(), req.DialogID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, msg)
}

func getAuthServiceURL() string {
	if url := os.Getenv("AUTH_SERVICE_URL"); url != "" {
		return url
	}
	if url := os.Getenv("GATEWAY_URL"); url != "" {
		return url
	}
	return "http://auth-service:8083"
}

// User fetching from auth service
func getCurrentUser(ctx context.Context, userID string, authHeader string) (*User, error) {
	authURL := getAuthServiceURL()

	req, err := http.NewRequestWithContext(ctx, "GET", authURL+"/api/users/me", nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("X-User-ID", userID)

	client := gatewayHTTPClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("auth service returned status: %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &user, nil
}

func getUsers(c *gin.Context) {
	authURL := getAuthServiceURL()

	req, err := http.NewRequest("GET", authURL+"/api/users", nil)
	if err != nil {
		log.Printf("Messaging getUsers: Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	userIDHeader := c.GetHeader("X-User-ID")
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("X-User-ID", userIDHeader)

	client := gatewayHTTPClient
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Messaging getUsers: Failed to connect to gateway: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to gateway"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Messaging getUsers: Gateway returned error status: %d", resp.StatusCode)
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to fetch users"})
		return
	}

	var response struct {
		Users []User `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Messaging getUsers: Failed to parse response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": response.Users})
}

// gatewayHTTPClient используется для походов в gateway за данными других
// сервисов. Таймаут обязателен: без него зависший upstream держит запрос
// вкладки «Сообщения» до победного конца, и пользователь видит бесконечную
// загрузку вместо ошибки.
var gatewayHTTPClient = &http.Client{Timeout: 10 * time.Second}
