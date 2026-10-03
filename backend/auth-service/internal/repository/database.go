package repository

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/security"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

// InitDB инициализирует пул соединений с базой данных и накатывает миграции
func InitDB(cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать DatabaseURL: %w", err)
	}

	poolConfig.MaxConns = 100
	poolConfig.MinConns = 5
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("не удалось проверить подключение к базе данных: %w", err)
	}

	logrus.Info("Database connection pool established")

	if err := runMigrations(pool); err != nil {
		logrus.WithError(err).Warn("Failed to run migrations")
	}

	return pool, nil
}

func runMigrations(pool *pgxpool.Pool) error {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("не удалось прочитать директорию миграций: %w", err)
	}

	var upFiles []string
	for _, entry := range entries {
		name := entry.Name()
		if len(name) > 7 && name[len(name)-7:] == ".up.sql" {
			upFiles = append(upFiles, name)
		}
	}
	sort.Strings(upFiles)

	for _, fileName := range upFiles {
		data, err := migrationsFS.ReadFile("migrations/" + fileName)
		if err != nil {
			return fmt.Errorf("не удалось прочитать миграцию %s: %w", fileName, err)
		}

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, err = pool.Exec(ctx, string(data))
		cancel()

		if err != nil {
			logrus.WithError(err).WithField("file", fileName).Warn("Warning executing migration")
		} else {
			logrus.WithField("file", fileName).Info("Migration applied successfully")
		}
	}

	return nil
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

	hashedPassword, err := security.HashPassword("qewret123")
	if err != nil {
		logrus.WithError(err).Error("Не удалось хэшировать пароль для default пользователя")
		return
	}

	defaultUser := &domain.User{
		Email:    "bibidatrbib@gmail.com",
		Name:     "Test User",
		Provider: "local",
		Role:     "admin",
		Password: hashedPassword,
	}

	if _, err := repo.Create(defaultUser); err != nil {
		logrus.WithError(err).Error("Не удалось создать default пользователя")
		return
	}

	logrus.Info("Default пользователь создан: bibidatrbib@gmail.com / qewret123")
}
