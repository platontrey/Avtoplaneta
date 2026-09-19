package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	partsv1 "avtoplaneta/gen/parts/v1"
)

var errBatchFailure = errors.New("ошибка базы данных при пакетной вставке")

type failingInventoryService struct {
	recordingInventoryService
}

func (failingInventoryService) AddPartsBatch(context.Context, []Part) ([]Part, error) {
	return nil, errBatchFailure
}

func postDefectReport(t *testing.T, handler *Handler) *httptest.ResponseRecorder {
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
	router.POST("/api/defect-reports", handler.CreateDefectReportHandler)

	request := httptest.NewRequest(http.MethodPost, "/api/defect-reports", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	return response
}

// Каталог не загружен — это отказ инфраструктуры, а не ошибка клиента.
func TestCreateDefectReportHandlerReportsCatalogUnavailable(t *testing.T) {
	handler := NewHandler(&recordingInventoryService{}, nil, nil, NewDefectReportWorkflow(nil, &recordingInventoryService{}))
	response := postDefectReport(t, handler)
	require.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
}

// Сервис не сконфигурирован — тоже 503, а не 500.
func TestCreateDefectReportHandlerReportsServiceUnavailable(t *testing.T) {
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	handler := NewHandler(&recordingInventoryService{}, catalog, nil, NewDefectReportWorkflow(catalog, nil))
	response := postDefectReport(t, handler)
	require.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
}

// Ошибка пакетной вставки в БД — это уже 500.
func TestCreateDefectReportHandlerReportsBatchFailure(t *testing.T) {
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	handler := NewHandler(&recordingInventoryService{}, catalog, nil, NewDefectReportWorkflow(catalog, &failingInventoryService{}))
	response := postDefectReport(t, handler)
	require.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
}

func TestPreviewDefectReportHandlerReportsCatalogUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(&recordingInventoryService{}, nil, nil, nil)
	router := gin.New()
	router.POST("/api/defect-reports/preview", handler.PreviewDefectReportHandler)

	body, err := json.Marshal(map[string]any{"brand": "Toyota", "model": "Camry", "year": 2015})
	require.NoError(t, err)

	request := httptest.NewRequest(http.MethodPost, "/api/defect-reports/preview", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
}

func TestPartsGRPCServerCreateDefectReportReportsUnavailable(t *testing.T) {
	server := NewPartsGRPCServer(&recordingInventoryService{}, nil)

	_, err := server.CreateDefectReport(context.Background(), &partsv1.CreateDefectReportRequest{
		Brand: "Toyota",
		Model: "Camry",
	})
	require.Error(t, err)
	require.Equal(t, codes.Unavailable, status.Code(err))
}

func TestPartsGRPCServerCreateDefectReportReportsBatchFailure(t *testing.T) {
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	server := NewPartsGRPCServer(&recordingInventoryService{}, NewDefectReportWorkflow(catalog, &failingInventoryService{}))

	_, err = server.CreateDefectReport(context.Background(), &partsv1.CreateDefectReportRequest{
		Brand: "Toyota",
		Model: "Camry",
	})
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))
}

// Preview строит набор сервером и ничего не создает.
func TestPartsGRPCServerPreviewDefectReportBuildsServerSet(t *testing.T) {
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	service := &recordingInventoryService{}
	server := NewPartsGRPCServer(service, NewDefectReportWorkflow(catalog, service))

	response, err := server.PreviewDefectReport(context.Background(), &partsv1.PreviewDefectReportRequest{
		Brand:             "Toyota",
		Model:             "Camry",
		Year:              "2015",
		Vin:               "PREVIEWVIN",
		CarReleasePeriod:  "2011-2017",
		BodyBrand:         "XV50",
		EngineBrand:       "2AR-FE",
		Transmission:      "АКПП",
		TransmissionModel: "U660E",
		Drive:             "Передний",
	})
	require.NoError(t, err)

	require.Equal(t, catalog.Version, response.CatalogVersion)
	require.EqualValues(t, len(catalog.Parts), response.Total)
	require.Len(t, response.Parts, len(catalog.Parts))
	require.Empty(t, service.snapshot(), "preview must not create anything")

	for _, part := range response.Parts {
		require.Equal(t, "PREVIEWVIN", part.Vin, part.Category)
		require.Equal(t, "2015", part.CarReleaseDate, part.Category)
		require.Equal(t, "2011-2017", part.CarReleasePeriod, part.Category)
		require.Equal(t, "U660E", part.TransmissionModel, part.Category)
		require.Equal(t, "Передний", part.Drive, part.Category)
	}
}

func TestPartsGRPCServerPreviewDefectReportReportsCatalogUnavailable(t *testing.T) {
	server := NewPartsGRPCServer(&recordingInventoryService{}, nil)

	_, err := server.PreviewDefectReport(context.Background(), &partsv1.PreviewDefectReportRequest{Brand: "Toyota"})
	require.Error(t, err)
	require.Equal(t, codes.Unavailable, status.Code(err))
}
