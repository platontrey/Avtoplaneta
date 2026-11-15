package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func parseUserID(c *gin.Context) (int64, bool) {
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return 0, false
	}

	userIDInt, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return 0, false
	}
	return userIDInt, true
}

// Conversations
func getConversations(c *gin.Context) {
	userIDInt, ok := parseUserID(c)
	if !ok {
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
	userIDInt, ok := parseUserID(c)
	if !ok {
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

	if err := svc.DeleteConversation(c.Request.Context(), conversationIDInt); err != nil {
		RecordDBError("delete_conversation", "messaging-service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete conversation"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Conversation deleted"})
}

func removeParticipant(c *gin.Context) {
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

// Notifications
func getNotifications(c *gin.Context) {
	userIDInt, ok := parseUserID(c)
	if !ok {
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
	userIDInt, ok := parseUserID(c)
	if !ok {
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
	userIDInt, ok := parseUserID(c)
	if !ok {
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
	userIDInt, ok := parseUserID(c)
	if !ok {
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
