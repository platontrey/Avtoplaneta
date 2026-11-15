package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"orders-service/internal/domain"
)

func computeOrderTotal(order *domain.Order) {
	var subtotal float64
	for _, item := range order.Items {
		subtotal += item.Price * float64(item.Quantity)
	}
	order.TotalAmount = math.Max(subtotal-order.Discount, 0)
}

func (s *ordersService) newPartCacheLookup(ctx context.Context) func(int64) (*domain.Part, error) {
	partCache := make(map[int64]*domain.Part)
	return func(partID int64) (*domain.Part, error) {
		if cached, ok := partCache[partID]; ok {
			if cached == nil {
				return nil, fmt.Errorf("part %d not found", partID)
			}
			return cached, nil
		}
		part, err := s.partRepo.FindByID(ctx, partID)
		if err != nil {
			partCache[partID] = nil
			return nil, err
		}
		partCache[partID] = part
		return part, nil
	}
}

func enrichOrderDates(order *domain.Order, now time.Time) {
	order.CreatedAtFormatted = order.CreatedAt.Format("2006-01-02 15:04:05")
	order.TimeAgo = formatTimeAgo(now.Sub(order.CreatedAt))
	if order.CompletedAt != nil {
		order.CompletedAtFormatted = order.CompletedAt.Format("2006-01-02 15:04:05")
	}
}

func (s *ordersService) enrichOrderItem(item *domain.OrderItem, lookupPart func(int64) (*domain.Part, error), fetchLiveParts bool, fallbackPart string) string {
	if fetchLiveParts && item.PartID > 0 {
		if part, err := lookupPart(item.PartID); err == nil && part != nil {
			item.Location = part.Location
			if part.Name != "" {
				item.PartName = part.Name
			}
		}
	}

	if item.PartName == "" {
		if item.PartNameSnapshot != "" {
			item.PartName = item.PartNameSnapshot
		} else if fallbackPart != "" {
			item.PartName = fallbackPart
		}
	}

	return item.Location
}

func (s *ordersService) enrichOrderFromItems(order *domain.Order, lookupPart func(int64) (*domain.Part, error), fetchLiveParts bool) {
	if order.PartID == 0 {
		order.PartID = order.Items[0].PartID
	}

	var locations []string
	seenLocations := make(map[string]bool)
	var names []string

	fallbackPart := ""
	if len(order.Items) == 1 {
		fallbackPart = order.Part
	}

	for j := range order.Items {
		loc := s.enrichOrderItem(&order.Items[j], lookupPart, fetchLiveParts, fallbackPart)
		if loc != "" && !seenLocations[loc] {
			seenLocations[loc] = true
			locations = append(locations, loc)
		}
		if order.Items[j].PartName != "" {
			names = append(names, order.Items[j].PartName)
		}
	}

	if len(locations) > 0 {
		order.Location = strings.Join(locations, ", ")
	} else if order.Location == "" {
		order.Location = "Неизвестно"
	}

	if order.Part == "" && len(names) > 0 {
		order.Part = strings.Join(names, ", ")
	}
}

func (s *ordersService) enrichOrderFromSinglePart(order *domain.Order, lookupPart func(int64) (*domain.Part, error), fetchLiveParts bool) {
	if order.PartID > 0 && fetchLiveParts {
		if part, err := lookupPart(order.PartID); err == nil && part != nil {
			order.Location = part.Location
			if order.Part == "" {
				order.Part = part.Name
			}
			return
		}
		if order.Location == "" {
			order.Location = "Неизвестно"
		}
		return
	}

	if order.Location == "" {
		order.Location = "Нет деталей"
	}
}

func (s *ordersService) enrichOrders(ctx context.Context, orders []domain.Order, fetchLiveParts bool) {
	now := time.Now()
	lookupPart := s.newPartCacheLookup(ctx)

	for i := range orders {
		order := &orders[i]
		enrichOrderDates(order, now)

		if len(order.Items) > 0 {
			s.enrichOrderFromItems(order, lookupPart, fetchLiveParts)
		} else {
			s.enrichOrderFromSinglePart(order, lookupPart, fetchLiveParts)
		}

		computeOrderTotal(order)
	}
}

func formatTimeAgo(duration time.Duration) string {
	hours := int(duration.Hours())
	if hours < 1 {
		return "менее часа назад"
	} else if hours < 24 {
		return fmt.Sprintf("%d ч. назад", hours)
	} else {
		days := hours / 24
		return fmt.Sprintf("%d дн. назад", days)
	}
}
