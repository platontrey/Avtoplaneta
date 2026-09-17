package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"avtoplaneta/pkg/httpcache"
)

// GetVehicleCatalogHandler отдаёт справочник марок и моделей обоим клиентам.
// Версия справочника служит ETag, поэтому повторный заход стоит один 304.
func (h *Handler) GetVehicleCatalogHandler(c *gin.Context) {
	if h.vehicleCatalog == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Справочник марок и моделей недоступен"})
		return
	}

	if httpcache.ServeVersioned(c.Writer, c.Request, h.vehicleCatalog.Version, 5*time.Minute) {
		return
	}
	c.JSON(http.StatusOK, h.vehicleCatalog)
}
