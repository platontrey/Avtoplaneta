package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func initDB() {
	var err error

	// Строка подключения к PostgreSQL
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		// Локальное подключение к PostgreSQL по умолчанию
		dsn = "host=localhost user=postgres password=qewret123 dbname=autoplanet port=5432 sslmode=disable"
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             time.Millisecond * 200, // Логировать запросы медленнее 200ms
				LogLevel:                  logger.Info,
				IgnoreRecordNotFoundError: true,
				Colorful:                  true,
			},
		),
	})
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	// Автоматическая миграция схемы пользователей и логов активности
	if err := db.AutoMigrate(&User{}, &UserActivityLog{}); err != nil {
		log.Fatal("Не удалось выполнить миграцию базы данных:", err)
	}
}

func createDefaultUser() {
	// Проверить, есть ли уже пользователи
	var count int64
	db.Model(&User{}).Count(&count)
	if count > 0 {
		log.Println("Пользователи уже существуют, пропускаем создание default пользователя")
		return
	}

	// Создать default пользователя
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("Не удалось хэшировать пароль для default пользователя: %v", err)
		return
	}

	defaultUser := User{
		Email:    "bibidatrbib@gmail.com",
		Name:     "Test User",
		Provider: "local",
		Role:     "admin",
		Password: string(hashedPassword),
	}

	if err := db.Create(&defaultUser).Error; err != nil {
		log.Printf("Не удалось создать default пользователя: %v", err)
		return
	}

	log.Println("Default пользователь создан: bibidatrbib@gmail.com / password123")
}

func main() {
	initDB()
	createDefaultUser()
	initAuth()

	log.Println("Сервис аутентификации готов к работе с пользователями.")

	r := gin.Default()

	// Установить доверенные прокси для безопасности
	r.SetTrustedProxies([]string{"127.0.0.1", "localhost"})

	log.Printf("Gin Engine создан. Middleware: Logger=%v, Recovery=%v", r.Use(gin.Logger()), r.Use(gin.Recovery()))

	// CORS middleware для кросс-доменных запросов
	r.Use(func(c *gin.Context) {
		allowedOrigins := []string{"http://localhost:5173", "http://192.168.1.63:5173", "http://192.168.56.1:5173", "http://192.168.51.2:5173"}
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
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Middleware защиты от CSRF атак
	r.Use(csrfMiddleware)

	// Основные маршруты аутентификации
	r.GET("/auth/google", googleAuth)
	r.GET("/auth/google/callback", googleAuthCallback)
	r.POST("/auth/login", userLogin)
	r.POST("/auth/logout", logout)
	r.GET("/auth/me", getCurrentUser)
	r.GET("/auth/csrf-token", getCsrfToken)

	// Маршруты панели администратора
	admin := r.Group("/admin")
	admin.Use(authMiddleware)
	{
		// Управление пользователями - требуется роль admin
		admin.POST("/users", requireRole("admin"), createUser)
		admin.GET("/users", requireMinRole("manager"), getUsers)
		admin.PUT("/users/:id", requireRole("admin"), updateUser)
		admin.DELETE("/users/:id", requireRole("admin"), deleteUser)

		// Управление сервером - требуется роль admin
		admin.GET("/status", requireRole("admin"), getServerStatus)
		admin.GET("/logs", requireRole("admin"), getServerLogs)

		// Логи активности пользователей - требуется роль admin
		admin.GET("/user-activity-logs", requireRole("admin"), getUserActivityLogs)
		admin.POST("/user-activity-logs", requireRole("admin"), logUserActivity)
	}

	log.Println("Сервис аутентификации запускается на порту :8083")
	if err := r.Run(":8083"); err != nil {
		log.Fatal("Не удалось запустить сервис аутентификации:", err)
	}
}
