package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestDefectReportHTTPSynchronousBatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	recordingService := &recordingInventoryService{}
	handler := NewHandler(recordingService, catalog, nil, NewDefectReportWorkflow(catalog, recordingService))
	router := gin.New()
	router.POST("/api/defect-reports", handler.CreateDefectReportHandler)

	payload := map[string]any{
		"brand":              "BMW",
		"model":              "E90",
		"year":               2011,
		"car_release_period": "2005-2011",
		"vin":                "TESTVIN123456789",
		"mileage":            123456,
		"engine_brand":       "N52B30",
		"body_brand":         "BMW E90",
		"interior_color":     "Бежевый",
		"body_color":         "Черный металлик",
		"transmission":       "АКПП",
		"transmission_model": "6HP19",
		"drive":              "Задний",
		"description":        "Интеграционный тест",
		"catalog_version":    catalog.Version,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/defect-reports", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	createdParts := recordingService.snapshot()
	require.Len(t, createdParts, len(catalog.Parts))
	require.Len(t, createdParts, 2992)

	var supplierCode string
	for _, part := range createdParts {
		require.Equal(t, "BMW", part.Brand)
		require.Equal(t, "E90", part.Model)
		require.Equal(t, "TESTVIN123456789", part.VIN)
		require.Equal(t, "BMW E90", part.BodyBrand)
		require.Equal(t, "N52B30", part.EngineBrand)
		require.Equal(t, "2011", part.CarReleaseDate)
		require.Equal(t, "2005-2011", part.CarReleasePeriod)
		require.Equal(t, "System", part.Salesman)
		require.EqualValues(t, 1, part.SellerID)
		require.EqualValues(t, 0, part.Price)
		require.EqualValues(t, 0, part.Quantity)

		if supplierCode == "" {
			supplierCode = part.SupplierCode
		}
		require.NotEmpty(t, part.SupplierCode)
		require.Equal(t, supplierCode, part.SupplierCode)

		require.Equal(t, "АКПП", part.Transmission, part.Category)
		require.Equal(t, "6HP19", part.TransmissionModel, part.Category)
		require.Equal(t, "Задний", part.Drive, part.Category)
	}

	require.Equal(t, "Бежевый", findRecordedPart(t, createdParts, "Электрооснащение").Color)
	require.Equal(t, "Черный металлик", findRecordedPart(t, createdParts, "Стекла").Color)
	require.Equal(t, "6HP19", findRecordedPart(t, createdParts, "Подвеска ДВС/КПП").TransmissionModel)
	require.Equal(t, "6HP19", findRecordedPart(t, createdParts, "Подвеска передних колес").TransmissionModel)
}

func TestDefectReportPreviewUsesServerCatalog(t *testing.T) {
	gin.SetMode(gin.TestMode)
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	recordingService := &recordingInventoryService{}
	handler := NewHandler(recordingService, catalog, nil, NewDefectReportWorkflow(catalog, recordingService))
	router := gin.New()
	router.POST("/api/defect-reports/preview", handler.PreviewDefectReportHandler)

	payload := map[string]any{
		"brand":              "Toyota",
		"model":              "Camry",
		"year":               2015,
		"car_release_period": "2011-2017",
		"vin":                "TESTVIN",
		"engine_brand":       "2AR-FE",
		"body_brand":         "XV50",
		"transmission":       "АКПП",
		"transmission_model": "U660E",
		"drive":              "Передний",
		// Клиентский список должен быть проигнорирован preview-ом.
		"selectedParts": []map[string]any{{
			"name":     "Подмененная деталь",
			"category": "Другое",
		}},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/defect-reports/preview", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	var preview DefectReportPreviewResponse
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &preview))
	require.Equal(t, catalog.Version, preview.CatalogVersion)
	require.Equal(t, len(catalog.Parts), preview.Total)
	require.Len(t, preview.Parts, len(catalog.Parts))

	part := findExpandedPart(t, preview.Parts, "Кузов снаружи")
	require.Equal(t, "2011-2017", part.CarReleasePeriod)
	require.Equal(t, "АКПП", part.Transmission)
	require.Equal(t, "U660E", part.TransmissionModel)
	require.Equal(t, "Передний", part.Drive)
	require.NotEqual(t, "Подмененная деталь", preview.Parts[0].Name)
}

func TestDefectReportHTTPSynchronousBatch_WithSelectedParts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	recordingService := &recordingInventoryService{}
	handler := NewHandler(recordingService, catalog, nil, NewDefectReportWorkflow(catalog, recordingService))
	router := gin.New()
	router.POST("/api/defect-reports", handler.CreateDefectReportHandler)

	// Клиент присылает сформированные selectedParts без некоторых характеристик
	payload := map[string]any{
		"brand":              "Audi",
		"model":              "A4",
		"year":               2015,
		"car_release_period": "2011-2016",
		"vin":                "WAUZZZ8K9FA123456",
		"mileage":            85000,
		"engine_brand":       "CDNC",
		"body_brand":         "Audi B8",
		"transmission":       "Роботизированная",
		"transmission_model": "DL501",
		"drive":              "Полный",
		"selectedParts": []map[string]any{
			{"name": "КПП в сборе", "category": "Трансмиссия", "quantity": 1, "price": 50000.0, "location": "Стеллаж А-12", "address": "Томск, Мира 1"},
			{"name": "Рычаг передний", "category": "Подвеска передних колес", "quantity": 1, "price": 3000.0},
			{"name": "Подрамник", "category": "Подвеска ДВС/КПП", "quantity": 1, "price": 8000.0},
			{"name": "Редуктор задний", "category": "Подвеска задних колес", "quantity": 1, "price": 15000.0},
			{"name": "Рулевая рейка", "category": "Рулевое управление", "quantity": 1, "price": 12000.0},
			{"name": "Двигатель без навесного", "category": "Двигатель", "quantity": 1, "price": 80000.0},
			{"name": "Бампер передний", "category": "Кузов снаружи", "quantity": 1, "price": 10000.0},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/defect-reports", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-User-ID", "42")
	request.Header.Set("X-User-Name", "AutoTester")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	createdParts := recordingService.snapshot()
	require.Len(t, createdParts, 7)

	for _, part := range createdParts {
		require.Equal(t, "Audi", part.Brand)
		require.Equal(t, "A4", part.Model)
		require.Equal(t, "WAUZZZ8K9FA123456", part.VIN)
		require.Equal(t, "2015", part.CarReleaseDate)
		require.Equal(t, "2011-2016", part.CarReleasePeriod)
		require.Equal(t, "Audi B8", part.BodyBrand)
		require.Equal(t, "CDNC", part.EngineBrand)
		require.Equal(t, "AutoTester", part.Salesman)
		require.EqualValues(t, 42, part.SellerID)
	}

	// Трансмиссионные данные описывают автомобиль и должны быть доступны для фильтрации любой запчасти.
	for _, part := range createdParts {
		require.Equal(t, "Роботизированная", part.Transmission, part.Category)
		require.Equal(t, "DL501", part.TransmissionModel, part.Category)
		require.Equal(t, "Полный", part.Drive, part.Category)
	}

	// Место хранения принадлежит конкретной позиции: доезжает до созданной запчасти
	// и не протекает на остальные.
	gearbox := findRecordedPart(t, createdParts, "Трансмиссия")
	require.Equal(t, "Стеллаж А-12", gearbox.Location)
	require.Equal(t, "Томск, Мира 1", gearbox.Address)

	bumper := findRecordedPart(t, createdParts, "Кузов снаружи")
	require.Empty(t, bumper.Location)
	require.Empty(t, bumper.Address)
}

func TestPartCatalogHTTPRevalidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	handler := NewHandler(nil, catalog, nil, nil)
	router := gin.New()
	router.GET("/api/part-catalog", handler.GetPartCatalogHandler)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/part-catalog", nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, `"`+catalog.Version+`"`, response.Header().Get("ETag"))

	var received PartCatalog
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &received))
	require.Equal(t, catalog.Version, received.Version)
	require.Len(t, received.Parts, 2992)
	require.Equal(t, []string{"Передний", "Задний", "Полный"}, attributeOptions(&received, "drive"))
	require.Equal(t, []string{"МКПП", "АКПП", "Роботизированная", "Вариатор"}, attributeOptions(&received, "transmission"))

	revalidationRequest := httptest.NewRequest(http.MethodGet, "/api/part-catalog", nil)
	revalidationRequest.Header.Set("If-None-Match", response.Header().Get("ETag"))
	revalidationResponse := httptest.NewRecorder()
	router.ServeHTTP(revalidationResponse, revalidationRequest)
	require.Equal(t, http.StatusNotModified, revalidationResponse.Code)
	require.Empty(t, revalidationResponse.Body.String())
}

type recordingInventoryService struct {
	mu    sync.Mutex
	parts []Part
}

func (service *recordingInventoryService) AddPart(_ context.Context, part *Part) (*Part, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	copyOfPart := *part
	copyOfPart.ID = int64(len(service.parts) + 1)
	service.parts = append(service.parts, copyOfPart)
	return &copyOfPart, nil
}

func (service *recordingInventoryService) snapshot() []Part {
	service.mu.Lock()
	defer service.mu.Unlock()
	return append([]Part(nil), service.parts...)
}

func (service *recordingInventoryService) GetInventory(context.Context, InventoryQueryParams) ([]Part, error) {
	return nil, nil
}
func (service *recordingInventoryService) UpdatePart(context.Context, int64, map[string]interface{}) error {
	return nil
}
func (service *recordingInventoryService) DeletePart(context.Context, int64) error { return nil }
func (service *recordingInventoryService) MarkPartForDeletion(context.Context, int64) error {
	return nil
}
func (service *recordingInventoryService) GetStatistics(context.Context) (StatisticsResponse, error) {
	return StatisticsResponse{}, nil
}
func (service *recordingInventoryService) InventoryVersion(context.Context) (string, error) {
	return "", nil
}
func (service *recordingInventoryService) BulkDeleteParts(context.Context, []int64) error {
	return nil
}
func (service *recordingInventoryService) BulkUpdateParts(context.Context, []map[string]interface{}) (int, error) {
	return 0, nil
}
func (service *recordingInventoryService) DeleteZeroQuantityPartsBySupplier(context.Context, string) (int64, error) {
	return 0, nil
}
func (service *recordingInventoryService) GetSupplierCodes(context.Context) ([]string, error) {
	return nil, nil
}
func (service *recordingInventoryService) UploadPartPhoto(context.Context, int64, *gin.Context) (string, error) {
	return "", nil
}
func (service *recordingInventoryService) DeletePartPhoto(context.Context, int64, string) error {
	return nil
}
func (service *recordingInventoryService) GetPartByID(context.Context, int64) (*Part, error) {
	return nil, nil
}
func (service *recordingInventoryService) AddPartsBatch(ctx context.Context, parts []Part) ([]Part, error) {
	service.mu.Lock()
	defer service.mu.Unlock()
	for i := range parts {
		service.parts = append(service.parts, parts[i])
	}
	return parts, nil
}
func (service *recordingInventoryService) DecreasePartQuantity(context.Context, int64, int, string) error {
	return nil
}
func (service *recordingInventoryService) IncreasePartQuantity(context.Context, int64, int, string) error {
	return nil
}
func (service *recordingInventoryService) UpdateEarnings(context.Context, float64) error {
	return nil
}

func attributeOptions(catalog *PartCatalog, code string) []string {
	for _, attribute := range catalog.Attributes {
		if attribute.Code == code {
			return attribute.Options
		}
	}
	return nil
}

func findRecordedPart(t *testing.T, parts []Part, category string) Part {
	t.Helper()
	for _, part := range parts {
		if part.Category == category {
			return part
		}
	}
	t.Fatalf("created parts contain no category %q", category)
	return Part{}
}

func (service *recordingInventoryService) RenameSeller(context.Context, int64, string) ([]Part, error) {
	return nil, nil
}

func (service *recordingInventoryService) PartsForExport(context.Context) ([]Part, error) {
	return nil, nil
}
