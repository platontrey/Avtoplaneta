package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"avtoplaneta/pkg/httpcache"
)

// GetStatisticsHandler возвращает статистику по запчастям
func (h *Handler) GetStatisticsHandler(c *gin.Context) {
	ctx := c.Request.Context()

	if version, err := h.inventoryService.InventoryVersion(ctx); err == nil {
		if httpcache.ServeVersioned(c.Writer, c.Request, version, time.Minute) {
			return
		}
	}

	stats, err := h.inventoryService.GetStatistics(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить статистику"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// UpdateEarningsHandler обновляет общий заработок
func (h *Handler) UpdateEarningsHandler(c *gin.Context) {
	var req struct {
		Amount float64 `json:"amount" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	ctx := c.Request.Context()
	if err := h.inventoryService.UpdateEarnings(ctx, req.Amount); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить заработок"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Заработок обновлен"})
}
