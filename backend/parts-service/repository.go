package main

import (
	"fmt"
	"time"

	"gorm.io/gorm"
)

// PartRepository определяет контракт для доступа к данным запчастей
// Это позволяет легко заменять реализацию (например, на другую БД)

// PartRepository определяет интерфейс для работы с запчастями в базе данных
// Это позволяет легко заменять реализацию (например, на другую БД)
type PartRepository interface {
	// Create Основные операции CRUD
	Create(part *Part) error                              // Создает новую запчасть
	FindByID(id uint) (*Part, error)                      // Находит запчасть по ID
	Update(id uint, updates map[string]interface{}) error // Обновляет запчасть
	Delete(id uint) error                                 // Удаляет запчасть

	// FindAll Поиск и фильтрация
	FindAll(query *gorm.DB) ([]Part, error)                         // Находит все запчасти по запросу
	FindWithFilters(filters map[string]interface{}) ([]Part, error) // Находит с фильтрами

	// MarkForDeletion Специфические операции
	MarkForDeletion(id uint, deleteAt time.Time) error // Отмечает для удаления
	DeleteExpiredParts(before time.Time) error         // Удаляет просроченные
	GetStatistics() (StatisticsResponse, error)        // Получает статистику
	BulkDelete(ids []uint) error                       // Массовое удаление
	BulkUpdate(updates []map[string]interface{}) error // Массовое обновление

	// DeleteZeroQuantityPartsBySupplier Supplier operations
	DeleteZeroQuantityPartsBySupplier(supplierCode string) (int64, error) // Удаляет запчасти с нулевым количеством по поставщику
	GetSupplierCodes() ([]string, error)                                  // Получает уникальные коды поставщиков
}

// partRepository реализует PartRepository
type partRepository struct {
	db *gorm.DB
}

// NewPartRepository создает новый экземпляр репозитория
func NewPartRepository(db *gorm.DB) PartRepository {
	return &partRepository{db: db}
}

// Create создает новую запчасть
func (r *partRepository) Create(part *Part) error {
	return r.db.Create(part).Error
}

// FindByID находит запчасть по ID
func (r *partRepository) FindByID(id uint) (*Part, error) {
	var part Part
	err := r.db.First(&part, id).Error
	if err != nil {
		return nil, err
	}
	return &part, nil
}

// Update обновляет запчасть по ID
func (r *partRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&Part{}).Where("id = ?", id).Updates(updates).Error
}

// Delete удаляет запчасть по ID
func (r *partRepository) Delete(id uint) error {
	return r.db.Delete(&Part{}, id).Error
}

// FindAll находит все запчасти с учетом запроса
func (r *partRepository) FindAll(query *gorm.DB) ([]Part, error) {
	var parts []Part
	err := query.Find(&parts).Error
	return parts, err
}

// FindWithFilters находит запчасти с фильтрами
func (r *partRepository) FindWithFilters(filters map[string]interface{}) ([]Part, error) {
	query := r.db.Model(&Part{})

	// Применяем фильтры
	for key, value := range filters {
		switch key {
		case "ids_in":
			query = query.Where("id IN ?", value)
		case "to_delete_at_is_null":
			if value.(bool) {
				query = query.Where("to_delete_at IS NULL")
			}
		case "quantity_gte":
			query = query.Where("quantity >= ?", value)
		case "category_ilike":
			query = query.Where("category ILIKE ?", "%"+value.(string)+"%")
		case "brand_ilike":
			query = query.Where("brand ILIKE ?", "%"+value.(string)+"%")
		case "model_ilike":
			query = query.Where("model ILIKE ?", "%"+value.(string)+"%")
		case "location_ilike":
			query = query.Where("location ILIKE ?", "%"+value.(string)+"%")
		case "salesman_ilike":
			query = query.Where("salesman ILIKE ?", "%"+value.(string)+"%")
		case "status":
			query = query.Where("status = ?", value)
		case "has_photo":
			if value.(bool) {
				query = query.Where("photo IS NOT NULL AND photo != ''")
			} else {
				query = query.Where("(photo IS NULL OR photo = '')")
			}
		case "search":
			searchTerm := "%" + value.(string) + "%"
			query = query.Where("name ILIKE ? OR description ILIKE ?", searchTerm, searchTerm)
		}
	}

	var parts []Part
	err := query.Find(&parts).Error
	return parts, err
}

// MarkForDeletion отмечает запчасть для удаления
func (r *partRepository) MarkForDeletion(id uint, deleteAt time.Time) error {
	return r.db.Model(&Part{}).Where("id = ?", id).Update("to_delete_at", deleteAt).Error
}

// DeleteExpiredParts удаляет просроченные запчасти
func (r *partRepository) DeleteExpiredParts(before time.Time) error {
	return r.db.Where("to_delete_at IS NOT NULL AND to_delete_at <= ?", before).Delete(&Part{}).Error
}

// GetStatistics получает статистику по запчастям
func (r *partRepository) GetStatistics() (StatisticsResponse, error) {
	var stats StatisticsResponse

	// Общее количество запчастей
	var totalParts int64
	r.db.Model(&Part{}).Where("to_delete_at IS NULL AND quantity >= 1").Count(&totalParts)
	stats.TotalParts = int(totalParts)

	// Общая стоимость
	r.db.Model(&Part{}).Where("to_delete_at IS NULL AND quantity >= 1").Select("COALESCE(SUM(price * quantity), 0)").Scan(&stats.TotalValue)

	// Категории
	r.db.Model(&Part{}).Select("category as name, COUNT(*) as count").Where("category != '' AND to_delete_at IS NULL AND quantity >= 1").Group("category").Scan(&stats.Categories)

	return stats, r.db.Error
}

// BulkDelete удаляет несколько запчастей
func (r *partRepository) BulkDelete(ids []uint) error {
	return r.db.Where("id IN ?", ids).Delete(&Part{}).Error
}

// BulkUpdate обновляет несколько запчастей
func (r *partRepository) BulkUpdate(updates []map[string]interface{}) error {
	for _, update := range updates {
		if id, ok := update["id"].(uint); ok {
			delete(update, "id")
			if err := r.Update(id, update); err != nil {
				return err
			}
		}
	}
	return nil
}

// DeleteZeroQuantityPartsBySupplier удаляет запчасти с нулевым количеством по коду поставщика
func (r *partRepository) DeleteZeroQuantityPartsBySupplier(supplierCode string) (int64, error) {
	fmt.Printf("Repository: DeleteZeroQuantityPartsBySupplier called with supplier_code='%s'\n", supplierCode)

	// Сначала посчитаем, сколько записей будет удалено
	var count int64
	err := r.db.Model(&Part{}).Where("supplier_code = ? AND quantity = 0 AND to_delete_at IS NULL", supplierCode).Count(&count).Error
	if err != nil {
		fmt.Printf("Repository: Error counting parts to delete: %v\n", err)
		return 0, err
	}
	fmt.Printf("Repository: Found %d parts to delete for supplier_code='%s'\n", count, supplierCode)

	// Выполним удаление
	result := r.db.Where("supplier_code = ? AND quantity = 0 AND to_delete_at IS NULL", supplierCode).Delete(&Part{})
	if result.Error != nil {
		fmt.Printf("Repository: Error deleting parts: %v\n", result.Error)
		return 0, result.Error
	}

	fmt.Printf("Repository: Successfully deleted %d parts for supplier_code='%s'\n", result.RowsAffected, supplierCode)
	return result.RowsAffected, nil
}

// GetSupplierCodes получает уникальные коды поставщиков, для которых есть запчасти с quantity=0
func (r *partRepository) GetSupplierCodes() ([]string, error) {
	fmt.Printf("Repository: GetSupplierCodes called\n")
	var codes []string
	err := r.db.Model(&Part{}).Where("supplier_code IS NOT NULL AND supplier_code != '' AND quantity = 0 AND to_delete_at IS NULL").Distinct("supplier_code").Pluck("supplier_code", &codes).Error
	if err != nil {
		fmt.Printf("Repository: GetSupplierCodes failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("Repository: GetSupplierCodes found %d supplier codes with zero quantity parts\n", len(codes))
	return codes, nil
}
