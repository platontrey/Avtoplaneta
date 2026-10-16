package domain

// PriceListMeta хранит отпечаток и статистику последнего собранного прайс-листа
type PriceListMeta struct {
	Version    string `json:"version"`
	PartsCount int    `json:"parts_count"`
	BuiltAt    string `json:"built_at"`
}
