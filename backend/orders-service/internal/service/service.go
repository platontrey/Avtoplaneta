package service

import (
	"context"

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
