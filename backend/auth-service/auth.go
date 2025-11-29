package main

import (
	"log"
	"net/http"
	"sync"
	"time"

	"os"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

var store *sessions.CookieStore

// Хранилище токенов CSRF с ограниченным сроком действия (в рабочей среде используйте Redis или аналог)
type csrfToken struct {
	token     string
	expiresAt time.Time
}

var csrfTokens = make(map[string]csrfToken)
var csrfMutex sync.RWMutex

func initAuth() {
	// Инициализация хранилища сессий с безопасным случайным ключом
	sessionKey := os.Getenv("SESSION_SECRET")
	if sessionKey == "" {
		// Генерация безопасного случайного ключа, если не предоставлен (минимум 32 байта)
		log.Println("ПРЕДУПРЕЖДЕНИЕ: SESSION_SECRET не установлен, используем ключ по умолчанию. УСТАНОВИТЕ ЭТО В ПРОДАКШЕНЕ!")
		sessionKey = "CHANGE_THIS_IN_PRODUCTION_TO_A_SECURE_RANDOM_KEY_32_CHARS_MIN"
	}

	// Проверка длины ключа сессии
	if len(sessionKey) < 32 {
		log.Fatal("БЕЗОПАСНОСТЬ: SESSION_SECRET должен быть не менее 32 символов")
	}

	store = sessions.NewCookieStore([]byte(sessionKey))

	// Параметры безопасного использования cookies
	isProduction := os.Getenv("NODE_ENV") == "production"
	store.Options = &sessions.Options{
		Path:     "/",
		Domain:   "",           // Пустой domain для local development
		MaxAge:   86400 * 7,    // 7 дней
		HttpOnly: true,         // Предотвращает XSS атаки
		Secure:   isProduction, // HTTPS только в продакшене
		SameSite: func() http.SameSite {
			if isProduction {
				return http.SameSiteLaxMode
			}
			return http.SameSiteLaxMode // Для разработки использовать Lax для HTTP
		}(),
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
	go cleanupExpiredTokens()
}

// Очистка просроченных токенов CSRF каждые 30 минут
func cleanupExpiredTokens() {
	ticker := time.NewTicker(30 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
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
