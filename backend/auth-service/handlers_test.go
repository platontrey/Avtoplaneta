package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ─── AuthService Mock ───────────────────────────────────────────────────────

type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) AuthenticateUser(email, password string) (*User, error) {
	args := m.Called(email, password)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockAuthService) CreateUserFromGoogle(email, name string) (*User, error) {
	args := m.Called(email, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockAuthService) GetCurrentUser(userID int64) (*User, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockAuthService) Logout(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) CreateUser(req CreateUserRequest) (*User, error) {
	args := m.Called(req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockAuthService) GetUsers() ([]User, error) {
	args := m.Called()
	return args.Get(0).([]User), args.Error(1)
}

func (m *MockAuthService) UpdateUser(userID int64, req UpdateUserRequest) (*User, error) {
	args := m.Called(userID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*User), args.Error(1)
}

func (m *MockAuthService) DeleteUser(userID int64) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockAuthService) GetTotalUsersCount() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockAuthService) LogUserActivity(user *User, action, resourceType string, resourceID *int64, details, ip, userAgent string) error {
	args := m.Called(user, action, resourceType, resourceID, details, ip, userAgent)
	return args.Error(0)
}

func (m *MockAuthService) GetUserActivityLogs(filters ActivityLogFilters) ([]UserActivityLog, error) {
	args := m.Called(filters)
	return args.Get(0).([]UserActivityLog), args.Error(1)
}

func (m *MockAuthService) GenerateCSRFToken(userID *int64) (string, error) {
	args := m.Called(userID)
	return args.String(0), args.Error(1)
}

// ─── Handlers Test Suite ────────────────────────────────────────────────────

type HandlersTestSuite struct {
	suite.Suite
	mockService  *MockAuthService
	mockUserRepo *MockUserRepository
	handler      *Handler
	router       *gin.Engine
}

func (suite *HandlersTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)
	suite.mockService = new(MockAuthService)
	suite.mockUserRepo = new(MockUserRepository)

	config := &Config{
		JWTSecret:       "test-secret-32-chars-minimum!!",
		PartsServiceURL: "http://localhost:8081",
	}

	maxAge := 1
	key := make([]byte, 32)
	_, _ = rand.Read(key)
	store = sessions.NewCookieStore(key)
	store.MaxAge(maxAge)

	SetUserRepo(suite.mockUserRepo)
	SetAuthConfig(config)

	suite.handler = NewHandler(suite.mockService, config)
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
	user := &User{ID: 1, Email: "test@example.com", Name: "Test", Role: "operator"}
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
	req := CreateUserRequest{
		Email:    "new@example.com",
		Name:     "New",
		Password: "pass123",
		Role:     "operator",
	}
	expected := &User{ID: 2, Email: "new@example.com", Name: "New", Role: "operator"}
	suite.mockService.On("CreateUser", req).Return(expected, nil)

	jsonData, _ := json.Marshal(req)
	c, w := newTestContext("POST", "/admin/users", jsonData)
	suite.handler.CreateUserHandler(c)

	assert.Equal(suite.T(), http.StatusCreated, w.Code)
}

func (suite *HandlersTestSuite) TestCreateUserHandler_Duplicate() {
	req := CreateUserRequest{Email: "exists@example.com", Name: "Exists", Password: "pass", Role: "operator"}
	suite.mockService.On("CreateUser", req).Return(nil, assert.AnError)

	jsonData, _ := json.Marshal(req)
	c, w := newTestContext("POST", "/admin/users", jsonData)
	suite.handler.CreateUserHandler(c)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func (suite *HandlersTestSuite) TestGetUsersHandler_Success() {
	users := []User{
		{ID: 1, Email: "user1@example.com", Name: "User 1"},
		{ID: 2, Email: "user2@example.com", Name: "User 2"},
	}
	suite.mockService.On("GetUsers").Return(users, nil)

	c, w := newTestContext("GET", "/admin/users", nil)
	suite.handler.GetUsersHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response map[string][]User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response["users"], 2)
}

func (suite *HandlersTestSuite) TestGetUsersHandler_Empty() {
	suite.mockService.On("GetUsers").Return([]User{}, nil)

	c, w := newTestContext("GET", "/admin/users", nil)
	suite.handler.GetUsersHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response map[string][]User
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Empty(suite.T(), response["users"])
}

func (suite *HandlersTestSuite) TestUpdateUserHandler_Success() {
	req := UpdateUserRequest{Name: "Updated", Role: "admin"}
	expected := &User{ID: 1, Name: "Updated", Email: "test@example.com", Role: "admin"}
	suite.mockService.On("UpdateUser", int64(1), req).Return(expected, nil)

	jsonData, _ := json.Marshal(req)
	c, w := newTestContext("PUT", "/admin/users/1", jsonData)
	c.Params = gin.Params{{Key: "id", Value: "1"}}
	suite.handler.UpdateUserHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response User
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
	req := UpdateUserRequest{Name: "Updated"}
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
	filters := ActivityLogFilters{
		Action: "login",
		Limit:  100,
	}
	logs := []UserActivityLog{
		{ID: 1, UserID: 1, Action: "login", ResourceType: "user", CreatedAt: time.Now()},
	}
	suite.mockService.On("GetUserActivityLogs", filters).Return(logs, nil)

	c, w := newTestContext("GET", "/admin/user-activity-logs?action=login", nil)
	c.Request = httptest.NewRequest("GET", "/admin/user-activity-logs?action=login", nil)
	suite.handler.GetUserActivityLogsHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response map[string][]UserActivityLog
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Len(suite.T(), response["logs"], 1)
	assert.Equal(suite.T(), "login", response["logs"][0].Action)
}

func (suite *HandlersTestSuite) TestGetUserActivityLogsHandler_WithUserID() {
	userID := int64(1)
	filters := ActivityLogFilters{
		UserID: &userID,
		Limit:  100,
	}
	logs := []UserActivityLog{
		{ID: 1, UserID: 1, Action: "create_part", ResourceType: "part", CreatedAt: time.Now()},
	}
	suite.mockService.On("GetUserActivityLogs", filters).Return(logs, nil)

	c, w := newTestContext("GET", "/admin/user-activity-logs?user_id=1", nil)
	c.Request = httptest.NewRequest("GET", "/admin/user-activity-logs?user_id=1", nil)
	suite.handler.GetUserActivityLogsHandler(c)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
	var response map[string][]UserActivityLog
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

	user := &User{ID: 10, Email: "test-google@example.com", Name: "Google User", Role: "operator"}
	suite.mockService.On("CreateUserFromGoogle", "test-google@example.com", "Google User").Return(user, nil)

	jsonData, _ := json.Marshal(map[string]string{"id_token": "test-id-token"})
	req, _ := http.NewRequest("POST", "/auth/google/mobile", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var resp LoginResponse
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

func TestHandlersTestSuite(t *testing.T) {
	suite.Run(t, new(HandlersTestSuite))
}

