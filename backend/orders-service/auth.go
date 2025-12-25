package main

import (
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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
func authMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Пропускаем OPTIONS запросы для CORS preflight
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Проверяем, установлены ли заголовки от gateway (для запросов через gateway)
		userIDStr := c.GetHeader("X-User-ID")
		userEmail := c.GetHeader("X-User-Email")
		userName := c.GetHeader("X-User-Name")

		if userIDStr != "" && userEmail != "" && userName != "" {
			// Запрос пришел через gateway, аутентификация уже проверена
			log.Printf("AUTH: Request authenticated via gateway headers for user ID %s from %s", userIDStr, c.ClientIP())
			c.Next()
			return
		}

		// Для прямых запросов проверяем сессию (fallback для тестирования)
		session, err := store.Get(c.Request, "auth-session")
		if err != nil {
			log.Printf("SECURITY: Invalid session from %s: %v", c.ClientIP(), err)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid session"})
			c.Abort()
			return
		}

		userID, ok := session.Values["user_id"]
		if !ok || userID == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		// Проверка времени сессии
		if loginTime, ok := session.Values["login_time"].(int64); ok {
			if time.Now().Unix()-loginTime > 86400*30 { // 30 days
				log.Printf("SECURITY: Session expired for user ID %v from %s", userID, c.ClientIP())
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Session expired"})
				c.Abort()
				return
			}
		}

		// For orders service, we just need to ensure user is authenticated
		// The actual user data would be fetched from auth service in production
		c.Request.Header.Set("X-User-ID", strconv.FormatUint(uint64(userID.(uint)), 10))

		c.Next()
	})
}

// adminMiddleware проверяет права администратора
func adminMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		userIDStr := c.GetHeader("X-User-ID")
		if userIDStr == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
			c.Abort()
			return
		}

		// In a real implementation, you'd fetch user data from auth service
		// For now, we'll assume admin check is done at auth service level
		// This is a simplified version for the orders service

		// For demo purposes, we'll allow all authenticated users to access admin routes
		// In production, this should validate admin status
		log.Printf("SECURITY: Admin access granted for user ID %d from %s", userID, c.ClientIP())

		c.Next()
	})
}