package main

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

// StartXMLGenerationScheduler запускает планировщик автоматической генерации XML прайс-листа
func StartXMLGenerationScheduler(ctx context.Context) {
	logrus.Info("Starting XML generation scheduler...")

	// Запускаем первую проверку при старте в отдельной горутине
	go checkAndGenerateXML(ctx)

	// Проверяем статус каждые 1 час
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Stopping XML generation scheduler")
			return
		case <-ticker.C:
			go checkAndGenerateXML(ctx)
		}
	}
}

// checkAndGenerateXML проверяет время последней генерации в Redis и при необходимости запускает новую
func checkAndGenerateXML(ctx context.Context) {
	if redisClient == nil {
		logrus.Warn("Redis client not initialized, skipping XML scheduler check")
		return
	}

	key := "xml_last_generated_at"
	lastGenStr, err := redisClient.Get(ctx, key).Result()

	var shouldGenerate bool
	if errors.Is(err, redis.Nil) {
		// Ключа еще нет в Redis — генерируем XML впервые
		shouldGenerate = true
	} else if err != nil {
		logrus.WithError(err).Error("Failed to get last XML generation timestamp from Redis")
		return
	} else {
		lastGenTime, parseErr := time.Parse(time.RFC3339, lastGenStr)
		if parseErr != nil {
			logrus.WithError(parseErr).Warn("Failed to parse last XML generation timestamp, forcing regeneration")
			shouldGenerate = true
		} else if time.Since(lastGenTime) >= 14*24*time.Hour {
			shouldGenerate = true
		}
	}

	if shouldGenerate {
		logrus.Info("Time threshold exceeded. Starting XML price list generation...")
		generateXMLPriceList()

		// Обновляем метку времени в Redis
		err := redisClient.Set(ctx, key, time.Now().Format(time.RFC3339), 0).Err()
		if err != nil {
			logrus.WithError(err).Error("Failed to save XML generation timestamp to Redis")
		}
	}
}

// generateXMLPriceList генерирует и сохраняет XML прайс-лист.
// Та же функция сборки, что и у ручного экспорта: если склад не менялся,
// планировщик не тратит время на пересборку.
func generateXMLPriceList() {
	meta, rebuilt, err := buildPriceList(context.Background())
	if err != nil {
		logrus.WithError(err).Error("Ошибка автоматической генерации XML прайс-листа")
		return
	}
	if !rebuilt {
		logrus.Info("XML прайс-лист актуален, пересборка пропущена")
		return
	}

	logrus.WithField("parts_count", meta.PartsCount).Info("Успешно автоматически сгенерирован и сохранен XML прайс-лист")
}
