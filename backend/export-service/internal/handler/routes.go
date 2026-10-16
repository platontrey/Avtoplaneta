package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"export-service/internal/service"
)

// SetupRoutes конфигурирует HTTP-маршруты сервиса экспорта
func SetupRoutes(router *gin.Engine, h *Handler) {
	router.GET("/health", h.HealthCheck)
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	router.GET("/api/export/xml", h.ExportXMLPriceList)
	router.POST("/api/export/drom", h.SendPriceListToDrom)
	router.GET("/api/export/status", h.GetExportStatus)

	router.GET(service.PriceListURL, h.ServePriceList)
}
