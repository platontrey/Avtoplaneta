package main

import (
	"encoding/xml"
	"os"
	"strings"
	"testing"
	"time"

	"export-service/internal/domain"
	"export-service/internal/service"
)

func parse(t *testing.T, data []byte) domain.Offers {
	t.Helper()
	var offers domain.Offers
	if err := xml.Unmarshal(data, &offers); err != nil {
		t.Fatalf("прайс-лист не разбирается как XML: %v", err)
	}
	return offers
}

// Запчасть, помеченную на удаление, площадка видеть не должна
func TestGenerateXMLPriceListSkipsPartsMarkedForDeletion(t *testing.T) {
	deleteAt := time.Now().Add(24 * time.Hour)
	parts := []domain.Part{
		{Name: "Живая", Status: true, Quantity: 1, Photos: []string{"/live.jpg"}},
		{Name: "Удаляется", Status: true, Quantity: 1, Photos: []string{"/del.jpg"}, ToDeleteAt: &deleteAt},
	}

	data, err := service.GenerateXMLPriceList(parts, nil, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 1 {
		t.Fatalf("ожидалось 1 предложение, получено %d", len(offers.Offers))
	}
	if offers.Offers[0].Name != "Живая" {
		t.Fatalf("в выгрузку попала не та запчасть: %q", offers.Offers[0].Name)
	}
}

// Запчасти без фото не должны передаваться на Drom
func TestGenerateXMLPriceListSkipsPartsWithoutPhotos(t *testing.T) {
	parts := []domain.Part{
		{Name: "С фото", Status: true, Quantity: 1, Photos: []string{"/photo.jpg"}},
		{Name: "Без фото (пустой слайс)", Status: true, Quantity: 1, Photos: []string{}},
		{Name: "Без фото (пробелы)", Status: true, Quantity: 1, Photos: []string{"   "}},
		{Name: "Без фото (nil)", Status: true, Quantity: 1, Photos: nil},
	}

	data, err := service.GenerateXMLPriceList(parts, nil, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 1 {
		t.Fatalf("ожидалось ровно 1 предложение с фото, получено %d", len(offers.Offers))
	}
	if offers.Offers[0].Name != "С фото" {
		t.Errorf("got %q, want 'С фото'", offers.Offers[0].Name)
	}
}

// Запчасти, у которых количество меньше 1 (quantity < 1), не должны передаваться
func TestGenerateXMLPriceListSkipsQuantityLessThanOne(t *testing.T) {
	parts := []domain.Part{
		{Name: "В наличии (1 шт)", Status: true, Quantity: 1, Photos: []string{"/1.jpg"}},
		{Name: "В наличии (5 шт)", Status: true, Quantity: 5, Photos: []string{"/5.jpg"}},
		{Name: "Нулевой остаток (0 шт)", Status: true, Quantity: 0, Photos: []string{"/0.jpg"}},
		{Name: "Проданная (-1 шт)", Status: true, Quantity: -1, Photos: []string{"/m1.jpg"}},
	}

	data, err := service.GenerateXMLPriceList(parts, nil, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 2 {
		t.Fatalf("ожидалось 2 предложения (с quantity >= 1), получено %d", len(offers.Offers))
	}
	if offers.Offers[0].Name != "В наличии (1 шт)" || offers.Offers[1].Name != "В наличии (5 шт)" {
		t.Errorf("неверный состав офферов: %+v", offers.Offers)
	}
}

// Запчасти с ценой 0 должны передаваться с явным тегом <price>0</price>
func TestGenerateXMLPriceListTransfersZeroPrice(t *testing.T) {
	parts := []domain.Part{
		{Name: "Бесплатная/без цены", Status: true, Quantity: 1, Price: 0.0, Photos: []string{"/free.jpg"}},
	}

	data, err := service.GenerateXMLPriceList(parts, nil, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	xmlStr := string(data)
	if !strings.Contains(xmlStr, "<price>0</price>") {
		t.Fatalf("в XML отсутствует обязательный тег <price>0</price>, XML:\n%s", xmlStr)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 1 {
		t.Fatalf("ожидался 1 оффер, получено %d", len(offers.Offers))
	}
	if offers.Offers[0].Price != 0 {
		t.Errorf("ожидалась цена 0, получено %v", offers.Offers[0].Price)
	}
}

// ИНН берётся из справочника auth-service по идентификатору продавца
func TestGenerateXMLPriceListFillsSupplierINN(t *testing.T) {
	parts := []domain.Part{
		{Name: "С реквизитами", SellerID: 7, Status: true, Quantity: 1, Photos: []string{"/p1.jpg"}},
		{Name: "Без реквизитов", SellerID: 9, Status: true, Quantity: 1, Photos: []string{"/p2.jpg"}},
	}

	data, err := service.GenerateXMLPriceList(parts, map[int64]string{7: "700123456789"}, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	offers := parse(t, data)
	if got := offers.Offers[0].SupplierInn; got != "700123456789" {
		t.Fatalf("ИНН продавца 7: ожидался 700123456789, получен %q", got)
	}
	if got := offers.Offers[1].SupplierInn; got != "" {
		t.Fatalf("ИНН продавца 9 неизвестен, но подставлен %q", got)
	}
}

// Пустое состояние площадка понимает как «новая», поэтому дефолт «Б/у»
func TestGenerateXMLPriceListDefaultsCondition(t *testing.T) {
	parts := []domain.Part{
		{Name: "Без состояния", Status: true, Quantity: 1, Photos: []string{"/p1.jpg"}},
		{Name: "С состоянием", Status: true, Quantity: 1, Condition: "Новая", Photos: []string{"/p2.jpg"}},
	}

	offers := parse(t, mustGenerate(t, parts))
	if got := offers.Offers[0].Condition; got != "Б/у" {
		t.Fatalf("ожидалось состояние по умолчанию Б/у, получено %q", got)
	}
	if got := offers.Offers[1].Condition; got != "Новая" {
		t.Fatalf("указанное состояние затёрто: %q", got)
	}
}

// Все фотографии детали попадают в выгрузку в виде отдельных тегов <picture>,
// а относительные пути преобразуются в абсолютные по baseImageURL.
func TestGenerateXMLPriceListMakesAbsolutePictureURL(t *testing.T) {
	parts := []domain.Part{
		{Name: "С фото", Status: true, Quantity: 1, Photos: []string{"/uploads/12_1700000000.jpg", "/uploads/12_1700000001.jpg"}},
		{Name: "Без фото", Status: true, Quantity: 1, Photos: nil},
	}

	base := "https://custom-domain.ru"
	data, err := service.GenerateXMLPriceList(parts, nil, base)
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 1 {
		t.Fatalf("ожидалась ровно 1 деталь с фото в выгрузке, получено %d", len(offers.Offers))
	}
	want1 := "https://custom-domain.ru/uploads/12_1700000000.jpg"
	want2 := "https://custom-domain.ru/uploads/12_1700000001.jpg"
	if got := offers.Offers[0].Pictures[0]; got != want1 {
		t.Fatalf("фото 1: ожидался адрес %q, получен %q", want1, got)
	}
	if got := offers.Offers[0].Pictures[1]; got != want2 {
		t.Fatalf("фото 2: ожидался адрес %q, получен %q", want2, got)
	}
}

func mustGenerate(t *testing.T, parts []domain.Part) []byte {
	t.Helper()
	data, err := service.GenerateXMLPriceList(parts, nil, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}
	return data
}

// Проверка наличия XML заголовка UTF-8 и корневого элемента offers
func TestGenerateXMLPriceListHeaderAndRoot(t *testing.T) {
	parts := []domain.Part{
		{Name: "Тестовая запчасть", Status: true, Quantity: 1, Price: 1500, Photos: []string{"/p.jpg"}},
	}
	data, err := service.GenerateXMLPriceList(parts, nil, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	xmlStr := string(data)
	if len(xmlStr) < len(xml.Header) || xmlStr[:len(xml.Header)] != xml.Header {
		t.Fatalf("отсутствует корректный заголовок xml.Header, получено: %s", xmlStr[:38])
	}

	offers := parse(t, data)
	if len(offers.Offers) != 1 {
		t.Fatalf("ожидалось 1 предложение в <offers>, получено: %d", len(offers.Offers))
	}
}

// Пустой склад формирует корректный XML без предложений
func TestGenerateXMLPriceListEmpty(t *testing.T) {
	data, err := service.GenerateXMLPriceList([]domain.Part{}, nil, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 0 {
		t.Fatalf("ожидался пустой список offers, получено: %d", len(offers.Offers))
	}
}

// Проверка маппинга всех полей запчасти на структуру Offer площадки Drom
func TestGenerateXMLPriceListAllFieldsMapping(t *testing.T) {
	sellerINNs := map[int64]string{
		42: "7017123456",
	}

	part := domain.Part{
		Name:           "Фара правая светодиодная",
		Quantity:       3,
		Description:    "Оригинал в отличном состоянии, все крепления целые",
		Price:          18500.50,
		Salesman:       "Автопланета Север",
		Location:       "Склад №2, ряд 4, полка 12",
		Status:         true,
		Brand:          "Toyota",
		Model:          "Camry",
		Photos:         []string{"/uploads/parts/camry_light_1.jpg", "/uploads/parts/camry_light_2.jpg"},
		SellerID:       42,
		BodyBrand:      "XV70",
		EngineBrand:    "2AR-FE",
		CarReleaseDate: "2019",
		FrontRear:      "перед",
		LeftRight:      "право",
		TopBottom:      "верх",
		Manufacturer:   "Koito",
		OEMCode:        "81110-33D30",
		Color:          "Черный",
		Condition:      "Б/у",
		SupplierCode:   "AP-778899",
	}

	base := "https://backend-server.ru"
	data, err := service.GenerateXMLPriceList([]domain.Part{part}, sellerINNs, base)
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 1 {
		t.Fatalf("ожидалось 1 предложение, получено %d", len(offers.Offers))
	}

	o := offers.Offers[0]
	if o.Name != "Фара правая светодиодная" {
		t.Errorf("Name: got %q, want %q", o.Name, "Фара правая светодиодная")
	}
	if !o.Available {
		t.Errorf("Available: got %v, want true", o.Available)
	}
	if o.Description != "Оригинал в отличном состоянии, все крепления целые" {
		t.Errorf("Description: got %q, want %q", o.Description, "Оригинал в отличном состоянии, все крепления целые")
	}
	if o.Price != 18500.50 {
		t.Errorf("Price: got %v, want 18500.50", o.Price)
	}
	if o.CurrencyId != "RUR" {
		t.Errorf("CurrencyId: got %q, want RUR", o.CurrencyId)
	}
	if len(o.Pictures) != 2 {
		t.Fatalf("Pictures: ожидалось 2 фото, получено %d", len(o.Pictures))
	}
	if o.Pictures[0] != "https://backend-server.ru/uploads/parts/camry_light_1.jpg" {
		t.Errorf("Pictures[0]: got %q, want https://backend-server.ru/uploads/parts/camry_light_1.jpg", o.Pictures[0])
	}
	if o.Pictures[1] != "https://backend-server.ru/uploads/parts/camry_light_2.jpg" {
		t.Errorf("Pictures[1]: got %q, want https://backend-server.ru/uploads/parts/camry_light_2.jpg", o.Pictures[1])
	}
	if o.Store != "Автопланета Север" {
		t.Errorf("Store: got %q, want %q", o.Store, "Автопланета Север")
	}
	if o.Quantity != 3 {
		t.Errorf("Quantity: got %d, want 3", o.Quantity)
	}
	if o.OemNumber != "81110-33D30" {
		t.Errorf("OemNumber: got %q, want %q", o.OemNumber, "81110-33D30")
	}
	if o.Manufacturer != "Koito" {
		t.Errorf("Manufacturer: got %q, want %q", o.Manufacturer, "Koito")
	}
	if o.Ordercode != "AP-778899" {
		t.Errorf("Ordercode: got %q, want %q", o.Ordercode, "AP-778899")
	}
	if o.Condition != "Б/у" {
		t.Errorf("Condition: got %q, want %q", o.Condition, "Б/у")
	}
	if o.Brandcars != "Toyota" {
		t.Errorf("Brandcars: got %q, want %q", o.Brandcars, "Toyota")
	}
	if o.Modelcars != "Camry" {
		t.Errorf("Modelcars: got %q, want %q", o.Modelcars, "Camry")
	}
	if o.Bodycars != "XV70" {
		t.Errorf("Bodycars: got %q, want %q", o.Bodycars, "XV70")
	}
	if o.Engine != "2AR-FE" {
		t.Errorf("Engine: got %q, want %q", o.Engine, "2AR-FE")
	}
	if o.Year != "2019" {
		t.Errorf("Year: got %q, want %q", o.Year, "2019")
	}
	if o.Lr != "право" {
		t.Errorf("Lr: got %q, want %q", o.Lr, "право")
	}
	if o.Fr != "перед" {
		t.Errorf("Fr: got %q, want %q", o.Fr, "перед")
	}
	if o.Ud != "верх" {
		t.Errorf("Ud: got %q, want %q", o.Ud, "верх")
	}
	if o.Color != "Черный" {
		t.Errorf("Color: got %q, want %q", o.Color, "Черный")
	}
	if o.Supplier != "Автопланета Север" {
		t.Errorf("Supplier: got %q, want %q", o.Supplier, "Автопланета Север")
	}
	if o.SupplierInn != "7017123456" {
		t.Errorf("SupplierInn: got %q, want %q", o.SupplierInn, "7017123456")
	}
	if o.Sklad != "Склад №2, ряд 4, полка 12" {
		t.Errorf("Sklad: got %q, want %q", o.Sklad, "Склад №2, ряд 4, полка 12")
	}
	if o.SupplierArt != "AP-778899" {
		t.Errorf("SupplierArt: got %q, want %q", o.SupplierArt, "AP-778899")
	}
}

// Проверка корректного экранирования спецсимволов XML в тексте
func TestGenerateXMLPriceListEscaping(t *testing.T) {
	parts := []domain.Part{
		{
			Name:        "Фара левая <Хрусталь> & \"Линза\" 2.0",
			Description: "Состояние: 100% OK & проверено > без трещин; цена < 5000",
			OEMCode:     "81150-12<34>&56",
			Status:      true,
			Quantity:    1,
			Photos:      []string{"/p.jpg"},
		},
	}

	data, err := service.GenerateXMLPriceList(parts, nil, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList с ошибкой: %v", err)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 1 {
		t.Fatalf("ожидалось 1 предложение, получено %d", len(offers.Offers))
	}

	o := offers.Offers[0]
	if o.Name != "Фара левая <Хрусталь> & \"Линза\" 2.0" {
		t.Errorf("спецсимволы в Name искажены: got %q", o.Name)
	}
	if o.Description != "Состояние: 100% OK & проверено > без трещин; цена < 5000" {
		t.Errorf("спецсимволы в Description искажены: got %q", o.Description)
	}
	if o.OemNumber != "81150-12<34>&56" {
		t.Errorf("спецсимволы в OEMCode искажены: got %q", o.OemNumber)
	}
}

// Тест генерирует файл sample_drom.xml с типовыми деталями
func TestGenerateSampleXMLFile(t *testing.T) {
	sellerINNs := map[int64]string{
		10: "7017998877",
		20: "7017665544",
	}

	parts := []domain.Part{
		{
			Name:           "Фара правая LED",
			Quantity:       2,
			Description:    "Оригинал, идеальное состояние. Без пробега по РФ.",
			Price:          25000,
			Salesman:       "Автопланета Томск",
			Location:       "Склад №1, Секция А",
			Status:         true,
			Brand:          "Toyota",
			Model:          "Camry",
			Photos:         []string{"/uploads/parts/camry_light_front.jpg", "/uploads/parts/camry_light_back.jpg"},
			SellerID:       10,
			BodyBrand:      "XV70",
			EngineBrand:    "2AR-FE",
			CarReleaseDate: "2020",
			FrontRear:      "перед",
			LeftRight:      "право",
			TopBottom:      "верх",
			Manufacturer:   "Koito",
			OEMCode:        "81110-33D30",
			Color:          "Черный",
			Condition:      "Б/у",
			SupplierCode:   "AP-1001",
		},
		{
			Name:           "Двигатель 1NZ-FE",
			Quantity:       1,
			Description:    "Контрактный ДВС в сборе с навесным оборудованием. Пробег 64 000 км по Японии. ГТД в комплекте.",
			Price:          65000,
			Salesman:       "Автопланета Север",
			Location:       "Склад агрегатов №3",
			Status:         true,
			Brand:          "Toyota",
			Model:          "Corolla",
			Photos:         []string{"/uploads/parts/engine_1nz.jpg"},
			SellerID:       20,
			BodyBrand:      "NZE121",
			EngineBrand:    "1NZ-FE",
			CarReleaseDate: "2005",
			Manufacturer:   "Toyota",
			OEMCode:        "19000-21010",
			Condition:      "Б/у",
			SupplierCode:   "ENG-1NZ-092",
		},
		{
			Name:           "Акт дефектовки: Бампер передний (трещина снизу)",
			Quantity:       0, // quantity = 0 не должно попасть в XML
			Description:    "Дефектная позиция: трещина 5 см на нижней губе под пайку. Крепления целые.",
			Price:          7000,
			Salesman:       "Автопланета Томск",
			Location:       "Зона дефектовки",
			Status:         true,
			Brand:          "Honda",
			Model:          "Civic",
			Photos:         []string{"/uploads/parts/civic_bumper.jpg"},
			SellerID:       10,
			BodyBrand:      "FK7",
			CarReleaseDate: "2018",
			FrontRear:      "перед",
			Condition:      "Б/у",
			SupplierCode:   "DEF-CIVIC-01",
		},
	}

	data, err := service.GenerateXMLPriceList(parts, sellerINNs, "")
	if err != nil {
		t.Fatalf("Ошибка генерации XML: %v", err)
	}

	_ = os.MkdirAll("testdata", 0o755)
	outPath := "testdata/sample_drom.xml"
	if err := os.WriteFile(outPath, data, 0o644); err != nil {
		t.Fatalf("Не удалось записать sample_drom.xml: %v", err)
	}

	offers := parse(t, data)
	// Третья позиция (quantity = 0) отфильтрована, должно быть ровно 2 позиции
	if len(offers.Offers) != 2 {
		t.Fatalf("Ожидалось 2 позиции (quantity >= 1), получено %d", len(offers.Offers))
	}
	if len(offers.Offers[0].Pictures) != 2 {
		t.Fatalf("Ожидалось 2 фото у первой детали, получено %d", len(offers.Offers[0].Pictures))
	}
}
