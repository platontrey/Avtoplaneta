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
func (h *Handler) logUserActivity(ctx context.Context, c *gin.Context, action, resourceType, details string, resourceID *int64) {
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

// GetOrdersHandler обрабатывает запрос на получение заказов (активных или завершённых по ?state=completed)
func (h *Handler) GetOrdersHandler(c *gin.Context) {
	ctx := c.Request.Context()
	state := c.Query("state")

	var (
		orders []Order
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

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	userName := c.GetHeader("X-User-Name")

	order, err := h.ordersService.CreateOrder(ctx, req, userID, userName)
	if err != nil {
		if IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
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
		if IsValidationError(err) {
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

	var req UpdateOrderDetailsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	updated, err := h.ordersService.UpdateOrderDetails(ctx, orderID, req)
	if err != nil {
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if IsValidationError(err) {
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
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to complete order"})
		}
		return
	}

	h.logUserActivity(ctx, c, "complete_order", "order", fmt.Sprintf("Completed order ID: %d", orderID), &orderID)

	c.JSON(http.StatusOK, gin.H{"message": "Order completed"})
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
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order"})
		}
		return
	}

	h.logUserActivity(ctx, c, "delete_order", "order", fmt.Sprintf("Deleted order ID: %d", orderID), &orderID)

	c.JSON(http.StatusOK, gin.H{"message": "Order deleted"})
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

	var req AddOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.ordersService.AddOrderItem(ctx, orderID, req); err != nil {
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to order"})
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

	var req UpdateOrderItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	if err := h.ordersService.UpdateOrderItem(ctx, orderID, itemID, req); err != nil {
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if IsValidationError(err) {
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
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if IsValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete order item"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order item deleted"})
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
		if IsNotFoundError(err) {
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
	var req CreateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	customer, err := h.ordersService.CreateCustomer(ctx, req)
	if err != nil {
		if IsValidationError(err) {
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

	var req UpdateCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	customer, err := h.ordersService.UpdateCustomer(ctx, id, req)
	if err != nil {
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else if IsValidationError(err) {
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
		if IsNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete customer"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Customer deleted successfully"})
}

