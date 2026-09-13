package main

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
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

	DecreasePartQuantity(ctx context.Context, id int64, amount int) error
	IncreasePartQuantity(ctx context.Context, id int64, amount int) error

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
	Address  string
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
		params.Model != "" || params.Location != "" || params.Address != "" || params.Salesman != "" ||
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

	validPartsMap := make(map[int64]Part, len(validParts))
	for _, part := range validParts {
		validPartsMap[part.ID] = part
	}

	parts = make([]Part, 0, len(esParts))
	for _, esPart := range esParts {
		if part, ok := validPartsMap[esPart.ID]; ok {
			parts = append(parts, part)
		}
	}

	return parts, nil
}

// TransliterateLatinToCyrillic преобразует транслит латиницы в кириллицу (sirena -> сирена, bamper -> бампер)
func TransliterateLatinToCyrillic(text string) string {
	text = strings.ToLower(text)
	replacements := []struct {
		from string
		to   string
	}{
		{"shch", "щ"}, {"sh", "ш"}, {"ch", "ч"}, {"zh", "ж"},
		{"ya", "я"}, {"yu", "ю"}, {"yo", "ё"}, {"ts", "ц"},
		{"a", "а"}, {"b", "б"}, {"v", "в"}, {"g", "г"}, {"d", "д"},
		{"e", "е"}, {"z", "з"}, {"i", "и"}, {"j", "й"}, {"k", "к"},
		{"l", "л"}, {"m", "м"}, {"n", "н"}, {"o", "о"}, {"p", "п"},
		{"r", "р"}, {"s", "с"}, {"t", "т"}, {"u", "у"}, {"f", "ф"},
		{"h", "х"}, {"c", "к"}, {"y", "ы"}, {"w", "в"}, {"x", "кс"},
	}

	res := text
	for _, r := range replacements {
		res = strings.ReplaceAll(res, r.from, r.to)
	}
	return res
}

// ConvertQwertyToRussian переводит текст с неверной QWERTY раскладки на кириллицу (gthtlybq -> передний)
func ConvertQwertyToRussian(text string) string {
	qwertyMap := map[rune]rune{
		'q': 'й', 'w': 'ц', 'e': 'у', 'r': 'к', 't': 'е', 'y': 'н', 'u': 'г', 'i': 'ш', 'o': 'щ', 'p': 'з', '[': 'х', ']': 'ъ',
		'a': 'ф', 's': 'ы', 'd': 'в', 'f': 'а', 'g': 'п', 'h': 'р', 'j': 'о', 'k': 'л', 'l': 'д', ';': 'ж', '\'': 'э',
		'z': 'я', 'x': 'ч', 'c': 'с', 'v': 'м', 'b': 'и', 'n': 'т', 'm': 'ь', ',': 'б', '.': 'ю',
	}

	var builder strings.Builder
	for _, char := range strings.ToLower(text) {
		if ruChar, ok := qwertyMap[char]; ok {
			builder.WriteRune(ruChar)
		} else {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

// buildElasticsearchQuery строит запрос для Elasticsearch
func (s *inventoryService) buildElasticsearchQuery(params InventoryQueryParams) map[string]interface{} {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{}

	// Фильтр для отображения валидных запчастей (quantity >= 0)
	filter = append(filter, map[string]interface{}{
		"range": map[string]interface{}{
			"quantity": map[string]interface{}{
				"gte": 0,
			},
		},
	})

	if params.Search != "" {
		transliteratedSearch := TransliterateLatinToCyrillic(params.Search)
		qwertySearch := ConvertQwertyToRussian(params.Search)

		shouldQueries := []map[string]interface{}{
			// 1. Абсолютное точное совпадение слова/слов в названии детали (Максимальный приоритет 100.0)
			{
				"match": map[string]interface{}{
					"name": map[string]interface{}{
						"query": params.Search,
						"boost": 100.0,
					},
				},
			},
			// 2. Фразовое совпадение с префиксом в названии
			{
				"match_phrase_prefix": map[string]interface{}{
					"name": map[string]interface{}{
						"query": params.Search,
						"boost": 50.0,
					},
				},
			},
			// 3. Кросс-полейный поиск по названию, брендам, моделям и артикулам
			{
				"multi_match": map[string]interface{}{
					"query":    params.Search,
					"fields":   []string{"name^10", "name.ngram^5", "brand.text^3", "model.text^3", "category.text^2", "description^1"},
					"type":     "cross_fields",
					"operator": "or",
					"boost":    10.0,
				},
			},
		}

		// Every term in a multi-word query must match somewhere in the same
		// part. Without this guard, "АКПП ACV30" matched any АКПП (for
		// example Nissan) even when ACV30 was absent.
		requiredFields := []string{
			"name", "name.ngram", "brand.text", "model.text", "body_brand",
			"engine_brand", "number", "oem_code", "manufacturer_code",
			"supplier_code", "vin", "category.text", "description",
		}
		requiredQueries := []map[string]interface{}{
			{"multi_match": map[string]interface{}{
				"query": params.Search, "fields": requiredFields,
				"type": "cross_fields", "operator": "and",
			}},
		}

		// Если введен латинский текст, добавляем варианты транслитерации и смены раскладки в кириллицу
		if transliteratedSearch != strings.ToLower(params.Search) {
			requiredQueries = append(requiredQueries, map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": transliteratedSearch, "fields": requiredFields,
					"type": "cross_fields", "operator": "and",
				},
			})
			shouldQueries = append(shouldQueries,
				map[string]interface{}{
					"match": map[string]interface{}{
						"name": map[string]interface{}{
							"query": transliteratedSearch,
							"boost": 80.0,
						},
					},
				},
				map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":    transliteratedSearch,
						"fields":   []string{"name^8", "name.ngram^4", "brand.text^3", "model.text^3", "category.text^2"},
						"type":     "cross_fields",
						"operator": "or",
						"boost":    8.0,
					},
				},
			)
		}

		if qwertySearch != strings.ToLower(params.Search) && qwertySearch != transliteratedSearch {
			requiredQueries = append(requiredQueries, map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query": qwertySearch, "fields": requiredFields,
					"type": "cross_fields", "operator": "and",
				},
			})
			shouldQueries = append(shouldQueries,
				map[string]interface{}{
					"match": map[string]interface{}{
						"name": map[string]interface{}{
							"query": qwertySearch,
							"boost": 70.0,
						},
					},
				},
			)
		}

		// 4. Фоновый нечёткий поиск для опечаток
		shouldQueries = append(shouldQueries, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":                params.Search,
				"fields":               []string{"name^2", "brand.text^1", "model.text^1"},
				"type":                 "best_fields",
				"fuzziness":            "AUTO:4,7",
				"prefix_length":        2,
				"minimum_should_match": "75%",
				"boost":                0.1,
			},
		})

		searchQuery := map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []map[string]interface{}{{
					"bool": map[string]interface{}{
						"should":               requiredQueries,
						"minimum_should_match": 1,
					},
				}},
				"should":               shouldQueries,
				"minimum_should_match": 1,
			},
		}
		must = append(must, searchQuery)
	}

	if params.Category != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"category.text": params.Category,
			},
		})
	}

	if params.HasPhoto != "" && params.HasPhoto != "all" {
		if params.HasPhoto == "with" {
			filter = append(filter, map[string]interface{}{
				"exists": map[string]interface{}{
					"field": "photos",
				},
			})
		} else if params.HasPhoto == "without" {
			filter = append(filter, map[string]interface{}{
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

	if params.Brand != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"brand.text": params.Brand,
			},
		})
	}

	if params.Model != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"model.text": params.Model,
			},
		})
	}

	if params.Location != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"location.text": params.Location,
			},
		})
	}

	if params.Address != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"address.text": params.Address,
			},
		})
	}

	if params.Salesman != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"salesman.text": params.Salesman,
			},
		})
	}

	if params.Status != "" {
		statusBool := params.Status == "true" || params.Status == "active" || params.Status == "1"
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{
				"status": statusBool,
			},
		})
	}

	boolQuery := map[string]interface{}{}
	if len(must) > 0 {
		boolQuery["must"] = must
	}
	if len(filter) > 0 {
		boolQuery["filter"] = filter
	}

	return map[string]interface{}{
		"bool": boolQuery,
	}
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
		filters["category_ilike"] = params.Category
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

func (s *inventoryService) DecreasePartQuantity(ctx context.Context, id int64, amount int) error {
	// Сброс кэша
	s.redis.Del(ctx, inventoryCacheKey)
	s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
	s.redis.Del(ctx, statisticsCacheKey)

	err := s.repo.DecreaseQuantity(ctx, id, amount)
	if err != nil {
		return err
	}

	part, err := s.GetPartByID(ctx, id)
	if err == nil && s.es != nil {
		s.es.IndexPart(part)
	}

	return nil
}

func (s *inventoryService) IncreasePartQuantity(ctx context.Context, id int64, amount int) error {
	// Сброс кэша
	s.redis.Del(ctx, inventoryCacheKey)
	s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
	s.redis.Del(ctx, statisticsCacheKey)

	err := s.repo.IncreaseQuantity(ctx, id, amount)
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

	logrus.WithField("totalEarnings", s.totalEarnings).Info("Updated total earnings in database")
	return nil
}
