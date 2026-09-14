package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupTestGatewayRouter() (*Gateway, *gin.Engine) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	g := &Gateway{
		router:    r,
		authCache: NewAuthCache(1 * time.Minute),
	}

	g.setupStaticRoutes()
	g.setupAppUpdateRoutes()
	g.setupClientLogRoutes()

	// Mock auth middleware для тестирования шлюза
	mockAuthMiddleware := func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		switch authHeader {
		case "Bearer admin-token":
			c.Set("user", User{ID: 1, Email: "admin@avtoplaneta.ru", Role: "admin"})
			c.Next()
		case "Bearer manager-token":
			c.Set("user", User{ID: 2, Email: "manager@avtoplaneta.ru", Role: "manager"})
			c.Next()
		case "Bearer operator-token":
			c.Set("user", User{ID: 3, Email: "operator@avtoplaneta.ru", Role: "operator"})
			c.Next()
		default:
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authentication required"})
			c.Abort()
		}
	}

	// Эндпоинты с проверкой ролей
	r.GET("/test/admin-only", mockAuthMiddleware, requireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "admin-ok"})
	})

	r.GET("/test/manager-min", mockAuthMiddleware, requireRole("manager"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "manager-ok"})
	})

	r.GET("/test/operator-min", mockAuthMiddleware, requireRole("operator"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "operator-ok"})
	})

	// Роуты админки
	r.GET("/admin/users", mockAuthMiddleware, requireRole("manager"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"users": []string{"admin", "manager", "operator"}})
	})
	r.POST("/admin/users", mockAuthMiddleware, requireRole("admin"), func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"status": "created"})
	})

	// Имитация auth эндпоинтов для проверки отсутствия конфликта методов в Gin
	r.POST("/auth/login", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "logged_in"})
	})

	// Имитация grpc-gateway wildcard
	r.Any("/api/v1/*any", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "grpc-gateway"})
	})

	return g, r
}

func TestGateway_RoleHierarchy(t *testing.T) {
	_, router := setupTestGatewayRouter()

	// 1. Admin имеет доступ ко ВСЕМ уровням (admin, manager, operator)
	for _, path := range []string{"/test/admin-only", "/test/manager-min", "/test/operator-min"} {
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer admin-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected Admin on %s to get 200, got %d", path, w.Code)
		}
	}

	// 2. Manager имеет доступ к manager и operator, но НЕ к admin
	reqMgrAdmin, _ := http.NewRequest(http.MethodGet, "/test/admin-only", nil)
	reqMgrAdmin.Header.Set("Authorization", "Bearer manager-token")
	wMgrAdmin := httptest.NewRecorder()
	router.ServeHTTP(wMgrAdmin, reqMgrAdmin)
	if wMgrAdmin.Code != http.StatusForbidden {
		t.Errorf("expected Manager on /test/admin-only to get 403, got %d", wMgrAdmin.Code)
	}

	for _, path := range []string{"/test/manager-min", "/test/operator-min"} {
		req, _ := http.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer manager-token")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
		if w.Code != http.StatusOK {
			t.Errorf("expected Manager on %s to get 200, got %d", path, w.Code)
		}
	}

	// 3. Operator имеет доступ только к operator
	reqOp, _ := http.NewRequest(http.MethodGet, "/test/operator-min", nil)
	reqOp.Header.Set("Authorization", "Bearer operator-token")
	wOp := httptest.NewRecorder()
	router.ServeHTTP(wOp, reqOp)
	if wOp.Code != http.StatusOK {
		t.Errorf("expected Operator on /test/operator-min to get 200, got %d", wOp.Code)
	}

	reqOpBlocked, _ := http.NewRequest(http.MethodGet, "/test/manager-min", nil)
	reqOpBlocked.Header.Set("Authorization", "Bearer operator-token")
	wOpBlocked := httptest.NewRecorder()
	router.ServeHTTP(wOpBlocked, reqOpBlocked)
	if wOpBlocked.Code != http.StatusForbidden {
		t.Errorf("expected Operator on /test/manager-min to get 403, got %d", wOpBlocked.Code)
	}
}

func TestGateway_NoRouteConflict_LoginAndAppVersion(t *testing.T) {
	_, router := setupTestGatewayRouter()

	// Проверяем, что POST /auth/login не падает с 405 Not Allowed из-за /api/v1/*any или /api/app/*
	loginReq, _ := http.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString(`{}`))
	loginReq.Header.Set("Content-Type", "application/json")
	wLogin := httptest.NewRecorder()
	router.ServeHTTP(wLogin, loginReq)

	if wLogin.Code != http.StatusOK {
		t.Errorf("expected POST /auth/login to be 200, got %d. Body: %s", wLogin.Code, wLogin.Body.String())
	}
	if !strings.Contains(wLogin.Body.String(), "logged_in") {
		t.Errorf("expected 'logged_in' in body, got: %s", wLogin.Body.String())
	}

	// Проверяем, что GET /api/app/version работает
	verReq, _ := http.NewRequest(http.MethodGet, "/api/app/version", nil)
	wVer := httptest.NewRecorder()
	router.ServeHTTP(wVer, verReq)
	if wVer.Code != http.StatusOK {
		t.Errorf("expected GET /api/app/version to be 200, got %d", wVer.Code)
	}

	// Проверяем, что GET /admin/users доступен для менеджера
	adminGetReq, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	adminGetReq.Header.Set("Authorization", "Bearer manager-token")
	wAdminGet := httptest.NewRecorder()
	router.ServeHTTP(wAdminGet, adminGetReq)
	if wAdminGet.Code != http.StatusOK {
		t.Errorf("expected GET /admin/users for manager to be 200, got %d", wAdminGet.Code)
	}

	// Проверяем, что POST /admin/users требует админа
	adminPostReq, _ := http.NewRequest(http.MethodPost, "/admin/users", nil)
	adminPostReq.Header.Set("Authorization", "Bearer manager-token")
	wAdminPost := httptest.NewRecorder()
	router.ServeHTTP(wAdminPost, adminPostReq)
	if wAdminPost.Code != http.StatusForbidden {
		t.Errorf("expected POST /admin/users for manager to be 403, got %d", wAdminPost.Code)
	}
}

func TestGateway_ClientErrorLogsApi(t *testing.T) {
	_, router := setupTestGatewayRouter()

	// 1. Отправка логов мобильного клиента (POST /api/app/logs)
	payload := ClientLogsBatchRequest{
		Client:  "AvtoplanetaApp",
		Version: "1.0.1",
		Build:   2,
		Device:  "Pixel 7 Android 14",
		Logs: []ClientErrorLog{
			{
				ID:         "test-err-1",
				Timestamp:  time.Now().Format(time.RFC3339),
				Type:       "network",
				Message:    "401 Unauthorized at /admin/users",
				StatusCode: 401,
				Endpoint:   "/admin/users",
				Method:     "GET",
			},
		},
	}
	bodyBytes, _ := json.Marshal(payload)

	postReq, _ := http.NewRequest(http.MethodPost, "/api/app/logs", bytes.NewReader(bodyBytes))
	postReq.Header.Set("Content-Type", "application/json")
	wPost := httptest.NewRecorder()
	router.ServeHTTP(wPost, postReq)

	if wPost.Code != http.StatusOK {
		t.Errorf("expected POST /api/app/logs to be 200, got %d. Body: %s", wPost.Code, wPost.Body.String())
	}
	if !strings.Contains(wPost.Body.String(), `"received":1`) {
		t.Errorf("expected received:1 in body, got: %s", wPost.Body.String())
	}

	// 2. Проверяем, что создался файл логов
	today := time.Now().Format("2006-01-02")
	logFile := "./logs/client_errors_" + today + ".log"
	if _, err := os.Stat(logFile); os.IsNotExist(err) {
		t.Errorf("expected log file %s to exist", logFile)
	}
	_ = os.Remove(logFile)
}
