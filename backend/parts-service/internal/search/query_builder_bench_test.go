package search

import (
	"testing"

	"parts-service/internal/domain"
)

func BenchmarkTransliterateLatinToCyrillic(b *testing.B) {
	input := "bamper perednij dlya toyota camry acv30"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = TransliterateLatinToCyrillic(input)
	}
}

func BenchmarkConvertQwertyToRussian(b *testing.B) {
	input := "gthtlybq ,fvgth lkz njqjnf rfvhb"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ConvertQwertyToRussian(input)
	}
}

func BenchmarkBuildElasticsearchQuery(b *testing.B) {
	params := domain.InventoryQueryParams{
		Search:   "бампер передний toyota camry",
		Category: "Кузовные детали",
		Page:     1,
		Limit:    50,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildElasticsearchQuery(params)
	}
}

func BenchmarkSplitMultiSearchQuery(b *testing.B) {
	query := "бампер передний, крыло левое; фара правая\nкапот ИЛИ дверь"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = SplitMultiSearchQuery(query)
	}
}
