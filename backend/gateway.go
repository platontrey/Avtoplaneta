package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"github.com/swaggo/files"
	"github.com/swaggo/gin-swagger"
)

// User представляет пользователя в системе
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

// Gateway представляет API Gateway
type Gateway struct {
	router             *gin.Engine
	authServiceURL     string
	partsServiceURL    string
	ordersServiceURL   string
	messagingServiceURL string
	allowedOrigins     []string
}

// NewGateway создает новый экземпляр Gateway
func NewGateway() *Gateway {
	// Инициализируем метрики
	InitMetrics()

	g := &Gateway{
		router: gin.Default(),
		allowedOrigins: []string{
			"http://localhost:5173",
			"http://localhost:5174",
			"http://localhost:3000",
			"http://192.168.1.63:8080",
			"http://10.0.2.2:8080",
			"http://192.168.1.63:5173",
			"http://192.168.56.1:5173",
			"http://192.168.51.2:5173",
			"https://spectrologically-seeable-zenobia.ngrok-free.dev",
		},
	}

	g.loadServiceURLs()
	g.setupMiddleware()
	g.setupRoutes()

	return g
}

// loadServiceURLs загружает URL сервисов из переменных окружения
func (g *Gateway) loadServiceURLs() {
	g.authServiceURL = getEnvOrDefault("AUTH_SERVICE_URL", "http://localhost:8083")
	g.partsServiceURL = getEnvOrDefault("PARTS_SERVICE_URL", "http://localhost:8081")
	g.ordersServiceURL = getEnvOrDefault("ORDERS_SERVICE_URL", "http://localhost:8082")
	g.messagingServiceURL = getEnvOrDefault("MESSAGING_SERVICE_URL", "http://localhost:8084")
}

// setupMiddleware настраивает middleware для gateway
func (g *Gateway) setupMiddleware() {
	// CORS middleware
	g.router.Use(cors.New(cors.Config{
		AllowOrigins:     g.allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token", "X-User-ID", "X-User-Email", "X-User-Name"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Security headers middleware
	g.router.Use(g.securityHeadersMiddleware())

	// Logging middleware
	g.router.Use(g.loggingMiddleware())

	// Metrics middleware
	g.router.Use(MetricsMiddleware())
}

// setupRoutes настраивает все маршруты
func (g *Gateway) setupRoutes() {
	// Auth routes (public)
	g.setupAuthRoutes()

	// Admin routes (protected)
	g.setupAdminRoutes()

	// Parts routes
	g.setupPartsRoutes()

	// Orders routes
	g.setupOrdersRoutes()

	// Messaging routes
	g.setupMessagingRoutes()

	// Static files
	g.setupStaticRoutes()

	// Add direct user route for messaging (temporarily without auth for debugging)
	g.router.GET("/api/users", getUsersHandler)

	// AI agent
	g.router.POST("/api/ai-agent/chat", handleAIAgentChat)

	// Swagger documentation
	g.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Monitoring and health checks
	g.router.GET("/health", HealthCheckHandler)
	g.router.GET("/ready", ReadinessHandler)
	g.router.GET("/metrics", MetricsHandler())
}

// setupAuthRoutes настраивает маршруты аутентификации
func (g *Gateway) setupAuthRoutes() {
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
		g.router.Any(route, g.proxyToAuthService)
	}
}

// setupAdminRoutes настраивает административные маршруты
func (g *Gateway) setupAdminRoutes() {
	adminRoutes := []string{
		"/admin/users",
		"/admin/users/*path",
		"/admin/status",
		"/admin/logs",
		"/admin/user-activity-logs",
	}

	for _, route := range adminRoutes {
		g.router.Any(route, authMiddleware, requireRole("admin"), g.proxyToAuthService)
	}
}

// setupPartsRoutes настраивает маршруты для работы с запчастями
func (g *Gateway) setupPartsRoutes() {
	// Public read routes
	readRoutes := []string{
		"/api/inventory",
		"/api/statistics",
	}

	for _, route := range readRoutes {
		g.router.Any(route, g.proxyToPartsService)
	}

	// Protected write routes
	writeRoutes := []string{
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

	for _, route := range writeRoutes {
		g.router.Any(route, authMiddleware, requireRole("operator"), g.proxyToPartsService)
	}

	// Uploads proxy
	g.router.Any("/uploads/*filepath", g.proxyToPartsService)
}

// setupOrdersRoutes настраивает маршруты для работы с заказами
func (g *Gateway) setupOrdersRoutes() {
	orderRoutes := []string{
		"/orders",
		"/admin/orders",
		"/admin/orders/*path",
	}

	for _, route := range orderRoutes {
		g.router.Any(route, authMiddleware, requireRole("manager"), g.proxyToOrdersService)
	}
}

// setupMessagingRoutes настраивает маршруты для работы с сообщениями
func (g *Gateway) setupMessagingRoutes() {
	messagingRoutes := []string{
		"/api/messaging/*path",
	}

	for _, route := range messagingRoutes {
		// Временно убрал authMiddleware для тестирования
		g.router.Any(route, g.proxyToMessagingService)
	}
}

// setupStaticRoutes настраивает статические маршруты
func (g *Gateway) setupStaticRoutes() {
	g.router.StaticFile("/robots.txt", "./robots.txt")
	g.router.StaticFile("/sitemap.xml", "./sitemap.xml")
}

// securityHeadersMiddleware добавляет заголовки безопасности
func (g *Gateway) securityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Content Security Policy
		csp := "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https: http://localhost:8080; font-src 'self' data:; connect-src 'self'"
		if os.Getenv("NODE_ENV") != "production" {
			csp += "; script-src 'self' 'unsafe-eval' 'unsafe-inline'; connect-src 'self' http://localhost:8080 http://localhost:5173 http://localhost:5174 http://192.168.1.63:8080 http://192.168.56.1:8080 http://192.168.51.2:8080 ws://localhost:5173 ws://localhost:5174 ws://192.168.1.63:5173 ws://192.168.56.1:5173 ws://192.168.51.2:5173"
		}
		csp += "; frame-ancestors 'none';"

		c.Header("Content-Security-Policy", csp)
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// HSTS for production
		if os.Getenv("NODE_ENV") == "production" {
			c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}

// loggingMiddleware логирует запросы безопасности
func (g *Gateway) loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method != "GET" && c.Request.Method != "OPTIONS" {
			logrus.WithFields(logrus.Fields{
				"method": c.Request.Method,
				"path":   c.Request.URL.Path,
				"ip":     c.ClientIP(),
			}).Info("Security: Request logged")
		}
		c.Next()
	}
}

// proxyToAuthService проксирует запросы к auth-service
func (g *Gateway) proxyToAuthService(c *gin.Context) {
	g.proxyToService(c, g.authServiceURL, c.Request.Method, c.Request.URL.Path)
}

// proxyToPartsService проксирует запросы к parts-service
func (g *Gateway) proxyToPartsService(c *gin.Context) {
	g.proxyToService(c, g.partsServiceURL, c.Request.Method, c.Request.URL.Path)
}

// proxyToOrdersService проксирует запросы к orders-service
func (g *Gateway) proxyToOrdersService(c *gin.Context) {
	g.proxyToService(c, g.ordersServiceURL, c.Request.Method, c.Request.URL.Path)
}

// proxyToMessagingService проксирует запросы к messaging-service
func (g *Gateway) proxyToMessagingService(c *gin.Context) {
	g.proxyToService(c, g.messagingServiceURL, c.Request.Method, c.Request.URL.Path)
}

// proxyToService универсальный метод проксирования
func (g *Gateway) proxyToService(c *gin.Context, serviceURL, method, path string) {
	// Build full URL with query parameters
	fullURL := serviceURL + path
	if c.Request.URL.RawQuery != "" {
		fullURL += "?" + c.Request.URL.RawQuery
	}

	// Prepare request body
	var body io.Reader
	if method == "POST" || method == "PUT" {
		if strings.Contains(c.GetHeader("Content-Type"), "multipart/form-data") {
			body = c.Request.Body
		} else {
			bodyBytes, err := io.ReadAll(c.Request.Body)
			if err != nil {
				logrus.WithError(err).Error("Failed to read request body")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read request body"})
				return
			}
			body = bytes.NewReader(bodyBytes)
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}
	}

	// Create request
	req, err := http.NewRequest(method, fullURL, body)
	if err != nil {
		logrus.WithError(err).Error("Failed to create proxy request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Add user info if authenticated
	if user, exists := c.Get("user"); exists {
		userObj := user.(User)
		req.Header.Set("X-User-ID", fmt.Sprintf("%d", userObj.ID))
		req.Header.Set("X-User-Email", userObj.Email)
		req.Header.Set("X-User-Name", userObj.Name)
	} else {
		// For testing messaging without auth, set a default user ID
		req.Header.Set("X-User-ID", "1")
		req.Header.Set("X-User-Email", "test@example.com")
		req.Header.Set("X-User-Name", "Test User")
	}

	// Copy cookies
	for _, cookie := range c.Request.Cookies() {
		req.AddCookie(cookie)
	}

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).WithField("service", serviceURL).Error("Failed to connect to service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to service"})
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close response body")
		}
	}()

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Set status and copy body
	c.Status(resp.StatusCode)
	if _, err := io.Copy(c.Writer, resp.Body); err != nil {
		logrus.WithError(err).Error("Failed to copy response body")
	}
}

// Run запускает gateway
func (g *Gateway) Run(port string) error {
	logrus.WithField("port", port).Info("API Gateway starting")
	return g.router.Run(port)
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// authMiddleware проверяет аутентификацию пользователя через auth-service
func authMiddleware(c *gin.Context) {
	logrus.WithFields(logrus.Fields{
		"method": c.Request.Method,
		"path":   c.Request.URL.Path,
		"ip":     c.ClientIP(),
	}).Debug("Auth middleware called")

	authHeader := c.GetHeader("Authorization")
	logrus.WithFields(logrus.Fields{
		"path": c.Request.URL.Path,
		"auth_header_present": authHeader != "",
		"auth_header_length": len(authHeader),
	}).Info("Auth middleware: checking authentication")

	// Создаем новый HTTP запрос к auth-service для проверки пользователя
	authServiceURL := getEnvOrDefault("AUTH_SERVICE_URL", "http://localhost:8083")
	logrus.WithField("auth_service_url", authServiceURL).Info("Auth middleware: auth service URL")

	req, err := http.NewRequest("GET", authServiceURL+"/auth/me", nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to create auth request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		c.Abort()
		return
	}

	// Копируем все cookies из оригинального запроса
	cookiesCount := len(c.Request.Cookies())
	logrus.WithField("cookies_count", cookiesCount).Info("Auth middleware: copying cookies")
	for _, cookie := range c.Request.Cookies() {
		req.AddCookie(cookie)
	}

	// Копируем заголовки аутентификации
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
		logrus.Info("Auth middleware: Authorization header set")
	} else {
		logrus.Warn("Auth middleware: No Authorization header found")
	}

	// Выполняем запрос к auth-service
	client := &http.Client{Timeout: 5 * time.Second}
	logrus.Info("Auth middleware: sending request to auth service")
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to connect to auth service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		c.Abort()
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close auth response body")
		}
	}()

	logrus.WithField("status_code", resp.StatusCode).Info("Auth middleware: auth service response")

	// Если auth-service вернул ошибку аутентификации
	if resp.StatusCode == http.StatusUnauthorized {
		logrus.WithField("ip", c.ClientIP()).Warn("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		c.Abort()
		return
	}

	if resp.StatusCode != http.StatusOK {
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"ip":     c.ClientIP(),
		}).Error("Auth service error")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		c.Abort()
		return
	}

	// Парсим ответ от auth-service
	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		logrus.WithError(err).Error("Failed to parse auth response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		c.Abort()
		return
	}

	logrus.WithFields(logrus.Fields{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}).Debug("User authenticated successfully")

	// Сохранить пользователя в контексте для последующего использования
	c.Set("user", user)
	c.Set("user_email", user.Email)
	c.Next()
}

// requireRole проверяет наличие требуемой роли
func requireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Not authenticated"})
			c.Abort()
			return
		}

		userObj := user.(User)
		if userObj.Role != requiredRole && userObj.Role != "admin" {
			logrus.WithFields(logrus.Fields{
				"path":      c.Request.URL.Path,
				"required":  requiredRole,
				"user_role": userObj.Role,
				"ip":        c.ClientIP(),
			}).Warn("Insufficient permissions")
			c.JSON(http.StatusForbidden, gin.H{"error": "Insufficient permissions"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// handleAIAgentChat обрабатывает запросы к ИИ агенту
func handleAIAgentChat(c *gin.Context) {
	var req struct {
		Message string                 `json:"message"`
		Context map[string]interface{} `json:"context,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		logrus.WithError(err).Error("Failed to parse AI agent request")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	if req.Message == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Message is required"})
		return
	}

	// Вызываем обработку ИИ команды
	response, action := processAICommand(req.Message, req.Context)

	// Возвращаем ответ
	c.JSON(http.StatusOK, gin.H{
		"response": response,
		"action":   action,
	})
}

// getUsersHandler возвращает список пользователей для messaging
func getUsersHandler(c *gin.Context) {
	// Get users from auth-service
	authServiceURL := getEnvOrDefault("AUTH_SERVICE_URL", "http://localhost:8083")

	req, err := http.NewRequest("GET", authServiceURL+"/internal/users", nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to create auth service request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to connect to auth service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to auth service"})
		return
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close auth service response body")
		}
	}()

	if resp.StatusCode != http.StatusOK {
		logrus.WithField("status", resp.StatusCode).Error("Auth service returned error")
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to fetch users"})
		return
	}

	var response struct {
		Users []User `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logrus.WithError(err).Error("Failed to parse auth service response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	logrus.WithField("users_count", len(response.Users)).Info("Fetched users from auth service")
	c.JSON(http.StatusOK, gin.H{"users": response.Users})
}
