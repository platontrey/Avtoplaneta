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

	"parts-service/internal/config"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

// InitDB инициализирует пул подключений к базе данных и выполняет миграции
func InitDB(cfg *config.Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать DatabaseURL: %w", err)
	}

	poolConfig.MaxConns = 100
	poolConfig.MinConns = 10
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к базе данных: %w", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("не удалось проверить подключение к базе данных: %w", err)
	}

	logrus.Info("Подключение к базе данных установлено")
	logrus.Info("Connection pooling настроен: MinConns=10, MaxConns=100, MaxConnLifetime=30m")

	if err := runMigrations(pool); err != nil {
		return nil, fmt.Errorf("ошибка выполнения миграций: %w", err)
	}

	return pool, nil
}

func runMigrations(pool *pgxpool.Pool) error {
	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("не удалось прочитать файлы миграций: %w", err)
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

		_, err = pool.Exec(context.Background(), string(data))
		if err != nil {
			logrus.WithError(err).Warnf("Предупреждение при выполнении миграции %s", fileName)
		} else {
			logrus.Infof("Миграция %s выполнена", fileName)
		}
	}
	return nil
}
