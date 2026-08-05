package main

import (
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

var csrfTokens = make(map[string]csrfToken)
var csrfMutex sync.RWMutex

type csrfToken struct {
	token     string
	expiresAt time.Time
}

var userRepo UserRepository

func SetUserRepo(repo UserRepository) {
	userRepo = repo
}

func CORSMiddleware(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var allowedOrigins []string
		if config.AllowedOrigins != "" {
			allowedOrigins = strings.Split(config.AllowedOrigins, ",")
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
			log.Printf("CORS: Обработка предварительного OPTIONS запроса к %s", c.Request.URL.Path)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func generateCsrfToken() string {
	token, err := generateSecureCsrfToken()
	if err != nil {
		return strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.FormatInt(time.Now().Unix(), 36)
	}
	return token
}

func getUserIDFromSessionValue(v interface{}) (int64, bool) {
	switch id := v.(type) {
	case int64:
		return id, true
	case int:
		return int64(id), true
	case uint:
		return int64(id), true
	case float64:
		return int64(id), true
	}
	return 0, false
}

func authMiddleware(c *gin.Context) {
	log.Printf("ОТЛАДКА: authMiddleware вызван для %s %s от %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())

	session, err := store.Get(c.Request, "auth-session")
	if err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Неверный сеанс от %s: %v", c.ClientIP(), err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сеанс"})
		c.Abort()
		return
	}

	val, ok := session.Values["user_id"]
	log.Printf("ОТЛАДКА: userID из сессии: %v, ok: %v", val, ok)
	if !ok || val == nil {
		log.Printf("ОТЛАДКА: userID отсутствует в сессии")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		c.Abort()
		return
	}

	userID, ok := getUserIDFromSessionValue(val)
	if !ok {
		log.Printf("ОТЛАДКА: userID неверного типа")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		c.Abort()
		return
	}

	if loginTime, ok := session.Values["login_time"].(int64); ok {
		log.Printf("ОТЛАДКА: loginTime из сессии: %d", loginTime)
		if time.Now().Unix()-loginTime > 86400*30 {
			log.Printf("БЕЗОПАСНОСТЬ: Истекший сеанс для пользователя ID %v от %s", userID, c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Сеанс истек, пожалуйста, войдите снова"})
			c.Abort()
			return
		}
	} else {
		log.Printf("ОТЛАДКА: loginTime отсутствует в сессии")
	}

	user, err := userRepo.FindByID(userID)
	if err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Пользователь не найден для сеанса %v от %s", userID, c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сеанс"})
		c.Abort()
		return
	}

	log.Printf("ОТЛАДКА: Пользователь найден: ID=%d, Email=%s, Role=%s", user.ID, user.Email, user.Role)

	c.Set("user", *user)
	c.Set("user_email", user.Email)
	log.Printf("ОТЛАДКА: authMiddleware завершен успешно для пользователя %s", user.Email)
	c.Next()
}

func internalOnlyMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		if clientIP != "127.0.0.1" && clientIP != "::1" && clientIP != "[::1]" {
			log.Printf("БЕЗОПАСНОСТЬ: Попытка доступа к внутреннему endpoint с внешнего IP %s", clientIP)
			c.JSON(http.StatusForbidden, gin.H{"error": "Доступ запрещен"})
			c.Abort()
			return
		}
		c.Next()
	}
}

func csrfMiddleware(c *gin.Context) {
	log.Printf("ОТЛАДКА CSRF: Начало проверки CSRF для %s %s от %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())

	if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
		log.Printf("ОТЛАДКА CSRF: Пропуск CSRF для %s запроса", c.Request.Method)
		c.Next()
		return
	}

	if c.Request.URL.Path == "/auth/login" ||
		c.Request.URL.Path == "/auth/logout" ||
		c.Request.URL.Path == "/auth/refresh" {
		log.Printf("ОТЛАДКА CSRF: Пропуск CSRF для endpoint входа/выхода/обновления: %s", c.Request.URL.Path)
		c.Next()
		return
	}

	if strings.HasPrefix(c.Request.URL.Path, "/internal/") {
		log.Printf("ОТЛАДКА CSRF: Пропуск CSRF для внутреннего endpoint: %s", c.Request.URL.Path)
		c.Next()
		return
	}

	if c.Request.URL.Path == "/admin/user-activity-logs" && (c.ClientIP() == "127.0.0.1" || c.ClientIP() == "::1" || c.ClientIP() == "[::1]") {
		log.Printf("ОТЛАДКА CSRF: Пропуск CSRF для внутреннего логирования активности от %s", c.ClientIP())
		c.Next()
		return
	}

	token := c.GetHeader("X-CSRF-Token")
	if token == "" {
		token = c.PostForm("csrf_token")
	}

	log.Printf("ОТЛАДКА CSRF: Запрос %s %s от %s, токен присутствует: %v, длина токена: %d",
		c.Request.Method, c.Request.URL.Path, c.ClientIP(), token != "", len(token))

	if token == "" {
		log.Printf("БЕЗОПАСНОСТЬ: Отсутствующий токен CSRF для %s %s от %s",
			c.Request.Method, c.Request.URL.Path, c.ClientIP())
		if c.ClientIP() == "127.0.0.1" || c.ClientIP() == "::1" {
			log.Printf("ОТЛАДКА: Внутренний запрос от %s без токена CSRF - возможно, внутренний сервис", c.ClientIP())
		}
		c.JSON(http.StatusForbidden, gin.H{"error": "Требуется токен CSRF"})
		c.Abort()
		return
	}

	session, _ := store.Get(c.Request, "auth-session")
	val, ok := session.Values["user_id"]

	log.Printf("ОТЛАДКА CSRF: user_id сеанса присутствует: %v, userID: %v", ok, val)

	if !ok || val == nil {
		sessionToken, exists := session.Values["csrf_token"]
		log.Printf("ОТЛАДКА CSRF: Неаутентифицированный запрос, токен сеанса существует: %v, совпадает: %v", exists, sessionToken == token)
		if !exists || sessionToken != token {
			log.Printf("БЕЗОПАСНОСТЬ: Неверный токен CSRF для %s %s от %s",
				c.Request.Method, c.Request.URL.Path, c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{"error": "Неверный токен CSRF"})
			c.Abort()
			return
		}
	} else {
		userID, _ := getUserIDFromSessionValue(val)
		userIDStr := strconv.FormatInt(userID, 10)
		log.Printf("ОТЛАДКА CSRF: userIDStr: %s", userIDStr)

		csrfMutex.RLock()
		csrfToken, exists := csrfTokens[userIDStr]
		csrfMutex.RUnlock()

		log.Printf("ОТЛАДКА CSRF: Аутентифицированный запрос для пользователя %s, токен существует: %v, токен совпадает: %v, истек: %v",
			userIDStr, exists, exists && csrfToken.token == token, exists && time.Now().After(csrfToken.expiresAt))

		if !exists || csrfToken.token != token {
			log.Printf("БЕЗОПАСНОСТЬ: Неверный токен CSRF для пользователя %v от %s", userID, c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{"error": "Неверный токен CSRF"})
			c.Abort()
			return
		}

		if time.Now().After(csrfToken.expiresAt) {
			log.Printf("БЕЗОПАСНОСТЬ: Истекший токен CSRF для пользователя %v от %s", userID, c.ClientIP())
			csrfMutex.Lock()
			delete(csrfTokens, userIDStr)
			csrfMutex.Unlock()
			c.JSON(http.StatusForbidden, gin.H{"error": "Токен CSRF истек"})
			c.Abort()
			return
		}
	}

	c.Next()
}

func requireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Printf("ОТЛАДКА: requireRole вызван для роли %s на пути %s", requiredRole, c.Request.URL.Path)
		user, exists := c.Get("user")
		if !exists {
			log.Printf("ОТЛАДКА: Пользователь не найден в контексте")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
			c.Abort()
			return
		}

		userObj := user.(User)
		log.Printf("ОТЛАДКА: Пользователь имеет роль: %s", userObj.Role)
		if userObj.Role != requiredRole && userObj.Role != "admin" {
			log.Printf("БЕЗОПАСНОСТЬ: Недостаточно прав для %s - требуется роль %s, пользователь имеет %s от %s",
				c.Request.URL.Path, requiredRole, userObj.Role, c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав для выполнения этой операции"})
			c.Abort()
			return
		}

		log.Printf("ОТЛАДКА: requireRole прошел успешно для пользователя %s", userObj.Email)
		c.Next()
	}
}

func requireMinRole(minRole string) gin.HandlerFunc {
	roleHierarchy := map[string]int{
		"operator": 1,
		"manager":  2,
		"admin":    3,
	}

	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
			c.Abort()
			return
		}

		userObj := user.(User)
		userLevel := roleHierarchy[userObj.Role]
		requiredLevel := roleHierarchy[minRole]

		if userLevel < requiredLevel {
			log.Printf("БЕЗОПАСНОСТЬ: Недостаточно прав для %s - требуется минимум %s, пользователь имеет %s от %s",
				c.Request.URL.Path, minRole, userObj.Role, c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав для выполнения этой операции"})
			c.Abort()
			return
		}

		c.Next()
	}
}
