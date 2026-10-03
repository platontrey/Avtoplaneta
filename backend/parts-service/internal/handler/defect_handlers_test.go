package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"parts-service/internal/catalog"
	"parts-service/internal/domain"
	"parts-service/internal/service"
)

type recordingInventoryService struct {
	MockInventoryService
	mu    sync.Mutex
	parts []domain.Part
}

func (s *recordingInventoryService) AddPart(_ context.Context, part *domain.Part) (*domain.Part, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	copyOfPart := *part
	copyOfPart.ID = int64(len(s.parts) + 1)
	s.parts = append(s.parts, copyOfPart)
	return &copyOfPart, nil
}

func (s *recordingInventoryService) snapshot() []domain.Part {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]domain.Part(nil), s.parts...)
}

func (s *recordingInventoryService) AddPartsBatch(_ context.Context, parts []domain.Part) ([]domain.Part, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range parts {
		s.parts = append(s.parts, parts[i])
	}
	return parts, nil
}

type failingInventoryService struct {
	recordingInventoryService
}

func (failingInventoryService) AddPartsBatch(context.Context, []domain.Part) ([]domain.Part, error) {
	return nil, errors.New("ошибка базы данных при пакетной вставке")
}

func findRecordedPart(t *testing.T, parts []domain.Part, category string) domain.Part {
	t.Helper()
	for _, part := range parts {
		if part.Category == category {
			return part
		}
	}
	t.Fatalf("created parts contain no category %q", category)
	return domain.Part{}
}

func findExpandedPart(t *testing.T, parts []catalog.DefectReportPart, category string) catalog.DefectReportPart {
	t.Helper()
	for _, part := range parts {
		if part.Category == category {
			return part
		}
	}
	t.Fatalf("expanded parts contain no category %q", category)
	return catalog.DefectReportPart{}
}

func TestDefectReportHTTPSynchronousBatch(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cat, err := catalog.LoadPartCatalog()
	require.NoError(t, err)

	recordingService := &recordingInventoryService{}
	h := NewHandler(recordingService, cat, nil, service.NewDefectReportWorkflow(cat, recordingService))
	router := gin.New()
	router.POST("/api/defect-reports", h.CreateDefectReportHandler)

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
		"catalog_version":    cat.Version,
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/defect-reports", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	createdParts := recordingService.snapshot()
	require.Len(t, createdParts, len(cat.Parts))

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
	cat, err := catalog.LoadPartCatalog()
	require.NoError(t, err)

	recordingService := &recordingInventoryService{}
	h := NewHandler(recordingService, cat, nil, service.NewDefectReportWorkflow(cat, recordingService))
	router := gin.New()
	router.POST("/api/defect-reports/preview", h.PreviewDefectReportHandler)

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
	require.Equal(t, cat.Version, preview.CatalogVersion)
	require.Equal(t, len(cat.Parts), preview.Total)
	require.Len(t, preview.Parts, len(cat.Parts))

	part := findExpandedPart(t, preview.Parts, "Кузов снаружи")
	require.Equal(t, "2011-2017", part.CarReleasePeriod)
	require.Equal(t, "АКПП", part.Transmission)
	require.Equal(t, "U660E", part.TransmissionModel)
	require.Equal(t, "Передний", part.Drive)
	require.NotEqual(t, "Подмененная деталь", preview.Parts[0].Name)
}

func TestDefectReportHTTPSynchronousBatch_WithSelectedParts(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cat, err := catalog.LoadPartCatalog()
	require.NoError(t, err)

	recordingService := &recordingInventoryService{}
	h := NewHandler(recordingService, cat, nil, service.NewDefectReportWorkflow(cat, recordingService))
	router := gin.New()
	router.POST("/api/defect-reports", h.CreateDefectReportHandler)

	payload := map[string]any{
		"brand":              "Audi",
		"model":              "A4",
		"year":               2015,
		"car_release_period": "2011-2015",
		"vin":                "WAUZZZTEST123",
		"body_brand":         "B8",
		"engine_brand":       "CJEB",
		"transmission":       "Вариатор",
		"transmission_model": "0AW",
		"drive":              "Передний",
		"selectedParts": []map[string]any{
			{
				"name":     "Крыло переднее левое",
				"category": "Кузов снаружи",
			},
		},
	}
	body, err := json.Marshal(payload)
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/defect-reports", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Equal(t, http.StatusOK, response.Code, response.Body.String())

	createdParts := recordingService.snapshot()
	require.Len(t, createdParts, 1)
	part := createdParts[0]
	require.Equal(t, "Крыло переднее левое", part.Name)
	require.Equal(t, "2011-2015", part.CarReleasePeriod)
	require.Equal(t, "Вариатор", part.Transmission)
}

func postDefectReport(t *testing.T, h *Handler) *httptest.ResponseRecorder {
	t.Helper()
	gin.SetMode(gin.TestMode)

	body, err := json.Marshal(map[string]any{
		"brand":              "Toyota",
		"model":              "Camry",
		"year":               2015,
		"car_release_period": "2011-2017",
		"vin":                "FAILVIN",
		"mileage":            100000,
	})
	require.NoError(t, err)

	router := gin.New()
	router.POST("/api/defect-reports", h.CreateDefectReportHandler)

	request := httptest.NewRequest(http.MethodPost, "/api/defect-reports", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

func TestCreateDefectReportHandlerReportsCatalogUnavailable(t *testing.T) {
	h := NewHandler(&recordingInventoryService{}, nil, nil, service.NewDefectReportWorkflow(nil, &recordingInventoryService{}))
	response := postDefectReport(t, h)
	require.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
}

func TestCreateDefectReportHandlerReportsServiceUnavailable(t *testing.T) {
	cat, err := catalog.LoadPartCatalog()
	require.NoError(t, err)

	h := NewHandler(&recordingInventoryService{}, cat, nil, service.NewDefectReportWorkflow(cat, nil))
	response := postDefectReport(t, h)
	require.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
}

func TestCreateDefectReportHandlerReportsBatchFailure(t *testing.T) {
	cat, err := catalog.LoadPartCatalog()
	require.NoError(t, err)

	h := NewHandler(&recordingInventoryService{}, cat, nil, service.NewDefectReportWorkflow(cat, &failingInventoryService{}))
	response := postDefectReport(t, h)
	require.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
}

func TestPreviewDefectReportHandlerReportsCatalogUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	h := NewHandler(&recordingInventoryService{}, nil, nil, nil)
	router := gin.New()
	router.POST("/api/defect-reports/preview", h.PreviewDefectReportHandler)

	body, err := json.Marshal(map[string]any{"brand": "Toyota", "model": "Camry", "year": 2015})
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/defect-reports/preview", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}
