package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"sort"
	"sync"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed db/migrations/*.up.sql
var migrationsFS embed.FS

var dbPool *pgxpool.Pool

// InitDB инициализирует подключение к базе данных и выполняет миграции
func InitDB(config *Config) {
	poolConfig, err := pgxpool.ParseConfig(config.DatabaseURL)
	if err != nil {
		log.Fatal("Не удалось разобрать DatabaseURL:", err)
	}

	poolConfig.MaxConns = 100
	poolConfig.MinConns = 10
	poolConfig.MaxConnLifetime = 30 * time.Minute
	poolConfig.MaxConnIdleTime = 5 * time.Minute

	dbPool, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных:", err)
	}

	if err := dbPool.Ping(context.Background()); err != nil {
		log.Fatal("Не удалось проверить подключение к базе данных:", err)
	}

	log.Println("Подключение к базе данных установлено")
	log.Println("Connection pooling настроен: MinConns=10, MaxConns=100, MaxConnLifetime=30m")

	runMigrations()
}

func runMigrations() {
	entries, err := fs.ReadDir(migrationsFS, "db/migrations")
	if err != nil {
		log.Fatal("Не удалось прочитать файлы миграций:", err)
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
		data, err := migrationsFS.ReadFile("db/migrations/" + fileName)
		if err != nil {
			log.Fatalf("Не удалось прочитать миграцию %s: %v", fileName, err)
		}

		_, err = dbPool.Exec(context.Background(), string(data))
		if err != nil {
			log.Printf("Предупреждение при выполнении миграции %s: %v", fileName, err)
		} else {
			log.Printf("Миграция %s выполнена", fileName)
		}
	}
}

// ReindexAllParts переиндексирует все существующие запчасти в Elasticsearch
func ReindexAllParts() error {
	repo := NewPartRepository(dbPool)
	parts, err := repo.FindAll(context.Background())
	if err != nil {
		return fmt.Errorf("не удалось получить запчасти: %v", err)
	}

	fmt.Printf("Переиндексация %d запчастей...\n", len(parts))

	// Используем пул воркеров для параллельной отправки запросов в Elasticsearch
	numWorkers := 16
	partsChan := make(chan Part, len(parts))
	for _, part := range parts {
		partsChan <- part
	}
	close(partsChan)

	var wg sync.WaitGroup
	wg.Add(numWorkers)

	for i := 0; i < numWorkers; i++ {
		go func() {
			defer wg.Done()
			for part := range partsChan {
				if err := IndexPart(&part); err != nil {
					log.Printf("Предупреждение: Не удалось проиндексировать запчасть %d: %v\n", part.ID, err)
				}
			}
		}()
	}

	wg.Wait()
	fmt.Println("Переиндексация завершена")
	return nil
}
