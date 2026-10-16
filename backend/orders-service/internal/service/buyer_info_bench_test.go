package service

import (
	"testing"
)

func BenchmarkParseBuyerInfo(b *testing.B) {
	input := "Получатель: +7 (999) 123-45-67 Иванов Иван Иванович"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseBuyerInfo(input)
	}
}

func BenchmarkParseBuyerInfoShort(b *testing.B) {
	input := "+79991234567"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseBuyerInfo(input)
	}
}

func BenchmarkParseBuyerInfoSpecial(b *testing.B) {
	input := "Самовывоз"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = parseBuyerInfo(input)
	}
}
