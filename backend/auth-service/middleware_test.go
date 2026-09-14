package main

import (
	"crypto/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupTestMiddlewareRouter(userRepo UserRepository, cfg *Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	key := make([]byte, 32)
	_, _ = rand.Read(key)
	store = sessions.NewCookieStore(key)

	SetUserRepo(userRepo)
	SetAuthConfig(cfg)

	r.Use(csrfMiddleware)

	admin := r.Group("/admin")
	admin.Use(authMiddleware)
	{
		admin.GET("/users", requireMinRole("manager"), func(c *gin.Context) {
			u, _ := c.Get("user")
			c.JSON(http.StatusOK, gin.H{"status": "ok", "user": u})
		})

		admin.POST("/users", requireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"status": "created"})
		})

		admin.DELETE("/users/:id", requireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "deleted"})
		})
	}

	return r
}

func TestAuthMiddleware_JWTBearer(t *testing.T) {
	cfg := &Config{
		JWTSecret: "test-jwt-secret-key-32byteslong!",
	}
	mockRepo := new(MockUserRepository)

	adminUser := &User{
		ID:    1,
		Email: "admin@avtoplaneta.ru",
		Name:  "Admin",
		Role:  "admin",
	}

	mockRepo.On("FindByID", int64(1)).Return(adminUser, nil)

	accessToken, _, err := GenerateJWTTokens(adminUser, cfg)
	assert.NoError(t, err)

	router := setupTestMiddlewareRouter(mockRepo, cfg)

	// 1. Успешный GET /admin/users с Bearer токеном
	req, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "admin@avtoplaneta.ru")

	// 2. Успешный POST /admin/users с Bearer токеном (проверка, что CSRF не блокирует Bearer)
	postReq, _ := http.NewRequest(http.MethodPost, "/admin/users", nil)
	postReq.Header.Set("Authorization", "Bearer "+accessToken)
	wPost := httptest.NewRecorder()
	router.ServeHTTP(wPost, postReq)

	assert.Equal(t, http.StatusCreated, wPost.Code)

	// 3. Успешный DELETE /admin/users/42 с Bearer токеном
	delReq, _ := http.NewRequest(http.MethodDelete, "/admin/users/42", nil)
	delReq.Header.Set("Authorization", "Bearer "+accessToken)
	wDel := httptest.NewRecorder()
	router.ServeHTTP(wDel, delReq)

	assert.Equal(t, http.StatusOK, wDel.Code)

	// 4. Ошибка 401 при недействительном токене
	badReq, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	badReq.Header.Set("Authorization", "Bearer invalid.jwt.token")
	wBad := httptest.NewRecorder()
	router.ServeHTTP(wBad, badReq)

	assert.Equal(t, http.StatusUnauthorized, wBad.Code)
	assert.Contains(t, wBad.Body.String(), "Недействительный токен")

	// 5. Ошибка 401 при отсутствии авторизации
	noAuthReq, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	wNoAuth := httptest.NewRecorder()
	router.ServeHTTP(wNoAuth, noAuthReq)

	assert.Equal(t, http.StatusUnauthorized, wNoAuth.Code)
}

func TestAuthMiddleware_RoleHierarchy(t *testing.T) {
	cfg := &Config{
		JWTSecret: "test-jwt-secret-key-32byteslong!",
	}
	mockRepo := new(MockUserRepository)

	managerUser := &User{
		ID:    2,
		Email: "manager@avtoplaneta.ru",
		Name:  "Manager",
		Role:  "manager",
	}

	operatorUser := &User{
		ID:    3,
		Email: "operator@avtoplaneta.ru",
		Name:  "Operator",
		Role:  "operator",
	}

	mockRepo.On("FindByID", int64(2)).Return(managerUser, nil)
	mockRepo.On("FindByID", int64(3)).Return(operatorUser, nil)

	managerToken, _, _ := GenerateJWTTokens(managerUser, cfg)
	operatorToken, _, _ := GenerateJWTTokens(operatorUser, cfg)

	router := setupTestMiddlewareRouter(mockRepo, cfg)

	// Менеджер МОЖЕТ смотреть список пользователей (GET /admin/users требует minRole "manager")
	req, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	req.Header.Set("Authorization", "Bearer "+managerToken)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Менеджер НЕ МОЖЕТ создавать пользователей (POST /admin/users требует роль "admin")
	postReq, _ := http.NewRequest(http.MethodPost, "/admin/users", nil)
	postReq.Header.Set("Authorization", "Bearer "+managerToken)
	wPost := httptest.NewRecorder()
	router.ServeHTTP(wPost, postReq)
	assert.Equal(t, http.StatusForbidden, wPost.Code)

	// Оператор НЕ МОЖЕТ смотреть список пользователей (GET /admin/users требует minRole "manager")
	opReq, _ := http.NewRequest(http.MethodGet, "/admin/users", nil)
	opReq.Header.Set("Authorization", "Bearer "+operatorToken)
	wOp := httptest.NewRecorder()
	router.ServeHTTP(wOp, opReq)
	assert.Equal(t, http.StatusForbidden, wOp.Code)
}

func TestCSRFMiddleware_BlocksCookiePostWithoutToken(t *testing.T) {
	cfg := &Config{JWTSecret: "test-secret"}
	mockRepo := new(MockUserRepository)
	router := setupTestMiddlewareRouter(mockRepo, cfg)

	// POST запрос без Bearer и без CSRF токена должен отклоняться
	req, _ := http.NewRequest(http.MethodPost, "/admin/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Требуется токен CSRF")
}

// Заглушка, чтобы избежать неиспользуемого импорта
var _ = mock.Anything
