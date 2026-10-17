package ginmw

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/authcontext"
)

const (
	defaultAllowedOrigin = "http://localhost:5173"
	defaultAllowedHeaders = "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, X-User-ID, X-User-Email, X-User-Name, X-User-Role, X-User-Username"
	defaultAllowedMethods = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
)

// CORS возвращает middleware для кросс-доменных запросов веб-клиента
func CORS(allowedOrigins ...string) gin.HandlerFunc {
	origin := defaultAllowedOrigin
	if envOrigin := os.Getenv("CORS_ALLOWED_ORIGIN"); envOrigin != "" {
		origin = envOrigin
	}
	if len(allowedOrigins) > 0 && allowedOrigins[0] != "" {
		origin = allowedOrigins[0]
	}

	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", defaultAllowedHeaders)
		c.Header("Access-Control-Allow-Methods", defaultAllowedMethods)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// RequireAuth проверяет аутентификацию пользователя через заголовки Traefik ForwardAuth
func RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodOptions {
			c.Next()
			return
		}

		user, ok := authcontext.FromRequest(c.Request)
		if !ok || !user.IsAuthenticated() {
			logrus.WithField("ip", c.ClientIP()).Warn("Unauthorized request rejected by RequireAuth middleware")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		c.Set("user", user)
		c.Set("user_id", user.ID)
		c.Set("user_name", user.Name)

		c.Next()
	}
}

// RequireRole проверяет, что роль пользователя совпадает с одной из разрешенных
func RequireRole(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := authcontext.FromRequest(c.Request)
		if !ok || !user.IsAuthenticated() {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			return
		}

		if len(allowedRoles) == 0 {
			c.Next()
			return
		}

		hasRole := false
		for _, role := range allowedRoles {
			if strings.EqualFold(user.Role, role) {
				hasRole = true
				break
			}
		}

		if !hasRole {
			logrus.WithFields(logrus.Fields{
				"user_id":       user.ID,
				"user_role":     user.Role,
				"allowed_roles": allowedRoles,
			}).Warn("Forbidden: insufficient role privileges")
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			return
		}

		c.Next()
	}
}

// RequireAdmin проверяет права администратора пользователя
func RequireAdmin() gin.HandlerFunc {
	return RequireRole("admin")
}

// RespondError стандартизирует JSON-ответ с ошибкой
func RespondError(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, gin.H{"error": message})
}

// RespondOK стандартизирует успешный JSON-ответ
func RespondOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, data)
}
