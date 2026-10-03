package security

import (
	"sync"
	"time"
)

// RateLimiter интерфейс ограничения частоты запросов
type RateLimiter interface {
	IsLimited(key string, limit int, window time.Duration) (bool, error)
}

type rateLimitEntry struct {
	attempts  int
	resetTime time.Time
}

type memoryRateLimiter struct {
	mu      sync.Mutex
	entries map[string]*rateLimitEntry
}

// NewMemoryRateLimiter создает in-memory ограничитель частоты запросов
func NewMemoryRateLimiter() RateLimiter {
	limiter := &memoryRateLimiter{
		entries: make(map[string]*rateLimitEntry),
	}
	return limiter
}

func (l *memoryRateLimiter) IsLimited(key string, limit int, window time.Duration) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	entry, exists := l.entries[key]
	if !exists || now.After(entry.resetTime) {
		l.entries[key] = &rateLimitEntry{
			attempts:  1,
			resetTime: now.Add(window),
		}
		return false, nil
	}

	entry.attempts++
	if entry.attempts > limit {
		return true, nil
	}

	return false, nil
}
