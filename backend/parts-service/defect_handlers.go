package main

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// CreateDefectReportHandler отправляет дефектную ведомость в Redis Streams для асинхронной обработки
func (h *Handler) CreateDefectReportHandler(c *gin.Context) {
	var defectReportData DefectReportRequest

	if err := c.ShouldBindJSON(&defectReportData); err != nil {
		fmt.Printf("Invalid JSON in createDefectReport: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Новые клиенты отправляют только данные автомобиля и характеристики.
	// Старый selectedParts временно поддерживается для уже установленных версий.
	if len(defectReportData.SelectedParts) == 0 {
		if h.partCatalog == nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Каталог шаблонов запчастей недоступен"})
			return
		}
		defectReportData.SelectedParts = h.partCatalog.ExpandDefectReport(defectReportData)
	} else if h.partCatalog != nil {
		h.partCatalog.ApplyBindingsToParts(defectReportData.SelectedParts, defectReportData)
	}

	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr != "" {
		fmt.Sscanf(userIDStr, "%d", &defectReportData.SellerID)
	}
	defectReportData.SellerName = c.GetHeader("X-User-Name")

	if defectReportData.SellerName == "" {
		defectReportData.SellerName = "System"
		defectReportData.SellerID = 1
	}

	fmt.Printf("Queueing defect report for %s %s %d with %d parts (Seller: %s)\n", defectReportData.Brand, defectReportData.Model, defectReportData.Year, len(defectReportData.SelectedParts), defectReportData.SellerName)

	payload, err := json.Marshal(defectReportData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка сериализации данных"})
		return
	}

	// Отправляем дефектную ведомость в Redis Stream
	err = redisClient.XAdd(c.Request.Context(), &redis.XAddArgs{
		Stream: "events:orders",
		Values: map[string]interface{}{
			"type": "defect_report_created",
			"data": string(payload),
		},
	}).Err()
	if err != nil {
		fmt.Printf("Failed to publish defect report event to Redis: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось отправить ведомость в очередь обработки"})
		return
	}

	// Логируем активность пользователя
	h.logUserActivity(c, "queue_defect_report", "part", fmt.Sprintf("Queued defect report with %d parts for %s %s %d", len(defectReportData.SelectedParts), defectReportData.Brand, defectReportData.Model, defectReportData.Year), nil)

	c.JSON(http.StatusAccepted, gin.H{
		"message": "Дефектная ведомость отправлена в очередь обработки",
	})
}
