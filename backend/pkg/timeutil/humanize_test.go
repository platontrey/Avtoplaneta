package timeutil

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFormatPlural(t *testing.T) {
	tests := []struct {
		count    int
		expected string
	}{
		{count: 1, expected: "1 день"},
		{count: 2, expected: "2 дня"},
		{count: 4, expected: "4 дня"},
		{count: 5, expected: "5 дней"},
		{count: 11, expected: "11 дней"},
		{count: 14, expected: "14 дней"},
		{count: 21, expected: "21 день"},
		{count: 22, expected: "22 дня"},
		{count: 25, expected: "25 дней"},
	}

	for _, tt := range tests {
		res := FormatPlural(tt.count, "день", "дня", "дней")
		assert.Equal(t, tt.expected, res)
	}
}

func TestFormatTimeUntil(t *testing.T) {
	now := time.Date(2026, 11, 19, 12, 0, 0, 0, time.UTC)

	t.Run("past time", func(t *testing.T) {
		target := now.Add(-1 * time.Hour)
		assert.Equal(t, "уже прошло", FormatTimeUntil(target, now))
	})

	t.Run("equal time", func(t *testing.T) {
		assert.Equal(t, "уже прошло", FormatTimeUntil(now, now))
	})

	t.Run("seconds left (soon)", func(t *testing.T) {
		target := now.Add(30 * time.Second)
		assert.Equal(t, "скоро", FormatTimeUntil(target, now))
	})

	t.Run("minutes left", func(t *testing.T) {
		target := now.Add(5 * time.Minute)
		assert.Equal(t, "через 5 минут", FormatTimeUntil(target, now))

		target2 := now.Add(1 * time.Minute)
		assert.Equal(t, "через 1 минуту", FormatTimeUntil(target2, now))

		target3 := now.Add(3 * time.Minute)
		assert.Equal(t, "через 3 минуты", FormatTimeUntil(target3, now))
	})

	t.Run("hours left", func(t *testing.T) {
		target := now.Add(2 * time.Hour)
		assert.Equal(t, "через 2 часа", FormatTimeUntil(target, now))

		target1 := now.Add(1 * time.Hour)
		assert.Equal(t, "через 1 час", FormatTimeUntil(target1, now))

		target5 := now.Add(5 * time.Hour)
		assert.Equal(t, "через 5 часов", FormatTimeUntil(target5, now))
	})

	t.Run("days left", func(t *testing.T) {
		target := now.Add(24 * time.Hour)
		assert.Equal(t, "через 1 день", FormatTimeUntil(target, now))

		target2 := now.Add(48 * time.Hour)
		assert.Equal(t, "через 2 дня", FormatTimeUntil(target2, now))

		target10 := now.Add(10 * 24 * time.Hour)
		assert.Equal(t, "через 10 дней", FormatTimeUntil(target10, now))
	})
}
