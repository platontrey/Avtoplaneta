package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB
var RedisClient *redis.Client

func InitDatabase() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=qewret123 dbname=autoplanet port=5432 sslmode=disable"
	}

	var err error
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Автомиграция
	err = DB.AutoMigrate(
		&Conversation{},
		&Message{},
		&Mention{},
		&Reaction{},
		&Notification{},
		&UserStatus{},
	)
	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	// Выполнение кастомных миграций
	RunMigrations(DB)

	fmt.Println("Database connected and migrated successfully")
}

func InitRedis() {
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379"
	}

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		log.Fatal("Failed to parse Redis URL:", err)
	}

	RedisClient = redis.NewClient(opt)

	// Test connection
	_, err = RedisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	fmt.Println("Redis connected successfully")
}