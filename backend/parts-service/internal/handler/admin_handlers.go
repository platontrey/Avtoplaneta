package handler

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"parts-service/internal/domain"
)

// BulkDeletePartsHandler удаляет несколько запчастей
func (h *Handler) BulkDeletePartsHandler(c *gin.Context) {
	var requestData struct {
		IDs []int64 `json:"ids"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		logrus.WithError(err).Error("BulkDeletePartsHandler: Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"ids":   requestData.IDs,
		"count": len(requestData.IDs),
	}).Info("BulkDeletePartsHandler: Received request")

	ctx := c.Request.Context()
	if err := h.inventoryService.BulkDeleteParts(ctx, requestData.IDs); err != nil {
		logrus.WithError(err).Error("BulkDeletePartsHandler: Failed to bulk delete parts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить запчасти"})
		return
	}

	// Логируем массовое удаление
	h.logUserActivity(c, "bulk_delete_parts", "part", fmt.Sprintf("Bulk deleted %d parts", len(requestData.IDs)), nil)

	logrus.Info("BulkDeletePartsHandler: Successfully deleted parts")
	c.JSON(http.StatusOK, gin.H{"message": "Запчасти удалены успешно"})
}

// BulkUpdatePartsHandler обновляет несколько запчастей
func (h *Handler) BulkUpdatePartsHandler(c *gin.Context) {
	var updates []map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		logrus.WithError(err).Error("BulkUpdatePartsHandler: Failed to bind JSON")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат данных"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"updates": updates,
		"count":   len(updates),
	}).Info("BulkUpdatePartsHandler: Received request")

	ctx := c.Request.Context()
	updatedCount, err := h.inventoryService.BulkUpdateParts(ctx, updates)
	if err != nil {
		logrus.WithError(err).Error("BulkUpdatePartsHandler: Failed to bulk update parts")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить запчасти"})
		return
	}

	// Логируем массовое обновление
	h.logUserActivity(c, "bulk_update_parts", "part", fmt.Sprintf("Bulk updated %d parts", updatedCount), nil)

	logrus.Info("BulkUpdatePartsHandler: Successfully updated parts")
	c.JSON(http.StatusOK, gin.H{
		"message":       "Запчасти обновлены успешно",
		"updated_count": updatedCount,
	})
}

// DeleteZeroQuantityPartsBySupplierHandler удаляет запчасти с нулевым количеством по поставщику
func (h *Handler) DeleteZeroQuantityPartsBySupplierHandler(c *gin.Context) {
	supplierCode := strings.TrimSpace(c.Param("supplier_code"))
	if supplierCode == "" {
		supplierCode = strings.TrimSpace(c.Query("supplier_code"))
	}
	if supplierCode == "" {
		var body struct {
			SupplierCode string `json:"supplier_code"`
		}
		if err := c.ShouldBindJSON(&body); err == nil {
			supplierCode = strings.TrimSpace(body.SupplierCode)
		}
	}
	if supplierCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указан код поставки"})
		return
	}

	fmt.Printf("DeleteZeroQuantityPartsBySupplierHandler: supplier_code='%s'\n", supplierCode)

	ctx := c.Request.Context()
	deletedCount, err := h.inventoryService.DeleteZeroQuantityPartsBySupplier(ctx, supplierCode)
	if err != nil {
		fmt.Printf("DeleteZeroQuantityPartsBySupplierHandler: error deleting parts for supplier_code='%s': %v\n", supplierCode, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить запчасти"})
		return
	}

	fmt.Printf("DeleteZeroQuantityPartsBySupplierHandler: successfully deleted %d parts for supplier_code='%s'\n", deletedCount, supplierCode)

	// Логируем массовое удаление запчастей с quantity=0
	h.logUserActivity(c, "delete_zero_quantity_parts", "part", fmt.Sprintf("Deleted %d zero-quantity parts for supplier code: %s", deletedCount, supplierCode), nil)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Запчасти с нулевым количеством удалены",
		"deleted_count": deletedCount,
	})
}

// GetSupplierCodesHandler получает коды поставщиков
func (h *Handler) GetSupplierCodesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	c.Header("Cache-Control", "no-store, no-cache, must-revalidate")

	var batches []domain.SupplierBatchInfo
	if batchProvider, ok := h.inventoryService.(interface {
		GetSupplierBatches(ctx context.Context) ([]domain.SupplierBatchInfo, error)
	}); ok {
		var err error
		batches, err = batchProvider.GetSupplierBatches(ctx)
		if err != nil {
			logrus.WithError(err).Error("Handler: GetSupplierBatches failed")
		}
	}

	codes, err := h.inventoryService.GetSupplierCodes(ctx)
	if err != nil {
		logrus.WithError(err).Error("Handler: GetSupplierCodes failed")
		codes = []string{}
	}

	if len(batches) == 0 && len(codes) > 0 {
		batches = make([]domain.SupplierBatchInfo, 0, len(codes))
		for _, code := range codes {
			batches = append(batches, domain.SupplierBatchInfo{Code: code, Label: code})
		}
	}
	if batches == nil {
		batches = []domain.SupplierBatchInfo{}
	}
	if codes == nil {
		codes = []string{}
	}

	c.JSON(http.StatusOK, gin.H{
		"supplier_codes": codes,
		"codes":          codes,
		"batches":        batches,
	})
}
