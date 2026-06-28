package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/redis/go-redis/v9"
)

type MessagingService interface {
	GetConversations(ctx context.Context, userID int64) ([]Conversation, error)
	GetConversation(ctx context.Context, id int64, userID int64) (*Conversation, error)
	CreateConversation(ctx context.Context, title string, participants []int64) (*Conversation, error)
	UpdateConversation(ctx context.Context, id int64, title string, participants []int64) (*Conversation, error)
	DeleteConversation(ctx context.Context, id int64) error

	GetMessages(ctx context.Context, conversationID int64) ([]Message, error)
	SendMessage(ctx context.Context, conversationID int64, senderID int64, content string, attachments []string) (*Message, error)
	SendVoiceMessage(ctx context.Context, conversationID int64, senderID int64, voiceURL string) (*Message, error)
	DeleteMessage(ctx context.Context, id int64, userID int64) error
	MarkMessageRead(ctx context.Context, messageID int64, userID int64) error

	AddReaction(ctx context.Context, messageID int64, userID int64, emoji string) (*Reaction, error)
	RemoveReaction(ctx context.Context, reactionID int64, userID int64) error

	GetNotifications(ctx context.Context, userID int64) ([]Notification, error)
	MarkNotificationRead(ctx context.Context, id int64) error
	MarkAllNotificationsRead(ctx context.Context, userID int64) error

	GetUserStatuses(ctx context.Context, userIDs []int64) ([]UserStatus, error)
	UpdateUserStatus(ctx context.Context, userID int64, isOnline bool) error
	GetUserStatus(ctx context.Context, userID int64) (*UserStatus, error)

	SearchMessages(ctx context.Context, userID int64, query string, limit, offset int) ([]Message, error)

	// Drom integration methods
	GetDromBriefs(ctx context.Context) ([]InboxBrief, error)
	GetDromMessages(ctx context.Context, dialogID string) ([]DromMessage, error)
	SendDromMessage(ctx context.Context, dialogID string, content string) (*DromMessage, error)
}

type messagingService struct {
	repo  MessagingRepository
	redis *redis.Client
}

func NewMessagingService(repo MessagingRepository, redisClient *redis.Client) MessagingService {
	return &messagingService{
		repo:  repo,
		redis: redisClient,
	}
}

// Cache helpers
func (s *messagingService) getMessagesFromCache(ctx context.Context, conversationID int64) ([]Message, error) {
	if s.redis == nil {
		return nil, errors.New("redis not initialized")
	}
	key := fmt.Sprintf("messages:%d", conversationID)
	val, err := s.redis.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var messages []Message
	if err := json.Unmarshal([]byte(val), &messages); err != nil {
		return nil, err
	}
	return messages, nil
}

func (s *messagingService) cacheMessages(ctx context.Context, conversationID int64, messages []Message) {
	if s.redis == nil {
		return
	}
	key := fmt.Sprintf("messages:%d", conversationID)
	data, err := json.Marshal(messages)
	if err != nil {
		return
	}
	s.redis.Set(ctx, key, string(data), 10*time.Minute)
}

func (s *messagingService) invalidateMessagesCache(ctx context.Context, conversationID int64) {
	if s.redis == nil {
		return
	}
	key := fmt.Sprintf("messages:%d", conversationID)
	s.redis.Del(ctx, key)
}

// Conversations
func (s *messagingService) GetConversations(ctx context.Context, userID int64) ([]Conversation, error) {
	return s.repo.GetConversations(ctx, userID)
}

func (s *messagingService) GetConversation(ctx context.Context, id int64, userID int64) (*Conversation, error) {
	conv, err := s.repo.GetConversation(ctx, id)
	if err != nil {
		return nil, err
	}

	// Validate access
	if userID > 0 {
		isParticipant := false
		for _, p := range conv.Participants {
			if p == userID {
				isParticipant = true
				break
			}
		}
		if !isParticipant {
			return nil, errors.New("access denied")
		}
	}

	return conv, nil
}

func (s *messagingService) CreateConversation(ctx context.Context, title string, participants []int64) (*Conversation, error) {
	conv := Conversation{
		Participants:  participants,
		Title:         title,
		CreatedAt:     time.Now(),
		LastMessageAt: time.Now(),
	}
	err := s.repo.CreateConversation(ctx, &conv)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

func (s *messagingService) UpdateConversation(ctx context.Context, id int64, title string, participants []int64) (*Conversation, error) {
	// GORM structure was update title or participants
	return s.repo.UpdateConversation(ctx, id, title, time.Now(), "", participants)
}

func (s *messagingService) DeleteConversation(ctx context.Context, id int64) error {
	// First invalidate cache
	s.invalidateMessagesCache(ctx, id)
	return s.repo.DeleteConversation(ctx, id)
}

// Messages
func (s *messagingService) GetMessages(ctx context.Context, conversationID int64) ([]Message, error) {
	// Try cache first
	cached, err := s.getMessagesFromCache(ctx, conversationID)
	if err == nil && len(cached) > 0 {
		return cached, nil
	}

	// Database query
	messages, err := s.repo.GetMessages(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	// Cache messages in background
	go s.cacheMessages(context.Background(), conversationID, messages)

	return messages, nil
}

func (s *messagingService) SendMessage(ctx context.Context, conversationID int64, senderID int64, content string, attachments []string) (*Message, error) {
	// 1. Verify conversation exists and sender is participant
	conv, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	isParticipant := false
	for _, p := range conv.Participants {
		if p == senderID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, errors.New("access denied")
	}

	// 2. Create message
	msg := Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        content,
		MessageType:    "text",
		CreatedAt:      time.Now(),
		ReadBy:         []int64{senderID},
		Attachments:    attachments,
	}

	if err := s.repo.CreateMessage(ctx, &msg); err != nil {
		return nil, err
	}

	// 3. Update last message at in conversation (fixed duplicate Save() bug here)
	if err := s.repo.UpdateConversationLastMessage(ctx, conversationID, msg.CreatedAt, msg.Content); err != nil {
		log.Printf("Warning: failed to update conversation last message: %v", err)
	}

	// 4. Invalidate cache
	go s.invalidateMessagesCache(context.Background(), conversationID)

	return &msg, nil
}

func (s *messagingService) SendVoiceMessage(ctx context.Context, conversationID int64, senderID int64, voiceURL string) (*Message, error) {
	// 1. Verify conversation exists
	conv, err := s.repo.GetConversation(ctx, conversationID)
	if err != nil {
		return nil, err
	}

	isParticipant := false
	for _, p := range conv.Participants {
		if p == senderID {
			isParticipant = true
			break
		}
	}
	if !isParticipant {
		return nil, errors.New("access denied")
	}

	// 2. Create message
	msg := Message{
		ConversationID: conversationID,
		SenderID:       senderID,
		Content:        "[Голосовое сообщение]",
		MessageType:    "voice",
		VoiceURL:       voiceURL,
		CreatedAt:      time.Now(),
		ReadBy:         []int64{senderID},
		Attachments:    []string{},
	}

	if err := s.repo.CreateMessage(ctx, &msg); err != nil {
		return nil, err
	}

	// 3. Update last message at
	if err := s.repo.UpdateConversationLastMessage(ctx, conversationID, msg.CreatedAt, msg.Content); err != nil {
		log.Printf("Warning: failed to update conversation last message: %v", err)
	}

	// 4. Invalidate cache
	go s.invalidateMessagesCache(context.Background(), conversationID)

	return &msg, nil
}

func (s *messagingService) DeleteMessage(ctx context.Context, id int64, userID int64) error {
	// First fetch the message to find the conversation ID
	// GORM code loaded first. We need to do raw query or sqlc or we can fetch conversation.
	// Actually we should write a query to delete message only if sender matches or check sender ID.
	// Since we delete by message ID in REST:
	// Let's get conversation ID to invalidate cache
	// We didn't define a GetMessageByID in sqlc, but we can easily add it or query directly,
	// or we can just use the squirrel builder / raw DB. pool.QueryRow().
	var conversationID int64
	var senderID int64
	err := s.repo.(*messagingRepository).pool.QueryRow(ctx, "SELECT conversation_id, sender_id FROM messages WHERE id = $1", id).Scan(&conversationID, &senderID)
	if err != nil {
		return err
	}

	if senderID != userID {
		return errors.New("access denied")
	}

	if err := s.repo.DeleteMessage(ctx, id); err != nil {
		return err
	}

	go s.invalidateMessagesCache(context.Background(), conversationID)
	return nil
}

func (s *messagingService) MarkMessageRead(ctx context.Context, messageID int64, userID int64) error {
	// Get message conversation ID to invalidate cache
	var conversationID int64
	err := s.repo.(*messagingRepository).pool.QueryRow(ctx, "SELECT conversation_id FROM messages WHERE id = $1", messageID).Scan(&conversationID)
	if err != nil {
		return err
	}

	if err := s.repo.AddUserToReadBy(ctx, messageID, userID); err != nil {
		return err
	}

	go s.invalidateMessagesCache(context.Background(), conversationID)
	return nil
}

// Reactions
func (s *messagingService) AddReaction(ctx context.Context, messageID int64, userID int64, emoji string) (*Reaction, error) {
	// Get conversation ID to invalidate cache
	var conversationID int64
	err := s.repo.(*messagingRepository).pool.QueryRow(ctx, "SELECT conversation_id FROM messages WHERE id = $1", messageID).Scan(&conversationID)
	if err != nil {
		return nil, err
	}

	reaction := Reaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}

	if err := s.repo.CreateReaction(ctx, &reaction); err != nil {
		return nil, err
	}

	go s.invalidateMessagesCache(context.Background(), conversationID)
	return &reaction, nil
}

func (s *messagingService) RemoveReaction(ctx context.Context, reactionID int64, userID int64) error {
	reaction, err := s.repo.GetReactionByID(ctx, reactionID)
	if err != nil {
		return err
	}

	if userID > 0 && reaction.UserID != userID {
		return errors.New("access denied")
	}

	var conversationID int64
	err = s.repo.(*messagingRepository).pool.QueryRow(ctx, "SELECT conversation_id FROM messages WHERE id = $1", reaction.MessageID).Scan(&conversationID)
	if err != nil {
		return err
	}

	if err := s.repo.DeleteReaction(ctx, reactionID); err != nil {
		return err
	}

	go s.invalidateMessagesCache(context.Background(), conversationID)
	return nil
}

// Notifications
func (s *messagingService) GetNotifications(ctx context.Context, userID int64) ([]Notification, error) {
	return s.repo.GetNotifications(ctx, userID)
}

func (s *messagingService) MarkNotificationRead(ctx context.Context, id int64) error {
	return s.repo.MarkNotificationRead(ctx, id)
}

func (s *messagingService) MarkAllNotificationsRead(ctx context.Context, userID int64) error {
	return s.repo.MarkAllNotificationsRead(ctx, userID)
}

// UserStatuses
func (s *messagingService) GetUserStatuses(ctx context.Context, userIDs []int64) ([]UserStatus, error) {
	return s.repo.GetUserStatuses(ctx, userIDs)
}

func (s *messagingService) UpdateUserStatus(ctx context.Context, userID int64, isOnline bool) error {
	status := UserStatus{
		UserID:   userID,
		IsOnline: isOnline,
		LastSeen: time.Now(),
	}
	return s.repo.UpsertUserStatus(ctx, &status)
}

func (s *messagingService) GetUserStatus(ctx context.Context, userID int64) (*UserStatus, error) {
	return s.repo.GetUserStatusByUserID(ctx, userID)
}

// Search
func (s *messagingService) SearchMessages(ctx context.Context, userID int64, query string, limit, offset int) ([]Message, error) {
	return s.repo.SearchMessages(ctx, userID, query, limit, offset)
}

// === Drom Scraper Integration ===
type DromDialog struct {
	ID            int64   `json:"id"`
	DialogID      string  `json:"dialog_id"`
	Interlocutor  string  `json:"interlocutor"`
	CreatedAt     string  `json:"created_at"`
	LastMessageAt string  `json:"last_message_at"`
	LastMessage   *string `json:"last_message,omitempty"`
}

type DromMessage struct {
	ID        int64  `json:"id"`
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
	IsUnread     bool   `json:"isUnread"`
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

func makeDromRequest(reqURL string) ([]byte, error) {
	// Read session file
	data, err := os.ReadFile(DromSessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read drom session: %v", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("failed to parse drom session: %v", err)
	}

	req, err := http.NewRequest("GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", DromUserAgent)
	for _, cookie := range session.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("drom API returned status: %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}

func (s *messagingService) GetDromBriefs(ctx context.Context) ([]InboxBrief, error) {
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

func (s *messagingService) GetDromMessages(ctx context.Context, dialogID string) ([]DromMessage, error) {
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

func (s *messagingService) SendDromMessage(ctx context.Context, dialogID string, content string) (*DromMessage, error) {
	if err := sendDromMessageToAPI(dialogID, content); err != nil {
		return nil, err
	}

	message := DromMessage{
		ID:        time.Now().Unix(),
		DialogID:  dialogID,
		MessageID: strconv.FormatInt(time.Now().Unix(), 10),
		Author:    "Я",
		Direction: "outgoing",
		Time:      time.Now().Format("2006-01-02 15:04:05"),
		Text:      content,
		IsRead:    true,
		CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"),
	}

	return &message, nil
}

func sendDromMessageToAPI(dialogID, text string) error {
	formData := url.Values{}
	formData.Set("dialogId", dialogID)
	formData.Set("text", text)
	formData.Set("ajax", "1")
	formData.Set("action", "send")

	// Read session file
	data, err := os.ReadFile(DromSessionFile)
	if err != nil {
		return fmt.Errorf("failed to read drom session: %v", err)
	}

	var session SessionData
	if err := json.Unmarshal(data, &session); err != nil {
		return fmt.Errorf("failed to parse drom session: %v", err)
	}

	req, err := http.NewRequest("POST", DromPostURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", DromUserAgent)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, cookie := range session.Cookies {
		req.AddCookie(&http.Cookie{
			Name:  cookie.Name,
			Value: cookie.Value,
		})
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("drom post message returned status: %d", resp.StatusCode)
	}

	return nil
}

func extractDromMessagesFromHTML(htmlContent, interlocutorName string) []DromMessage {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		log.Printf("Failed to parse Drom HTML: %v", err)
		return nil
	}

	var messages []DromMessage
	doc.Find(".message-item").Each(func(i int, s *goquery.Selection) {
		messageID, _ := s.Attr("data-message-id")
		text := s.Find(".message-text").Text()
		timeStr := s.Find(".message-time").Text()
		isOutgoing := s.HasClass("message-outgoing")

		author := interlocutorName
		direction := "incoming"
		if isOutgoing {
			author = "Я"
			direction = "outgoing"
		}

		messages = append(messages, DromMessage{
			ID:        int64(i + 1),
			MessageID: messageID,
			Author:    author,
			Direction: direction,
			Time:      timeStr,
			Text:      strings.TrimSpace(text),
			IsRead:    true,
			CreatedAt: time.Now().Format("2006-01-02T15:04:05Z"), // Fallback
		})
	})

	return messages
}
