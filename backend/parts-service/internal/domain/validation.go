package domain

import (
	"errors"
	"strconv"
	"time"

	"avtoplaneta/pkg/timeutil"
)

// ValidatePart выполняет валидацию данных запчасти
func ValidatePart(part *Part) error {
	if len(part.Name) > 255 {
		return errors.New("название слишком длинное")
	}
	if len(part.Description) > 1000 {
		return errors.New("описание слишком длинное")
	}
	if part.Quantity < 0 {
		return errors.New("количество должно быть не менее 0")
	}
	if part.Price < 0 {
		return errors.New("цена не может быть отрицательной")
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
		return nil, errors.New("количество должно быть не менее 0")
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
		return nil, errors.New("неверный формат цены")
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

// FormatTimeUntil форматирует человекочитаемое время до будущей даты.
// Делегирует платформенной функции avtoplaneta/pkg/timeutil.FormatTimeUntil.
func FormatTimeUntil(targetTime time.Time, now time.Time) string {
	return timeutil.FormatTimeUntil(targetTime, now)
}
