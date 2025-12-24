package main

import (
	"time"

	"github.com/lib/pq"
)

// Conversation представляет чат между пользователями
type Conversation struct {
	ID             int            `json:"id" gorm:"primaryKey"`
	Participants   pq.Int64Array  `json:"participants" gorm:"type:integer[]"` // Массив ID пользователей
	Title          string         `json:"title,omitempty"`                    // Название чата (опционально)
	CreatedAt      time.Time      `json:"created_at"`
	LastMessageAt  time.Time      `json:"last_message_at"`
	LastMessage    string         `json:"last_message,omitempty"` // Последнее сообщение для превью
	UnreadCount    int            `json:"unread_count,omitempty" gorm:"-"` // Для фронтенда
}

// Message представляет сообщение в чате
type Message struct {
	ID             int            `json:"id" gorm:"primaryKey"`
	ConversationID int            `json:"conversation_id"`
	SenderID       int            `json:"sender_id"`
	Content        string         `json:"content"`
	MessageType    string         `json:"message_type" gorm:"default:'text'"` // text, voice, image, file
	VoiceURL       string         `json:"voice_url,omitempty"`                // URL для голосового сообщения
	CreatedAt      time.Time      `json:"created_at"`
	ReadBy         pq.Int64Array  `json:"read_by" gorm:"type:integer[]"` // Кто прочитал
	Attachments    []string `json:"attachments,omitempty" gorm:"type:text[]"` // Файлы
	Mentions       []Mention      `json:"mentions,omitempty" gorm:"foreignKey:MessageID"` // Ссылки на запчасти/заказы
	Reactions      []Reaction     `json:"reactions,omitempty" gorm:"foreignKey:MessageID"` // Реакции
}

// Mention представляет ссылку на запчасть или заказ в сообщении
type Mention struct {
	ID         int    `json:"id" gorm:"primaryKey"`
	MessageID  int    `json:"message_id"`
	Type       string `json:"type"` // "part" или "order"
	ResourceID int    `json:"resource_id"`
	Text       string `json:"text"` // Отображаемый текст
}

// Reaction представляет реакцию на сообщение
type Reaction struct {
	ID        int    `json:"id" gorm:"primaryKey"`
	MessageID int    `json:"message_id"`
	UserID    int    `json:"user_id"`
	Emoji     string `json:"emoji"`
}

// Notification представляет уведомление пользователя
type Notification struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	UserID     int       `json:"user_id"`
	Type       string    `json:"type"` // "message", "mention", "reaction"
	Message    string    `json:"message"`
	Read       bool      `json:"read" gorm:"default:false"`
	CreatedAt  time.Time `json:"created_at"`
	RelatedID  int       `json:"related_id"` // ID связанного объекта (conversation, message)
}

// UserStatus представляет онлайн статус пользователя
type UserStatus struct {
	UserID    int       `json:"user_id" gorm:"primaryKey"`
	IsOnline  bool      `json:"is_online"`
	LastSeen  time.Time `json:"last_seen"`
}
