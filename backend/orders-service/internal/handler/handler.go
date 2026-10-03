package handler

import (
	"context"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/authcontext"
	"orders-service/internal/domain"
	"orders-service/internal/service"
)

// Handler содержит все HTTP handlers для orders-service
type Handler struct {
	ordersService service.OrdersService
	publisher     domain.EventPublisher
}

// NewHandler создает новый handler с dependency injection
func NewHandler(ordersService service.OrdersService, publisher domain.EventPublisher) *Handler {
	return &Handler{
		ordersService: ordersService,
		publisher:     publisher,
	}
}

// logUserActivity логирует активность пользователя через Redis Streams
func (h *Handler) logUserActivity(ctx context.Context, c *gin.Context, action, resourceType, details string, resourceID *int64) {
	if h.publisher == nil {
		return
	}

	user, ok := authcontext.FromRequest(c.Request)
	if !ok || !user.IsAuthenticated() {
		logrus.Warn("Cannot log activity - no authenticated user")
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
		logrus.WithError(err).Warn("Failed to publish user action event")
	}
}
