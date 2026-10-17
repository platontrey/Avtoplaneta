package main

import (
	"context"
	"embed"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/postgres"
)

//go:embed db/migrations/*.up.sql
var migrationsFS embed.FS

var dbPool *pgxpool.Pool
var RedisClient *redis.Client
var repo MessagingRepository
var svc MessagingService

func InitDatabase() {
	dsn := os.Getenv("MESSAGING_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "host=localhost user=postgres dbname=autoplanet port=5432 sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	pgCfg := postgres.Config{
		DatabaseURL:     dsn,
		MaxConns:        100,
		MinConns:        10,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
	}

	var err error
	dbPool, err = postgres.Connect(ctx, pgCfg)
	if err != nil {
		log.Fatal("Не удалось подключиться к базе данных messaging:", err)
	}

	// Выполнение миграций с версионированием через schema_migrations
	if err := postgres.RunMigrations(ctx, dbPool, migrationsFS, "db/migrations"); err != nil {
		dbPool.Close()
		log.Fatal("Ошибка выполнения миграций messaging:", err)
	}

	// Инициализация репозитория и сервиса
	repo = NewMessagingRepository(dbPool)
	svc = NewMessagingService(repo, RedisClient)
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