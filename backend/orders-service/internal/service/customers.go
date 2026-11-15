package service

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"orders-service/internal/domain"
)

// parseBuyerInfo извлекает нормализованный телефон и имя покупателя из строки buyer_number
func parseBuyerInfo(raw string) (phone string, name string) {
	raw = strings.TrimSpace(raw)
	if raw == "" || raw == "Самовывоз" || raw == "Продажа на месте" || raw == "Без контакта" {
		return "", ""
	}

	// Извлекаем все цифры с предварительно выделенной емкостью
	var digits strings.Builder
	digits.Grow(len(raw))
	for i := 0; i < len(raw); i++ {
		ch := raw[i]
		if ch >= '0' && ch <= '9' {
			digits.WriteByte(ch)
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
		lower = strings.ToLower(namePart)
	}
	if strings.HasPrefix(lower, "фио") {
		namePart = strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(namePart[len("фио"):], ":"), " "))
	}

	// Парсим слова имени без аллокаций среза strings.Fields
	var nameBuilder strings.Builder
	nameBuilder.Grow(len(namePart))

	for i := 0; i < len(namePart); {
		// Пропускаем пробелы
		for i < len(namePart) && (namePart[i] == ' ' || namePart[i] == '\t' || namePart[i] == '\n' || namePart[i] == '\r') {
			i++
		}
		if i >= len(namePart) {
			break
		}
		start := i
		for i < len(namePart) && !(namePart[i] == ' ' || namePart[i] == '\t' || namePart[i] == '\n' || namePart[i] == '\r') {
			i++
		}
		w := namePart[start:i]

		cleanWord := strings.Trim(w, "+()- —–")
		isNumeric := true
		for _, ch := range cleanWord {
			if ch < '0' || ch > '9' {
				isNumeric = false
				break
			}
		}
		if !isNumeric || len(cleanWord) < 6 {
			if nameBuilder.Len() > 0 {
				nameBuilder.WriteByte(' ')
			}
			nameBuilder.WriteString(w)
		}
	}

	name = strings.TrimSpace(nameBuilder.String())
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
