package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
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
)

func RecordDBError(operation, service string) {
	dbErrorsTotal.WithLabelValues(operation, service).Inc()
}
func RecordBusinessOperation(operation, service, status string) {
	businessOperationsTotal.WithLabelValues(operation, service, status).Inc()
}
func init() {
	prometheus.MustRegister(dbErrorsTotal, businessOperationsTotal)
}

func setupRoutes(r *gin.Engine) {
	api := r.Group("/api/messaging")
	{
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

// Conversations
func getConversations(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var conversations []Conversation
	if err := DB.Raw(`
		SELECT c.*, COALESCE(u.unread_count, 0) as unread_count
		FROM conversations c
		LEFT JOIN (
			SELECT conversation_id, COUNT(*) as unread_count
			FROM messages
			WHERE sender_id != ? AND NOT (read_by @> ARRAY[?]::integer[])
			GROUP BY conversation_id
		) u ON c.id = u.conversation_id
		WHERE c.participants @> ARRAY[?]::integer[]
		ORDER BY c.last_message_at DESC
	`, userIDInt, userIDInt, userIDInt).Scan(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversations": conversations})
}

func createConversation(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var req struct {
		Participants []int  `json:"participants"`
		Title        string `json:"title,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	if len(req.Participants) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one participant required"})
		return
	}

	// Ensure creator is in participants
	creatorInParticipants := false
	for _, p := range req.Participants {
		if p == (userIDInt) {
			creatorInParticipants = true
			break
		}
	}
	if !creatorInParticipants {
		req.Participants = append(req.Participants, userIDInt)
	}

	participants := make([]int64, len(req.Participants))
	for i, p := range req.Participants {
		participants[i] = int64(p)
	}

	conversation := Conversation{
		Participants:  participants,
		Title:         req.Title,
		CreatedAt:     time.Now(),
		LastMessageAt: time.Now(),
	}

	log.Printf("Creating conversation with participants: %v", conversation.Participants)
	if err := DB.Debug().Create(&conversation).Error; err != nil {
		RecordDBError("create", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

	RecordBusinessOperation(
		"create_conversation", "messaging-service", "success")

	c.JSON(http.StatusCreated, conversation)
}

func getConversation(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var conversation Conversation
	if err := DB.First(&conversation, conversationIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// Check if user is participant
	isParticipant := false
	for _, p := range conversation.Participants {
		if p == int64(userIDInt) {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"conversation": conversation})
}

func updateConversation(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var req struct {
		Title string `json:"title"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	var conversation Conversation
	if err := DB.First(&conversation, conversationIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// Check if user is participant
	isParticipant := false
	for _, p := range conversation.Participants {
		if p == int64(userIDInt) {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Update title
	conversation.Title = req.Title
	if err := DB.Save(&conversation).Error; err != nil {
		RecordDBError("update", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update conversation"})
		return
	}

	RecordBusinessOperation(
		"update_conversation", "messaging-service", "success")

	c.JSON(http.StatusOK, conversation)
}

func deleteConversation(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	var conversation Conversation
	if err := DB.First(&conversation, conversationIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// Check if user is participant
	isParticipant := false
	for _, p := range conversation.Participants {
		if p == int64(userIDInt) {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Delete messages first
	if err := DB.Where("conversation_id = ?", conversationIDInt).Delete(&Message{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete messages"})
		return
	}

	// Delete conversation
	if err := DB.Delete(&conversation).Error; err != nil {
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

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	participantID := c.Param("userId")
	participantIDInt, err := strconv.Atoi(participantID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid participant ID"})
		return
	}

	// Get current user to check role
	currentUser, err := getCurrentUser(userID, c.GetHeader("Authorization"))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get user info"})
		return
	}

	if currentUser.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Only admins can remove participants"})
		return
	}

	var conversation Conversation
	if err := DB.First(&conversation, conversationIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	// Check if participant is in conversation
	isParticipant := false
	for _, p := range conversation.Participants {
		if p == int64(participantIDInt) {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User is not a participant"})
		return
	}

	// Cannot remove yourself
	if participantIDInt == userIDInt {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove yourself"})
		return
	}

	// Remove participant
	newParticipants := []int64{}
	for _, p := range conversation.Participants {
		if p != int64(participantIDInt) {
			newParticipants = append(newParticipants, p)
		}
	}
	conversation.Participants = newParticipants

	if err := DB.Save(&conversation).Error; err != nil {
		RecordDBError("update", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove participant"})
		return
	}

	RecordBusinessOperation(
		"remove_participant", "messaging-service", "success")

	c.JSON(http.StatusOK, gin.H{"message": "Participant removed"})
}

// Messages
func getMessages(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Check if user is participant
	var conversation Conversation
	if err := DB.First(&conversation, conversationIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	isParticipant := false
	for _, p := range conversation.Participants {
		if p == int64(userIDInt) {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Try to get from cache first
	cacheKey := fmt.Sprintf("messages:%d", conversationIDInt)
	cachedMessages, err := getMessagesFromCache(cacheKey)
	if err == nil && len(cachedMessages) > 0 {
		c.JSON(http.StatusOK, gin.H{"messages": cachedMessages})
		return
	}

	// Get from database
	var messages []Message
	if err := DB.Where("conversation_id = ?", conversationIDInt).Order("created_at ASC").Preload("Reactions").Preload("Mentions").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	// Cache the messages
	go cacheMessages(cacheKey, messages)

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// === DROM HANDLERS ===

// DromDialog Drom structures
type DromDialog struct {
	ID            int     `json:"id"`
	DialogID      string  `json:"dialog_id"`
	Interlocutor  string  `json:"interlocutor"`
	CreatedAt     string  `json:"created_at"`
	LastMessageAt string  `json:"last_message_at"`
	LastMessage   *string `json:"last_message,omitempty"`
}

type DromMessage struct {
	ID        int    `json:"id"`
	DialogID  string `json:"dialog_id"`
	MessageID string `json:"message_id"`
	Author    string `json:"author"`
	Direction string `json:"direction"`
	Time      string `json:"time"`
	Text      string `json:"text"`
	IsRead    bool   `json:"is_read"`
	CreatedAt string `json:"created_at"`
}

type InboxBrief struct {
	DialogID     int    `json:"dialogId"`
	Interlocutor string `json:"interlocutor"`
}

type InboxListResponse struct {
	Briefs []InboxBrief `json:"briefs"`
}

type DromApiResponse struct {
	Interlocutor string `json:"interlocutor"`
	DialogHTML   string `json:"dialog"`
}

type SessionCookie struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type SessionData struct {
	Cookies []SessionCookie `json:"cookies"`
}

const (
	DromSessionFile = "drom_session.json"
	DromListURL     = "https://my.drom.ru/personal/messaging/inbox-list?ajax=1&fromIndex=0&count=50&list=personal"
	DromViewURL     = "https://my.drom.ru/personal/messaging/view?dialogId=%s&json=true&flat-layout=false&ajax=1"
	DromPostURL     = "https://my.drom.ru/personal/messaging/view"
	DromUserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"
)

func getDromDialogs(c *gin.Context) {
	briefs, err := fetchDromBriefs()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var dialogs []DromDialog
	for _, brief := range briefs {
		dialog := DromDialog{
			ID:            brief.DialogID,
			DialogID:      strconv.Itoa(brief.DialogID),
			Interlocutor:  brief.Interlocutor,
			CreatedAt:     time.Now().Format("2006-01-02T15:04:05Z"),
			LastMessageAt: time.Now().Format("2006-01-02T15:04:05Z"),
		}
		// Try to get last message from dialog
		if msgs, err := fetchDromMessages(strconv.Itoa(brief.DialogID)); err == nil && len(msgs) > 0 {
			lastMsg := msgs[len(msgs)-1]
			dialog.LastMessage = &lastMsg.Text
			if t, err := time.Parse("2006-01-02 15:04:05", lastMsg.Time); err == nil {
				dialog.LastMessageAt = t.Format("2006-01-02T15:04:05Z")
			}
		}
		dialogs = append(dialogs, dialog)
	}

	c.JSON(http.StatusOK, gin.H{"dialogs": dialogs})
}

func getDromMessages(c *gin.Context) {
	dialogID := c.Query("dialog_id")
	if dialogID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "dialog_id required"})
		return
	}

	messages, err := fetchDromMessages(dialogID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}

// Cache helper functions
func getMessagesFromCache(key string) ([]Message, error) {
	val, err := RedisClient.Get(context.Background(), key).Result()
	if err != nil {
		return nil, err
	}

	var messages []Message
	err = json.Unmarshal([]byte(val), &messages)
	return messages, err
}

func cacheMessages(key string, messages []Message) {
	data, err := json.Marshal(messages)
	if err != nil {
		log.Printf("Failed to marshal messages for cache: %v", err)
		return
	}

	err = RedisClient.Set(context.Background(), key, data, 10*time.Minute).Err()
	if err != nil {
		log.Printf("Failed to cache messages: %v", err)
	}
}

func invalidateMessagesCache(conversationID int) {
	key := fmt.Sprintf("messages:%d", conversationID)
	RedisClient.Del(context.Background(), key)
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

	err := sendDromMessageToAPI(req.DialogID, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Return a dummy message for now
	message := DromMessage{
		ID:        int(time.Now().Unix()),
		DialogID:  req.DialogID,
		MessageID: strconv.Itoa(int(time.Now().Unix())),
		Author:    "Я",
		Direction: "outgoing",
		Time:      time.Now().Format("2006-01-02 15:04:05"),
		Text:      req.Content,
		IsRead:    true,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
	}

	c.JSON(http.StatusCreated, message)
}

// Helper functions
func fetchDromBriefs() ([]InboxBrief, error) {
	body, err := makeDromRequest(DromListURL)
	if err != nil {
		return nil, err
	}

	var response InboxListResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse briefs: %v", err)
	}

	return response.Briefs, nil
}

func fetchDromMessages(dialogID string) ([]DromMessage, error) {
	url := fmt.Sprintf(DromViewURL, dialogID)
	body, err := makeDromRequest(url)
	if err != nil {
		return nil, err
	}

	var apiResp DromApiResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		// If not JSON, assume HTML
		apiResp.DialogHTML = string(body)
	}

	messages := extractDromMessagesFromHTML(apiResp.DialogHTML, apiResp.Interlocutor)
	for i := range messages {
		messages[i].DialogID = dialogID
	}
	return messages, nil
}

func sendDromMessageToAPI(dialogID, text string) error {
	formData := url.Values{}
	formData.Set("message", text)
	formData.Set("post", "Отправить")

	reqUrl, _ := url.Parse(DromPostURL)
	q := reqUrl.Query()
	q.Add("dialogId", dialogID)
	q.Add("json", "true")
	q.Add("ajax", "1")
	q.Add("flat-layout", "false")
	reqUrl.RawQuery = q.Encode()

	req, err := http.NewRequest("POST", reqUrl.String(), strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", DromUserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")
	req.Header.Set("Origin", "https://my.drom.ru")
	req.Header.Set("Referer", fmt.Sprintf("https://my.drom.ru/personal/messaging-modal/dialog-%s", dialogID))

	// Load cookies
	data, err := os.ReadFile(DromSessionFile)
	if err != nil {
		return fmt.Errorf("failed to read session file: %v", err)
	}
	var session SessionData
	err = json.Unmarshal(data, &session)
	if err != nil {
		return err
	}
	for _, c := range session.Cookies {
		req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value})
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return nil
}

func makeDromRequest(url string) ([]byte, error) {
	req, _ := http.NewRequest("GET", url, nil)
	req.Header.Set("User-Agent", DromUserAgent)
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	data, err := os.ReadFile(DromSessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %v", err)
	}
	var session SessionData
	err = json.Unmarshal(data, &session)
	if err != nil {
		return nil, err
	}
	for _, c := range session.Cookies {
		req.AddCookie(&http.Cookie{Name: c.Name, Value: c.Value})
	}

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func extractDromMessagesFromHTML(htmlContent, interlocutorName string) []DromMessage {
	var msgs []DromMessage
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return msgs
	}

	doc.Find(".bzr-dialog__msg-container").Each(func(i int, s *goquery.Selection) {
		id, _ := s.Attr("data-message-id")

		block := s.Find(".bzr-dialog__message")
		if block.Length() == 0 {
			return
		}

		text := strings.TrimSpace(block.Find(".bzr-dialog__text").Text())
		if text == "" {
			text = "[Вложение]"
		}

		timeVal := strings.TrimSpace(block.Find(".bzr-dialog__message-dt").Text())

		author := interlocutorName
		direction := "incoming"

		if block.HasClass("bzr-dialog__message_out") {
			author = "Я"
			direction = "outgoing"
		}

		isRead := false
		if val, ok := block.Find(".bzr-dialog__message-check").Attr("data-state"); ok && val == "read" {
			isRead = true
		}

		msgs = append(msgs, DromMessage{
			ID:        i + 1,
			DialogID:  "", // Will be set by caller
			MessageID: id,
			Author:    author,
			Direction: direction,
			Time:      timeVal,
			Text:      text,
			IsRead:    isRead,
			CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
		})
	})
	return msgs
}

// User Users
type User struct {
	ID       int    `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Initials string `json:"initials,omitempty"`
	INN      string `json:"inn,omitempty"`
	Provider string `json:"provider"`
	Role     string `json:"role"`
}

func getCurrentUser(userID string, authHeader string) (*User, error) {
	gatewayURL := "http://localhost:8080" // Assuming gateway is on 8080

	req, err := http.NewRequest("GET", gatewayURL+"/api/users/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Authorization", authHeader)
	req.Header.Set("X-User-ID", userID)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to gateway: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gateway returned error status: %d", resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &user, nil
}

func getUsers(c *gin.Context) {
	// Get users from gateway
	gatewayURL := "http://localhost:8080" // Assuming gateway is on 8080

	req, err := http.NewRequest("GET", gatewayURL+"/api/users", nil)
	if err != nil {
		log.Printf("Messaging getUsers: Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers from original request
	authHeader := c.GetHeader("Authorization")
	userIDHeader := c.GetHeader("X-User-ID")
	log.Printf("Messaging getUsers: Authorization header: %s, X-User-ID: %s", authHeader, userIDHeader)
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("X-User-ID", userIDHeader)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Messaging getUsers: Failed to connect to gateway: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to gateway"})
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	log.Printf("Messaging getUsers: Gateway response status: %d", resp.StatusCode)
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

	log.Printf("Messaging getUsers: Successfully fetched %d users", len(response.Users))
	c.JSON(http.StatusOK, gin.H{"users": response.Users})
}

func sendMessage(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Check if user is participant
	var conversation Conversation
	if err := DB.First(&conversation, conversationIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	isParticipant := false
	for _, p := range conversation.Participants {
		if p == int64(userIDInt) {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
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

	readBy := []int64{int64(userIDInt)}
	attachments := req.Attachments

	message := Message{
		ConversationID: conversationIDInt,
		SenderID:       userIDInt,
		Content:        req.Content,
		CreatedAt:      time.Now(),
		ReadBy:         readBy, // Sender has read it
		Attachments:    attachments,
	}

	if err := DB.Create(&message).Error; err != nil {
		RecordDBError("create", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Update conversation's last message info
	conversation.LastMessageAt = message.CreatedAt
	conversation.LastMessage = message.Content
	if err := DB.Save(&conversation).Error; err != nil {
		RecordDBError("update", "messaging-service")
		// Continue anyway
	}
	DB.Save(&conversation)

	RecordBusinessOperation(
		"send_message", "messaging-service", "success")

	// Invalidate cache for this conversation
	go invalidateMessagesCache(conversationIDInt)

	c.JSON(http.StatusCreated, message)
}

func sendVoiceMessage(c *gin.Context) {
	log.Printf("sendVoiceMessage: Received request for conversation %s", c.Param("id"))
	log.Printf("sendVoiceMessage: Headers: %v", c.Request.Header)
	log.Printf("sendVoiceMessage: Content-Type: %s", c.ContentType())

	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		log.Printf("sendVoiceMessage: Missing X-User-ID header")
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		log.Printf("sendVoiceMessage: Invalid user ID: %s", userID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conversationID := c.Param("id")
	conversationIDInt, err := strconv.Atoi(conversationID)
	if err != nil {
		log.Printf("sendVoiceMessage: Invalid conversation ID: %s", conversationID)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid conversation ID"})
		return
	}

	// Check if user is participant
	var conversation Conversation
	if err := DB.First(&conversation, conversationIDInt).Error; err != nil {
		log.Printf("sendVoiceMessage: Conversation not found: %d", conversationIDInt)
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	isParticipant := false
	for _, p := range conversation.Participants {
		if p == int64(userIDInt) {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		log.Printf("sendVoiceMessage: User %d not participant in conversation %d", userIDInt, conversationIDInt)
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Handle file upload
	file, err := c.FormFile("voice")
	if err != nil {
		log.Printf("sendVoiceMessage: Failed to get voice file: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Voice file is required"})
		return
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		log.Printf("sendVoiceMessage: Failed to create uploads directory: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create uploads directory"})
		return
	}

	// Generate unique filename
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".wav" // default extension
	}
	filename := fmt.Sprintf("voice_%d_%d%s", conversationIDInt, time.Now().Unix(), ext)
	filePath := filepath.Join("./uploads", filename)

	log.Printf("sendVoiceMessage: Saving file to: %s", filePath)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		log.Printf("sendVoiceMessage: Failed to save file: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save voice file"})
		return
	}

	voiceURL := fmt.Sprintf("%s/uploads/%s", os.Getenv("API_BASE_URL"), filename)
	log.Printf("sendVoiceMessage: Voice URL: %s", voiceURL)

	readBy := []int64{int64(userIDInt)}

	message := Message{
		ConversationID: conversationIDInt,
		SenderID:       userIDInt,
		Content:        "Voice message",
		MessageType:    "voice",
		VoiceURL:       voiceURL,
		CreatedAt:      time.Now(),
		ReadBy:         readBy, // Sender has read it
	}

	if err := DB.Create(&message).Error; err != nil {
		RecordDBError("create", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send voice message"})
		return
	}

	// Update conversation's last message info
	conversation.LastMessageAt = message.CreatedAt
	conversation.LastMessage = "Voice message"
	if err := DB.Save(&conversation).Error; err != nil {
		RecordDBError("update", "messaging-service")
		// Continue anyway
	}
	DB.Save(&conversation)

	RecordBusinessOperation(
		"send_voice_message", "messaging-service", "success")

	// Invalidate cache for this conversation
	go invalidateMessagesCache(conversationIDInt)

	c.JSON(http.StatusCreated, message)
}

func markMessageRead(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	messageID := c.Param("id")
	messageIDInt, err := strconv.Atoi(messageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var message Message
	if err := DB.First(&message, messageIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Check if user is participant in conversation
	var conversation Conversation
	if err := DB.First(&conversation, message.ConversationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	isParticipant := false
	for _, p := range conversation.Participants {
		if p == int64(userIDInt) {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Add user to read_by if not already there
	alreadyRead := false
	for _, reader := range message.ReadBy {
		if reader == int64(userIDInt) {
			alreadyRead = true
			break
		}
	}

	if !alreadyRead {
		message.ReadBy = append(message.ReadBy, int64(userIDInt))
		if err := DB.Save(&message).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark message as read"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"message": "Message marked as read"})
}

func deleteMessage(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	messageID := c.Param("id")
	messageIDInt, err := strconv.Atoi(messageID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid message ID"})
		return
	}

	var message Message
	if err := DB.First(&message, messageIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Check if user is the sender of the message
	if message.SenderID != userIDInt {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := DB.Delete(&message).Error; err != nil {
		RecordDBError("delete", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete message"})
		return
	}

	RecordBusinessOperation(
		"delete_message", "messaging-service", "success")

	c.JSON(http.StatusOK, gin.H{"message": "Message deleted"})
}

// Reactions
func addReaction(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	messageID := c.Param("id")
	messageIDInt, err := strconv.Atoi(messageID)
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

	if req.Emoji == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Emoji is required"})
		return
	}

	// Check if message exists
	var message Message
	if err := DB.First(&message, messageIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Message not found"})
		return
	}

	// Check if user is participant in conversation
	var conversation Conversation
	if err := DB.First(&conversation, message.ConversationID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Conversation not found"})
		return
	}

	isParticipant := false
	for _, p := range conversation.Participants {
		if p == int64(userIDInt) {
			isParticipant = true
			break
		}
	}

	if !isParticipant {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	// Check if reaction already exists
	var existingReaction Reaction
	if err := DB.Where("message_id = ? AND user_id = ? AND emoji = ?", messageIDInt, userIDInt, req.Emoji).First(&existingReaction).Error; err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "Reaction already exists"})
		return
	}

	reaction := Reaction{
		MessageID: messageIDInt,
		UserID:    userIDInt,
		Emoji:     req.Emoji,
	}

	if err := DB.Create(&reaction).Error; err != nil {
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

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	reactionId := c.Param("reactionId")
	reactionIdInt, err := strconv.Atoi(reactionId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reaction ID"})
		return
	}

	var reaction Reaction
	if err := DB.First(&reaction, reactionIdInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reaction not found"})
		return
	}

	// Check if user owns the reaction
	if reaction.UserID != userIDInt {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	if err := DB.Delete(&reaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove reaction"})
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

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var notifications []Notification
	if err := DB.Where("user_id = ?", userIDInt).Order("created_at DESC").Limit(50).Find(&notifications).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch notifications"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"notifications": notifications})
}

func markNotificationRead(c *gin.Context) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	notificationID := c.Param("id")
	notificationIDInt, err := strconv.Atoi(notificationID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid notification ID"})
		return
	}

	var notification Notification
	if err := DB.First(&notification, notificationIDInt).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Notification not found"})
		return
	}

	// Check if user owns the notification
	if notification.UserID != userIDInt {
		c.JSON(http.StatusForbidden, gin.H{"error": "Access denied"})
		return
	}

	notification.Read = true
	if err := DB.Save(&notification).Error; err != nil {
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

	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	if err := DB.Model(&Notification{}).Where("user_id = ? AND read = ?", userIDUint, false).Update("read", true).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to mark all notifications as read"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "All notifications marked as read"})
}

// User status
func getUserStatuses(c *gin.Context) {
	var statuses []UserStatus
	if err := DB.Find(&statuses).Error; err != nil {
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

	var req struct {
		IsOnline bool `json:"is_online"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}

	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	status := UserStatus{
		UserID:   int(uint(userIDUint)),
		IsOnline: req.IsOnline,
		LastSeen: time.Now(),
	}

	if err := DB.Save(&status).Error; err != nil {
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

	userIDUint, err := strconv.ParseUint(userID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query required"})
		return
	}

	// Search in conversations where user is participant
	var messages []Message
	if err := DB.Joins("JOIN conversations c ON messages.conversation_id = c.id").
		Where("c.participants @> ARRAY[?]::integer[] AND messages.search_vector @@ plainto_tsquery('russian', ?)", userIDUint, query).
		Order("messages.created_at DESC").
		Limit(50).
		Preload("Reactions").
		Preload("Mentions").
		Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to search messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
}
