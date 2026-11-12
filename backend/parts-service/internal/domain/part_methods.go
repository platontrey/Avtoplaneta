package domain

import (
	"errors"

	"avtoplaneta/pkg/constants"
)

var (
	ErrInsufficientQuantity = errors.New("insufficient part quantity for decrease operation")
	ErrPartAlreadySold      = errors.New("part is already sold out and depleted")
)

// IsVisible возвращает true, если деталь видима в поисковой выдаче и каталогах (quantity >= 0)
func (p *Part) IsVisible() bool {
	return constants.IsVisiblePart(int32(p.Quantity))
}

// IsSoldOut возвращает true, если деталь списана/продана (quantity == -1)
func (p *Part) IsSoldOut() bool {
	return p.Quantity == constants.PartQuantitySold
}

// IsDefective возвращает true, если позиция из дефектной ведомости (quantity == 0)
func (p *Part) IsDefective() bool {
	return p.Quantity == constants.PartQuantityDefective
}

// CanDecreaseQuantity проверяет, достаточно ли остатка для списания количества amount
func (p *Part) CanDecreaseQuantity(amount int) error {
	if p.IsSoldOut() {
		return ErrPartAlreadySold
	}
	if p.Quantity < amount {
		return ErrInsufficientQuantity
	}
	return nil
}
