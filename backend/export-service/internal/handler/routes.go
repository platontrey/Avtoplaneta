package handler

import (
	"avtoplaneta/pkg/metrics"
	"github.com/gin-gonic/gin"

	"export-service/internal/service"
)

// SetupRoutes конфигурирует HTTP-маршруты сервиса экспорта
func SetupRoutes(router *gin.Engine, h *Handler) {
	router.Use(metrics.GinMiddleware("export-service"))

	router.GET("/health", h.HealthCheck)
	router.GET("/metrics", metrics.Handler())

	router.GET("/api/export/xml", h.ExportXMLPriceList)
	router.POST("/api/export/drom", h.SendPriceListToDrom)
	router.GET("/api/export/status", h.GetExportStatus)

	router.GET(service.PriceListURL, h.ServePriceList)
	router.HEAD(service.PriceListURL, h.ServePriceList)
}
