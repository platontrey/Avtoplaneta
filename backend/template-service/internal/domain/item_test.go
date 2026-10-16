package domain

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestItem_Validate(t *testing.T) {
	t.Run("валидный элемент с дефолтным статусом", func(t *testing.T) {
		item := &Item{
			Name: "Тестовая позиция",
		}
		err := item.Validate()
		require.NoError(t, err)
		require.Equal(t, "active", item.Status)
	})

	t.Run("пустое название возвращает ошибку", func(t *testing.T) {
		item := &Item{
			Name: "   ",
		}
		err := item.Validate()
		require.Error(t, err)
	})

	t.Run("слишком длинное название возвращает ошибку", func(t *testing.T) {
		item := &Item{
			Name: strings.Repeat("A", 256),
		}
		err := item.Validate()
		require.Error(t, err)
	})
}
