package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"orders-service/internal/domain"
)

// ListCustomersHandler возвращает список клиентов с фильтрацией и поиском
func (h *Handler) ListCustomersHandler(c *gin.Context) {
	ctx := c.Request.Context()
	category := c.Query("category")
	search := c.Query("q")
	if search == "" {
		search = c.Query("search")
	}

	limit := int32(50)
	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = int32(parsed)
		}
	}

	offset := int32(0)
	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = int32(parsed)
		}
	}

	customers, err := h.ordersService.ListCustomers(ctx, category, search, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customers"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"customers": customers,
		"total":     len(customers),
	})
}

// GetCustomerHandler возвращает карточку клиента и историю его заказов
func (h *Handler) GetCustomerHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	details, err := h.ordersService.GetCustomer(ctx, id)
	if err != nil {
		if domain.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch customer details"})
		}
		return
	}

	c.JSON(http.StatusOK, details)
}

// CreateCustomerHandler создаёт карточку клиента
func (h *Handler) CreateCustomerHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var req domain.CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	customer, err := h.ordersService.CreateCustomer(ctx, req)
	if err != nil {
		if domain.IsConflictError(err) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		} else if domain.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create customer"})
		}
		return
	}

	c.JSON(http.StatusCreated, customer)
}

// UpdateCustomerHandler обновляет карточку клиента
func (h *Handler) UpdateCustomerHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	var req domain.UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	customer, err := h.ordersService.UpdateCustomer(ctx, id, req)
	if err != nil {
		if domain.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if domain.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update customer"})
		}
		return
	}

	c.JSON(http.StatusOK, customer)
}

// DeleteCustomerHandler удаляет карточку клиента
func (h *Handler) DeleteCustomerHandler(c *gin.Context) {
	ctx := c.Request.Context()
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid customer ID"})
		return
	}

	if err := h.ordersService.DeleteCustomer(ctx, id); err != nil {
		if domain.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}
