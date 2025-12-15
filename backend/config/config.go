package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config представляет основную конфигурацию приложения
type Config struct {
	// Server settings
	Server ServerConfig

	// Database settings
	Database DatabaseConfig

	// Services URLs
	Services ServicesConfig

	// Security settings
	Security SecurityConfig

	// Monitoring settings
	Monitoring MonitoringConfig

	// External services
	External ExternalConfig
}

// ServerConfig конфигурация сервера
type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	Mode         string // "development", "production"
}

// DatabaseConfig конфигурация базы данных
type DatabaseConfig struct {
	Host         string
	Port         string
	User         string
	Password     string
	DBName       string
	SSLMode      string
	MaxOpenConns int
	MaxIdleConns int
	MaxLifetime  time.Duration
}

// ServicesConfig URL микросервисов
type ServicesConfig struct {
	AuthServiceURL   string
	PartsServiceURL  string
	OrdersServiceURL string
}

// SecurityConfig настройки безопасности
type SecurityConfig struct {
	JWTSecret          string
	CSRFSecret         string
	SessionSecret      string
	GoogleClientID     string
	GoogleClientSecret string
	AllowedOrigins     []string
	CORSMaxAge         time.Duration
}

// MonitoringConfig настройки мониторинга
type MonitoringConfig struct {
	PrometheusPort string
	EnableMetrics  bool
	LogLevel       string
}

// ExternalConfig внешние сервисы
type ExternalConfig struct {
	ElasticsearchURL string
	RedisURL         string
}

// Validate проверяет корректность конфигурации
func (c *Config) Validate() error {
	// Проверяем обязательные поля
	if c.Database.Password == "" && c.Server.Mode == "production" {
		return ErrMissingRequiredField("DB_PASSWORD")
	}

	if c.Security.GoogleClientID == "" && c.Server.Mode == "production" {
		return ErrMissingRequiredField("GOOGLE_CLIENT_ID")
	}

	return nil
}

// GetDSN возвращает строку подключения к базе данных
func (c *Config) GetDSN() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		c.Database.Host,
		c.Database.Port,
		c.Database.User,
		c.Database.Password,
		c.Database.DBName,
		c.Database.SSLMode,
	)
}

// IsProduction проверяет, запущено ли приложение в продакшен режиме
func (c *Config) IsProduction() bool {
	return c.Server.Mode == "production"
}

// IsDevelopment проверяет, запущено ли приложение в режиме разработки
func (c *Config) IsDevelopment() bool {
	return c.Server.Mode == "development" || c.Server.Mode == ""
}

// Helper functions

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}

func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if boolValue, err := strconv.ParseBool(value); err == nil {
			return boolValue
		}
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

// ErrMissingRequiredField Errors
var (
	ErrMissingRequiredField = func(field string) error {
		return fmt.Errorf("required configuration field is missing: %s", field)
	}
)
