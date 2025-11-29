package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// Обработчики инвентаря
func getInventory(c *gin.Context) {
	// Delete expired parts (older than 14 days)
	now := time.Now()
	fourteenDaysAgo := now.AddDate(0, 0, -14)
	if err := db.Where("to_delete_at IS NOT NULL AND to_delete_at <= ?", fourteenDaysAgo).Delete(&Part{}).Error; err != nil {
		fmt.Printf("Error deleting expired parts: %v\n", err)
		// Continue anyway, don't fail the request
	}

	// Parse query parameters for filtering
	search := c.Query("search")
	category := c.Query("category")
	brand := c.Query("brand")
	model := c.Query("model")
	location := c.Query("location")
	salesman := c.Query("salesman")
	status := c.Query("status")
	hasPhoto := c.Query("hasPhoto")

	// Check if we have any search parameters that would benefit from Elasticsearch
	useElasticsearch := search != "" || category != "" || brand != "" || model != "" || location != "" || salesman != "" || status != "" || hasPhoto != ""

	fmt.Printf("getInventory: search='%s', category='%s', brand='%s', model='%s', location='%s', salesman='%s', status='%s', hasPhoto='%s', useElasticsearch=%v\n", search, category, brand, model, location, salesman, status, hasPhoto, useElasticsearch)
	fmt.Printf("getInventory: Request URL: %s\n", c.Request.URL.String())
	fmt.Printf("getInventory: Request headers: %+v\n", c.Request.Header)

	if useElasticsearch {
		// Use Elasticsearch for search
		esQuery := BuildSearchQuery(search, category, brand, model, location, salesman, status, hasPhoto)
		fmt.Printf("getInventory: Elasticsearch query: %+v\n", esQuery)

		// Parse pagination parameters
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
		from := (page - 1) * limit

		esParts, _, err := SearchParts(esQuery, from, limit)
		if err != nil {
			fmt.Printf("Error searching with Elasticsearch: %v\n", err)
			// Fallback to database search if Elasticsearch fails
			fallbackToDatabaseSearch(c, search, category, brand, model, location, salesman, status, hasPhoto)
			return
		}

		fmt.Printf("getInventory: Found %d parts in Elasticsearch\n", len(esParts))

		// Преобразовать запчасти Elasticsearch в обычные запчасти и отфильтровать запчасти, отмеченные для удаления
		parts := make([]Part, 0, len(esParts))
		fmt.Printf("getInventory: Processing %d ES parts\n", len(esParts))

		// Получаем ID всех запчастей
		partIDs := make([]uint, len(esParts))
		for i, esPart := range esParts {
			partIDs[i] = esPart.ID
		}

		// Проверяем to_delete_at для всех запчастей одним запросом и исключаем quantity < 0
		var dbParts []Part
		db.Where("id IN ? AND to_delete_at IS NULL AND quantity >= 0", partIDs).Find(&dbParts)

		// Создаем map для быстрого поиска
		validIDs := make(map[uint]bool)
		for _, dbPart := range dbParts {
			validIDs[dbPart.ID] = true
		}

		// Фильтруем запчасти из Elasticsearch
		for _, esPart := range esParts {
			// Пропускаем части, помеченные для удаления
			if !validIDs[esPart.ID] {
				continue
			}

			part := Part{
				ID:          esPart.ID,
				Name:        esPart.Name,
				Quantity:    esPart.Quantity,
				Description: esPart.Description,
				Category:    esPart.Category,
				Price:       esPart.Price,
				Salesman:    esPart.Salesman,
				Location:    esPart.Location,
				Status:      esPart.Status,
				Brand:       esPart.Brand,
				Model:       esPart.Model,
				Photo:       esPart.Photo,
			}
			parts = append(parts, part)
		}

		fmt.Printf("getInventory: After filtering marked parts: %d remaining\n", len(parts))
		c.JSON(http.StatusOK, parts)
	} else {
		// Use database for simple listing without filters
		fmt.Printf("getInventory: Using database search\n")
		fallbackToDatabaseSearch(c, search, category, brand, model, location, salesman, status, hasPhoto)
	}
}

func fallbackToDatabaseSearch(c *gin.Context, search, category, brand, model, location, salesman, status, hasPhoto string) {
	// Parse limit and page parameters
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "0")) // 0 means no limit
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))

	// Построить запрос с фильтрами - исключить запчасти с количеством < 0
	query := db.Where("to_delete_at IS NULL AND quantity >= 0")

	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+search+"%", "%"+search+"%")
	}
	if category != "" {
		query = query.Where("category ILIKE ?", "%"+category+"%")
	}
	if brand != "" {
		query = query.Where("brand ILIKE ?", "%"+brand+"%")
	}
	if model != "" {
		query = query.Where("model ILIKE ?", "%"+model+"%")
	}
	if location != "" {
		query = query.Where("location ILIKE ?", "%"+location+"%")
	}
	if salesman != "" {
		query = query.Where("salesman ILIKE ?", "%"+salesman+"%")
	}
	if status != "" {
		if status == "true" {
			query = query.Where("status = ?", true)
		} else if status == "false" {
			query = query.Where("status = ?", false)
		}
	}
	if hasPhoto != "" {
		if hasPhoto == "with" {
			query = query.Where("photo IS NOT NULL AND photo != ''")
		} else if hasPhoto == "without" {
			query = query.Where("(photo IS NULL OR photo = '')")
		}
		// "all" - не добавляем фильтр, показываем все
	}

	var parts []Part
	var err error
	if limit > 0 {
		offset := (page - 1) * limit
		err = query.Limit(limit).Offset(offset).Find(&parts).Error
	} else {
		err = query.Find(&parts).Error
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить инвентарь"})
		return
	}

	// Заполнить форматированные поля дат для частей из базы данных
	now := time.Now()
	for i := range parts {
		if parts[i].ToDeleteAt != nil {
			parts[i].ToDeleteAtFormatted = parts[i].ToDeleteAt.Format("2006-01-02 15:04:05")
			parts[i].TimeUntilDeletion = formatTimeUntil(*parts[i].ToDeleteAt, now)
		}
	}

	c.JSON(http.StatusOK, parts)
}

func addPart(c *gin.Context) {
	var part Part
	if err := c.ShouldBindJSON(&part); err != nil {
		fmt.Printf("Invalid JSON in addPart: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	// Базовая валидация
	if len(part.Name) > 255 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Название слишком длинное"})
		return
	}
	if len(part.Description) > 1000 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Описание слишком длинное"})
		return
	}
	if part.Quantity < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Количество должно быть не менее 1"})
		return
	}
	if part.Price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Цена не может быть отрицательной"})
		return
	}

	fmt.Printf("Received part data: %+v\n", part)

	if err := db.Create(&part).Error; err != nil {
		fmt.Printf("Error creating part: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add part"})
		return
	}

	// Получить созданную запчасть для возврата её ID
	var createdPart Part
	if err := db.Last(&createdPart).Error; err != nil {
		fmt.Printf("Ошибка получения созданной части: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Часть создана, но не удалось получить ID"})
		return
	}

	// Заполнить форматированные поля дат для новой запчасти
	now := time.Now()
	if createdPart.ToDeleteAt != nil {
		createdPart.ToDeleteAtFormatted = createdPart.ToDeleteAt.Format("2006-01-02 15:04:05")
		createdPart.TimeUntilDeletion = formatTimeUntil(*createdPart.ToDeleteAt, now)
	}

	// Индексировать запчасть в Elasticsearch
	if err := IndexPart(&createdPart); err != nil {
		fmt.Printf("Warning: Failed to index part in Elasticsearch: %v\n", err)
		// Don't fail the request, just log the warning
	} else {
		fmt.Printf("Successfully indexed part %d in Elasticsearch\n", createdPart.ID)
	}

	fmt.Printf("Успешно создана часть с ID: %d\n", createdPart.ID)
	c.JSON(http.StatusCreated, gin.H{
		"message": "Часть добавлена успешно",
		"id":      createdPart.ID,
	})
}

func deletePart(c *gin.Context) {
	id := c.Param("id")
	fmt.Printf("=== DELETE PART START ===\n")
	fmt.Printf("deletePart called for part ID: %s\n", id)

	// Преобразовать id в uint для правильного запроса GORM
	partID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	// Сначала проверить, существует ли запчасть
	var part Part
	if err := db.First(&part, uint(partID)).Error; err != nil {
		fmt.Printf("Часть %s не найдена для удаления: %v\n", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Часть не найдена"})
		return
	}

	// Удалить связанный файл фото, если он существует
	if part.Photo != "" {
		// Извлечь имя файла из пути фото (удалить префикс "/uploads/")
		photoPath := strings.TrimPrefix(part.Photo, "/")
		if err := os.Remove(photoPath); err != nil && !os.IsNotExist(err) {
			// Записать ошибку, но не прерывать удаление - очистка фото не критична
			fmt.Printf("Предупреждение: Не удалось удалить файл фото %s: %v\n", photoPath, err)
		} else if err == nil {
			fmt.Printf("Успешно удален файл фото: %s\n", photoPath)
		}
	}

	if err := db.Delete(&Part{}, uint(partID)).Error; err != nil {
		fmt.Printf("Ошибка удаления части %s: %v\n", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить часть"})
		return
	}

	// Логировать удаление запчасти
	logUserActivity(c, "delete_part", "part", fmt.Sprintf("Deleted part: %s (ID: %s)", part.Name, id), &partID)

	// Удалить запчасть из индекса Elasticsearch
	if err := DeletePartFromIndex(uint(partID)); err != nil {
		fmt.Printf("Warning: Failed to remove part from Elasticsearch index: %v\n", err)
		// Не отклоняйте запрос, просто запишите предупреждение.
	} else {
		fmt.Printf("Successfully removed part %s from Elasticsearch index\n", id)
	}

	fmt.Printf("Успешно удалена часть %s\n", id)
	fmt.Printf("=== DELETE PART END ===\n")
	c.JSON(http.StatusOK, gin.H{"message": "Запчасть удалена успешно"})
}

func updatePart(c *gin.Context) {
	fmt.Printf("=== STARTING UPDATE PART REQUEST ===\n")
	id := c.Param("id")
	fmt.Printf("Part ID from URL: %s\n", id)

	var updates map[string]interface{}
	if err := c.ShouldBindJSON(&updates); err != nil {
		fmt.Printf("Invalid JSON in updatePart: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	fmt.Printf("Received updates for part %s: %+v\n", id, updates)
	rawData, _ := c.GetRawData()
	fmt.Printf("Raw request body: %s\n", string(rawData))

	// Convert id to uint for proper GORM query
	partID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	// Сначала проверить, существует ли запчасть
	var part Part
	if err := db.First(&part, uint(partID)).Error; err != nil {
		fmt.Printf("Часть %s не найдена: %v\n", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Часть не найдена"})
		return
	}
	fmt.Printf("Найдена существующая часть: %+v\n", part)
	fmt.Printf("Existing part details - ID: %d, Name: %s, Quantity: %d, Price: %.2f\n", part.ID, part.Name, part.Quantity, part.Price)

	// Валидировать и конвертировать типы данных
	processedUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case "quantity":
			if qty, ok := value.(float64); ok {
				intQty := int(qty)
				if intQty < 0 {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Количество должно быть не менее 0"})
					return
				}
				processedUpdates[key] = intQty
			} else {
				processedUpdates[key] = value
			}
		case "price":
			if price, ok := value.(string); ok {
				if parsedPrice, err := strconv.ParseFloat(price, 64); err == nil {
					processedUpdates[key] = parsedPrice
				} else {
					fmt.Printf("Неверный формат цены: %s\n", price)
					c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный формат цены"})
					return
				}
			} else {
				processedUpdates[key] = value
			}
		case "status":
			// Убедиться, что статус является булевым
			if status, ok := value.(bool); ok {
				processedUpdates[key] = status
			} else {
				processedUpdates[key] = value
			}
		case "photo":
			// Обработка поля photo - если пустая строка, удалить файл и установить NULL
			if photoStr, ok := value.(string); ok {
				if photoStr == "" {
					// Удалить файл, если он существует
					if part.Photo != "" {
						photoPath := strings.TrimPrefix(part.Photo, "/")
						if err := os.Remove(photoPath); err != nil && !os.IsNotExist(err) {
							fmt.Printf("Warning: Failed to remove photo file %s: %v\n", photoPath, err)
						} else if err == nil {
							fmt.Printf("Successfully removed photo file: %s\n", photoPath)
						}
					}
					processedUpdates[key] = nil // Установить NULL
				} else {
					processedUpdates[key] = photoStr
				}
			} else {
				processedUpdates[key] = value
			}
		default:
			processedUpdates[key] = value
		}
	}
	fmt.Printf("Обработанные обновления: %+v\n", processedUpdates)

	// Попытаться обновить с индивидуальными обновлениями полей для перехвата конкретных ошибок
	fmt.Printf("Starting individual field updates...\n")
	for key, value := range processedUpdates {
		fmt.Printf("Updating field %s with value %v (type: %T)\n", key, value, value)
		if err := db.Model(&Part{}).Where("id = ?", uint(partID)).Update(key, value).Error; err != nil {
			fmt.Printf("Ошибка обновления поля %s со значением %v: %v\n", key, value, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Не удалось обновить поле %s", key)})
			return
		}
		fmt.Printf("Successfully updated field %s\n", key)
	}
	fmt.Printf("All field updates completed\n")

	// Переиндексация обновлённой запчасти в Elasticsearch
	var updatedPart Part
	if err := db.First(&updatedPart, uint(partID)).Error; err != nil {
		fmt.Printf("Warning: Failed to fetch updated part for indexing: %v\n", err)
	} else {
		// Заполнить форматированные поля дат для обновленной части
		now := time.Now()
		if updatedPart.ToDeleteAt != nil {
			updatedPart.ToDeleteAtFormatted = updatedPart.ToDeleteAt.Format("2006-01-02 15:04:05")
			updatedPart.TimeUntilDeletion = formatTimeUntil(*updatedPart.ToDeleteAt, now)
		}

		if err := IndexPart(&updatedPart); err != nil {
			fmt.Printf("Warning: Failed to re-index part in Elasticsearch: %v\n", err)
			// Don't fail the request, just log the warning
		} else {
			fmt.Printf("Successfully re-indexed part %d in Elasticsearch\n", updatedPart.ID)
		}
	}

	fmt.Printf("=== UPDATE PART COMPLETED SUCCESSFULLY ===\n")
	c.JSON(http.StatusOK, gin.H{"message": "Часть обновлена успешно"})
}

func getStatistics(c *gin.Context) {
	fmt.Printf("=== GET STATISTICS START ===\n")

	var totalParts int64
	db.Model(&Part{}).Where("to_delete_at IS NULL AND quantity >= 1").Count(&totalParts)
	fmt.Printf("Total parts in inventory: %d\n", totalParts)

	var totalValue float64
	db.Model(&Part{}).Where("to_delete_at IS NULL AND quantity >= 1").Select("COALESCE(SUM(price * quantity), 0)").Scan(&totalValue)
	fmt.Printf("Total inventory value: %.2f\n", totalValue)

	// Calculate total sales from auto-deleted orders and completed orders (green status)
	var totalSales float64
	result := db.Table("orders").
		Select("COALESCE(SUM(oi.quantity * oi.price), 0)").
		Joins("JOIN order_items oi ON orders.id = oi.order_id").
		Where("orders.auto_deleted = ? OR orders.status = ?", true, "green").
		Scan(&totalSales)

	if result.Error != nil {
		fmt.Printf("ERROR calculating total sales: %v\n", result.Error)
	} else {
		fmt.Printf("Total sales calculated: %.2f\n", totalSales)
	}

	// Debug: Check how many orders match the criteria
	var orderCount int64
	db.Table("orders").Where("auto_deleted = ? OR status = ?", true, "green").Count(&orderCount)
	fmt.Printf("Orders matching sales criteria (auto_deleted=true OR status=green): %d\n", orderCount)

	// Debug: Check individual counts
	var autoDeletedCount int64
	db.Table("orders").Where("auto_deleted = ?", true).Count(&autoDeletedCount)
	fmt.Printf("Orders with auto_deleted=true: %d\n", autoDeletedCount)

	var greenStatusCount int64
	db.Table("orders").Where("status = ?", "green").Count(&greenStatusCount)
	fmt.Printf("Orders with status=green: %d\n", greenStatusCount)

	// Debug: Check if there are any orders at all
	var allOrdersCount int64
	db.Table("orders").Count(&allOrdersCount)
	fmt.Printf("Total orders in database: %d\n", allOrdersCount)

	// Debug: Check order_items count
	var orderItemsCount int64
	db.Table("order_items").Count(&orderItemsCount)
	fmt.Printf("Total order_items in database: %d\n", orderItemsCount)

	var categories []CategoryCount
	db.Model(&Part{}).Select("category as name, COUNT(*) as count").Where("category != '' AND to_delete_at IS NULL AND quantity >= 1").Group("category").Scan(&categories)
	fmt.Printf("Inventory categories: %d\n", len(categories))

	var monthlySales []MonthlySales
	// Calculate monthly sales for the current year from auto-deleted and completed orders
	currentYear := time.Now().Year()
	for month := 1; month <= 12; month++ {
		startOfMonth := time.Date(currentYear, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		endOfMonth := startOfMonth.AddDate(0, 1, 0).Add(-time.Nanosecond)

		var monthlyTotal float64
		result := db.Table("orders").
			Select("COALESCE(SUM(oi.quantity * oi.price), 0)").
			Joins("JOIN order_items oi ON orders.id = oi.order_id").
			Where("(orders.auto_deleted = ? OR orders.status = ?) AND orders.created_at >= ? AND orders.created_at <= ?", true, "green", startOfMonth, endOfMonth).
			Scan(&monthlyTotal)

		if result.Error != nil {
			fmt.Printf("ERROR calculating monthly sales for month %d: %v\n", month, result.Error)
			monthlyTotal = 0
		}

		monthName := startOfMonth.Format("2006-01")
		monthlySales = append(monthlySales, MonthlySales{
			Month: monthName,
			Sales: monthlyTotal,
		})
		fmt.Printf("Month %s: sales %.2f\n", monthName, monthlyTotal)
	}
	fmt.Printf("Monthly sales calculated: %d months\n", len(monthlySales))

	response := StatisticsResponse{
		TotalParts:    int(totalParts),
		TotalValue:    totalValue,
		TotalEarnings: totalSales,
		Categories:    categories,
		MonthlySales:  monthlySales,
	}

	fmt.Printf("Final statistics response: %+v\n", response)
	fmt.Printf("=== GET STATISTICS END ===\n")

	c.JSON(http.StatusOK, response)
}

func markPartForDeletion(c *gin.Context) {
	partID := c.Param("id")

	// Convert partID to uint for proper GORM query
	id, err := strconv.ParseUint(partID, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	// Проверить, существует ли часть
	var part Part
	if err := db.First(&part, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Часть не найдена"})
		return
	}

	// Mark part for deletion in 14 days
	fourteenDaysFromNow := time.Now().AddDate(0, 0, 14)
	if err := db.Model(&Part{}).Where("id = ?", uint(id)).Update("to_delete_at", fourteenDaysFromNow).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось отметить часть для удаления"})
		return
	}

	// Получить обновленную часть для возврата форматированных полей
	var updatedPart Part
	if err := db.First(&updatedPart, uint(id)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить обновленную часть"})
		return
	}

	// Заполнить форматированные поля дат
	now := time.Now()
	if updatedPart.ToDeleteAt != nil {
		updatedPart.ToDeleteAtFormatted = updatedPart.ToDeleteAt.Format("2006-01-02 15:04:05")
		updatedPart.TimeUntilDeletion = formatTimeUntil(*updatedPart.ToDeleteAt, now)
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Часть отмечена для удаления через 14 дней",
		"part":    updatedPart,
	})
}

func deletePartPhoto(c *gin.Context) {
	partID := c.Param("id")
	fmt.Printf("=== DELETE PART PHOTO START ===\n")
	fmt.Printf("deletePartPhoto called for part ID: %s\n", partID)

	// Convert partID to uint for proper GORM query
	id, err := strconv.ParseUint(partID, 10, 32)
	if err != nil {
		fmt.Printf("Invalid part ID: %s, error: %v\n", partID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	// Проверить, существует ли часть
	var part Part
	if err := db.First(&part, uint(id)).Error; err != nil {
		fmt.Printf("Part not found: %d, error: %v\n", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Часть не найдена"})
		return
	}
	fmt.Printf("Found part: %+v\n", part)

	// Удалить файл фото, если он существует
	if part.Photo != "" {
		// Извлечь имя файла из пути фото (удалить префикс "/uploads/")
		photoPath := strings.TrimPrefix(part.Photo, "/")
		fmt.Printf("Attempting to remove photo: %s\n", photoPath)
		if err := os.Remove(photoPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Warning: Failed to remove photo file %s: %v\n", photoPath, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить файл фото"})
			return
		} else if err == nil {
			fmt.Printf("Successfully removed photo file: %s\n", photoPath)
		}
	}

	// Обновить часть, установив photo в NULL
	if err := db.Model(&Part{}).Where("id = ?", uint(id)).Update("photo", nil).Error; err != nil {
		fmt.Printf("Failed to update part photo to NULL: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить часть"})
		return
	}
	fmt.Printf("Part photo updated to NULL successfully\n")

	// Re-index the updated part in Elasticsearch
	var updatedPart Part
	if err := db.First(&updatedPart, uint(id)).Error; err != nil {
		fmt.Printf("Warning: Failed to fetch updated part for indexing: %v\n", err)
	} else {
		if err := IndexPart(&updatedPart); err != nil {
			fmt.Printf("Warning: Failed to re-index part in Elasticsearch: %v\n", err)
		} else {
			fmt.Printf("Successfully re-indexed part %d in Elasticsearch after photo deletion\n", updatedPart.ID)
		}
	}

	fmt.Printf("Photo deletion completed successfully for part %d\n", id)
	c.JSON(http.StatusOK, gin.H{"message": "Фото удалено успешно"})
}

// formatTimeUntil форматирует время до будущей даты (например, "через 3 дня")
func formatTimeUntil(targetTime time.Time, now time.Time) string {
	duration := targetTime.Sub(now)

	if duration <= 0 {
		return "уже прошло"
	}

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		if days == 1 {
			return "через 1 день"
		} else if days < 5 {
			return fmt.Sprintf("через %d дня", days)
		} else {
			return fmt.Sprintf("через %d дней", days)
		}
	} else if hours > 0 {
		if hours == 1 {
			return "через 1 час"
		} else if hours < 5 {
			return fmt.Sprintf("через %d часа", hours)
		} else {
			return fmt.Sprintf("через %d часов", hours)
		}
	} else if minutes > 0 {
		if minutes == 1 {
			return "через 1 минуту"
		} else if minutes < 5 {
			return fmt.Sprintf("через %d минуты", minutes)
		} else {
			return fmt.Sprintf("через %d минут", minutes)
		}
	} else {
		return "скоро"
	}
}

func uploadPartPhoto(c *gin.Context) {
	partID := c.Param("id")
	fmt.Printf("=== UPLOAD PART PHOTO START ===\n")
	fmt.Printf("uploadPartPhoto called for part ID: %s\n", partID)
	fmt.Printf("Request method: %s\n", c.Request.Method)
	fmt.Printf("Content-Type: %s\n", c.GetHeader("Content-Type"))
	fmt.Printf("Request headers: %+v\n", c.Request.Header)

	// Convert partID to uint for proper GORM query
	id, err := strconv.ParseUint(partID, 10, 32)
	if err != nil {
		fmt.Printf("Invalid part ID: %s, error: %v\n", partID, err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	// Проверить, существует ли часть
	var part Part
	if err := db.First(&part, uint(id)).Error; err != nil {
		fmt.Printf("Part not found: %d, error: %v\n", id, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Часть не найдена"})
		return
	}
	fmt.Printf("Found part: %+v\n", part)

	// Удалить старый файл фото, если он существует
	if part.Photo != "" {
		// Извлечь имя файла из пути фото (удалить префикс "/uploads/")
		oldPhotoPath := strings.TrimPrefix(part.Photo, "/")
		fmt.Printf("Attempting to remove old photo: %s\n", oldPhotoPath)
		if err := os.Remove(oldPhotoPath); err != nil && !os.IsNotExist(err) {
			// Записать ошибку, но не прерывать загрузку - очистка старого файла не критична
			fmt.Printf("Предупреждение: Не удалось удалить старый файл фото %s: %v\n", oldPhotoPath, err)
		} else if err == nil {
			fmt.Printf("Успешно удален старый файл фото: %s\n", oldPhotoPath)
		}
	}

	// Получить загруженный файл
	file, err := c.FormFile("photo")
	if err != nil {
		fmt.Printf("No photo file provided, error: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл фото не предоставлен"})
		return
	}
	fmt.Printf("Received file: %s, size: %d, content-type: %s\n", file.Filename, file.Size, file.Header.Get("Content-Type"))

	// Валидировать тип файла
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		fmt.Printf("Invalid file type: %s\n", file.Header.Get("Content-Type"))
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл должен быть изображением"})
		return
	}

	// Валидировать размер файла (макс 5МБ)
	if file.Size > 5*1024*1024 {
		fmt.Printf("File too large: %d bytes\n", file.Size)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Размер файла должен быть менее 5МБ"})
		return
	}

	// Создать директорию uploads, если она не существует
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		fmt.Printf("Failed to create uploads directory: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать директорию uploads"})
		return
	}

	// Сгенерировать уникальное имя файла
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg" // расширение по умолчанию
	}
	filename := fmt.Sprintf("%s_%d%s", partID, time.Now().Unix(), ext)
	filePath := filepath.Join("./uploads", filename)
	fmt.Printf("Generated filename: %s, path: %s\n", filename, filePath)

	// Сохранить файл с лучшей обработкой ошибок
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		fmt.Printf("Failed to save file: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":    "Не удалось сохранить фото",
			"details":  err.Error(),
			"filename": filename,
		})
		return
	}
	fmt.Printf("File saved successfully to: %s\n", filePath)

	// Обновить часть с путем к фото
	photoPath := "/uploads/" + filename
	fmt.Printf("Updating part %d with photo path: %s\n", id, photoPath)

	// Сначала проверить текущее значение photo
	var currentPart Part
	if err := db.First(&currentPart, uint(id)).Error; err != nil {
		fmt.Printf("Failed to fetch current part before update: %v\n", err)
	} else {
		fmt.Printf("Current part photo before update: '%s'\n", currentPart.Photo)
	}

	if err := db.Model(&Part{}).Where("id = ?", uint(id)).Update("photo", photoPath).Error; err != nil {
		fmt.Printf("Failed to update part in database: %v\n", err)
		// Попытаться очистить загруженный файл, если обновление базы данных не удалось
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Warning: Failed to remove uploaded file after database error: %v\n", err)
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Не удалось обновить часть с фото",
			"details": err.Error(),
		})
		return
	}
	fmt.Printf("Part updated successfully with photo path: %s\n", photoPath)

	// Проверить, что обновление действительно произошло
	if err := db.First(&currentPart, uint(id)).Error; err != nil {
		fmt.Printf("Failed to fetch updated part: %v\n", err)
	} else {
		fmt.Printf("Updated part photo: '%s'\n", currentPart.Photo)
	}

	// Re-index the updated part in Elasticsearch
	var updatedPart Part
	if err := db.First(&updatedPart, uint(id)).Error; err != nil {
		fmt.Printf("Warning: Failed to fetch updated part for indexing: %v\n", err)
	} else {
		if err := IndexPart(&updatedPart); err != nil {
			fmt.Printf("Warning: Failed to re-index part in Elasticsearch: %v\n", err)
			// Don't fail the request, just log the warning
		} else {
			fmt.Printf("Successfully re-indexed part %d in Elasticsearch after photo upload\n", updatedPart.ID)
		}
	}

	// Вернуть более подробный ответ
	fmt.Printf("Photo upload completed successfully for part %d\n", id)
	c.JSON(http.StatusOK, gin.H{
		"message":  "Фото загружено успешно",
		"photo":    photoPath,
		"filename": filename,
		"size":     file.Size,
		"type":     file.Header.Get("Content-Type"),
	})
}

// exportXMLPriceList экспортирует прайс-лист в формате XML для Drom
func exportXMLPriceList(c *gin.Context) {
	// Получить все доступные части для экспорта
	parts, err := GetPartsForXML()
	if err != nil {
		fmt.Printf("Ошибка получения частей для XML экспорта: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить данные для экспорта"})
		return
	}

	// Генерировать XML
	xmlData, err := GenerateXMLPriceList(parts)
	if err != nil {
		fmt.Printf("Ошибка генерации XML: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать XML"})
		return
	}

	// Установить заголовки для скачивания файла
	filename := fmt.Sprintf("price-list-%s.xml", time.Now().Format("2006-01-02"))
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Header("Content-Type", "application/xml; charset=utf-8")
	c.Header("Content-Length", fmt.Sprintf("%d", len(xmlData)))

	// Отправить XML данные
	c.Data(http.StatusOK, "application/xml; charset=utf-8", xmlData)

	fmt.Printf("Успешно экспортирован XML прайс-лист с %d предложениями\n", len(parts))
}

// Характеристики теперь хранятся в основной таблице Part, поэтому эти обработчики больше не нужны

// createDefectReport создает дефектную ведомость и массово добавляет выбранные запчасти
func createDefectReport(c *gin.Context) {
	var defectReportData struct {
		Brand         string `json:"brand"`
		Model         string `json:"model"`
		Year          int    `json:"year"`
		VIN           string `json:"vin"`
		Mileage       int    `json:"mileage"`
		Description   string `json:"description"`
		SelectedParts []struct {
			Name        string  `json:"name"`
			Category    string  `json:"category"`
			Description string  `json:"description"`
			Quantity    int     `json:"quantity"`
			Price       float64 `json:"price"`
			// Характеристики запчасти
			BodyBrand          string `json:"body_brand,omitempty"`
			EngineBrand        string `json:"engine_brand,omitempty"`
			CarReleaseDate     string `json:"car_release_date,omitempty"`
			FrontRear          string `json:"front_rear,omitempty"`
			LeftRight          string `json:"left_right,omitempty"`
			TopBottom          string `json:"top_bottom,omitempty"`
			Number             string `json:"number,omitempty"`
			Manufacturer       string `json:"manufacturer,omitempty"`
			ManufacturerCode   string `json:"manufacturer_code,omitempty"`
			OEMCode            string `json:"oem_code,omitempty"`
			Color              string `json:"color,omitempty"`
			Condition          string `json:"condition,omitempty"`
			SupplierCode       string `json:"supplier_code,omitempty"`
			Defect             string `json:"defect,omitempty"`
			Transmission       string `json:"transmission,omitempty"`
			Drive              string `json:"drive,omitempty"`
			WearPercentage     string `json:"wear_percentage,omitempty"`
			Season             string `json:"season,omitempty"`
			Diameter           string `json:"diameter,omitempty"`
			Width              string `json:"width,omitempty"`
			Profile            string `json:"profile,omitempty"`
			TireQuantity       string `json:"tire_quantity,omitempty"`
			Drilling           string `json:"drilling,omitempty"`
			Offset             string `json:"offset,omitempty"`
			CenterHoleDiameter string `json:"center_hole_diameter,omitempty"`
			TireModel          string `json:"tire_model,omitempty"`
			VIN                string `json:"vin,omitempty"`
		} `json:"selectedParts"`
	}

	if err := c.ShouldBindJSON(&defectReportData); err != nil {
		fmt.Printf("Invalid JSON in createDefectReport: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Printf("Creating defect report for %s %s %d with %d parts\n", defectReportData.Brand, defectReportData.Model, defectReportData.Year, len(defectReportData.SelectedParts))

	// Получить первого пользователя как продавца по умолчанию
	// Поскольку User модель из другого сервиса, используем простую структуру
	type User struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	}

	var defaultUser User
	if err := db.Table("users").First(&defaultUser).Error; err != nil {
		fmt.Printf("Error getting default user: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить пользователя"})
		return
	}

	var createdParts []Part
	var createdPartIDs []uint

	// Создать запчасти массово
	for _, selectedPart := range defectReportData.SelectedParts {
		part := Part{
			Name:        selectedPart.Name,
			Brand:       defectReportData.Brand,
			Model:       defectReportData.Model,
			Category:    selectedPart.Category,
			Description: selectedPart.Description,
			Quantity:    selectedPart.Quantity,
			Price:       selectedPart.Price,
			Salesman:    defaultUser.Name,
			Status:      true,             // Доступна
			VIN:         selectedPart.VIN, // VIN автомобиля
			// Характеристики запчасти
			BodyBrand:          selectedPart.BodyBrand,
			EngineBrand:        selectedPart.EngineBrand,
			CarReleaseDate:     selectedPart.CarReleaseDate,
			FrontRear:          selectedPart.FrontRear,
			LeftRight:          selectedPart.LeftRight,
			TopBottom:          selectedPart.TopBottom,
			Number:             selectedPart.Number,
			Manufacturer:       selectedPart.Manufacturer,
			ManufacturerCode:   selectedPart.ManufacturerCode,
			OEMCode:            selectedPart.OEMCode,
			Color:              selectedPart.Color,
			Condition:          selectedPart.Condition,
			SupplierCode:       selectedPart.SupplierCode,
			Defect:             selectedPart.Defect,
			Transmission:       selectedPart.Transmission,
			Drive:              selectedPart.Drive,
			WearPercentage:     selectedPart.WearPercentage,
			Season:             selectedPart.Season,
			Diameter:           selectedPart.Diameter,
			Width:              selectedPart.Width,
			Profile:            selectedPart.Profile,
			TireQuantity:       selectedPart.TireQuantity,
			Drilling:           selectedPart.Drilling,
			Offset:             selectedPart.Offset,
			CenterHoleDiameter: selectedPart.CenterHoleDiameter,
			TireModel:          selectedPart.TireModel,
		}

		// Санитизация
		part.Name = strings.TrimSpace(part.Name)
		part.Description = strings.TrimSpace(part.Description)
		part.Category = strings.TrimSpace(part.Category)
		part.Salesman = strings.TrimSpace(part.Salesman)
		part.Location = strings.TrimSpace(part.Location)
		part.Brand = strings.TrimSpace(part.Brand)
		part.Model = strings.TrimSpace(part.Model)

		if err := db.Create(&part).Error; err != nil {
			fmt.Printf("Error creating part %s: %v\n", selectedPart.Name, err)
			continue // Продолжить с другими запчастями
		}

		// Получить созданную запчасть для индексации
		var createdPart Part
		if err := db.Last(&createdPart).Error; err != nil {
			fmt.Printf("Error getting created part: %v\n", err)
			continue
		}

		createdParts = append(createdParts, createdPart)
		createdPartIDs = append(createdPartIDs, createdPart.ID)

		// Индексировать запчасть в Elasticsearch
		if err := IndexPart(&createdPart); err != nil {
			fmt.Printf("Warning: Failed to index part %d in Elasticsearch: %v\n", createdPart.ID, err)
		} else {
			fmt.Printf("Successfully indexed part %d in Elasticsearch\n", createdPart.ID)
		}
	}

	fmt.Printf("Successfully created %d parts for defect report\n", len(createdParts))

	c.JSON(http.StatusCreated, gin.H{
		"message":      "Дефектная ведомость создана успешно",
		"createdParts": len(createdParts),
		"partIDs":      createdPartIDs,
	})
}

// deleteZeroQuantityPartsBySupplier удаляет запчасти с quantity=0 для указанного supplier_code
func deleteZeroQuantityPartsBySupplier(c *gin.Context) {
	supplierCode := c.Param("supplier_code")

	if supplierCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Код поставки обязателен"})
		return
	}

	fmt.Printf("Удаление запчастей с quantity=0 для supplier_code: %s\n", supplierCode)

	// Найти все запчасти с quantity=0 для данного supplier_code
	var partsToDelete []Part
	if err := db.Where("supplier_code = ? AND quantity = 0 AND to_delete_at IS NULL", supplierCode).Find(&partsToDelete).Error; err != nil {
		fmt.Printf("Ошибка поиска запчастей для удаления: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось найти запчасти для удаления"})
		return
	}

	if len(partsToDelete) == 0 {
		fmt.Printf("Не найдено запчастей с quantity=0 для supplier_code: %s\n", supplierCode)
		c.JSON(http.StatusOK, gin.H{
			"message":       "Запчасти не найдены",
			"deleted_count": 0,
		})
		return
	}

	fmt.Printf("Найдено %d запчастей для удаления\n", len(partsToDelete))

	deletedCount := 0
	var deletedIDs []uint

	// Удалить каждую запчасть
	for _, part := range partsToDelete {
		// Удалить связанный файл фото, если он существует
		if part.Photo != "" {
			photoPath := strings.TrimPrefix(part.Photo, "/")
			if err := os.Remove(photoPath); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Предупреждение: Не удалось удалить файл фото %s: %v\n", photoPath, err)
			} else if err == nil {
				fmt.Printf("Успешно удален файл фото: %s\n", photoPath)
			}
		}

		// Удалить запчасть из базы данных
		if err := db.Delete(&Part{}, part.ID).Error; err != nil {
			fmt.Printf("Ошибка удаления запчасти %d: %v\n", part.ID, err)
			continue
		}

		// Удалить запчасть из индекса Elasticsearch
		if err := DeletePartFromIndex(part.ID); err != nil {
			fmt.Printf("Warning: Failed to remove part from Elasticsearch index: %v\n", err)
		} else {
			fmt.Printf("Successfully removed part %d from Elasticsearch index\n", part.ID)
		}

		deletedIDs = append(deletedIDs, part.ID)
		deletedCount++
	}

	fmt.Printf("Успешно удалено %d запчастей для supplier_code: %s\n", deletedCount, supplierCode)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Запчасти удалены успешно",
		"deleted_count": deletedCount,
		"supplier_code": supplierCode,
		"deleted_ids":   deletedIDs,
	})
}

// getSupplierCodes возвращает список уникальных кодов поставки
func getSupplierCodes(c *gin.Context) {
	var supplierCodes []string
	if err := db.Model(&Part{}).Where("supplier_code IS NOT NULL AND supplier_code != ''").Distinct("supplier_code").Pluck("supplier_code", &supplierCodes).Error; err != nil {
		fmt.Printf("Ошибка получения кодов поставки: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить коды поставки"})
		return
	}

	fmt.Printf("Найдено %d уникальных кодов поставки\n", len(supplierCodes))

	c.JSON(http.StatusOK, gin.H{
		"supplier_codes": supplierCodes,
	})
}

// bulkDeleteParts массово удаляет запчасти по их ID
func bulkDeleteParts(c *gin.Context) {
	var requestData struct {
		PartIds []uint `json:"partIds"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		fmt.Printf("Invalid JSON in bulkDeleteParts: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(requestData.PartIds) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указаны ID запчастей для удаления"})
		return
	}

	fmt.Printf("Массовое удаление %d запчастей: %v\n", len(requestData.PartIds), requestData.PartIds)

	deletedCount := 0
	var deletedIDs []uint

	// Удалить каждую запчасть
	for _, partID := range requestData.PartIds {
		var part Part
		if err := db.First(&part, partID).Error; err != nil {
			fmt.Printf("Запчасть с ID %d не найдена: %v\n", partID, err)
			continue
		}

		// Удалить связанный файл фото, если он существует
		if part.Photo != "" {
			photoPath := strings.TrimPrefix(part.Photo, "/")
			if err := os.Remove(photoPath); err != nil && !os.IsNotExist(err) {
				fmt.Printf("Предупреждение: Не удалось удалить файл фото %s: %v\n", photoPath, err)
			} else if err == nil {
				fmt.Printf("Успешно удален файл фото: %s\n", photoPath)
			}
		}

		// Удалить запчасть из базы данных
		if err := db.Delete(&Part{}, partID).Error; err != nil {
			fmt.Printf("Ошибка удаления запчасти %d: %v\n", partID, err)
			continue
		}

		// Удалить запчасть из индекса Elasticsearch
		if err := DeletePartFromIndex(partID); err != nil {
			fmt.Printf("Warning: Failed to remove part from Elasticsearch index: %v\n", err)
		} else {
			fmt.Printf("Successfully removed part %d from Elasticsearch index\n", partID)
		}

		deletedIDs = append(deletedIDs, partID)
		deletedCount++
	}

	fmt.Printf("Успешно удалено %d запчастей\n", deletedCount)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Запчасти удалены успешно",
		"deleted_count": deletedCount,
		"deleted_ids":   deletedIDs,
	})
}

// bulkUpdateParts массово обновляет запчасти
func bulkUpdateParts(c *gin.Context) {
	var requestData struct {
		PartIds []uint                 `json:"partIds"`
		Updates map[string]interface{} `json:"updates"`
	}

	if err := c.ShouldBindJSON(&requestData); err != nil {
		fmt.Printf("Invalid JSON in bulkUpdateParts: %v\n", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(requestData.PartIds) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указаны ID запчастей для обновления"})
		return
	}

	if len(requestData.Updates) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Не указаны поля для обновления"})
		return
	}

	fmt.Printf("Массовое обновление %d запчастей: %v с обновлениями: %v\n", len(requestData.PartIds), requestData.PartIds, requestData.Updates)

	updatedCount := 0
	var updatedIDs []uint

	// Обновить каждую запчасть
	for _, partID := range requestData.PartIds {
		updates := make(map[string]interface{})
		for k, v := range requestData.Updates {
			updates[k] = v
		}

		// Обновить запчасть в базе данных
		if err := db.Model(&Part{}).Where("id = ?", partID).Updates(updates).Error; err != nil {
			fmt.Printf("Ошибка обновления запчасти %d: %v\n", partID, err)
			continue
		}

		// Получить обновленную запчасть для переиндексации
		var updatedPart Part
		if err := db.First(&updatedPart, partID).Error; err != nil {
			fmt.Printf("Ошибка получения обновленной запчасти %d: %v\n", partID, err)
			continue
		}

		// Переиндексировать запчасть в Elasticsearch
		if err := IndexPart(&updatedPart); err != nil {
			fmt.Printf("Warning: Failed to reindex part %d in Elasticsearch: %v\n", partID, err)
		} else {
			fmt.Printf("Successfully reindexed part %d in Elasticsearch\n", partID)
		}

		updatedIDs = append(updatedIDs, partID)
		updatedCount++
	}

	fmt.Printf("Успешно обновлено %d запчастей\n", updatedCount)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Запчасти обновлены успешно",
		"updated_count": updatedCount,
		"updated_ids":   updatedIDs,
	})
}

// logUserActivity логирует активность пользователя
func logUserActivity(c *gin.Context, action, resourceType, details string, resourceID *uint) {
	userIDStr := c.GetHeader("X-User-ID")
	userEmail := c.GetHeader("X-User-Email")
	userName := c.GetHeader("X-User-Name")

	if userIDStr == "" {
		fmt.Printf("Warning: Cannot log activity - no user ID in headers\n")
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		fmt.Printf("Warning: Invalid user ID in headers: %s\n", userIDStr)
		return
	}

	logData := map[string]interface{}{
		"action":        action,
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"details":       details,
	}

	jsonData, err := json.Marshal(logData)
	if err != nil {
		fmt.Printf("Warning: Failed to marshal log data: %v\n", err)
		return
	}

	req, err := http.NewRequest("POST", "http://localhost:8083/admin/user-activity-logs", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Warning: Failed to create log request: %v\n", err)
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-ID", userIDStr)
	req.Header.Set("X-User-Email", userEmail)
	req.Header.Set("X-User-Name", userName)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Warning: Failed to send log request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Warning: Log request failed with status %d\n", resp.StatusCode)
	}
}
