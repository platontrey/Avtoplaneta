package domain

import (
	"fmt"
	"strconv"
	"time"
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
			// Обработка одиночного поля "photo" для обратной совместимости с фронтендом
			if photoStr, ok := value.(string); ok && photoStr != "" {
				// Если уже есть поле "photos" в обновлениях, игнорируем одиночное "photo"
				if _, hasPhotos := updates["photos"]; !hasPhotos {
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
