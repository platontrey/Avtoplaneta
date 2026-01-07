package main

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
)

// OrdersService определяет интерфейс для бизнес-логики управления заказами
type OrdersService interface {
	// GetOrders Основные операции с заказами
	GetOrders(ctx context.Context) ([]Order, error)
	CreateOrder(ctx context.Context, req CreateOrderRequest, userID uint, userName string) (*Order, error)
	UpdateOrderStatus(ctx context.Context, orderID uint, status string) error
	CompleteOrder(ctx context.Context, orderID uint) error
	DeleteOrder(ctx context.Context, orderID uint) error
	AddOrderItem(ctx context.Context, orderID uint, req AddOrderItemRequest) error
	GetMonthlySales(ctx context.Context) ([]MonthlySales, error)
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
	publisher EventPublisher
}

// NewOrdersService создает новый сервис заказов
func NewOrdersService(orderRepo OrderRepository, partRepo PartRepositoryForOrders, cache CacheService, publisher EventPublisher) OrdersService {
	return &ordersService{
		orderRepo: orderRepo,
		partRepo:  partRepo,
		cache:     cache,
		publisher: publisher,
	}
}

// GetOrders получает все активные заказы с использованием кеша
func (s *ordersService) GetOrders(ctx context.Context) ([]Order, error) {
	// Сначала пытаемся получить из кеша
	cachedOrders, err := s.cache.GetOrders()
	if err != nil {
		logrus.WithError(err).Warn("Failed to get orders from cache, falling back to database")
	} else if cachedOrders != nil {
		logrus.Info("Returning orders from cache")
		return cachedOrders, nil
	}

	// Если в кеше нет данных, получаем из базы данных
	orders, err := s.orderRepo.FindActive(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get orders from database")
		return nil, fmt.Errorf("failed to get orders from database: %w", err)
	}

	// Форматируем данные для отображения
	now := time.Now()
	for i := range orders {
		orders[i].CreatedAtFormatted = orders[i].CreatedAt.Format("2006-01-02 15:04:05")
		orders[i].TimeAgo = formatTimeAgo(now.Sub(orders[i].CreatedAt))

		// Получаем location из первой позиции заказа
		if len(orders[i].Items) > 0 {
			part, err := s.partRepo.FindByID(ctx, orders[i].Items[0].PartID)
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
func (s *ordersService) CreateOrder(ctx context.Context, req CreateOrderRequest, userID uint, userName string) (*Order, error) {
	if req.BuyerNumber == "" {
		return nil, ValidationError{Field: "buyer_number", Message: "buyer number is required"}
	}

	if len(req.Items) == 0 {
		return nil, ValidationError{Field: "items", Message: "at least one part must be selected"}
	}

	// Создаем заказ
	order := &Order{
		CustomerID:  req.CustomerID,
		SellerID:    userID,
		Seller:      userName,
		Part:        req.Part,
		PartID:      req.PartID,
		BuyerNumber: req.BuyerNumber,
		Status:      "red",
		StatusText:  "Need to order transport company",
		CreatedAt:   time.Now(),
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		logrus.WithError(err).Error("Failed to create order in database")
		return nil, fmt.Errorf("failed to create order: %w", err)
	}

	// Создаем позиции заказа и уменьшаем количество запчастей
	for _, item := range req.Items {
		part, err := s.partRepo.FindByID(ctx, item.PartID)
		if err != nil {
			logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to get part for order item")
			return nil, fmt.Errorf("failed to get part information for part %d: %w", item.PartID, err)
		}

		orderItem := &OrderItem{
			OrderID:  order.ID,
			PartID:   item.PartID,
			Quantity: item.Quantity,
			Price:    part.Price,
		}

		if err := s.orderRepo.CreateItem(ctx, orderItem); err != nil {
			logrus.WithError(err).Error("Failed to create order item in database")
			return nil, fmt.Errorf("failed to create order item: %w", err)
		}

		// Уменьшаем количество запчасти
		if err := s.partRepo.DecreaseQuantity(ctx, item.PartID, item.Quantity); err != nil {
			logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to decrease part quantity")
			return nil, fmt.Errorf("failed to decrease part quantity for part %d: %w", item.PartID, err)
		}

		logrus.WithFields(logrus.Fields{
			"part_id":  item.PartID,
			"quantity": item.Quantity,
		}).Info("Decreased part quantity for order")
	}

	// Загружаем полный заказ с позициями
	completeOrder, err := s.orderRepo.FindByID(ctx, order.ID)
	if err != nil {
		logrus.WithError(err).WithField("order_id", order.ID).Error("Failed to load created order")
		return nil, fmt.Errorf("failed to load created order %d: %w", order.ID, err)
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
func (s *ordersService) UpdateOrderStatus(ctx context.Context, orderID uint, status string) error {
	validStatuses := map[string]bool{
		"red":    true,
		"brown":  true,
		"yellow": true,
		"green":  true,
	}

	if !validStatuses[status] {
		return ValidationError{Field: "status", Message: fmt.Sprintf("invalid status: %s", status)}
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, status); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to update order status in database")
		return fmt.Errorf("failed to update order status for order %d: %w", orderID, err)
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
func (s *ordersService) CompleteOrder(ctx context.Context, orderID uint) error {
	logrus.WithField("order_id", orderID).Info("Starting order completion")

	// Получаем заказ с позициями
	order, err := s.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to find order for completion")
		return fmt.Errorf("failed to find order %d for completion: %w", orderID, err)
	}

	// Меняем статус
	err = s.UpdateOrderStatus(ctx, orderID, "green")
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to complete order")
		return fmt.Errorf("failed to complete order %d: %w", orderID, err)
	}

	// Рассчитываем сумму заказа
	var totalAmount float64
	for _, item := range order.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	// Удаляем запчасти полностью
	for _, item := range order.Items {
		if err := s.partRepo.DeletePart(ctx, item.PartID); err != nil {
			logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to delete part after order completion")
			// Продолжаем, не прерываем
		} else {
			logrus.WithField("part_id", item.PartID).Info("Part deleted after order completion")
		}
	}

	// Добавляем запись в историю продаж
	month := time.Now().Format("2006-01")
	salesHistory := &SalesHistory{
		Month: month,
		Sales: totalAmount,
	}
	if err := s.orderRepo.CreateSalesHistory(ctx, salesHistory); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to create sales history")
		// Продолжаем, не прерываем
	} else {
		logrus.WithFields(logrus.Fields{
			"order_id": orderID,
			"month":    month,
			"sales":    totalAmount,
		}).Info("Sales history created")
	}

	// Помечаем заказ как автоматически удаленный (для статистики)
	if err := s.orderRepo.Update(ctx, orderID, map[string]interface{}{"auto_deleted": true}); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to mark order as auto-deleted")
		return err
	}

	// Публикуем событие завершения заказа для обновления статистики
	if err := s.publisher.PublishOrderCompleted(ctx, orderID, totalAmount); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Warn("Failed to publish order completed event")
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
func (s *ordersService) DeleteOrder(ctx context.Context, orderID uint) error {
	logrus.WithField("order_id", orderID).Info("Starting order deletion")
	// Получаем заказ с позициями перед удалением
	orderWithItems, err := s.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to find order with items for deletion")
		return NotFoundError{Resource: "order", ID: orderID}
	}
	logrus.WithField("order_id", orderID).WithField("items_count", len(orderWithItems.Items)).Info("Found order with items for deletion")

	// Возвращаем запчасти в инвентарь
	for _, item := range orderWithItems.Items {
		if err := s.partRepo.IncreaseQuantity(ctx, item.PartID, item.Quantity); err != nil {
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
	if err := s.orderRepo.DeleteItemsByOrderID(ctx, orderID); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to delete order items from database")
		return fmt.Errorf("failed to delete order items for order %d: %w", orderID, err)
	}

	logrus.WithField("order_id", orderID).Info("Attempting to delete order from database")
	if err := s.orderRepo.Delete(ctx, orderID); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to delete order from database")
		return fmt.Errorf("failed to delete order %d: %w", orderID, err)
	}

	// Инвалидируем кеш заказов
	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after deleting order")
	}

	logrus.WithField("order_id", orderID).Info("Order deleted and parts remain deducted")
	return nil
}

// AddOrderItem добавляет позицию в существующий заказ
func (s *ordersService) AddOrderItem(ctx context.Context, orderID uint, req AddOrderItemRequest) error {
	// Проверяем существование заказа
	_, err := s.orderRepo.FindByID(ctx, orderID)
	if err != nil {
		return NotFoundError{Resource: "order", ID: orderID}
	}

	// Проверяем существование запчасти
	part, err := s.partRepo.FindByID(ctx, req.PartID)
	if err != nil {
		return NotFoundError{Resource: "part", ID: req.PartID}
	}

	// Проверяем, есть ли уже такая позиция в заказе
	existingItem, err := s.orderRepo.FindOrderItem(ctx, orderID, req.PartID)
	if err == nil {
		// Обновляем количество существующей позиции
		existingItem.Quantity += req.Quantity
		if err := s.orderRepo.UpdateItem(ctx, existingItem); err != nil {
			logrus.WithError(err).Error("Failed to update order item in database")
			return fmt.Errorf("failed to update order item: %w", err)
		}
	} else {
		// Создаем новую позицию
		orderItem := &OrderItem{
			OrderID:  orderID,
			PartID:   req.PartID,
			Quantity: req.Quantity,
			Price:    part.Price,
		}

		if err := s.orderRepo.CreateItem(ctx, orderItem); err != nil {
			logrus.WithError(err).Error("Failed to create order item in database")
			return fmt.Errorf("failed to create order item: %w", err)
		}
	}

	// Уменьшаем количество запчасти
	if err := s.partRepo.DecreaseQuantity(ctx, req.PartID, req.Quantity); err != nil {
		logrus.WithError(err).WithField("part_id", req.PartID).Error("Failed to decrease part quantity")
		return fmt.Errorf("failed to decrease part quantity for part %d: %w", req.PartID, err)
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

// GetMonthlySales получает продажи по месяцам
func (s *ordersService) GetMonthlySales(ctx context.Context) ([]MonthlySales, error) {
	return s.orderRepo.GetMonthlySales(ctx)
}
