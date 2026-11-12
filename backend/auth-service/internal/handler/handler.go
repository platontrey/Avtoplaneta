package handler

import (
	partsv1 "avtoplaneta/gen/parts/v1"
	"avtoplaneta/pkg/health"

	"auth-service/internal/config"
	"auth-service/internal/repository"
	"auth-service/internal/security"
	"auth-service/internal/service"
)

// Handler инкапсулирует зависимости для HTTP хэндлеров auth-service
type Handler struct {
	authService  service.AuthService
	sessionStore security.SessionStore
	csrfManager  security.CSRFManager
	userRepo     repository.UserRepository
	config       *config.Config
	partsClient  partsv1.PartsServiceClient
	health       *health.Checker
}

// NewHandler создает новый экземпляр Handler со всеми необходимыми зависимостями
func NewHandler(
	authService service.AuthService,
	sessionStore security.SessionStore,
	csrfManager security.CSRFManager,
	userRepo repository.UserRepository,
	cfg *config.Config,
) *Handler {
	return &Handler{
		authService:  authService,
		sessionStore: sessionStore,
		csrfManager:  csrfManager,
		userRepo:     userRepo,
		config:       cfg,
		health:       health.NewChecker("auth-service"),
	}
}

// HealthChecker возвращает чекер для регистрации readiness/liveness проверок
func (h *Handler) HealthChecker() *health.Checker {
	if h.health == nil {
		h.health = health.NewChecker("auth-service")
	}
	return h.health
}

// SetPartsClient устанавливает gRPC клиент parts-service
func (h *Handler) SetPartsClient(client partsv1.PartsServiceClient) {
	h.partsClient = client
}
