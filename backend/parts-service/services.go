package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// ValidatePart выполняет валидацию данных запчасти
func ValidatePart(part *Part) error {
	if len(part.Name) > 255 {
		return fmt.Errorf("название слишком длинное")
	}
	if len(part.Description) > 1000 {
		return fmt.Errorf("описание слишком длинное")
	}
	if part.Quantity < 0 {
		return fmt.Errorf("количество должно быть не менее 0")
	}
	if part.Price < 0 {
		return fmt.Errorf("цена не может быть отрицательной")
	}
	return nil
}

// ProcessPartUpdates обрабатывает и валидирует обновления полей запчасти
func ProcessPartUpdates(updates map[string]interface{}, part *Part) (map[string]interface{}, error) {
	processedUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case "quantity":
			if qty, ok := value.(float64); ok {
				intQty := int(qty)
				if intQty < 0 {
					return nil, fmt.Errorf("количество должно быть не менее 0")
				}
				processedUpdates[key] = intQty
			} else {
				processedUpdates[key] = value
			}
		case "price":
			if price, ok := value.(string); ok {
				if parsedPrice, err := strconv.ParseFloat(price, 64); err == nil {
					processedUpdates[key] = parsedPrice
				} else {
					return nil, fmt.Errorf("неверный формат цены")
				}
			} else {
				processedUpdates[key] = value
			}
		case "status":
			if status, ok := value.(bool); ok {
				processedUpdates[key] = status
			} else {
				processedUpdates[key] = value
			}
		case "photo":
			// ВАЖНО: Поле photo больше не существует в БД после миграции
			// Преобразуем его в массив photos для обратной совместимости
			if photoStr, ok := value.(string); ok {
				if photoStr == "" {
					// Удалить все фото
					if part != nil && len(part.Photos) > 0 {
						// Удалить файлы
						for _, p := range part.Photos {
							photoPath := strings.TrimPrefix(p, "/")
							if err := os.Remove(photoPath); err != nil && !os.IsNotExist(err) {
								fmt.Printf("Warning: Failed to remove photo file %s: %v\n", photoPath, err)
							}
						}
					}
					processedUpdates["photos"] = StringArray{}
				} else {
					// Преобразовать одиночное фото в массив (если это первое фото)
					// или добавить в существующий массив
					if part != nil && len(part.Photos) > 0 {
						// Заменить первый элемент массива (совместимость с legacy behavior)
						newPhotos := make(StringArray, len(part.Photos))
						copy(newPhotos, part.Photos)
						newPhotos[0] = photoStr
						processedUpdates["photos"] = newPhotos
					} else {
						// Создать новый массив с одним элементом
						processedUpdates["photos"] = StringArray{photoStr}
					}
				}
			}
			// НЕ добавляем "photo" в processedUpdates - этого поля нет в БД
		case "photos":
			if photosArray, ok := value.([]interface{}); ok {
				photos := make(StringArray, len(photosArray))
				for i, v := range photosArray {
					if str, ok := v.(string); ok {
						photos[i] = str
					}
				}
				processedUpdates[key] = photos
			} else {
				processedUpdates[key] = value
			}
		case "id", "created_at", "updated_at", "deleted_at", "to_delete_at_formatted", "time_until_deletion", "toDeleteAtFormatted", "timeUntilDeletion", "photoPreview":
			// Игнорируем поля, которые не должны обновляться напрямую или существуют только на фронтенде
			continue
		default:
			processedUpdates[key] = value
		}
	}
	return processedUpdates, nil
}

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
	if !strings.HasPrefix(file.Header.Get("Content-Type"), "image/") {
		fmt.Printf("DEBUG HandlePhotoUpload: Invalid file type: %s\n", file.Header.Get("Content-Type"))
		return "", fmt.Errorf("файл должен быть изображением")
	}

	// Валидировать размер файла (макс 5МБ)
	if file.Size > 5*1024*1024 {
		fmt.Printf("DEBUG HandlePhotoUpload: File too large: %d bytes\n", file.Size)
		return "", fmt.Errorf("размер файла должен быть менее 5МБ")
	}

	// Создать директорию uploads, если она не существует
	fmt.Printf("DEBUG HandlePhotoUpload: Creating uploads directory\n")
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		fmt.Printf("DEBUG HandlePhotoUpload: Failed to create uploads directory: %v\n", err)
		return "", fmt.Errorf("не удалось создать директорию uploads")
	}

	// Сгенерировать уникальное имя файла
	ext := filepath.Ext(file.Filename)
	if ext == "" {
		ext = ".jpg" // расширение по умолчанию
	}
	filename := fmt.Sprintf("%d_%d%s", partID, time.Now().Unix(), ext)
	filePath := filepath.Join("./uploads", filename)
	fmt.Printf("DEBUG HandlePhotoUpload: Generated filename: %s, path: %s\n", filename, filePath)

	// Сохранить файл
	fmt.Printf("DEBUG HandlePhotoUpload: Saving file to: %s\n", filePath)
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		fmt.Printf("DEBUG HandlePhotoUpload: Failed to save file: %v\n", err)
		return "", fmt.Errorf("не удалось сохранить фото: %v", err)
	}
	fmt.Printf("DEBUG HandlePhotoUpload: File saved successfully\n")

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

	// Валидировать размер (макс 5МБ)
	if len(data) > 5*1024*1024 {
		return "", fmt.Errorf("размер файла должен быть менее 5МБ")
	}

	// Создать директорию uploads
	if err := os.MkdirAll("./uploads", 0755); err != nil {
		return "", fmt.Errorf("не удалось создать директорию uploads")
	}

	// Сгенерировать имя файла
	filename := fmt.Sprintf("%d_%d.jpg", partID, time.Now().Unix())
	filePath := filepath.Join("./uploads", filename)

	// Сохранить файл
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", fmt.Errorf("не удалось сохранить фото: %v", err)
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

// FormatTimeUntil форматирует время до будущей даты
func FormatTimeUntil(targetTime time.Time, now time.Time) string {
	duration := targetTime.Sub(now)

	if duration <= 0 {
		return "уже прошло"
	}

	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	if days > 0 {
		if days == 1 {
			return "через 1 день"
		} else if days < 5 {
			return fmt.Sprintf("через %d дня", days)
		} else {
			return fmt.Sprintf("через %d дней", days)
		}
	} else if hours > 0 {
		if hours == 1 {
			return "через 1 час"
		} else if hours < 5 {
			return fmt.Sprintf("через %d часа", hours)
		} else {
			return fmt.Sprintf("через %d часов", hours)
		}
	} else if minutes > 0 {
		if minutes == 1 {
			return "через 1 минуту"
		} else if minutes < 5 {
			return fmt.Sprintf("через %d минуты", minutes)
		} else {
			return fmt.Sprintf("через %d минут", minutes)
		}
	} else {
		return "скоро"
	}
}
