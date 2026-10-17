package repository

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"avtoplaneta/pkg/postgres"
	"template-service/internal/config"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

// InitDB инициализирует пул подключений к PostgreSQL и накатывает миграции через pkg/postgres
func InitDB(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := postgres.Connect(ctx, postgres.DefaultConfig(cfg.DatabaseURL))
	if err != nil {
		return nil, fmt.Errorf("ошибка инициализации PostgreSQL: %w", err)
	}

	if err := postgres.RunMigrations(ctx, pool, migrationsFS, "migrations"); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ошибка запуска миграций: %w", err)
	}

	return pool, nil
}
