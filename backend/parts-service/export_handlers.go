package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// exportXMLPriceList экспортирует прайс-лист в формате XML для Drom.
// Файл пересобирается только если склад менялся с прошлой сборки.
func exportXMLPriceList(c *gin.Context) {
	meta, rebuilt, err := buildPriceList(c.Request.Context())
	if err != nil {
		fmt.Printf("Ошибка подготовки XML прайс-листа: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось подготовить прайс-лист"})
		return
	}

	message := "Прайс-лист актуален, пересборка не потребовалась"
	if rebuilt {
		message = "Прайс-лист успешно сгенерирован"
		fmt.Printf("Сгенерирован XML прайс-лист с %d предложениями: %s\n", meta.PartsCount, priceListPath)
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     message,
		"file_url":    priceListURL,
		"parts_count": meta.PartsCount,
		"rebuilt":     rebuilt,
	})
}

// sendPriceListToDrom отправляет прайс-лист на API Drom.ru
func sendPriceListToDrom(c *gin.Context) {
	// Получить все доступные части для экспорта
	parts, err := GetPartsForXML()
	if err != nil {
		fmt.Printf("Ошибка получения частей для отправки на Drom: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить данные для отправки"})
		return
	}

	// Генерировать XML
	xmlData, err := GenerateXMLPriceList(parts)
	if err != nil {
		fmt.Printf("Ошибка генерации XML для Drom: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать XML"})
		return
	}

	// Отправить на API Drom
	if err := sendToDromAPI(xmlData); err != nil {
		fmt.Printf("Ошибка отправки на Drom API: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось отправить прайс-лист на Drom"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":     "Прайс-лист успешно отправлен на Drom",
		"parts_count": len(parts),
	})
}
