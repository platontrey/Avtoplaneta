package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"

	"orders-service/db/sqlc"
)

type txKeyType struct{}
var txKey = txKeyType{}

// RunInTransaction выполняет функцию fn в рамках транзакции
func RunInTransaction(ctx context.Context, pool *pgxpool.Pool, fn func(ctx context.Context) error) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	txCtx := context.WithValue(ctx, txKey, tx)
	if err := fn(txCtx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// OrderRepository определяет интерфейс для работы с заказами
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	CreateItem(ctx context.Context, item *OrderItem) error
	FindByID(ctx context.Context, id int64) (*Order, error)
	FindWithItemsByID(ctx context.Context, id int64) (*Order, error)
	FindAll(ctx context.Context) ([]Order, error)
	FindActive(ctx context.Context) ([]Order, error)
	FindWithItems(ctx context.Context) ([]Order, error)
	FindOrderItem(ctx context.Context, orderID, partID int64) (*OrderItem, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	UpdateStatus(ctx context.Context, id int64, status string) error
	UpdateItem(ctx context.Context, item *OrderItem) error
	Delete(ctx context.Context, id int64) error
	DeleteItemsByOrderID(ctx context.Context, orderID int64) error
	MarkExpiredAsAutoDeleted(ctx context.Context, before time.Time) error
	GetMonthlySales(ctx context.Context) ([]MonthlySales, error)
	CreateSalesHistory(ctx context.Context, history *SalesHistory) error
	UpdateSellerName(ctx context.Context, sellerID int64, name string) error
	GetPool() *pgxpool.Pool
}

// PartRepositoryForOrders определяет интерфейс для работы с запчастями (для orders-service)
type PartRepositoryForOrders interface {
	FindByID(ctx context.Context, id int64) (*Part, error)
	UpdateQuantity(ctx context.Context, id int64, newQuantity int) error
	DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	DeletePart(ctx context.Context, id int64) error
}

// orderRepository реализует OrderRepository
type orderRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

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

func toTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: !t.IsZero()}
}

func sqlcOrderToDomain(o sqlc.Order) Order {
	return Order{
		ID:          o.ID,
		CustomerID:  o.CustomerID,
		SellerID:    o.SellerID,
		Seller:      o.Seller,
		Part:        o.Part,
		PartID:      o.PartID,
		Location:    o.Location,
		BuyerNumber: o.BuyerNumber,
		Status:      o.Status,
		StatusText:  o.StatusText,
		AutoDeleted: o.AutoDeleted,
		CreatedAt:   toTime(o.CreatedAt),
		Items:       []OrderItem{},
	}
}

func sqlcOrderItemToDomain(oi sqlc.OrderItem) OrderItem {
	return OrderItem{
		ID:       oi.ID,
		OrderID:  oi.OrderID,
		PartID:   oi.PartID,
		Quantity: int(oi.Quantity),
		Price:    oi.Price,
	}
}

func (r *orderRepository) Create(ctx context.Context, order *Order) error {
	params := sqlc.CreateOrderParams{
		CustomerID:  order.CustomerID,
		SellerID:    order.SellerID,
		Seller:      order.Seller,
		Part:        order.Part,
		PartID:      order.PartID,
		Location:    order.Location,
		BuyerNumber: order.BuyerNumber,
		Status:      order.Status,
		StatusText:  order.StatusText,
		AutoDeleted: order.AutoDeleted,
		CreatedAt:   toTimestamptz(order.CreatedAt),
	}

	o, err := r.getQueries(ctx).CreateOrder(ctx, params)
	if err != nil {
		return err
	}
	order.ID = o.ID
	return nil
}

func (r *orderRepository) FindByID(ctx context.Context, id int64) (*Order, error) {
	o, err := r.getQueries(ctx).GetOrderByID(ctx, id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("failed to find order with ID %d: record not found", id)
		}
		return nil, fmt.Errorf("failed to find order with ID %d: %w", id, err)
	}
	domain := sqlcOrderToDomain(o)
	return &domain, nil
}

func (r *orderRepository) FindWithItemsByID(ctx context.Context, id int64) (*Order, error) {
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

	domain := sqlcOrderToDomain(o)
	domain.Items = make([]OrderItem, len(items))
	for i, item := range items {
		domain.Items[i] = sqlcOrderItemToDomain(item)
	}
	return &domain, nil
}

func (r *orderRepository) FindAll(ctx context.Context) ([]Order, error) {
	rows, err := r.getQueries(ctx).FindAllOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find all orders: %w", err)
	}

	orders := make([]Order, len(rows))
	for i, row := range rows {
		orders[i] = sqlcOrderToDomain(row)
	}
	return orders, nil
}

func (r *orderRepository) FindWithItems(ctx context.Context) ([]Order, error) {
	rows, err := r.getQueries(ctx).FindAllOrders(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to find orders with items: %w", err)
	}

	if len(rows) == 0 {
		return []Order{}, nil
	}

	orderIDs := make([]int64, len(rows))
	for i, row := range rows {
		orderIDs[i] = row.ID
	}

	items, err := r.getQueries(ctx).GetOrderItemsByOrderIDs(ctx, orderIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to load items for orders: %w", err)
	}

	itemsMap := make(map[int64][]OrderItem)
	for _, item := range items {
		itemsMap[item.OrderID] = append(itemsMap[item.OrderID], sqlcOrderItemToDomain(item))
	}

	orders := make([]Order, len(rows))
	for i, row := range rows {
		domain := sqlcOrderToDomain(row)
		domain.Items = itemsMap[row.ID]
		if domain.Items == nil {
			domain.Items = []OrderItem{}
		}
		orders[i] = domain
	}

	return orders, nil
}

func (r *orderRepository) FindActive(ctx context.Context) ([]Order, error) {
	return r.FindWithItems(ctx)
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

func (r *orderRepository) CreateItem(ctx context.Context, item *OrderItem) error {
	params := sqlc.CreateOrderItemParams{
		OrderID:  item.OrderID,
		PartID:   item.PartID,
		Quantity: int32(item.Quantity),
		Price:    item.Price,
	}

	oi, err := r.getQueries(ctx).CreateOrderItem(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}
	item.ID = oi.ID
	return nil
}

func (r *orderRepository) FindOrderItem(ctx context.Context, orderID, partID int64) (*OrderItem, error) {
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
	domain := sqlcOrderItemToDomain(oi)
	return &domain, nil
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

func (r *orderRepository) UpdateItem(ctx context.Context, item *OrderItem) error {
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

func (r *orderRepository) DeleteItemsByOrderID(ctx context.Context, orderID int64) error {
	err := r.getQueries(ctx).DeleteOrderItemsByOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("failed to delete order items for order %d: %w", orderID, err)
	}
	return nil
}

func (r *orderRepository) GetMonthlySales(ctx context.Context) ([]MonthlySales, error) {
	rows, err := r.getQueries(ctx).GetMonthlySales(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly sales: %w", err)
	}

	monthlySales := make([]MonthlySales, len(rows))
	for i, row := range rows {
		monthlySales[i] = MonthlySales{Month: row.Month, Sales: row.Sales}
	}
	return monthlySales, nil
}

func (r *orderRepository) CreateSalesHistory(ctx context.Context, history *SalesHistory) error {
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

func (r *orderRepository) GetPool() *pgxpool.Pool {
	return r.pool
}

// partRepositoryForOrders реализует PartRepositoryForOrders
type partRepositoryForOrders struct {
	client PartsGRPCClient
}

func NewPartRepositoryForOrders(client PartsGRPCClient) PartRepositoryForOrders {
	return &partRepositoryForOrders{
		client: client,
	}
}

func (r *partRepositoryForOrders) FindByID(ctx context.Context, id int64) (*Part, error) {
	p, err := r.client.GetPartByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find part with ID %d via gRPC: %w", id, err)
	}
	return p, nil
}

func (r *partRepositoryForOrders) UpdateQuantity(ctx context.Context, id int64, newQuantity int) error {
	// Not atomic, but this method seems unused for actual order placements.
	// Often it's better to rely on Decrease/Increase.
	// For now, if we need UpdateQuantity, we don't have it exposed via gRPC natively as "SetQuantity".
	// But let's check if it's used. If it is, we need to implement it.
	return fmt.Errorf("UpdateQuantity is not supported via gRPC yet")
}

func (r *partRepositoryForOrders) DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	err := r.client.DecreaseQuantity(ctx, id, amount, operationID)
	if err != nil {
		return fmt.Errorf("failed to decrease quantity for part %d by %d: %w", id, amount, err)
	}
	return nil
}

func (r *partRepositoryForOrders) IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	err := r.client.IncreaseQuantity(ctx, id, amount, operationID)
	if err != nil {
		return fmt.Errorf("failed to increase quantity for part %d by %d: %w", id, amount, err)
	}
	return nil
}

func (r *partRepositoryForOrders) DeletePart(ctx context.Context, id int64) error {
	logrus.WithField("part_id", id).Info("Starting part deletion via gRPC")
	err := r.client.DeletePart(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete part with ID %d via gRPC: %w", id, err)
	}
	return nil
}
