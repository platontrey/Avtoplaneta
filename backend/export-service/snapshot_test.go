package main

import (
	"encoding/xml"
	"os"
	"path/filepath"
	"testing"

	"export-service/internal/domain"
	"export-service/internal/repository"
)

func TestSnapshotStore_UpsertAndWriteFullXML(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_catalog.db")

	store, err := repository.NewBoltSnapshotRepository(dbPath)
	if err != nil {
		t.Fatalf("NewBoltSnapshotRepository: %v", err)
	}
	defer store.Close()

	parts := []domain.Part{
		{
			ID:           1,
			Name:         "Амортизатор передний",
			Status:       true,
			Quantity:     2,
			Price:        3500.0,
			SellerID:     10,
			Brand:        "Toyota",
			Model:        "Corolla",
			Condition:    "Новый",
			SupplierCode: "AM-100",
			Photos:       []string{"/photo1.jpg"},
		},
		{
			ID:           2,
			Name:         "Фара левая LED",
			Status:       true,
			Quantity:     1,
			Price:        12000.0,
			SellerID:     10,
			Brand:        "Toyota",
			Model:        "Corolla",
			Condition:    "Б/у",
			SupplierCode: "HL-200",
			Photos:       []string{"/photo2.jpg"},
		},
	}
	sellerINNs := map[int64]string{10: "7701123456"}
	baseURL := "https://backend-server.ru"

	// 1. Первичная загрузка BulkLoad
	if err := store.BulkLoad(parts, sellerINNs, baseURL, "v1"); err != nil {
		t.Fatalf("BulkLoad: %v", err)
	}

	if store.Count() != 2 {
		t.Fatalf("ожидалось 2 оффера в базе, получено %d", store.Count())
	}
	if store.GetVersion() != "v1" {
		t.Fatalf("ожидалась версия v1, получено %q", store.GetVersion())
	}

	// 2. Сборка полного XML на диск
	xmlPath := filepath.Join(tmpDir, "pricelist.xml")
	count, err := store.WriteFullXML(xmlPath)
	if err != nil {
		t.Fatalf("WriteFullXML: %v", err)
	}
	if count != 2 {
		t.Fatalf("ожидалось 2 оффера в XML, получено %d", count)
	}

	xmlBytes, err := os.ReadFile(xmlPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	var parsed domain.YMLCatalog
	if err := xml.Unmarshal(xmlBytes, &parsed); err != nil {
		t.Fatalf("ошибка разбора сгенерированного XML: %v", err)
	}
	if len(parsed.Shop.Offers) != 2 {
		t.Fatalf("ожидалось 2 предложения в XML, получено %d", len(parsed.Shop.Offers))
	}
	if parsed.Shop.Offers[0].Name != "Амортизатор передний" || parsed.Shop.Offers[0].Price != 3500.0 {
		t.Errorf("неверные данные первого оффера: %+v", parsed.Shop.Offers[0])
	}
	if parsed.Shop.Offers[1].Name != "Фара левая LED" || parsed.Shop.Offers[1].Price != 12000.0 {
		t.Errorf("неверные данные второго оффера: %+v", parsed.Shop.Offers[1])
	}

	// 3. Точечное обновление (SaveOffer) детали №2: цена изменилась с 12000 на 14500
	updatedPart2 := parts[1]
	updatedPart2.Price = 14500.0
	if err := store.SaveOffer(updatedPart2, sellerINNs[10], baseURL); err != nil {
		t.Fatalf("SaveOffer: %v", err)
	}

	// Точечное добавление детали №3
	part3 := domain.Part{
		ID:           3,
		Name:         "Колодки тормозные",
		Status:       true,
		Quantity:     4,
		Price:        1800.0,
		SellerID:     10,
		Condition:    "Новый",
		SupplierCode: "BP-300",
		Photos:       []string{"/photo3.jpg"},
	}
	if err := store.SaveOffer(part3, sellerINNs[10], baseURL); err != nil {
		t.Fatalf("SaveOffer part3: %v", err)
	}

	if store.Count() != 3 {
		t.Fatalf("ожидалось 3 оффера в базе после добавления, получено %d", store.Count())
	}

	// 4. Повторная генерация полного XML из обновленного слепка
	count, err = store.WriteFullXML(xmlPath)
	if err != nil {
		t.Fatalf("WriteFullXML после обновления: %v", err)
	}
	if count != 3 {
		t.Fatalf("ожидалось 3 оффера в XML, получено %d", count)
	}

	xmlBytes, _ = os.ReadFile(xmlPath)
	var parsed2 domain.YMLCatalog
	_ = xml.Unmarshal(xmlBytes, &parsed2)
	if len(parsed2.Shop.Offers) != 3 {
		t.Fatalf("ожидалось 3 оффера, получено %d", len(parsed2.Shop.Offers))
	}
	if parsed2.Shop.Offers[1].Price != 14500.0 {
		t.Errorf("цена детали 2 не обновилась: got %v, want 14500.0", parsed2.Shop.Offers[1].Price)
	}

	// 5. Точечное удаление детали №1 (DeleteOffer)
	if err := store.DeleteOffer(1); err != nil {
		t.Fatalf("DeleteOffer: %v", err)
	}
	if store.Count() != 2 {
		t.Fatalf("ожидалось 2 оффера после удаления, получено %d", store.Count())
	}

	count, _ = store.WriteFullXML(xmlPath)
	if count != 2 {
		t.Fatalf("ожидалось 2 оффера в финальном XML, получено %d", count)
	}

	xmlBytes, _ = os.ReadFile(xmlPath)
	var parsed3 domain.YMLCatalog
	_ = xml.Unmarshal(xmlBytes, &parsed3)
	for _, o := range parsed3.Shop.Offers {
		if o.Name == "Амортизатор передний" {
			t.Fatal("удаленная деталь всё ещё присутствует в полном XML")
		}
	}
}

func TestSnapshotStore_PersistenceAcrossReopen(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "persist.db")

	store1, err := repository.NewBoltSnapshotRepository(dbPath)
	if err != nil {
		t.Fatalf("NewBoltSnapshotRepository: %v", err)
	}

	part := domain.Part{
		ID:        100,
		Name:      "Стойка стабилизатора",
		Status:    true,
		Quantity:  2,
		Price:     950.0,
		Condition: "Новый",
		Photos:    []string{"/photo100.jpg"},
	}
	_ = store1.SaveOffer(part, "1234567890", "https://backend-server.ru")
	_ = store1.SetVersion("version-xyz")
	_ = store1.Close()

	// Открываем заново (симуляция рестарта контейнера)
	store2, err := repository.NewBoltSnapshotRepository(dbPath)
	if err != nil {
		t.Fatalf("reopen NewBoltSnapshotRepository: %v", err)
	}
	defer store2.Close()

	if store2.Count() != 1 {
		t.Fatalf("после переоткрытия ожидали 1 оффер, получили %d", store2.Count())
	}
	if store2.GetVersion() != "version-xyz" {
		t.Fatalf("после переоткрытия ожидали версию version-xyz, получили %q", store2.GetVersion())
	}

	// Проверяем генерацию байт XML
	xmlData, count, err := store2.GenerateFullXMLBytes()
	if err != nil || count != 1 {
		t.Fatalf("GenerateFullXMLBytes: count=%d, err=%v", count, err)
	}
	var parsed domain.YMLCatalog
	if err := xml.Unmarshal(xmlData, &parsed); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if parsed.Shop.Offers[0].Name != "Стойка стабилизатора" {
		t.Errorf("got %q, want Стойка стабилизатора", parsed.Shop.Offers[0].Name)
	}
}
