package main

import (
	"time"
)

type User struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	Email    string `json:"email" gorm:"unique;not null"`
	Name     string `json:"name"`
	Initials string `json:"initials,omitempty"`             // Инициалы пользователя
	INN      string `json:"inn,omitempty"`                  // ИНН пользователя
	Provider string `json:"provider"`                       // "google" или "local"
	Role     string `json:"role" gorm:"default:'operator'"` // "admin", "manager", "operator"
	Password string `json:"-" gorm:"not null;default:''"`   // Пароль не сериализуется в JSON
}

type LoginResponse struct {
	User    User   `json:"user"`
	Message string `json:"message"`
}

type UserActivityLog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	UserID       uint      `json:"user_id" gorm:"not null"`
	UserName     string    `json:"user_name" gorm:"not null"`
	UserEmail    string    `json:"user_email" gorm:"not null"`
	Action       string    `json:"action" gorm:"not null"`        // login, logout, create_part, etc.
	ResourceType string    `json:"resource_type" gorm:"not null"` // part, order, user, photo, system
	ResourceID   *uint     `json:"resource_id,omitempty"`         // ID ресурса, если применимо
	Details      string    `json:"details,omitempty"`             // Дополнительная информация
	IPAddress    string    `json:"ip_address,omitempty"`          // IP адрес пользователя
	UserAgent    string    `json:"user_agent,omitempty"`          // User-Agent браузера
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
}

