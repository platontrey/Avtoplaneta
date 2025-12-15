package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
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
func (h *Handler) logUserActivity(r *http.Request, action, resourceType, details string, resourceID *uint) {
	userIDStr := r.Header.Get("X-User-ID")
	userEmail := r.Header.Get("X-User-Email")
	userName := r.Header.Get("X-User-Name")

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

	req, err := http.NewRequest("POST", "http://localhost:8083/internal/log-activity", bytes.NewBuffer(jsonData))
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
func (h *Handler) GetOrdersHandler(w http.ResponseWriter, r *http.Request) {
	orders, err := h.ordersService.GetOrders()
	if err != nil {
		http.Error(w, "Failed to fetch orders", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// CreateOrderHandler создает новый заказ
func (h *Handler) CreateOrderHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	order, err := h.ordersService.CreateOrder(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Логируем создание заказа
	h.logUserActivity(r, "create_order", "order", fmt.Sprintf("Created order for buyer: %s", req.BuyerNumber), &order.ID)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(order); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

// UpdateOrderStatusHandler обновляет статус заказа
func (h *Handler) UpdateOrderStatusHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"]
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req struct {
		Status string `json:"status"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	if err := h.ordersService.UpdateOrderStatus(uint(orderID), req.Status); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Order status updated"}`))
}

// CompleteOrderHandler завершает заказ
func (h *Handler) CompleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"]
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if err := h.ordersService.CompleteOrder(uint(orderID)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Логируем завершение заказа
	orderIDUint := uint(orderID)
	h.logUserActivity(r, "complete_order", "order", fmt.Sprintf("Completed order ID: %d", orderID), &orderIDUint)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Order completed"}`))
}

// DeleteOrderHandler удаляет заказ
func (h *Handler) DeleteOrderHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"]
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	if err := h.ordersService.DeleteOrder(uint(orderID)); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Логируем удаление заказа
	orderIDUint := uint(orderID)
	h.logUserActivity(r, "delete_order", "order", fmt.Sprintf("Deleted order ID: %d", orderID), &orderIDUint)

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Order deleted"}`))
}

// AddOrderItemHandler добавляет позицию в заказ
func (h *Handler) AddOrderItemHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	orderIDStr := vars["id"]
	orderID, err := strconv.ParseUint(orderIDStr, 10, 32)
	if err != nil {
		http.Error(w, "Invalid order ID", http.StatusBadRequest)
		return
	}

	var req AddOrderItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request data", http.StatusBadRequest)
		return
	}

	if err := h.ordersService.AddOrderItem(uint(orderID), req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Item added to order"}`))
}
