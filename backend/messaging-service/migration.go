package main

import (
	"log"

	"gorm.io/gorm"
)

// RunMigrations выполняет кастомные миграции для messaging-service
func RunMigrations(db *gorm.DB) {
	// Индексы для conversations
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_conversations_participants ON conversations USING GIN (participants)`).Error; err != nil {
		log.Printf("Failed to create index idx_conversations_participants: %v", err)
	}

	// Индексы для messages
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_conversation_id ON messages (conversation_id)`).Error; err != nil {
		log.Printf("Failed to create index idx_messages_conversation_id: %v", err)
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages (sender_id)`).Error; err != nil {
		log.Printf("Failed to create index idx_messages_sender_id: %v", err)
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_read_by ON messages USING GIN (read_by)`).Error; err != nil {
		log.Printf("Failed to create index idx_messages_read_by: %v", err)
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_created_at ON messages (created_at DESC)`).Error; err != nil {
		log.Printf("Failed to create index idx_messages_created_at: %v", err)
	}

	// Полнотекстовый поиск для messages
	if err := db.Exec(`ALTER TABLE messages ADD COLUMN IF NOT EXISTS search_vector tsvector`).Error; err != nil {
		log.Printf("Failed to add search_vector column: %v", err)
	}
	if err := db.Exec(`UPDATE messages SET search_vector = to_tsvector('russian', content) WHERE search_vector IS NULL`).Error; err != nil {
		log.Printf("Failed to update search_vector: %v", err)
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_messages_search ON messages USING GIN (search_vector)`).Error; err != nil {
		log.Printf("Failed to create index idx_messages_search: %v", err)
	}

	// Индексы для notifications
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_user_id ON notifications (user_id)`).Error; err != nil {
		log.Printf("Failed to create index idx_notifications_user_id: %v", err)
	}
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_notifications_created_at ON notifications (created_at DESC)`).Error; err != nil {
		log.Printf("Failed to create index idx_notifications_created_at: %v", err)
	}

	// Индексы для user_status
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_user_status_user_id ON user_status (user_id)`).Error; err != nil {
		log.Printf("Failed to create index idx_user_status_user_id: %v", err)
	}

	// Индексы для reactions
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_reactions_message_id ON reactions (message_id)`).Error; err != nil {
		log.Printf("Failed to create index idx_reactions_message_id: %v", err)
	}

	// Индексы для mentions
	if err := db.Exec(`CREATE INDEX IF NOT EXISTS idx_mentions_message_id ON mentions (message_id)`).Error; err != nil {
		log.Printf("Failed to create index idx_mentions_message_id: %v", err)
	}

	// Уникальный индекс для предотвращения дубликатов реакций
	if err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_reactions_unique ON reactions (message_id, user_id, emoji)`).Error; err != nil {
		log.Printf("Failed to create unique index idx_reactions_unique: %v", err)
	}

	log.Println("Messaging service migrations completed")
}