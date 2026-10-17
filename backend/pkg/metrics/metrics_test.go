package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestGinMiddleware_RecordsMetrics(t *testing.T) {
	router := gin.New()
	router.Use(GinMiddleware("test-service"))

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.POST("/items/:id", func(c *gin.Context) {
		c.JSON(http.StatusCreated, gin.H{"id": c.Param("id")})
	})

	router.GET("/error", func(c *gin.Context) {
		c.AbortWithStatus(http.StatusInternalServerError)
	})

	router.GET("/metrics", Handler())

	// Request 1: GET /health
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	// Request 2: POST /items/123 -> matches parameterized route /items/:id
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodPost, "/items/123", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)

	// Request 3: GET /error -> triggers error counter
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/error", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusInternalServerError, w.Code)

	// Request 4: GET /metrics -> ensure prometheus output includes the recorded metrics
	w = httptest.NewRecorder()
	req, _ = http.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, "http_requests_total")
	assert.Contains(t, body, `service="test-service"`)
	assert.Contains(t, body, `path="/health"`)
	assert.Contains(t, body, `path="/items/:id"`)
	assert.Contains(t, body, "http_requests_errors_total")
	assert.Contains(t, body, `path="/error"`)
	assert.Contains(t, body, "http_request_duration_seconds_bucket")
}

func TestOutboxMetrics_RecordEvents(t *testing.T) {
	outbox := NewOutboxMetrics("test-parts-service")
	require.NotNil(t, outbox)

	outbox.IncProcessed("part_created")
	outbox.IncFailed("part_updated")
	outbox.ObserveDuration("part_created", 15*time.Millisecond)

	// Verify metrics endpoint serves them
	router := gin.New()
	router.GET("/metrics", Handler())

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/metrics", nil)
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusOK, w.Code)

	body := w.Body.String()
	assert.Contains(t, body, "outbox_events_processed_total")
	assert.Contains(t, body, `service="test-parts-service"`)
	assert.Contains(t, body, `event_type="part_created"`)
	assert.Contains(t, body, "outbox_events_failed_total")
	assert.Contains(t, body, `event_type="part_updated"`)
	assert.Contains(t, body, "outbox_processing_duration_seconds")
}
