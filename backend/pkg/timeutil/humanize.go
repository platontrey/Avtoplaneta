package timeutil

import (
	"fmt"
	"time"
)

// FormatPlural склоняет числительные на русском языке ("1 день", "2 дня", "5 дней")
func FormatPlural(count int, one, few, many string) string {
	abs := count
	if abs < 0 {
		abs = -abs
	}

	rem100 := abs % 100
	rem10 := abs % 10

	if rem100 >= 11 && rem100 <= 19 {
		return fmt.Sprintf("%d %s", count, many)
	}

	switch rem10 {
	case 1:
		return fmt.Sprintf("%d %s", count, one)
	case 2, 3, 4:
		return fmt.Sprintf("%d %s", count, few)
	default:
		return fmt.Sprintf("%d %s", count, many)
	}
}

// FormatTimeUntil форматирует человекочитаемый интервал до будущей даты ("через 1 день", "через 2 часа", "скоро", "уже прошло")
func FormatTimeUntil(targetTime time.Time, now time.Time) string {
	duration := targetTime.Sub(now)
	if duration <= 0 {
		return "уже прошло"
	}

	days := int(duration.Hours() / 24)
	if days > 0 {
		return "через " + FormatPlural(days, "день", "дня", "дней")
	}

	hours := int(duration.Hours()) % 24
	if hours > 0 {
		return "через " + FormatPlural(hours, "час", "часа", "часов")
	}

	minutes := int(duration.Minutes()) % 60
	if minutes > 0 {
		return "через " + FormatPlural(minutes, "минуту", "минуты", "минут")
	}

	return "скоро"
}
