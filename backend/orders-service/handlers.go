package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Handler содержит все HTTP handlers для orders-service
type Handler struct {
	ordersService OrdersService
}

// NewHandler создает новый handler с dependency injection
func NewHandler(ordersService OrdersService) *Handler {
	return &Handler{
		ordersService: ordersService,
	}
}

// logUserActivity логирует активность пользователя, отправляя запрос к auth-service
func (h *Handler) logUserActivity(ctx context.Context, c *gin.Context, action, resourceType, details string, resourceID *uint) {
	userIDStr := c.GetHeader("X-User-ID")
	userEmail := c.GetHeader("X-User-Email")
	userName := c.GetHeader("X-User-Name")

	if userIDStr == "" {
		logrus.Warn("Cannot log activity - no user ID in headers")
		return
	}

	logData := map[string]interface{}{
		"action":        action,
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"details":       details,
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		logrus.WithError(err).Warn("Failed to marshal log data")
		return
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "http://localhost:8083/internal/log-activity", bytes.NewBuffer(jsonData))
	if err != nil {logrus.WithError(err).Warn("Failed to create log request")
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userIDStr)
	req.Header.Set("X-User-Email", userEmail)
	req.Header.Set("X-User-Name", userName)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Warn("Failed to send log request")
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status":   resp.StatusCode,
			"response": string(body),
		}).Warn("Log request failed")
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

	order, err := h.ordersService.CreateOrder(ctx, req)
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
