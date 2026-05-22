package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var store *sessions.CookieStore

func InitAuth(ctx context.Context, config *Config) {
	// Инициализация хранилища сессий с безопасным случайным ключом
	sessionKey := config.SessionSecret

	// Проверка длины ключа сессии
	if len(sessionKey) < 32 {
		log.Fatal("БЕЗОПАСНОСТЬ: SESSION_SECRET должен быть не менее 32 символов")
	}

	store = sessions.NewCookieStore([]byte(sessionKey))

	store.Options = &sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	// Настройка OAuth провайдера Google
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	callbackURL := os.Getenv("GOOGLE_CALLBACK_URL")

	if callbackURL == "" {
		callbackURL = "http://localhost:8082/auth/google/callback"
	}

	if clientID != "" && clientSecret != "" {
		goth.UseProviders(
			google.New(clientID, clientSecret, callbackURL),
		)
		gothic.Store = store
	} else {
		log.Println("ПРЕДУПРЕЖДЕНИЕ: Google OAuth не настроен (отсутствует GOOGLE_CLIENT_ID или GOOGLE_CLIENT_SECRET)")
	}

	// Запуск горутины для очистки просроченных токенов CSRF
	go cleanupExpiredTokens(ctx)
}

// Очистка просроченных токенов CSRF каждые 30 минут
func cleanupExpiredTokens(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Остановка очистки просроченных токенов CSRF")
			return
		case <-ticker.C:
			now := time.Now()
			csrfMutex.Lock()
			for key, token := range csrfTokens {
				if now.After(token.expiresAt) {
					delete(csrfTokens, key)
				}
			}
			csrfMutex.Unlock()

			// Очистка ограничений скорости
			rateLimitMutex.Lock()
			for key, entry := range rateLimits {
				if now.After(entry.resetTime) {
					delete(rateLimits, key)
				}
			}
			rateLimitMutex.Unlock()
		}
	}
}
