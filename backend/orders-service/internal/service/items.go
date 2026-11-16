package service

import (
	"context"
	"fmt"
	"log/slog"

	"avtoplaneta/pkg/logger"
	"orders-service/internal/domain"
	"orders-service/internal/repository"
)

// AddOrderItem добавляет позицию в существующий заказ
func (s *ordersService) AddOrderItem(ctx context.Context, orderID int64, req domain.AddOrderItemRequest) error {
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	part, err := s.partRepo.FindByID(ctx, req.PartID)
	if err != nil {
		return domain.NotFoundError{Resource: "part", ID: req.PartID}
	}

	if part.Quantity <= 0 {
		return domain.ValidationError{
			Field:   "quantity",
			Message: fmt.Sprintf("деталь '%s' (ID %d) не может быть добавлена в заказ: количество на складе %d (нет в наличии)", part.Name, part.ID, part.Quantity),
		}
	}

	if part.Quantity < req.Quantity {
		return domain.ValidationError{
			Field:   "quantity",
			Message: fmt.Sprintf("деталь '%s' (ID %d) не может быть добавлена в заказ: запрошено %d, на складе %d", part.Name, part.ID, req.Quantity, part.Quantity),
		}
	}

	itemPrice := part.Price
	if req.Price != nil && *req.Price >= 0 {
		itemPrice = *req.Price
	}

	err = repository.RunInTransaction(ctx, s.orderRepo.GetPool(), func(txCtx context.Context) error {
		if _, err := s.orderRepo.FindByID(txCtx, orderID); err != nil {
			return domain.NotFoundError{Resource: "order", ID: orderID}
		}

		existingItem, err := s.orderRepo.FindOrderItem(txCtx, orderID, req.PartID)
		if err == nil {
			existingItem.Quantity += req.Quantity
			if req.Price != nil && *req.Price >= 0 {
				existingItem.Price = *req.Price
			}
			if err := s.orderRepo.UpdateItem(txCtx, existingItem); err != nil {
				return fmt.Errorf("failed to update order item: %w", err)
			}
		} else {
			orderItem := &domain.OrderItem{
				OrderID:          orderID,
				PartID:           req.PartID,
				PartName:         part.Name,
				PartNameSnapshot: part.Name,
				Quantity:         req.Quantity,
				Price:            itemPrice,
			}

			if err := s.orderRepo.CreateItem(txCtx, orderItem); err != nil {
				return fmt.Errorf("failed to create order item: %w", err)
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	// Списываем остаток со склада
	opID := fmt.Sprintf("order-%d-add-part-%d", orderID, req.PartID)
	if err := s.partRepo.DecreaseQuantity(ctx, req.PartID, req.Quantity, opID); err != nil {
		slog.WarnContext(ctx, "Failed to decrease quantity when adding item to order", logger.Err(err), slog.Int64("part_id", req.PartID))
	}

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			slog.WarnContext(ctx, "Failed to invalidate orders cache after adding item", logger.Err(err))
		}
	}

	return nil
}

// UpdateOrderItem изменяет количество или цену позиции в заказе
func (s *ordersService) UpdateOrderItem(ctx context.Context, orderID, itemID int64, req domain.UpdateOrderItemRequest) error {
	order, err := s.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		return domain.NotFoundError{Resource: "order", ID: orderID}
	}

	var target *domain.OrderItem
	for i := range order.Items {
		if order.Items[i].ID == itemID {
			target = &order.Items[i]
			break
		}
	}
	if target == nil {
		return domain.NotFoundError{Resource: "order_item", ID: itemID}
	}

	oldQuantity := target.Quantity
	if req.Quantity != nil && *req.Quantity > 0 {
		target.Quantity = *req.Quantity
	}
	if req.Price != nil && *req.Price >= 0 {
		target.Price = *req.Price
	}

	if err := s.orderRepo.UpdateItem(ctx, target); err != nil {
		return err
	}

	if target.PartID > 0 && req.Quantity != nil && *req.Quantity != oldQuantity {
		diff := *req.Quantity - oldQuantity
		if diff > 0 {
			opID := fmt.Sprintf("order-%d-item-%d-inc-qty-%d", orderID, itemID, *req.Quantity)
			if err := s.partRepo.DecreaseQuantity(ctx, target.PartID, diff, opID); err != nil {
				slog.WarnContext(ctx, "Failed to decrease stock on item quantity increase", logger.Err(err))
			}
		} else if diff < 0 {
			opID := fmt.Sprintf("order-%d-item-%d-dec-qty-%d", orderID, itemID, *req.Quantity)
			if err := s.partRepo.IncreaseQuantity(ctx, target.PartID, -diff, opID); err != nil {
				slog.WarnContext(ctx, "Failed to restore stock on item quantity decrease", logger.Err(err))
			}
		}
	}

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			slog.WarnContext(ctx, "Failed to invalidate orders cache after updating item", logger.Err(err))
		}
	}
	return nil
}

// DeleteOrderItem удаляет позицию из заказа и возвращает остаток на склад
func (s *ordersService) DeleteOrderItem(ctx context.Context, orderID, itemID int64) error {
	order, err := s.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		return domain.NotFoundError{Resource: "order", ID: orderID}
	}
	if len(order.Items) <= 1 {
		return domain.ValidationError{Field: "items", Message: "cannot remove the last item of an order; delete the order instead"}
	}

	var targetItem *domain.OrderItem
	for i := range order.Items {
		if order.Items[i].ID == itemID {
			targetItem = &order.Items[i]
			break
		}
	}

	if err := s.orderRepo.DeleteItem(ctx, orderID, itemID); err != nil {
		return err
	}

	// Возвращаем позицию на склад
	if targetItem != nil && targetItem.PartID > 0 {
		opID := fmt.Sprintf("order-%d-del-item-%d-part-%d", orderID, itemID, targetItem.PartID)
		if err := s.partRepo.IncreaseQuantity(ctx, targetItem.PartID, targetItem.Quantity, opID); err != nil {
			slog.WarnContext(ctx, "Failed to restore quantity when deleting item from order", logger.Err(err), slog.Int64("part_id", targetItem.PartID))
		}
	}

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			slog.WarnContext(ctx, "Failed to invalidate orders cache after deleting item", logger.Err(err))
		}
	}
	return nil
}
