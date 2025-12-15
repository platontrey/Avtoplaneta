package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// exportXMLPriceList экспортирует прайс-лист в формате XML для Drom
func exportXMLPriceList(c *gin.Context) {
	// Получить все доступные части для экспорта
	parts, err := GetPartsForXML()
	if err != nil {
		fmt.Printf("Ошибка получения частей для XML экспорта: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить данные для экспорта"})
		return
	}

	// Генерировать XML
	xmlData, err := GenerateXMLPriceList(parts)
	if err != nil {
		fmt.Printf("Ошибка генерации XML: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать XML"})
		return
	}

	// Сохранить XML файл на сервере
	filename := "pricelist.xml"
	filepath := "./uploads/" + filename

	// Создать директорию uploads, если она не существует
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		fmt.Printf("Ошибка создания директории uploads: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать директорию uploads"})
		return
	}

	// Записать файл
	if err := os.WriteFile(filepath, xmlData, 0644); err != nil {
		fmt.Printf("Ошибка сохранения XML файла: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить XML файл"})
		return
	}

	// Ссылка на файл
	fileURL := "/uploads/" + filename

	fmt.Printf("Успешно сгенерирован и сохранен XML прайс-лист с %d предложениями по пути: %s\n", len(parts), filepath)

	c.JSON(http.StatusOK, gin.H{
		"message":     "Прайс-лист успешно сгенерирован",
		"file_url":    fileURL,
		"parts_count": len(parts),
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