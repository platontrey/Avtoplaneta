package postgres

import (
	"context"
	"fmt"
	"io/fs"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
)

// Config настройки пула подключений к PostgreSQL
type Config struct {
	DatabaseURL     string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// DefaultConfig возвращает настройки пула по стандартам Avtoplaneta
func DefaultConfig(databaseURL string) Config {
	return Config{
		DatabaseURL:     databaseURL,
		MaxConns:        50,
		MinConns:        5,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
	}
}

// Connect создает и проверяет пул подключений к PostgreSQL
func Connect(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("postgres parse config: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolConfig.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolConfig.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		poolConfig.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolConfig.MaxConnIdleTime = cfg.MaxConnIdleTime
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("postgres connect: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres ping: %w", err)
	}

	logrus.Info("Пул соединений PostgreSQL успешно инициализирован")
	return pool, nil
}

// RunMigrations применяет миграции из fs.FS с версионированием через schema_migrations
func RunMigrations(ctx context.Context, pool *pgxpool.Pool, fsys fs.FS, dir string) error {
	if pool == nil {
		return fmt.Errorf("postgres pool cannot be nil")
	}

	// 1. Создаем таблицу schema_migrations при ее отсутствии
	_, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			version VARCHAR(255) PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to init schema_migrations: %w", err)
	}

	// 2. Читаем файлы миграций
	entries, err := fs.ReadDir(fsys, dir)
	if err != nil {
		return fmt.Errorf("failed to read migrations dir '%s': %w", dir, err)
	}

	var upFiles []string
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasSuffix(name, ".up.sql") {
			upFiles = append(upFiles, name)
		}
	}
	sort.Strings(upFiles)

	// 3. Последовательно накатываем непримененные миграции
	for _, fileName := range upFiles {
		var exists bool
		err := pool.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE version = $1)", fileName).Scan(&exists)
		if err != nil {
			return fmt.Errorf("check migration status '%s': %w", fileName, err)
		}

		if exists {
			continue // Миграция уже была применена ранее
		}

		filePath := fileName
		if dir != "" && dir != "." {
			filePath = strings.TrimSuffix(dir, "/") + "/" + fileName
		}

		data, err := fs.ReadFile(fsys, filePath)
		if err != nil {
			return fmt.Errorf("read migration file '%s': %w", filePath, err)
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			return fmt.Errorf("begin migration tx '%s': %w", fileName, err)
		}

		if _, err := tx.Exec(ctx, string(data)); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("execute migration '%s': %w", fileName, err)
		}

		if _, err := tx.Exec(ctx, "INSERT INTO schema_migrations (version) VALUES ($1)", fileName); err != nil {
			_ = tx.Rollback(ctx)
			return fmt.Errorf("record migration '%s': %w", fileName, err)
		}

		if err := tx.Commit(ctx); err != nil {
			return fmt.Errorf("commit migration '%s': %w", fileName, err)
		}

		logrus.WithField("version", fileName).Info("Миграция БД успешно применена")
	}

	return nil
}

type txContextKey struct{}

var txKey = txContextKey{}

// WithTx помещает транзакцию pgx.Tx в контекст
func WithTx(ctx context.Context, tx pgx.Tx) context.Context {
	return context.WithValue(ctx, txKey, tx)
}

// TxFromContext извлекает транзакцию pgx.Tx из контекста, если она присутствует
func TxFromContext(ctx context.Context) (pgx.Tx, bool) {
	tx, ok := ctx.Value(txKey).(pgx.Tx)
	return tx, ok
}

// RunInTransaction выполняет функцию fn в рамках транзакции базы данных.
// Если в контексте уже открыта транзакция, fn выполняется в ней без открытия вложенной.
func RunInTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
	if pool == nil {
		return fn(ctx)
	}

	if _, ok := TxFromContext(ctx); ok {
		return fn(ctx)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("postgres begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if err := fn(WithTx(ctx, tx)); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("postgres commit tx: %w", err)
	}

	return nil
}
