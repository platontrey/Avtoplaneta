package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestDefectReportHTTPThroughRedisConsumer(t *testing.T) {
	gin.SetMode(gin.TestMode)
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	redisServer := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	t.Cleanup(func() { _ = client.Close() })

	previousRedisClient := redisClient
	redisClient = client
	t.Cleanup(func() { redisClient = previousRedisClient })

	recordingService := &recordingInventoryService{}
	handler := NewHandler(recordingService, catalog)
	router := gin.New()
	router.POST("/api/defect-reports", handler.CreateDefectReportHandler)

	consumer := &RedisEventConsumer{
		client:   client,
		service:  recordingService,
		group:    "parts-service-integration",
		consumer: "worker-integration",
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

	payload := map[string]any{
		"brand":              "BMW",
		"model":              "E90",
		"year":               2011,
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
	require.Equal(t, http.StatusAccepted, response.Code, response.Body.String())

	messages, err := client.XRange(context.Background(), "events:orders", "-", "+").Result()
	require.NoError(t, err)
	require.Len(t, messages, 1)
	require.Equal(t, "defect_report_created", messages[0].Values["type"])

	require.Eventually(t, func() bool {
		return len(recordingService.snapshot()) == len(catalog.Parts)
	}, 10*time.Second, 10*time.Millisecond, "consumer did not create all catalog parts")
	pending, err := client.XPending(context.Background(), "events:orders", consumer.group).Result()
	require.NoError(t, err)
	require.Zero(t, pending.Count, "processed event must be acknowledged")

	cancelConsumer()
	select {
	case consumerErr := <-consumerDone:
		require.NoError(t, consumerErr)
	case <-time.After(6 * time.Second):
		t.Fatal("consumer did not stop after context cancellation")
	}

	createdParts := recordingService.snapshot()
	require.Len(t, createdParts, len(catalog.Parts))
	require.Len(t, createdParts, 1442)

	transmissionModelCategories := bindingCategories(catalog, "transmission_model")
	driveCategories := bindingCategories(catalog, "drive")
	require.Len(t, transmissionModelCategories, 3)
	require.Len(t, driveCategories, 9)

	var supplierCode string
	for _, part := range createdParts {
		require.Equal(t, "BMW", part.Brand)
		require.Equal(t, "E90", part.Model)
		require.Equal(t, "TESTVIN123456789", part.VIN)
		require.Equal(t, "BMW E90", part.BodyBrand)
		require.Equal(t, "N52B30", part.EngineBrand)
		require.Equal(t, "2011", part.CarReleaseDate)
		require.Equal(t, "System", part.Salesman)
		require.EqualValues(t, 1, part.SellerID)

		if supplierCode == "" {
			supplierCode = part.SupplierCode
		}
		require.NotEmpty(t, part.SupplierCode)
		require.Equal(t, supplierCode, part.SupplierCode)

		if transmissionModelCategories[part.Category] {
			require.Equal(t, "6HP19", part.TransmissionModel, part.Category)
		} else {
			require.Empty(t, part.TransmissionModel, part.Category)
		}

		if driveCategories[part.Category] {
			require.Equal(t, "Задний", part.Drive, part.Category)
		} else {
			require.Empty(t, part.Drive, part.Category)
		}

		if part.Category == "Трансмиссия" {
			require.Equal(t, "АКПП", part.Transmission)
		} else {
			require.Empty(t, part.Transmission, part.Category)
		}
	}

	require.Equal(t, "Бежевый", findRecordedPart(t, createdParts, "Электрооснащение").Color)
	require.Equal(t, "Черный металлик", findRecordedPart(t, createdParts, "Стекла").Color)
	require.Equal(t, "6HP19", findRecordedPart(t, createdParts, "Подвеска ДВС/КПП").TransmissionModel)
	require.Equal(t, "6HP19", findRecordedPart(t, createdParts, "Подвеска передних колес").TransmissionModel)
}

func TestPartCatalogHTTPRevalidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	catalog, err := LoadPartCatalog()
	require.NoError(t, err)

	handler := NewHandler(nil, catalog)
	router := gin.New()
	router.GET("/api/part-catalog", handler.GetPartCatalogHandler)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/part-catalog", nil))
	require.Equal(t, http.StatusOK, response.Code)
	require.Equal(t, `"`+catalog.Version+`"`, response.Header().Get("ETag"))

	var received PartCatalog
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &received))
	require.Equal(t, catalog.Version, received.Version)
	require.Len(t, received.Parts, 1442)

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
func (service *recordingInventoryService) DecreasePartQuantity(context.Context, int64, int) error {
	return nil
}
func (service *recordingInventoryService) IncreasePartQuantity(context.Context, int64, int) error {
	return nil
}
func (service *recordingInventoryService) UpdateEarnings(context.Context, float64) error {
	return nil
}

func bindingCategories(catalog *PartCatalog, source string) map[string]bool {
	result := make(map[string]bool)
	for _, binding := range catalog.ReportBindings {
		if binding.Source == source {
			for _, category := range binding.Categories {
				result[category] = true
			}
		}
	}
	return result
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
