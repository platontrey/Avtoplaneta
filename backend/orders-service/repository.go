package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// OrderRepository определяет интерфейс для работы с заказами
type OrderRepository interface {
	Create(ctx context.Context, order *Order) error
	CreateItem(ctx context.Context, item *OrderItem) error
	FindByID(ctx context.Context, id uint) (*Order, error)
	FindWithItemsByID(ctx context.Context, id uint) (*Order, error)
	FindAll(ctx context.Context) ([]Order, error)
	FindActive(ctx context.Context) ([]Order, error)
	FindWithItems(ctx context.Context) ([]Order, error)
	FindOrderItem(ctx context.Context, orderID, partID uint) (*OrderItem, error)
	Update(ctx context.Context, id uint, updates map[string]interface{}) error
	UpdateStatus(ctx context.Context, id uint, status string) error
	UpdateItem(ctx context.Context, item *OrderItem) error
	Delete(ctx context.Context, id uint) error
	DeleteItemsByOrderID(ctx context.Context, orderID uint) error
	MarkExpiredAsAutoDeleted(ctx context.Context, before time.Time) error
	GetMonthlySales(ctx context.Context) ([]MonthlySales, error)
	CreateSalesHistory(ctx context.Context, history *SalesHistory) error
}

// PartRepositoryForOrders определяет интерфейс для работы с запчастями (для orders-service)
type PartRepositoryForOrders interface {
	FindByID(ctx context.Context, id uint) (*Part, error)
	UpdateQuantity(ctx context.Context, id uint, newQuantity int) error
	DecreaseQuantity(ctx context.Context, id uint, amount int) error
	IncreaseQuantity(ctx context.Context, id uint, amount int) error
	DeletePart(ctx context.Context, id uint) error
}

// orderRepository реализует OrderRepository
type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(ctx context.Context, order *Order) error {
	return r.db.WithContext(ctx).Create(order).Error
}

func (r *orderRepository) FindByID(ctx context.Context, id uint) (*Order, error) {
	var order Order
	err := r.db.WithContext(ctx).First(&order, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find order with ID %d: %w", id, err)
	}
	return &order, nil
}

func (r *orderRepository) FindWithItemsByID(ctx context.Context, id uint) (*Order, error) {
	var order Order
	err := r.db.WithContext(ctx).Preload("Items").First(&order, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find order with items for ID %d: %w", id, err)
	}
	return &order, nil
}

func (r *orderRepository) FindAll(ctx context.Context) ([]Order, error) {
	var orders []Order
	err := r.db.WithContext(ctx).Where("auto_deleted = ?", false).Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find all orders: %w", err)
	}
	return orders, nil
}

func (r *orderRepository) FindWithItems(ctx context.Context) ([]Order, error) {
	var orders []Order
	err := r.db.WithContext(ctx).Preload("Items").Where("auto_deleted = ?", false).Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find orders with items: %w", err)
	}
	return orders, nil
}

func (r *orderRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	err := r.db.WithContext(ctx).Model(&Order{}).Where("id = ?", id).Updates(updates).Error
	if err != nil {
		return fmt.Errorf("failed to update order %d: %w", id, err)
	}
	return nil
}

func (r *orderRepository) Delete(ctx context.Context, id uint) error {
	err := r.db.WithContext(ctx).Delete(&Order{}, id).Error
	if err != nil {
		return fmt.Errorf("failed to delete order %d: %w", id, err)
	}
	return nil
}

func (r *orderRepository) MarkExpiredAsAutoDeleted(ctx context.Context, before time.Time) error {
	err := r.db.WithContext(ctx).Model(&Order{}).Where("created_at <= ? AND auto_deleted = ?", before, false).Update("auto_deleted", true).Error
	if err != nil {
		return fmt.Errorf("failed to mark expired orders as auto-deleted: %w", err)
	}
	return nil
}

func (r *orderRepository) CreateItem(ctx context.Context, item *OrderItem) error {
	err := r.db.WithContext(ctx).Create(item).Error
	if err != nil {
		return fmt.Errorf("failed to create order item: %w", err)
	}
	return nil
}

func (r *orderRepository) FindActive(ctx context.Context) ([]Order, error) {
	var orders []Order
	err := r.db.WithContext(ctx).Preload("Items").Where("auto_deleted = ?", false).Find(&orders).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find active orders: %w", err)
	}
	return orders, nil
}

func (r *orderRepository) FindOrderItem(ctx context.Context, orderID, partID uint) (*OrderItem, error) {
	var item OrderItem
	err := r.db.WithContext(ctx).Where("order_id = ? AND part_id = ?", orderID, partID).First(&item).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find order item for order %d and part %d: %w", orderID, partID, err)
	}
	return &item, nil
}

func (r *orderRepository) UpdateStatus(ctx context.Context, id uint, status string) error {
	err := r.db.WithContext(ctx).Model(&Order{}).Where("id = ?", id).Update("status", status).Error
	if err != nil {
		return fmt.Errorf("failed to update status for order %d: %w", id, err)
	}
	return nil
}

func (r *orderRepository) UpdateItem(ctx context.Context, item *OrderItem) error {
	err := r.db.WithContext(ctx).Save(item).Error
	if err != nil {
		return fmt.Errorf("failed to update order item: %w", err)
	}
	return nil
}

func (r *orderRepository) DeleteItemsByOrderID(ctx context.Context, orderID uint) error {
	err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Delete(&OrderItem{}).Error
	if err != nil {
		return fmt.Errorf("failed to delete order items for order %d: %w", orderID, err)
	}
	return nil
}

func (r *orderRepository) GetMonthlySales(ctx context.Context) ([]MonthlySales, error) {
	var results []struct {
		Month string
		Sales float64
	}
	query, args, err := squirrel.Select("TO_CHAR(orders.created_at, 'YYYY-MM') as month", "SUM(order_items.price * order_items.quantity) as sales").
		From("order_items").
		Join("orders ON order_items.order_id = orders.id").
		Where(squirrel.Eq{"orders.status": "green", "orders.auto_deleted": true}).
		GroupBy("TO_CHAR(orders.created_at, 'YYYY-MM')").
		OrderBy("month DESC").
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build monthly sales query: %w", err)
	}
	err = r.db.WithContext(ctx).Raw(query, args...).Scan(&results).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly sales: %w", err)
	}
	monthlySales := make([]MonthlySales, len(results))
	for i, res := range results {
		monthlySales[i] = MonthlySales{Month: res.Month, Sales: res.Sales}
	}
	return monthlySales, nil
}

func (r *orderRepository) CreateSalesHistory(ctx context.Context, history *SalesHistory) error {
	return r.db.WithContext(ctx).Create(history).Error
}

// partRepositoryForOrders реализует PartRepositoryForOrders
type partRepositoryForOrders struct {
	db *gorm.DB
}

func NewPartRepositoryForOrders(db *gorm.DB) PartRepositoryForOrders {
	return &partRepositoryForOrders{db: db}
}

func (r *partRepositoryForOrders) FindByID(ctx context.Context, id uint) (*Part, error) {
	var part Part
	err := r.db.WithContext(ctx).First(&part, id).Error
	if err != nil {
		return nil, fmt.Errorf("failed to find part with ID %d: %w", id, err)
	}
	return &part, nil
}

func (r *partRepositoryForOrders) UpdateQuantity(ctx context.Context, id uint, newQuantity int) error {
	err := r.db.WithContext(ctx).Model(&Part{}).Where("id = ?", id).Update("quantity", newQuantity).Error
	if err != nil {
		return fmt.Errorf("failed to update quantity for part %d: %w", id, err)
	}
	return nil
}

func (r *partRepositoryForOrders) DecreaseQuantity(ctx context.Context, id uint, amount int) error {
	err := r.db.WithContext(ctx).Model(&Part{}).Where("id = ?", id).Update("quantity", r.db.Raw("CASE WHEN quantity - ? <= 0 THEN -1 ELSE quantity - ? END", amount, amount)).Error
	if err != nil {
		return fmt.Errorf("failed to decrease quantity for part %d by %d: %w", id, amount, err)
	}
	return nil
}

func (r *partRepositoryForOrders) IncreaseQuantity(ctx context.Context, id uint, amount int) error {
	err := r.db.WithContext(ctx).Model(&Part{}).Where("id = ?", id).Update("quantity", r.db.Raw("quantity + ?", amount)).Error
	if err != nil {
		return fmt.Errorf("failed to increase quantity for part %d by %d: %w", id, amount, err)
	}
	return nil
}

func (r *partRepositoryForOrders) DeletePart(ctx context.Context, id uint) error {
	logrus.WithField("part_id", id).Info("Starting part deletion")

	// Получаем запчасть для удаления фото
	var part Part
	if err := r.db.WithContext(ctx).First(&part, id).Error; err != nil {
		logrus.WithError(err).WithField("part_id", id).Error("Failed to find part for deletion")
		return err
	}

	logrus.WithFields(logrus.Fields{
		"part_id": id,
		"photo":   part.Photo,
	}).Info("Part found for deletion, checking photo")

	// Удаляем фото если есть
	if part.Photo != "" {
		// Извлечь имя файла из пути (удалить префикс "/uploads/")
		filename := strings.TrimPrefix(part.Photo, "/uploads/")
		filePath := filepath.Join("../parts-service/uploads", filename)

		logrus.WithFields(logrus.Fields{
			"part_id": id,
			"photo":   part.Photo,
			"file_path": filePath,
		}).Info("Attempting to delete photo file")

		if err := os.Remove(filePath); err != nil {
			// Логируем, но не прерываем
			logrus.WithError(err).WithFields(logrus.Fields{
				"part_id": id,
				"photo":   part.Photo,
				"file_path": filePath,
			}).Warn("Failed to delete photo file")
		} else {
			logrus.WithFields(logrus.Fields{
				"part_id": id,
				"photo":   part.Photo,
				"file_path": filePath,
			}).Info("Photo file deleted successfully")
		}
	} else {
		logrus.WithField("part_id", id).Info("No photo to delete for part")
	}

	if err := r.db.WithContext(ctx).Delete(&Part{}, id).Error; err != nil {
		logrus.WithError(err).WithField("part_id", id).Error("Failed to delete part from database")
		return fmt.Errorf("failed to delete part %d from database: %w", id, err)
	}

	logrus.WithField("part_id", id).Info("Part deleted from database successfully")
	return nil
}
