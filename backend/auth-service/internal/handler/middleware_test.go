package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"auth-service/internal/config"
	"auth-service/internal/domain"
	"auth-service/internal/repository"
	"auth-service/internal/security"
)

func setupTestMiddlewareRouter(userRepo repository.UserRepository, cfg *config.Config) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	sessionStore, _ := security.NewCookieSessionStore("test-jwt-secret-key-32byteslong!")
	csrfManager := security.NewCSRFManager()
	h := NewHandler(nil, sessionStore, csrfManager, userRepo, cfg)

	r.Use(h.CSRFMiddleware)

	admin := r.Group("/admin")
	admin.Use(h.AuthMiddleware)
	{
		admin.GET("/users", h.RequireMinRole("manager"), func(c *gin.Context) {
			u, _ := c.Get("user")
			c.JSON(http.StatusOK, gin.H{"status": "ok", "user": u})
		})

		admin.POST("/users", h.RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusCreated, gin.H{"status": "created"})
		})

		admin.DELETE("/users/:id", h.RequireRole("admin"), func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "deleted"})
		})
	}

	return r
}

func TestAuthMiddleware_JWTBearer(t *testing.T) {
	cfg := &config.Config{
		JWTSecret: "test-jwt-secret-key-32byteslong!",
	}
	mockRepo := new(repository.MockUserRepository)

	adminUser := &domain.User{
		ID:    1,
		Email: "admin@avtoplaneta.ru",
		Name:  "Admin",
		Role:  "admin",
	}

	mockRepo.On("FindByID", int64(1)).Return(adminUser, nil)

	accessToken, _, err := security.GenerateJWTTokens(adminUser, cfg.JWTSecret)
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
	cfg := &config.Config{
		JWTSecret: "test-jwt-secret-key-32byteslong!",
	}
	mockRepo := new(repository.MockUserRepository)

	managerUser := &domain.User{
		ID:    2,
		Email: "manager@avtoplaneta.ru",
		Name:  "Manager",
		Role:  "manager",
	}

	operatorUser := &domain.User{
		ID:    3,
		Email: "operator@avtoplaneta.ru",
		Name:  "Operator",
		Role:  "operator",
	}

	mockRepo.On("FindByID", int64(2)).Return(managerUser, nil)
	mockRepo.On("FindByID", int64(3)).Return(operatorUser, nil)

	managerToken, _, _ := security.GenerateJWTTokens(managerUser, cfg.JWTSecret)
	operatorToken, _, _ := security.GenerateJWTTokens(operatorUser, cfg.JWTSecret)

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
	cfg := &config.Config{JWTSecret: "test-secret-at-least-32-bytes-long!"}
	mockRepo := new(repository.MockUserRepository)
	router := setupTestMiddlewareRouter(mockRepo, cfg)

	// POST запрос без Bearer и без CSRF токена должен отклоняться
	req, _ := http.NewRequest(http.MethodPost, "/admin/users", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "Требуется токен CSRF")
}
