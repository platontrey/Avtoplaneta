package health

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

const (
	StatusHealthy   = "healthy"
	StatusUnhealthy = "unhealthy"

	DefaultTimeout = 3 * time.Second
)

// CheckFunc функция проверки отдельного компонента (БД, кэш, внешние API)
type CheckFunc func(ctx context.Context) error

// Pinger интерфейс компонентов с возможностью ping (например, *pgxpool.Pool)
type Pinger interface {
	Ping(ctx context.Context) error
}

// Status DTO ответа проверки здоровья
type Status struct {
	Status    string            `json:"status"`
	Service   string            `json:"service,omitempty"`
	Timestamp string            `json:"timestamp"`
	Checks    map[string]string `json:"checks,omitempty"`
}

// Checker агрегатор проверок liveness и readiness сервиса
type Checker struct {
	serviceName string
	timeout     time.Duration
	mu          sync.RWMutex
	readiness   map[string]CheckFunc
	liveness    map[string]CheckFunc
}

// Option функциональная опция для настройки Checker
type Option func(*Checker)

// WithTimeout задает предельное время на выполнение проверок
func WithTimeout(timeout time.Duration) Option {
	return func(c *Checker) {
		if timeout > 0 {
			c.timeout = timeout
		}
	}
}

// NewChecker инициализирует чекер для конкретного микросервиса
func NewChecker(serviceName string, opts ...Option) *Checker {
	c := &Checker{
		serviceName: serviceName,
		timeout:     DefaultTimeout,
		readiness:   make(map[string]CheckFunc),
		liveness:    make(map[string]CheckFunc),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// AddReadinessCheck регистрирует зависимость, необходимую для приема боевого трафика
func (c *Checker) AddReadinessCheck(name string, fn CheckFunc) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.readiness[name] = fn
}

// AddLivenessCheck регистрирует проверку жизнеспособности процесса
func (c *Checker) AddLivenessCheck(name string, fn CheckFunc) {
	if c == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.liveness[name] = fn
}

// AddPinger регистрирует Pinger (например, пул БД) в качестве readiness-проверки
func (c *Checker) AddPinger(name string, pinger Pinger) {
	if c == nil || pinger == nil {
		return
	}
	c.AddReadinessCheck(name, func(ctx context.Context) error {
		return pinger.Ping(ctx)
	})
}

// CheckReadiness выполняет все readiness-проверки параллельно
func (c *Checker) CheckReadiness(ctx context.Context) (Status, bool) {
	if c == nil {
		return Status{Status: StatusHealthy, Timestamp: time.Now().UTC().Format(time.RFC3339)}, true
	}

	c.mu.RLock()
	checks := make(map[string]CheckFunc, len(c.readiness))
	for k, v := range c.readiness {
		checks[k] = v
	}
	c.mu.RUnlock()

	return c.runChecks(ctx, checks)
}

// CheckLiveness выполняет все liveness-проверки параллельно
func (c *Checker) CheckLiveness(ctx context.Context) (Status, bool) {
	if c == nil {
		return Status{Status: StatusHealthy, Timestamp: time.Now().UTC().Format(time.RFC3339)}, true
	}

	c.mu.RLock()
	checks := make(map[string]CheckFunc, len(c.liveness))
	for k, v := range c.liveness {
		checks[k] = v
	}
	c.mu.RUnlock()

	return c.runChecks(ctx, checks)
}

func (c *Checker) runChecks(ctx context.Context, checks map[string]CheckFunc) (Status, bool) {
	timeoutCtx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	results := make(map[string]string, len(checks))
	var resultsMu sync.Mutex
	var wg sync.WaitGroup
	allHealthy := true

	for name, fn := range checks {
		wg.Add(1)
		go func(checkName string, checkFn CheckFunc) {
			defer wg.Done()

			errChan := make(chan error, 1)
			go func() {
				errChan <- checkFn(timeoutCtx)
			}()

			var checkErr error
			select {
			case <-timeoutCtx.Done():
				checkErr = timeoutCtx.Err()
			case err := <-errChan:
				checkErr = err
			}

			resultsMu.Lock()
			defer resultsMu.Unlock()
			if checkErr != nil {
				allHealthy = false
				results[checkName] = checkErr.Error()
			} else {
				results[checkName] = "ok"
			}
		}(name, fn)
	}

	wg.Wait()

	overallStatus := StatusHealthy
	if !allHealthy {
		overallStatus = StatusUnhealthy
	}

	serviceName := ""
	if c != nil {
		serviceName = c.serviceName
	}

	return Status{
		Status:    overallStatus,
		Service:   serviceName,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Checks:    results,
	}, allHealthy
}

// LiveHandler возвращает Gin-обработчик для проверки жизнеспособности (liveness)
func (c *Checker) LiveHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		status, healthy := c.CheckLiveness(ctx.Request.Context())
		code := http.StatusOK
		if !healthy {
			code = http.StatusServiceUnavailable
		}
		ctx.JSON(code, status)
	}
}

// ReadyHandler возвращает Gin-обработчик для проверки готовности (readiness)
func (c *Checker) ReadyHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		status, healthy := c.CheckReadiness(ctx.Request.Context())
		code := http.StatusOK
		if !healthy {
			code = http.StatusServiceUnavailable
		}
		ctx.JSON(code, status)
	}
}

// RegisterRoutes регистрирует стандартный набор health-эндпоинтов на роутере:
// - GET /health       -> readiness (обратная совместимость с Dockerfile/LoadBalancer)
// - GET /health/ready -> readiness
// - GET /health/live  -> liveness
// - GET /ready        -> readiness алиас
// - GET /live         -> liveness алиас
func (c *Checker) RegisterRoutes(r gin.IRoutes) {
	if c == nil {
		c = NewChecker("service")
	}
	readyHandler := c.ReadyHandler()
	liveHandler := c.LiveHandler()

	r.GET("/health", readyHandler)
	r.GET("/health/ready", readyHandler)
	r.GET("/ready", readyHandler)

	r.GET("/health/live", liveHandler)
	r.GET("/live", liveHandler)
}
