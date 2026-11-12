package handler

import (
	"avtoplaneta/pkg/health"
	"template-service/internal/service"
)

type Handler struct {
	itemService service.ItemService
	health      *health.Checker
}

func NewHandler(itemService service.ItemService) *Handler {
	return &Handler{
		itemService: itemService,
		health:      health.NewChecker("template-service"),
	}
}

func (h *Handler) HealthChecker() *health.Checker {
	if h.health == nil {
		h.health = health.NewChecker("template-service")
	}
	return h.health
}
