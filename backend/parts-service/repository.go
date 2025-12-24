package main

import (
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
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
	MarkForDeletion(id uint, deleteAt time.Time) error        // Отмечает для удаления
	DeleteExpiredParts(before time.Time) error                // Удаляет просроченные
	GetStatistics() (StatisticsResponse, error)               // Получает статистику
	BulkDelete(ids []uint) error                              // Массовое удаление
	BulkUpdate(updates []map[string]interface{}) (int, error) // Массовое обновление

	// DeleteZeroQuantityPartsBySupplier Supplier operations
	DeleteZeroQuantityPartsBySupplier(supplierCode string) (int64, error) // Удаляет запчасти с нулевым количеством по поставщику
	GetSupplierCodes() ([]string, error)                                  // Получает уникальные коды поставщиков

	// Earnings operations
	GetTotalEarnings() (float64, error)     // Получает общий заработок
	UpdateTotalEarnings(amount float64) error // Обновляет общий заработок
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
				query = query.Where("photos IS NOT NULL AND jsonb_array_length(photos) > 0")
			} else {
				query = query.Where("(photos IS NULL OR jsonb_array_length(photos) = 0)")
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

	// Проверяем существование столбца price
	var hasPriceColumn bool
	checkQuery := `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='parts' AND column_name='price')`
	err := r.db.Raw(checkQuery).Scan(&hasPriceColumn).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to check if price column exists")
		return StatisticsResponse{}, err
	}
	logrus.WithField("has_price_column", hasPriceColumn).Info("Checked price column existence")

	// Получаем totals
	type Totals struct {
		TotalParts int     `json:"total_parts"`
		TotalValue float64 `json:"total_value"`
	}
	var totals Totals
	totalsQuery := `
		SELECT COUNT(*) as total_parts, COALESCE(SUM(price * quantity), 0) as total_value
		FROM parts
		WHERE to_delete_at IS NULL AND quantity >= 1
	`
	err = r.db.Raw(totalsQuery).Scan(&totals).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to get totals")
		return StatisticsResponse{}, err
	}
	stats.TotalParts = totals.TotalParts
	stats.TotalValue = totals.TotalValue

	// Получаем categories
	categoriesQuery := `
		SELECT category as name, COUNT(*) as count
		FROM parts
		WHERE to_delete_at IS NULL AND quantity >= 1 AND category != ''
		GROUP BY category
		ORDER BY count DESC
	`
	var categories []CategoryCount
	err = r.db.Raw(categoriesQuery).Scan(&categories).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to get categories")
		return StatisticsResponse{}, err
	}
	stats.Categories = categories

	logrus.WithFields(logrus.Fields{
		"total_parts": stats.TotalParts,
		"total_value": stats.TotalValue,
		"categories_count": len(stats.Categories),
	}).Info("Statistics retrieved successfully")

	return stats, nil
}

// BulkDelete удаляет несколько запчастей
func (r *partRepository) BulkDelete(ids []uint) error {
	logrus.WithFields(logrus.Fields{
		"ids":   ids,
		"count": len(ids),
	}).Info("PartRepository.BulkDelete: Starting bulk delete")

	// Используем транзакцию для атомарности
	tx := r.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	err := tx.Where("id IN ?", ids).Delete(&Part{}).Error
	if err != nil {
		logrus.WithError(err).Error("PartRepository.BulkDelete: Failed to execute delete query")
		tx.Rollback()
		return err
	}

	if err := tx.Commit().Error; err != nil {
		logrus.WithError(err).Error("PartRepository.BulkDelete: Failed to commit transaction")
		return err
	}

	logrus.Info("PartRepository.BulkDelete: Successfully completed bulk delete")
	return nil
}

// BulkUpdate обновляет несколько запчастей
func (r *partRepository) BulkUpdate(updates []map[string]interface{}) (int, error) {
	logrus.WithFields(logrus.Fields{
		"updates": updates,
		"count":   len(updates),
	}).Info("PartRepository.BulkUpdate: Starting bulk update")

	updatedCount := 0
	for i, update := range updates {
		var id uint
		if idVal, exists := update["id"]; exists {
			switch v := idVal.(type) {
			case float64:
				id = uint(v)
			case int:
				id = uint(v)
			case uint:
				id = v
			default:
				logrus.WithFields(logrus.Fields{
					"index": i,
					"idVal": idVal,
					"type": fmt.Sprintf("%T", idVal),
				}).Error("PartRepository.BulkUpdate: Invalid id type")
				return updatedCount, fmt.Errorf("update at index %d has invalid id type", i)
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"index": i,
				"update": update,
			}).Error("PartRepository.BulkUpdate: Update missing 'id' field")
			return updatedCount, fmt.Errorf("update at index %d missing 'id' field", i)
		}

		delete(update, "id")
		logrus.WithFields(logrus.Fields{
			"index": i,
			"id": id,
			"update": update,
		}).Debug("PartRepository.BulkUpdate: Processing update")

		if err := r.Update(id, update); err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"index": i,
				"id": id,
			}).Error("PartRepository.BulkUpdate: Failed to update part")
			return updatedCount, err
		}
		updatedCount++
	}

	logrus.Info("PartRepository.BulkUpdate: Successfully completed bulk update")
	return updatedCount, nil
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

// GetTotalEarnings получает общий заработок из базы данных
func (r *partRepository) GetTotalEarnings() (float64, error) {
	var earnings Earnings
	err := r.db.First(&earnings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Если записи нет, создаем новую с нулевым значением
			newEarnings := Earnings{TotalAmount: 0}
			if createErr := r.db.Create(&newEarnings).Error; createErr != nil {
				return 0, createErr
			}
			return 0, nil
		}
		return 0, err
	}
	return earnings.TotalAmount, nil
}

// UpdateTotalEarnings обновляет общий заработок в базе данных
func (r *partRepository) UpdateTotalEarnings(amount float64) error {
	var earnings Earnings
	err := r.db.First(&earnings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Создаем новую запись
			newEarnings := Earnings{TotalAmount: amount}
			return r.db.Create(&newEarnings).Error
		}
		return err
	}

	// Обновляем существующую запись
	return r.db.Model(&earnings).Update("total_amount", amount).Error
}
