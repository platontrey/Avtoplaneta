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

	// Алиасы для совместимости с внешними сборщиками (Drom) и старыми ссылками Joomla
	router.GET("/pricelist.xml", h.ServePriceList)
	router.HEAD("/pricelist.xml", h.ServePriceList)
	router.GET("/components/com_jshopping/files/importexport/multi_yml_export/Multi_YML_Export.xml", h.ServePriceList)
	router.HEAD("/components/com_jshopping/files/importexport/multi_yml_export/Multi_YML_Export.xml", h.ServePriceList)
}
