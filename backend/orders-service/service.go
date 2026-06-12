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
	CreateOrder(ctx context.Context, req CreateOrderRequest, userID int64, userName string) (*Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	CompleteOrder(ctx context.Context, orderID int64) error
	DeleteOrder(ctx context.Context, orderID int64) error
	AddOrderItem(ctx context.Context, orderID int64, req AddOrderItemRequest) error
	GetMonthlySales(ctx context.Context) ([]MonthlySales, error)
}

// CreateOrderRequest запрос на создание заказа
type CreateOrderRequest struct {
	CustomerID  int64  `json:"customer_id"`
	Part        string `json:"part"`
	PartID      int64  `json:"part_id"`
	BuyerNumber string `json:"buyer_number"`
	Items       []struct {
		PartID   int64 `json:"part_id"`
		Quantity int   `json:"quantity"`
	} `json:"items"`
}

// AddOrderItemRequest запрос на добавление позиции в заказ
type AddOrderItemRequest struct {
	PartID   int64 `json:"part_id"`
	Quantity int   `json:"quantity"`
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
	cachedOrders, err := s.cache.GetOrders()
	if err != nil {
		logrus.WithError(err).Warn("Failed to get orders from cache, falling back to database")
	} else if cachedOrders != nil {
		logrus.Info("Returning orders from cache")
		return cachedOrders, nil
	}

	orders, err := s.orderRepo.FindActive(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get orders from database")
		return nil, fmt.Errorf("failed to get orders from database: %w", err)
	}

	now := time.Now()
	for i := range orders {
		orders[i].CreatedAtFormatted = orders[i].CreatedAt.Format("2006-01-02 15:04:05")
		orders[i].TimeAgo = formatTimeAgo(now.Sub(orders[i].CreatedAt))

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

	if err := s.cache.SetOrders(orders); err != nil {
		logrus.WithError(err).Warn("Failed to cache orders")
	}

	return orders, nil
}

// CreateOrder создает новый заказ (выполняется в транзакции)
func (s *ordersService) CreateOrder(ctx context.Context, req CreateOrderRequest, userID int64, userName string) (*Order, error) {
	if req.BuyerNumber == "" {
		return nil, ValidationError{Field: "buyer_number", Message: "buyer number is required"}
	}

	if len(req.Items) == 0 {
		return nil, ValidationError{Field: "items", Message: "at least one part must be selected"}
	}

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

	var completeOrder *Order
	err := RunInTransaction(ctx, s.orderRepo.GetPool(), func(txCtx context.Context) error {
		if err := s.orderRepo.Create(txCtx, order); err != nil {
			logrus.WithError(err).Error("Failed to create order in database")
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, item := range req.Items {
			part, err := s.partRepo.FindByID(txCtx, item.PartID)
			if err != nil {
				logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to get part for order item")
				return fmt.Errorf("failed to get part information for part %d: %w", item.PartID, err)
			}

			orderItem := &OrderItem{
				OrderID:  order.ID,
				PartID:   item.PartID,
				Quantity: item.Quantity,
				Price:    part.Price,
			}

			if err := s.orderRepo.CreateItem(txCtx, orderItem); err != nil {
				logrus.WithError(err).Error("Failed to create order item in database")
				return fmt.Errorf("failed to create order item: %w", err)
			}

			if err := s.partRepo.DecreaseQuantity(txCtx, item.PartID, item.Quantity); err != nil {
				logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to decrease part quantity")
				return fmt.Errorf("failed to decrease part quantity for part %d: %w", item.PartID, err)
			}

			logrus.WithFields(logrus.Fields{
				"part_id":  item.PartID,
				"quantity": item.Quantity,
			}).Info("Decreased part quantity for order")
		}

		var errLoad error
		completeOrder, errLoad = s.orderRepo.FindWithItemsByID(txCtx, order.ID)
		return errLoad
	})

	if err != nil {
		return nil, err
	}

	now := time.Now()
	completeOrder.CreatedAtFormatted = completeOrder.CreatedAt.Format("2006-01-02 15:04:05")
	completeOrder.TimeAgo = formatTimeAgo(now.Sub(completeOrder.CreatedAt))

	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after creating order")
	}

	return completeOrder, nil
}

// UpdateOrderStatus обновляет статус заказа
func (s *ordersService) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
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

	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after updating status")
	}

	return nil
}

// CompleteOrder завершает заказ (переводит в статус green, удаляет запчасти и пишет в историю продаж в одной транзакции)
func (s *ordersService) CompleteOrder(ctx context.Context, orderID int64) error {
	logrus.WithField("order_id", orderID).Info("Starting order completion")

	var totalAmount float64
	err := RunInTransaction(ctx, s.orderRepo.GetPool(), func(txCtx context.Context) error {
		order, err := s.orderRepo.FindWithItemsByID(txCtx, orderID)
		if err != nil {
			logrus.WithError(err).WithField("order_id", orderID).Error("Failed to find order for completion")
			return fmt.Errorf("failed to find order %d for completion: %w", orderID, err)
		}

		if err := s.orderRepo.UpdateStatus(txCtx, orderID, "green"); err != nil {
			logrus.WithError(err).WithField("order_id", orderID).Error("Failed to update status to green")
			return fmt.Errorf("failed to complete order %d: %w", orderID, err)
		}

		for _, item := range order.Items {
			totalAmount += item.Price * float64(item.Quantity)
			if err := s.partRepo.DeletePart(txCtx, item.PartID); err != nil {
				logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to delete part after order completion")
				return fmt.Errorf("failed to delete part %d: %w", item.PartID, err)
			}
			logrus.WithField("part_id", item.PartID).Info("Part deleted after order completion")
		}

		month := time.Now().Format("2006-01")
		salesHistory := &SalesHistory{
			Month:     month,
			Sales:     totalAmount,
			CreatedAt: time.Now(),
		}
		if err := s.orderRepo.CreateSalesHistory(txCtx, salesHistory); err != nil {
			logrus.WithError(err).WithField("order_id", orderID).Error("Failed to create sales history")
			return fmt.Errorf("failed to create sales history: %w", err)
		}

		if err := s.orderRepo.Update(txCtx, orderID, map[string]interface{}{"auto_deleted": true}); err != nil {
			logrus.WithError(err).WithField("order_id", orderID).Error("Failed to mark order as auto-deleted")
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	if err := s.publisher.PublishOrderCompleted(ctx, orderID, totalAmount); err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Warn("Failed to publish order completed event")
	}

	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after completing order")
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"amount":   totalAmount,
	}).Info("Order completed successfully, order and parts deleted")

	return nil
}

// DeleteOrder удаляет заказ и возвращает запчасти в инвентарь (в транзакции)
func (s *ordersService) DeleteOrder(ctx context.Context, orderID int64) error {
	logrus.WithField("order_id", orderID).Info("Starting order deletion")

	err := RunInTransaction(ctx, s.orderRepo.GetPool(), func(txCtx context.Context) error {
		orderWithItems, err := s.orderRepo.FindWithItemsByID(txCtx, orderID)
		if err != nil {
			logrus.WithError(err).WithField("order_id", orderID).Error("Failed to find order with items for deletion")
			return NotFoundError{Resource: "order", ID: orderID}
		}
		logrus.WithField("order_id", orderID).WithField("items_count", len(orderWithItems.Items)).Info("Found order with items for deletion")

		for _, item := range orderWithItems.Items {
			if err := s.partRepo.IncreaseQuantity(txCtx, item.PartID, item.Quantity); err != nil {
				logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to return part quantity to inventory")
				return fmt.Errorf("failed to restore part quantity for part %d: %w", item.PartID, err)
			}
			logrus.WithFields(logrus.Fields{
				"part_id":  item.PartID,
				"quantity": item.Quantity,
			}).Info("Returning part quantity to inventory")
		}

		logrus.WithField("order_id", orderID).Info("Deleting order items")
		if err := s.orderRepo.DeleteItemsByOrderID(txCtx, orderID); err != nil {
			logrus.WithError(err).WithField("order_id", orderID).Error("Failed to delete order items from database")
			return fmt.Errorf("failed to delete order items for order %d: %w", orderID, err)
		}

		logrus.WithField("order_id", orderID).Info("Attempting to delete order from database")
		if err := s.orderRepo.Delete(txCtx, orderID); err != nil {
			logrus.WithError(err).WithField("order_id", orderID).Error("Failed to delete order from database")
			return fmt.Errorf("failed to delete order %d: %w", orderID, err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after deleting order")
	}

	logrus.WithField("order_id", orderID).Info("Order deleted and parts restored successfully")
	return nil
}

// AddOrderItem добавляет позицию в существующий заказ (в транзакции)
func (s *ordersService) AddOrderItem(ctx context.Context, orderID int64, req AddOrderItemRequest) error {
	err := RunInTransaction(ctx, s.orderRepo.GetPool(), func(txCtx context.Context) error {
		_, err := s.orderRepo.FindByID(txCtx, orderID)
		if err != nil {
			return NotFoundError{Resource: "order", ID: orderID}
		}

		part, err := s.partRepo.FindByID(txCtx, req.PartID)
		if err != nil {
			return NotFoundError{Resource: "part", ID: req.PartID}
		}

		existingItem, err := s.orderRepo.FindOrderItem(txCtx, orderID, req.PartID)
		if err == nil {
			existingItem.Quantity += req.Quantity
			if err := s.orderRepo.UpdateItem(txCtx, existingItem); err != nil {
				logrus.WithError(err).Error("Failed to update order item in database")
				return fmt.Errorf("failed to update order item: %w", err)
			}
		} else {
			orderItem := &OrderItem{
				OrderID:  orderID,
				PartID:   req.PartID,
				Quantity: req.Quantity,
				Price:    part.Price,
			}

			if err := s.orderRepo.CreateItem(txCtx, orderItem); err != nil {
				logrus.WithError(err).Error("Failed to create order item in database")
				return fmt.Errorf("failed to create order item: %w", err)
			}
		}

		if err := s.partRepo.DecreaseQuantity(txCtx, req.PartID, req.Quantity); err != nil {
			logrus.WithError(err).WithField("part_id", req.PartID).Error("Failed to decrease part quantity")
			return fmt.Errorf("failed to decrease part quantity for part %d: %w", req.PartID, err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"part_id":  req.PartID,
		"quantity": req.Quantity,
	}).Info("Added item to order")

	if err := s.cache.InvalidateOrders(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate orders cache after adding item")
	}

	return nil
}

func formatTimeAgo(duration time.Duration) string {
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
