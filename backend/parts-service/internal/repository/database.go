package repository

import (
	"context"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"avtoplaneta/pkg/postgres"
	"parts-service/internal/config"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

// InitDB инициализирует пул подключений к базе данных и выполняет миграции через pkg/postgres
func InitDB(cfg *config.Config) (*pgxpool.Pool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pgCfg := postgres.Config{
		DatabaseURL:     cfg.DatabaseURL,
		MaxConns:        100,
		MinConns:        10,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
	}

	pool, err := postgres.Connect(ctx, pgCfg)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	if err := postgres.RunMigrations(ctx, pool, migrationsFS, "migrations"); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ошибка выполнения миграций: %w", err)
	}

	return pool, nil
}
