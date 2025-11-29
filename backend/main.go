package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type User struct {
	ID       uint   `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Initials string `json:"initials,omitempty"`
	INN      string `json:"inn,omitempty"`
	Provider string `json:"provider"`
	Role     string `json:"role"`
	Password string `json:"-"`
}

func main() {
	// Загрузка переменных окружения из .env файла
	err := godotenv.Load()
	if err != nil {
		log.Printf("Предупреждение: не удалось загрузить .env файл: %v", err)
		log.Println("Переменные окружения будут взяты из системных переменных")
	}

	r := gin.Default()

	// Настройка CORS - более ограничительная для продакшена
	allowedOrigins := []string{"http://localhost:5173", "http://localhost:5174", "http://localhost:3000", "http://192.168.1.63:8080", "http://10.0.2.2:8080", "http://192.168.1.63:5173", "http://192.168.56.1:5173", "http://192.168.51.2:5173", "https://spectrologically-seeable-zenobia.ngrok-free.dev"}
	// if origin := os.Getenv("ALLOWED_ORIGIN"); origin != "" {
	// 	allowedOrigins = append(allowedOrigins, origin)
	// }

	log.Println("Allowed origins:", allowedOrigins)

	r.Use(cors.New(cors.Config{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Middleware заголовков безопасности
	r.Use(func(c *gin.Context) {
		// Content Security Policy - строгий для продакшена, разрешить eval для разработки
		csp := "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; connect-src 'self'"
		if os.Getenv("NODE_ENV") != "production" {
			csp += "; script-src 'self' 'unsafe-eval' 'unsafe-inline'; connect-src 'self' http://localhost:8080 http://localhost:5173 http://localhost:5174 http://192.168.1.63:8080 http://192.168.56.1:8080 http://192.168.51.2:8080 ws://localhost:5173 ws://localhost:5174 ws://192.168.1.63:5173 ws://192.168.56.1:5173 ws://192.168.51.2:5173"
		}
		csp += "; frame-ancestors 'none';"

		// Установить заголовок CSP (перезаписать существующие для избежания дубликатов)
		c.Header("Content-Security-Policy", csp)

		// Дополнительные заголовки безопасности
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Добавить HSTS в продакшен
		if os.Getenv("NODE_ENV") == "production" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		// ЛОГИРОВАНИЕ БЕЗОПАСНОСТИ: регистрировать потенциальные события безопасности
		if c.Request.Method != "GET" && c.Request.Method != "OPTIONS" {
			log.Printf("БЕЗОПАСНОСТЬ: %s запрос к %s от %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		}

		c.Next()
	})

	// Прокси маршруты аутентификации к auth-service
	authRoutes := []string{
		"/auth/google",
		"/auth/google/callback",
		"/auth/admin",
		"/auth/login",
		"/auth/logout",
		"/auth/me",
		"/auth/csrf-token",
	}

	for _, route := range authRoutes {
		r.Any(route, func(c *gin.Context) {
			proxyToService(c, "AUTH_SERVICE_URL", "http://localhost:8083", c.Request.Method, c.Request.URL.Path)
		})
	}

	// Прокси маршруты панели администратора к auth-service (только для админов)
	r.Any("/admin/users", authMiddleware, requireRole("admin"), func(c *gin.Context) {
		proxyToService(c, "AUTH_SERVICE_URL", "http://localhost:8083", c.Request.Method, c.Request.URL.Path)
	})
	r.Any("/admin/users/*path", authMiddleware, requireRole("admin"), func(c *gin.Context) {
		proxyToService(c, "AUTH_SERVICE_URL", "http://localhost:8083", c.Request.Method, c.Request.URL.Path)
	})
	r.Any("/admin/status", authMiddleware, requireRole("admin"), func(c *gin.Context) {
		proxyToService(c, "AUTH_SERVICE_URL", "http://localhost:8083", c.Request.Method, c.Request.URL.Path)
	})
	r.Any("/admin/logs", authMiddleware, requireRole("admin"), func(c *gin.Context) {
		proxyToService(c, "AUTH_SERVICE_URL", "http://localhost:8083", c.Request.Method, c.Request.URL.Path)
	})

	// Прокси маршруты инвентаря к parts-service
	// Чтение доступно всем аутентифицированным, изменения - только операторам
	readPartsRoutes := []string{
		"/api/inventory",
		"/api/statistics",
	}

	writePartsRoutes := []string{
		"/api/addpart",
		"/api/deletepart/:id",
		"/api/updatepart/:id",
		"/api/uploadpartphoto/:id",
		"/api/deletepartphoto/:id",
		"/api/markpartfordeletion/:id",
		"/api/admin/delete-zero-quantity-parts/:supplier_code",
		"/api/admin/supplier-codes",
		"/api/admin/bulk-delete-parts",
		"/api/admin/bulk-update-parts",
	}

	for _, route := range readPartsRoutes {
		r.Any(route, func(c *gin.Context) {
			proxyToService(c, "PARTS_SERVICE_URL", "http://localhost:8081", c.Request.Method, c.Request.URL.Path)
		})
	}

	for _, route := range writePartsRoutes {
		r.Any(route, authMiddleware, requireRole("operator"), func(c *gin.Context) {
			proxyToService(c, "PARTS_SERVICE_URL", "http://localhost:8081", c.Request.Method, c.Request.URL.Path)
		})
	}

	// Прокси маршруты заказов к orders-service (только для менеджеров)
	r.Any("/orders", authMiddleware, requireRole("manager"), func(c *gin.Context) {
		proxyToService(c, "ORDERS_SERVICE_URL", "http://localhost:8082", c.Request.Method, c.Request.URL.Path)
	})
	r.Any("/admin/orders", authMiddleware, requireRole("manager"), func(c *gin.Context) {
		proxyToService(c, "ORDERS_SERVICE_URL", "http://localhost:8082", c.Request.Method, c.Request.URL.Path)
	})
	r.Any("/admin/orders/*path", authMiddleware, requireRole("manager"), func(c *gin.Context) {
		proxyToService(c, "ORDERS_SERVICE_URL", "http://localhost:8082", c.Request.Method, c.Request.URL.Path)
	})

	// Статическое обслуживание файлов безопасности и загрузок
	r.StaticFile("/robots.txt", "./robots.txt")
	r.StaticFile("/sitemap.xml", "./sitemap.xml")

	// Прокси статических файлов загрузок к parts-service
	r.Any("/uploads/*filepath", func(c *gin.Context) {
		proxyToService(c, "PARTS_SERVICE_URL", "http://localhost:8081", c.Request.Method, c.Request.URL.Path)
	})

	// API для ИИ агента
	r.POST("/api/ai-agent/chat", handleAIAgentChat)

	log.Println("API Gateway запущен на порту :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal("Не удалось запустить API Gateway:", err)
	}
}

// authMiddleware проверяет аутентификацию пользователя через auth-service
func authMiddleware(c *gin.Context) {
	log.Printf("ОТЛАДКА: authMiddleware вызван для %s %s от %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())

	// Создаем новый HTTP запрос к auth-service для проверки пользователя
	authServiceURL := os.Getenv("AUTH_SERVICE_URL")
	if authServiceURL == "" {
		authServiceURL = "http://localhost:8083"
	}

	req, err := http.NewRequest("GET", authServiceURL+"/auth/me", nil)
	if err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось создать запрос к auth-service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка аутентификации"})
		c.Abort()
		return
	}

	// Копируем все cookies из оригинального запроса
	for _, cookie := range c.Request.Cookies() {
		req.AddCookie(cookie)
	}

	// Копируем заголовки аутентификации
	if authHeader := c.GetHeader("Authorization"); authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	// Выполняем запрос к auth-service
	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось подключиться к auth-service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка аутентификации"})
		c.Abort()
		return
	}
	defer resp.Body.Close()

	// Если auth-service вернул ошибку аутентификации
	if resp.StatusCode == http.StatusUnauthorized {
		log.Printf("БЕЗОПАСНОСТЬ: Пользователь не аутентифицирован от %s", c.ClientIP())
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется аутентификация"})
		c.Abort()
		return
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("БЕЗОПАСНОСТЬ: Ошибка от auth-service: статус %d от %s", resp.StatusCode, c.ClientIP())
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка аутентификации"})
		c.Abort()
		return
	}

	// Парсим ответ от auth-service
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось распарсить ответ от auth-service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка аутентификации"})
		c.Abort()
		return
	}

	log.Printf("ОТЛАДКА: Пользователь аутентифицирован: ID=%d, Email=%s, Role=%s", user.ID, user.Email, user.Role)

	// Сохранить пользователя в контексте для последующего использования
	c.Set("user", user)
	c.Set("user_email", user.Email) // Для логирования
	log.Printf("ОТЛАДКА: authMiddleware завершен успешно для пользователя %s", user.Email)
	c.Next()
}

func requireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
			c.Abort()
			return
		}

		userObj := user.(User)
		if userObj.Role != requiredRole && userObj.Role != "admin" {
			log.Printf("БЕЗОПАСНОСТЬ: Недостаточно прав для %s - требуется роль %s, пользователь имеет %s от %s",
				c.Request.URL.Path, requiredRole, userObj.Role, c.ClientIP())
			c.JSON(http.StatusForbidden, gin.H{"error": "Недостаточно прав для выполнения этой операции"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func proxyToService(c *gin.Context, serviceURLEnv, defaultURL, method, path string) {
	serviceURL := os.Getenv(serviceURLEnv)
	if serviceURL == "" {
		serviceURL = defaultURL
	}

	// Собрать полный URL с query параметрами
	fullURL := serviceURL + path
	if c.Request.URL.RawQuery != "" {
		fullURL += "?" + c.Request.URL.RawQuery
	}

	// Создать новый HTTP запрос
	var body io.Reader
	if method == "POST" || method == "PUT" {
		// Для multipart form data (загрузка файлов) не читать тело
		if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
			// Позволить прокси копировать multipart данные напрямую
			body = c.Request.Body
		} else {
			// Для JSON/text данных прочитать и буферизовать
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось прочитать тело запроса"})
				return
			}
			body = bytes.NewReader(bodyBytes)
			// Восстановить тело для потенциального дальнейшего использования
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать запрос"})
		return
	}

	// Копировать заголовки
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Добавить user ID в заголовок, если пользователь аутентифицирован
	if user, exists := c.Get("user"); exists {
		userObj := user.(User)
		req.Header.Set("X-User-ID", fmt.Sprintf("%d", userObj.ID))
		req.Header.Set("X-User-Email", userObj.Email)
		req.Header.Set("X-User-Name", userObj.Name)
	}

	// Копировать cookies для управления сессиями
	for _, cookie := range c.Request.Cookies() {
		req.AddCookie(cookie)
	}

	// Выполнить запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось подключиться к сервису"})
		return
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			log.Printf("Ошибка закрытия тела ответа: %v", closeErr)
		}
	}()

	// Копировать заголовки ответа (включая Set-Cookie для управления сессиями)
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Установить код статуса
	c.Status(resp.StatusCode)

	// Скопировать response body
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		log.Printf("Ошибка копирования response body: %v", err)
	}
}
