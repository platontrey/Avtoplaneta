package repository

import (
	"context"
	"embed"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/security"
	"avtoplaneta/pkg/logger"
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
		slog.WarnContext(ctx, "Не удалось проверить количество пользователей", logger.Err(err))
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
		slog.ErrorContext(ctx, "Не удалось хэшировать пароль для default пользователя", logger.Err(err))
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
		slog.ErrorContext(ctx, "Не удалось создать default пользователя", logger.Err(err))
		return
	}

	slog.InfoContext(ctx, "Default администратор успешно создан (@admin). Пароль установлен из переменной окружения DEFAULT_ADMIN_PASSWORD.",
		slog.String("email", adminEmail),
	)
}
