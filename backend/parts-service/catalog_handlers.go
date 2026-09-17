package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"avtoplaneta/pkg/httpcache"
)

func (h *Handler) GetPartCatalogHandler(c *gin.Context) {
	if h.partCatalog == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Каталог шаблонов запчастей недоступен"})
		return
	}

	if httpcache.ServeVersioned(c.Writer, c.Request, h.partCatalog.Version, 5*time.Minute) {
		return
	}
	c.JSON(http.StatusOK, h.partCatalog)
}
