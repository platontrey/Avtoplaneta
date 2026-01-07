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

// CORSMiddleware добавляет CORS заголовки для кросс-доменных запросов
func CORSMiddleware(config *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var allowedOrigins []string
		if config.AllowedOrigins != "" {
			allowedOrigins = strings.Split(config.AllowedOrigins, ",")
			// Trim spaces
			for i := range allowedOrigins {
				allowedOrigins[i] = strings.TrimSpace(allowedOrigins[i])
			}
		} else {
			allowedOrigins = []string{"http://localhost:5173", "http://192.168.1.63:5173", "http://192.168.56.1:5173", "http://192.168.51.2:5173"}
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

// generateCsrfToken генерирует криптографически безопасный токен CSRF
func generateCsrfToken() string {
	token, err := generateSecureCsrfToken()
	if err != nil {
		// Fallback to less secure method if crypto fails
		return strconv.FormatInt(time.Now().UnixNano(), 36) + strconv.FormatInt(time.Now().Unix(), 36)
	}
	return token
}

// authMiddleware проверяет аутентификацию пользователя
func authMiddleware(c *gin.Context) {
	log.Printf("ОТЛАДКА: authMiddleware вызван для %s %s от %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())

	session, err := store.Get(c.Request, "auth-session")
	if err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Неверный сеанс от %s: %v", c.ClientIP(), err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сеанс"})
		c.Abort()
		return
	}

	userID, ok := session.Values["user_id"]
	log.Printf("ОТЛАДКА: userID из сессии: %v, ok: %v", userID, ok)
	if !ok || userID == nil {
		log.Printf("ОТЛАДКА: userID отсутствует в сессии")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		c.Abort()
		return
	}

	// Проверить возраст сеанса
	if loginTime, ok := session.Values["login_time"].(int64); ok {
		log.Printf("ОТЛАДКА: loginTime из сессии: %d", loginTime)
		if time.Now().Unix()-loginTime > 86400*30 { // 30 дней
			log.Printf("БЕЗОПАСНОСТЬ: Истекший сеанс для пользователя ID %v от %s", userID, c.ClientIP())
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Сеанс истек, пожалуйста, войдите снова"})
			c.Abort()
			return
		}
	} else {
		log.Printf("ОТЛАДКА: loginTime отсутствует в сессии")
	}

	// Проверить, существует ли пользователь
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Пользователь не найден для сеанса %v от %s", userID, c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сеанс"})
		c.Abort()
		return
	}

	log.Printf("ОТЛАДКА: Пользователь найден: ID=%d, Email=%s, Role=%s", user.ID, user.Email, user.Role)

	// Сохранить пользователя в контексте для последующего использования
	c.Set("user", user)
	c.Set("user_email", user.Email) // Для логирования
	log.Printf("ОТЛАДКА: authMiddleware завершен успешно для пользователя %s", user.Email)
	c.Next()
}

// internalOnlyMiddleware проверяет, что запрос приходит только от внутренних IP
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

// csrfMiddleware проверяет токены CSRF для защиты от атак
func csrfMiddleware(c *gin.Context) {
	log.Printf("ОТЛАДКА CSRF: Начало проверки CSRF для %s %s от %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())

	// Пропустить CSRF для GET запросов и предварительных OPTIONS
	if c.Request.Method == "GET" || c.Request.Method == "OPTIONS" {
		log.Printf("ОТЛАДКА CSRF: Пропуск CSRF для %s запроса", c.Request.Method)
		c.Next()
		return
	}

	// Пропустить проверку CSRF для endpoints входа
	if c.Request.URL.Path == "/auth/login" ||
		c.Request.URL.Path == "/auth/logout" {
		log.Printf("ОТЛАДКА CSRF: Пропуск CSRF для endpoint входа: %s", c.Request.URL.Path)
		c.Next()
		return
	}

	// Пропустить проверку CSRF для внутренних endpoints
	if strings.HasPrefix(c.Request.URL.Path, "/internal/") {
		log.Printf("ОТЛАДКА CSRF: Пропуск CSRF для внутреннего endpoint: %s", c.Request.URL.Path)
		c.Next()
		return
	}

	// Пропустить проверку CSRF для внутренних запросов к логированию активности
	if c.Request.URL.Path == "/admin/user-activity-logs" && (c.ClientIP() == "127.0.0.1" || c.ClientIP() == "::1" || c.ClientIP() == "[::1]") {
		log.Printf("ОТЛАДКА CSRF: Пропуск CSRF для внутреннего логирования активности от %s", c.ClientIP())
		c.Next()
		return
	}

	// Проверить токен CSRF
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

	// Получить сеанс для поиска пользователя
	session, _ := store.Get(c.Request, "auth-session")
	userID, ok := session.Values["user_id"]

	log.Printf("ОТЛАДКА CSRF: user_id сеанса присутствует: %v, userID: %v", ok, userID)
	log.Printf("ОТЛАДКА CSRF: Токен CSRF: %s", token)

	if !ok || userID == nil {
		// Для не аутентифицированных запросов, проверить токен на основе сеанса
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
		// Для аутентифицированных запросов, проверить специфичный для пользователя токен
		userIDStr := strconv.FormatUint(uint64(userID.(uint)), 10)
		log.Printf("ОТЛАДКА CSRF: userIDStr: %s", userIDStr)

		csrfMutex.RLock()
		csrfToken, exists := csrfTokens[userIDStr]
		csrfMutex.RUnlock()

		log.Printf("ОТЛАДКА CSRF: Аутентифицированный запрос для пользователя %s, токен существует: %v, токен совпадает: %v, истек: %v",
			userIDStr, exists, exists && csrfToken.token == token, exists && time.Now().After(csrfToken.expiresAt))

		if !exists {
			log.Printf("ОТЛАДКА CSRF: Токен CSRF не существует для пользователя %s", userIDStr)
		} else if csrfToken.token != token {
			log.Printf("ОТЛАДКА CSRF: Токен CSRF не совпадает. Ожидалось: %s, Получено: %s", csrfToken.token, token)
		}

		if !exists || csrfToken.token != token {
			log.Printf("БЕЗОПАСНОСТЬ: Неверный токен CSRF для пользователя %v от %s", userID, c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{"error": "Неверный токен CSRF"})
			c.Abort()
			return
		}

		// Проверить, истек ли токен
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

// requireRole проверяет, что пользователь имеет требуемую роль
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

// requireMinRole проверяет минимальную роль пользователя
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
