package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOrder_RecalculateTotal(t *testing.T) {
	order := &Order{
		Discount: 200,
		Items: []OrderItem{
			{Price: 500, Quantity: 2}, // 1000
			{Price: 300, Quantity: 1}, // 300
		},
	}

	order.RecalculateTotal()
	assert.Equal(t, 1100.0, order.TotalAmount)

	// Проверка, что скидка не уводит сумму в минус
	order.Discount = 2000
	order.RecalculateTotal()
	assert.Equal(t, 0.0, order.TotalAmount)
}

func TestOrder_StatusTransitions(t *testing.T) {
	order := &Order{
		Status: OrderStatusRed,
	}

	assert.False(t, order.IsCompleted())
	assert.NoError(t, order.CanTransitionTo(OrderStatusBrown))
	assert.NoError(t, order.CanTransitionTo(OrderStatusYellow))
	assert.NoError(t, order.CanTransitionTo(OrderStatusGreen))
	assert.ErrorIs(t, order.CanTransitionTo("invalid_status"), ErrInvalidStatusTransition)

	// Завершенный заказ нельзя менять
	order.Status = OrderStatusGreen
	assert.True(t, order.IsCompleted())
	assert.ErrorIs(t, order.CanTransitionTo(OrderStatusRed), ErrOrderAlreadyCompleted)
}

func TestOrder_CanBeCompleted(t *testing.T) {
	emptyOrder := &Order{
		Status: OrderStatusYellow,
		Items:  []OrderItem{},
		PartID: 0,
	}
	assert.ErrorIs(t, emptyOrder.CanBeCompleted(), ErrOrderEmpty)

	validOrder := &Order{
		Status: OrderStatusYellow,
		Items: []OrderItem{
			{PartID: 10, Quantity: 1, Price: 100},
		},
	}
	assert.NoError(t, validOrder.CanBeCompleted())
}
