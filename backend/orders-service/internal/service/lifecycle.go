package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"strings"
	"time"

	"avtoplaneta/pkg/logger"
	"orders-service/internal/domain"
	"orders-service/internal/repository"
)

// GetOrders получает все активные заказы с использованием кеша
func (s *ordersService) GetOrders(ctx context.Context) ([]domain.Order, error) {
	if s.cache != nil {
		cachedOrders, err := s.cache.GetOrders()
		if err != nil {
			slog.WarnContext(ctx, "Failed to get orders from cache, falling back to database", logger.Err(err))
		} else if cachedOrders != nil {
			slog.InfoContext(ctx, "Returning orders from cache")
			return cachedOrders, nil
		}
	}

	orders, err := s.orderRepo.FindActive(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get orders from database", logger.Err(err))
		return nil, fmt.Errorf("failed to get orders from database: %w", err)
	}

	s.enrichOrders(ctx, orders, true)

	if s.cache != nil {
		if err := s.cache.SetOrders(orders); err != nil {
			slog.WarnContext(ctx, "Failed to cache orders", logger.Err(err))
		}
	}

	return orders, nil
}

// GetCompletedOrders получает завершённые заказы (историю продаж)
func (s *ordersService) GetCompletedOrders(ctx context.Context) ([]domain.Order, error) {
	orders, err := s.orderRepo.FindCompleted(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get completed orders from database", logger.Err(err))
		return nil, fmt.Errorf("failed to get completed orders from database: %w", err)
	}

	s.enrichOrders(ctx, orders, false)
	return orders, nil
}

// CreateOrder создает новый заказ
func (s *ordersService) CreateOrder(ctx context.Context, req domain.CreateOrderRequest, userID int64, userName string) (*domain.Order, error) {
	if len(req.Items) == 0 {
		return nil, domain.ValidationError{Field: "items", Message: "at least one part must be selected"}
	}

	buyerNumber := strings.TrimSpace(req.BuyerNumber)
	if buyerNumber == "" {
		if req.QuickSale {
			buyerNumber = "Самовывоз"
		} else {
			return nil, domain.ValidationError{Field: "buyer_number", Message: "buyer number is required"}
		}
	}

	type resolvedItem struct {
		partID   int64
		quantity int
		price    float64
		name     string
		location string
	}

	resolvedItems := make([]resolvedItem, 0, len(req.Items))
	var locations []string
	seenLoc := make(map[string]bool)
	var partNames []string

	for _, item := range req.Items {
		if item.Quantity <= 0 {
			item.Quantity = 1
		}
		part, err := s.partRepo.FindByID(ctx, item.PartID)
		if err != nil {
			slog.ErrorContext(ctx, "Failed to get part for order item", logger.Err(err), slog.Int64("part_id", item.PartID))
			return nil, fmt.Errorf("failed to get part information for part %d: %w", item.PartID, err)
		}

		if part.Quantity <= 0 {
			return nil, domain.ValidationError{
				Field:   "quantity",
				Message: fmt.Sprintf("деталь '%s' (ID %d) не может быть оформлена в заказ: количество на складе %d (нет в наличии)", part.Name, part.ID, part.Quantity),
			}
		}

		itemPrice := part.Price
		if item.Price != nil && *item.Price >= 0 {
			itemPrice = *item.Price
		}

		resolvedItems = append(resolvedItems, resolvedItem{
			partID:   item.PartID,
			quantity: item.Quantity,
			price:    itemPrice,
			name:     part.Name,
			location: part.Location,
		})

		if part.Location != "" && !seenLoc[part.Location] {
			seenLoc[part.Location] = true
			locations = append(locations, part.Location)
		}
		if part.Name != "" {
			partNames = append(partNames, part.Name)
		}
	}

	primaryPartID := req.PartID
	if primaryPartID == 0 && len(resolvedItems) > 0 {
		primaryPartID = resolvedItems[0].partID
	}

	partDisplay := strings.TrimSpace(req.Part)
	if partDisplay == "" && len(partNames) > 0 {
		partDisplay = strings.Join(partNames, ", ")
	}

	paymentStatus := strings.TrimSpace(req.PaymentStatus)
	if paymentStatus == "" {
		if req.QuickSale {
			paymentStatus = "paid"
		} else {
			paymentStatus = "unpaid"
		}
	}

	warehouseStatus := strings.TrimSpace(req.WarehouseStatus)
	if warehouseStatus == "" {
		if req.QuickSale {
			warehouseStatus = "ready"
		} else {
			warehouseStatus = "inspecting"
		}
	}

	deliveryMethod := strings.TrimSpace(req.DeliveryMethod)
	if deliveryMethod == "" {
		deliveryMethod = "pickup"
	}

	source := strings.TrimSpace(req.Source)
	if source == "" {
		if req.QuickSale {
			source = "pickup"
		} else {
			source = "drom"
		}
	}

	initialStatus := "Принят в обработку"
	if paymentStatus == "unpaid" && !req.QuickSale {
		initialStatus = "Ожидает предоплаты"
	}

	customerID := s.resolveCustomerForOrder(ctx, req, buyerNumber)

	now := time.Now()
	order := &domain.Order{
		CustomerID:       customerID,
		OrderNumber:      strings.TrimSpace(req.OrderNumber),
		Source:           source,
		SellerID:         userID,
		Seller:           userName,
		Part:             partDisplay,
		PartID:           primaryPartID,
		Location:         strings.Join(locations, ", "),
		BuyerNumber:      buyerNumber,
		Status:           initialStatus,
		StatusText:       initialStatus,
		PaymentStatus:    paymentStatus,
		WarehouseStatus:  warehouseStatus,
		DeliveryMethod:   deliveryMethod,
		TransportCompany: strings.TrimSpace(req.TransportCompany),
		TrackingNumber:   strings.TrimSpace(req.TrackingNumber),
		Notes:            strings.TrimSpace(req.Notes),
		Discount:         math.Max(req.Discount, 0),
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	var completeOrder *domain.Order
	err := repository.RunInTransaction(ctx, s.orderRepo.GetPool(), func(txCtx context.Context) error {
		if err := s.orderRepo.Create(txCtx, order); err != nil {
			slog.ErrorContext(ctx, "Failed to create order in database", logger.Err(err))
			return fmt.Errorf("failed to create order: %w", err)
		}

		for _, rItem := range resolvedItems {
			orderItem := &domain.OrderItem{
				OrderID:          order.ID,
				PartID:           rItem.partID,
				PartName:         rItem.name,
				PartNameSnapshot: rItem.name,
				Quantity:         rItem.quantity,
				Price:            rItem.price,
			}

			if err := s.orderRepo.CreateItem(txCtx, orderItem); err != nil {
				slog.ErrorContext(ctx, "Failed to create order item in database", logger.Err(err))
				return fmt.Errorf("failed to create order item: %w", err)
			}
		}

		var errLoad error
		completeOrder, errLoad = s.orderRepo.FindWithItemsByID(txCtx, order.ID)
		return errLoad
	})
	if err != nil {
		return nil, err
	}

	// Если быстрая продажа, сразу завершаем заказ
	if req.QuickSale {
		if err := s.CompleteOrder(ctx, completeOrder.ID); err != nil {
			return nil, fmt.Errorf("order created (#%d), but failed to complete quick sale: %w", completeOrder.ID, err)
		}
		if updated, errLoad := s.orderRepo.FindWithItemsByID(ctx, completeOrder.ID); errLoad == nil {
			completeOrder = updated
		}
	}

	ordersSlice := []domain.Order{*completeOrder}
	s.enrichOrders(ctx, ordersSlice, !req.QuickSale)
	*completeOrder = ordersSlice[0]

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			slog.WarnContext(ctx, "Failed to invalidate orders cache after creating order", logger.Err(err))
		}
	}

	return completeOrder, nil
}

// UpdateOrderStatus обновляет статус заказа
func (s *ordersService) UpdateOrderStatus(ctx context.Context, orderID int64, status string) error {
	validStatuses := map[string]bool{
		"yellow":                      true,
		"green":                       true,
		"red":                         true,
		"Принят в обработку":          true,
		"Ожидает предоплаты":          true,
		"Оплачен":                     true,
		"На фотофиксации":             true,
		"Перемещение между складами":  true,
		"Проверен":                    true,
		"Упаковывается":               true,
		"Готов к выдаче":              true,
		"Отправлен":                   true,
		"Выдан клиенту":               true,
		"Возврат":                     true,
		"Отменен":                     true,
	}

	if !validStatuses[status] {
		return domain.ValidationError{Field: "status", Message: "invalid status"}
	}

	updates := map[string]interface{}{
		"status":      status,
		"status_text": status,
	}

	switch status {
	case "Оплачен":
		updates["payment_status"] = "paid"
	case "Ожидает предоплаты":
		updates["payment_status"] = "unpaid"
	case "На фотофиксации":
		updates["warehouse_status"] = "inspecting"
	case "Перемещение между складами":
		updates["warehouse_status"] = "transfer"
	case "Проверен":
		updates["warehouse_status"] = "ready"
	case "Упаковывается":
		updates["warehouse_status"] = "packing"
	case "Готов к выдаче":
		updates["warehouse_status"] = "ready"
		updates["payment_status"] = "paid"
	case "Отправлен":
		updates["payment_status"] = "paid"
	case "Выдан клиенту":
		updates["payment_status"] = "paid"
	}

	if err := s.orderRepo.Update(ctx, orderID, updates); err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			slog.WarnContext(ctx, "Failed to invalidate orders cache after updating status", logger.Err(err))
		}
	}

	return nil
}

// UpdateOrderDetails обновляет реквизиты сделки
func (s *ordersService) UpdateOrderDetails(ctx context.Context, orderID int64, req domain.UpdateOrderDetailsRequest) (*domain.Order, error) {
	if _, err := s.orderRepo.FindByID(ctx, orderID); err != nil {
		return nil, domain.NotFoundError{Resource: "order", ID: orderID}
	}

	updates := make(map[string]interface{})
	if req.BuyerNumber != nil {
		updates["buyer_number"] = strings.TrimSpace(*req.BuyerNumber)
	}
	if req.OrderNumber != nil {
		updates["order_number"] = strings.TrimSpace(*req.OrderNumber)
	}
	if req.Source != nil {
		updates["source"] = strings.TrimSpace(*req.Source)
	}
	if req.Status != nil && strings.TrimSpace(*req.Status) != "" {
		st := strings.TrimSpace(*req.Status)
		updates["status"] = st
		updates["status_text"] = st
	}
	if req.PaymentStatus != nil {
		updates["payment_status"] = strings.TrimSpace(*req.PaymentStatus)
	}
	if req.WarehouseStatus != nil {
		ws := strings.TrimSpace(*req.WarehouseStatus)
		updates["warehouse_status"] = ws
		if req.Status == nil {
			switch ws {
			case "inspecting":
				updates["status"] = "На фотофиксации"
				updates["status_text"] = "На фотофиксации"
			case "transfer":
				updates["status"] = "Перемещение между складами"
				updates["status_text"] = "Перемещение между складами"
			case "ready":
				updates["status"] = "Проверен"
				updates["status_text"] = "Проверен"
			}
		}
	}
	if req.DeliveryMethod != nil {
		updates["delivery_method"] = strings.TrimSpace(*req.DeliveryMethod)
	}
	if req.TransportCompany != nil {
		updates["transport_company"] = strings.TrimSpace(*req.TransportCompany)
	}
	if req.TrackingNumber != nil {
		tn := strings.TrimSpace(*req.TrackingNumber)
		updates["tracking_number"] = tn
		if tn != "" && req.Status == nil {
			updates["status"] = "Отправлен"
			updates["status_text"] = "Отправлен"
		}
	}
	if req.Notes != nil {
		updates["notes"] = strings.TrimSpace(*req.Notes)
	}
	if req.Discount != nil {
		updates["discount"] = math.Max(*req.Discount, 0)
	}

	if err := s.orderRepo.Update(ctx, orderID, updates); err != nil {
		return nil, err
	}

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			slog.WarnContext(ctx, "Failed to invalidate orders cache after updating order details", logger.Err(err))
		}
	}

	updated, err := s.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		return nil, err
	}
	slice := []domain.Order{*updated}
	s.enrichOrders(ctx, slice, !updated.AutoDeleted)
	return &slice[0], nil
}

// CompleteOrder завершает заказ и списывает остатки
func (s *ordersService) CompleteOrder(ctx context.Context, orderID int64) error {
	slog.InfoContext(ctx, "Starting order completion", slog.Int64("order_id", orderID))

	order, err := s.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to find order for completion", logger.Err(err), slog.Int64("order_id", orderID))
		return domain.NotFoundError{Resource: "order", ID: orderID}
	}

	order.RecalculateTotal()
	totalAmount := order.TotalAmount

	// 1. Атомарно завершаем заказ и создаем Outbox-событие в единой транзакции
	err = repository.RunInTransaction(ctx, s.orderRepo.GetPool(), func(txCtx context.Context) error {
		if err := s.orderRepo.CompleteRecord(txCtx, orderID); err != nil {
			return err
		}

		month := time.Now().Format("2006-01")
		salesHistory := &domain.SalesHistory{
			Month:     month,
			Sales:     totalAmount,
			CreatedAt: time.Now(),
		}
		if err := s.orderRepo.CreateSalesHistory(txCtx, salesHistory); err != nil {
			slog.WarnContext(ctx, "Failed to create sales history entry", logger.Err(err), slog.Int64("order_id", orderID))
		}

		payload, _ := json.Marshal(map[string]interface{}{
			"order_id": orderID,
			"amount":   totalAmount,
		})
		return s.orderRepo.CreateOutboxEvent(txCtx, "order", orderID, "order_completed", payload)
	})
	if err != nil {
		return err
	}

	// 2. Быстрый путь (Fast-path): списываем остатки и отправляем событие сразу,
	// а Outbox-диспетчер гарантирует повторы и доставку в случае любых сетевых сбоев
	for _, item := range order.Items {
		if item.PartID <= 0 {
			continue
		}

		part, errPart := s.partRepo.FindByID(ctx, item.PartID)
		if errPart != nil {
			slog.WarnContext(ctx, "Part not found during order completion (may already be deleted)", logger.Err(errPart), slog.Int64("part_id", item.PartID))
			continue
		}

		if part.Quantity > item.Quantity {
			opID := fmt.Sprintf("complete-order-%d-part-%d", orderID, item.PartID)
			if err := s.partRepo.DecreaseQuantity(ctx, item.PartID, item.Quantity, opID); err != nil {
				slog.WarnContext(ctx, "Fast-path decrease failed, Outbox Dispatcher will retry", logger.Err(err), slog.Int64("part_id", item.PartID))
			}
		} else {
			if err := s.partRepo.DeletePart(ctx, item.PartID); err != nil {
				slog.WarnContext(ctx, "Fast-path delete failed, Outbox Dispatcher will retry", logger.Err(err), slog.Int64("part_id", item.PartID))
			}
		}
	}

	if s.publisher != nil {
		if err := s.publisher.PublishOrderCompleted(ctx, orderID, totalAmount); err != nil {
			slog.WarnContext(ctx, "Fast-path publish failed, Outbox Dispatcher will retry", logger.Err(err), slog.Int64("order_id", orderID))
		}
	}

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			slog.WarnContext(ctx, "Failed to invalidate orders cache after completing order", logger.Err(err))
		}
	}

	slog.InfoContext(ctx, "Order completed successfully",
		slog.Int64("order_id", orderID),
		slog.Float64("amount", totalAmount))

	return nil
}

// DeleteOrder удаляет заказ
func (s *ordersService) DeleteOrder(ctx context.Context, orderID int64) error {
	slog.InfoContext(ctx, "Starting order deletion", slog.Int64("order_id", orderID))

	err := repository.RunInTransaction(ctx, s.orderRepo.GetPool(), func(txCtx context.Context) error {
		if _, err := s.orderRepo.FindByID(txCtx, orderID); err != nil {
			return domain.NotFoundError{Resource: "order", ID: orderID}
		}

		if err := s.orderRepo.DeleteItemsByOrderID(txCtx, orderID); err != nil {
			return fmt.Errorf("failed to delete order items for order %d: %w", orderID, err)
		}

		if err := s.orderRepo.Delete(txCtx, orderID); err != nil {
			return fmt.Errorf("failed to delete order %d: %w", orderID, err)
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			slog.WarnContext(ctx, "Failed to invalidate orders cache after deleting order", logger.Err(err))
		}
	}

	return nil
}

// GetMonthlySales получает продажи по месяцам
func (s *ordersService) GetMonthlySales(ctx context.Context) ([]domain.MonthlySales, error) {
	return s.orderRepo.GetMonthlySales(ctx)
}
