package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
)

// OrdersService определяет интерфейс для бизнес-логики управления заказами
type OrdersService interface {
	// GetOrders Основные операции с заказами
	GetOrders() ([]Order, error)
	CreateOrder(req CreateOrderRequest) (*Order, error)
	UpdateOrderStatus(orderID uint, status string) error
	CompleteOrder(orderID uint) error
	DeleteOrder(orderID uint) error
	AddOrderItem(orderID uint, req AddOrderItemRequest) error
}

// CreateOrderRequest запрос на создание заказа
type CreateOrderRequest struct {
	CustomerID  int    `json:"customer_id"`
	Part        string `json:"part"`
	PartID      uint   `json:"part_id"`
	BuyerNumber string `json:"buyer_number"`
	Items       []struct {
		PartID   uint `json:"part_id"`
		Quantity int  `json:"quantity"`
	} `json:"items"`
}

// AddOrderItemRequest запрос на добавление позиции в заказ
type AddOrderItemRequest struct {
	PartID   uint `json:"part_id"`
	Quantity int  `json:"quantity"`
}

// ordersService реализует OrdersService
type ordersService struct {
	orderRepo OrderRepository
	partRepo  PartRepositoryForOrders
	cache     CacheService
}

// NewOrdersService создает новый сервис заказов
func NewOrdersService(orderRepo OrderRepository, partRepo PartRepositoryForOrders, cache CacheService) OrdersService {
	return &ordersService{
		orderRepo: orderRepo,
		partRepo:  partRepo,
		cache:     cache,
	}
}

// GetOrders получает все активные заказы с использованием кеша
func (s *ordersService) GetOrders() ([]Order, error) {
	// Сначала пытаемся получить из кеша
	cachedOrders, err := s.cache.GetOrders()
	if err != nil {
		logrus.WithError(err).Warn("Failed to get orders from cache, falling back to database")
	} else if cachedOrders != nil {
		logrus.Info("Returning orders from cache")
		return cachedOrders, nil
	}

	// Если в кеше нет данных, получаем из базы данных
	orders, err := s.orderRepo.FindActive()
	if err != nil {
		logrus.WithError(err).Error("Failed to get orders from database")
		return nil, err
	}

	// Форматируем данные для отображения
	now := time.Now()
	for i := range orders {
		orders[i].CreatedAtFormatted = orders[i].CreatedAt.Format("2006-01-02 15:04:05")
		orders[i].TimeAgo = formatTimeAgo(now.Sub(orders[i].CreatedAt))

		// Получаем location из первой позиции заказа
		if len(orders[i].Items) > 0 {
			part, err := s.partRepo.FindByID(orders[i].Items[0].PartID)
			if err != nil {
				logrus.WithError(err).WithField("part_id", orders[i].Items[0].PartID).Warn("Failed to get part location")
				orders[i].Location = "Неизвестно"
			} else {
				orders[i].Location = part.Location
			}
		} else {
			orders[i].Location = "Нет деталей"
		}
	}

	// Сохраняем в кеш
	if err := s.cache.SetOrders(orders); err != nil {
		logrus.WithError(err).Warn("Failed to cache orders")
	}

	return orders, nil
}

// CreateOrder создает новый заказ
func (s *ordersService) CreateOrder(req CreateOrderRequest) (*Order, error) {
	if req.BuyerNumber == "" {
		return nil, fmt.Errorf("buyer number is required")
	}

	if len(req.Items) == 0 {
		return nil, fmt.Errorf("at least one part must be selected")
	}

	// Получаем ID пользователя из контекста (предполагаем, что он установлен middleware)
	userIDStr := getUserIDFromContext() // TODO: реализовать получение из контекста
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID")
	}

	// Получаем имя пользователя
	userName := getUserNameFromContext() // TODO: реализовать

	// Создаем заказ
	order := &Order{
		CustomerID:  req.CustomerID,
		SellerID:    uint(userID),
		Seller:      userName,
		Part:        req.Part,
		PartID:      req.PartID,
		BuyerNumber: req.BuyerNumber,
		Status:      "red",
		StatusText:  "Need to order transport company",
		CreatedAt:   time.Now(),
	}

	if err := s.orderRepo.Create(order); err != nil {
		logrus.WithError(err).Error("Failed to create order")
		return nil, err
	}

	// Создаем позиции заказа и уменьшаем количество запчастей
	for _, item := range req.Items {
		part, err := s.partRepo.FindByID(item.PartID)
		if err != nil {
			logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to get part for order item")
			return nil, fmt.Errorf("failed to get part information")
		}

		orderItem := &OrderItem{
			OrderID:  order.ID,
			PartID:   item.PartID,
			Quantity: item.Quantity,
			Price:    part.Price,
		}

		if err := s.orderRepo.CreateItem(orderItem); err != nil {
			logrus.WithError(err).Error("Failed to create order item")
			return nil, err
		}

		// Уменьшаем количество запчасти
		if err := s.partRepo.DecreaseQuantity(item.PartID, item.Quantity); err != nil {
			logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to decrease part quantity")
			return nil, err
		}

		logrus.WithFields(logrus.Fields{
			"part_id":  item.PartID,
			"quantity": item.Quantity,
		}).Info("Decreased part quantity for order")
	}

	// Загружаем полный заказ с позициями
	completeOrder, err := s.orderRepo.FindByID(order.ID)
	if err != nil {
		logrus.WithError(err).WithField("order_id", order.ID).Error("Failed to load created order")
		return nil, err
	}

	// Форматируем данные
	now := time.Now()
	completeOrder.CreatedAtFormatted = completeOrder.CreatedAt.Format("2006-01-02 15:04:05")
	completeOrder.TimeAgo = formatTimeAgo(now.Sub(completeOrder.CreatedAt))

	// Инвалидируем кеш заказов
	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after creating order")
	}

	return completeOrder, nil
}

// UpdateOrderStatus обновляет статус заказа
func (s *ordersService) UpdateOrderStatus(orderID uint, status string) error {
	validStatuses := map[string]bool{
		"red":    true,
		"brown":  true,
		"yellow": true,
		"green":  true,
	}

	if !validStatuses[status] {
		return fmt.Errorf("invalid status: %s", status)
	}

	if err := s.orderRepo.UpdateStatus(orderID, status); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to update order status")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"status":   status,
	}).Info("Order status updated")

	// Инвалидируем кеш заказов
	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after updating status")
	}

	return nil
}

// CompleteOrder завершает заказ (переводит в статус green) и удаляет запчасти
func (s *ordersService) CompleteOrder(orderID uint) error {
	logrus.WithField("order_id", orderID).Info("Starting order completion")

	// Получаем заказ с позициями
	order, err := s.orderRepo.FindWithItemsByID(orderID)
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to find order for completion")
		return err
	}

	// Меняем статус
	err = s.UpdateOrderStatus(orderID, "green")
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to complete order")
		return err
	}

	// Рассчитываем сумму заказа
	var totalAmount float64
	for _, item := range order.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	// Удаляем запчасти полностью
	for _, item := range order.Items {
		if err := s.partRepo.DeletePart(item.PartID); err != nil {
			logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to delete part after order completion")
			// Продолжаем, не прерываем
		} else {
			logrus.WithField("part_id", item.PartID).Info("Part deleted after order completion")
		}
	}

	// Удаляем позиции заказа
	if err := s.orderRepo.DeleteItemsByOrderID(orderID); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to delete order items")
		return err
	}

	// Удаляем заказ
	if err := s.orderRepo.Delete(orderID); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to delete order")
		return err
	}

	// Обновляем статистику в parts-service
	if err := s.updatePartsStatistics(totalAmount); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Warn("Failed to update parts statistics")
		// Не прерываем, заказ завершен
	}

	// Инвалидируем кеш заказов
	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after completing order")
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"amount":   totalAmount,
	}).Info("Order completed successfully, order and parts deleted")

	return nil
}

// DeleteOrder удаляет заказ и возвращает запчасти в инвентарь
func (s *ordersService) DeleteOrder(orderID uint) error {
	logrus.WithField("order_id", orderID).Info("Starting order deletion")
	// Получаем заказ с позициями перед удалением
	orderWithItems, err := s.orderRepo.FindWithItemsByID(orderID)
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to find order with items for deletion")
		return fmt.Errorf("заказ не найден")
	}
	logrus.WithField("order_id", orderID).WithField("items_count", len(orderWithItems.Items)).Info("Found order with items for deletion")

	// Возвращаем запчасти в инвентарь
	for _, item := range orderWithItems.Items {
		if err := s.partRepo.IncreaseQuantity(item.PartID, item.Quantity); err != nil {
			logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to return part quantity to inventory")
			// Продолжаем, не прерываем
		} else {
			logrus.WithFields(logrus.Fields{
				"part_id":  item.PartID,
				"quantity": item.Quantity,
			}).Info("Returning part quantity to inventory")
		}
	}
	logrus.WithField("order_id", orderID).Info("Part quantities returned to inventory")

	// Удаляем позиции заказа перед удалением самого заказа
	logrus.WithField("order_id", orderID).Info("Deleting order items")
	if err := s.orderRepo.DeleteItemsByOrderID(orderID); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to delete order items")
		return err
	}

	logrus.WithField("order_id", orderID).Info("Attempting to delete order from database")
	if err := s.orderRepo.Delete(orderID); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to delete order")
		return err
	}

	// Инвалидируем кеш заказов
	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after deleting order")
	}

	logrus.WithField("order_id", orderID).Info("Order deleted and parts remain deducted")
	return nil
}

// AddOrderItem добавляет позицию в существующий заказ
func (s *ordersService) AddOrderItem(orderID uint, req AddOrderItemRequest) error {
	// Проверяем существование заказа
	_, err := s.orderRepo.FindByID(orderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	// Проверяем существование запчасти
	part, err := s.partRepo.FindByID(req.PartID)
	if err != nil {
		return fmt.Errorf("part not found")
	}

	// Проверяем, есть ли уже такая позиция в заказе
	existingItem, err := s.orderRepo.FindOrderItem(orderID, req.PartID)
	if err == nil {
		// Обновляем количество существующей позиции
		existingItem.Quantity += req.Quantity
		if err := s.orderRepo.UpdateItem(existingItem); err != nil {
			logrus.WithError(err).Error("Failed to update order item")
			return err
		}
	} else {
		// Создаем новую позицию
		orderItem := &OrderItem{
			OrderID:  orderID,
			PartID:   req.PartID,
			Quantity: req.Quantity,
			Price:    part.Price,
		}

		if err := s.orderRepo.CreateItem(orderItem); err != nil {
			logrus.WithError(err).Error("Failed to create order item")
			return err
		}
	}

	// Уменьшаем количество запчасти
	if err := s.partRepo.DecreaseQuantity(req.PartID, req.Quantity); err != nil {
		logrus.WithError(err).WithField("part_id", req.PartID).Error("Failed to decrease part quantity")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"part_id":  req.PartID,
		"quantity": req.Quantity,
	}).Info("Added item to order")

	// Инвалидируем кеш заказов
	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after adding item")
	}

	return nil
}

// Вспомогательные функции (нужно реализовать получение данных из контекста)
func getUserIDFromContext() string {
	// TODO: реализовать получение из gorilla/mux context или gin context
	return "1" // заглушка
}

func getUserNameFromContext() string {
	// TODO: реализовать получение из контекста
	return "Test User" // заглушка
}

func formatTimeAgo(duration time.Duration) string {
	// Простая реализация форматирования времени
	hours := int(duration.Hours())
	if hours < 1 {
		return "менее часа назад"
	} else if hours < 24 {
		return fmt.Sprintf("%d часов назад", hours)
	} else {
		days := hours / 24
		return fmt.Sprintf("%d дней назад", days)
	}
}

// updatePartsStatistics обновляет статистику в parts-service
func (s *ordersService) updatePartsStatistics(amount float64) error {
	reqData := map[string]interface{}{
		"amount": amount,
	}
	jsonData, err := json.Marshal(reqData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "http://localhost:8081/api/statistics/update-earnings", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("parts service returned status %d", resp.StatusCode)
	}

	return nil
}
