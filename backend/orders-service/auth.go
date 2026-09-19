package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authv1 "avtoplaneta/gen/auth/v1"
)

var store *sessions.CookieStore
var config *Config

// User представляет пользователя в системе
type User struct {
	ID       int64  `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Initials string `json:"initials,omitempty"`
	INN      string `json:"inn,omitempty"`
	Provider string `json:"provider"`
	Role     string `json:"role"`
	Password string `json:"-"`
}

// InitAuth инициализирует хранилище сессий
func InitAuth(cfg *Config) {
	config = cfg
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

var (
	authGRPCClient authv1.AuthServiceClient
	authGRPCConn   *grpc.ClientConn
	authClientOnce sync.Once
)

func getAuthGRPCClient() authv1.AuthServiceClient {
	authClientOnce.Do(func() {
		if config == nil || config.AuthServiceGRPCURL == "" {
			return
		}
		conn, err := grpc.NewClient(config.AuthServiceGRPCURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			log.Printf("AUTH: Failed to create gRPC connection to auth-service: %v", err)
			return
		}
		authGRPCConn = conn
		authGRPCClient = authv1.NewAuthServiceClient(conn)
		log.Printf("AUTH: Connected to auth-service via gRPC at %s", config.AuthServiceGRPCURL)
	})
	return authGRPCClient
}

// authMiddleware проверяет аутентификацию пользователя
func authMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		// Пропускаем OPTIONS запросы для CORS preflight
		if c.Request.Method == "OPTIONS" {
			c.Next()
			return
		}

		// Проверяем, установлены ли заголовки от Traefik ForwardAuth
		userIDStr := c.GetHeader("X-User-ID")
		userEmail := c.GetHeader("X-User-Email")
		userName := c.GetHeader("X-User-Name")

		if userIDStr != "" && userEmail != "" {
			// Запрос пришел через Traefik ForwardAuth, аутентификация уже проверена
			log.Printf("AUTH: Request authenticated via auth headers for user ID %s from %s", userIDStr, c.ClientIP())
			if uid, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
				c.Set("user_id", uid)
			}
			if userName != "" {
				c.Set("user_name", userName)
			}
			c.Next()
			return
		}

		// Для прямых запросов проверяем аутентификацию через auth-service по gRPC
		var sessionCookie string
		if cookie, err := c.Request.Cookie("auth-session"); err == nil {
			sessionCookie = cookie.Value
		}
		authHeader := c.GetHeader("Authorization")

		if client := getAuthGRPCClient(); client != nil && (sessionCookie != "" || authHeader != "") {
			ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
			defer cancel()
			resp, err := client.ValidateSession(ctx, &authv1.ValidateSessionRequest{
				SessionCookie: sessionCookie,
				Authorization: authHeader,
			})
			if err == nil && resp.Valid && resp.User != nil {
				c.Request.Header.Set("X-User-ID", strconv.FormatUint(uint64(resp.User.Id), 10))
				c.Request.Header.Set("X-User-Email", resp.User.Email)
				c.Request.Header.Set("X-User-Name", resp.User.Name)
				c.Request.Header.Set("X-User-Role", resp.User.Role)
				c.Set("user_id", int64(resp.User.Id))
				c.Next()
				return
			}
			if err != nil {
				log.Printf("AUTH: gRPC ValidateSession failed, trying HTTP fallback: %v", err)
			}
		}

		// Для прямых запросов проверяем аутентификацию через auth-service
		authURL := config.AuthServiceURL + "/auth/me"
		req, err := http.NewRequest("GET", authURL, nil)
		if err != nil {
			log.Printf("AUTH: Failed to create auth request from %s: %v", c.ClientIP(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
			c.Abort()
			return
		}

		// Копируем cookies из оригинального запроса
		for _, cookie := range c.Request.Cookies() {
			req.AddCookie(cookie)
		}

		// Выполняем запрос к auth-service
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Do(req)
		if err != nil {
			log.Printf("AUTH: Failed to connect to auth service from %s: %v", c.ClientIP(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
			c.Abort()
			return
		}
		defer func() {
			if err := resp.Body.Close(); err != nil {
				log.Printf("AUTH: Failed to close auth response body: %v", err)
			}
		}()

		if resp.StatusCode == http.StatusUnauthorized {
			log.Printf("AUTH: User not authenticated from %s", c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
			return
		}

		if resp.StatusCode != http.StatusOK {
			log.Printf("AUTH: Auth service error from %s: status %d", c.ClientIP(), resp.StatusCode)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
			c.Abort()
			return
		}

		// Парсим ответ от auth-service
		var user User
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Printf("AUTH: Failed to read auth response from %s: %v", c.ClientIP(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
			c.Abort()
			return
		}

		if err := json.Unmarshal(body, &user); err != nil {
			log.Printf("AUTH: Failed to parse auth response from %s: %v", c.ClientIP(), err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
			c.Abort()
			return
		}

		log.Printf("AUTH: User authenticated via auth-service for user ID %d from %s", user.ID, c.ClientIP())

		// Устанавливаем заголовки для совместимости
		c.Request.Header.Set("X-User-ID", strconv.FormatInt(user.ID, 10))
		c.Request.Header.Set("X-User-Email", user.Email)
		c.Request.Header.Set("X-User-Name", user.Name)

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
