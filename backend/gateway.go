package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/eapache/go-resiliency/breaker"
	"github.com/eapache/go-resiliency/retrier"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"golang.org/x/time/rate"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"

	authv1 "avtoplaneta/gen/auth/v1"
	partsv1 "avtoplaneta/gen/parts/v1"
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
	router              *gin.Engine
	authServiceURL      string
	partsServiceURL     string
	ordersServiceURL    string
	messagingServiceURL string
	allowedOrigins      []string

	// gRPC clients (замена HTTP proxy)
	authConn *grpc.ClientConn
	authGRPC authv1.AuthServiceClient
	
	partsConn *grpc.ClientConn
	partsGRPC partsv1.PartsServiceClient

	// Resiliency patterns
	authBreaker      *breaker.Breaker
	partsBreaker     *breaker.Breaker
	ordersBreaker    *breaker.Breaker
	messagingBreaker *breaker.Breaker

	authLimiter      *rate.Limiter
	partsLimiter     *rate.Limiter
	ordersLimiter    *rate.Limiter
	messagingLimiter *rate.Limiter

	authSem      chan struct{} // bulkhead for auth service
	partsSem     chan struct{} // bulkhead for parts service
	ordersSem    chan struct{} // bulkhead for orders service
	messagingSem chan struct{} // bulkhead for messaging service

	authCache *AuthCache
}

// NewGateway создает новый экземпляр Gateway
func NewGateway() *Gateway {
	// Инициализируем метрики
	InitMetrics()

	domain := getEnvOrDefault("DOMAIN", "localhost")

	g := &Gateway{
		router: gin.Default(),
		authCache: NewAuthCache(5 * time.Minute), // Кэшируем авторизацию на 5 минут
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
			"https://192.168.1.3",
			"https://" + domain,
			"http://" + domain,
		},
	}

	g.loadServiceURLs()
	g.initResiliencyPatterns()
	g.initGRPCClients()
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

// initResiliencyPatterns инициализирует паттерны устойчивости
func (g *Gateway) initResiliencyPatterns() {
	// Circuit Breakers: 5 ошибок подряд вызывают открытие на 10 секунд
	g.authBreaker = breaker.New(5, 1, 10*time.Second)
	g.partsBreaker = breaker.New(5, 1, 10*time.Second)
	g.ordersBreaker = breaker.New(5, 1, 10*time.Second)
	g.messagingBreaker = breaker.New(5, 1, 10*time.Second)

	// Rate Limiters: 100 запросов в секунду на сервис
	g.authLimiter = rate.NewLimiter(rate.Limit(100), 100)
	g.partsLimiter = rate.NewLimiter(rate.Limit(100), 100)
	g.ordersLimiter = rate.NewLimiter(rate.Limit(100), 100)
	g.messagingLimiter = rate.NewLimiter(rate.Limit(100), 100)

	// Bulkhead: максимум 50 одновременных соединений на сервис
	g.authSem = make(chan struct{}, 50)
	g.partsSem = make(chan struct{}, 50)
	g.ordersSem = make(chan struct{}, 50)
	g.messagingSem = make(chan struct{}, 50)
}

// initGRPCClients инициализирует gRPC-соединения к сервисам
func (g *Gateway) initGRPCClients() {
	authGRPCAddr := getEnvOrDefault("AUTH_GRPC_ADDR", "localhost:9083")
	partsGRPCAddr := getEnvOrDefault("PARTS_GRPC_ADDR", "localhost:9081")

	var err error
	g.authConn, err = grpc.NewClient(authGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to create gRPC connection to auth-service")
	} else {
		g.authGRPC = authv1.NewAuthServiceClient(g.authConn)
		logrus.WithField("addr", authGRPCAddr).Info("gRPC client connected to auth-service")
	}

	g.partsConn, err = grpc.NewClient(partsGRPCAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
	)
	if err != nil {
		logrus.WithError(err).Error("Failed to create gRPC connection to parts-service")
	} else {
		g.partsGRPC = partsv1.NewPartsServiceClient(g.partsConn)
		logrus.WithField("addr", partsGRPCAddr).Info("gRPC client connected to parts-service")
	}
}

// Close закрывает все gRPC соединения
func (g *Gateway) Close() {
	if g.authConn != nil {
		if err := g.authConn.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close auth gRPC connection")
		}
	}
	if g.partsConn != nil {
		if err := g.partsConn.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close parts gRPC connection")
		}
	}
}

// setupMiddleware настраивает middleware для gateway
func (g *Gateway) setupMiddleware() {
	if trustedProxies := os.Getenv("FORWARDED_ALLOW_IPS"); trustedProxies != "" {
		if trustedProxies == "*" {
			g.router.ForwardedByClientIP = true
		} else {
			g.router.SetTrustedProxies(strings.Split(trustedProxies, ","))
		}
	}

	g.router.Use(cors.New(cors.Config{
		AllowOrigins:     g.allowedOrigins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-CSRF-Token", "X-User-ID", "X-User-Email", "X-User-Name", "X-Forwarded-For", "X-Forwarded-Proto", "X-Forwarded-Host"},
		ExposeHeaders:    []string{"Content-Length", "X-Request-Id"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	g.router.Use(g.securityHeadersMiddleware())
	g.router.Use(g.loggingMiddleware())
	g.router.Use(g.rateLimitingMiddleware())
	g.router.Use(MetricsMiddleware())
	g.router.Use(otelgin.Middleware("api-gateway"))
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

	// Список пользователей нужен клиентам мессенджера и не должен быть публичным.
	g.router.GET("/api/users", g.authMiddleware, requireRole("operator"), g.getUsersHandler)

	// AI agent доступен только авторизованным пользователям приложения.
	g.router.POST("/api/ai-agent/chat", g.authMiddleware, requireRole("operator"), handleAIAgentChat)

	// Swagger documentation
	g.router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Monitoring and health checks
	g.router.GET("/health", HealthCheckHandler)
	g.router.GET("/ready", ReadinessHandler)
	g.router.GET("/metrics", MetricsHandler())

	// Профилирование (только для админов)
	adminDebugGroup := g.router.Group("/admin/debug", g.authMiddleware, requireRole("admin"))
	pprof.RouteRegister(adminDebugGroup, "pprof")

	// Настройка grpc-gateway
	g.setupGRPCGatewayRoutes()
}

// setupGRPCGatewayRoutes настраивает маршрутизацию для grpc-gateway
func (g *Gateway) setupGRPCGatewayRoutes() {
	if g.authConn == nil {
		logrus.Warn("Cannot setup grpc-gateway: auth gRPC connection is nil")
		return
	}

	// Создаем мультиплексор grpc-gateway
	gwmux := runtime.NewServeMux(
		// Маппинг заголовков HTTP -> gRPC metadata (например для Cookie и Authorization)
		runtime.WithIncomingHeaderMatcher(func(key string) (string, bool) {
			if strings.ToLower(key) == "cookie" || strings.ToLower(key) == "authorization" {
				return key, true
			}
			return runtime.DefaultHeaderMatcher(key)
		}),
	)

	// Регистрируем auth-service хендлеры в grpc-gateway
	err := authv1.RegisterAuthServiceHandler(context.Background(), gwmux, g.authConn)
	if err != nil {
		logrus.WithError(err).Fatal("Failed to register auth service handler in grpc-gateway")
	}

	// Регистрируем parts-service хендлеры в grpc-gateway
	if g.partsConn != nil {
		err = partsv1.RegisterPartsServiceHandler(context.Background(), gwmux, g.partsConn)
		if err != nil {
			logrus.WithError(err).Fatal("Failed to register parts service handler in grpc-gateway")
		}
	} else {
		logrus.Warn("Cannot register parts service in grpc-gateway: partsConn is nil")
	}

	// Монтируем grpc-gateway внутри Gin-движка по пути /api/v1/*
	// Мы оборачиваем его в authMiddleware для защищенных маршрутов, или оставляем публичным,
	// но grpc-gateway сам проксирует заголовки, поэтому auth-service сможет валидировать их внутри.
	g.router.Any("/api/v1/*any", gin.WrapH(gwmux))
}

// setupAuthRoutes настраивает маршруты аутентификации
func (g *Gateway) setupAuthRoutes() {
	authRoutes := []string{
		"/auth/google",
		"/auth/google/callback",
		"/auth/google/mobile",
		"/auth/admin",
		"/auth/login",
		"/auth/logout",
		"/auth/me",
		"/auth/csrf-token",
		"/auth/refresh",
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
		g.router.Any(route, g.authMiddleware, requireRole("admin"), g.proxyToAuthService)
	}
}

// setupPartsRoutes настраивает маршруты для работы с запчастями
func (g *Gateway) setupPartsRoutes() {
	// Public read routes
	readRoutes := []string{
		"/api/inventory",
		"/api/inventory/:id",
		"/api/statistics",
		"/api/part-catalog",
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
		"/api/defect-reports",
		"/api/admin/delete-zero-quantity-parts/:supplier_code",
		"/api/admin/supplier-codes",
		"/api/admin/bulk-delete-parts",
		"/api/admin/bulk-update-parts",
	}

	for _, route := range writeRoutes {
		g.router.Any(route, g.authMiddleware, requireRole("operator"), g.proxyToPartsService)
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
		g.router.Any(route, g.authMiddleware, requireRole("operator"), g.proxyToOrdersService)
	}
}

// setupMessagingRoutes настраивает маршруты для работы с сообщениями
func (g *Gateway) setupMessagingRoutes() {
	messagingRoutes := []string{
		"/api/messaging/*path",
	}

	for _, route := range messagingRoutes {
		g.router.Any(route, g.authMiddleware, requireRole("operator"), g.proxyToMessagingService)
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

// rateLimitingMiddleware ограничивает частоту запросов
func (g *Gateway) rateLimitingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		var limiter *rate.Limiter

		// Определяем лимитер на основе пути запроса
		path := c.Request.URL.Path
		if strings.HasPrefix(path, "/auth/") {
			limiter = g.authLimiter
		} else if strings.HasPrefix(path, "/api/") && (strings.Contains(path, "/inventory") || strings.Contains(path, "/addpart") || strings.Contains(path, "/deletepart") || strings.Contains(path, "/updatepart")) {
			limiter = g.partsLimiter
		} else if strings.HasPrefix(path, "/orders/") || strings.HasPrefix(path, "/admin/orders") {
			limiter = g.ordersLimiter
		} else if strings.HasPrefix(path, "/api/messaging/") {
			limiter = g.messagingLimiter
		} else {
			// Для остальных запросов используем общий лимитер
			limiter = rate.NewLimiter(rate.Limit(1000), 1000)
		}

		if !limiter.Allow() {
			logrus.WithFields(logrus.Fields{
				"path": path,
				"ip":   c.ClientIP(),
			}).Warn("Rate limit exceeded")
			c.JSON(http.StatusTooManyRequests, gin.H{"error": "Rate limit exceeded"})
			c.Abort()
			return
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

// proxyToService универсальный метод проксирования с паттернами устойчивости
func (g *Gateway) proxyToService(c *gin.Context, serviceURL, method, path string) {
	// Определяем компоненты resiliency на основе serviceURL
	var breaker *breaker.Breaker
	var sem chan struct{}

	switch serviceURL {
	case g.authServiceURL:
		breaker = g.authBreaker
		sem = g.authSem
	case g.partsServiceURL:
		breaker = g.partsBreaker
		sem = g.partsSem
	case g.ordersServiceURL:
		breaker = g.ordersBreaker
		sem = g.ordersSem
	case g.messagingServiceURL:
		breaker = g.messagingBreaker
		sem = g.messagingSem
	default:
		logrus.WithField("serviceURL", serviceURL).Error("Unknown service URL")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unknown service"})
		return
	}

	// Bulkhead: ограничиваем количество одновременных соединений
	select {
	case sem <- struct{}{}:
		defer func() { <-sem }()
	default:
		logrus.WithField("service", serviceURL).Warn("Bulkhead: too many concurrent requests")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service temporarily unavailable"})
		return
	}

	// Retry: создаем retrier с экспоненциальной задержкой
	r := retrier.New(retrier.ConstantBackoff(3, 100*time.Millisecond), nil)

	// Circuit Breaker + Retry: выполняем запрос через breaker с retry
	err := breaker.Run(func() error {
		return r.Run(func() error {
			return g.executeRequest(c, serviceURL, method, path)
		})
	})

	if err != nil {
		logrus.WithError(err).WithField("service", serviceURL).Error("Failed to execute request after retries and circuit breaker")
		// Если breaker.Run вернул ошибку, значит circuit breaker открыт или все попытки failed
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Service temporarily unavailable"})
	}
}

// executeRequest выполняет один запрос с возможностью повторных попыток
func (g *Gateway) executeRequest(c *gin.Context, serviceURL, method, path string) error {
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
				return err
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
		return err
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
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).WithField("service", serviceURL).Error("Failed to connect to service")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to service"})
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.WithError(err).Warn("Failed to close response body")
		}
	}()

	// Для circuit breaker: считаем неудачей только сетевые ошибки или 5xx статусы
	if resp.StatusCode >= 500 {
		return fmt.Errorf("server error: %d", resp.StatusCode)
	}

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
		return err
	}

	return nil
}

// Run запускает gateway с поддержкой graceful shutdown
func (g *Gateway) Run(ctx context.Context, port string) error {
	logrus.WithField("port", port).Info("API Gateway starting")

	srv := &http.Server{
		Addr:    port,
		Handler: g.router,
	}

	// Канал для ошибок сервера
	errChan := make(chan error, 1)

	go func() {
		if os.Getenv("NODE_ENV") == "production" {
			logrus.Info("Production mode: TLS handled by Traefik/Reverse Proxy")
		} else {
			logrus.Info("Development mode: starting Gateway without TLS...")
		}
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errChan <- err
		}
	}()

	// Ожидание сигнала отмены или ошибки
	select {
	case <-ctx.Done():
		logrus.Info("Shutting down API Gateway gracefully...")
		defer g.Close()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := srv.Shutdown(shutdownCtx); err != nil {
			logrus.WithError(err).Error("Server forced to shutdown")
			return err
		}
		logrus.Info("API Gateway stopped")
		return nil
	case err := <-errChan:
		return err
	}
}

// getEnvOrDefault возвращает значение переменной окружения или значение по умолчанию
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// authMiddleware проверяет аутентификацию пользователя через gRPC (auth-service)
func (g *Gateway) authMiddleware(c *gin.Context) {
	if g.authGRPC == nil {
		logrus.Warn("Auth middleware: gRPC client not initialized, using HTTP fallback")
		g.authMiddlewareHTTP(c)
		return
	}

	authHeader := c.GetHeader("Authorization")

	// Собираем session cookie для сессионной авторизации
	var sessionCookie string
	for _, cookie := range c.Request.Cookies() {
		if cookie.Name == "auth-session" {
			sessionCookie = cookie.Value
		}
	}
	cookieHeader := c.GetHeader("Cookie")
	
	// Create a unique cache key based on token or session
	cacheKey := authHeader
	if cacheKey == "" {
		cacheKey = sessionCookie
	}
	
	// Check cache first if we have a key
	if cacheKey != "" {
		if cachedUser, ok := g.authCache.Get(cacheKey); ok {
			logrus.WithFields(logrus.Fields{
				"id":    cachedUser.ID,
				"email": cachedUser.Email,
				"role":  cachedUser.Role,
			}).Debug("User authenticated successfully via CACHE")
			
			c.Set("user", cachedUser)
			c.Set("user_email", cachedUser.Email)
			c.Next()
			return
		}
	}

	resp, err := g.authGRPC.ValidateSession(c.Request.Context(), &authv1.ValidateSessionRequest{
		SessionCookie: cookieHeader,
		Authorization: authHeader,
	})

	if err != nil {
		logrus.WithError(err).Warn("Auth middleware: gRPC ValidateSession failed")
		_ = sessionCookie // в gRPC передаётся полный Cookie header
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		c.Abort()
		return
	}

	if !resp.Valid {
		logrus.WithField("ip", c.ClientIP()).Warn("User not authenticated")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		c.Abort()
		return
	}

	user := User{
		ID:    uint(resp.User.Id),
		Email: resp.User.Email,
		Name:  resp.User.Name,
		Role:  resp.User.Role,
	}
	
	// Save to cache
	if cacheKey != "" {
		g.authCache.Set(cacheKey, user)
	}

	logrus.WithFields(logrus.Fields{
		"id":    user.ID,
		"email": user.Email,
		"role":  user.Role,
	}).Debug("User authenticated successfully via gRPC")

	c.Set("user", user)
	c.Set("user_email", user.Email)
	c.Next()
}

// authMiddlewareHTTP — fallback HTTP-версия на случай, если gRPC не работает
func (g *Gateway) authMiddlewareHTTP(c *gin.Context) {
	authHeader := c.GetHeader("Authorization")

	authServiceURL := getEnvOrDefault("AUTH_SERVICE_URL", "http://localhost:8083")

	req, err := http.NewRequest("GET", authServiceURL+"/auth/me", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		c.Abort()
		return
	}

	for _, cookie := range c.Request.Cookies() {
		req.AddCookie(cookie)
	}

	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		c.Abort()
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		c.JSON(resp.StatusCode, gin.H{"error": "Authentication required"})
		c.Abort()
		return
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Authentication error"})
		c.Abort()
		return
	}

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

// getUsersHandler возвращает список пользователей для messaging (через gRPC)
func (g *Gateway) getUsersHandler(c *gin.Context) {
	if g.authGRPC != nil {
		resp, err := g.authGRPC.GetUsers(c.Request.Context(), &authv1.GetUsersRequest{})
		if err == nil {
			users := make([]User, len(resp.Users))
			for i, u := range resp.Users {
				users[i] = User{
					ID:       uint(u.Id),
					Email:    u.Email,
					Name:     u.Name,
					Role:     u.Role,
					Initials: u.Initials,
					INN:      u.Inn,
					Provider: u.Provider,
				}
			}
			logrus.WithField("users_count", len(users)).Info("Fetched users from auth service via gRPC")
			c.JSON(http.StatusOK, gin.H{"users": users})
			return
		}
		logrus.WithError(err).Warn("gRPC GetUsers failed, falling back to HTTP")
	}

	// HTTP fallback
	authServiceURL := getEnvOrDefault("AUTH_SERVICE_URL", "http://localhost:8083")
	req, err := http.NewRequest("GET", authServiceURL+"/internal/users", nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to auth service"})
		return
	}
	defer resp.Body.Close()

	var response struct {
		Users []User `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": response.Users})
}
