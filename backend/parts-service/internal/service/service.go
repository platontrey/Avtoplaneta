package service

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/userdirectory"
	"parts-service/internal/catalog"
	"parts-service/internal/config"
	"parts-service/internal/domain"
	"parts-service/internal/repository"
	"parts-service/internal/search"
)

const (
	statisticsCacheKey             = "parts:statistics"
	statisticsCacheTTL             = 5 * time.Minute
	inventoryCacheKey              = "parts:inventory"
	inventoryCacheKeyWithoutPhotos = "parts:inventory:without_photos"
	inventoryCacheTTL              = 10 * time.Minute
)

// InventoryService определяет интерфейс для бизнес-логики управления инвентарем запчастей
// Содержит всю логику валидации, обработки и координации между репозиторием и внешними сервисами
type InventoryService interface {
	// GetInventory Основные операции с запчастями
	GetInventory(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, error)  // Получает список запчастей с фильтрами
	AddPart(ctx context.Context, part *domain.Part) (*domain.Part, error)                         // Добавляет новую запчасть
	UpdatePart(ctx context.Context, id int64, updates map[string]interface{}) error // Обновляет существующую запчасть
	DeletePart(ctx context.Context, id int64) error                                 // Удаляет запчасть
	MarkPartForDeletion(ctx context.Context, id int64) error                        // Отмечает запчасть для отложенного удаления
	GetStatistics(ctx context.Context) (domain.StatisticsResponse, error)                  // Получает статистику по инвентарю
	InventoryVersion(ctx context.Context) (string, error)                           // Отпечаток состояния склада для условных запросов
	RenameSeller(ctx context.Context, sellerID int64, name string) ([]domain.Part, error)  // Приводит копию имени продавца в строках к справочнику
	PartsForExport(ctx context.Context) ([]domain.Part, error)                             // Запчасти для выгрузки прайс-листа

	// BulkDeleteParts Админ операции
	BulkDeleteParts(ctx context.Context, ids []int64) error                                    // Массовое удаление запчастей
	BulkUpdateParts(ctx context.Context, updates []map[string]interface{}) (int, error)        // Массовое обновление запчастей
	DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) // Удаление по поставщику
	GetSupplierCodes(ctx context.Context) ([]string, error)                                    // Получение кодов поставщиков
	GetNextSupplierCode(ctx context.Context) (string, error)                                   // Получение следующего номера поставки
	PeekNextSupplierCode(ctx context.Context) (string, error)                                  // Просмотр следующего номера поставки без инкремента

	// UploadPartPhoto Фото операции
	UploadPartPhoto(ctx context.Context, id int64, c *gin.Context) (string, error) // Загрузка фото запчасти
	DeletePartPhoto(ctx context.Context, id int64, photoPath string) error         // Удаление фото запчасти (если photoPath пустой - удаляет все)
	SavePhotoFromBytes(ctx context.Context, id int64, data []byte) (string, error)

	// GetPartByID Получение запчасти по ID
	GetPartByID(ctx context.Context, id int64) (*domain.Part, error)
	AddPartsBatch(ctx context.Context, parts []domain.Part) ([]domain.Part, error)

	DecreasePartQuantity(ctx context.Context, id int64, amount int, operationID string) error
	IncreasePartQuantity(ctx context.Context, id int64, amount int, operationID string) error

	// UpdateEarnings Обновление общего заработка
	UpdateEarnings(ctx context.Context, amount float64) error
	ReindexAllParts(ctx context.Context) error
	SetMonthlySalesProvider(fn func(ctx context.Context) ([]domain.MonthlySales, error))
}

// inventoryService реализует InventoryService
type inventoryService struct {
	repo                 repository.PartRepository
	es                   search.ElasticsearchClient
	redis                *redis.Client
	users                *userdirectory.Directory
	totalEarnings        float64
	queryPool            sync.Pool // Pool для повторного использования map для Elasticsearch queries
	monthlySalesProvider func(ctx context.Context) ([]domain.MonthlySales, error)
}

// NewInventoryService создает новый сервис инвентаря
func NewInventoryService(repo repository.PartRepository, es search.ElasticsearchClient, config *config.Config, users *userdirectory.Directory) InventoryService {
	// Инициализация Redis клиента с настройками для IPv4
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.RedisURL,
		Network:      "tcp", // Явно указываем TCP для IPv4
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})

	service := &inventoryService{
		repo:          repo,
		es:            es,
		redis:         rdb,
		users:         users,
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
func (s *inventoryService) GetInventory(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, error) {
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
func (s *inventoryService) fetchParts(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, error) {
	if s.shouldUseElasticsearch(params) && s.es != nil {
		return s.getInventoryFromElasticsearch(ctx, params)
	}
	return s.getInventoryFromDatabase(ctx, params)
}

// formatPartsForDisplay добавляет форматированные поля для отображения
func (s *inventoryService) formatPartsForDisplay(parts []domain.Part) {
	now := time.Now()
	for i := range parts {
		if parts[i].ToDeleteAt != nil {
			parts[i].ToDeleteAtFormatted = parts[i].ToDeleteAt.Format("2006-01-02 15:04:05")
			parts[i].TimeUntilDeletion = domain.FormatTimeUntil(*parts[i].ToDeleteAt, now)
		}
	}
}

// shouldUseElasticsearch определяет, использовать ли Elasticsearch
func (s *inventoryService) shouldUseElasticsearch(params domain.InventoryQueryParams) bool {
	return params.Search != "" || params.Category != "" || params.Brand != "" ||
		params.Model != "" || params.Location != "" || params.Address != "" || params.Salesman != "" ||
		params.Status != "" || params.HasPhoto != "" || params.Number != "" || params.OEMCode != "" ||
		params.VIN != "" || params.BodyBrand != "" || params.EngineBrand != "" ||
		params.CarReleaseDate != "" || params.CarReleasePeriod != "" || params.Transmission != "" || params.Drive != "" ||
		params.Condition != "" || params.Manufacturer != "" || params.Defect != "" || params.Color != "" ||
		params.MinPrice != "" || params.MaxPrice != "" || params.MinQuantity != "" || params.MaxQuantity != "" ||
		params.FrontRear != "" || params.LeftRight != "" || params.TopBottom != "" ||
		params.ManufacturerCode != "" || params.SupplierCode != "" || params.TransmissionModel != "" ||
		params.WearPercentage != "" || params.Season != "" || params.Diameter != "" ||
		params.Width != "" || params.Profile != "" || params.TireQuantity != "" ||
		params.Drilling != "" || params.Offset != "" || params.CenterHoleDiameter != "" || params.TireModel != ""
}

// getInventoryFromElasticsearch получает данные из Elasticsearch
func (s *inventoryService) getInventoryFromElasticsearch(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, error) {
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
		// Fallback to database to ensure we don't return 0 if DB actually has matching parts
		dbParts, dbErr := s.getInventoryFromDatabase(ctx, params)
		if dbErr == nil && len(dbParts) > 0 {
			return dbParts, nil
		}
		return []domain.Part{}, nil
	}

	parts := make([]domain.Part, 0, len(esParts))
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

	validPartsMap := make(map[int64]domain.Part, len(validParts))
	for _, part := range validParts {
		validPartsMap[part.ID] = part
	}

	parts = make([]domain.Part, 0, len(esParts))
	for _, esPart := range esParts {
		if part, ok := validPartsMap[esPart.ID]; ok {
			parts = append(parts, part)
		}
	}

	return parts, nil
}


// buildElasticsearchQuery строит запрос для Elasticsearch (делегирует в search.BuildElasticsearchQuery)
func (s *inventoryService) buildElasticsearchQuery(params domain.InventoryQueryParams) map[string]interface{} {
	return search.BuildElasticsearchQuery(params)
}


// getInventoryFromDatabase получает данные из базы данных
func (s *inventoryService) getInventoryFromDatabase(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, error) {
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
func (s *inventoryService) buildDatabaseFilters(params domain.InventoryQueryParams) map[string]interface{} {
	filters := make(map[string]interface{})

	if params.Search != "" {
		filters["search"] = params.Search
	}
	if params.Category != "" {
		filters["category_ilike"] = catalog.ExpandCategorySynonyms(params.Category)
	}
	if params.Brand != "" {
		filters["brand_ilike"] = params.Brand
	}
	if params.Model != "" {
		filters["model_ilike"] = params.Model
	}
	if params.Location != "" {
		filters["location_ilike"] = params.Location
	}
	if params.Address != "" {
		filters["address_ilike"] = params.Address
	}
	if params.Salesman != "" {
		filters["salesman_ilike"] = params.Salesman
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
	if params.Number != "" {
		filters["number_ilike"] = params.Number
	}
	if params.OEMCode != "" {
		filters["oem_code_ilike"] = params.OEMCode
	}
	if params.VIN != "" {
		filters["vin_ilike"] = params.VIN
	}
	if params.BodyBrand != "" {
		filters["body_brand_ilike"] = params.BodyBrand
	}
	if params.EngineBrand != "" {
		filters["engine_brand_ilike"] = params.EngineBrand
	}
	if params.CarReleaseDate != "" {
		filters["car_release_date_ilike"] = params.CarReleaseDate
	}
	if params.CarReleasePeriod != "" {
		filters["car_release_period_ilike"] = params.CarReleasePeriod
	}
	if params.Transmission != "" {
		filters["transmission_ilike"] = params.Transmission
	}
	if params.Drive != "" {
		filters["drive_ilike"] = params.Drive
	}
	if params.Condition != "" {
		filters["condition_ilike"] = params.Condition
	}
	if params.Manufacturer != "" {
		filters["manufacturer_ilike"] = params.Manufacturer
	}
	if params.Defect != "" {
		filters["defect_ilike"] = params.Defect
	}
	if params.Color != "" {
		filters["color_ilike"] = params.Color
	}
	if params.MinPrice != "" {
		if val, err := strconv.ParseFloat(params.MinPrice, 64); err == nil {
			filters["min_price"] = val
		}
	}
	if params.MaxPrice != "" {
		if val, err := strconv.ParseFloat(params.MaxPrice, 64); err == nil {
			filters["max_price"] = val
		}
	}
	if params.MinQuantity != "" {
		if val, err := strconv.Atoi(params.MinQuantity); err == nil {
			filters["min_quantity"] = val
		}
	}
	if params.MaxQuantity != "" {
		if val, err := strconv.Atoi(params.MaxQuantity); err == nil {
			filters["max_quantity"] = val
		}
	}
	if params.FrontRear != "" {
		filters["front_rear_ilike"] = params.FrontRear
	}
	if params.LeftRight != "" {
		filters["left_right_ilike"] = params.LeftRight
	}
	if params.TopBottom != "" {
		filters["top_bottom_ilike"] = params.TopBottom
	}
	if params.ManufacturerCode != "" {
		filters["manufacturer_code_ilike"] = params.ManufacturerCode
	}
	if params.SupplierCode != "" {
		filters["supplier_code_ilike"] = params.SupplierCode
	}
	if params.TransmissionModel != "" {
		filters["transmission_model_ilike"] = params.TransmissionModel
	}
	if params.WearPercentage != "" {
		filters["wear_percentage_ilike"] = params.WearPercentage
	}
	if params.Season != "" {
		filters["season_ilike"] = params.Season
	}
	if params.Diameter != "" {
		filters["diameter_ilike"] = params.Diameter
	}
	if params.Width != "" {
		filters["width_ilike"] = params.Width
	}
	if params.Profile != "" {
		filters["profile_ilike"] = params.Profile
	}
	if params.TireQuantity != "" {
		filters["tire_quantity_ilike"] = params.TireQuantity
	}
	if params.Drilling != "" {
		filters["drilling_ilike"] = params.Drilling
	}
	if params.Offset != "" {
		filters["offset_ilike"] = params.Offset
	}
	if params.CenterHoleDiameter != "" {
		filters["center_hole_diameter_ilike"] = params.CenterHoleDiameter
	}
	if params.TireModel != "" {
		filters["tire_model_ilike"] = params.TireModel
	}

	return filters
}

// AddPart добавляет новую запчасть
func (s *inventoryService) AddPart(ctx context.Context, part *domain.Part) (*domain.Part, error) {
	if err := domain.ValidatePart(part); err != nil {
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

// AddPartsBatch пакетно валидирует, создает и индексирует запчасти (для дефектных ведомостей)
func (s *inventoryService) AddPartsBatch(ctx context.Context, parts []domain.Part) ([]domain.Part, error) {
	if len(parts) == 0 {
		return nil, nil
	}

	for i := range parts {
		if err := domain.ValidatePart(&parts[i]); err != nil {
			return nil, fmt.Errorf("ошибка валидации детали #%d (%s): %w", i, parts[i].Name, err)
		}
	}

	createdParts, err := s.repo.CreateBatch(ctx, parts)
	if err != nil {
		return nil, fmt.Errorf("ошибка пакетной вставки запчастей: %w", err)
	}

	// Пакетная индексация всех созданных деталей в Elasticsearch
	if err := search.BulkIndexParts(ctx, createdParts); err != nil {
		logrus.WithError(err).Warn("Не удалось пакетно проиндексировать созданные запчасти в Elasticsearch")
	}

	// Сброс кэша один раз на весь батч
	s.redis.Del(ctx, inventoryCacheKey)
	s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
	s.invalidateStatisticsCache(ctx)

	return createdParts, nil
}

// UpdatePart обновляет запчасть
func (s *inventoryService) UpdatePart(ctx context.Context, id int64, updates map[string]interface{}) error {
	// Получаем существующую часть для обработки обновлений
	existingPart, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return err
	}

	processedUpdates, err := domain.ProcessPartUpdates(updates, existingPart)
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
func (s *inventoryService) GetStatistics(ctx context.Context) (domain.StatisticsResponse, error) {
	// Проверяем кэш
	cachedData, err := s.redis.Get(ctx, statisticsCacheKey).Result()
	if err == nil {
		// Данные найдены в кэше
		var stats domain.StatisticsResponse
		if err := json.Unmarshal([]byte(cachedData), &stats); err == nil {
			logrus.Info("Statistics retrieved from cache")
			return stats, nil
		}
		logrus.WithError(err).Warn("Failed to unmarshal cached statistics, falling back to database")
	}

	// Получаем данные из базы данных
	stats, err := s.repo.GetStatistics(ctx)
	if err != nil {
		return domain.StatisticsResponse{}, err
	}

	// Устанавливаем общий заработок
	stats.TotalEarnings = s.totalEarnings

	// Получаем месячные продажи из orders-service
	monthlySales, err := s.getMonthlySalesFromOrdersService(ctx)
	if err != nil {
		logrus.WithError(err).Warn("Failed to get monthly sales from orders service, using empty list")
		stats.MonthlySales = []domain.MonthlySales{}
	} else {
		stats.MonthlySales = monthlySales
	}

	// Рассчитываем динамику продаж по месяцам
	if len(stats.MonthlySales) >= 2 {
		sorted := make([]domain.MonthlySales, len(stats.MonthlySales))
		copy(sorted, stats.MonthlySales)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Month < sorted[j].Month
		})
		last := sorted[len(sorted)-1].Sales
		prev := sorted[len(sorted)-2].Sales
		if prev > 0 {
			stats.EarningsGrowth = math.Round(((last-prev)/prev)*1000) / 10
		} else if last > 0 {
			stats.EarningsGrowth = 100.0
		}
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

// SetMonthlySalesProvider задает поставщика месячных продаж
func (s *inventoryService) SetMonthlySalesProvider(fn func(ctx context.Context) ([]domain.MonthlySales, error)) {
	s.monthlySalesProvider = fn
}

// getMonthlySalesFromOrdersService получает месячные продажи через зарегистрированного провайдера
func (s *inventoryService) getMonthlySalesFromOrdersService(ctx context.Context) ([]domain.MonthlySales, error) {
	if s.monthlySalesProvider != nil {
		sales, err := s.monthlySalesProvider(ctx)
		if err == nil {
			return sales, nil
		}
		logrus.WithError(err).Warn("Failed to get monthly sales via provider, returning empty list")
	}
	return []domain.MonthlySales{}, nil
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

	var (
		deletedCount int64
		deletedIDs   []int64
		err          error
	)

	if idDeleter, ok := s.repo.(interface {
		DeleteZeroQuantityPartsWithIDs(ctx context.Context, supplierCode string) ([]int64, error)
	}); ok {
		deletedIDs, err = idDeleter.DeleteZeroQuantityPartsWithIDs(ctx, supplierCode)
		deletedCount = int64(len(deletedIDs))
	} else {
		deletedCount, err = s.repo.DeleteZeroQuantityPartsBySupplier(ctx, supplierCode)
	}
	if err != nil {
		fmt.Printf("Service: DeleteZeroQuantityPartsBySupplier failed for supplier_code='%s': %v\n", supplierCode, err)
		return 0, err
	}

	if len(deletedIDs) > 0 {
		if esErr := search.BulkDeletePartsFromIndex(ctx, deletedIDs); esErr != nil {
			logrus.WithError(esErr).Warn("Failed to bulk delete zero-quantity parts from Elasticsearch")
		}
	}

	if s.redis != nil {
		s.redis.Del(ctx, inventoryCacheKey)
		s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
		s.invalidateStatisticsCache(ctx)
	}

	fmt.Printf("Service: DeleteZeroQuantityPartsBySupplier completed successfully for supplier_code='%s', deleted %d parts\n", supplierCode, deletedCount)
	return deletedCount, nil
}

// GetSupplierBatches получает расширенную информацию о поставках / ведомостях с нулевым количеством
func (s *inventoryService) GetSupplierBatches(ctx context.Context) ([]domain.SupplierBatchInfo, error) {
	if batchProvider, ok := s.repo.(interface {
		GetSupplierBatches(ctx context.Context) ([]domain.SupplierBatchInfo, error)
	}); ok {
		return batchProvider.GetSupplierBatches(ctx)
	}
	codes, err := s.repo.GetSupplierCodes(ctx)
	if err != nil {
		return nil, err
	}
	batches := make([]domain.SupplierBatchInfo, 0, len(codes))
	for _, code := range codes {
		batches = append(batches, domain.SupplierBatchInfo{Code: code, Label: code})
	}
	return batches, nil
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

// GetNextSupplierCode получает следующий номер поставки через репозиторий
func (s *inventoryService) GetNextSupplierCode(ctx context.Context) (string, error) {
	return s.repo.GetNextSupplierCode(ctx)
}

// PeekNextSupplierCode просматривает следующий номер поставки без инкремента
func (s *inventoryService) PeekNextSupplierCode(ctx context.Context) (string, error) {
	return s.repo.PeekNextSupplierCode(ctx)
}

// UploadPartPhoto загружает фото
func (s *inventoryService) UploadPartPhoto(ctx context.Context, id int64, c *gin.Context) (string, error) {
	return HandlePhotoUpload(ctx, c, id, s.repo)
}

// DeletePartPhoto удаляет фото
func (s *inventoryService) DeletePartPhoto(ctx context.Context, id int64, photoPath string) error {
	return DeletePhoto(ctx, id, photoPath, s.repo)
}

// SavePhotoFromBytes сохраняет фото из байтов (для gRPC streaming upload)
func (s *inventoryService) SavePhotoFromBytes(ctx context.Context, id int64, data []byte) (string, error) {
	return SavePhotoFromBytes(ctx, id, data, s.repo)
}

// GetPartByID получает запчасть по ID
func (s *inventoryService) GetPartByID(ctx context.Context, id int64) (*domain.Part, error) {
	part, err := s.repo.FindByID(ctx, id)
	if err != nil || part == nil {
		return part, err
	}

	return part, nil
}

func (s *inventoryService) DecreasePartQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	// Сброс кэша
	s.redis.Del(ctx, inventoryCacheKey)
	s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
	s.invalidateStatisticsCache(ctx)

	err := s.repo.DecreaseQuantity(ctx, id, amount, operationID)
	if err != nil {
		return err
	}

	part, err := s.GetPartByID(ctx, id)
	if err == nil && s.es != nil {
		s.es.IndexPart(part)
	}

	return nil
}

func (s *inventoryService) IncreasePartQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	// Сброс кэша
	s.redis.Del(ctx, inventoryCacheKey)
	s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
	s.invalidateStatisticsCache(ctx)

	err := s.repo.IncreaseQuantity(ctx, id, amount, operationID)
	if err != nil {
		return err
	}

	part, err := s.GetPartByID(ctx, id)
	if err == nil && s.es != nil {
		s.es.IndexPart(part)
	}

	return nil
}

// UpdateEarnings обновляет общий заработок
func (s *inventoryService) UpdateEarnings(ctx context.Context, amount float64) error {
	s.totalEarnings += amount

	// Сохраняем в базу данных для персистентности
	if err := s.repo.UpdateTotalEarnings(ctx, s.totalEarnings); err != nil {
		logrus.WithError(err).Error("Failed to save total earnings to database")
		return err
	}

	s.invalidateStatisticsCache(ctx)
	logrus.WithField("totalEarnings", s.totalEarnings).Info("Updated total earnings in database")
	return nil
}

// InventoryVersion пробрасывает отпечаток склада из репозитория. Кэшировать его
// в Redis не нужно: запрос и так дешёвый, а лишний слой добавил бы окно, в
// котором клиент получал бы 304 на уже изменившиеся данные.
func (s *inventoryService) InventoryVersion(ctx context.Context) (string, error) {
	return s.repo.InventoryVersion(ctx)
}

// RenameSeller пробрасывает починку копии имени продавца в репозиторий.
func (s *inventoryService) RenameSeller(ctx context.Context, sellerID int64, name string) ([]domain.Part, error) {
	return s.repo.RenameSeller(ctx, sellerID, name)
}

// PartsForExport отдаёт запчасти, пригодные для выгрузки прайс-листа.
// Сам прайс-лист собирает export-service; сервис запчастей только отдаёт данные.
func (s *inventoryService) PartsForExport(ctx context.Context) ([]domain.Part, error) {
	return s.repo.GetPartsForXML(ctx)
}

// ReindexAllParts переиндексирует все запчасти в Elasticsearch
func (s *inventoryService) ReindexAllParts(ctx context.Context) error {
	parts, err := s.repo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("не удалось получить запчасти: %w", err)
	}
	if s.es == nil {
		return nil
	}
	return search.BulkIndexParts(ctx, parts)
}

