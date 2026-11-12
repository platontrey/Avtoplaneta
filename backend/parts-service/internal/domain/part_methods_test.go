package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"avtoplaneta/pkg/constants"
)

func TestPart_InventoryInvariants(t *testing.T) {
	// 1. Обычная деталь на складе
	partInStock := &Part{
		PartCore: PartCore{
			ID:       1,
			Quantity: 5,
		},
	}
	assert.True(t, partInStock.IsVisible())
	assert.False(t, partInStock.IsSoldOut())
	assert.False(t, partInStock.IsDefective())
	assert.NoError(t, partInStock.CanDecreaseQuantity(2))
	assert.ErrorIs(t, partInStock.CanDecreaseQuantity(10), ErrInsufficientQuantity)

	// 2. Дефектная ведомость (quantity == 0) — обязана быть видимой!
	partDefective := &Part{
		PartCore: PartCore{
			ID:       2,
			Quantity: constants.PartQuantityDefective,
		},
	}
	assert.True(t, partDefective.IsVisible())
	assert.False(t, partDefective.IsSoldOut())
	assert.True(t, partDefective.IsDefective())

	// 3. Полностью списанная/проданная деталь (quantity == -1) — сокрыта!
	partSold := &Part{
		PartCore: PartCore{
			ID:       3,
			Quantity: constants.PartQuantitySold,
		},
	}
	assert.False(t, partSold.IsVisible())
	assert.True(t, partSold.IsSoldOut())
	assert.False(t, partSold.IsDefective())
	assert.ErrorIs(t, partSold.CanDecreaseQuantity(1), ErrPartAlreadySold)
}
