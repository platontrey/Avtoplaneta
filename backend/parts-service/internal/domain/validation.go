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

func normalizeQuantity(value interface{}) (interface{}, error) {
	qty, ok := value.(float64)
	if !ok {
		return value, nil
	}
	intQty := int(qty)
	if intQty < 0 {
		return nil, fmt.Errorf("количество должно быть не менее 0")
	}
	return intQty, nil
}

func normalizePrice(value interface{}) (interface{}, error) {
	price, ok := value.(string)
	if !ok {
		return value, nil
	}
	parsedPrice, err := strconv.ParseFloat(price, 64)
	if err != nil {
		return nil, fmt.Errorf("неверный формат цены")
	}
	return parsedPrice, nil
}

func normalizePhotos(value interface{}) interface{} {
	photosArray, ok := value.([]interface{})
	if !ok {
		return value
	}
	photos := make(StringArray, len(photosArray))
	for i, v := range photosArray {
		if str, ok := v.(string); ok {
			photos[i] = str
		}
	}
	return photos
}

func resolveSinglePhoto(value interface{}, part *Part, updates map[string]interface{}) (StringArray, bool) {
	photoStr, ok := value.(string)
	if !ok || photoStr == "" {
		return nil, false
	}
	if _, hasPhotos := updates["photos"]; hasPhotos {
		return nil, false
	}
	if part != nil && len(part.Photos) > 0 {
		newPhotos := make(StringArray, len(part.Photos))
		copy(newPhotos, part.Photos)
		newPhotos[0] = photoStr
		return newPhotos, true
	}
	return StringArray{photoStr}, true
}

// ProcessPartUpdates обрабатывает и валидирует обновления полей запчасти
func ProcessPartUpdates(updates map[string]interface{}, part *Part) (map[string]interface{}, error) {
	processedUpdates := make(map[string]interface{})
	for key, value := range updates {
		switch key {
		case "quantity":
			val, err := normalizeQuantity(value)
			if err != nil {
				return nil, err
			}
			processedUpdates[key] = val
		case "price":
			val, err := normalizePrice(value)
			if err != nil {
				return nil, err
			}
			processedUpdates[key] = val
		case "status":
			if status, ok := value.(bool); ok {
				processedUpdates[key] = status
			} else {
				processedUpdates[key] = value
			}
		case "photo":
			if photos, ok := resolveSinglePhoto(value, part, updates); ok {
				processedUpdates["photos"] = photos
			}
			// НЕ добавляем "photo" в processedUpdates - этого поля нет в БД
		case "photos":
			processedUpdates[key] = normalizePhotos(value)
		case "id", "created_at", "updated_at", "deleted_at", "to_delete_at_formatted", "time_until_deletion", "toDeleteAtFormatted", "timeUntilDeletion", "photoPreview":
			// Игнорируем поля, которые не должны обновляться напрямую или существуют только на фронтенде
			continue
		default:
			processedUpdates[key] = value
		}
	}
	return processedUpdates, nil
}

func formatPlural(count int, one, few, many string) string {
	if count == 1 {
		return fmt.Sprintf("через 1 %s", one)
	}
	if count < 5 {
		return fmt.Sprintf("через %d %s", count, few)
	}
	return fmt.Sprintf("через %d %s", count, many)
}

// FormatTimeUntil форматирует время до будущей даты
func FormatTimeUntil(targetTime time.Time, now time.Time) string {
	duration := targetTime.Sub(now)
	if duration <= 0 {
		return "уже прошло"
	}

	days := int(duration.Hours() / 24)
	if days > 0 {
		return formatPlural(days, "день", "дня", "дней")
	}

	hours := int(duration.Hours()) % 24
	if hours > 0 {
		return formatPlural(hours, "час", "часа", "часов")
	}

	minutes := int(duration.Minutes()) % 60
	if minutes > 0 {
		return formatPlural(minutes, "минуту", "минуты", "минут")
	}

	return "скоро"
}
