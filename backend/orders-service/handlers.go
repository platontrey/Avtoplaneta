package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler содержит все HTTP handlers для orders-service
type Handler struct {
	ordersService OrdersService
	publisher     EventPublisher
}

// NewHandler создает новый handler с dependency injection
func NewHandler(ordersService OrdersService, publisher EventPublisher) *Handler {
	return &Handler{
		ordersService: ordersService,
		publisher:     publisher,
	}
}

// logUserActivity логирует активность пользователя через Redis Streams
func (h *Handler) logUserActivity(ctx context.Context, c *gin.Context, action, resourceType, details string, resourceID *uint) {
	userIDStr := c.GetHeader("X-User-ID")
	userEmail := c.GetHeader("X-User-Email")
	userName := c.GetHeader("X-User-Name")

	if userIDStr == "" {
		logrus.Warn("Cannot log activity - no user ID in headers")
		return
	}

	eventDetails := map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"details":       details,
		"user_email":    userEmail,
		"user_name":     userName,
	}

	if err := h.publisher.PublishUserAction(ctx, userIDStr, action, eventDetails); err != nil {
		logrus.WithError(err).Warn("Failed to publish user action event")
	}
}

// GetOrdersHandler обрабатывает запрос на получение заказов
func (h *Handler) GetOrdersHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orders, err := h.ordersService.GetOrders(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// CreateOrderHandler создает новый заказ
func (h *Handler) CreateOrderHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	userName := c.GetHeader("X-User-Name")

	order, err := h.ordersService.CreateOrder(ctx, req, uint(userID), userName)
	if err != nil {
		if IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		}
		return
	}

	// Логируем создание заказа
	h.logUserActivity(ctx, c, "create_order", "order", fmt.Sprintf("Created order for buyer: %s", req.BuyerNumber), &order.ID)

	c.JSON(http.StatusCreated, order)
}

// UpdateOrderStatusHandler обновляет статус заказа
func (h *Handler) UpdateOrderStatusHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.ordersService.UpdateOrderStatus(ctx, uint(orderID), req.Status); err != nil {
		if IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated"})
}

// CompleteOrderHandler завершает заказ
func (h *Handler) CompleteOrderHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	if err := h.ordersService.CompleteOrder(ctx, uint(orderID)); err != nil {
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete order"})
		}
		return
	}

	// Логируем завершение заказа
	orderIDUint := uint(orderID)
	h.logUserActivity(ctx, c, "complete_order", "order", fmt.Sprintf("Completed order ID: %d", orderID), &orderIDUint)

	c.JSON(http.StatusOK, gin.H{"message": "Order completed"})
}

// DeleteOrderHandler удаляет заказ
func (h *Handler) DeleteOrderHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	if err := h.ordersService.DeleteOrder(ctx, uint(orderID)); err != nil {
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		}
		return
	}

	// Логируем удаление заказа
	orderIDUint := uint(orderID)
	h.logUserActivity(ctx, c, "delete_order", "order", fmt.Sprintf("Deleted order ID: %d", orderID), &orderIDUint)

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted"})
}

// AddOrderItemHandler добавляет позицию в заказ
func (h *Handler) AddOrderItemHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req AddOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.ordersService.AddOrderItem(ctx, uint(orderID), req); err != nil {
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to order"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item added to order"})
}

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
