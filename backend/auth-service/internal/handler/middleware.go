package handler

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"auth-service/internal/domain"
	"auth-service/internal/security"
)

// CORSMiddleware настраивает CORS заголовки
func (h *Handler) CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var allowedOrigins []string
		if h.config != nil && h.config.AllowedOrigins != "" {
			allowedOrigins = strings.Split(h.config.AllowedOrigins, ",")
			for i := range allowedOrigins {
				allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
			}
		} else {
			allowedOrigins = []string{
				"http://localhost:5173",
				"http://192.168.1.63:5173",
				"http://192.168.56.1:5173",
				"http://192.168.51.2:5173",
				"https://backend-server.ru",
				"https://www.backend-server.ru",
				"https://192.168.1.3",
			}
		}

		origin := c.GetHeader("Origin")
		for _, o := range allowedOrigins {
			if o == origin {
				c.Header("Access-Control-Allow-Origin", origin)
				break
			}
		}
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthMiddleware проверяет аутентификацию по Bearer JWT токену или сессионной cookie
func (h *Handler) AuthMiddleware(c *gin.Context) {
	// 1. Проверяем заголовок Authorization (JWT Bearer токен от мобильного приложения / API-клиентов)
	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		var secret string
		if h.config != nil && h.config.JWTSecret != "" {
			secret = h.config.JWTSecret
		}

		claims, err := security.ValidateJWTToken(tokenString, secret)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			c.Abort()
			return
		}

		if h.userRepo == nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Хранилище пользователей не инициализировано"})
			c.Abort()
			return
		}

		user, err := h.userRepo.FindByID(claims.UserID)
		if err != nil || user == nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
			c.Abort()
			return
		}

		c.Set("user", *user)
		c.Set("user_email", user.Email)
		c.Next()
		return
	}

	// 2. Проверяем cookie сессии
	if h.sessionStore == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		c.Abort()
		return
	}

	session, err := h.sessionStore.Get(c.Request, "auth-session")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сеанс"})
		c.Abort()
		return
	}

	val, ok := session.Values["user_id"]
	if !ok || val == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		c.Abort()
		return
	}

	userID, ok := security.GetUserIDFromSessionValue(val)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		c.Abort()
		return
	}

	if loginTime, ok := session.Values["login_time"].(int64); ok {
		if time.Now().Unix()-loginTime > 86400*30 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Сеанс истек, пожалуйста, войдите снова"})
			c.Abort()
			return
		}
	}

	if h.userRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Хранилище пользователей не инициализировано"})
		c.Abort()
		return
	}

	user, err := h.userRepo.FindByID(userID)
	if err != nil || user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сеанс"})
		c.Abort()
		return
	}

	c.Set("user", *user)
	c.Set("user_email", user.Email)
	c.Next()
}

// InternalOnlyMiddleware ограничивает доступ только запросам с loopback IP
func (h *Handler) InternalOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if clientIP != "127.0.0.1" && clientIP != "::1" && clientIP != "[::1]" {
			logrus.WithField("ip", clientIP).Warn("Access denied to internal endpoint")
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CSRFMiddleware проверяет токен CSRF для модифицирующих HTTP методов
func (h *Handler) CSRFMiddleware(c *gin.Context) {
	if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
		c.Next()
		return
	}

	path := c.Request.URL.Path
	if path == "/auth/login" || path == "/auth/logout" || path == "/auth/refresh" || strings.HasPrefix(path, "/auth/google") {
		c.Next()
		return
	}

	// Пропуск CSRF для запросов с Bearer токеном
	if strings.HasPrefix(c.GetHeader("Authorization"), "Bearer ") {
		c.Next()
		return
	}

	if strings.HasPrefix(path, "/internal/") {
		c.Next()
		return
	}

	if path == "/admin/user-activity-logs" && (c.ClientIP() == "127.0.0.1" || c.ClientIP() == "::1" || c.ClientIP() == "[::1]") {
		c.Next()
		return
	}

	token := c.GetHeader("X-CSRF-Token")
	if token == "" {
		token = c.PostForm("csrf_token")
	}

	if token == "" {
		c.JSON(http.StatusForbidden, gin.H{"error": "Требуется токен CSRF"})
		c.Abort()
		return
	}

	var userID *int64
	var sessionToken string

	if h.sessionStore != nil {
		session, _ := h.sessionStore.Get(c.Request, "auth-session")
		if session != nil {
			if val, ok := session.Values["user_id"]; ok && val != nil {
				if uid, ok := security.GetUserIDFromSessionValue(val); ok {
					userID = &uid
				}
			}
			if st, ok := session.Values["csrf_token"].(string); ok {
				sessionToken = st
			}
		}
	}

	if h.csrfManager != nil && !h.csrfManager.ValidateToken(userID, sessionToken, token) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Неверный токен CSRF"})
		c.Abort()
		return
	}

	c.Next()
}

// RequireRole проверяет, что пользователь имеет указанную роль или является admin
func (h *Handler) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		userVal, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
			c.Abort()
			return
		}

		userObj, ok := userVal.(domain.User)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
			c.Abort()
			return
		}

		if userObj.Role != requiredRole && userObj.Role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав для выполнения этой операции"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireMinRole проверяет иерархический минимальный уровень роли (operator < manager < admin)
func (h *Handler) RequireMinRole(minRole string) gin.HandlerFunc {
	roleHierarchy := map[string]int{
		"operator": 1,
		"manager":  2,
		"admin":    3,
	}

	return func(c *gin.Context) {
		userVal, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
			c.Abort()
			return
		}

		userObj, ok := userVal.(domain.User)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
			c.Abort()
			return
		}

		userLevel := roleHierarchy[userObj.Role]
		requiredLevel := roleHierarchy[minRole]

		if userLevel < requiredLevel {
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав для выполнения этой операции"})
			c.Abort()
			return
		}

		c.Next()
	}
}
