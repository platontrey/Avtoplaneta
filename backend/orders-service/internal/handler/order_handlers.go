package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"avtoplaneta/pkg/authcontext"
	"avtoplaneta/pkg/logger"
	"orders-service/internal/domain"
)

// GetOrdersHandler обрабатывает запрос на получение заказов (активных или завершённых по ?state=completed)
func (h *Handler) GetOrdersHandler(c *gin.Context) {
	ctx := c.Request.Context()
	state := c.Query("state")

	var (
		orders []domain.Order
		err    error
	)
	if state == "completed" {
		orders, err = h.ordersService.GetCompletedOrders(ctx)
	} else {
		orders, err = h.ordersService.GetOrders(ctx)
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	c.JSON(http.StatusOK, orders)
}

// GetCompletedOrdersHandler обрабатывает запрос на получение истории завершённых заказов
func (h *Handler) GetCompletedOrdersHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orders, err := h.ordersService.GetCompletedOrders(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch completed orders"})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// CreateOrderHandler создает новый заказ (или проводит быструю продажу)
func (h *Handler) CreateOrderHandler(c *gin.Context) {
	ctx := c.Request.Context()
	var req domain.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	user, ok := authcontext.FromRequest(c.Request)
	if !ok || !user.IsAuthenticated() {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
		return
	}

	order, err := h.ordersService.CreateOrder(ctx, req, user.ID, user.Name)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to create order",
			slog.Int64("user_id", user.ID),
			slog.String("user_name", user.Name),
			slog.String("buyer_number", req.BuyerNumber),
			slog.String("order_number", req.OrderNumber),
			slog.Int("items_count", len(req.Items)),
			logger.Err(err),
		)

		if domain.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to create order",
				"details": err.Error(),
			})
		}
		return
	}

	actionName := "create_order"
	if req.QuickSale {
		actionName = "quick_sale_order"
	}
	h.logUserActivity(ctx, c, actionName, "order", fmt.Sprintf("Created order for buyer: %s", order.BuyerNumber), &order.ID)

	c.JSON(http.StatusCreated, order)
}

// UpdateOrderStatusHandler обновляет статус заказа
func (h *Handler) UpdateOrderStatusHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
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

	if err := h.ordersService.UpdateOrderStatus(ctx, orderID, req.Status); err != nil {
		if domain.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order status updated"})
}

// UpdateOrderDetailsHandler обновляет детали заказа (оплата, склад, доставка, ТК, трек-номер, заметки, скидка)
func (h *Handler) UpdateOrderDetailsHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req domain.UpdateOrderDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	updated, err := h.ordersService.UpdateOrderDetails(ctx, orderID, req)
	if err != nil {
		if domain.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if domain.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order"})
		}
		return
	}

	c.JSON(http.StatusOK, updated)
}

// CompleteOrderHandler завершает заказ
func (h *Handler) CompleteOrderHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	if err := h.ordersService.CompleteOrder(ctx, orderID); err != nil {
		if domain.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete order"})
		}
		return
	}

	h.logUserActivity(ctx, c, "complete_order", "order", fmt.Sprintf("Completed order: %d", orderID), &orderID)

	c.JSON(http.StatusOK, gin.H{"message": "Order completed successfully"})
}

// DeleteOrderHandler удаляет заказ
func (h *Handler) DeleteOrderHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	if err := h.ordersService.DeleteOrder(ctx, orderID); err != nil {
		if domain.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		}
		return
	}

	h.logUserActivity(ctx, c, "delete_order", "order", fmt.Sprintf("Deleted order: %d", orderID), &orderID)

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted successfully"})
}

// AddOrderItemHandler добавляет позицию в заказ
func (h *Handler) AddOrderItemHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderIDStr := c.Param("id")
	orderID, err := strconv.ParseInt(orderIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}

	var req domain.AddOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.ordersService.AddOrderItem(ctx, orderID, req); err != nil {
		slog.ErrorContext(ctx, "Failed to add item to order",
			slog.Int64("order_id", orderID),
			slog.Int64("part_id", req.PartID),
			slog.Int("quantity", req.Quantity),
			logger.Err(err),
		)

		if domain.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if domain.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Failed to add item to order",
				"details": err.Error(),
			})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item added to order"})
}

// UpdateOrderItemHandler обновляет количество или цену позиции в заказе
func (h *Handler) UpdateOrderItemHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	var req domain.UpdateOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.ordersService.UpdateOrderItem(ctx, orderID, itemID, req); err != nil {
		if domain.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if domain.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order item"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order item updated"})
}

// DeleteOrderItemHandler удаляет позицию из заказа
func (h *Handler) DeleteOrderItemHandler(c *gin.Context) {
	ctx := c.Request.Context()
	orderID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid order ID"})
		return
	}
	itemID, err := strconv.ParseInt(c.Param("itemId"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid item ID"})
		return
	}

	if err := h.ordersService.DeleteOrderItem(ctx, orderID, itemID); err != nil {
		if domain.IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if domain.IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order item"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order item deleted"})
}
