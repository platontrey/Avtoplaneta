package handler

import (
	"template-service/internal/service"
)

type Handler struct {
	itemService service.ItemService
}

func NewHandler(itemService service.ItemService) *Handler {
	return &Handler{
		itemService: itemService,
	}
}
