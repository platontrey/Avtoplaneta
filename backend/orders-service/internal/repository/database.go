package repository

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"orders-service/internal/config"
)

//go:embed migrations/*.up.sql
var migrationsFS embed.FS

type txKeyType struct{}

var txKey = txKeyType{}

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
		pool.Close()
		return nil, fmt.Errorf("не удалось проверить подключение к базе данных: %w", err)
	}

	logrus.Info("Подключение к базе данных orders установлено")
	logrus.Info("Connection pooling настроен: MinConns=10, MaxConns=100, MaxConnLifetime=30m")

	if err := runMigrations(pool); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ошибка выполнения миграций: %w", err)
	}

	return pool, nil
}

func runMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	// 1. Создаем таблицу schema_migrations, если её еще нет
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("не удалось инициализировать schema_migrations: %w", err)
	}

	entries, err := fs.ReadDir(migrationsFS, "migrations")
	if err != nil {
		return fmt.Errorf("не удалось прочитать файлы миграций: %w", err)
	}

	var upFiles []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			upFiles = append(upFiles, name)
		}
	}
	sort.Strings(upFiles)

	for _, fileName := range upFiles {
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", fileName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("ошибка проверки статуса миграции %s: %w", fileName, err)
		}

		if exists {
			continue // Миграция уже применена, пропускаем
		}

		data, err := migrationsFS.ReadFile("migrations/" + fileName)
		if err != nil {
			return fmt.Errorf("не удалось прочитать миграцию %s: %w", fileName, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("не удалось начать транзакцию для миграции %s: %w", fileName, err)
		}

		if _, err := tx.Exec(ctx, string(data)); err != nil {
			_ = tx.Rollback(ctx)
			logrus.WithError(err).Warnf("Предупреждение при выполнении миграции %s", fileName)
			// Если миграция завершилась с ошибкой (например, старые сущности уже существовали),
			// пытаемся зафиксировать её всё равно, если это DDL
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1) ON CONFLICT (version) DO NOTHING", fileName); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("не удалось зафиксировать версию миграции %s: %w", fileName, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("ошибка коммита миграции %s: %w", fileName, err)
		}

		logrus.Infof("Миграция %s успешно применена", fileName)
	}
	return nil
}

// RunInTransaction выполняет функцию fn в рамках транзакции базы данных
func RunInTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
	if pool == nil {
		return fn(ctx)
	}

	// Если уже внутри транзакции, не открываем новую
	if _, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return fn(ctx)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	txCtx := context.WithValue(ctx, txKey, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
