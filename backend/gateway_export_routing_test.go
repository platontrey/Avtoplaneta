package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// Выгрузка живёт в отдельном сервисе, но публичные адреса остались прежними:
// /uploads/pricelist.xml как был, так и есть. Развести их отдельными маршрутами
// нельзя — gin не даёт зарегистрировать конкретный путь рядом с catch-all, —
// поэтому выбор делается кодом, а код нужно проверять.
//
// Отдельно это важно потому, что уже было: /api/vehicle-catalog работал в
// сервисе, но не был прописан в шлюзе, и снаружи отдавал 404. Тест ловит ровно
// такую ошибку.
func setupUploadsRoutingGateway(t *testing.T) (*gin.Engine, func() string, func() string) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	var partsHit, exportHit string
	partsService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		partsHit = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("parts"))
	}))
	t.Cleanup(partsService.Close)

	exportService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		exportHit = r.URL.Path
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("export"))
	}))
	t.Cleanup(exportService.Close)

	gateway := &Gateway{
		router:           gin.New(),
		authCache:        NewAuthCache(time.Minute),
		partsServiceURL:  partsService.URL,
		exportServiceURL: exportService.URL,
	}
	gateway.initResiliencyPatterns()
	gateway.router.Any("/uploads/*filepath", gateway.proxyToUploads)

	return gateway.router, func() string { return partsHit }, func() string { return exportHit }
}

func TestUploadsPriceListGoesToExportService(t *testing.T) {
	router, partsHit, exportHit := setupUploadsRoutingGateway(t)

	request, _ := http.NewRequest(http.MethodGet, "/uploads/pricelist.xml", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("ожидался 200, получен %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := exportHit(); got != "/uploads/pricelist.xml" {
		t.Fatalf("прайс-лист не дошёл до export-service, получено %q", got)
	}
	if got := partsHit(); got != "" {
		t.Fatalf("прайс-лист ушёл в parts-service: %q", got)
	}
}

func TestUploadsPhotosStayInPartsService(t *testing.T) {
	router, partsHit, exportHit := setupUploadsRoutingGateway(t)

	request, _ := http.NewRequest(http.MethodGet, "/uploads/12_1700000000.jpg", nil)
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("ожидался 200, получен %d: %s", recorder.Code, recorder.Body.String())
	}
	if got := partsHit(); got != "/uploads/12_1700000000.jpg" {
		t.Fatalf("фотография не дошла до parts-service, получено %q", got)
	}
	if got := exportHit(); got != "" {
		t.Fatalf("фотография ушла в export-service: %q", got)
	}
}

// Пересборка прайс-листа — это полный проход по складу и перезапись публичного
// файла, а отправка на Drom вдобавок уходит наружу от имени компании.
// Анонимно дёргать их нельзя.
func TestExportRoutesRequireAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	reached := false
	exportService := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		reached = true
		w.WriteHeader(http.StatusOK)
	}))
	defer exportService.Close()

	gateway := &Gateway{
		router:           gin.New(),
		authCache:        NewAuthCache(time.Minute),
		exportServiceURL: exportService.URL,
	}
	gateway.initResiliencyPatterns()
	gateway.setupExportRoutes()

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/export/xml"},
		{http.MethodPost, "/api/export/drom"},
	}

	for _, testCase := range cases {
		request, _ := http.NewRequest(testCase.method, testCase.path, nil)
		recorder := httptest.NewRecorder()
		gateway.router.ServeHTTP(recorder, request)

		// 404 значит, что маршрут забыли прописать в шлюзе, и снаружи выгрузки
		// просто нет — так уже было с /api/vehicle-catalog.
		if recorder.Code == http.StatusNotFound {
			t.Errorf("%s %s не зарегистрирован в шлюзе", testCase.method, testCase.path)
			continue
		}
		// Конкретный код зависит от того, ответил ли auth-service, а вот то,
		// что запрос без токена до сервиса не доходит, зависеть ни от чего
		// не должно.
		if recorder.Code < 400 {
			t.Errorf("%s %s без токена вернул %d", testCase.method, testCase.path, recorder.Code)
		}
		if reached {
			t.Fatalf("%s %s без токена дошёл до export-service", testCase.method, testCase.path)
		}
	}
}
