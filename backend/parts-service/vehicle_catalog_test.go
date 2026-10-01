package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestVehicleCatalogLoadsAndValidates(t *testing.T) {
	catalog, err := LoadVehicleCatalog()
	require.NoError(t, err)
	require.NotEmpty(t, catalog.Version)
	require.NotEmpty(t, catalog.Brands)

	for _, brand := range catalog.Brands {
		require.NotEmpty(t, brand.Name)
		for _, model := range brand.Models {
			require.NotEmpty(t, model.Name, brand.Name)
		}
	}
}

func TestVehicleCatalogRejectsBrokenData(t *testing.T) {
	cases := map[string]VehicleCatalog{
		"без версии": {Brands: []VehicleBrand{{Name: "Toyota"}}},
		"без марок":  {Version: "1"},
		"дубль марки": {Version: "1", Brands: []VehicleBrand{
			{Name: "Toyota"}, {Name: "toyota"},
		}},
		"дубль модели": {Version: "1", Brands: []VehicleBrand{
			{Name: "Toyota", Models: []VehicleModel{{Name: "Camry"}, {Name: "camry"}}},
		}},
		"марки не отсортированы": {Version: "1", Brands: []VehicleBrand{
			{Name: "Toyota"}, {Name: "Audi"},
		}},
		"модели не отсортированы": {Version: "1", Brands: []VehicleBrand{
			{Name: "Toyota", Models: []VehicleModel{{Name: "Corolla"}, {Name: "Camry"}}},
		}},
	}

	for name, broken := range cases {
		t.Run(name, func(t *testing.T) {
			require.Error(t, validateVehicleCatalog(&broken))
		})
	}
}

func TestVehicleCatalogModelsOfIgnoresCase(t *testing.T) {
	catalog := &VehicleCatalog{Version: "1", Brands: []VehicleBrand{
		{Name: "Toyota", Models: []VehicleModel{{Name: "Camry"}}},
	}}

	require.Len(t, catalog.ModelsOf("toyota"), 1)
	require.Len(t, catalog.ModelsOf("  TOYOTA "), 1)
	require.Empty(t, catalog.ModelsOf("Audi"))

	var missing *VehicleCatalog
	require.Empty(t, missing.ModelsOf("Toyota"))
}

func TestVehicleCatalogBodiesAndEnginesOf(t *testing.T) {
	catalog := &VehicleCatalog{Version: "1", Brands: []VehicleBrand{
		{
			Name:    "Toyota",
			Bodies:  []string{"ACV40", "NZE121"},
			Engines: []string{"1NZ-FE", "2AZ-FE"},
			Models: []VehicleModel{
				{
					Name:    "Corolla",
					Bodies:  []string{"NZE121"},
					Engines: []string{"1NZ-FE"},
				},
				{
					Name: "EmptyModel",
				},
			},
		},
	}}

	// Проверка тел модели
	require.Equal(t, []string{"NZE121"}, catalog.BodiesOf("toyota", "corolla"))
	require.Equal(t, []string{"1NZ-FE"}, catalog.EnginesOf("toyota", "corolla"))

	// Модель без своих данных откатывается на марку
	require.Equal(t, []string{"ACV40", "NZE121"}, catalog.BodiesOf("toyota", "EmptyModel"))
	require.Equal(t, []string{"1NZ-FE", "2AZ-FE"}, catalog.EnginesOf("toyota", "EmptyModel"))

	// Поиск без модели возвращает данные марки
	require.Equal(t, []string{"ACV40", "NZE121"}, catalog.BodiesOf("toyota", ""))
	require.Equal(t, []string{"1NZ-FE", "2AZ-FE"}, catalog.EnginesOf("toyota", ""))

	// Несуществующая марка
	require.Nil(t, catalog.BodiesOf("unknown", "corolla"))
	require.Nil(t, catalog.EnginesOf("unknown", "corolla"))
}

// Справочник кэшируется клиентами по версии, поэтому ETag и 304 — часть контракта.
func TestVehicleCatalogHTTPRevalidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	catalog, err := LoadVehicleCatalog()
	require.NoError(t, err)

	handler := NewHandler(nil, nil, catalog, nil)
	router := gin.New()
	router.GET("/api/vehicle-catalog", handler.GetVehicleCatalogHandler)

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/vehicle-catalog", nil))
	require.Equal(t, http.StatusOK, response.Code)

	etag := response.Header().Get("ETag")
	require.Equal(t, `"`+catalog.Version+`"`, etag)

	var received VehicleCatalog
	require.NoError(t, json.Unmarshal(response.Body.Bytes(), &received))
	require.Equal(t, catalog.Version, received.Version)
	require.Len(t, received.Brands, len(catalog.Brands))

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
