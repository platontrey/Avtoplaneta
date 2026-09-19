package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DefectReportPreviewResponse struct {
	CatalogVersion string             `json:"catalog_version"`
	Total          int                `json:"total"`
	Parts          []DefectReportPart `json:"parts"`
}

// PreviewDefectReportHandler возвращает ровно тот набор запчастей, который будет позже создан.
func (h *Handler) PreviewDefectReportHandler(c *gin.Context) {
	var report DefectReportRequest
	if err := c.ShouldBindJSON(&report); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	prepared, err := h.defectReports.Preview(report)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, DefectReportPreviewResponse{
		CatalogVersion: prepared.CatalogVersion,
		Total:          len(prepared.SelectedParts),
		Parts:          prepared.SelectedParts,
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

	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr != "" {
		fmt.Sscanf(userIDStr, "%d", &defectReportData.SellerID)
	}
	defectReportData.SellerName = c.GetHeader("X-User-Name")

	if defectReportData.SellerName == "" {
		defectReportData.SellerName = "System"
		defectReportData.SellerID = 1
	}

	// TODO(legacy): заменить true на false вместе со снятием allowLegacyClientParts.
	createdParts, err := h.defectReports.Create(c.Request.Context(), &defectReportData, true)
	if err != nil {
		fmt.Printf("Failed to process defect report: %v\n", err)
		if defectReportUnavailable(err) {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обработать дефектную ведомость"})
		return
	}
	fmt.Printf("Created defect report for %s %s %d with %d parts (Seller: %s)\n", defectReportData.Brand, defectReportData.Model, defectReportData.Year, len(createdParts), defectReportData.SellerName)

	// Логируем активность пользователя
	h.logUserActivity(c, "create_defect_report", "part", fmt.Sprintf("Created defect report with %d parts for %s %s %d", len(createdParts), defectReportData.Brand, defectReportData.Model, defectReportData.Year), nil)

	c.JSON(http.StatusOK, gin.H{
		"message": "Дефектная ведомость успешно обработана",
		"total":   len(createdParts),
	})
}
