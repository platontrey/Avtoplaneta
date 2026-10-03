package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetMonthlySalesHandler получает продажи по месяцам
func (h *Handler) GetMonthlySalesHandler(c *gin.Context) {
	ctx := c.Request.Context()
	sales, err := h.ordersService.GetMonthlySales(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch monthly sales"})
		return
	}

	c.JSON(http.StatusOK, sales)
}
