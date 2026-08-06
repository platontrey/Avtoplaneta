package main

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// SetupRoutes настраивает маршруты для приложения с dependency injection
func SetupRoutes(r *gin.Engine, handler *Handler) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "healthy"})
	})

	// Add metrics middleware
	r.Use(metricsMiddleware())

	// Add /metrics endpoint
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Основные маршруты сервиса запчастей
	r.GET("/api/inventory", handler.GetInventoryHandler)
	r.GET("/api/inventory/:id", handler.GetPartByIDHandler)
	r.POST("/api/addpart", handler.AddPartHandler)
	r.DELETE("/api/deletepart/:id", handler.DeletePartHandler)
	r.PUT("/api/updatepart/:id", handler.UpdatePartHandler)
	r.POST("/api/uploadpartphoto/:id", handler.UploadPartPhotoHandler)
	r.DELETE("/api/deletepartphoto/:id", handler.DeletePartPhotoHandler)
	r.POST("/api/markpartfordeletion/:id", handler.MarkPartForDeletionHandler)
	r.GET("/api/statistics", handler.GetStatisticsHandler)
	r.GET("/api/part-catalog", handler.GetPartCatalogHandler)
	r.POST("/api/statistics/update-earnings", handler.UpdateEarningsHandler)
	r.GET("/api/export/xml", exportXMLPriceList)
	r.POST("/api/export/drom", sendPriceListToDrom)

	// Маршруты для характеристик запчастей больше не нужны - характеристики хранятся в основной таблице Part

	// Маршруты для дефектных ведомостей
	r.POST("/api/defect-reports", handler.CreateDefectReportHandler)

	// Админ маршруты
	r.DELETE("/api/admin/delete-zero-quantity-parts/:supplier_code", handler.DeleteZeroQuantityPartsBySupplierHandler)
	r.GET("/api/admin/supplier-codes", handler.GetSupplierCodesHandler)
	r.DELETE("/api/admin/bulk-delete-parts", handler.BulkDeletePartsHandler)
	r.PUT("/api/admin/bulk-update-parts", handler.BulkUpdatePartsHandler)

	// Статическое обслуживание файлов для загрузок
	r.Static("/uploads", "./uploads")
}
