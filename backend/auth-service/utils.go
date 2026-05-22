package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"sync"
	"time"
)

// generateSecureCsrfToken генерирует криптографически безопасный токен CSRF
func generateSecureCsrfToken() (string, error) {
	bytes := make([]byte, 32) // 256 бит энтропии
	if _, err := rand.Read(bytes); err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Критическая ошибка генерации CSRF токена: %v", err)
		return "", errors.New("не удалось сгенерировать безопасный токен CSRF")
	}
	return hex.EncodeToString(bytes), nil
}

// rateLimitEntry структура для хранения информации об ограничении скорости
type rateLimitEntry struct {
	attempts   int
	resetTime  time.Time
	blockUntil time.Time
}

var rateLimits = make(map[string]*rateLimitEntry)
var rateLimitMutex sync.RWMutex
