package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"parts-service/internal/domain"
)

// UploadPartPhotoHandler загружает фото для запчасти
func (h *Handler) UploadPartPhotoHandler(c *gin.Context) {
	idStr := c.Param("id")
	partID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	ctx := c.Request.Context()
	photoPath, err := h.inventoryService.UploadPartPhoto(ctx, int64(partID), c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Логируем загрузку фото
	partIDUint := int64(partID)
	h.logUserActivity(c, "upload_photo", "part", fmt.Sprintf("Uploaded photo for part ID: %d", partID), &partIDUint)

	// Получить файл для размера и типа
	file, _ := c.FormFile("photo")

	// Получить обновленную часть для возврата всех фото
	updatedPart, err := h.inventoryService.GetPartByID(ctx, int64(partID))
	if err != nil {
		slog.DebugContext(ctx, "UploadPartPhotoHandler: failed to fetch updated part", "error", err)
		updatedPart = &domain.Part{}
	} else {
		slog.DebugContext(ctx, "UploadPartPhotoHandler: updated part photos", "photos", updatedPart.Photos)
	}

	response := gin.H{
		"message": "Фото загружено успешно",
		"photo":   photoPath,
		"photos":  updatedPart.Photos,
	}

	// Добавить информацию о файле, если он существует
	if file != nil {
		response["filename"] = file.Filename
		response["size"] = file.Size
		response["type"] = file.Header.Get("Content-Type")
	}

	c.JSON(http.StatusOK, response)
}

// DeletePartPhotoHandler удаляет фото запчасти
func (h *Handler) DeletePartPhotoHandler(c *gin.Context) {
	idStr := c.Param("id")
	partID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID части"})
		return
	}

	photoPath := c.Query("photo")
	if photoPath == "" {
		photoPath = c.Query("photo_url")
	}
	slog.DebugContext(c.Request.Context(), "DeletePartPhotoHandler: started", "part_id", partID, "photo_path", photoPath)

	ctx := c.Request.Context()
	if err := h.inventoryService.DeletePartPhoto(ctx, int64(partID), photoPath); err != nil {
		slog.ErrorContext(ctx, "DeletePartPhotoHandler: error deleting photo", "error", err, "part_id", partID, "photo_path", photoPath)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем удаление фото
	partIDUint := int64(partID)
	photoInfo := "all photos"
	if photoPath != "" {
		photoInfo = photoPath
	}
	slog.InfoContext(ctx, "DeletePartPhotoHandler: photo deleted", "photo_info", photoInfo, "part_id", partID)
	h.logUserActivity(c, "delete_photo", "part", fmt.Sprintf("Deleted photo %s for part ID: %d", photoInfo, partID), &partIDUint)

	c.JSON(http.StatusOK, gin.H{"message": "Фото удалено успешно"})
}
