package main

import (
	"time"
)

type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Initials string `json:"initials,omitempty"`
	INN      string `json:"inn,omitempty"`
	Provider string `json:"provider"`
	Role     string `json:"role"`
	Password string `json:"-"`
}

type LoginResponse struct {
	User         User   `json:"user"`
	Message      string `json:"message"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

type UserActivityLog struct {
	ID           int64      `json:"id"`
	UserID       int64      `json:"user_id"`
	UserName     string     `json:"user_name"`
	UserEmail    string     `json:"user_email"`
	Action       string     `json:"action"`
	ResourceType string     `json:"resource_type"`
	ResourceID   *int64     `json:"resource_id,omitempty"`
	Details      string     `json:"details,omitempty"`
	IPAddress    string     `json:"ip_address,omitempty"`
	UserAgent    string     `json:"user_agent,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}
