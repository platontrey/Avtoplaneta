package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func setupTestUpdateRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	g := &Gateway{router: r}
	g.setupAppUpdateRoutes()
	return r
}

func TestGetAppVersionHandler(t *testing.T) {
	router := setupTestUpdateRouter()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/app/version", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", w.Code, w.Body.String())
	}

	contentType := w.Header().Get("Content-Type")
	if !strings.Contains(contentType, "application/json") {
		t.Errorf("expected Content-Type application/json, got %s", contentType)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to parse JSON response: %v", err)
	}

	if response["version"] == nil || response["version"] == "" {
		t.Errorf("expected non-empty version in response")
	}
	if response["build_number"] == nil {
		t.Errorf("expected build_number in response")
	}
	if response["download_url"] == nil || response["download_url"] == "" {
		t.Errorf("expected download_url in response")
	}
}

func TestDownloadAppHandler_NotFound(t *testing.T) {
	router := setupTestUpdateRouter()

	req, err := http.NewRequest(http.MethodGet, "/api/v1/app/download", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	_ = os.Remove("downloads/avtoplaneta-release.apk")
	_ = os.Remove("downloads/app-release.apk")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
}

func TestDownloadAppHandler_SuccessWithAPK(t *testing.T) {
	router := setupTestUpdateRouter()

	_ = os.MkdirAll("downloads", 0755)
	testApkPath := filepath.Join("downloads", "avtoplaneta-release.apk")
	testData := []byte("PK\x03\x04fake-apk-binary-content-for-testing")
	if err := os.WriteFile(testApkPath, testData, 0644); err != nil {
		t.Fatalf("failed to create test apk: %v", err)
	}
	defer os.Remove(testApkPath)

	req, err := http.NewRequest(http.MethodGet, "/api/v1/app/download", nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	expectedContentType := "application/vnd.android.package-archive"
	if w.Header().Get("Content-Type") != expectedContentType {
		t.Errorf("expected Content-Type %s, got %s", expectedContentType, w.Header().Get("Content-Type"))
	}

	disposition := w.Header().Get("Content-Disposition")
	if !strings.Contains(disposition, "avtoplaneta-release.apk") {
		t.Errorf("expected Content-Disposition to contain avtoplaneta-release.apk, got %s", disposition)
	}

	if !bytes.Equal(w.Body.Bytes(), testData) {
		t.Errorf("expected body to match test data")
	}
}
