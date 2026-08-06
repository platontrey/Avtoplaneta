package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func (h *Handler) GetPartCatalogHandler(c *gin.Context) {
	if h.partCatalog == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Каталог шаблонов запчастей недоступен"})
		return
	}

	etag := `"` + h.partCatalog.Version + `"`
	c.Header("Cache-Control", "public, max-age=300, must-revalidate")
	c.Header("ETag", etag)
	if c.GetHeader("If-None-Match") == etag {
		c.Status(http.StatusNotModified)
		return
	}
	c.JSON(http.StatusOK, h.partCatalog)
}
