package main

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	statisticsCacheKey             = "parts:statistics"
	statisticsCacheTTL             = 5 * time.Minute
	inventoryCacheKey              = "parts:inventory"
	inventoryCacheKeyWithoutPhotos = "parts:inventory:without_photos"
	inventoryCacheTTL              = 10 * time.Minute
)

// ElasticsearchClient определяет интерфейс для работы с Elasticsearch
type ElasticsearchClient interface {
	IndexPart(part *Part) error
	DeletePartFromIndex(partID int64) error
	SearchParts(query map[string]interface{}, from, size int) ([]ElasticsearchPart, int64, error)
}

// InventoryService определяет интерфейс для бизнес-логики управления инвентарем запчастей
// Содержит всю логику валидации, обработки и координации между репозиторием и внешними сервисами
type InventoryService interface {
	// GetInventory Основные операции с запчастями
	GetInventory(ctx context.Context, params InventoryQueryParams) ([]Part, error)  // Получает список запчастей с фильтрами
	AddPart(ctx context.Context, part *Part) (*Part, error)                         // Добавляет новую запчасть
	UpdatePart(ctx context.Context, id int64, updates map[string]interface{}) error // Обновляет существующую запчасть
	DeletePart(ctx context.Context, id int64) error                                 // Удаляет запчасть
	MarkPartForDeletion(ctx context.Context, id int64) error                        // Отмечает запчасть для отложенного удаления
	GetStatistics(ctx context.Context) (StatisticsResponse, error)                  // Получает статистику по инвентарю

	// BulkDeleteParts Админ операции
	BulkDeleteParts(ctx context.Context, ids []int64) error                                    // Массовое удаление запчастей
	BulkUpdateParts(ctx context.Context, updates []map[string]interface{}) (int, error)        // Массовое обновление запчастей
	DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) // Удаление по поставщику
	GetSupplierCodes(ctx context.Context) ([]string, error)                                    // Получение кодов поставщиков

	// UploadPartPhoto Фото операции
	UploadPartPhoto(ctx context.Context, id int64, c *gin.Context) (string, error) // Загрузка фото запчасти
	DeletePartPhoto(ctx context.Context, id int64, photoPath string) error         // Удаление фото запчасти (если photoPath пустой - удаляет все)

	// GetPartByID Получение запчасти по ID
	GetPartByID(ctx context.Context, id int64) (*Part, error)

	// UpdateEarnings Обновление общего заработка
	UpdateEarnings(ctx context.Context, amount float64) error
}

// InventoryQueryParams параметры запроса для инвентаря
type InventoryQueryParams struct {
	Search   string
	Category string
	Brand    string
	Model    string
	Location string
	Salesman string
	Status   string
	HasPhoto string
	Page     int
	Limit    int
}

// inventoryService реализует InventoryService
type inventoryService struct {
	repo          PartRepository
	es            ElasticsearchClient
	redis         *redis.Client
	totalEarnings float64
	queryPool     sync.Pool // Pool для повторного использования map для Elasticsearch queries
}

// NewInventoryService создает новый сервис инвентаря
func NewInventoryService(repo PartRepository, es ElasticsearchClient, config *Config) InventoryService {
	// Инициализация Redis клиента с настройками для IPv4
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.RedisURL,
		Network:      "tcp", // Явно указываем TCP для IPv4
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	service := &inventoryService{
		repo:  repo,
		es:    es,
		redis: rdb,
		queryPool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{})
			},
		},
	}

	// Инициализируем totalEarnings из базы данных
	ctx := context.Background()
	if earnings, err := repo.GetTotalEarnings(ctx); err == nil {
		service.totalEarnings = earnings
	} else {
		logrus.WithError(err).Warn("Failed to load total earnings from database, starting with 0")
		service.totalEarnings = 0
	}

	return service
}

// GetInventory получает инвентарь запчастей с учетом фильтров и пагинации
func (s *inventoryService) GetInventory(ctx context.Context, params InventoryQueryParams) ([]Part, error) {
	// Очищаем просроченные запчасти перед поиском
	if err := s.cleanupExpiredParts(ctx); err != nil {
		// Логируем ошибку, но продолжаем выполнение
		logrus.WithError(err).Warn("Failed to cleanup expired parts during inventory fetch")
	}

	// Выбираем источник данных на основе параметров поиска
	parts, err := s.fetchParts(ctx, params)
	if err != nil {
		return nil, err
	}

	// Форматируем дополнительные поля для фронтенда
	s.formatPartsForDisplay(parts)

	return parts, nil
}

// cleanupExpiredParts удаляет запчасти, отмеченные для удаления более 14 дней назад
func (s *inventoryService) cleanupExpiredParts(ctx context.Context) error {
	fourteenDaysAgo := time.Now().AddDate(0, 0, -14)
	err := s.repo.DeleteExpiredParts(ctx, fourteenDaysAgo)
	if err != nil {
		logrus.WithError(err).Warn("Failed to cleanup expired parts")
	}
	return err
}

// fetchParts выбирает оптимальный источник данных для поиска
func (s *inventoryService) fetchParts(ctx context.Context, params InventoryQueryParams) ([]Part, error) {
	if s.shouldUseElasticsearch(params) && s.es != nil {
		return s.getInventoryFromElasticsearch(ctx, params)
	}
	return s.getInventoryFromDatabase(ctx, params)
}

// formatPartsForDisplay добавляет форматированные поля для отображения
func (s *inventoryService) formatPartsForDisplay(parts []Part) {
	now := time.Now()
	for i := range parts {
		if parts[i].ToDeleteAt != nil {
			parts[i].ToDeleteAtFormatted = parts[i].ToDeleteAt.Format("2006-01-02 15:04:05")
			parts[i].TimeUntilDeletion = FormatTimeUntil(*parts[i].ToDeleteAt, now)
		}
	}
}

// shouldUseElasticsearch определяет, использовать ли Elasticsearch
func (s *inventoryService) shouldUseElasticsearch(params InventoryQueryParams) bool {
	return params.Search != "" || params.Category != "" || params.Brand != "" ||
		params.Model != "" || params.Location != "" || params.Salesman != "" ||
		params.Status != "" || params.HasPhoto != ""
}

// getInventoryFromElasticsearch получает данные из Elasticsearch
func (s *inventoryService) getInventoryFromElasticsearch(ctx context.Context, params InventoryQueryParams) ([]Part, error) {
	esQuery := s.buildElasticsearchQuery(params)
	from := (params.Page - 1) * params.Limit

	esParts, _, err := s.es.SearchParts(esQuery, from, params.Limit)
	if err != nil {
		fmt.Printf("Error searching with Elasticsearch: %v\n", err)
		// Fallback to database
		return s.getInventoryFromDatabase(ctx, params)
	}

	// Преобразуем и фильтруем
	if len(esParts) == 0 {
		return []Part{}, nil
	}

	parts := make([]Part, 0, len(esParts))
	partIDs := make([]int64, len(esParts))
	for i, esPart := range esParts {
		partIDs[i] = esPart.ID
	}

	// Получаем валидные части из БД
	filters := map[string]interface{}{
		"ids_in":               partIDs,
		"to_delete_at_is_null": true,
		"quantity_gte":         0,
	}
	validParts, err := s.repo.FindWithFilters(ctx, filters, 0, 0)
	if err != nil {
		return nil, err
	}

	validIDs := make(map[int64]bool)
	for _, part := range validParts {
		validIDs[part.ID] = true
	}

	// Фильтруем
	for _, esPart := range esParts {
		if validIDs[esPart.ID] {
			part := Part{
				PartCore: PartCore{
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
					Photos:      esPart.Photos,
				},
			}
			parts = append(parts, part)
		}
	}

	return parts, nil
}

// buildElasticsearchQuery строит запрос для Elasticsearch
func (s *inventoryService) buildElasticsearchQuery(params InventoryQueryParams) map[string]interface{} {
	query := map[string]interface{}{
		"bool": map[string]interface{}{
			"must": []map[string]interface{}{},
		},
	}

	must := query["bool"].(map[string]interface{})["must"].([]map[string]interface{})

	if params.Search != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  params.Search,
				"fields": []string{"name", "description"},
			},
		})
	}

	if params.Category != "" {
		must = append(must, map[string]interface{}{
			"match": map[string]interface{}{
				"category": params.Category,
			},
		})
	}

	if params.HasPhoto != "" && params.HasPhoto != "all" {
		if params.HasPhoto == "with" {
			must = append(must, map[string]interface{}{
				"exists": map[string]interface{}{
					"field": "photos",
				},
			})
		} else if params.HasPhoto == "without" {
			must = append(must, map[string]interface{}{
				"bool": map[string]interface{}{
					"must_not": []map[string]interface{}{
						{
							"exists": map[string]interface{}{
								"field": "photos",
							},
						},
					},
				},
			})
		}
	}

	// Аналогично для других фильтров...

	query["bool"].(map[string]interface{})["must"] = must
	return query
}

// getInventoryFromDatabase получает данные из базы данных
func (s *inventoryService) getInventoryFromDatabase(ctx context.Context, params InventoryQueryParams) ([]Part, error) {
	filters := s.buildDatabaseFilters(params)
	// Add default filters
	filters["to_delete_at_is_null"] = true
	filters["quantity_gte"] = 0

	// Calculate offset for pagination
	offset := 0
	limit := 0
	if params.Limit > 0 {
		offset = (params.Page - 1) * params.Limit
		limit = params.Limit
	}

	parts, err := s.repo.FindWithFilters(ctx, filters, offset, limit)
	if err != nil {
		return nil, err
	}

	return parts, nil
}

// buildDatabaseFilters строит фильтры для базы данных
func (s *inventoryService) buildDatabaseFilters(params InventoryQueryParams) map[string]interface{} {
	filters := make(map[string]interface{})

	if params.Search != "" {
		filters["search"] = params.Search
	}
	if params.Category != "" {
		filters["category"] = params.Category
	}
	if params.Brand != "" {
		filters["brand"] = params.Brand
	}
	if params.Model != "" {
		filters["model"] = params.Model
	}
	if params.Location != "" {
		filters["location"] = params.Location
	}
	if params.Salesman != "" {
		filters["salesman"] = params.Salesman
	}
	if params.Status != "" {
		filters["status"] = params.Status
	}
	if params.HasPhoto != "" && params.HasPhoto != "all" {
		var hasPhotoBool bool
		if params.HasPhoto == "with" {
			hasPhotoBool = true
		} else if params.HasPhoto == "without" {
			hasPhotoBool = false
		}
		filters["has_photo"] = hasPhotoBool
	}

	return filters
}

// AddPart добавляет новую запчасть
func (s *inventoryService) AddPart(ctx context.Context, part *Part) (*Part, error) {
	if err := ValidatePart(part); err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, part); err != nil {
		return nil, err
	}

	// Получаем созданную запчасть
	createdPart, err := s.repo.FindByID(ctx, part.ID)
	if err != nil {
		return nil, err
	}

	// Отправляем событие индексации в Redis Stream
	err = s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: "events:orders",
		Values: map[string]interface{}{
			"type":    "part_index_requested",
			"part_id": fmt.Sprintf("%d", createdPart.ID),
		},
	}).Err()
	if err != nil {
		logrus.WithError(err).Warn("Failed to publish part_index_requested to Redis")
	}

	// Инвалидируем кэш статистики
	s.invalidateStatisticsCache(ctx)

	return createdPart, nil
}

// UpdatePart обновляет запчасть
func (s *inventoryService) UpdatePart(ctx context.Context, id int64, updates map[string]interface{}) error {
	// Получаем существующую часть для обработки обновлений
	existingPart, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	processedUpdates, err := ProcessPartUpdates(updates, existingPart)
	if err != nil {
		return err
	}

	if err := s.repo.Update(ctx, id, processedUpdates); err != nil {
		return err
	}

	// Отправляем событие индексации в Redis Stream
	err = s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: "events:orders",
		Values: map[string]interface{}{
			"type":    "part_index_requested",
			"part_id": fmt.Sprintf("%d", id),
		},
	}).Err()
	if err != nil {
		logrus.WithError(err).Warn("Failed to publish part_index_requested (update) to Redis")
	}

	// Инвалидируем кэш статистики
	s.invalidateStatisticsCache(ctx)

	return nil
}

// DeletePart удаляет запчасть
func (s *inventoryService) DeletePart(ctx context.Context, id int64) error {
	part, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	// Удаляем все фото если есть
	for _, photoPath := range part.Photos {
		if photoPath != "" {
			if err := DeletePhotoFile(photoPath); err != nil {
				fmt.Printf("Warning: Failed to delete photo file: %v\n", err)
			}
		}
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Отправляем событие удаления из индекса в Redis Stream
	err = s.redis.XAdd(ctx, &redis.XAddArgs{
		Stream: "events:orders",
		Values: map[string]interface{}{
			"type":    "part_delete_requested",
			"part_id": fmt.Sprintf("%d", id),
		},
	}).Err()
	if err != nil {
		logrus.WithError(err).Warn("Failed to publish part_delete_requested to Redis")
	}

	// Инвалидируем кэш статистики
	s.invalidateStatisticsCache(ctx)

	return nil
}

// MarkPartForDeletion отмечает запчасть для удаления
func (s *inventoryService) MarkPartForDeletion(ctx context.Context, id int64) error {
	fourteenDaysFromNow := time.Now().AddDate(0, 0, 14)
	return s.repo.MarkForDeletion(ctx, id, fourteenDaysFromNow)
}

// GetStatistics получает статистику с кэшированием
func (s *inventoryService) GetStatistics(ctx context.Context) (StatisticsResponse, error) {
	// Проверяем кэш
	cachedData, err := s.redis.Get(ctx, statisticsCacheKey).Result()
	if err == nil {
		// Данные найдены в кэше
		var stats StatisticsResponse
		if err := json.Unmarshal([]byte(cachedData), &stats); err == nil {
			logrus.Info("Statistics retrieved from cache")
			return stats, nil
		}
		logrus.WithError(err).Warn("Failed to unmarshal cached statistics, falling back to database")
	}

	// Получаем данные из базы данных
	stats, err := s.repo.GetStatistics(ctx)
	if err != nil {
		return StatisticsResponse{}, err
	}

	// Устанавливаем общий заработок
	stats.TotalEarnings = s.totalEarnings

	// Получаем месячные продажи из orders-service
	monthlySales, err := s.getMonthlySalesFromOrdersService(ctx)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get monthly sales from orders service, using empty list")
		stats.MonthlySales = []MonthlySales{}
	} else {
		stats.MonthlySales = monthlySales
	}

	// Кэшируем результат
	if data, err := json.Marshal(stats); err == nil {
		if err := s.redis.Set(ctx, statisticsCacheKey, data, statisticsCacheTTL).Err(); err != nil {
			logrus.WithError(err).Warn("Failed to cache statistics")
		} else {
			logrus.Info("Statistics cached successfully")
		}
	} else {
		logrus.WithError(err).Warn("Failed to marshal statistics for caching")
	}

	return stats, nil
}

// getMonthlySalesFromOrdersService получает месячные продажи из orders-service через gRPC
func (s *inventoryService) getMonthlySalesFromOrdersService(ctx context.Context) ([]MonthlySales, error) {
	sales, err := getMonthlySalesGRPC(ctx)
	if err == nil {
		return sales, nil
	}

	logrus.WithError(err).Warn("Failed to get monthly sales via gRPC, returning empty list")
	return []MonthlySales{}, nil
}

// invalidateStatisticsCache инвалидирует кэш статистики
func (s *inventoryService) invalidateStatisticsCache(ctx context.Context) {
	if err := s.redis.Del(ctx, statisticsCacheKey).Err(); err != nil {
		logrus.WithError(err).Warn("Failed to invalidate statistics cache")
	} else {
		logrus.Info("Statistics cache invalidated")
	}
}

// BulkDeleteParts удаляет несколько запчастей с использованием worker pool для параллельной обработки
func (s *inventoryService) BulkDeleteParts(ctx context.Context, ids []int64) error {
	logrus.WithFields(logrus.Fields{
		"ids":   ids,
		"count": len(ids),
	}).Info("InventoryService.BulkDeleteParts: Starting bulk delete")

	// Worker pool для параллельной обработки удаления фото и ES индекса
	numWorkers := runtime.NumCPU() // Используем все доступные ядра
	jobs := make(chan int64, len(ids))
	var wg sync.WaitGroup

	// Запуск воркеров
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for id := range jobs {
				part, err := s.repo.FindByID(ctx, id)
				if err != nil {
					logrus.WithError(err).WithField("id", id).Warn("InventoryService.BulkDeleteParts: Failed to find part for photo deletion")
					continue
				}

				// Удаляем все фото если есть
				for _, photoPath := range part.Photos {
					if photoPath != "" {
						if err := DeletePhotoFile(photoPath); err != nil {
							logrus.WithError(err).WithFields(logrus.Fields{
								"id":        id,
								"photoPath": photoPath,
							}).Warn("InventoryService.BulkDeleteParts: Failed to delete photo file")
						}
					}
				}

				// Отправляем событие удаления из индекса в Redis Stream
				err = s.redis.XAdd(ctx, &redis.XAddArgs{
					Stream: "events:orders",
					Values: map[string]interface{}{
						"type":    "part_delete_requested",
						"part_id": fmt.Sprintf("%d", id),
					},
				}).Err()
				if err != nil {
					logrus.WithError(err).WithField("id", id).Warn("Failed to publish part_delete_requested to Redis")
				}
			}
		}()
	}

	// Отправка заданий
	for _, id := range ids {
		jobs <- id
	}
	close(jobs)

	// Ожидание завершения воркеров
	wg.Wait()

	err := s.repo.BulkDelete(ctx, ids)
	if err != nil {
		logrus.WithError(err).Error("InventoryService.BulkDeleteParts: Failed to bulk delete")
		return err
	}

	// Инвалидируем кэш статистики
	s.invalidateStatisticsCache(ctx)

	logrus.Info("InventoryService.BulkDeleteParts: Successfully completed bulk delete")
	return nil
}

// BulkUpdateParts обновляет несколько запчастей
func (s *inventoryService) BulkUpdateParts(ctx context.Context, updates []map[string]interface{}) (int, error) {
	logrus.WithFields(logrus.Fields{
		"updates": updates,
		"count":   len(updates),
	}).Info("InventoryService.BulkUpdateParts: Starting bulk update")

	updatedCount, err := s.repo.BulkUpdate(ctx, updates)
	if err != nil {
		logrus.WithError(err).Error("InventoryService.BulkUpdateParts: Failed to bulk update")
		return 0, err
	}

	// Инвалидируем кэш статистики
	s.invalidateStatisticsCache(ctx)

	logrus.Info("InventoryService.BulkUpdateParts: Successfully completed bulk update")
	return updatedCount, nil
}

// DeleteZeroQuantityPartsBySupplier удаляет запчасти с нулевым количеством по поставщику
func (s *inventoryService) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	fmt.Printf("Service: DeleteZeroQuantityPartsBySupplier called with supplier_code='%s'\n", supplierCode)

	deletedCount, err := s.repo.DeleteZeroQuantityPartsBySupplier(ctx, supplierCode)
	if err != nil {
		fmt.Printf("Service: DeleteZeroQuantityPartsBySupplier failed for supplier_code='%s': %v\n", supplierCode, err)
		return 0, err
	}

	fmt.Printf("Service: DeleteZeroQuantityPartsBySupplier completed successfully for supplier_code='%s', deleted %d parts\n", supplierCode, deletedCount)
	return deletedCount, nil
}

// GetSupplierCodes получает коды поставщиков
func (s *inventoryService) GetSupplierCodes(ctx context.Context) ([]string, error) {
	codes, err := s.repo.GetSupplierCodes(ctx)
	if err != nil {
		fmt.Printf("Service: GetSupplierCodes failed: %v\n", err)
		return nil, err
	}
	fmt.Printf("Service: GetSupplierCodes returned %d codes\n", len(codes))
	return codes, nil
}

// UploadPartPhoto загружает фото
func (s *inventoryService) UploadPartPhoto(ctx context.Context, id int64, c *gin.Context) (string, error) {
	return HandlePhotoUpload(c, id)
}

// DeletePartPhoto удаляет фото
func (s *inventoryService) DeletePartPhoto(ctx context.Context, id int64, photoPath string) error {
	return DeletePhoto(id, photoPath)
}

// GetPartByID получает запчасть по ID
func (s *inventoryService) GetPartByID(ctx context.Context, id int64) (*Part, error) {
	return s.repo.FindByID(ctx, id)
}

// UpdateEarnings обновляет общий заработок
func (s *inventoryService) UpdateEarnings(ctx context.Context, amount float64) error {
	s.totalEarnings += amount

	// Сохраняем в базу данных для персистентности
	if err := s.repo.UpdateTotalEarnings(ctx, s.totalEarnings); err != nil {
		logrus.WithError(err).Error("Failed to save total earnings to database")
		return err
	}

	logrus.WithField("totalEarnings", s.totalEarnings).Info("Updated total earnings in database")
	return nil
}
