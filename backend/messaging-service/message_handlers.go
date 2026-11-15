package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// Messages
func getMessages(c *gin.Context) {
	userIDInt, ok := parseUserID(c)
	if !ok {
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
	userIDInt, ok := parseUserID(c)
	if !ok {
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
	userIDInt, ok := parseUserID(c)
	if !ok {
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
	userIDInt, ok := parseUserID(c)
	if !ok {
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
	userIDInt, ok := parseUserID(c)
	if !ok {
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
	userIDInt, ok := parseUserID(c)
	if !ok {
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
	userIDInt, ok := parseUserID(c)
	if !ok {
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
