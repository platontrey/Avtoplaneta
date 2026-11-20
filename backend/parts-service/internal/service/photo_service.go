package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"parts-service/internal/domain"
	"parts-service/internal/repository"
)

// HandlePhotoUpload обрабатывает загрузку фото для запчасти
func HandlePhotoUpload(ctx context.Context, c *gin.Context, partID int64, repo repository.PartRepository) (string, error) {
	slog.DebugContext(ctx, "HandlePhotoUpload: starting upload", "part_id", partID)

	// Проверить, существует ли часть
	part, err := repo.FindByID(ctx, int64(partID))
	if err != nil {
		slog.WarnContext(ctx, "HandlePhotoUpload: part not found", "part_id", partID, "error", err)
		return "", errors.New("часть не найдена")
	}
	slog.DebugContext(ctx, "HandlePhotoUpload: part found", "part_id", part.ID, "photos", part.Photos)

	// Получить загруженный файл
	file, err := c.FormFile("photo")
	if err != nil {
		slog.WarnContext(ctx, "HandlePhotoUpload: no photo file provided", "error", err)
		return "", errors.New("файл фото не предоставлен")
	}
	slog.DebugContext(ctx, "HandlePhotoUpload: file received", "filename", file.Filename, "size", file.Size, "content_type", file.Header.Get("Content-Type"))

	// Валидировать тип файла
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".gif" {
			slog.WarnContext(ctx, "HandlePhotoUpload: invalid file type", "content_type", contentType, "ext", ext)
			return "", errors.New("файл должен быть изображением")
		}
	}

	// Валидировать размер файла (макс MaxUploadFileSize)
	if file.Size > MaxUploadFileSize {
		slog.WarnContext(ctx, "HandlePhotoUpload: file too large", "size", file.Size)
		return "", fmt.Errorf("размер файла должен быть менее %dМБ", MaxUploadFileSize/(1024*1024))
	}

	// Открыть загруженный файл для оптимизации
	src, err := file.Open()
	if err != nil {
		slog.ErrorContext(ctx, "HandlePhotoUpload: failed to open uploaded file", "error", err)
		return "", fmt.Errorf("не удалось открыть файл: %w", err)
	}
	defer src.Close()

	// Сгенерировать уникальное имя файла (.jpg для полной совместимости со всеми сервисами и Drom)
	filename := fmt.Sprintf("%d_%d.jpg", partID, time.Now().UnixNano()/1000000)
	filePath := filepath.Join("./uploads", filename)
	slog.DebugContext(ctx, "HandlePhotoUpload: generated filename", "filename", filename, "path", filePath)

	// Оптимизировать и сохранить файл в Progressive JPEG (авто-ориентация по EXIF, макс 2048px)
	slog.DebugContext(ctx, "HandlePhotoUpload: optimizing and saving file", "path", filePath)
	if err := OptimizeAndSaveImage(src, filePath); err != nil {
		slog.ErrorContext(ctx, "HandlePhotoUpload: failed to process and save file", "error", err)
		return "", fmt.Errorf("не удалось обработать и сохранить фото: %w", err)
	}
	slog.DebugContext(ctx, "HandlePhotoUpload: file optimized and saved successfully")

	// Обновить часть с путем к фото (добавить в массив photos)
	photoPath := "/uploads/" + filename
	slog.DebugContext(ctx, "HandlePhotoUpload: adding photo path", "photo_path", photoPath)

	// Получить текущий массив фото
	currentPhotos := make(domain.StringArray, len(part.Photos))
	copy(currentPhotos, part.Photos)

	// Для обратной совместимости, если есть старое поле photo и его нет в массиве
	if part.Photo != "" {
		found := false
		for _, photo := range currentPhotos {
			if photo == part.Photo {
				found = true
				break
			}
		}
		if !found {
			currentPhotos = append(currentPhotos, part.Photo)
		}
	}

	// Добавить новое фото, если его еще нет в массиве
	photoExists := false
	for _, existingPhoto := range currentPhotos {
		if existingPhoto == photoPath {
			photoExists = true
			break
		}
	}
	if !photoExists {
		currentPhotos = append(currentPhotos, photoPath)
		slog.DebugContext(ctx, "HandlePhotoUpload: photo added to array")
	} else {
		slog.DebugContext(ctx, "HandlePhotoUpload: photo already exists in array, skipping duplicate")
	}

	// Обновить используя repository - domain.StringArray автоматически сериализуется
	if err := repo.UpdatePartPhotos(ctx, int64(partID), currentPhotos); err != nil {
		slog.ErrorContext(ctx, "HandlePhotoUpload: failed to update database", "error", err)
		// Попытаться очистить загруженный файл, если обновление базы данных не удалось
		if removeErr := os.Remove(filePath); removeErr != nil {
			slog.WarnContext(ctx, "HandlePhotoUpload: failed to remove uploaded file after database error", "error", removeErr)
		}
		return "", fmt.Errorf("не удалось обновить часть с фото: %w", err)
	}
	slog.InfoContext(ctx, "HandlePhotoUpload: photo uploaded successfully", "part_id", partID, "photo_path", photoPath)

	return photoPath, nil
}

func deleteAllPartPhotos(ctx context.Context, partID int64, photos domain.StringArray, repo repository.PartRepository) error {
	slog.DebugContext(ctx, "DeletePhoto: deleting all photos", "part_id", partID, "count", len(photos))
	for _, p := range photos {
		if p == "" {
			continue
		}
		filePath := strings.TrimPrefix(p, "/")
		slog.DebugContext(ctx, "DeletePhoto: removing file", "path", filePath)
		if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
			slog.WarnContext(ctx, "DeletePhoto: failed to remove photo file", "path", filePath, "error", err)
		}
	}

	if err := repo.UpdatePartPhotos(ctx, partID, domain.StringArray{}); err != nil {
		slog.ErrorContext(ctx, "DeletePhoto: failed to update database", "error", err)
		return errors.New("не удалось обновить часть")
	}
	slog.InfoContext(ctx, "DeletePhoto: all photos deleted successfully", "part_id", partID)
	return nil
}

func deleteSpecificPartPhoto(ctx context.Context, partID int64, photoPath string, photos domain.StringArray, repo repository.PartRepository) error {
	slog.DebugContext(ctx, "DeletePhoto: deleting specific photo", "part_id", partID, "photo_path", photoPath)
	newPhotos := make(domain.StringArray, 0, len(photos))
	photoDeleted := false

	for _, p := range photos {
		if p == photoPath && !photoDeleted {
			filePath := strings.TrimPrefix(p, "/")
			slog.DebugContext(ctx, "DeletePhoto: removing file", "path", filePath)
			if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
				slog.WarnContext(ctx, "DeletePhoto: failed to remove photo file", "path", filePath, "error", err)
			}
			photoDeleted = true
			slog.DebugContext(ctx, "DeletePhoto: photo found and file removed", "photo_path", photoPath)
			continue
		}
		newPhotos = append(newPhotos, p)
	}

	if !photoDeleted {
		slog.WarnContext(ctx, "DeletePhoto: photo not found in array", "photo_path", photoPath)
		return errors.New("фото не найдено в массиве")
	}

	slog.DebugContext(ctx, "DeletePhoto: new photos array", "photos", newPhotos)
	if err := repo.UpdatePartPhotos(ctx, partID, newPhotos); err != nil {
		slog.ErrorContext(ctx, "DeletePhoto: failed to update database", "error", err)
		return errors.New("не удалось обновить часть")
	}
	slog.InfoContext(ctx, "DeletePhoto: specific photo deleted successfully", "part_id", partID, "photo_path", photoPath)
	return nil
}

// DeletePhoto удаляет конкретное фото или все фото запчасти
// Если photoPath пустой, удаляет все фото
func DeletePhoto(ctx context.Context, partID int64, photoPath string, repo repository.PartRepository) error {
	slog.DebugContext(ctx, "DeletePhoto: starting deletion", "part_id", partID, "photo_path", photoPath)

	part, err := repo.FindByID(ctx, partID)
	if err != nil {
		slog.WarnContext(ctx, "DeletePhoto: part not found", "part_id", partID, "error", err)
		return errors.New("часть не найдена")
	}

	slog.DebugContext(ctx, "DeletePhoto: part found", "part_id", part.ID, "photos", part.Photos)

	if photoPath == "" {
		return deleteAllPartPhotos(ctx, partID, part.Photos, repo)
	}
	return deleteSpecificPartPhoto(ctx, partID, photoPath, part.Photos, repo)
}

// contains проверяет, содержит ли слайс строку
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// SavePhotoFromBytes сохраняет фото из байтов (для gRPC streaming upload)
func SavePhotoFromBytes(ctx context.Context, partID int64, data []byte, repo repository.PartRepository) (string, error) {
	if len(data) == 0 {
		return "", errors.New("пустые данные фото")
	}

	// Проверить, существует ли часть
	part, err := repo.FindByID(ctx, int64(partID))
	if err != nil {
		return "", errors.New("часть не найдена")
	}

	// Валидировать размер (макс MaxUploadFileSize)
	if len(data) > MaxUploadFileSize {
		return "", fmt.Errorf("размер файла должен быть менее %dМБ", MaxUploadFileSize/(1024*1024))
	}

	// Сгенерировать имя файла (.jpg)
	filename := fmt.Sprintf("%d_%d.jpg", partID, time.Now().UnixNano()/1000000)
	filePath := filepath.Join("./uploads", filename)

	// Оптимизировать и сохранить файл
	if err := OptimizeAndSaveImageBytes(data, filePath); err != nil {
		return "", fmt.Errorf("не удалось обработать и сохранить фото: %w", err)
	}

	// Обновить массив фото
	photoPath := "/uploads/" + filename
	currentPhotos := make(domain.StringArray, len(part.Photos))
	copy(currentPhotos, part.Photos)
	currentPhotos = append(currentPhotos, photoPath)

	if err := repo.UpdatePartPhotos(ctx, int64(partID), currentPhotos); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("не удалось обновить часть с фото: %w", err)
	}

	return photoPath, nil
}

// DeletePhotoFile удаляет файл фото
func DeletePhotoFile(photoPath string) error {
	if photoPath == "" {
		return nil
	}
	// Извлечь имя файла из пути (удалить префикс "/uploads/")
	filePath := strings.TrimPrefix(photoPath, "/")
	if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("не удалось удалить файл фото: %w", err)
	}
	return nil
}
