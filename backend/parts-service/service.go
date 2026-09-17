package main

import (
	"context"
	"encoding/json"
	"fmt"
	"runtime"
	"strconv"
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
	InventoryVersion(ctx context.Context) (string, error)                           // Отпечаток состояния склада для условных запросов

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
	Search             string
	Category           string
	Brand              string
	Model              string
	Location           string
	Address            string
	Salesman           string
	Status             string
	HasPhoto           string
	Number             string
	OEMCode            string
	VIN                string
	BodyBrand          string
	EngineBrand        string
	CarReleaseDate     string
	CarReleasePeriod   string
	Transmission       string
	Drive              string
	Condition          string
	Manufacturer       string
	Defect             string
	Color              string
	MinPrice           string
	MaxPrice           string
	MinQuantity        string
	MaxQuantity        string
	FrontRear          string
	LeftRight          string
	TopBottom          string
	ManufacturerCode   string
	SupplierCode       string
	TransmissionModel  string
	WearPercentage     string
	Season             string
	Diameter           string
	Width              string
	Profile            string
	TireQuantity       string
	Drilling           string
	Offset             string
	CenterHoleDiameter string
	TireModel          string
	Page               int
	Limit              int
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

// escapeESQuery экранирует спецсимволы синтаксиса Lucene для безопасного использования в query_string
func escapeESQuery(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '+', '-', '=', '!', '(', ')', '{', '}', '[', ']', '^', '"', '~', '*', '?', ':', '\\', '/', '&', '|', '<', '>':
			sb.WriteRune('\\')
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// ExpandFrontRearSynonyms разворачивает любое обозначение положения перед/зад (русское, английское, F/R)
// во все возможные синонимы для поиска в БД и Elasticsearch.
func ExpandFrontRearSynonyms(val string) []string {
	clean := strings.ToLower(strings.TrimSpace(val))
	switch {
	case clean == "f" || clean == "front" || strings.HasPrefix(clean, "перед"):
		return []string{"F", "f", "Front", "front", "перед", "передний", "передняя", "переднее", "передние", "перед / зад", "перед/зад", "F/R", "F / R"}
	case clean == "r" || clean == "rear" || strings.HasPrefix(clean, "зад"):
		return []string{"R", "r", "Rear", "rear", "зад", "задний", "задняя", "заднее", "задние", "перед / зад", "перед/зад", "F/R", "F / R"}
	case clean == "перед / зад" || clean == "перед/зад" || clean == "f/r" || clean == "f / r":
		return []string{"перед / зад", "перед/зад", "F/R", "F / R", "F", "R"}
	default:
		return []string{val}
	}
}

// ExpandLeftRightSynonyms разворачивает любое обозначение стороны право/лево (русское, английское, L/R)
// во все возможные синонимы для поиска в БД и Elasticsearch.
func ExpandLeftRightSynonyms(val string) []string {
	clean := strings.ToLower(strings.TrimSpace(val))
	switch {
	case clean == "r" || clean == "right" || strings.HasPrefix(clean, "прав"):
		return []string{"R", "r", "Right", "right", "право", "правый", "правая", "правое", "правые", "прав", "лево / право", "лево/право", "L/R", "L / R"}
	case clean == "l" || clean == "left" || strings.HasPrefix(clean, "лев"):
		return []string{"L", "l", "Left", "left", "лево", "левый", "левая", "левое", "левые", "лев", "лево / право", "лево/право", "L/R", "L / R"}
	case clean == "лево / право" || clean == "лево/право" || clean == "l/r" || clean == "l / r":
		return []string{"лево / право", "лево/право", "L/R", "L / R", "L", "R"}
	default:
		return []string{val}
	}
}

// ExpandTopBottomSynonyms разворачивает любое обозначение вертикального положения верх/низ
// во все возможные синонимы для поиска в БД и Elasticsearch.
func ExpandTopBottomSynonyms(val string) []string {
	clean := strings.ToLower(strings.TrimSpace(val))
	switch {
	case clean == "t" || clean == "u" || clean == "top" || clean == "upper" || strings.HasPrefix(clean, "верх"):
		return []string{"T", "t", "U", "u", "Top", "top", "Upper", "upper", "верх", "верхний", "верхняя", "верхнее", "верхние", "верх / низ", "верх/низ"}
	case clean == "b" || clean == "bottom" || clean == "lower" || strings.HasPrefix(clean, "низ"):
		return []string{"B", "b", "L", "l", "Bottom", "bottom", "Lower", "lower", "низ", "нижний", "нижняя", "нижнее", "нижние", "верх / низ", "верх/низ"}
	case clean == "верх / низ" || clean == "верх/низ" || clean == "t/b" || clean == "u/l":
		return []string{"верх / низ", "верх/низ", "T/B", "U/L", "T", "B"}
	default:
		return []string{val}
	}
}

// GetPositionTermQueries проверяет, является ли терм маркером расположения/стороны,
// и возвращает запросы к колонкам front_rear, left_right, top_bottom.
func GetPositionTermQueries(term string) []map[string]interface{} {
	clean := strings.ToLower(strings.TrimSpace(term))
	var queries []map[string]interface{}

	// Проверка на Перед: "перед", "передний", "передняя", "переднее", "передние", "front"
	if strings.HasPrefix(clean, "перед") || clean == "front" {
		synonyms := []interface{}{"F", "f", "Front", "front", "перед", "передний", "передняя", "переднее", "перед / зад"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"front_rear": synonyms,
			},
		})
	}

	// Проверка на Зад: "зад", "задний", "задняя", "заднее", "задние", "rear"
	if strings.HasPrefix(clean, "зад") || clean == "rear" {
		synonyms := []interface{}{"R", "r", "Rear", "rear", "зад", "задний", "задняя", "заднее", "перед / зад"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"front_rear": synonyms,
			},
		})
	}

	// Проверка на Право: "прав", "правый", "правая", "правое", "правые", "право", "right"
	if strings.HasPrefix(clean, "прав") || clean == "right" {
		synonyms := []interface{}{"R", "r", "Right", "right", "право", "правый", "правая", "правое", "лево / право"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"left_right": synonyms,
			},
		})
	}

	// Проверка на Лево: "лев", "левый", "левая", "левое", "левые", "лево", "left"
	if strings.HasPrefix(clean, "лев") || clean == "left" {
		synonyms := []interface{}{"L", "l", "Left", "left", "лево", "левый", "левая", "левое", "лево / право"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"left_right": synonyms,
			},
		})
	}

	// Проверка на Верх: "верх", "верхний", "верхняя", "верхнее", "top", "upper"
	if strings.HasPrefix(clean, "верх") || clean == "top" || clean == "upper" {
		synonyms := []interface{}{"T", "t", "U", "u", "Top", "top", "Upper", "upper", "верх", "верхний", "верхняя", "верх / низ"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"top_bottom": synonyms,
			},
		})
	}

	// Проверка на Низ: "низ", "нижний", "нижняя", "нижнее", "bottom", "lower"
	if strings.HasPrefix(clean, "низ") || clean == "bottom" || clean == "lower" {
		synonyms := []interface{}{"B", "b", "L", "l", "Bottom", "bottom", "Lower", "lower", "низ", "нижний", "нижняя", "верх / низ"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"top_bottom": synonyms,
			},
		})
	}

	return queries
}

// buildElasticsearchQuery строит запрос для Elasticsearch
func (s *inventoryService) buildElasticsearchQuery(params InventoryQueryParams) map[string]interface{} {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{}

	should := []map[string]interface{}{}

	// Фильтр для отображения валидных запчастей (quantity >= 0)
	filter = append(filter, map[string]interface{}{
		"range": map[string]interface{}{
			"quantity": map[string]interface{}{
				"gte": 0,
			},
		},
	})

	if params.Search != "" {
		terms := strings.Fields(params.Search)
		transliteratedSearch := TransliterateLatinToCyrillic(params.Search)
		qwertySearch := ConvertQwertyToRussian(params.Search)

		searchableFields := []string{
			"name^10", "name.ngram^5",
			"brand^4", "brand.text^4", "brand.ngram^3",
			"model^4", "model.text^4", "model.ngram^3",
			"body_brand^3", "body_brand.text^3", "body_brand.ngram^2",
			"engine_brand^3", "engine_brand.text^3", "engine_brand.ngram^2",
			"number^4", "number.text^4",
			"oem_code^4", "oem_code.text^4",
			"manufacturer_code^3", "manufacturer_code.text^3",
			"supplier_code^2", "supplier_code.text^2",
			"vin^3", "vin.text^3",
			"category^3", "category.text^3", "category.ngram^2",
			"car_release_date^3", "car_release_date.text^3", "car_release_date.ngram^2",
			"car_release_period^3", "car_release_period.text^3", "car_release_period.ngram^2",
			"front_rear^3", "front_rear.text^3", "front_rear.ngram^2",
			"left_right^3", "left_right.text^3", "left_right.ngram^2",
			"top_bottom^3", "top_bottom.text^3", "top_bottom.ngram^2",
			"color^3", "color.text^3", "color.ngram^2",
			"condition^2", "condition.text^2", "condition.ngram^1",
			"transmission^3", "transmission.text^3", "transmission.ngram^2",
			"transmission_model^3", "transmission_model.text^3", "transmission_model.ngram^2",
			"drive^3", "drive.text^3", "drive.ngram^2",
			"manufacturer^2", "manufacturer.text^2", "manufacturer.ngram^1",
			"defect^2", "defect.text^2",
			"season^3", "season.text^3", "season.ngram^2",
			"diameter^2", "diameter.text^2",
			"width^2", "width.text^2",
			"profile^2", "profile.text^2",
			"drilling^2", "drilling.text^2",
			"offset^2", "offset.text^2",
			"center_hole_diameter^2", "center_hole_diameter.text^2",
			"tire_model^3", "tire_model.text^3", "tire_model.ngram^2",
			"tire_quantity^1", "tire_quantity.text^1",
			"wear_percentage^1", "wear_percentage.text^1",
			"location^1", "location.text^1",
			"address^1", "address.text^1",
			"salesman^1", "salesman.text^1",
			"description^1",
		}

		// 1. Обязательное совпадение: каждый терм поискового запроса должен присутствовать
		// в запчасти (по любому из полей или как префикс не завершенного слова).
		for _, term := range terms {
			if strings.TrimSpace(term) == "" {
				continue
			}

			variants := []string{term}
			tTrans := TransliterateLatinToCyrillic(term)
			if tTrans != strings.ToLower(term) {
				variants = append(variants, tTrans)
			}
			tQwerty := ConvertQwertyToRussian(term)
			if tQwerty != strings.ToLower(term) && tQwerty != tTrans {
				variants = append(variants, tQwerty)
			}

			termQueries := []map[string]interface{}{}
			for _, variant := range variants {
				escapedVar := escapeESQuery(variant)

				// 1. Точное / ngram / стеммированное совпадение по всем полям
				termQueries = append(termQueries, map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":  variant,
						"fields": searchableFields,
						"type":   "best_fields",
					},
				})

				// 2. Префиксный поиск через match_bool_prefix (когда слово не дописано)
				termQueries = append(termQueries, map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":  variant,
						"fields": searchableFields,
						"type":   "bool_prefix",
					},
				})

				// 3. Префиксный wildcard (слово*) через query_string по всем полям
				termQueries = append(termQueries, map[string]interface{}{
					"query_string": map[string]interface{}{
						"query":            escapedVar + "*",
						"fields":           searchableFields,
						"default_operator": "OR",
						"analyze_wildcard": true,
						"boost":            2.0,
					},
				})

				// 4. Подстрочный wildcard (*слово*) для фрагментов от 3 символов
				if len([]rune(variant)) >= 3 {
					termQueries = append(termQueries, map[string]interface{}{
						"query_string": map[string]interface{}{
							"query":            "*" + escapedVar + "*",
							"fields":           searchableFields,
							"default_operator": "OR",
							"analyze_wildcard": true,
							"boost":            1.0,
						},
					})
				}

				// 5. Позиционные синонимы (F/R/L и перед/зад/право/лево/верх/низ)
				if posQueries := GetPositionTermQueries(variant); len(posQueries) > 0 {
					termQueries = append(termQueries, posQueries...)
				}
			}

			must = append(must, map[string]interface{}{
				"bool": map[string]interface{}{
					"should":               termQueries,
					"minimum_should_match": 1,
				},
			})
		}

		// 2. Для ранжирования и релевантности добавляем should-запросы
		// с высокими весами для точного совпадения названия, фразового поиска и опечаток
		should = append(should,
			// Точное совпадение в названии (Максимальный приоритет 100.0)
			map[string]interface{}{
				"match": map[string]interface{}{
					"name": map[string]interface{}{
						"query": params.Search,
						"boost": 100.0,
					},
				},
			},
			// Фразовое совпадение с префиксом в названии
			map[string]interface{}{
				"match_phrase_prefix": map[string]interface{}{
					"name": map[string]interface{}{
						"query": params.Search,
						"boost": 50.0,
					},
				},
			},
			// Фразовое совпадение с префиксом по всей строке по всем полям
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  params.Search,
					"fields": searchableFields,
					"type":   "bool_prefix",
					"boost":  40.0,
				},
			},
			// Совпадение всей поисковой фразы целиком по всем полям
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  params.Search,
					"fields": searchableFields,
					"type":   "best_fields",
					"boost":  15.0,
				},
			},
		)

		if transliteratedSearch != strings.ToLower(params.Search) {
			should = append(should,
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
						"query":  transliteratedSearch,
						"fields": searchableFields,
						"type":   "best_fields",
						"boost":  10.0,
					},
				},
			)
		}

		if qwertySearch != strings.ToLower(params.Search) && qwertySearch != transliteratedSearch {
			should = append(should,
				map[string]interface{}{
					"match": map[string]interface{}{
						"name": map[string]interface{}{
							"query": qwertySearch,
							"boost": 70.0,
						},
					},
				},
				map[string]interface{}{
					"multi_match": map[string]interface{}{
						"query":  qwertySearch,
						"fields": searchableFields,
						"type":   "best_fields",
						"boost":  8.0,
					},
				},
			)
		}

		// Нечёткий поиск для компенсации опечаток
		should = append(should, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":                params.Search,
				"fields":               []string{"name^2", "brand.text^1", "model.text^1", "car_release_date.text^1", "car_release_period.text^1", "front_rear.text^1", "color.text^1", "transmission.text^1"},
				"type":                 "best_fields",
				"fuzziness":            "AUTO:4,7",
				"prefix_length":        2,
				"minimum_should_match": "75%",
				"boost":                0.1,
			},
		})
	}

	if params.Category != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{
						"term": map[string]interface{}{
							"category": params.Category,
						},
					},
					{
						"match": map[string]interface{}{
							"category.text": params.Category,
						},
					},
				},
				"minimum_should_match": 1,
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
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"brand": params.Brand}},
					{"match": map[string]interface{}{"brand.text": params.Brand}},
					{"match": map[string]interface{}{"brand.ngram": params.Brand}},
					{"wildcard": map[string]interface{}{"brand": "*" + strings.ToLower(params.Brand) + "*"}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Model != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"model": params.Model}},
					{"match": map[string]interface{}{"model.text": params.Model}},
					{"match": map[string]interface{}{"model.ngram": params.Model}},
					{"wildcard": map[string]interface{}{"model": "*" + strings.ToLower(params.Model) + "*"}},
				},
				"minimum_should_match": 1,
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

	if params.Number != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"number": params.Number}},
					{"match": map[string]interface{}{"number.text": params.Number}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.OEMCode != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"oem_code": params.OEMCode}},
					{"match": map[string]interface{}{"oem_code.text": params.OEMCode}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.VIN != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"vin": params.VIN}},
					{"match": map[string]interface{}{"vin.text": params.VIN}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.BodyBrand != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"body_brand.text": params.BodyBrand,
			},
		})
	}

	if params.EngineBrand != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"engine_brand.text": params.EngineBrand,
			},
		})
	}

	if params.CarReleaseDate != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"car_release_date": params.CarReleaseDate}},
					{"match": map[string]interface{}{"car_release_date.text": params.CarReleaseDate}},
					{"match": map[string]interface{}{"car_release_date.ngram": params.CarReleaseDate}},
					{"wildcard": map[string]interface{}{"car_release_date": "*" + params.CarReleaseDate + "*"}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.CarReleasePeriod != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"car_release_period": params.CarReleasePeriod}},
					{"match": map[string]interface{}{"car_release_period.text": params.CarReleasePeriod}},
					{"match": map[string]interface{}{"car_release_period.ngram": params.CarReleasePeriod}},
					{"wildcard": map[string]interface{}{"car_release_period": "*" + params.CarReleasePeriod + "*"}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Transmission != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"transmission": params.Transmission}},
					{"match": map[string]interface{}{"transmission.text": params.Transmission}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Drive != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"drive": params.Drive}},
					{"match": map[string]interface{}{"drive.text": params.Drive}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Condition != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"condition": params.Condition,
			},
		})
	}

	if params.Manufacturer != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"manufacturer.text": params.Manufacturer,
			},
		})
	}

	if params.Defect != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"defect": params.Defect,
			},
		})
	}

	if params.Color != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"color.text": params.Color,
			},
		})
	}

	if params.MinPrice != "" || params.MaxPrice != "" {
		priceRange := map[string]interface{}{}
		if params.MinPrice != "" {
			if minP, err := strconv.ParseFloat(params.MinPrice, 64); err == nil {
				priceRange["gte"] = minP
			}
		}
		if params.MaxPrice != "" {
			if maxP, err := strconv.ParseFloat(params.MaxPrice, 64); err == nil {
				priceRange["lte"] = maxP
			}
		}
		if len(priceRange) > 0 {
			filter = append(filter, map[string]interface{}{
				"range": map[string]interface{}{
					"price": priceRange,
				},
			})
		}
	}

	if params.MinQuantity != "" || params.MaxQuantity != "" {
		qtyRange := map[string]interface{}{}
		if params.MinQuantity != "" {
			if minQ, err := strconv.Atoi(params.MinQuantity); err == nil {
				qtyRange["gte"] = minQ
			}
		}
		if params.MaxQuantity != "" {
			if maxQ, err := strconv.Atoi(params.MaxQuantity); err == nil {
				qtyRange["lte"] = maxQ
			}
		}
		if len(qtyRange) > 0 {
			filter = append(filter, map[string]interface{}{
				"range": map[string]interface{}{
					"quantity": qtyRange,
				},
			})
		}
	}

	if params.FrontRear != "" {
		synonyms := ExpandFrontRearSynonyms(params.FrontRear)
		synInterfaces := make([]interface{}, len(synonyms))
		for i, v := range synonyms {
			synInterfaces[i] = v
		}
		shouldClauses := []map[string]interface{}{
			{"terms": map[string]interface{}{"front_rear": synInterfaces}},
		}
		for _, syn := range synonyms {
			shouldClauses = append(shouldClauses, map[string]interface{}{
				"match": map[string]interface{}{"front_rear.text": syn},
			})
		}
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		})
	}

	if params.LeftRight != "" {
		synonyms := ExpandLeftRightSynonyms(params.LeftRight)
		synInterfaces := make([]interface{}, len(synonyms))
		for i, v := range synonyms {
			synInterfaces[i] = v
		}
		shouldClauses := []map[string]interface{}{
			{"terms": map[string]interface{}{"left_right": synInterfaces}},
		}
		for _, syn := range synonyms {
			shouldClauses = append(shouldClauses, map[string]interface{}{
				"match": map[string]interface{}{"left_right.text": syn},
			})
		}
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		})
	}

	if params.TopBottom != "" {
		synonyms := ExpandTopBottomSynonyms(params.TopBottom)
		synInterfaces := make([]interface{}, len(synonyms))
		for i, v := range synonyms {
			synInterfaces[i] = v
		}
		shouldClauses := []map[string]interface{}{
			{"terms": map[string]interface{}{"top_bottom": synInterfaces}},
		}
		for _, syn := range synonyms {
			shouldClauses = append(shouldClauses, map[string]interface{}{
				"match": map[string]interface{}{"top_bottom.text": syn},
			})
		}
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		})
	}

	if params.ManufacturerCode != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"manufacturer_code": params.ManufacturerCode}},
					{"match": map[string]interface{}{"manufacturer_code.text": params.ManufacturerCode}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.SupplierCode != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"supplier_code": params.SupplierCode}},
					{"match": map[string]interface{}{"supplier_code.text": params.SupplierCode}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.TransmissionModel != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"transmission_model": params.TransmissionModel}},
					{"match": map[string]interface{}{"transmission_model.text": params.TransmissionModel}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.WearPercentage != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"wear_percentage": params.WearPercentage}},
					{"match": map[string]interface{}{"wear_percentage.text": params.WearPercentage}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Season != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"season": params.Season}},
					{"match": map[string]interface{}{"season.text": params.Season}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Diameter != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"diameter": params.Diameter}},
					{"match": map[string]interface{}{"diameter.text": params.Diameter}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Width != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"width": params.Width}},
					{"match": map[string]interface{}{"width.text": params.Width}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Profile != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"profile": params.Profile}},
					{"match": map[string]interface{}{"profile.text": params.Profile}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.TireQuantity != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"tire_quantity": params.TireQuantity}},
					{"match": map[string]interface{}{"tire_quantity.text": params.TireQuantity}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Drilling != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"drilling": params.Drilling}},
					{"match": map[string]interface{}{"drilling.text": params.Drilling}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Offset != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"offset": params.Offset}},
					{"match": map[string]interface{}{"offset.text": params.Offset}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.CenterHoleDiameter != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"center_hole_diameter": params.CenterHoleDiameter}},
					{"match": map[string]interface{}{"center_hole_diameter.text": params.CenterHoleDiameter}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.TireModel != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"tire_model": params.TireModel}},
					{"match": map[string]interface{}{"tire_model.text": params.TireModel}},
				},
				"minimum_should_match": 1,
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
	if len(should) > 0 {
		boolQuery["should"] = should
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

// InventoryVersion пробрасывает отпечаток склада из репозитория. Кэшировать его
// в Redis не нужно: запрос и так дешёвый, а лишний слой добавил бы окно, в
// котором клиент получал бы 304 на уже изменившиеся данные.
func (s *inventoryService) InventoryVersion(ctx context.Context) (string, error) {
	return s.repo.InventoryVersion(ctx)
}
