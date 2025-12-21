package main

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func setupRoutes(r *gin.Engine) {
	api := r.Group("/api/messaging")
	{
		// Conversations
		api.GET("/conversations", getConversations)
		api.POST("/conversations", createConversation)
		api.GET("/conversations/:id", getConversation)
		api.DELETE("/conversations/:id", deleteConversation)

		// Messages
		api.GET("/conversations/:id/messages", getMessages)
		api.POST("/conversations/:id/messages", sendMessage)
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
	query := DB.Where("participants @> ARRAY[?]::integer[]", userIDInt).Order("last_message_at DESC")

	if err := query.Find(&conversations).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch conversations"})
		return
	}

	// Calculate unread count for each conversation
	for i := range conversations {
		var unreadCount int64
		DB.Model(&Message{}).Where("conversation_id = ? AND sender_id != ? AND NOT (read_by @> ARRAY[?]::integer[])", conversations[i].ID, userIDInt, userIDInt).Count(&unreadCount)
		conversations[i].UnreadCount = int(unreadCount)
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create conversation"})
		return
	}

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

	var messages []Message
	if err := DB.Where("conversation_id = ?", conversationIDInt).Order("created_at ASC").Preload("Reactions").Preload("Mentions").Find(&messages).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch messages"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": messages})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to send message"})
		return
	}

	// Update conversation's last message info
	conversation.LastMessageAt = message.CreatedAt
	conversation.LastMessage = message.Content
	DB.Save(&conversation)

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
		Where("c.participants @> ARRAY[?]::integer[] AND messages.content ILIKE ?", userIDUint, "%"+query+"%").
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
