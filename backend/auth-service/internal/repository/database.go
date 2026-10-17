package repository

import (
	"context"
	"embed"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/security"
	"avtoplaneta/pkg/postgres"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

// InitDB инициализирует пул соединений с базой данных и накатывает миграции через pkg/postgres
func InitDB(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pgCfg := postgres.Config{
		DatabaseURL:     cfg.DatabaseURL,
		MaxConns:        100,
		MinConns:        5,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
	}

	pool, err := postgres.Connect(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных auth: %w", err)
	}

	if err := postgres.RunMigrations(ctx, pool, migrationsFS, "migrations"); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ошибка выполнения миграций auth: %w", err)
	}

	return pool, nil
}

// CreateDefaultUser создает администратора по умолчанию, если таблица пользователей пуста
func CreateDefaultUser(ctx context.Context, repo UserRepository) {
	count, err := repo.CountAll()
	if err != nil {
		logrus.WithError(err).Warn("Не удалось проверить количество пользователей")
		return
	}
	if count > 0 {
		return
	}

	adminEmail := os.Getenv("DEFAULT_ADMIN_EMAIL")
	if adminEmail == "" {
		adminEmail = "bibidatrbib@gmail.com"
	}

	adminPassword := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if adminPassword == "" {
		adminPassword = "qewret123"
	}

	adminName := os.Getenv("DEFAULT_ADMIN_NAME")
	if adminName == "" {
		adminName = "Administrator"
	}

	hashedPassword, err := security.HashPassword(adminPassword)
	if err != nil {
		logrus.WithError(err).Error("Не удалось хэшировать пароль для default пользователя")
		return
	}

	defaultUser := &domain.User{
		Username: "admin",
		Email:    adminEmail,
		Name:     adminName,
		Provider: "local",
		Role:     "admin",
		Password: hashedPassword,
	}

	if _, err := repo.Create(defaultUser); err != nil {
		logrus.WithError(err).Error("Не удалось создать default пользователя")
		return
	}

	logrus.Infof("Default администратор успешно создан (%s / @admin). Пароль установлен из переменной окружения DEFAULT_ADMIN_PASSWORD.", adminEmail)
}
