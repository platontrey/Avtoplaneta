package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/authcontext"
)

// CORSMiddleware добавляет CORS заголовки для кросс-доменных запросов
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-ID, X-User-Email, X-User-Name, X-User-Role")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// AuthMiddleware проверяет аутентификацию пользователя через Traefik ForwardAuth заголовки
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		user, ok := authcontext.FromRequest(c.Request)
		if !ok || !user.IsAuthenticated() {
			logrus.WithField("ip", c.ClientIP()).Warn("Unauthorized request to orders-service")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_name", user.Name)

		c.Next()
	}
}

// AdminMiddleware проверяет права администратора пользователя
func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authcontext.FromRequest(c.Request)
		if !ok || !user.IsAuthenticated() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		// Если роль передана в заголовках, проверяем права администратора
		if user.Role != "" && !user.IsAdmin() {
			logrus.WithFields(logrus.Fields{
				"user_id": user.ID,
				"role":    user.Role,
			}).Warn("Forbidden: Admin access required")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Admin access required"})
			return
		}

		c.Next()
	}
}
