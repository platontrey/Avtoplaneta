package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// HandlePhotoUpload обрабатывает загрузку фото для запчасти
func HandlePhotoUpload(c *gin.Context, partID int64) (string, error) {
	fmt.Printf("DEBUG HandlePhotoUpload: Starting upload for partID %d\n", partID)

	// Проверить, существует ли часть
	repo := NewPartRepository(dbPool)
	part, err := repo.FindByID(c.Request.Context(), int64(partID))
	if err != nil {
		fmt.Printf("DEBUG HandlePhotoUpload: Part not found for ID %d: %v\n", partID, err)
		return "", fmt.Errorf("часть не найдена")
	}
	fmt.Printf("DEBUG HandlePhotoUpload: Part found: ID=%d, current photos=%v\n", part.ID, part.Photos)

	// Получить загруженный файл
	file, err := c.FormFile("photo")
	if err != nil {
		fmt.Printf("DEBUG HandlePhotoUpload: No photo file provided: %v\n", err)
		return "", fmt.Errorf("файл фото не предоставлен")
	}
	fmt.Printf("DEBUG HandlePhotoUpload: File received: name='%s', size=%d, content-type='%s'\n", file.Filename, file.Size, file.Header.Get("Content-Type"))

	// Валидировать тип файла
	contentType := file.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" && ext != ".webp" && ext != ".gif" {
			fmt.Printf("DEBUG HandlePhotoUpload: Invalid file type: %s (ext: %s)\n", contentType, ext)
			return "", fmt.Errorf("файл должен быть изображением")
		}
	}

	// Валидировать размер файла (макс MaxUploadFileSize)
	if file.Size > MaxUploadFileSize {
		fmt.Printf("DEBUG HandlePhotoUpload: File too large: %d bytes\n", file.Size)
		return "", fmt.Errorf("размер файла должен быть менее %dМБ", MaxUploadFileSize/(1024*1024))
	}

	// Открыть загруженный файл для оптимизации
	src, err := file.Open()
	if err != nil {
		fmt.Printf("DEBUG HandlePhotoUpload: Failed to open uploaded file: %v\n", err)
		return "", fmt.Errorf("не удалось открыть файл: %v", err)
	}
	defer src.Close()

	// Сгенерировать уникальное имя файла (.jpg для полной совместимости со всеми сервисами и Drom)
	filename := fmt.Sprintf("%d_%d.jpg", partID, time.Now().UnixNano()/1000000)
	filePath := filepath.Join("./uploads", filename)
	fmt.Printf("DEBUG HandlePhotoUpload: Generated filename: %s, path: %s\n", filename, filePath)

	// Оптимизировать и сохранить файл в Progressive JPEG (авто-ориентация по EXIF, макс 2048px)
	fmt.Printf("DEBUG HandlePhotoUpload: Optimizing and saving file to: %s\n", filePath)
	if err := OptimizeAndSaveImage(src, filePath); err != nil {
		fmt.Printf("DEBUG HandlePhotoUpload: Failed to process and save file: %v\n", err)
		return "", fmt.Errorf("не удалось обработать и сохранить фото: %v", err)
	}
	fmt.Printf("DEBUG HandlePhotoUpload: File optimized and saved successfully\n")

	// Обновить часть с путем к фото (добавить в массив photos)
	photoPath := "/uploads/" + filename
	fmt.Printf("DEBUG HandlePhotoUpload: Adding photo path to array: %s\n", photoPath)

	// Получить текущий массив фото
	currentPhotos := make(StringArray, len(part.Photos))
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
		fmt.Printf("DEBUG HandlePhotoUpload: Photo added to array\n")
	} else {
		fmt.Printf("DEBUG HandlePhotoUpload: Photo already exists in array, not adding duplicate\n")
	}

	// Обновить используя repository - StringArray автоматически сериализуется
	if err := repo.UpdatePartPhotos(c.Request.Context(), int64(partID), currentPhotos); err != nil {
		fmt.Printf("DEBUG HandlePhotoUpload: Failed to update database: %v\n", err)
		// Попытаться очистить загруженный файл, если обновление базы данных не удалось
		if err := os.Remove(filePath); err != nil {
			fmt.Printf("Warning: Failed to remove uploaded file after database error: %v\n", err)
		}
		return "", fmt.Errorf("не удалось обновить часть с фото: %v", err)
	}
	fmt.Printf("DEBUG HandlePhotoUpload: Database updated successfully\n")

	return photoPath, nil
}

// DeletePhoto удаляет конкретное фото или все фото запчасти
// Если photoPath пустой, удаляет все фото
func DeletePhoto(partID int64, photoPath string) error {
	fmt.Printf("DeletePhoto: Starting deletion for partID=%d, photoPath='%s'\n", partID, photoPath)

	// Проверить, существует ли часть
	repo := NewPartRepository(dbPool)
	part, err := repo.FindByID(context.Background(), int64(partID))
	if err != nil {
		fmt.Printf("DeletePhoto: Part not found: %v\n", err)
		return fmt.Errorf("часть не найдена")
	}

	fmt.Printf("DeletePhoto: Part found, current photos: %v\n", part.Photos)

	if photoPath == "" {
		fmt.Printf("DeletePhoto: Deleting all photos\n")
		// Удалить все фото
		for _, p := range part.Photos {
			if p != "" {
				filePath := strings.TrimPrefix(p, "/")
				fmt.Printf("DeletePhoto: Removing file: %s\n", filePath)
				if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
					fmt.Printf("Warning: Failed to remove photo file %s: %v\n", filePath, err)
				}
			}
		}

		// Обновить часть, установив photos в пустой массив
		if err := repo.UpdatePartPhotos(context.Background(), int64(partID), StringArray{}); err != nil {
			fmt.Printf("DeletePhoto: Failed to update database: %v\n", err)
			return fmt.Errorf("не удалось обновить часть")
		}
		fmt.Printf("DeletePhoto: All photos deleted successfully\n")
	} else {
		fmt.Printf("DeletePhoto: Deleting specific photo: %s\n", photoPath)
		// Удалить только первое вхождение конкретного фото из массива
		newPhotos := make(StringArray, 0)
		photoDeleted := false

		for _, p := range part.Photos {
			if p == photoPath && !photoDeleted {
				// Удалить файл только для первого вхождения
				filePath := strings.TrimPrefix(p, "/")
				fmt.Printf("DeletePhoto: Removing file: %s\n", filePath)
				if err := os.Remove(filePath); err != nil && !os.IsNotExist(err) {
					fmt.Printf("Warning: Failed to remove photo file %s: %v\n", filePath, err)
				}
				photoDeleted = true
				fmt.Printf("DeletePhoto: Photo found and file removed\n")
				// Не добавлять это фото в newPhotos
			} else {
				newPhotos = append(newPhotos, p)
			}
		}

		if !photoDeleted {
			fmt.Printf("DeletePhoto: Photo not found in array\n")
			return fmt.Errorf("фото не найдено в массиве")
		}

		fmt.Printf("DeletePhoto: New photos array: %v\n", newPhotos)
		// Обновить массив фото
		if err := repo.UpdatePartPhotos(context.Background(), int64(partID), newPhotos); err != nil {
			fmt.Printf("DeletePhoto: Failed to update database: %v\n", err)
			return fmt.Errorf("не удалось обновить часть")
		}
		fmt.Printf("DeletePhoto: Specific photo deleted successfully\n")
	}

	return nil
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
func SavePhotoFromBytes(partID int64, data []byte) (string, error) {
	if len(data) == 0 {
		return "", fmt.Errorf("пустые данные фото")
	}

	// Проверить, существует ли часть
	repo := NewPartRepository(dbPool)
	part, err := repo.FindByID(context.Background(), int64(partID))
	if err != nil {
		return "", fmt.Errorf("часть не найдена")
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
		return "", fmt.Errorf("не удалось обработать и сохранить фото: %v", err)
	}

	// Обновить массив фото
	photoPath := "/uploads/" + filename
	currentPhotos := make(StringArray, len(part.Photos))
	copy(currentPhotos, part.Photos)
	currentPhotos = append(currentPhotos, photoPath)

	if err := repo.UpdatePartPhotos(context.Background(), int64(partID), currentPhotos); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("не удалось обновить часть с фото: %v", err)
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
		return fmt.Errorf("не удалось удалить файл фото: %v", err)
	}
	return nil
}
