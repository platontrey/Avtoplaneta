package main

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// MockElasticsearchClient - мок для Elasticsearch клиента
type MockElasticsearchClient struct {
	mock.Mock
}

func (m *MockElasticsearchClient) IndexPart(part *Part) error {
	args := m.Called(part)
	return args.Error(0)
}

func (m *MockElasticsearchClient) DeletePartFromIndex(partID int64) error {
	args := m.Called(partID)
	return args.Error(0)
}

func (m *MockElasticsearchClient) SearchParts(query map[string]interface{}, from, size int) ([]ElasticsearchPart, int64, error) {
	args := m.Called(query, from, size)
	return args.Get(0).([]ElasticsearchPart), args.Get(1).(int64), args.Error(2)
}

// ServiceTestSuite - набор тестов для сервиса
type ServiceTestSuite struct {
	suite.Suite
	mockRepo *MockPartRepository
	mockES   *MockElasticsearchClient
	service  InventoryService
	testPart *Part
}

func (suite *ServiceTestSuite) SetupTest() {
	suite.mockRepo = new(MockPartRepository)
	suite.mockES = new(MockElasticsearchClient)
	config := &Config{RedisURL: "127.0.0.1:6379"}

	// Настраиваем mock для GetTotalEarnings, который вызывается в NewInventoryService
	suite.mockRepo.On("GetTotalEarnings", mock.Anything).Return(500.0, nil)

	suite.service = NewInventoryService(suite.mockRepo, suite.mockES, config, nil)

	// Создаем тестовую запчасть
	suite.testPart = &Part{
		PartCore: PartCore{
			ID:          1,
			Name:        "Test Part",
			Quantity:    10,
			Description: "Test description",
			Category:    "Test Category",
			Price:       100.0,
		},
	}
}

func (suite *ServiceTestSuite) TearDownTest() {
	suite.mockRepo.AssertExpectations(suite.T())
	suite.mockES.AssertExpectations(suite.T())
}

// TestAddPart - тест добавления запчасти
func (suite *ServiceTestSuite) TestAddPart() {
	newPart := &Part{
		PartCore: PartCore{
			Name:     "New Part",
			Quantity: 5,
			Price:    50.0,
		},
	}

	// Настраиваем моки
	suite.mockRepo.On("Create", mock.Anything, newPart).Return(nil).Run(func(args mock.Arguments) {
		part := args.Get(1).(*Part)
		part.ID = 2
	})
	suite.mockRepo.On("FindByID", mock.Anything, int64(2)).Return(newPart, nil)

	result, err := suite.service.AddPart(context.Background(), newPart)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), "New Part", result.Name)
}

// TestAddPart_InvalidData - тест добавления с невалидными данными
func (suite *ServiceTestSuite) TestAddPart_InvalidData() {
	invalidPart := &Part{
		PartCore: PartCore{
			Name:     "", // пустое имя - невалидно
			Quantity: 5,
		},
	}

	suite.mockRepo.On("Create", mock.Anything, invalidPart).Return(assert.AnError)

	result, err := suite.service.AddPart(context.Background(), invalidPart)
	assert.Error(suite.T(), err)
	assert.Nil(suite.T(), result)
}

// TestUpdatePart - тест обновления запчасти
func (suite *ServiceTestSuite) TestUpdatePart() {
	updates := map[string]interface{}{
		"name":     "Updated Name",
		"quantity": 20,
	}

	// Настраиваем моки
	suite.mockRepo.On("Update", mock.Anything, suite.testPart.ID, updates).Return(nil)
	suite.mockRepo.On("FindByID", mock.Anything, suite.testPart.ID).Return(suite.testPart, nil)

	err := suite.service.UpdatePart(context.Background(), suite.testPart.ID, updates)
	assert.NoError(suite.T(), err)
}

// TestDeletePart - тест удаления запчасти
func (suite *ServiceTestSuite) TestDeletePart() {
	// Настраиваем моки
	suite.mockRepo.On("Delete", mock.Anything, suite.testPart.ID).Return(nil)
	suite.mockRepo.On("FindByID", mock.Anything, suite.testPart.ID).Return(suite.testPart, nil)

	err := suite.service.DeletePart(context.Background(), suite.testPart.ID)
	assert.NoError(suite.T(), err)
}

// TestMarkPartForDeletion - тест отметки для удаления
func (suite *ServiceTestSuite) TestMarkPartForDeletion() {
	suite.mockRepo.On("MarkForDeletion", mock.Anything, suite.testPart.ID, mock.AnythingOfType("time.Time")).Return(nil)

	err := suite.service.MarkPartForDeletion(context.Background(), suite.testPart.ID)
	assert.NoError(suite.T(), err)
}

// TestGetInventory - тест получения инвентаря
func (suite *ServiceTestSuite) TestGetInventory() {
	params := InventoryQueryParams{
		Page:  1,
		Limit: 10,
	}

	expectedParts := []Part{*suite.testPart}
	suite.mockRepo.On("DeleteExpiredParts", mock.Anything, mock.Anything).Return(nil)
	suite.mockRepo.On("FindWithFilters", mock.Anything, mock.Anything, 0, 10).Return(expectedParts, nil)

	parts, err := suite.service.GetInventory(context.Background(), params)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), parts)
	assert.Equal(suite.T(), suite.testPart.Name, parts[0].Name)
}

// TestGetInventory_WithSearch - тест поиска в инвентаре
func (suite *ServiceTestSuite) TestGetInventory_WithSearch() {
	params := InventoryQueryParams{
		Search: "Test",
		Page:   1,
		Limit:  10,
	}

	expectedParts := []Part{*suite.testPart}
	suite.mockRepo.On("DeleteExpiredParts", mock.Anything, mock.Anything).Return(nil)
	suite.mockES.On("SearchParts", mock.AnythingOfType("map[string]interface {}"), 0, 10).Return([]ElasticsearchPart{{ID: suite.testPart.ID, Name: suite.testPart.Name}}, int64(1), nil)
	suite.mockRepo.On("FindWithFilters", mock.Anything, mock.Anything, 0, 0).Return(expectedParts, nil)

	parts, err := suite.service.GetInventory(context.Background(), params)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), parts, 1)
	assert.Equal(suite.T(), suite.testPart.Name, parts[0].Name)
}

// TestGetStatistics - тест получения статистики
func (suite *ServiceTestSuite) TestGetStatistics() {
	suite.mockRepo.On("GetStatistics", mock.Anything).Return(StatisticsResponse{}, nil)
	suite.mockRepo.On("GetTotalEarnings", mock.Anything).Return(500.0, nil)

	stats, err := suite.service.GetStatistics(context.Background())
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), stats)
	assert.GreaterOrEqual(suite.T(), stats.TotalEarnings, 500.0)
}

// TestBulkDeleteParts - тест массового удаления
func (suite *ServiceTestSuite) TestBulkDeleteParts() {
	part2 := &Part{PartCore: PartCore{ID: 2, Name: "Part 2", Quantity: 5}}
	ids := []int64{suite.testPart.ID, part2.ID}

	suite.mockRepo.On("FindByID", mock.Anything, suite.testPart.ID).Return(suite.testPart, nil)
	suite.mockRepo.On("FindByID", mock.Anything, part2.ID).Return(part2, nil)
	suite.mockRepo.On("BulkDelete", mock.Anything, ids).Return(nil)

	err := suite.service.BulkDeleteParts(context.Background(), ids)
	assert.NoError(suite.T(), err)
}

// TestBulkUpdateParts - тест массового обновления
func (suite *ServiceTestSuite) TestBulkUpdateParts() {
	updates := []map[string]interface{}{
		{
			"id":       suite.testPart.ID,
			"name":     "Bulk Updated",
			"quantity": 99,
		},
	}

	suite.mockRepo.On("BulkUpdate", mock.Anything, updates).Return(1, nil)

	count, err := suite.service.BulkUpdateParts(context.Background(), updates)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 1, count)
}

// TestGetPartByID - тест получения запчасти по ID
func (suite *ServiceTestSuite) TestGetPartByID() {
	suite.mockRepo.On("FindByID", mock.Anything, suite.testPart.ID).Return(suite.testPart, nil)

	part, err := suite.service.GetPartByID(context.Background(), suite.testPart.ID)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), part)
	assert.Equal(suite.T(), suite.testPart.Name, part.Name)
}

// TestUpdateEarnings - тест обновления заработка
func (suite *ServiceTestSuite) TestUpdateEarnings() {
	suite.mockRepo.On("UpdateTotalEarnings", mock.Anything, mock.Anything).Return(nil)

	err := suite.service.UpdateEarnings(context.Background(), 100.0)
	assert.NoError(suite.T(), err)
}

// TestBuildElasticsearchQuery_MultiTerm - тест генерации запроса с раздельными словами
func (suite *ServiceTestSuite) TestBuildElasticsearchQuery_MultiTerm() {
	s := suite.service.(*inventoryService)
	params := InventoryQueryParams{
		Search: "АКПП ACV30",
	}
	query := s.buildElasticsearchQuery(params)
	assert.NotNil(suite.T(), query)

	boolQuery, ok := query["bool"].(map[string]interface{})
	assert.True(suite.T(), ok)

	mustClauses, ok := boolQuery["must"].([]map[string]interface{})
	assert.True(suite.T(), ok)
	// Должно быть 2 must clauses: по одному для каждого терма ("АКПП" и "ACV30")
	assert.Equal(suite.T(), 2, len(mustClauses))

	// Должны присутствовать should-клаузы верхнего уровня для релевантности
	shouldClauses, ok := boolQuery["should"].([]map[string]interface{})
	assert.True(suite.T(), ok)
	assert.NotEmpty(suite.T(), shouldClauses)
}

// TestBuildElasticsearchQuery_Category - тест фильтрации по категории
func (suite *ServiceTestSuite) TestBuildElasticsearchQuery_Category() {
	s := suite.service.(*inventoryService)
	params := InventoryQueryParams{
		Category: "Тормозная система",
	}
	query := s.buildElasticsearchQuery(params)
	assert.NotNil(suite.T(), query)

	boolQuery, ok := query["bool"].(map[string]interface{})
	assert.True(suite.T(), ok)

	filters, ok := boolQuery["filter"].([]map[string]interface{})
	assert.True(suite.T(), ok)
	assert.GreaterOrEqual(suite.T(), len(filters), 2)

	// Проверяем наличие фильтрации и по keyword, и по text
	categoryFilter := filters[1]
	catBool, ok := categoryFilter["bool"].(map[string]interface{})
	assert.True(suite.T(), ok)
	shouldList, ok := catBool["should"].([]map[string]interface{})
	assert.True(suite.T(), ok)
	assert.Len(suite.T(), shouldList, 2)
}

// TestBuildElasticsearchQuery_TransliterationAndQwerty - тест транслитерации и исправления раскладки
func (suite *ServiceTestSuite) TestBuildElasticsearchQuery_TransliterationAndQwerty() {
	s := suite.service.(*inventoryService)

	// Тест латиницы (должна транслитерироваться в кириллицу)
	paramsLatin := InventoryQueryParams{
		Search: "bamper toyota",
	}
	queryLatin := s.buildElasticsearchQuery(paramsLatin)
	boolLatin := queryLatin["bool"].(map[string]interface{})
	mustLatin := boolLatin["must"].([]map[string]interface{})
	assert.Equal(suite.T(), 2, len(mustLatin))

	// В первом терме "bamper" должно быть условие для "бампер"
	term1Bool := mustLatin[0]["bool"].(map[string]interface{})
	term1Should := term1Bool["should"].([]map[string]interface{})
	assert.GreaterOrEqual(suite.T(), len(term1Should), 2) // оригинал + транслитерация

	// Тест неверной раскладки QWERTY: "gthtlybq" -> "передний"
	paramsQwerty := InventoryQueryParams{
		Search: "gthtlybq",
	}
	queryQwerty := s.buildElasticsearchQuery(paramsQwerty)
	boolQwerty := queryQwerty["bool"].(map[string]interface{})
	mustQwerty := boolQwerty["must"].([]map[string]interface{})
	assert.Equal(suite.T(), 1, len(mustQwerty))

	termQwertyBool := mustQwerty[0]["bool"].(map[string]interface{})
	termQwertyShould := termQwertyBool["should"].([]map[string]interface{})
	assert.GreaterOrEqual(suite.T(), len(termQwertyShould), 2)
}

// TestBuildElasticsearchQuery_AllFilters - тест добавления всех фильтров
func (suite *ServiceTestSuite) TestBuildElasticsearchQuery_AllFilters() {
	s := suite.service.(*inventoryService)
	params := InventoryQueryParams{
		Category:         "Кузов снаружи",
		Brand:            "Toyota",
		Model:            "Camry",
		Location:         "Стеллаж A-1",
		Address:          "Склад 2",
		Salesman:         "Иванов",
		Status:           "true",
		HasPhoto:         "with",
		Number:           "52119-33939",
		OEMCode:          "5211933939",
		VIN:              "JTM53REV",
		BodyBrand:        "ACV40",
		EngineBrand:      "2AZ-FE",
		CarReleaseDate:   "2008",
		CarReleasePeriod: "2006-2011",
		Transmission:     "АКПП",
		Drive:            "Передний",
		Condition:        "Контрактная",
		Manufacturer:     "Toyota",
		Defect:           "Царапина",
		Color:            "Белый",
	}
	query := s.buildElasticsearchQuery(params)
	boolQuery := query["bool"].(map[string]interface{})
	filters := boolQuery["filter"].([]map[string]interface{})

	// range(quantity >= 0) + category + hasPhoto + brand + model + location + address + salesman + status + 13 характеристик = 22 фильтра
	assert.Equal(suite.T(), 22, len(filters))

	// Проверяем hasPhoto == "without"
	paramsWithoutPhoto := InventoryQueryParams{
		HasPhoto: "without",
	}
	queryWithoutPhoto := s.buildElasticsearchQuery(paramsWithoutPhoto)
	filtersWithout := queryWithoutPhoto["bool"].(map[string]interface{})["filter"].([]map[string]interface{})
	assert.Equal(suite.T(), 2, len(filtersWithout)) // quantity >= 0 + must_not exists photo
}

// TestTransliterationHelpers - тест функций транслитерации и переключения раскладки
func (suite *ServiceTestSuite) TestTransliterationHelpers() {
	assert.Equal(suite.T(), "бампер", TransliterateLatinToCyrillic("bamper"))
	assert.Equal(suite.T(), "сирена", TransliterateLatinToCyrillic("sirena"))
	assert.Equal(suite.T(), "щетка", TransliterateLatinToCyrillic("shchetka"))

	assert.Equal(suite.T(), "передний", ConvertQwertyToRussian("gthtlybq"))
	assert.Equal(suite.T(), "тойота", ConvertQwertyToRussian("njqjnf"))
	assert.Equal(suite.T(), "капот", ConvertQwertyToRussian("rfgjn"))
}

// TestBuildDatabaseFilters - тест сборки фильтров для базы данных
func (suite *ServiceTestSuite) TestBuildDatabaseFilters() {
	s := suite.service.(*inventoryService)
	params := InventoryQueryParams{
		Search:             "АКПП",
		Category:           "Трансмиссия",
		Brand:              "Honda",
		Model:              "Civic",
		Location:           "Стеллаж 5",
		Address:            "Центральный склад",
		Salesman:           "Петров",
		Status:             "true",
		HasPhoto:           "with",
		Number:             "12345",
		OEMCode:            "OEM123",
		VIN:                "VIN123",
		BodyBrand:          "EK3",
		EngineBrand:        "D15B",
		CarReleaseDate:     "1999",
		CarReleasePeriod:   "1995-2000",
		Transmission:       "МКПП",
		Drive:              "Передний",
		Condition:          "Б/у",
		Manufacturer:       "Honda",
		Defect:             "Нет",
		Color:              "Черный",
		MinPrice:           "1000",
		MaxPrice:           "50000",
		MinQuantity:        "1",
		MaxQuantity:        "10",
		FrontRear:          "Передний",
		LeftRight:          "Левый",
		TopBottom:          "Верхний",
		ManufacturerCode:   "MF123",
		SupplierCode:       "SUP456",
		TransmissionModel:  "W58",
		WearPercentage:     "15%",
		Season:             "Зима",
		Diameter:           "R16",
		Width:              "205",
		Profile:            "55",
		TireQuantity:       "4",
		Drilling:           "5x114.3",
		Offset:             "45",
		CenterHoleDiameter: "60.1",
		TireModel:          "Hakkapeliitta 9",
	}

	filters := s.buildDatabaseFilters(params)
	assert.Equal(suite.T(), "АКПП", filters["search"])
	assert.Equal(suite.T(), "Трансмиссия", filters["category_ilike"])
	assert.Equal(suite.T(), "Honda", filters["brand_ilike"])
	assert.Equal(suite.T(), "Civic", filters["model_ilike"])
	assert.Equal(suite.T(), "Стеллаж 5", filters["location_ilike"])
	assert.Equal(suite.T(), "Центральный склад", filters["address_ilike"])
	assert.Equal(suite.T(), "Петров", filters["salesman_ilike"])
	assert.Equal(suite.T(), "true", filters["status"])
	assert.Equal(suite.T(), true, filters["has_photo"])
	assert.Equal(suite.T(), "12345", filters["number_ilike"])
	assert.Equal(suite.T(), "OEM123", filters["oem_code_ilike"])
	assert.Equal(suite.T(), "VIN123", filters["vin_ilike"])
	assert.Equal(suite.T(), "EK3", filters["body_brand_ilike"])
	assert.Equal(suite.T(), "D15B", filters["engine_brand_ilike"])
	assert.Equal(suite.T(), "1999", filters["car_release_date_ilike"])
	assert.Equal(suite.T(), "1995-2000", filters["car_release_period_ilike"])
	assert.Equal(suite.T(), "МКПП", filters["transmission_ilike"])
	assert.Equal(suite.T(), "Передний", filters["drive_ilike"])
	assert.Equal(suite.T(), "Б/у", filters["condition_ilike"])
	assert.Equal(suite.T(), "Honda", filters["manufacturer_ilike"])
	assert.Equal(suite.T(), "Нет", filters["defect_ilike"])
	assert.Equal(suite.T(), "Черный", filters["color_ilike"])
	assert.Equal(suite.T(), float64(1000), filters["min_price"])
	assert.Equal(suite.T(), float64(50000), filters["max_price"])
	assert.Equal(suite.T(), 1, filters["min_quantity"])
	assert.Equal(suite.T(), 10, filters["max_quantity"])
	assert.Equal(suite.T(), "Передний", filters["front_rear_ilike"])
	assert.Equal(suite.T(), "Левый", filters["left_right_ilike"])
	assert.Equal(suite.T(), "Верхний", filters["top_bottom_ilike"])
	assert.Equal(suite.T(), "MF123", filters["manufacturer_code_ilike"])
	assert.Equal(suite.T(), "SUP456", filters["supplier_code_ilike"])
	assert.Equal(suite.T(), "W58", filters["transmission_model_ilike"])
	assert.Equal(suite.T(), "15%", filters["wear_percentage_ilike"])
	assert.Equal(suite.T(), "Зима", filters["season_ilike"])
	assert.Equal(suite.T(), "R16", filters["diameter_ilike"])
	assert.Equal(suite.T(), "205", filters["width_ilike"])
	assert.Equal(suite.T(), "55", filters["profile_ilike"])
	assert.Equal(suite.T(), "4", filters["tire_quantity_ilike"])
	assert.Equal(suite.T(), "5x114.3", filters["drilling_ilike"])
	assert.Equal(suite.T(), "45", filters["offset_ilike"])
	assert.Equal(suite.T(), "60.1", filters["center_hole_diameter_ilike"])
	assert.Equal(suite.T(), "Hakkapeliitta 9", filters["tire_model_ilike"])
}

// TestBuildElasticsearchQuery_NewSpecificationAndRangeFilters - тест новых фильтров в Elasticsearch
func (suite *ServiceTestSuite) TestBuildElasticsearchQuery_NewSpecificationAndRangeFilters() {
	s := suite.service.(*inventoryService)
	params := InventoryQueryParams{
		MinPrice:          "500",
		MaxPrice:          "15000",
		MinQuantity:       "2",
		MaxQuantity:       "20",
		FrontRear:         "Передний",
		Season:            "Зима",
		TransmissionModel: "W58",
	}
	query := s.buildElasticsearchQuery(params)
	boolQuery := query["bool"].(map[string]interface{})
	filters := boolQuery["filter"].([]map[string]interface{})

	// quantity >= 0 + price range + quantity range + front_rear + season + transmission_model = 6
	assert.Equal(suite.T(), 6, len(filters))

	// Проверяем фильтр по диапазону цены
	foundPriceRange := false
	for _, f := range filters {
		if r, ok := f["range"].(map[string]interface{}); ok {
			if pr, ok := r["price"].(map[string]interface{}); ok {
				assert.Equal(suite.T(), float64(500), pr["gte"])
				assert.Equal(suite.T(), float64(15000), pr["lte"])
				foundPriceRange = true
			}
		}
	}
	assert.True(suite.T(), foundPriceRange, "Должен присутствовать фильтр по цене")
}

// TestBuildElasticsearchQuery_DigitsAndReleaseDate - тест поиска с цифрами и годом выпуска
func (suite *ServiceTestSuite) TestBuildElasticsearchQuery_DigitsAndReleaseDate() {
	s := suite.service.(*inventoryService)
	params := InventoryQueryParams{
		Search: "Audi 2010",
	}
	query := s.buildElasticsearchQuery(params)
	assert.NotNil(suite.T(), query)

	boolQuery := query["bool"].(map[string]interface{})
	mustClauses := boolQuery["must"].([]map[string]interface{})
	assert.Equal(suite.T(), 2, len(mustClauses))

	// Проверяем, что car_release_date и model.ngram присутствуют в searchableFields
	term0 := mustClauses[0]["bool"].(map[string]interface{})
	shouldList := term0["should"].([]map[string]interface{})
	firstMatch := shouldList[0]["multi_match"].(map[string]interface{})
	fields := firstMatch["fields"].([]string)

	assert.Contains(suite.T(), fields, "car_release_date^3")
	assert.Contains(suite.T(), fields, "car_release_date.text^3")
	assert.Contains(suite.T(), fields, "car_release_date.ngram^2")
	assert.Contains(suite.T(), fields, "model.ngram^3")
	assert.Contains(suite.T(), fields, "brand.ngram^3")
	assert.Contains(suite.T(), fields, "front_rear^3")
	assert.Contains(suite.T(), fields, "color^3")
	assert.Contains(suite.T(), fields, "transmission^3")
	assert.Contains(suite.T(), fields, "season^3")
	assert.Contains(suite.T(), fields, "tire_model^3")

	// Проверяем наличие префиксного поиска bool_prefix и wildcard в termQueries
	hasBoolPrefix := false
	hasWildcardQuery := false
	for _, q := range shouldList {
		if mm, ok := q["multi_match"].(map[string]interface{}); ok {
			if mm["type"] == "bool_prefix" {
				hasBoolPrefix = true
			}
		}
		if qs, ok := q["query_string"].(map[string]interface{}); ok {
			if queryString, ok := qs["query"].(string); ok && strings.HasSuffix(queryString, "*") {
				hasWildcardQuery = true
			}
		}
	}
	assert.True(suite.T(), hasBoolPrefix, "должен присутствовать bool_prefix для недописанных слов")
	assert.True(suite.T(), hasWildcardQuery, "должен присутствовать wildcard query для префиксов недописанных слов")
}

// TestBuildElasticsearchQuery_IncompleteWords - тест поиска по недописанным словам (автодополнение/префикс)
func (suite *ServiceTestSuite) TestBuildElasticsearchQuery_IncompleteWords() {
	s := suite.service.(*inventoryService)
	// Пользователь ввел "Toyota cam" - второе слово недописано
	params := InventoryQueryParams{
		Search: "Toyota cam",
	}
	query := s.buildElasticsearchQuery(params)
	assert.NotNil(suite.T(), query)

	boolQuery := query["bool"].(map[string]interface{})
	mustClauses := boolQuery["must"].([]map[string]interface{})
	assert.Equal(suite.T(), 2, len(mustClauses))

	// Клауза для "cam"
	termCam := mustClauses[1]["bool"].(map[string]interface{})
	shouldList := termCam["should"].([]map[string]interface{})

	// Должен быть query_string с cam*
	foundPrefix := false
	for _, q := range shouldList {
		if qs, ok := q["query_string"].(map[string]interface{}); ok {
			if qs["query"] == "cam*" {
				foundPrefix = true
			}
		}
	}
	assert.True(suite.T(), foundPrefix, "должен присутствовать query_string с cam* для поиска недописанного слова")
}

// TestBuildElasticsearchQuery_BrandAndModelFilters - тест гибкой фильтрации бренда, модели и года
func (suite *ServiceTestSuite) TestBuildElasticsearchQuery_BrandAndModelFilters() {
	s := suite.service.(*inventoryService)
	params := InventoryQueryParams{
		Brand:          "Audi",
		Model:          "A6",
		CarReleaseDate: "2015",
	}
	query := s.buildElasticsearchQuery(params)
	assert.NotNil(suite.T(), query)

	boolQuery := query["bool"].(map[string]interface{})
	filters := boolQuery["filter"].([]map[string]interface{})
	// range(quantity >= 0) + brand + model + car_release_date = 4 фильтра
	assert.Equal(suite.T(), 4, len(filters))

	// Проверяем структуру фильтра Brand (должен быть bool should с 4 вариантами: term, text, ngram, wildcard)
	brandFilter := filters[1]["bool"].(map[string]interface{})
	brandShould := brandFilter["should"].([]map[string]interface{})
	assert.Equal(suite.T(), 4, len(brandShould))

	// Проверяем структуру фильтра Model
	modelFilter := filters[2]["bool"].(map[string]interface{})
	modelShould := modelFilter["should"].([]map[string]interface{})
	assert.Equal(suite.T(), 4, len(modelShould))

	// Проверяем структуру фильтра CarReleaseDate
	yearFilter := filters[3]["bool"].(map[string]interface{})
	yearShould := yearFilter["should"].([]map[string]interface{})
	assert.Equal(suite.T(), 4, len(yearShould))
}

// TestShouldUseElasticsearch_Specifications - тест переключения на Elasticsearch при фильтрации по спецификациям
func (suite *ServiceTestSuite) TestShouldUseElasticsearch_Specifications() {
	s := suite.service.(*inventoryService)

	// Пустые параметры -> false
	assert.False(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{}))

	// Спецификации по отдельности -> true
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{CarReleaseDate: "2010"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{CarReleasePeriod: "2001-2007"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{Number: "12345"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{OEMCode: "OEM999"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{VIN: "VIN123"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{BodyBrand: "E90"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{EngineBrand: "1NZ"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{Transmission: "АКПП"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{Drive: "Передний"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{Condition: "Б/у"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{Manufacturer: "Toyota"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{Defect: "Царапина"}))
	assert.True(suite.T(), s.shouldUseElasticsearch(InventoryQueryParams{Color: "Белый"}))
}

// TestPositionSynonyms_AndSearch проверяет нормализацию синонимов перед/зад, право/лево, верх/низ и полнотекстовый поиск
func (suite *ServiceTestSuite) TestPositionSynonyms_AndSearch() {
	// 1. Проверяем разворачивание синонимов
	frSyns := ExpandFrontRearSynonyms("зад")
	assert.Contains(suite.T(), frSyns, "R")
	assert.Contains(suite.T(), frSyns, "Rear")
	assert.Contains(suite.T(), frSyns, "зад")
	assert.Contains(suite.T(), frSyns, "задний")

	frontSyns := ExpandFrontRearSynonyms("передний")
	assert.Contains(suite.T(), frontSyns, "F")
	assert.Contains(suite.T(), frontSyns, "Front")
	assert.Contains(suite.T(), frontSyns, "перед")

	lrSyns := ExpandLeftRightSynonyms("право")
	assert.Contains(suite.T(), lrSyns, "R")
	assert.Contains(suite.T(), lrSyns, "Right")
	assert.Contains(suite.T(), lrSyns, "правый")

	leftSyns := ExpandLeftRightSynonyms("левый")
	assert.Contains(suite.T(), leftSyns, "L")
	assert.Contains(suite.T(), leftSyns, "Left")
	assert.Contains(suite.T(), leftSyns, "лево")

	tbSyns := ExpandTopBottomSynonyms("верхний")
	assert.Contains(suite.T(), tbSyns, "T")
	assert.Contains(suite.T(), tbSyns, "U")
	assert.Contains(suite.T(), tbSyns, "верх")

	// 2. Проверяем построение фильтра Elasticsearch для "Зад" и "Право"
	s := suite.service.(*inventoryService)
	params := InventoryQueryParams{
		FrontRear: "Зад",
		LeftRight: "Право",
	}
	esQuery := s.buildElasticsearchQuery(params)
	boolQ := esQuery["bool"].(map[string]interface{})
	filters := boolQ["filter"].([]map[string]interface{})

	// Должны присутствовать фильтры front_rear и left_right
	foundFR := false
	foundLR := false
	for _, f := range filters {
		if b, ok := f["bool"].(map[string]interface{}); ok {
			if shouldList, ok := b["should"].([]map[string]interface{}); ok {
				for _, sh := range shouldList {
					if termsQ, ok := sh["terms"].(map[string]interface{}); ok {
						if frList, ok := termsQ["front_rear"].([]interface{}); ok {
							assert.Contains(suite.T(), frList, "R")
							assert.Contains(suite.T(), frList, "зад")
							foundFR = true
						}
						if lrList, ok := termsQ["left_right"].([]interface{}); ok {
							assert.Contains(suite.T(), lrList, "R")
							assert.Contains(suite.T(), lrList, "право")
							foundLR = true
						}
					}
				}
			}
		}
	}
	assert.True(suite.T(), foundFR, "Фильтр front_rear должен содержать синонимы")
	assert.True(suite.T(), foundLR, "Фильтр left_right должен содержать синонимы")

	// 3. Проверяем, что при текстовом поиске со словами "бампер задний"
	// терм "задний" генерирует terms запрос для front_rear
	searchParams := InventoryQueryParams{
		Search: "бампер задний",
	}
	searchESQuery := s.buildElasticsearchQuery(searchParams)
	searchBoolQ := searchESQuery["bool"].(map[string]interface{})
	mustClauses := searchBoolQ["must"].([]map[string]interface{})
	assert.Equal(suite.T(), 2, len(mustClauses))

	// Ищем clause для "задний"
	term1 := mustClauses[1]["bool"].(map[string]interface{})
	shoulds1 := term1["should"].([]map[string]interface{})
	foundPosTerm := false
	for _, sh := range shoulds1 {
		if termsQ, ok := sh["terms"].(map[string]interface{}); ok {
			if frList, ok := termsQ["front_rear"].([]interface{}); ok {
				assert.Contains(suite.T(), frList, "R")
				assert.Contains(suite.T(), frList, "зад")
				foundPosTerm = true
			}
		}
	}
	assert.True(suite.T(), foundPosTerm, "Текстовый поиск 'задний' должен включать terms front_rear со значением 'R'")
}

// TestRunSuite - запуск всех тестов сервиса
func TestServiceTestSuite(t *testing.T) {
	suite.Run(t, new(ServiceTestSuite))
}
