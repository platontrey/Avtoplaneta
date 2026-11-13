package handler

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/gin-gonic/gin"

	"avtoplaneta/pkg/authcontext"
	"avtoplaneta/pkg/health"
	"avtoplaneta/pkg/logger"
	"orders-service/internal/domain"
	"orders-service/internal/service"
)

// Handler содержит все HTTP handlers для orders-service
type Handler struct {
	ordersService service.OrdersService
	publisher     domain.EventPublisher
	health        *health.Checker
}

// NewHandler создает новый handler с dependency injection
func NewHandler(ordersService service.OrdersService, publisher domain.EventPublisher) *Handler {
	return &Handler{
		ordersService: ordersService,
		publisher:     publisher,
		health:        health.NewChecker("orders-service"),
	}
}

// HealthChecker возвращает чекер для регистрации проверок
func (h *Handler) HealthChecker() *health.Checker {
	if h.health == nil {
		h.health = health.NewChecker("orders-service")
	}
	return h.health
}

// logUserActivity логирует активность пользователя через Redis Streams
func (h *Handler) logUserActivity(ctx context.Context, c *gin.Context, action, resourceType, details string, resourceID *int64) {
	if h.publisher == nil {
		return
	}

	user, ok := authcontext.FromRequest(c.Request)
	if !ok || !user.IsAuthenticated() {
		slog.WarnContext(ctx, "Cannot log activity - no authenticated user")
		return
	}

	eventDetails := map[string]interface{}{
		"resource_type": resourceType,
		"resource_id":   resourceID,
		"details":       details,
		"user_email":    user.Email,
		"user_name":     user.Name,
	}

	if err := h.publisher.PublishUserAction(ctx, strconv.FormatInt(user.ID, 10), action, eventDetails); err != nil {
		slog.WarnContext(ctx, "Failed to publish user action event", logger.Err(err))
	}
}
