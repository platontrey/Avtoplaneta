package ginmw

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCORS_Preflight(t *testing.T) {
	r := gin.New()
	r.Use(CORS("http://example.com"))
	r.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodOptions, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "http://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestRequireAuth_Unauthorized(t *testing.T) {
	r := gin.New()
	r.Use(RequireAuth())
	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "Authentication required")
}

func TestRequireAuth_Success(t *testing.T) {
	r := gin.New()
	r.Use(RequireAuth())
	r.GET("/protected", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user_id": c.GetInt64("user_id")})
	})

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("X-User-ID", "42")
	req.Header.Set("X-User-Name", "Ivan")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"user_id":42`)
}

func TestRequireRole_Forbidden(t *testing.T) {
	r := gin.New()
	r.Use(RequireRole("admin"))
	r.GET("/admin-only", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/admin-only", nil)
	req.Header.Set("X-User-ID", "42")
	req.Header.Set("X-User-Role", "operator")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestRequireRole_Allowed(t *testing.T) {
	r := gin.New()
	r.Use(RequireRole("manager", "admin"))
	r.GET("/staff-only", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/staff-only", nil)
	req.Header.Set("X-User-ID", "42")
	req.Header.Set("X-User-Role", "Manager")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
