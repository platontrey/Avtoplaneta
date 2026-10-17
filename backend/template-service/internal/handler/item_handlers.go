package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"avtoplaneta/pkg/httpcache"
	"template-service/internal/domain"
)

type CreateItemRequest struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Metadata    map[string]any `json:"metadata"`
}

type UpdateItemRequest struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description"`
	Status      string         `json:"status"`
	Metadata    map[string]any `json:"metadata"`
}

func (h *Handler) CreateItem(c *gin.Context) {
	var req CreateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный запрос", "details": err.Error()})
		return
	}

	item := &domain.Item{
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Metadata:    req.Metadata,
	}

	created, err := h.itemService.CreateItem(c.Request.Context(), item)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, created)
}

func (h *Handler) GetItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID элемента"})
		return
	}

	item, err := h.itemService.GetItem(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Элемент не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
		return
	}

	// Использование pkg/httpcache: если клиент прислал подходящий If-None-Match, возвращаем 304 Not Modified
	version := fmt.Sprintf("%d", item.UpdatedAt.UnixNano())
	if httpcache.ServeVersioned(c.Writer, c.Request, version, 5*time.Minute) {
		return
	}

	c.JSON(http.StatusOK, item)
}

func (h *Handler) ListItems(c *gin.Context) {
	page, _ := strconv.ParseInt(c.DefaultQuery("page", "1"), 10, 32)
	limit, _ := strconv.ParseInt(c.DefaultQuery("limit", "20"), 10, 32)

	items, err := h.itemService.ListItems(c.Request.Context(), int32(page), int32(limit))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка получения списка элементов"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"page":  page,
		"limit": limit,
	})
}

func (h *Handler) UpdateItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID элемента"})
		return
	}

	var req UpdateItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный запрос", "details": err.Error()})
		return
	}

	item := &domain.Item{
		ID:          id,
		Name:        req.Name,
		Description: req.Description,
		Status:      req.Status,
		Metadata:    req.Metadata,
	}

	updated, err := h.itemService.UpdateItem(c.Request.Context(), item)
	if err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Элемент не найден"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, updated)
}

func (h *Handler) DeleteItem(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Некорректный ID элемента"})
		return
	}

	if err := h.itemService.DeleteItem(c.Request.Context(), id); err != nil {
		if errors.Is(err, domain.ErrItemNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "Элемент не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка удаления элемента"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "deleted", "id": id})
}
