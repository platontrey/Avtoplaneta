package main

import (
	"context"
	"fmt"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

// PartRepository определяет контракт для доступа к данным запчастей
// Это позволяет легко заменять реализацию (например, на другую БД)
type PartRepository interface {
	// Create Основные операции CRUD
	Create(ctx context.Context, part *Part) error                              // Создает новую запчасть
	FindByID(ctx context.Context, id uint) (*Part, error)                      // Находит запчасть по ID
	Update(ctx context.Context, id uint, updates map[string]interface{}) error // Обновляет запчасть
	Delete(ctx context.Context, id uint) error                                 // Удаляет запчасть

	// FindAll Поиск и фильтрация
	FindAll(ctx context.Context, query *gorm.DB) ([]Part, error)                         // Находит все запчасти по запросу
	FindWithFilters(ctx context.Context, filters map[string]interface{}) ([]Part, error) // Находит с фильтрами

	// MarkForDeletion Специфические операции
	MarkForDeletion(ctx context.Context, id uint, deleteAt time.Time) error        // Отмечает для удаления
	DeleteExpiredParts(ctx context.Context, before time.Time) error                // Удаляет просроченные
	GetStatistics(ctx context.Context) (StatisticsResponse, error)                 // Получает статистику
	BulkDelete(ctx context.Context, ids []uint) error                              // Массовое удаление
	BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error) // Массовое обновление

	// DeleteZeroQuantityPartsBySupplier Supplier operations
	DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) // Удаляет запчасти с нулевым количеством по поставщику
	GetSupplierCodes(ctx context.Context) ([]string, error)                                    // Получает уникальные коды поставщиков

	// Earnings operations
	GetTotalEarnings(ctx context.Context) (float64, error)         // Получает общий заработок
	UpdateTotalEarnings(ctx context.Context, amount float64) error // Обновляет общий заработок
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
func (r *partRepository) Create(ctx context.Context, part *Part) error {
	return r.db.WithContext(ctx).Create(part).Error
}

// FindByID находит запчасть по ID
func (r *partRepository) FindByID(ctx context.Context, id uint) (*Part, error) {
	var part Part
	err := r.db.WithContext(ctx).First(&part, id).Error
	if err != nil {
		return nil, err
	}
	return &part, nil
}

// Update обновляет запчасть по ID
func (r *partRepository) Update(ctx context.Context, id uint, updates map[string]interface{}) error {
	return r.db.WithContext(ctx).Model(&Part{}).Where("id = ?", id).Updates(updates).Error
}

// Delete удаляет запчасть по ID
func (r *partRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&Part{}, id).Error
}

// FindAll находит все запчасти с учетом запроса
func (r *partRepository) FindAll(ctx context.Context, query *gorm.DB) ([]Part, error) {
	var parts []Part
	err := query.WithContext(ctx).Find(&parts).Error
	return parts, err
}

// FindWithFilters находит запчасти с фильтрами
func (r *partRepository) FindWithFilters(ctx context.Context, filters map[string]interface{}) ([]Part, error) {
	query := r.db.WithContext(ctx).Model(&Part{})

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
func (r *partRepository) MarkForDeletion(ctx context.Context, id uint, deleteAt time.Time) error {
	return r.db.WithContext(ctx).Model(&Part{}).Where("id = ?", id).Update("to_delete_at", deleteAt).Error
}

// DeleteExpiredParts удаляет просроченные запчасти
func (r *partRepository) DeleteExpiredParts(ctx context.Context, before time.Time) error {
	return r.db.WithContext(ctx).Where("to_delete_at IS NOT NULL AND to_delete_at <= ?", before).Delete(&Part{}).Error
}

// GetStatistics получает статистику по запчастям
func (r *partRepository) GetStatistics(ctx context.Context) (StatisticsResponse, error) {
	var stats StatisticsResponse

	// Проверяем существование столбца price
	var hasPriceColumn bool
	checkQuery := `SELECT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='parts' AND column_name='price')`
	err := r.db.WithContext(ctx).Raw(checkQuery).Scan(&hasPriceColumn).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to check if price column exists")
		return StatisticsResponse{}, err
	}
	logrus.WithField("has_price_column", hasPriceColumn).Info("Checked price column existence")

	// Получаем totals
	type Totals struct {
		TotalParts    int     `json:"total_parts"`
		TotalQuantity int     `json:"total_quantity"`
		TotalValue    float64 `json:"total_value"`
	}
	var totals Totals
	totalsQuery, args, err := squirrel.Select("COUNT(*) as total_parts", "COALESCE(SUM(quantity), 0) as total_quantity", "COALESCE(SUM(price * quantity), 0) as total_value").
		From("parts").
		Where(squirrel.Eq{"to_delete_at": nil}).
		Where(squirrel.GtOrEq{"quantity": 1}).
		ToSql()
	if err != nil {
		logrus.WithError(err).Error("Failed to build totals query")
		return StatisticsResponse{}, err
	}
	err = r.db.WithContext(ctx).Raw(totalsQuery, args...).Scan(&totals).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to get totals")
		return StatisticsResponse{}, err
	}
	stats.TotalParts = totals.TotalParts
	stats.TotalQuantity = totals.TotalQuantity
	stats.TotalValue = totals.TotalValue

	// Получаем categories
	categoriesQuery, args, err := squirrel.Select("category as name", "COUNT(*) as count").
		From("parts").
		Where(squirrel.Eq{"to_delete_at": nil}).
		Where(squirrel.GtOrEq{"quantity": 1}).
		Where(squirrel.NotEq{"category": ""}).
		GroupBy("category").
		OrderBy("count DESC").
		ToSql()
	if err != nil {
		logrus.WithError(err).Error("Failed to build categories query")
		return StatisticsResponse{}, err
	}
	var categories []CategoryCount
	err = r.db.WithContext(ctx).Raw(categoriesQuery, args...).Scan(&categories).Error
	if err != nil {
		logrus.WithError(err).Error("Failed to get categories")
		return StatisticsResponse{}, err
	}
	stats.Categories = categories

	logrus.WithFields(logrus.Fields{
		"total_parts":      stats.TotalParts,
		"total_value":      stats.TotalValue,
		"categories_count": len(stats.Categories),
	}).Info("Statistics retrieved successfully")

	return stats, nil
}

// BulkDelete удаляет несколько запчастей
func (r *partRepository) BulkDelete(ctx context.Context, ids []uint) error {
	logrus.WithFields(logrus.Fields{
		"ids":   ids,
		"count": len(ids),
	}).Info("PartRepository.BulkDelete: Starting bulk delete")

	// Используем транзакцию для атомарности
	tx := r.db.WithContext(ctx).Begin()
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

// BulkUpdate обновляет несколько запчастей с использованием batch-операций
func (r *partRepository) BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error) {
	logrus.WithFields(logrus.Fields{
		"updates": updates,
		"count":   len(updates),
	}).Info("PartRepository.BulkUpdate: Starting bulk update")

	if len(updates) == 0 {
		return 0, nil
	}

	// Используем транзакцию для атомарности
	tx := r.db.WithContext(ctx).Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Создаем временную таблицу для batch update
	tempTableName := "temp_parts_update"
	createTempTable := fmt.Sprintf(`
		CREATE TEMP TABLE %s (
			id BIGINT PRIMARY KEY,
			name TEXT,
			quantity INTEGER,
			description TEXT,
			category TEXT,
			price DECIMAL(10,2),
			brand TEXT,
			model TEXT,
			location TEXT,
			salesman TEXT,
			status TEXT,
			photos JSONB,
			body_brand TEXT,
			engine_brand TEXT,
			car_release_date TEXT,
			front_rear TEXT,
			left_right TEXT,
			top_bottom TEXT,
			number TEXT,
			manufacturer TEXT,
			manufacturer_code TEXT,
			oem_code TEXT,
			color TEXT,
			condition TEXT,
			supplier_code TEXT,
			defect TEXT,
			transmission TEXT,
			drive TEXT,
			wear_percentage DECIMAL(5,2),
			season TEXT,
			diameter TEXT,
			width TEXT,
			profile TEXT,
			tire_quantity INTEGER,
			drilling TEXT,
			offset TEXT,
			center_hole_diameter TEXT,
			tire_model TEXT
		) ON COMMIT DROP
	`, tempTableName)

	if err := tx.Exec(createTempTable).Error; err != nil {
		logrus.WithError(err).Error("PartRepository.BulkUpdate: Failed to create temp table")
		tx.Rollback()
		return 0, err
	}

	// Вставляем данные во временную таблицу
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
					"type":  fmt.Sprintf("%T", idVal),
				}).Error("PartRepository.BulkUpdate: Invalid id type")
				tx.Rollback()
				return 0, fmt.Errorf("update at index %d has invalid id type", i)
			}
		} else {
			logrus.WithFields(logrus.Fields{
				"index":  i,
				"update": update,
			}).Error("PartRepository.BulkUpdate: Update missing 'id' field")
			tx.Rollback()
			return 0, fmt.Errorf("update at index %d missing 'id' field", i)
		}

		delete(update, "id")

		// Вставляем в temp таблицу
		insertSQL := fmt.Sprintf(`
			INSERT INTO %s (id, name, quantity, description, category, price, brand, model, location, salesman, status, photos,
				body_brand, engine_brand, car_release_date, front_rear, left_right, top_bottom, number, manufacturer,
				manufacturer_code, oem_code, color, condition, supplier_code, defect, transmission, drive, wear_percentage,
				season, diameter, width, profile, tire_quantity, drilling, offset, center_hole_diameter, tire_model)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, tempTableName)

		if err := tx.Exec(insertSQL,
			id,
			update["name"],
			update["quantity"],
			update["description"],
			update["category"],
			update["price"],
			update["brand"],
			update["model"],
			update["location"],
			update["salesman"],
			update["status"],
			update["photos"],
			update["body_brand"],
			update["engine_brand"],
			update["car_release_date"],
			update["front_rear"],
			update["left_right"],
			update["top_bottom"],
			update["number"],
			update["manufacturer"],
			update["manufacturer_code"],
			update["oem_code"],
			update["color"],
			update["condition"],
			update["supplier_code"],
			update["defect"],
			update["transmission"],
			update["drive"],
			update["wear_percentage"],
			update["season"],
			update["diameter"],
			update["width"],
			update["profile"],
			update["tire_quantity"],
			update["drilling"],
			update["offset"],
			update["center_hole_diameter"],
			update["tire_model"],
		).Error; err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"index": i,
				"id":    id,
			}).Error("PartRepository.BulkUpdate: Failed to insert into temp table")
			tx.Rollback()
			return 0, err
		}
	}

	// Выполняем batch update из temp таблицы
	updateSQL := fmt.Sprintf(`
		UPDATE parts
		SET
			name = COALESCE(t.name, parts.name),
			quantity = COALESCE(t.quantity, parts.quantity),
			description = COALESCE(t.description, parts.description),
			category = COALESCE(t.category, parts.category),
			price = COALESCE(t.price, parts.price),
			brand = COALESCE(t.brand, parts.brand),
			model = COALESCE(t.model, parts.model),
			location = COALESCE(t.location, parts.location),
			salesman = COALESCE(t.salesman, parts.salesman),
			status = COALESCE(t.status, parts.status),
			photos = COALESCE(t.photos, parts.photos),
			body_brand = COALESCE(t.body_brand, parts.body_brand),
			engine_brand = COALESCE(t.engine_brand, parts.engine_brand),
			car_release_date = COALESCE(t.car_release_date, parts.car_release_date),
			front_rear = COALESCE(t.front_rear, parts.front_rear),
			left_right = COALESCE(t.left_right, parts.left_right),
			top_bottom = COALESCE(t.top_bottom, parts.top_bottom),
			number = COALESCE(t.number, parts.number),
			manufacturer = COALESCE(t.manufacturer, parts.manufacturer),
			manufacturer_code = COALESCE(t.manufacturer_code, parts.manufacturer_code),
			oem_code = COALESCE(t.oem_code, parts.oem_code),
			color = COALESCE(t.color, parts.color),
			condition = COALESCE(t.condition, parts.condition),
			supplier_code = COALESCE(t.supplier_code, parts.supplier_code),
			defect = COALESCE(t.defect, parts.defect),
			transmission = COALESCE(t.transmission, parts.transmission),
			drive = COALESCE(t.drive, parts.drive),
			wear_percentage = COALESCE(t.wear_percentage, parts.wear_percentage),
			season = COALESCE(t.season, parts.season),
			diameter = COALESCE(t.diameter, parts.diameter),
			width = COALESCE(t.width, parts.width),
			profile = COALESCE(t.profile, parts.profile),
			tire_quantity = COALESCE(t.tire_quantity, parts.tire_quantity),
			drilling = COALESCE(t.drilling, parts.drilling),
			offset = COALESCE(t.offset, parts.offset),
			center_hole_diameter = COALESCE(t.center_hole_diameter, parts.center_hole_diameter),
			tire_model = COALESCE(t.tire_model, parts.tire_model),
			updated_at = NOW()
		FROM %s t
		WHERE parts.id = t.id
	`, tempTableName)

	if err := tx.Exec(updateSQL).Error; err != nil {
		logrus.WithError(err).Error("PartRepository.BulkUpdate: Failed to execute batch update")
		tx.Rollback()
		return 0, err
	}

	// Получаем количество обновленных строк
	var updatedCount int64
	countSQL := fmt.Sprintf("SELECT COUNT(*) FROM %s", tempTableName)
	if err := tx.Raw(countSQL).Scan(&updatedCount).Error; err != nil {
		logrus.WithError(err).Error("PartRepository.BulkUpdate: Failed to get updated count")
		tx.Rollback()
		return 0, err
	}

	if err := tx.Commit().Error; err != nil {
		logrus.WithError(err).Error("PartRepository.BulkUpdate: Failed to commit transaction")
		return 0, err
	}

	logrus.WithField("updated_count", updatedCount).Info("PartRepository.BulkUpdate: Successfully completed bulk update")
	return int(updatedCount), nil
}

// DeleteZeroQuantityPartsBySupplier удаляет запчасти с нулевым количеством по коду поставщика
func (r *partRepository) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	fmt.Printf("Repository: DeleteZeroQuantityPartsBySupplier called with supplier_code='%s'\n", supplierCode)

	// Сначала посчитаем, сколько записей будет удалено
	var count int64
	err := r.db.WithContext(ctx).Model(&Part{}).Where("supplier_code = ? AND quantity = 0 AND to_delete_at IS NULL", supplierCode).Count(&count).Error
	if err != nil {
		fmt.Printf("Repository: Error counting parts to delete: %v\n", err)
		return 0, err
	}
	fmt.Printf("Repository: Found %d parts to delete for supplier_code='%s'\n", count, supplierCode)

	// Выполним удаление
	result := r.db.WithContext(ctx).Where("supplier_code = ? AND quantity = 0 AND to_delete_at IS NULL", supplierCode).Delete(&Part{})
	if result.Error != nil {
		fmt.Printf("Repository: Error deleting parts: %v\n", result.Error)
		return 0, result.Error
	}

	fmt.Printf("Repository: Successfully deleted %d parts for supplier_code='%s'\n", result.RowsAffected, supplierCode)
	return result.RowsAffected, nil
}

// GetSupplierCodes получает уникальные коды поставщиков, для которых есть запчасти с quantity=0
func (r *partRepository) GetSupplierCodes(ctx context.Context) ([]string, error) {
	fmt.Printf("Repository: GetSupplierCodes called\n")
	var codes []string
	err := r.db.WithContext(ctx).Model(&Part{}).Where("supplier_code IS NOT NULL AND supplier_code != '' AND quantity = 0 AND to_delete_at IS NULL").Distinct("supplier_code").Pluck("supplier_code", &codes).Error
	if err != nil {
		fmt.Printf("Repository: GetSupplierCodes failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("Repository: GetSupplierCodes found %d supplier codes with zero quantity parts\n", len(codes))
	return codes, nil
}

// GetTotalEarnings получает общий заработок из базы данных
func (r *partRepository) GetTotalEarnings(ctx context.Context) (float64, error) {
	var earnings Earnings
	err := r.db.WithContext(ctx).First(&earnings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Если записи нет, создаем новую с нулевым значением
			newEarnings := Earnings{TotalAmount: 0}
			if createErr := r.db.WithContext(ctx).Create(&newEarnings).Error; createErr != nil {
				return 0, createErr
			}
			return 0, nil
		}
		return 0, err
	}
	return earnings.TotalAmount, nil
}

// UpdateTotalEarnings обновляет общий заработок в базе данных
func (r *partRepository) UpdateTotalEarnings(ctx context.Context, amount float64) error {
	var earnings Earnings
	err := r.db.WithContext(ctx).First(&earnings).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Создаем новую запись
			newEarnings := Earnings{TotalAmount: amount}
			return r.db.WithContext(ctx).Create(&newEarnings).Error
		}
		return err
	}

	// Обновляем существующую запись
	return r.db.WithContext(ctx).Model(&earnings).Update("total_amount", amount).Error
}
