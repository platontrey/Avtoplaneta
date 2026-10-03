package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"orders-service/db/sqlc"
	"orders-service/internal/domain"
)

// OrderRepository определяет интерфейс для работы с заказами и клиентами
type OrderRepository interface {
	Create(ctx context.Context, order *domain.Order) error
	CreateItem(ctx context.Context, item *domain.OrderItem) error
	FindByID(ctx context.Context, id int64) (*domain.Order, error)
	FindWithItemsByID(ctx context.Context, id int64) (*domain.Order, error)
	FindAll(ctx context.Context) ([]domain.Order, error)
	FindActive(ctx context.Context) ([]domain.Order, error)
	FindCompleted(ctx context.Context) ([]domain.Order, error)
	FindWithItems(ctx context.Context) ([]domain.Order, error)
	FindOrderItem(ctx context.Context, orderID, partID int64) (*domain.OrderItem, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	CompleteRecord(ctx context.Context, id int64) error
	UpdateItem(ctx context.Context, item *domain.OrderItem) error
	DeleteItem(ctx context.Context, orderID, itemID int64) error
	Delete(ctx context.Context, id int64) error
	DeleteItemsByOrderID(ctx context.Context, orderID int64) error
	MarkExpiredAsAutoDeleted(ctx context.Context, before time.Time) error
	GetMonthlySales(ctx context.Context) ([]domain.MonthlySales, error)
	CreateSalesHistory(ctx context.Context, history *domain.SalesHistory) error
	UpdateSellerName(ctx context.Context, sellerID int64, name string) error
	GetPool() *pgxpool.Pool

	// Customer methods
	CreateCustomer(ctx context.Context, customer *domain.Customer) error
	GetCustomerByID(ctx context.Context, id int64) (*domain.Customer, error)
	GetCustomerByPhone(ctx context.Context, phone string) (*domain.Customer, error)
	ListCustomersWithStats(ctx context.Context, category, search string, limit, offset int32) ([]domain.CustomerWithStats, error)
	UpdateCustomer(ctx context.Context, customer *domain.Customer) error
	DeleteCustomer(ctx context.Context, id int64) error
	GetOrdersByCustomerID(ctx context.Context, customerID int64) ([]domain.Order, error)
}

// orderRepository реализует OrderRepository с использованием sqlc и pgx
type orderRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewOrderRepository создает экземпляр OrderRepository
func NewOrderRepository(pool *pgxpool.Pool) OrderRepository {
	return &orderRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *orderRepository) getQueries(ctx context.Context) *sqlc.Queries {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return r.queries.WithTx(tx)
	}
	return r.queries
}

func (r *orderRepository) getExec(ctx context.Context) sqlc.DBTX {
	if tx, ok := ctx.Value(txKey).(pgx.Tx); ok {
		return tx
	}
	return r.pool
}

// Helpers for time conversion
func toTime(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func toTimePtr(t pgtype.Timestamptz) *time.Time {
	if !t.Valid || t.Time.IsZero() {
		return nil
	}
	tm := t.Time
	return &tm
}

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()}
}

func toTimestamptzPtr(t *time.Time) pgtype.Timestamptz {
	if t == nil || t.IsZero() {
		return pgtype.Timestamptz{Valid: false}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func sqlcOrderToDomain(o sqlc.Order) domain.Order {
	completedAt := toTimePtr(o.CompletedAt)
	completedAtFormatted := ""
	if completedAt != nil {
		completedAtFormatted = completedAt.Format("2006-01-02 15:04:05")
	}

	return domain.Order{
		ID:                   o.ID,
		CustomerID:           o.CustomerID,
		OrderNumber:          o.OrderNumber,
		Source:               o.Source,
		SellerID:             o.SellerID,
		Seller:               o.Seller,
		Part:                 o.Part,
		PartID:               o.PartID,
		Location:             o.Location,
		BuyerNumber:          o.BuyerNumber,
		Status:               o.Status,
		StatusText:           o.StatusText,
		PaymentStatus:        o.PaymentStatus,
		WarehouseStatus:      o.WarehouseStatus,
		DeliveryMethod:       o.DeliveryMethod,
		TransportCompany:     o.TransportCompany,
		TrackingNumber:       o.TrackingNumber,
		Notes:                o.Notes,
		Discount:             o.Discount,
		AutoDeleted:          o.AutoDeleted,
		CreatedAt:            toTime(o.CreatedAt),
		CompletedAt:          completedAt,
		CompletedAtFormatted: completedAtFormatted,
		UpdatedAt:            toTime(o.UpdatedAt),
		Items:                []domain.OrderItem{},
	}
}

func sqlcOrderItemToDomain(oi sqlc.OrderItem) domain.OrderItem {
	return domain.OrderItem{
		ID:               oi.ID,
		OrderID:          oi.OrderID,
		PartID:           oi.PartID,
		PartName:         oi.PartNameSnapshot,
		PartNameSnapshot: oi.PartNameSnapshot,
		Quantity:         int(oi.Quantity),
		Price:            oi.Price,
	}
}

func sqlcCustomerToDomain(c sqlc.Customer) domain.Customer {
	return domain.Customer{
		ID:              c.ID,
		Name:            c.Name,
		Phone:           c.Phone,
		City:            c.City,
		PreferredTk:     c.PreferredTk,
		PassportOrInn:   c.PassportOrInn,
		Category:        c.Category,
		DiscountPercent: c.DiscountPercent,
		Notes:           c.Notes,
		CreatedAt:       toTime(c.CreatedAt),
		UpdatedAt:       toTime(c.UpdatedAt),
	}
}

func sqlcCustomerRowToDomain(r sqlc.ListCustomersWithStatsRow) domain.CustomerWithStats {
	return domain.CustomerWithStats{
		Customer: domain.Customer{
			ID:              r.ID,
			Name:            r.Name,
			Phone:           r.Phone,
			City:            r.City,
			PreferredTk:     r.PreferredTk,
			PassportOrInn:   r.PassportOrInn,
			Category:        r.Category,
			DiscountPercent: r.DiscountPercent,
			Notes:           r.Notes,
			CreatedAt:       toTime(r.CreatedAt),
			UpdatedAt:       toTime(r.UpdatedAt),
		},
		TotalOrders: r.TotalOrders,
		TotalSpent:  r.TotalSpent,
		LastOrderAt: toTimePtr(r.LastOrderAt),
	}
}

func (r *orderRepository) Create(ctx context.Context, order *domain.Order) error {
	now := time.Now()
	if order.CreatedAt.IsZero() {
		order.CreatedAt = now
	}
	if order.UpdatedAt.IsZero() {
		order.UpdatedAt = now
	}
	if order.PaymentStatus == "" {
		order.PaymentStatus = "unpaid"
	}
	if order.WarehouseStatus == "" {
		order.WarehouseStatus = "inspecting"
	}
	if order.DeliveryMethod == "" {
		order.DeliveryMethod = "pickup"
	}

	params := sqlc.CreateOrderParams{
		CustomerID:       order.CustomerID,
		SellerID:         order.SellerID,
		Seller:           order.Seller,
		Part:             order.Part,
		PartID:           order.PartID,
		Location:         order.Location,
		BuyerNumber:      order.BuyerNumber,
		Status:           order.Status,
		StatusText:       order.StatusText,
		AutoDeleted:      order.AutoDeleted,
		CreatedAt:        toTimestamptz(order.CreatedAt),
		OrderNumber:      order.OrderNumber,
		Source:           order.Source,
		PaymentStatus:    order.PaymentStatus,
		WarehouseStatus:  order.WarehouseStatus,
		DeliveryMethod:   order.DeliveryMethod,
		TransportCompany: order.TransportCompany,
		TrackingNumber:   order.TrackingNumber,
		Notes:            order.Notes,
		Discount:         order.Discount,
		CompletedAt:      toTimestamptzPtr(order.CompletedAt),
		UpdatedAt:        toTimestamptz(order.UpdatedAt),
	}

	o, err := r.getQueries(ctx).CreateOrder(ctx, params)
	if err != nil {
		return err
	}
	order.ID = o.ID
	return nil
}

func (r *orderRepository) FindByID(ctx context.Context, id int64) (*domain.Order, error) {
	o, err := r.getQueries(ctx).GetOrderByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("failed to find order with ID %d: record not found", id)
		}
		return nil, fmt.Errorf("failed to find order with ID %d: %w", id, err)
	}
	d := sqlcOrderToDomain(o)
	return &d, nil
}

func (r *orderRepository) FindWithItemsByID(ctx context.Context, id int64) (*domain.Order, error) {
	o, err := r.getQueries(ctx).GetOrderByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("failed to find order with ID %d: record not found", id)
		}
		return nil, fmt.Errorf("failed to find order with ID %d: %w", id, err)
	}

	items, err := r.getQueries(ctx).GetOrderItemsByOrderID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find items for order %d: %w", id, err)
	}

	d := sqlcOrderToDomain(o)
	d.Items = make([]domain.OrderItem, len(items))
	for i, item := range items {
		d.Items[i] = sqlcOrderItemToDomain(item)
	}
	return &d, nil
}

func (r *orderRepository) FindAll(ctx context.Context) ([]domain.Order, error) {
	rows, err := r.getQueries(ctx).FindAllOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all orders: %w", err)
	}

	orders := make([]domain.Order, len(rows))
	for i, row := range rows {
		orders[i] = sqlcOrderToDomain(row)
	}
	return orders, nil
}

func (r *orderRepository) attachItems(ctx context.Context, rows []sqlc.Order) ([]domain.Order, error) {
	if len(rows) == 0 {
		return []domain.Order{}, nil
	}

	orderIDs := make([]int64, len(rows))
	for i, row := range rows {
		orderIDs[i] = row.ID
	}

	items, err := r.getQueries(ctx).GetOrderItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load items for orders: %w", err)
	}

	itemsMap := make(map[int64][]domain.OrderItem)
	for _, item := range items {
		itemsMap[item.OrderID] = append(itemsMap[item.OrderID], sqlcOrderItemToDomain(item))
	}

	orders := make([]domain.Order, len(rows))
	for i, row := range rows {
		d := sqlcOrderToDomain(row)
		d.Items = itemsMap[row.ID]
		if d.Items == nil {
			d.Items = []domain.OrderItem{}
		}
		orders[i] = d
	}

	return orders, nil
}

func (r *orderRepository) FindWithItems(ctx context.Context) ([]domain.Order, error) {
	rows, err := r.getQueries(ctx).FindAllOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders with items: %w", err)
	}
	return r.attachItems(ctx, rows)
}

func (r *orderRepository) FindActive(ctx context.Context) ([]domain.Order, error) {
	return r.FindWithItems(ctx)
}

func (r *orderRepository) FindCompleted(ctx context.Context) ([]domain.Order, error) {
	rows, err := r.getQueries(ctx).FindCompletedOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find completed orders: %w", err)
	}
	return r.attachItems(ctx, rows)
}

func (r *orderRepository) Update(ctx context.Context, id int64, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return nil
	}

	builder := squirrel.Update("orders").Where(squirrel.Eq{"id": id})
	for k, v := range updates {
		if t, ok := v.(time.Time); ok {
			builder = builder.Set(k, toTimestamptz(t))
		} else {
			builder = builder.Set(k, v)
		}
	}
	builder = builder.Set("updated_at", toTimestamptz(time.Now()))

	query, args, err := builder.PlaceholderFormat(squirrel.Dollar).ToSql()
	if err != nil {
		return fmt.Errorf("failed to build update query: %w", err)
	}

	_, err = r.getExec(ctx).Exec(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("failed to update order %d: %w", id, err)
	}
	return nil
}

func (r *orderRepository) CompleteRecord(ctx context.Context, id int64) error {
	if err := r.getQueries(ctx).CompleteOrderRecord(ctx, id); err != nil {
		return fmt.Errorf("failed to complete order record %d: %w", id, err)
	}
	return nil
}

func (r *orderRepository) Delete(ctx context.Context, id int64) error {
	err := r.getQueries(ctx).DeleteOrder(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete order %d: %w", id, err)
	}
	return nil
}

func (r *orderRepository) MarkExpiredAsAutoDeleted(ctx context.Context, before time.Time) error {
	err := r.getQueries(ctx).MarkExpiredAsAutoDeleted(ctx, toTimestamptz(before))
	if err != nil {
		return fmt.Errorf("failed to mark expired orders as auto-deleted: %w", err)
	}
	return nil
}

func (r *orderRepository) CreateItem(ctx context.Context, item *domain.OrderItem) error {
	snapshot := item.PartNameSnapshot
	if snapshot == "" {
		snapshot = item.PartName
	}
	params := sqlc.CreateOrderItemParams{
		OrderID:          item.OrderID,
		PartID:           item.PartID,
		Quantity:         int32(item.Quantity),
		Price:            item.Price,
		PartNameSnapshot: snapshot,
	}

	oi, err := r.getQueries(ctx).CreateOrderItem(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}
	item.ID = oi.ID
	return nil
}

func (r *orderRepository) FindOrderItem(ctx context.Context, orderID, partID int64) (*domain.OrderItem, error) {
	params := sqlc.FindOrderItemParams{
		OrderID: orderID,
		PartID:  partID,
	}

	oi, err := r.getQueries(ctx).FindOrderItem(ctx, params)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("failed to find order item for order %d and part %d: record not found", orderID, partID)
		}
		return nil, fmt.Errorf("failed to find order item: %w", err)
	}
	d := sqlcOrderItemToDomain(oi)
	return &d, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id int64, status string) error {
	params := sqlc.UpdateOrderStatusParams{
		ID:     id,
		Status: status,
	}
	err := r.getQueries(ctx).UpdateOrderStatus(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update status for order %d: %w", id, err)
	}
	return nil
}

func (r *orderRepository) UpdateItem(ctx context.Context, item *domain.OrderItem) error {
	params := sqlc.UpdateOrderItemParams{
		ID:       item.ID,
		Quantity: int32(item.Quantity),
		Price:    item.Price,
	}
	err := r.getQueries(ctx).UpdateOrderItem(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update order item: %w", err)
	}
	return nil
}

func (r *orderRepository) DeleteItem(ctx context.Context, orderID, itemID int64) error {
	err := r.getQueries(ctx).DeleteOrderItem(ctx, sqlc.DeleteOrderItemParams{
		ID:      itemID,
		OrderID: orderID,
	})
	if err != nil {
		return fmt.Errorf("failed to delete order item %d: %w", itemID, err)
	}
	return nil
}

func (r *orderRepository) DeleteItemsByOrderID(ctx context.Context, orderID int64) error {
	err := r.getQueries(ctx).DeleteOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order items for order %d: %w", orderID, err)
	}
	return nil
}

func (r *orderRepository) GetMonthlySales(ctx context.Context) ([]domain.MonthlySales, error) {
	rows, err := r.getQueries(ctx).GetMonthlySales(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly sales: %w", err)
	}

	monthlySales := make([]domain.MonthlySales, len(rows))
	for i, row := range rows {
		monthlySales[i] = domain.MonthlySales{Month: row.Month, Sales: row.Sales}
	}
	return monthlySales, nil
}

func (r *orderRepository) CreateSalesHistory(ctx context.Context, history *domain.SalesHistory) error {
	params := sqlc.CreateSalesHistoryParams{
		Month:     history.Month,
		Sales:     history.Sales,
		CreatedAt: toTimestamptz(history.CreatedAt),
	}

	sh, err := r.getQueries(ctx).CreateSalesHistory(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create sales history: %w", err)
	}
	history.ID = sh.ID
	return nil
}

func (r *orderRepository) UpdateSellerName(ctx context.Context, sellerID int64, name string) error {
	return r.getQueries(ctx).UpdateSellerName(ctx, sqlc.UpdateSellerNameParams{
		SellerID: sellerID,
		Seller:   name,
	})
}

func (r *orderRepository) CreateCustomer(ctx context.Context, customer *domain.Customer) error {
	now := time.Now()
	if customer.CreatedAt.IsZero() {
		customer.CreatedAt = now
	}
	customer.UpdatedAt = now

	params := sqlc.CreateCustomerParams{
		Name:            customer.Name,
		Phone:           customer.Phone,
		City:            customer.City,
		PreferredTk:     customer.PreferredTk,
		PassportOrInn:   customer.PassportOrInn,
		Category:        customer.Category,
		DiscountPercent: customer.DiscountPercent,
		Notes:           customer.Notes,
		CreatedAt:       toTimestamptz(customer.CreatedAt),
		UpdatedAt:       toTimestamptz(customer.UpdatedAt),
	}

	c, err := r.getQueries(ctx).CreateCustomer(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create customer: %w", err)
	}

	customer.ID = c.ID
	return nil
}

func (r *orderRepository) GetCustomerByID(ctx context.Context, id int64) (*domain.Customer, error) {
	c, err := r.getQueries(ctx).GetCustomerByID(ctx, id)
	if err != nil {
		return nil, err
	}
	d := sqlcCustomerToDomain(c)
	return &d, nil
}

func (r *orderRepository) GetCustomerByPhone(ctx context.Context, phone string) (*domain.Customer, error) {
	c, err := r.getQueries(ctx).GetCustomerByPhone(ctx, phone)
	if err != nil {
		return nil, err
	}
	d := sqlcCustomerToDomain(c)
	return &d, nil
}

func (r *orderRepository) ListCustomersWithStats(ctx context.Context, category, search string, limit, offset int32) ([]domain.CustomerWithStats, error) {
	if limit <= 0 {
		limit = 50
	}
	params := sqlc.ListCustomersWithStatsParams{
		Category:  category,
		Search:    search,
		LimitVal:  limit,
		OffsetVal: offset,
	}

	rows, err := r.getQueries(ctx).ListCustomersWithStats(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list customers: %w", err)
	}

	customers := make([]domain.CustomerWithStats, len(rows))
	for i, row := range rows {
		customers[i] = sqlcCustomerRowToDomain(row)
	}
	return customers, nil
}

func (r *orderRepository) UpdateCustomer(ctx context.Context, customer *domain.Customer) error {
	params := sqlc.UpdateCustomerParams{
		ID:              customer.ID,
		Name:            customer.Name,
		Phone:           customer.Phone,
		City:            customer.City,
		PreferredTk:     customer.PreferredTk,
		PassportOrInn:   customer.PassportOrInn,
		Category:        customer.Category,
		DiscountPercent: customer.DiscountPercent,
		Notes:           customer.Notes,
	}

	c, err := r.getQueries(ctx).UpdateCustomer(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update customer: %w", err)
	}

	*customer = sqlcCustomerToDomain(c)
	return nil
}

func (r *orderRepository) DeleteCustomer(ctx context.Context, id int64) error {
	return r.getQueries(ctx).DeleteCustomer(ctx, id)
}

func (r *orderRepository) GetOrdersByCustomerID(ctx context.Context, customerID int64) ([]domain.Order, error) {
	orders, err := r.getQueries(ctx).GetOrdersByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	domainOrders := make([]domain.Order, len(orders))
	for i, o := range orders {
		domainOrders[i] = sqlcOrderToDomain(o)
	}
	return domainOrders, nil
}

func (r *orderRepository) GetPool() *pgxpool.Pool {
	return r.pool
}
