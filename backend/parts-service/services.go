package main

import (
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
	if part.Quantity < 1 {
		return fmt.Errorf("количество должно быть не менее 1")
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
			if photoStr, ok := value.(string); ok {
				if photoStr == "" {
					// Удалить файл, если он существует
					if part != nil && part.Photo != "" {
						photoPath := strings.TrimPrefix(part.Photo, "/")
						if err := os.Remove(photoPath); err != nil && !os.IsNotExist(err) {
							fmt.Printf("Warning: Failed to remove photo file %s: %v\n", photoPath, err)
						} else if err == nil {
							fmt.Printf("Successfully removed photo file: %s\n", photoPath)
						}
					}
					processedUpdates[key] = nil // Установить NULL
				} else {
					processedUpdates[key] = photoStr
				}
			} else {
				processedUpdates[key] = value
			}
		default:
			processedUpdates[key] = value
		}
	}
	return processedUpdates, nil
}

// HandlePhotoUpload обрабатывает загрузку фото для запчасти
func HandlePhotoUpload(c *gin.Context, partID uint) (string, error) {
	fmt.Printf("DEBUG HandlePhotoUpload: Starting upload for partID %d\n", partID)

	// Проверить, существует ли часть
	var part Part
	if err := db.First(&part, partID).Error; err != nil {
		fmt.Printf("DEBUG HandlePhotoUpload: Part not found for ID %d: %v\n", partID, err)
		return "", fmt.Errorf("часть не найдена")
	}
	fmt.Printf("DEBUG HandlePhotoUpload: Part found: ID=%d, current photo='%s'\n", part.ID, part.Photo)

	// Удалить старый файл фото, если он существует
	if part.Photo != "" {
		oldPhotoPath := strings.TrimPrefix(part.Photo, "/")
		fmt.Printf("DEBUG HandlePhotoUpload: Removing old photo file: %s\n", oldPhotoPath)
		if err := os.Remove(oldPhotoPath); err != nil && !os.IsNotExist(err) {
			fmt.Printf("Предупреждение: Не удалось удалить старый файл фото %s: %v\n", oldPhotoPath, err)
		} else if err == nil {
			fmt.Printf("Успешно удален старый файл фото: %s\n", oldPhotoPath)
		}
	}

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

	// Обновить часть с путем к фото
	photoPath := "/uploads/" + filename
	fmt.Printf("DEBUG HandlePhotoUpload: Updating database with photo path: %s\n", photoPath)
	if err := db.Model(&Part{}).Where("id = ?", partID).Update("photo", photoPath).Error; err != nil {
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

// DeletePhoto удаляет фото запчасти
func DeletePhoto(partID uint) error {
	// Проверить, существует ли часть
	var part Part
	if err := db.First(&part, partID).Error; err != nil {
		return fmt.Errorf("часть не найдена")
	}

	// Удалить файл фото, если он существует
	if part.Photo != "" {
		photoPath := strings.TrimPrefix(part.Photo, "/")
		if err := os.Remove(photoPath); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("не удалось удалить файл фото")
		}
	}

	// Обновить часть, установив photo в NULL
	if err := db.Model(&Part{}).Where("id = ?", partID).Update("photo", nil).Error; err != nil {
		return fmt.Errorf("не удалось обновить часть")
	}

	return nil
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
