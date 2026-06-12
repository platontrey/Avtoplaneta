package main

import (
	"context"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

//go:embed db/migrations/*.up.sql
var migrationsFS embed.FS

var dbPool *pgxpool.Pool
var RedisClient *redis.Client
var repo MessagingRepository
var svc MessagingService

func InitDatabase() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres dbname=autoplanet port=5432 sslmode=disable"
	}

	poolConfig, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		log.Fatal("Не удалось разобрать DATABASE_URL:", err)
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
	log.Println("Connection pooling настроен: MinConns=10, MaxConns=100")

	// Выполнение миграций
	runMigrations()

	// Инициализация репозитория и сервиса
	repo = NewMessagingRepository(dbPool)
	svc = NewMessagingService(repo, RedisClient)
}

func runMigrations() {
	entries, err := fs.ReadDir(migrationsFS, "db/migrations")
	if err != nil {
		log.Fatal("Не удалось прочитать файлы миграций:", err)
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

func InitRedis() {
	redisAddr := os.Getenv("REDIS_URL")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
	})

	// Test connection
	_, err := RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	fmt.Println("Redis connected successfully")
	// Обновляем ссылку на RedisClient в сервисе
	if svc != nil {
		svc.(*messagingService).redis = RedisClient
	}
}