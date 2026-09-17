package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	partsv1 "avtoplaneta/gen/parts/v1"
)

var errPublishRejected = errors.New("очередь отклонила публикацию")

type failingDefectReportPublisher struct{}

func (failingDefectReportPublisher) PublishDefectReport(context.Context, DefectReportRequest) error {
	return errPublishRejected
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
	handler := NewHandler(&recordingInventoryService{}, nil, NewDefectReportWorkflow(nil, failingDefectReportPublisher{}))
	response := postDefectReport(t, handler)
	require.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
}

// Очередь не сконфигурирована — тоже 503, а не 500.
func TestCreateDefectReportHandlerReportsQueueUnavailable(t *testing.T) {
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	handler := NewHandler(&recordingInventoryService{}, catalog, NewDefectReportWorkflow(catalog, nil))
	response := postDefectReport(t, handler)
	require.Equal(t, http.StatusServiceUnavailable, response.Code, response.Body.String())
}

// Публикация сорвалась на живой очереди — это уже 500.
func TestCreateDefectReportHandlerReportsPublishFailure(t *testing.T) {
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	handler := NewHandler(&recordingInventoryService{}, catalog, NewDefectReportWorkflow(catalog, failingDefectReportPublisher{}))
	response := postDefectReport(t, handler)
	require.Equal(t, http.StatusInternalServerError, response.Code, response.Body.String())
}

func TestPreviewDefectReportHandlerReportsCatalogUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(&recordingInventoryService{}, nil, nil)
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

func TestPartsGRPCServerCreateDefectReportReportsPublishFailure(t *testing.T) {
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	server := NewPartsGRPCServer(&recordingInventoryService{}, NewDefectReportWorkflow(catalog, failingDefectReportPublisher{}))

	_, err = server.CreateDefectReport(context.Background(), &partsv1.CreateDefectReportRequest{
		Brand: "Toyota",
		Model: "Camry",
	})
	require.Error(t, err)
	require.Equal(t, codes.Internal, status.Code(err))
}

// Preview строит набор сервером и ничего не публикует.
func TestPartsGRPCServerPreviewDefectReportBuildsServerSet(t *testing.T) {
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	publisher := &recordingDefectReportPublisher{}
	server := NewPartsGRPCServer(&recordingInventoryService{}, NewDefectReportWorkflow(catalog, publisher))

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
	require.Empty(t, publisher.report.SelectedParts, "preview must not publish anything")

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

// Совместимость: события, попавшие в очередь до выката (event_version отсутствует),
// дополняются характеристиками автомобиля на стороне consumer.
// Удалить вместе с веткой EventVersion < defectReportEventVersion в events_consumer.go.
func TestConsumerFillsVehicleSpecsForLegacyEvents(t *testing.T) {
	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	recordingService := &recordingInventoryService{}
	consumer := &RedisEventConsumer{
		client:   client,
		service:  recordingService,
		group:    "parts-service-legacy",
		consumer: "worker-legacy",
		stream:   "events:orders",
	}
	consumerContext, cancelConsumer := context.WithCancel(context.Background())
	t.Cleanup(cancelConsumer)
	consumerDone := make(chan error, 1)
	go func() { consumerDone <- consumer.Start(consumerContext) }()
	require.Eventually(t, func() bool {
		groups, groupErr := client.XInfoGroups(context.Background(), "events:orders").Result()
		return groupErr == nil && len(groups) == 1
	}, 5*time.Second, 10*time.Millisecond, "consumer group was not created")

	// Ровно тот формат, который писали продюсеры до централизации workflow:
	// без event_version и без характеристик в самих позициях.
	payload, err := json.Marshal(map[string]any{
		"brand":              "Nissan",
		"model":              "X-Trail",
		"year":               2012,
		"car_release_period": "2007-2013",
		"vin":                "LEGACYVIN",
		"mileage":            150000,
		"engine_brand":       "MR20DE",
		"body_brand":         "T31",
		"transmission":       "Вариатор",
		"transmission_model": "JF011E",
		"drive":              "Полный",
		"body_color":         "Серебристый",
		"interior_color":     "Серый",
		"seller_name":        "Legacy",
		"seller_id":          7,
		"selectedParts": []map[string]any{
			{"name": "Бампер передний", "category": "Кузов снаружи", "quantity": 1, "price": 5000.0},
			{"name": "Блок управления", "category": "Электрооснащение", "quantity": 1, "price": 7000.0},
		},
	})
	require.NoError(t, err)

	require.NoError(t, client.XAdd(context.Background(), &redis.XAddArgs{
		Stream: "events:orders",
		Values: map[string]interface{}{
			"type": "defect_report_created",
			"data": string(payload),
		},
	}).Err())

	require.Eventually(t, func() bool {
		return len(recordingService.snapshot()) == 2
	}, 10*time.Second, 10*time.Millisecond, "consumer did not create legacy parts")

	cancelConsumer()
	select {
	case consumerErr := <-consumerDone:
		require.NoError(t, consumerErr)
	case <-time.After(6 * time.Second):
		t.Fatal("consumer did not stop after context cancellation")
	}

	createdParts := recordingService.snapshot()
	for _, part := range createdParts {
		require.Equal(t, "LEGACYVIN", part.VIN, part.Category)
		require.Equal(t, "2012", part.CarReleaseDate, part.Category)
		require.Equal(t, "2007-2013", part.CarReleasePeriod, part.Category)
		require.Equal(t, "T31", part.BodyBrand, part.Category)
		require.Equal(t, "MR20DE", part.EngineBrand, part.Category)
		require.Equal(t, "Вариатор", part.Transmission, part.Category)
		require.Equal(t, "JF011E", part.TransmissionModel, part.Category)
		require.Equal(t, "Полный", part.Drive, part.Category)
	}

	require.Equal(t, "Серебристый", findRecordedPart(t, createdParts, "Кузов снаружи").Color)
	require.Equal(t, "Серый", findRecordedPart(t, createdParts, "Электрооснащение").Color)
}
