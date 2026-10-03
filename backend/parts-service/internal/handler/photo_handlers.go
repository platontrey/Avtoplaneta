package handler

import (
	"fmt"
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
		fmt.Printf("DEBUG UploadPartPhotoHandler: Failed to fetch updated part: %v\n", err)
		updatedPart = &domain.Part{}
	} else {
		fmt.Printf("DEBUG UploadPartPhotoHandler: Updated part photos: %v\n", updatedPart.Photos)
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
	fmt.Printf("DeletePartPhotoHandler: partID=%d, photoPath='%s'\n", partID, photoPath)

	ctx := c.Request.Context()
	if err := h.inventoryService.DeletePartPhoto(ctx, int64(partID), photoPath); err != nil {
		fmt.Printf("DeletePartPhotoHandler: Error deleting photo: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Логируем удаление фото
	partIDUint := int64(partID)
	photoInfo := "all photos"
	if photoPath != "" {
		photoInfo = photoPath
	}
	fmt.Printf("DeletePartPhotoHandler: Successfully deleted %s for part ID %d\n", photoInfo, partID)
	h.logUserActivity(c, "delete_photo", "part", fmt.Sprintf("Deleted photo %s for part ID: %d", photoInfo, partID), &partIDUint)

	c.JSON(http.StatusOK, gin.H{"message": "Фото удалено успешно"})
}
