package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
)

var store *sessions.CookieStore

// InitAuth инициализирует хранилище сессий
func InitAuth(config *Config) {
	sessionKey := config.SessionSecret

	// Проверка длины ключа сессии
	if len(sessionKey) < 32 {
		log.Fatal("БЕЗОПАСНОСТЬ: SESSION_SECRET должен быть не менее 32 символов")
	}

	store = sessions.NewCookieStore([]byte(sessionKey))

	// Параметры безопасного использования cookies
	isProduction := config.NodeEnv == "production"
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7,            // 7 дней
		HttpOnly: true,                 // Предотвращает XSS атаки
		Secure:   isProduction,         // HTTPS только в продакшене
		SameSite: http.SameSiteLaxMode, // Защита от CSRF
	}
}

// authMiddleware проверяет аутентификацию пользователя
func authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Пропускаем OPTIONS запросы для CORS preflight
		if r.Method == "OPTIONS" {
			next.ServeHTTP(w, r)
			return
		}

		// Проверяем, установлены ли заголовки от gateway (для запросов через gateway)
		userIDStr := r.Header.Get("X-User-ID")
		userEmail := r.Header.Get("X-User-Email")
		userName := r.Header.Get("X-User-Name")

		if userIDStr != "" && userEmail != "" && userName != "" {
			// Запрос пришел через gateway, аутентификация уже проверена
			log.Printf("AUTH: Request authenticated via gateway headers for user ID %s from %s", userIDStr, r.RemoteAddr)
			next.ServeHTTP(w, r)
			return
		}

		// Для прямых запросов проверяем сессию (fallback для тестирования)
		session, err := store.Get(r, "auth-session")
		if err != nil {
			log.Printf("SECURITY: Invalid session from %s: %v", r.RemoteAddr, err)
			http.Error(w, "Invalid session", http.StatusUnauthorized)
			return
		}

		userID, ok := session.Values["user_id"]
		if !ok || userID == nil {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		// Проверка времени сессии
		if loginTime, ok := session.Values["login_time"].(int64); ok {
			if time.Now().Unix()-loginTime > 86400*30 { // 30 days
				log.Printf("SECURITY: Session expired for user ID %v from %s", userID, r.RemoteAddr)
				http.Error(w, "Session expired", http.StatusUnauthorized)
				return
			}
		}

		// For orders service, we just need to ensure user is authenticated
		// The actual user data would be fetched from auth service in production
		r.Header.Set("X-User-ID", strconv.FormatUint(uint64(userID.(uint)), 10))

		next.ServeHTTP(w, r)
	})
}

// adminMiddleware проверяет права администратора
func adminMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr == "" {
			http.Error(w, "Authentication required", http.StatusUnauthorized)
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			http.Error(w, "Invalid user ID", http.StatusBadRequest)
			return
		}

		// In a real implementation, you'd fetch user data from auth service
		// For now, we'll assume admin check is done at auth service level
		// This is a simplified version for the orders service

		// For demo purposes, we'll allow all authenticated users to access admin routes
		// In production, this should validate admin status
		log.Printf("SECURITY: Admin access granted for user ID %d from %s", userID, r.RemoteAddr)

		next.ServeHTTP(w, r)
	})
}