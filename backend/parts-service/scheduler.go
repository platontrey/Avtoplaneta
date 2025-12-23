package main

import (
	"context"
	"os"
	"runtime"
	"time"

	"github.com/sirupsen/logrus"
)

// StartXMLGenerationScheduler запускает планировщик автоматической генерации XML прайс-листа каждые 14 дней
func StartXMLGenerationScheduler(ctx context.Context) {
	ticker := time.NewTicker(14 * 24 * time.Hour) // 14 дней
	defer ticker.Stop()

	// Канал для worker pool (ограничение concurrency до 4 для использования многопроцессорности)
	jobs := make(chan func(), 10)

	// Worker goroutines (используем количество ядер для оптимальной производительности)
	numWorkers := runtime.NumCPU()
	for i := 0; i < numWorkers; i++ {
		go func() {
			for job := range jobs {
				job()
			}
		}()
	}

	// Генерируем XML сразу при запуске
	logrus.Info("Запуск начальной генерации XML прайс-листа...")
	jobs <- generateXMLPriceList

	for {
		select {
		case <-ctx.Done():
			close(jobs)
			logrus.Info("Остановка планировщика генерации XML")
			return
		case <-ticker.C:
			logrus.Info("Автоматическая генерация XML прайс-листа каждые 14 дней...")
			jobs <- generateXMLPriceList
		}
	}
}

// generateXMLPriceList генерирует и сохраняет XML прайс-лист
func generateXMLPriceList() {
	parts, err := GetPartsForXML()
	if err != nil {
		logrus.WithError(err).Error("Ошибка получения частей для автоматической генерации XML")
		return
	}

	xmlData, err := GenerateXMLPriceList(parts)
	if err != nil {
		logrus.WithError(err).Error("Ошибка генерации XML для автоматической генерации")
		return
	}

	// Сохранить XML файл на сервере
	filename := "pricelist.xml"
	filepath := "./uploads/" + filename

	// Создать директорию uploads, если она не существует
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		logrus.WithError(err).Error("Ошибка создания директории uploads для автоматической генерации")
		return
	}

	// Записать файл
	if err := os.WriteFile(filepath, xmlData, 0644); err != nil {
		logrus.WithError(err).Error("Ошибка сохранения XML файла для автоматической генерации")
		return
	}

	logrus.WithField("parts_count", len(parts)).Info("Успешно автоматически сгенерирован и сохранен XML прайс-лист")
}
