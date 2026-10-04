package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"orders-service/internal/domain"
	"orders-service/internal/repository"
)

// OrdersService определяет интерфейс для бизнес-логики управления заказами
type OrdersService interface {
	GetOrders(ctx context.Context) ([]domain.Order, error)
	GetCompletedOrders(ctx context.Context) ([]domain.Order, error)
	CreateOrder(ctx context.Context, req domain.CreateOrderRequest, userID int64, userName string) (*domain.Order, error)
	UpdateOrderStatus(ctx context.Context, orderID int64, status string) error
	UpdateOrderDetails(ctx context.Context, orderID int64, req domain.UpdateOrderDetailsRequest) (*domain.Order, error)
	CompleteOrder(ctx context.Context, orderID int64) error
	DeleteOrder(ctx context.Context, orderID int64) error
	AddOrderItem(ctx context.Context, orderID int64, req domain.AddOrderItemRequest) error
	UpdateOrderItem(ctx context.Context, orderID, itemID int64, req domain.UpdateOrderItemRequest) error
	DeleteOrderItem(ctx context.Context, orderID, itemID int64) error
	GetMonthlySales(ctx context.Context) ([]domain.MonthlySales, error)

	// Customer operations
	ListCustomers(ctx context.Context, category, search string, limit, offset int32) ([]domain.CustomerWithStats, error)
	GetCustomer(ctx context.Context, id int64) (*domain.CustomerDetails, error)
	CreateCustomer(ctx context.Context, req domain.CreateCustomerRequest) (*domain.Customer, error)
	UpdateCustomer(ctx context.Context, id int64, req domain.UpdateCustomerRequest) (*domain.Customer, error)
	DeleteCustomer(ctx context.Context, id int64) error
}

// ordersService реализует OrdersService
type ordersService struct {
	orderRepo repository.OrderRepository
	partRepo  repository.PartRepositoryForOrders
	cache     repository.CacheService
	publisher domain.EventPublisher
}

// NewOrdersService создает новый сервис заказов
func NewOrdersService(orderRepo repository.OrderRepository, partRepo repository.PartRepositoryForOrders, cache repository.CacheService, publisher domain.EventPublisher) OrdersService {
	return &ordersService{
		orderRepo: orderRepo,
		partRepo:  partRepo,
		cache:     cache,
		publisher: publisher,
	}
}

func computeOrderTotal(order *domain.Order) {
	var subtotal float64
	for _, item := range order.Items {
		subtotal += item.Price * float64(item.Quantity)
	}
	order.TotalAmount = math.Max(subtotal-order.Discount, 0)
}

func (s *ordersService) enrichOrders(ctx context.Context, orders []domain.Order, fetchLiveParts bool) {
	now := time.Now()
	partCache := make(map[int64]*domain.Part)
	lookupPart := func(partID int64) (*domain.Part, error) {
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

	for i := range orders {
		orders[i].CreatedAtFormatted = orders[i].CreatedAt.Format("2006-01-02 15:04:05")
		orders[i].TimeAgo = formatTimeAgo(now.Sub(orders[i].CreatedAt))
		if orders[i].CompletedAt != nil {
			orders[i].CompletedAtFormatted = orders[i].CompletedAt.Format("2006-01-02 15:04:05")
		}

		if orders[i].PartID == 0 && len(orders[i].Items) > 0 {
			orders[i].PartID = orders[i].Items[0].PartID
		}

		if len(orders[i].Items) > 0 {
			var locations []string
			seenLocations := make(map[string]bool)

			for j := range orders[i].Items {
				itemPartID := orders[i].Items[j].PartID
				if fetchLiveParts && itemPartID > 0 {
					part, err := lookupPart(itemPartID)
					if err == nil && part != nil {
						orders[i].Items[j].Location = part.Location
						if part.Location != "" && !seenLocations[part.Location] {
							seenLocations[part.Location] = true
							locations = append(locations, part.Location)
						}
						if part.Name != "" {
							orders[i].Items[j].PartName = part.Name
						}
					}
				}

				if orders[i].Items[j].PartName == "" {
					if orders[i].Items[j].PartNameSnapshot != "" {
						orders[i].Items[j].PartName = orders[i].Items[j].PartNameSnapshot
					} else if len(orders[i].Items) == 1 && orders[i].Part != "" {
						orders[i].Items[j].PartName = orders[i].Part
					}
				}
			}

			if len(locations) > 0 {
				orders[i].Location = strings.Join(locations, ", ")
			} else if orders[i].Location == "" {
				orders[i].Location = "Неизвестно"
			}

			if orders[i].Part == "" {
				var names []string
				for _, it := range orders[i].Items {
					if it.PartName != "" {
						names = append(names, it.PartName)
					}
				}
				orders[i].Part = strings.Join(names, ", ")
			}
		} else if orders[i].PartID > 0 && fetchLiveParts {
			if part, err := lookupPart(orders[i].PartID); err == nil && part != nil {
				orders[i].Location = part.Location
				if orders[i].Part == "" {
					orders[i].Part = part.Name
				}
			} else if orders[i].Location == "" {
				orders[i].Location = "Неизвестно"
			}
		} else if orders[i].Location == "" {
			orders[i].Location = "Нет деталей"
		}

		computeOrderTotal(&orders[i])
	}
}

// GetOrders получает все активные заказы с использованием кеша
func (s *ordersService) GetOrders(ctx context.Context) ([]domain.Order, error) {
	if s.cache != nil {
		cachedOrders, err := s.cache.GetOrders()
		if err != nil {
			logrus.WithError(err).Warn("Failed to get orders from cache, falling back to database")
		} else if cachedOrders != nil {
			logrus.Info("Returning orders from cache")
			return cachedOrders, nil
		}
	}

	orders, err := s.orderRepo.FindActive(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get orders from database")
		return nil, fmt.Errorf("failed to get orders from database: %w", err)
	}

	s.enrichOrders(ctx, orders, true)

	if s.cache != nil {
		if err := s.cache.SetOrders(orders); err != nil {
			logrus.WithError(err).Warn("Failed to cache orders")
		}
	}

	return orders, nil
}

// GetCompletedOrders получает завершённые заказы (историю продаж)
func (s *ordersService) GetCompletedOrders(ctx context.Context) ([]domain.Order, error) {
	orders, err := s.orderRepo.FindCompleted(ctx)
	if err != nil {
		logrus.WithError(err).Error("Failed to get completed orders from database")
		return nil, fmt.Errorf("failed to get completed orders from database: %w", err)
	}

	s.enrichOrders(ctx, orders, false)
	return orders, nil
}

// parseBuyerInfo извлекает нормализованный телефон и имя покупателя из строки buyer_number
func parseBuyerInfo(raw string) (phone string, name string) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "Самовывоз" || raw == "Продажа на месте" || raw == "Без контакта" {
		return "", ""
	}

	// Извлекаем все цифры
	var digits strings.Builder
	for _, r := range raw {
		if r >= '0' && r <= '9' {
			digits.WriteRune(r)
		}
	}
	digitStr := digits.String()

	// Если цифр меньше 7, это не телефон (например, короткое число "1" или артикул)
	if len(digitStr) < 7 {
		return "", ""
	}

	// Нормализация российского/международного номера
	if len(digitStr) == 10 {
		phone = "+7" + digitStr
	} else if len(digitStr) == 11 && (digitStr[0] == '7' || digitStr[0] == '8') {
		phone = "+7" + digitStr[1:]
	} else if len(digitStr) >= 11 {
		phone = "+" + digitStr
	} else {
		phone = digitStr
	}

	// Извлекаем имя: убираем префиксы "Получатель", "ФИО:" и сам телефон из имени
	namePart := raw
	lower := strings.ToLower(namePart)
	if strings.HasPrefix(lower, "получатель") {
		namePart = strings.TrimSpace(namePart[len("получатель"):])
	}
	if strings.HasPrefix(strings.ToLower(namePart), "фио") {
		namePart = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(namePart, "ФИО"), ":"))
	}

	words := strings.Fields(namePart)
	var nameWords []string
	for _, w := range words {
		cleanWord := strings.Trim(w, "+()- —–")
		isNumeric := true
		for _, ch := range cleanWord {
			if ch < '0' || ch > '9' {
				isNumeric = false
				break
			}
		}
		if !isNumeric || len(cleanWord) < 6 {
			nameWords = append(nameWords, w)
		}
	}

	name = strings.Join(nameWords, " ")
	name = strings.TrimSpace(name)
	if name == "" {
		name = phone
	}

	return phone, name
}

func (s *ordersService) resolveCustomerForOrder(ctx context.Context, req domain.CreateOrderRequest, buyerNumber string) int64 {
	customerID := req.CustomerID
	if customerID > 0 {
		return customerID
	}

	phone, name := parseBuyerInfo(buyerNumber)
	if phone == "" {
		return 0
	}

	// 1. Ищем существующего клиента по нормализованному номеру и возможным вариациям (8..., 7...)
	existingCustomer, err := s.orderRepo.GetCustomerByPhone(ctx, phone)
	if err != nil && phone != buyerNumber {
		existingCustomer, err = s.orderRepo.GetCustomerByPhone(ctx, buyerNumber)
	}
	if err != nil && strings.HasPrefix(phone, "+7") && len(phone) == 12 {
		existingCustomer, err = s.orderRepo.GetCustomerByPhone(ctx, "8"+phone[2:])
		if err != nil {
			existingCustomer, err = s.orderRepo.GetCustomerByPhone(ctx, phone[1:])
		}
	}

	if err == nil && existingCustomer != nil {
		// Обогащаем имя, если у клиента был только номер
		if (existingCustomer.Name == "" || existingCustomer.Name == existingCustomer.Phone) && name != phone {
			existingCustomer.Name = name
			_ = s.orderRepo.UpdateCustomer(ctx, existingCustomer)
		}
		return existingCustomer.ID
	}

	// 2. Создаем нового клиента с защитой от гонки (race conditions)
	newCust := &domain.Customer{
		Name:            name,
		Phone:           phone,
		City:            "",
		PreferredTk:     strings.TrimSpace(req.TransportCompany),
		Category:        "regular",
		DiscountPercent: 0,
		Notes:           "Создан автоматически из заказа",
	}

	if err := s.orderRepo.CreateCustomer(ctx, newCust); err == nil {
		return newCust.ID
	}

	// Если произошел конфликт уникальности из-за параллельного запроса
	if conflictCust, errConf := s.orderRepo.GetCustomerByPhone(ctx, phone); errConf == nil && conflictCust != nil {
		return conflictCust.ID
	}
	if strings.HasPrefix(phone, "+7") && len(phone) == 12 {
		if conflictCust, errConf := s.orderRepo.GetCustomerByPhone(ctx, "8"+phone[2:]); errConf == nil && conflictCust != nil {
			return conflictCust.ID
		}
		if conflictCust, errConf := s.orderRepo.GetCustomerByPhone(ctx, phone[1:]); errConf == nil && conflictCust != nil {
			return conflictCust.ID
		}
	}
	if conflictCust, errConf := s.orderRepo.GetCustomerByPhone(ctx, buyerNumber); errConf == nil && conflictCust != nil {
		return conflictCust.ID
	}

	return 0
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
			logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to get part for order item")
			return nil, fmt.Errorf("failed to get part information for part %d: %w", item.PartID, err)
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
			logrus.WithError(err).Error("Failed to create order in database")
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
				logrus.WithError(err).Error("Failed to create order item in database")
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
			logrus.WithError(err).Warn("Failed to invalidate orders cache after creating order")
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
			logrus.WithError(err).Warn("Failed to invalidate orders cache after updating status")
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
			logrus.WithError(err).Warn("Failed to invalidate orders cache after updating order details")
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
	logrus.WithField("order_id", orderID).Info("Starting order completion")

	order, err := s.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		logrus.WithError(err).WithField("order_id", orderID).Error("Failed to find order for completion")
		return domain.NotFoundError{Resource: "order", ID: orderID}
	}

	var subtotal float64
	for _, item := range order.Items {
		subtotal += item.Price * float64(item.Quantity)
		if item.PartID <= 0 {
			continue
		}

		part, errPart := s.partRepo.FindByID(ctx, item.PartID)
		if errPart != nil {
			logrus.WithError(errPart).WithField("part_id", item.PartID).Warn("Part not found during order completion (may already be deleted)")
			continue
		}

		if part.Quantity > item.Quantity {
			opID := fmt.Sprintf("complete-order-%d-part-%d", orderID, item.PartID)
			if err := s.partRepo.DecreaseQuantity(ctx, item.PartID, item.Quantity, opID); err != nil {
				logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to decrease part quantity on order completion")
				return fmt.Errorf("failed to decrease part %d quantity: %w", item.PartID, err)
			}
			logrus.WithFields(logrus.Fields{
				"part_id":       item.PartID,
				"sold_quantity": item.Quantity,
				"remaining":     part.Quantity - item.Quantity,
			}).Info("Part quantity decreased after order completion")
		} else {
			if err := s.partRepo.DeletePart(ctx, item.PartID); err != nil {
				logrus.WithError(err).WithField("part_id", item.PartID).Error("Failed to delete depleted part after order completion")
				return fmt.Errorf("failed to delete part %d: %w", item.PartID, err)
			}
			logrus.WithField("part_id", item.PartID).Info("Depleted part deleted after order completion")
		}
	}

	totalAmount := math.Max(subtotal-order.Discount, 0)

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
			logrus.WithError(err).WithField("order_id", orderID).Warn("Failed to create sales history entry")
		}

		return nil
	})
	if err != nil {
		return err
	}

	if s.publisher != nil {
		if err := s.publisher.PublishOrderCompleted(ctx, orderID, totalAmount); err != nil {
			logrus.WithError(err).WithField("order_id", orderID).Warn("Failed to publish order completed event")
		}
	}

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			logrus.WithError(err).Warn("Failed to invalidate orders cache after completing order")
		}
	}

	logrus.WithFields(logrus.Fields{
		"order_id": orderID,
		"amount":   totalAmount,
	}).Info("Order completed successfully")

	return nil
}

// DeleteOrder удаляет заказ
func (s *ordersService) DeleteOrder(ctx context.Context, orderID int64) error {
	logrus.WithField("order_id", orderID).Info("Starting order deletion")

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
			logrus.WithError(err).Warn("Failed to invalidate orders cache after deleting order")
		}
	}

	return nil
}

// AddOrderItem добавляет позицию в существующий заказ
func (s *ordersService) AddOrderItem(ctx context.Context, orderID int64, req domain.AddOrderItemRequest) error {
	if req.Quantity <= 0 {
		req.Quantity = 1
	}

	part, err := s.partRepo.FindByID(ctx, req.PartID)
	if err != nil {
		return domain.NotFoundError{Resource: "part", ID: req.PartID}
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

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			logrus.WithError(err).Warn("Failed to invalidate orders cache after adding item")
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

	if req.Quantity != nil && *req.Quantity > 0 {
		target.Quantity = *req.Quantity
	}
	if req.Price != nil && *req.Price >= 0 {
		target.Price = *req.Price
	}

	if err := s.orderRepo.UpdateItem(ctx, target); err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			logrus.WithError(err).Warn("Failed to invalidate orders cache after updating item")
		}
	}
	return nil
}

// DeleteOrderItem удаляет позицию из заказа
func (s *ordersService) DeleteOrderItem(ctx context.Context, orderID, itemID int64) error {
	order, err := s.orderRepo.FindWithItemsByID(ctx, orderID)
	if err != nil {
		return domain.NotFoundError{Resource: "order", ID: orderID}
	}
	if len(order.Items) <= 1 {
		return domain.ValidationError{Field: "items", Message: "cannot remove the last item of an order; delete the order instead"}
	}

	if err := s.orderRepo.DeleteItem(ctx, orderID, itemID); err != nil {
		return err
	}

	if s.cache != nil {
		if err := s.cache.InvalidateOrders(); err != nil {
			logrus.WithError(err).Warn("Failed to invalidate orders cache after deleting item")
		}
	}
	return nil
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

// GetMonthlySales получает продажи по месяцам
func (s *ordersService) GetMonthlySales(ctx context.Context) ([]domain.MonthlySales, error) {
	return s.orderRepo.GetMonthlySales(ctx)
}

// ListCustomers возвращает список клиентов с фильтрацией и поиском
func (s *ordersService) ListCustomers(ctx context.Context, category, search string, limit, offset int32) ([]domain.CustomerWithStats, error) {
	return s.orderRepo.ListCustomersWithStats(ctx, strings.TrimSpace(category), strings.TrimSpace(search), limit, offset)
}

// GetCustomer возвращает карточку клиента и историю его заказов
func (s *ordersService) GetCustomer(ctx context.Context, id int64) (*domain.CustomerDetails, error) {
	customer, err := s.orderRepo.GetCustomerByID(ctx, id)
	if err != nil {
		return nil, domain.NotFoundError{Resource: "customer", ID: id}
	}

	orders, err := s.orderRepo.GetOrdersByCustomerID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch customer orders: %w", err)
	}

	s.enrichOrders(ctx, orders, false)

	var totalSpent float64
	var lastOrderAt *time.Time
	for _, o := range orders {
		totalSpent += o.TotalAmount
		if lastOrderAt == nil || o.CreatedAt.After(*lastOrderAt) {
			t := o.CreatedAt
			lastOrderAt = &t
		}
	}

	return &domain.CustomerDetails{
		CustomerWithStats: domain.CustomerWithStats{
			Customer:    *customer,
			TotalOrders: int64(len(orders)),
			TotalSpent:  totalSpent,
			LastOrderAt: lastOrderAt,
		},
		Orders: orders,
	}, nil
}

// CreateCustomer создаёт нового клиента вручную
func (s *ordersService) CreateCustomer(ctx context.Context, req domain.CreateCustomerRequest) (*domain.Customer, error) {
	phone := strings.TrimSpace(req.Phone)
	name := strings.TrimSpace(req.Name)
	if phone == "" && name == "" {
		return nil, domain.ValidationError{Field: "phone", Message: "укажите телефон или имя клиента"}
	}

	if phone != "" {
		if normPhone, normName := parseBuyerInfo(phone); normPhone != "" {
			phone = normPhone
			if name == "" && normName != "" && normName != normPhone {
				name = normName
			}
		}

		existing, err := s.orderRepo.GetCustomerByPhone(ctx, phone)
		if err == nil && existing != nil {
			return nil, domain.ConflictError{
				Resource: "customer",
				Message:  fmt.Sprintf("клиент с номером %s уже существует (ID %d: %s)", phone, existing.ID, existing.Name),
			}
		}
	}

	if name == "" {
		name = phone
	}

	category := strings.TrimSpace(req.Category)
	if category == "" {
		category = "regular"
	}

	customer := &domain.Customer{
		Name:            name,
		Phone:           phone,
		City:            strings.TrimSpace(req.City),
		PreferredTk:     strings.TrimSpace(req.PreferredTk),
		PassportOrInn:   strings.TrimSpace(req.PassportOrInn),
		Category:        category,
		DiscountPercent: math.Max(req.DiscountPercent, 0),
		Notes:           strings.TrimSpace(req.Notes),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	if err := s.orderRepo.CreateCustomer(ctx, customer); err != nil {
		return nil, fmt.Errorf("не удалось создать клиента: %w", err)
	}

	return customer, nil
}

// UpdateCustomer обновляет данные клиента
func (s *ordersService) UpdateCustomer(ctx context.Context, id int64, req domain.UpdateCustomerRequest) (*domain.Customer, error) {
	customer, err := s.orderRepo.GetCustomerByID(ctx, id)
	if err != nil {
		return nil, domain.NotFoundError{Resource: "customer", ID: id}
	}

	if req.Name != "" {
		customer.Name = strings.TrimSpace(req.Name)
	}
	if req.Phone != "" {
		customer.Phone = strings.TrimSpace(req.Phone)
	}
	customer.City = strings.TrimSpace(req.City)
	customer.PreferredTk = strings.TrimSpace(req.PreferredTk)
	customer.PassportOrInn = strings.TrimSpace(req.PassportOrInn)
	if req.Category != "" {
		customer.Category = strings.TrimSpace(req.Category)
	}
	customer.DiscountPercent = math.Max(req.DiscountPercent, 0)
	customer.Notes = strings.TrimSpace(req.Notes)
	customer.UpdatedAt = time.Now()

	if err := s.orderRepo.UpdateCustomer(ctx, customer); err != nil {
		return nil, fmt.Errorf("не удалось обновить клиента: %w", err)
	}

	return customer, nil
}

// DeleteCustomer удаляет клиента
func (s *ordersService) DeleteCustomer(ctx context.Context, id int64) error {
	_, err := s.orderRepo.GetCustomerByID(ctx, id)
	if err != nil {
		return domain.NotFoundError{Resource: "customer", ID: id}
	}
	return s.orderRepo.DeleteCustomer(ctx, id)
}
