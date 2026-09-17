package main

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// versionedInventoryService отдаёт фиксированный отпечаток склада, чтобы
// проверить условные ответы без базы.
type versionedInventoryService struct {
	*recordingInventoryService
	version string
}

func (service *versionedInventoryService) InventoryVersion(context.Context) (string, error) {
	return service.version, nil
}

func conditionalRouter(t *testing.T, version string) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	service := &versionedInventoryService{
		recordingInventoryService: &recordingInventoryService{},
		version:                   version,
	}
	handler := NewHandler(service, nil, nil, nil)

	router := gin.New()
	router.GET("/api/statistics", handler.GetStatisticsHandler)
	router.GET("/api/admin/supplier-codes", handler.GetSupplierCodesHandler)
	return router
}

// Дорогие ответы должны переставать собираться, когда склад не менялся.
func TestExpensiveEndpointsRevalidate(t *testing.T) {
	for _, path := range []string{"/api/statistics", "/api/admin/supplier-codes"} {
		t.Run(path, func(t *testing.T) {
			router := conditionalRouter(t, "1700000000-42")

			first := httptest.NewRecorder()
			router.ServeHTTP(first, httptest.NewRequest(http.MethodGet, path, nil))
			require.Equal(t, http.StatusOK, first.Code)

			etag := first.Header().Get("ETag")
			require.Equal(t, `"1700000000-42"`, etag)
			require.NotEmpty(t, first.Header().Get("Cache-Control"))

			second := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, path, nil)
			request.Header.Set("If-None-Match", etag)
			router.ServeHTTP(second, request)

			require.Equal(t, http.StatusNotModified, second.Code)
			require.Empty(t, second.Body.Bytes())

			// Прокси вправе ослабить валидатор — это та же версия.
			weak := httptest.NewRecorder()
			weakRequest := httptest.NewRequest(http.MethodGet, path, nil)
			weakRequest.Header.Set("If-None-Match", `W/`+etag)
			router.ServeHTTP(weak, weakRequest)
			require.Equal(t, http.StatusNotModified, weak.Code)

			// Склад изменился — отдаём тело.
			stale := httptest.NewRecorder()
			staleRequest := httptest.NewRequest(http.MethodGet, path, nil)
			staleRequest.Header.Set("If-None-Match", `"1699999999-41"`)
			router.ServeHTTP(stale, staleRequest)
			require.Equal(t, http.StatusOK, stale.Code)
		})
	}
}

// Пустая версия означает «валидатора нет»: обработчик работает как раньше.
func TestExpensiveEndpointsWithoutVersion(t *testing.T) {
	router := conditionalRouter(t, "")

	response := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/statistics", nil)
	request.Header.Set("If-None-Match", "*")
	router.ServeHTTP(response, request)

	require.Equal(t, http.StatusOK, response.Code)
	require.Empty(t, response.Header().Get("ETag"))
}
