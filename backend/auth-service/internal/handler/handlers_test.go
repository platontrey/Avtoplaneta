package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"auth-service/internal/security"
	"auth-service/internal/service"
)

type HandlersTestSuite struct {
	suite.Suite
	mockService  *service.MockAuthService
	mockUserRepo *repository.MockUserRepository
	handler      *Handler
	router       *gin.Engine
}

func (suite *HandlersTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.mockService = new(service.MockAuthService)
	suite.mockUserRepo = new(repository.MockUserRepository)

	cfg := &config.Config{
		JWTSecret:       "test-secret-32-chars-minimum!!",
		PartsServiceURL: "http://localhost:8081",
	}

	sessionStore, _ := security.NewCookieSessionStore("test-session-secret-32-bytes-long!!")
	csrfManager := security.NewCSRFManager()

	suite.handler = NewHandler(suite.mockService, sessionStore, csrfManager, suite.mockUserRepo, cfg)
	suite.router = gin.New()
	SetupRoutes(suite.router, suite.handler)
}

func (suite *HandlersTestSuite) TearDownTest() {
	suite.mockService.AssertExpectations(suite.T())
	suite.mockUserRepo.AssertExpectations(suite.T())
}

// ─── Public endpoint tests (through router) ─────────────────────────────────

func (suite *HandlersTestSuite) TestHealthEndpoint() {
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *HandlersTestSuite) TestUserLoginHandler_Success() {
	user := &domain.User{ID: 1, Email: "test@example.com", Name: "Test", Role: "operator"}
	suite.mockService.On("AuthenticateUser", "test@example.com", "password123").Return(user, nil)

	jsonData, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "password123"})
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *HandlersTestSuite) TestUserLoginHandler_Unauthorized() {
	suite.mockService.On("AuthenticateUser", "test@example.com", "wrong").Return(nil, assert.AnError)
	jsonData, _ := json.Marshal(map[string]string{"email": "test@example.com", "password": "wrong"})
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *HandlersTestSuite) TestUserLoginHandler_EmptyFields() {
	req, _ := http.NewRequest("POST", "/auth/login", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *HandlersTestSuite) TestLogoutHandler() {
	req, _ := http.NewRequest("POST", "/auth/logout", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *HandlersTestSuite) TestGetCurrentUserHandler_NoAuth() {
	req, _ := http.NewRequest("GET", "/auth/me", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *HandlersTestSuite) TestRefreshTokenHandler_EmptyToken() {
	req, _ := http.NewRequest("POST", "/auth/refresh", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *HandlersTestSuite) TestInternalLogUserActivityHandler_NoUserID() {
	req, _ := http.NewRequest("POST", "/internal/log-activity", bytes.NewBufferString(`{"action":"test"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

// ─── Admin handler tests (called directly, middleware bypassed) ──────────────

func newTestContext(method, path string, body []byte) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	req := httptest.NewRequest(method, path, nil)
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	}
	c.Request = req
	return c, w
}

func (suite *HandlersTestSuite) TestCreateUserHandler_Success() {
	req := domain.CreateUserRequest{
		Email:    "new@example.com",
		Name:     "New",
		Password: "pass123",
		Role:     "operator",
	}
	expected := &domain.User{ID: 2, Email: "new@example.com", Name: "New", Role: "operator"}
	suite.mockService.On("CreateUser", req).Return(expected, nil)

	jsonData, _ := json.Marshal(req)
	c, w := newTestContext("POST", "/admin/users", jsonData)
	suite.handler.CreateUserHandler(c)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

func (suite *HandlersTestSuite) TestCreateUserHandler_Duplicate() {
	req := domain.CreateUserRequest{Email: "exists@example.com", Name: "Exists", Password: "pass", Role: "operator"}
	suite.mockService.On("CreateUser", req).Return(nil, assert.AnError)

	jsonData, _ := json.Marshal(req)
	c, w := newTestContext("POST", "/admin/users", jsonData)
	suite.handler.CreateUserHandler(c)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *HandlersTestSuite) TestGetUsersHandler_Success() {
	users := []domain.User{
		{ID: 1, Email: "user1@example.com", Name: "User 1"},
		{ID: 2, Email: "user2@example.com", Name: "User 2"},
	}
	suite.mockService.On("GetUsers").Return(users, nil)

	c, w := newTestContext("GET", "/admin/users", nil)
	suite.handler.GetUsersHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response map[string][]domain.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response["users"], 2)
}

func (suite *HandlersTestSuite) TestGetUsersHandler_Empty() {
	suite.mockService.On("GetUsers").Return([]domain.User{}, nil)

	c, w := newTestContext("GET", "/admin/users", nil)
	suite.handler.GetUsersHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response map[string][]domain.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), response["users"])
}

func (suite *HandlersTestSuite) TestUpdateUserHandler_Success() {
	req := domain.UpdateUserRequest{Name: "Updated", Role: "admin"}
	expected := &domain.User{ID: 1, Name: "Updated", Email: "test@example.com", Role: "admin"}
	suite.mockService.On("UpdateUser", int64(1), req).Return(expected, nil)

	jsonData, _ := json.Marshal(req)
	c, w := newTestContext("PUT", "/admin/users/1", jsonData)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	suite.handler.UpdateUserHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response domain.User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Updated", response.Name)
}

func (suite *HandlersTestSuite) TestUpdateUserHandler_InvalidID() {
	c, w := newTestContext("PUT", "/admin/users/abc", []byte(`{"name":"test"}`))
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	suite.handler.UpdateUserHandler(c)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *HandlersTestSuite) TestUpdateUserHandler_NotFound() {
	req := domain.UpdateUserRequest{Name: "Updated"}
	suite.mockService.On("UpdateUser", int64(999), req).Return(nil, assert.AnError)

	jsonData, _ := json.Marshal(req)
	c, w := newTestContext("PUT", "/admin/users/999", jsonData)
	c.Params = gin.Params{{Key: "id", Value: "999"}}
	suite.handler.UpdateUserHandler(c)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *HandlersTestSuite) TestDeleteUserHandler_Success() {
	suite.mockService.On("DeleteUser", int64(1)).Return(nil)

	c, w := newTestContext("DELETE", "/admin/users/1", nil)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	suite.handler.DeleteUserHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *HandlersTestSuite) TestDeleteUserHandler_InvalidID() {
	c, w := newTestContext("DELETE", "/admin/users/abc", nil)
	c.Params = gin.Params{{Key: "id", Value: "abc"}}
	suite.handler.DeleteUserHandler(c)
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *HandlersTestSuite) TestGetServerStatusHandler() {
	suite.mockService.On("GetTotalUsersCount").Return(5, nil)

	c, w := newTestContext("GET", "/admin/status", nil)
	suite.handler.GetServerStatusHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *HandlersTestSuite) TestGetServerLogsHandler() {
	c, w := newTestContext("GET", "/admin/logs", nil)
	suite.handler.GetServerLogsHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response map[string][]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotEmpty(suite.T(), response["logs"])
}

func (suite *HandlersTestSuite) TestGetUserActivityLogsHandler() {
	filters := domain.ActivityLogFilters{
		Action: "login",
		Limit:  100,
	}
	logs := []domain.UserActivityLog{
		{ID: 1, UserID: 1, Action: "login", ResourceType: "user", CreatedAt: time.Now()},
	}
	suite.mockService.On("GetUserActivityLogs", filters).Return(logs, nil)

	c, w := newTestContext("GET", "/admin/user-activity-logs?action=login", nil)
	c.Request = httptest.NewRequest("GET", "/admin/user-activity-logs?action=login", nil)
	suite.handler.GetUserActivityLogsHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response map[string][]domain.UserActivityLog
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response["logs"], 1)
	assert.Equal(suite.T(), "login", response["logs"][0].Action)
}

func (suite *HandlersTestSuite) TestGetUserActivityLogsHandler_WithUserID() {
	userID := int64(1)
	filters := domain.ActivityLogFilters{
		UserID: &userID,
		Limit:  100,
	}
	logs := []domain.UserActivityLog{
		{ID: 1, UserID: 1, Action: "create_part", ResourceType: "part", CreatedAt: time.Now()},
	}
	suite.mockService.On("GetUserActivityLogs", filters).Return(logs, nil)

	c, w := newTestContext("GET", "/admin/user-activity-logs?user_id=1", nil)
	c.Request = httptest.NewRequest("GET", "/admin/user-activity-logs?user_id=1", nil)
	suite.handler.GetUserActivityLogsHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response map[string][]domain.UserActivityLog
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response["logs"], 1)
}

func (suite *HandlersTestSuite) TestGoogleMobileAuthHandler_Success() {
	mockGoogleServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(suite.T(), "GET", r.Method)
		assert.Equal(suite.T(), "test-id-token", r.URL.Query().Get("id_token"))

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"email": "test-google@example.com",
			"email_verified": "true",
			"name": "Google User"
		}`))
	}))
	defer mockGoogleServer.Close()

	oldURL := googleTokenInfoURL
	googleTokenInfoURL = mockGoogleServer.URL
	defer func() { googleTokenInfoURL = oldURL }()

	user := &domain.User{ID: 10, Email: "test-google@example.com", Name: "Google User", Role: "operator"}
	suite.mockService.On("CreateUserFromGoogle", "test-google@example.com", "Google User").Return(user, nil)

	jsonData, _ := json.Marshal(map[string]string{"id_token": "test-id-token"})
	req, _ := http.NewRequest("POST", "/auth/google/mobile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var resp domain.LoginResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "test-google@example.com", resp.User.Email)
	assert.NotEmpty(suite.T(), resp.Token)
	assert.NotEmpty(suite.T(), resp.RefreshToken)
}

func (suite *HandlersTestSuite) TestGoogleMobileAuthHandler_InvalidToken() {
	mockGoogleServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"error": "invalid_token"}`))
	}))
	defer mockGoogleServer.Close()

	oldURL := googleTokenInfoURL
	googleTokenInfoURL = mockGoogleServer.URL
	defer func() { googleTokenInfoURL = oldURL }()

	jsonData, _ := json.Marshal(map[string]string{"id_token": "invalid-token"})
	req, _ := http.NewRequest("POST", "/auth/google/mobile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func (suite *HandlersTestSuite) TestVerifyAuthHandler_Success() {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/auth/verify", nil)
	c.Set("user", domain.User{
		ID:    42,
		Email: "verify@example.com",
		Name:  "Verify User",
		Role:  "operator",
	})

	suite.handler.VerifyAuthHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	assert.Equal(suite.T(), "42", w.Header().Get("X-User-ID"))
	assert.Equal(suite.T(), "verify@example.com", w.Header().Get("X-User-Email"))
	assert.Equal(suite.T(), "Verify User", w.Header().Get("X-User-Name"))
	assert.Equal(suite.T(), "operator", w.Header().Get("X-User-Role"))
}

func (suite *HandlersTestSuite) TestVerifyAuthHandler_Unauthorized() {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/auth/verify", nil)

	suite.handler.VerifyAuthHandler(c)

	assert.Equal(suite.T(), http.StatusUnauthorized, w.Code)
}

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}
