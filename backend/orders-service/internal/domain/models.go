package domain

import (
	"context"
	"time"
)

// EventPublisher определяет интерфейс для публикации событий
type EventPublisher interface {
	PublishOrderCompleted(ctx context.Context, orderID int64, amount float64) error
	PublishUserAction(ctx context.Context, userID, action string, details map[string]interface{}) error
}

// Order представляет заказ в системе
type Order struct {
	ID                   int64       `json:"id"`
	CustomerID           int64       `json:"customer_id"`
	OrderNumber          string      `json:"order_number"`
	Source               string      `json:"source"`
	SellerID             int64       `json:"seller_id"` // ID пользователя-продавца
	Seller               string      `json:"seller"`    // Имя продавца (для отображения)
	Part                 string      `json:"part"`      // Название детали (для отображения)
	PartID               int64       `json:"part_id"`   // ID детали из инвентаря
	Location             string      `json:"location"`  // Склад, где находится запчасть
	BuyerNumber          string      `json:"buyer_number"`
	Status               string      `json:"status"`
	StatusText           string      `json:"status_text"`
	PaymentStatus        string      `json:"payment_status"`
	WarehouseStatus      string      `json:"warehouse_status"`
	DeliveryMethod       string      `json:"delivery_method"`
	TransportCompany     string      `json:"transport_company"`
	TrackingNumber       string      `json:"tracking_number"`
	Notes                string      `json:"notes"`
	Discount             float64     `json:"discount"`
	TotalAmount          float64     `json:"total_amount"`
	AutoDeleted          bool        `json:"auto_deleted"` // Завершённый заказ (в архиве)
	CreatedAt            time.Time   `json:"created_at"`
	CreatedAtFormatted   string      `json:"created_at_formatted"` // Форматированная дата для фронтенда
	TimeAgo              string      `json:"time_ago"`             // Относительное время (например, "3 дня назад")
	CompletedAt          *time.Time  `json:"completed_at,omitempty"`
	CompletedAtFormatted string      `json:"completed_at_formatted,omitempty"`
	UpdatedAt            time.Time   `json:"updated_at"`
	Items                []OrderItem `json:"items"`
}

// OrderItem представляет позицию в заказе
type OrderItem struct {
	ID               int64   `json:"id"`
	OrderID          int64   `json:"order_id"`
	PartID           int64   `json:"part_id"`
	PartName         string  `json:"part_name,omitempty"`
	PartNameSnapshot string  `json:"part_name_snapshot,omitempty"`
	Location         string  `json:"location,omitempty"`
	Quantity         int     `json:"quantity"`
	Price            float64 `json:"price"` // Цена продажи позиции в заказе
}

// Part представляет запчасть в системе (упрощенная версия для orders-service)
type Part struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name,omitempty"`
	Quantity int     `json:"quantity"`
	Price    float64 `json:"price"`
	Location string  `json:"location,omitempty"`
	Photo    string  `json:"photo,omitempty"`
}

// SalesHistory хранит исторические данные продаж по месяцам
type SalesHistory struct {
	ID        int64     `json:"id"`
	Month     string    `json:"month"` // Формат: "2023-12"
	Sales     float64   `json:"sales"`
	CreatedAt time.Time `json:"created_at"`
}

// MonthlySales представляет продажи за месяц
type MonthlySales struct {
	Month string  `json:"month"`
	Sales float64 `json:"sales"`
}

// Customer представляет клиента в системе
type Customer struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Phone           string    `json:"phone"`
	City            string    `json:"city"`
	PreferredTk     string    `json:"preferred_tk"`
	PassportOrInn   string    `json:"passport_or_inn"`
	Category        string    `json:"category"` // regular, vip, wholesale, blacklist
	DiscountPercent float64   `json:"discount_percent"`
	Notes           string    `json:"notes"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// CustomerWithStats представляет клиента с агрегированной статистикой заказов
type CustomerWithStats struct {
	Customer
	TotalOrders int64      `json:"total_orders"`
	TotalSpent  float64    `json:"total_spent"`
	LastOrderAt *time.Time `json:"last_order_at,omitempty"`
}

// CustomerDetails представляет карточку клиента с историей его заказов
type CustomerDetails struct {
	CustomerWithStats
	Orders []Order `json:"orders"`
}

// CreateOrderItemInput входные данные позиции заказа
type CreateOrderItemInput struct {
	PartID   int64    `json:"part_id"`
	Quantity int      `json:"quantity"`
	Price    *float64 `json:"price,omitempty"`
}

// CreateOrderRequest запрос на создание заказа
type CreateOrderRequest struct {
	CustomerID       int64                  `json:"customer_id"`
	OrderNumber      string                 `json:"order_number"`
	Source           string                 `json:"source"`
	Part             string                 `json:"part"`
	PartID           int64                  `json:"part_id"`
	BuyerNumber      string                 `json:"buyer_number"`
	PaymentStatus    string                 `json:"payment_status"`
	WarehouseStatus  string                 `json:"warehouse_status"`
	DeliveryMethod   string                 `json:"delivery_method"`
	TransportCompany string                 `json:"transport_company"`
	TrackingNumber   string                 `json:"tracking_number"`
	Notes            string                 `json:"notes"`
	Discount         float64                `json:"discount"`
	QuickSale        bool                   `json:"quick_sale"`
	Items            []CreateOrderItemInput `json:"items"`
}

// UpdateOrderDetailsRequest запрос на частичное обновление деталей заказа
type UpdateOrderDetailsRequest struct {
	BuyerNumber      *string  `json:"buyer_number,omitempty"`
	OrderNumber      *string  `json:"order_number,omitempty"`
	Source           *string  `json:"source,omitempty"`
	Status           *string  `json:"status,omitempty"`
	PaymentStatus    *string  `json:"payment_status,omitempty"`
	WarehouseStatus  *string  `json:"warehouse_status,omitempty"`
	DeliveryMethod   *string  `json:"delivery_method,omitempty"`
	TransportCompany *string  `json:"transport_company,omitempty"`
	TrackingNumber   *string  `json:"tracking_number,omitempty"`
	Notes            *string  `json:"notes,omitempty"`
	Discount         *float64 `json:"discount,omitempty"`
}

// AddOrderItemRequest запрос на добавление позиции в заказ
type AddOrderItemRequest struct {
	PartID   int64    `json:"part_id"`
	Quantity int      `json:"quantity"`
	Price    *float64 `json:"price,omitempty"`
}

// UpdateOrderItemRequest запрос на редактирование позиции в заказе
type UpdateOrderItemRequest struct {
	Quantity *int     `json:"quantity,omitempty"`
	Price    *float64 `json:"price,omitempty"`
}

// CreateCustomerRequest запрос на создание клиента
type CreateCustomerRequest struct {
	Name            string  `json:"name"`
	Phone           string  `json:"phone"`
	City            string  `json:"city"`
	PreferredTk     string  `json:"preferred_tk"`
	PassportOrInn   string  `json:"passport_or_inn"`
	Category        string  `json:"category"`
	DiscountPercent float64 `json:"discount_percent"`
	Notes           string  `json:"notes"`
}

// UpdateCustomerRequest запрос на обновление клиента
type UpdateCustomerRequest struct {
	Name            string  `json:"name"`
	Phone           string  `json:"phone"`
	City            string  `json:"city"`
	PreferredTk     string  `json:"preferred_tk"`
	PassportOrInn   string  `json:"passport_or_inn"`
	Category        string  `json:"category"`
	DiscountPercent float64 `json:"discount_percent"`
	Notes           string  `json:"notes"`
}
