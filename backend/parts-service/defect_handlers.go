package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var errPartCatalogUnavailable = errors.New("каталог шаблонов запчастей недоступен")

const defectReportEventVersion = 1

type DefectReportPreviewResponse struct {
	CatalogVersion string             `json:"catalog_version"`
	Total          int                `json:"total"`
	Parts          []DefectReportPart `json:"parts"`
}

// prepareDefectReport централизует разворачивание каталога. Новые клиенты не передают
// selectedParts; поддержка clientParts оставлена временно для уже установленных версий.
func (h *Handler) prepareDefectReport(report *DefectReportRequest, clientParts bool) error {
	if h.partCatalog == nil {
		return errPartCatalogUnavailable
	}

	if !clientParts || len(report.SelectedParts) == 0 {
		report.SelectedParts = h.partCatalog.ExpandDefectReport(*report)
	} else {
		h.partCatalog.ApplyBindingsToParts(report.SelectedParts, *report)
	}
	report.CatalogVersion = h.partCatalog.Version
	report.EventVersion = defectReportEventVersion
	return nil
}

// PreviewDefectReportHandler возвращает ровно тот набор запчастей, который будет позже создан.
func (h *Handler) PreviewDefectReportHandler(c *gin.Context) {
	var report DefectReportRequest
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Preview всегда строится сервером и не доверяет selectedParts из запроса.
	report.SelectedParts = nil
	if err := h.prepareDefectReport(&report, false); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DefectReportPreviewResponse{
		CatalogVersion: report.CatalogVersion,
		Total:          len(report.SelectedParts),
		Parts:          report.SelectedParts,
	})
}

// CreateDefectReportHandler отправляет дефектную ведомость в Redis Streams для асинхронной обработки
func (h *Handler) CreateDefectReportHandler(c *gin.Context) {
	var defectReportData DefectReportRequest

	if err := c.ShouldBindJSON(&defectReportData); err != nil {
		fmt.Printf("Invalid JSON in createDefectReport: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.prepareDefectReport(&defectReportData, true); err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
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
