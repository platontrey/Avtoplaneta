package domain

import (
	"errors"
	"math"
)

var (
	ErrOrderAlreadyCompleted   = errors.New("order is already completed")
	ErrOrderEmpty              = errors.New("order has no items to complete")
	ErrInvalidStatusTransition = errors.New("invalid order status transition")
)

const (
	OrderStatusRed    = "red"
	OrderStatusBrown  = "brown"
	OrderStatusYellow = "yellow"
	OrderStatusGreen  = "green"
)

// RecalculateTotal рассчитывает и обновляет итоговую стоимость заказа с учетом скидки
func (o *Order) RecalculateTotal() {
	var subtotal float64
	for _, item := range o.Items {
		subtotal += item.Price * float64(item.Quantity)
	}
	o.TotalAmount = math.Max(subtotal-o.Discount, 0)
}

// IsCompleted возвращает true, если заказ завершен (green) или помечен как удаленный
func (o *Order) IsCompleted() bool {
	return o.Status == OrderStatusGreen || o.AutoDeleted
}

// CanTransitionTo проверяет допустимость перехода заказа в целевой статус
func (o *Order) CanTransitionTo(targetStatus string) error {
	if o.IsCompleted() {
		return ErrOrderAlreadyCompleted
	}

	switch targetStatus {
	case OrderStatusRed, OrderStatusBrown, OrderStatusYellow, OrderStatusGreen:
		return nil
	default:
		return ErrInvalidStatusTransition
	}
}

// CanBeCompleted проверяет, готов ли заказ к завершению (списанию)
func (o *Order) CanBeCompleted() error {
	if o.IsCompleted() {
		return ErrOrderAlreadyCompleted
	}
	if len(o.Items) == 0 && o.PartID <= 0 {
		return ErrOrderEmpty
	}
	return nil
}
