package main

import (
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

// SetupRoutes описывает весь внешний договор сервиса.
//
// Адреса намеренно оставлены прежними: снаружи ничего не меняется, прайс-лист
// как лежал по /uploads/pricelist.xml, так и лежит. Меняется только то, кто его
// отдаёт, — и это внутреннее дело, а не повод ломать ссылку у Drom.
func SetupRoutes(router *gin.Engine, handler *Handler) {
	router.GET("/health", handler.healthCheck)

	router.GET("/api/export/xml", handler.exportXMLPriceList)
	router.POST("/api/export/drom", handler.sendPriceListToDrom)

	router.GET(priceListURL, handler.servePriceList)
}

// servePriceList отдаёт готовый файл.
//
// Прайс-лист живёт по постоянному адресу и регулярно перезаписывается, поэтому
// immutable ему нельзя: Drom получил бы прошлогодний файл. Оставляем
// обязательную перепроверку — она дешёвая, ServeFile отдаёт Last-Modified и
// отвечает 304.
func (h *Handler) servePriceList(c *gin.Context) {
	path := h.builder.Path()

	if _, err := os.Stat(path); err != nil {
		// Файла ещё нет — первый запрос собирает его сам, вместо того чтобы
		// отдавать 404 и ждать планировщика.
		if _, _, buildErr := h.builder.Build(c.Request.Context()); buildErr != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Прайс-лист пока недоступен"})
			return
		}
	}

	c.Header("Cache-Control", "no-cache")
	c.File(path)
}
