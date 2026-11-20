package postgres

import (
	"encoding/json"
	"fmt"
)

// StringArray кастомный тип для работы с JSONB массивами строк ([]string) в PostgreSQL.
// Использует нативные интерфейсы json.Marshaler и json.Unmarshaler для pgx/v5 (pgtype.JSONBCodec)
// и метод Scan(any) для прямого сканирования строк из pgx без зависимости от legacy database/sql/driver.
type StringArray []string

// MarshalJSON сериализует срез в JSON-массив, гарантируя "[]" вместо "null" для пустого среза.
func (a StringArray) MarshalJSON() ([]byte, error) {
	if len(a) == 0 {
		return []byte("[]"), nil
	}
	return json.Marshal([]string(a))
}

// UnmarshalJSON десериализует JSON-массив строк.
func (a *StringArray) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*a = StringArray{}
		return nil
	}
	var arr []string
	if err := json.Unmarshal(data, &arr); err != nil {
		return err
	}
	*a = arr
	return nil
}

// Scan реализует сканирование значений из PostgreSQL (JSONB, TEXT, BYTEA) для pgx.
// Принимает []byte или string без необходимости импорта стандартного пакета database/sql.
func (a *StringArray) Scan(value any) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	switch v := value.(type) {
	case []byte:
		return a.UnmarshalJSON(v)
	case string:
		return a.UnmarshalJSON([]byte(v))
	default:
		return fmt.Errorf("failed to scan StringArray: unsupported type %T", value)
	}
}
