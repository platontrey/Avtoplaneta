package main

import (
	"fmt"
	"testing"
)

func BenchmarkGenerateXMLPriceList1000(b *testing.B) {
	parts := make([]Part, 1000)
	sellerINNs := make(map[int64]string)
	for i := 0; i < 1000; i++ {
		sellerID := int64(i % 10)
		sellerINNs[sellerID] = fmt.Sprintf("770123456%02d", sellerID)
		parts[i] = Part{
			Name:         fmt.Sprintf("Бампер передний для Toyota Camry #%d", i),
			Status:       true,
			Quantity:     1,
			Price:        15000.0,
			SellerID:     sellerID,
			Salesman:     "ООО Автопланета",
			Brand:        "Toyota",
			Model:        "Camry",
			OEMCode:      fmt.Sprintf("52119-339%02d", i%100),
			SupplierCode: fmt.Sprintf("SUP-%d", i),
			Location:     "Склад 1",
			Photos: []string{
				fmt.Sprintf("/uploads/part_%d_1.jpg", i),
				fmt.Sprintf("/uploads/part_%d_2.jpg", i),
			},
		}
	}

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_, err := GenerateXMLPriceList(parts, sellerINNs, "https://backend-server.ru")
		if err != nil {
			b.Fatal(err)
		}
	}
}
