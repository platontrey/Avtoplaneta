package domain

import (
	"time"
)

// User модель учетной записи пользователя
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

// LoginResponse ответ при успешной аутентификации
type LoginResponse struct {
	User         User   `json:"user"`
	Message      string `json:"message"`
	Token        string `json:"token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// UserActivityLog запись в журнале активности пользователей
type UserActivityLog struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	UserName     string    `json:"user_name"`
	UserEmail    string    `json:"user_email"`
	Action       string    `json:"action"`
	ResourceType string    `json:"resource_type"`
	ResourceID   *int64    `json:"resource_id,omitempty"`
	Details      string    `json:"details,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
	UserAgent    string    `json:"user_agent,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// CreateUserRequest запрос на создание пользователя
type CreateUserRequest struct {
	Email    string
	Name     string
	Initials string
	INN      string
	Password string
	Role     string
}

// UpdateUserRequest запрос на обновление пользователя
type UpdateUserRequest struct {
	Name     string
	Email    string
	Initials string
	INN      string
	Role     string
}

// UpdateUserParams параметры для репозитория обновления пользователя
type UpdateUserParams struct {
	Name     string
	Email    string
	Initials string
	INN      string
	Role     string
}

// ActivityLogFilters фильтры для выборки логов активности
type ActivityLogFilters struct {
	UserID       *int64
	Action       string
	ResourceType string
	UsefulOnly   bool
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

// NonUsefulActions действия, отфильтровываемые при UsefulOnly
var NonUsefulActions = []string{
	"view_parts",
	"search_parts",
	"view_part_details",
	"view_vehicle_catalog",
	"view_part_catalog",
	"get_user_info",
}
