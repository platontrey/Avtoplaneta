package health

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type dummyPinger struct {
	err error
}

func (p *dummyPinger) Ping(ctx context.Context) error {
	return p.err
}

func TestChecker_AllHealthy(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checker := NewChecker("test-service")

	checker.AddPinger("database", &dummyPinger{err: nil})
	checker.AddReadinessCheck("cache", func(ctx context.Context) error {
		return nil
	})

	r := gin.New()
	checker.RegisterRoutes(r)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"healthy"`)
	assert.Contains(t, w.Body.String(), `"service":"test-service"`)
	assert.Contains(t, w.Body.String(), `"database":"ok"`)
	assert.Contains(t, w.Body.String(), `"cache":"ok"`)
}

func TestChecker_ReadinessFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checker := NewChecker("test-service")

	checker.AddPinger("database", &dummyPinger{err: errors.New("db connection refused")})
	checker.AddReadinessCheck("cache", func(ctx context.Context) error {
		return nil
	})

	r := gin.New()
	checker.RegisterRoutes(r)

	// Readiness check should return 503
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health/ready", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusServiceUnavailable, w.Code)
	assert.Contains(t, w.Body.String(), `"status":"unhealthy"`)
	assert.Contains(t, w.Body.String(), `"database":"db connection refused"`)
	assert.Contains(t, w.Body.String(), `"cache":"ok"`)

	// Liveness check should still return 200 OK
	wLive := httptest.NewRecorder()
	reqLive, _ := http.NewRequest(http.MethodGet, "/health/live", nil)
	r.ServeHTTP(wLive, reqLive)

	assert.Equal(t, http.StatusOK, wLive.Code)
	assert.Contains(t, wLive.Body.String(), `"status":"healthy"`)
}

func TestChecker_Timeout(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checker := NewChecker("slow-service", WithTimeout(50*time.Millisecond))

	checker.AddReadinessCheck("slow_service", func(ctx context.Context) error {
		select {
		case <-time.After(500 * time.Millisecond):
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	})

	status, healthy := checker.CheckReadiness(context.Background())
	require.False(t, healthy)
	assert.Equal(t, StatusUnhealthy, status.Status)
	assert.Contains(t, status.Checks["slow_service"], "context deadline exceeded")
}
