package redisclient

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewRedisClientInvalidAddr(t *testing.T) {
	// Подключение к заведомо недоступному порту должно возвращать ошибку, а не паниковать
	client, err := New("127.0.0.1:99", "")
	require.Error(t, err)
	require.Nil(t, client)
}
