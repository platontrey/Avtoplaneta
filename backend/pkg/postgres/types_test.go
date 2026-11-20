package postgres

import (
	"encoding/json"
	"testing"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringArray_Scan(t *testing.T) {
	t.Run("nil value", func(t *testing.T) {
		var sa StringArray
		err := sa.Scan(nil)
		require.NoError(t, err)
		assert.Equal(t, StringArray{}, sa)
	})

	t.Run("valid json bytes", func(t *testing.T) {
		var sa StringArray
		data := []byte(`["url1.jpg", "url2.jpg"]`)
		err := sa.Scan(data)
		require.NoError(t, err)
		assert.Equal(t, StringArray{"url1.jpg", "url2.jpg"}, sa)
	})

	t.Run("valid json string", func(t *testing.T) {
		var sa StringArray
		data := `["url1.jpg", "url2.jpg"]`
		err := sa.Scan(data)
		require.NoError(t, err)
		assert.Equal(t, StringArray{"url1.jpg", "url2.jpg"}, sa)
	})

	t.Run("empty json array", func(t *testing.T) {
		var sa StringArray
		data := []byte(`[]`)
		err := sa.Scan(data)
		require.NoError(t, err)
		assert.Equal(t, StringArray{}, sa)
	})

	t.Run("invalid json", func(t *testing.T) {
		var sa StringArray
		data := []byte(`invalid-json`)
		err := sa.Scan(data)
		require.Error(t, err)
	})

	t.Run("unsupported type", func(t *testing.T) {
		var sa StringArray
		err := sa.Scan(12345)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unsupported type")
	})
}

func TestStringArray_JSON(t *testing.T) {
	t.Run("marshal empty slice", func(t *testing.T) {
		var sa StringArray
		bytes, err := json.Marshal(sa)
		require.NoError(t, err)
		assert.Equal(t, "[]", string(bytes))
	})

	t.Run("marshal with elements", func(t *testing.T) {
		sa := StringArray{"item1", "item2"}
		bytes, err := json.Marshal(sa)
		require.NoError(t, err)
		assert.JSONEq(t, `["item1", "item2"]`, string(bytes))
	})

	t.Run("unmarshal null or empty", func(t *testing.T) {
		var sa StringArray
		err := json.Unmarshal([]byte("null"), &sa)
		require.NoError(t, err)
		assert.Equal(t, StringArray{}, sa)
	})

	t.Run("unmarshal elements", func(t *testing.T) {
		var sa StringArray
		err := json.Unmarshal([]byte(`["foo", "bar"]`), &sa)
		require.NoError(t, err)
		assert.Equal(t, StringArray{"foo", "bar"}, sa)
	})
}

func TestStringArray_PGX_JSONB_Scan(t *testing.T) {
	// Проверяем, что pgtype.Map (движок сканирования pgx/v5) нативно сканирует JSONB в StringArray
	m := pgtype.NewMap()
	var sa StringArray
	err := m.Scan(pgtype.JSONBOID, pgtype.TextFormatCode, []byte(`["img1.png", "img2.png"]`), &sa)
	require.NoError(t, err)
	assert.Equal(t, StringArray{"img1.png", "img2.png"}, sa)
}
