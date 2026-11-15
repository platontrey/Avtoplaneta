package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

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
	GetDromBriefs(ctx context.Context) ([]DromDialog, error)
	GetDromMessages(ctx context.Context, dialogID string) ([]DromMessage, *DromDialog, error)
	SendDromMessage(ctx context.Context, dialogID string, content string) (*DromMessage, error)
	GetDromAuthStatus(ctx context.Context) (map[string]interface{}, error)
	LoginDrom(ctx context.Context, login, password, code string) (map[string]interface{}, error)
	SaveDromSession(ctx context.Context, rawSession []byte) error
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
	msg, err := s.repo.GetMessageByID(ctx, id)
	if err != nil {
		return err
	}
	if msg == nil {
		return errors.New("message not found")
	}

	if msg.SenderID != userID {
		return errors.New("access denied")
	}

	if err := s.repo.DeleteMessage(ctx, id); err != nil {
		return err
	}

	go s.invalidateMessagesCache(context.Background(), msg.ConversationID)
	return nil
}

func (s *messagingService) MarkMessageRead(ctx context.Context, messageID int64, userID int64) error {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return err
	}
	if msg == nil {
		return errors.New("message not found")
	}

	if err := s.repo.AddUserToReadBy(ctx, messageID, userID); err != nil {
		return err
	}

	go s.invalidateMessagesCache(context.Background(), msg.ConversationID)
	return nil
}

// Reactions
func (s *messagingService) AddReaction(ctx context.Context, messageID int64, userID int64, emoji string) (*Reaction, error) {
	msg, err := s.repo.GetMessageByID(ctx, messageID)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, errors.New("message not found")
	}

	reaction := Reaction{
		MessageID: messageID,
		UserID:    userID,
		Emoji:     emoji,
	}

	if err := s.repo.CreateReaction(ctx, &reaction); err != nil {
		return nil, err
	}

	go s.invalidateMessagesCache(context.Background(), msg.ConversationID)
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

	msg, err := s.repo.GetMessageByID(ctx, reaction.MessageID)
	if err != nil {
		return err
	}
	if msg == nil {
		return errors.New("message not found")
	}

	if err := s.repo.DeleteReaction(ctx, reactionID); err != nil {
		return err
	}

	go s.invalidateMessagesCache(context.Background(), msg.ConversationID)
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
