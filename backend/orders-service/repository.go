package main

import (
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// OrderRepository определяет интерфейс для работы с заказами
type OrderRepository interface {
	Create(order *Order) error
	CreateItem(item *OrderItem) error
	FindByID(id uint) (*Order, error)
	FindWithItemsByID(id uint) (*Order, error)
	FindAll() ([]Order, error)
	FindActive() ([]Order, error)
	FindWithItems() ([]Order, error)
	FindOrderItem(orderID, partID uint) (*OrderItem, error)
	Update(id uint, updates map[string]interface{}) error
	UpdateStatus(id uint, status string) error
	UpdateItem(item *OrderItem) error
	Delete(id uint) error
	DeleteItemsByOrderID(orderID uint) error
	MarkExpiredAsAutoDeleted(before time.Time) error
}

// PartRepositoryForOrders определяет интерфейс для работы с запчастями (для orders-service)
type PartRepositoryForOrders interface {
	FindByID(id uint) (*Part, error)
	UpdateQuantity(id uint, newQuantity int) error
	DecreaseQuantity(id uint, amount int) error
	IncreaseQuantity(id uint, amount int) error
	DeletePart(id uint) error
}

// orderRepository реализует OrderRepository
type orderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) OrderRepository {
	return &orderRepository{db: db}
}

func (r *orderRepository) Create(order *Order) error {
	return r.db.Create(order).Error
}

func (r *orderRepository) FindByID(id uint) (*Order, error) {
	var order Order
	err := r.db.First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindWithItemsByID(id uint) (*Order, error) {
	var order Order
	err := r.db.Preload("Items").First(&order, id).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) FindAll() ([]Order, error) {
	var orders []Order
	err := r.db.Where("auto_deleted = ?", false).Find(&orders).Error
	return orders, err
}

func (r *orderRepository) FindWithItems() ([]Order, error) {
	var orders []Order
	err := r.db.Preload("Items").Where("auto_deleted = ?", false).Find(&orders).Error
	return orders, err
}

func (r *orderRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&Order{}).Where("id = ?", id).Updates(updates).Error
}

func (r *orderRepository) Delete(id uint) error {
	return r.db.Delete(&Order{}, id).Error
}

func (r *orderRepository) MarkExpiredAsAutoDeleted(before time.Time) error {
	return r.db.Model(&Order{}).Where("created_at <= ? AND auto_deleted = ?", before, false).Update("auto_deleted", true).Error
}

func (r *orderRepository) CreateItem(item *OrderItem) error {
	return r.db.Create(item).Error
}

func (r *orderRepository) FindActive() ([]Order, error) {
	var orders []Order
	err := r.db.Preload("Items").Where("auto_deleted = ?", false).Find(&orders).Error
	return orders, err
}

func (r *orderRepository) FindOrderItem(orderID, partID uint) (*OrderItem, error) {
	var item OrderItem
	err := r.db.Where("order_id = ? AND part_id = ?", orderID, partID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *orderRepository) UpdateStatus(id uint, status string) error {
	return r.db.Model(&Order{}).Where("id = ?", id).Update("status", status).Error
}

func (r *orderRepository) UpdateItem(item *OrderItem) error {
	return r.db.Save(item).Error
}

func (r *orderRepository) DeleteItemsByOrderID(orderID uint) error {
	return r.db.Where("order_id = ?", orderID).Delete(&OrderItem{}).Error
}

// partRepositoryForOrders реализует PartRepositoryForOrders
type partRepositoryForOrders struct {
	db *gorm.DB
}

func NewPartRepositoryForOrders(db *gorm.DB) PartRepositoryForOrders {
	return &partRepositoryForOrders{db: db}
}

func (r *partRepositoryForOrders) FindByID(id uint) (*Part, error) {
	var part Part
	err := r.db.First(&part, id).Error
	if err != nil {
		return nil, err
	}
	return &part, nil
}

func (r *partRepositoryForOrders) UpdateQuantity(id uint, newQuantity int) error {
	return r.db.Model(&Part{}).Where("id = ?", id).Update("quantity", newQuantity).Error
}

func (r *partRepositoryForOrders) DecreaseQuantity(id uint, amount int) error {
	return r.db.Model(&Part{}).Where("id = ?", id).Update("quantity", r.db.Raw("CASE WHEN quantity - ? <= 0 THEN -1 ELSE quantity - ? END", amount, amount)).Error
}

func (r *partRepositoryForOrders) IncreaseQuantity(id uint, amount int) error {
	return r.db.Model(&Part{}).Where("id = ?", id).Update("quantity", r.db.Raw("quantity + ?", amount)).Error
}

func (r *partRepositoryForOrders) DeletePart(id uint) error {
	logrus.WithField("part_id", id).Info("Starting part deletion")

	// Получаем запчасть для удаления фото
	var part Part
	if err := r.db.First(&part, id).Error; err != nil {
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

	if err := r.db.Delete(&Part{}, id).Error; err != nil {
		logrus.WithError(err).WithField("part_id", id).Error("Failed to delete part from database")
		return err
	}

	logrus.WithField("part_id", id).Info("Part deleted from database successfully")
	return nil
}
