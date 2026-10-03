package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"parts-service/internal/catalog"
)

func TestVehicleCatalogHTTPRevalidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	vcat, err := catalog.LoadVehicleCatalog()
	require.NoError(t, err)

	handler := NewHandler(nil, nil, vcat, nil)
	router := gin.New()
	router.GET("/api/vehicle-catalog", handler.GetVehicleCatalogHandler)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/vehicle-catalog", nil))
	require.Equal(t, http.StatusOK, response.Code)

	etag := response.Header().Get("ETag")
	require.Equal(t, `"`+vcat.Version+`"`, etag)

	var received catalog.VehicleCatalog
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &received))
	require.Equal(t, vcat.Version, received.Version)
	require.Len(t, received.Brands, len(vcat.Brands))

	revalidated := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/vehicle-catalog", nil)
	request.Header.Set("If-None-Match", etag)
	router.ServeHTTP(revalidated, request)
	require.Equal(t, http.StatusNotModified, revalidated.Code)
	require.Empty(t, revalidated.Body.Bytes())
}

func TestVehicleCatalogHandlerReportsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil, nil, nil, nil)
	router := gin.New()
	router.GET("/api/vehicle-catalog", handler.GetVehicleCatalogHandler)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/vehicle-catalog", nil))
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}

func TestPartCatalogHTTPRevalidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pcat, err := catalog.LoadPartCatalog()
	require.NoError(t, err)

	handler := NewHandler(nil, pcat, nil, nil)
	router := gin.New()
	router.GET("/api/part-catalog", handler.GetPartCatalogHandler)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/part-catalog", nil))
	require.Equal(t, http.StatusOK, response.Code)

	etag := response.Header().Get("ETag")
	require.Equal(t, `"`+pcat.Version+`"`, etag)

	revalidated := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/part-catalog", nil)
	request.Header.Set("If-None-Match", etag)
	router.ServeHTTP(revalidated, request)
	require.Equal(t, http.StatusNotModified, revalidated.Code)
	require.Empty(t, revalidated.Body.Bytes())
}

func TestPartCatalogHandlerReportsUnavailable(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := NewHandler(nil, nil, nil, nil)
	router := gin.New()
	router.GET("/api/part-catalog", handler.GetPartCatalogHandler)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/part-catalog", nil))
	require.Equal(t, http.StatusServiceUnavailable, response.Code)
}
