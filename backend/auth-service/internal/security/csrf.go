package security

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"strconv"
	"sync"
	"time"
)

type csrfToken struct {
	token     string
	expiresAt time.Time
}

// CSRFManager управляет генерацией, валидацией и очисткой токенов CSRF
type CSRFManager interface {
	GenerateToken(userID *int64) (string, error)
	ValidateToken(userID *int64, sessionToken, incomingToken string) bool
	Invalidate(userID int64)
	StartCleanup(ctx context.Context)
}

type memoryCSRFManager struct {
	mu     sync.RWMutex
	tokens map[string]csrfToken
}

// NewCSRFManager создает новый менеджер CSRF токенов
func NewCSRFManager() CSRFManager {
	return &memoryCSRFManager{
		tokens: make(map[string]csrfToken),
	}
}

func (m *memoryCSRFManager) GenerateToken(userID *int64) (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		fallback := strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.FormatInt(time.Now().Unix(), 36)
		m.saveToken(userID, fallback)
		return fallback, nil
	}

	token := hex.EncodeToString(bytes)
	m.saveToken(userID, token)
	return token, nil
}

func (m *memoryCSRFManager) saveToken(userID *int64, token string) {
	if userID == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()

	key := strconv.FormatInt(*userID, 10)
	m.tokens[key] = csrfToken{
		token:     token,
		expiresAt: time.Now().Add(24 * time.Hour),
	}
}

func (m *memoryCSRFManager) ValidateToken(userID *int64, sessionToken, incomingToken string) bool {
	if incomingToken == "" {
		return false
	}

	if userID == nil {
		return sessionToken != "" && sessionToken == incomingToken
	}

	key := strconv.FormatInt(*userID, 10)
	m.mu.RLock()
	stored, exists := m.tokens[key]
	m.mu.RUnlock()

	if !exists {
		// Fallback к проверке токена из сессии
		return sessionToken != "" && sessionToken == incomingToken
	}

	if time.Now().After(stored.expiresAt) {
		m.Invalidate(*userID)
		return false
	}

	return stored.token == incomingToken
}

func (m *memoryCSRFManager) Invalidate(userID int64) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.tokens, strconv.FormatInt(userID, 10))
}

func (m *memoryCSRFManager) StartCleanup(ctx context.Context) {
	go func() {
		ticker := time.NewTicker(30 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				now := time.Now()
				m.mu.Lock()
				for k, tok := range m.tokens {
					if now.After(tok.expiresAt) {
						delete(m.tokens, k)
					}
				}
				m.mu.Unlock()
			}
		}
	}()
}
