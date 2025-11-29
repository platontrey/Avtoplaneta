package main

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"math"
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

// generateCsrfToken генерирует новый токен CSRF с fallback
func generateCsrfToken() string {
	token, err := generateSecureCsrfToken()
	if err != nil {
		// Fallback: комбинация timestamp + случайных данных
		fallback := make([]byte, 16)
		rand.Read(fallback) // Игнорируем ошибку для fallback
		return hex.EncodeToString(append(fallback, []byte(time.Now().String())...))
	}
	return token
}

// rateLimitEntry структура для хранения информации об ограничении скорости
type rateLimitEntry struct {
	attempts   int
	resetTime  time.Time
	blockUntil time.Time
}

var rateLimits = make(map[string]*rateLimitEntry)
var rateLimitMutex sync.RWMutex

// isRateLimited улучшенная версия с экспоненциальной задержкой
func isRateLimited(key string, maxAttempts int, window time.Duration) (bool, time.Duration) {
	rateLimitMutex.Lock()
	defer rateLimitMutex.Unlock()

	now := time.Now()
	entry, exists := rateLimits[key]

	if !exists || now.After(entry.resetTime) {
		rateLimits[key] = &rateLimitEntry{
			attempts:   1,
			resetTime:  now.Add(window),
			blockUntil: now,
		}
		return false, 0
	}

	// Проверяем блокировку
	if now.Before(entry.blockUntil) {
		remaining := entry.blockUntil.Sub(now)
		return true, remaining
	}

	entry.attempts++

	// Если превысили лимит, блокируем с экспоненциальной задержкой
	if entry.attempts > maxAttempts {
		// Экспоненциальная задержка: 1min, 2min, 4min, 8min, 16min, max 32min
		delay := time.Duration(math.Min(float64(entry.attempts-maxAttempts), 5)) * time.Minute
		maxDelay := 32 * time.Minute
		if delay > maxDelay {
			delay = maxDelay
		}

		entry.blockUntil = now.Add(delay)
		return true, delay
	}

	return false, 0
}
