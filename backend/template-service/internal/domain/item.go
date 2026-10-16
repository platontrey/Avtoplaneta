package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrItemNotFound = errors.New("item not found")
	ErrInvalidInput = errors.New("invalid item input")
)

// Item представляет основную доменную модель сервиса
type Item struct {
	ID          int64             `json:"id"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Metadata    map[string]any    `json:"metadata"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	DeletedAt   *time.Time        `json:"deleted_at,omitempty"`
}

// Validate проверяет доменные инварианты сущности
func (i *Item) Validate() error {
	if strings.TrimSpace(i.Name) == "" {
		return errors.New("название элемента не может быть пустым")
	}
	if len(i.Name) > 255 {
		return errors.New("название элемента не должно превышать 255 символов")
	}
	if i.Status == "" {
		i.Status = "active"
	}
	return nil
}
