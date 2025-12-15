package main

import (
	"fmt"
	"os"
	"time"
)

// StartXMLGenerationScheduler запускает планировщик автоматической генерации XML прайс-листа каждые 14 дней
func StartXMLGenerationScheduler() {
	ticker := time.NewTicker(14 * 24 * time.Hour) // 14 дней
	defer ticker.Stop()

	// Генерируем XML сразу при запуске
	fmt.Println("Запуск начальной генерации XML прайс-листа...")
	generateXMLPriceList()

	for {
		select {
		case <-ticker.C:
			fmt.Println("Автоматическая генерация XML прайс-листа каждые 14 дней...")
			generateXMLPriceList()
		}
	}
}

// generateXMLPriceList генерирует и сохраняет XML прайс-лист
func generateXMLPriceList() {
	parts, err := GetPartsForXML()
	if err != nil {
		fmt.Printf("Ошибка получения частей для автоматической генерации XML: %v\n", err)
		return
	}

	xmlData, err := GenerateXMLPriceList(parts)
	if err != nil {
		fmt.Printf("Ошибка генерации XML для автоматической генерации: %v\n", err)
		return
	}

	// Сохранить XML файл на сервере
	filename := "pricelist.xml"
	filepath := "./uploads/" + filename

	// Создать директорию uploads, если она не существует
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		fmt.Printf("Ошибка создания директории uploads для автоматической генерации: %v\n", err)
		return
	}

	// Записать файл
	if err := os.WriteFile(filepath, xmlData, 0644); err != nil {
		fmt.Printf("Ошибка сохранения XML файла для автоматической генерации: %v\n", err)
		return
	}

	fmt.Printf("Успешно автоматически сгенерирован и сохранен XML прайс-лист с %d предложениями\n", len(parts))
}