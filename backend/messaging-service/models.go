package main

import (
	"time"
)

// Conversation представляет чат между пользователями
type Conversation struct {
	ID             int64     `json:"id"`
	Participants   []int64   `json:"participants"` // Массив ID пользователей
	Title          string    `json:"title,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	LastMessageAt  time.Time `json:"last_message_at"`
	LastMessage    string    `json:"last_message,omitempty"`
	UnreadCount    int64     `json:"unread_count,omitempty"`
}

// Message представляет сообщение в чате
type Message struct {
	ID             int64      `json:"id"`
	ConversationID int64      `json:"conversation_id"`
	SenderID       int64      `json:"sender_id"`
	Content        string     `json:"content"`
	MessageType    string     `json:"message_type"` // text, voice, image, file
	VoiceURL       string     `json:"voice_url,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	ReadBy         []int64    `json:"read_by"`
	Attachments    []string   `json:"attachments,omitempty"`
	Mentions       []Mention  `json:"mentions,omitempty"`
	Reactions      []Reaction `json:"reactions,omitempty"`
}

// Mention представляет ссылку на запчасть или заказ в сообщении
type Mention struct {
	ID         int64  `json:"id"`
	MessageID  int64  `json:"message_id"`
	Type       string `json:"type"` // "part" или "order"
	ResourceID int64  `json:"resource_id"`
	Text       string `json:"text"`
}

// Reaction представляет реакцию на сообщение
type Reaction struct {
	ID        int64  `json:"id"`
	MessageID int64  `json:"message_id"`
	UserID    int64  `json:"user_id"`
	Emoji     string `json:"emoji"`
}

// Notification представляет уведомление пользователя
type Notification struct {
	ID        int64     `json:"id"`
	UserID    int64     `json:"user_id"`
	Type      string    `json:"type"` // "message", "mention", "reaction"
	Message   string    `json:"message"`
	Read      bool      `json:"read"`
	CreatedAt time.Time `json:"created_at"`
	RelatedID int64     `json:"related_id"` // ID связанного объекта (conversation, message)
}

// UserStatus представляет онлайн статус пользователя
type UserStatus struct {
	ID       int64     `json:"id"`
	UserID   int64     `json:"user_id"`
	IsOnline bool      `json:"is_online"`
	LastSeen time.Time `json:"last_seen"`
}

// User represents a user from Auth service (fetched via gateway)
type User struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	Role  string `json:"role"`
}
