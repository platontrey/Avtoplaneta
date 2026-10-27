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

type parsedResult struct {
	Catalog domain.YMLCatalog
	Offers  []domain.YMLOffer
}

func parse(t *testing.T, data []byte) parsedResult {
	t.Helper()
	var catalog domain.YMLCatalog
	if err := xml.Unmarshal(data, &catalog); err != nil {
		t.Fatalf("прайс-лист не разбирается как YML каталог: %v", err)
	}
	return parsedResult{
		Catalog: catalog,
		Offers:  catalog.Shop.Offers,
	}
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
	if got := offers.Offers[0].GetParam("ИНН поставщика"); got != "700123456789" {
		t.Fatalf("ИНН продавца 7: ожидался 700123456789, получен %q", got)
	}
	if got := offers.Offers[1].GetParam("ИНН поставщика"); got != "" {
		t.Fatalf("ИНН продавца 9 неизвестен, но подставлен %q", got)
	}
}

// Пустое состояние площадка понимает как «Б/У»
func TestGenerateXMLPriceListDefaultsCondition(t *testing.T) {
	parts := []domain.Part{
		{Name: "Без состояния", Status: true, Quantity: 1, Photos: []string{"/p1.jpg"}},
		{Name: "С состоянием", Status: true, Quantity: 1, Condition: "Новая", Photos: []string{"/p2.jpg"}},
	}

	offers := parse(t, mustGenerate(t, parts))
	if got := offers.Offers[0].GetParam("Б/у или новый(-ая)"); got != "Б/У" {
		t.Fatalf("ожидалось состояние по умолчанию Б/У, получено %q", got)
	}
	if got := offers.Offers[1].GetParam("Б/у или новый(-ая)"); got != "Новая" {
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

// Проверка наличия XML заголовка UTF-8, DOCTYPE и корневого элемента yml_catalog
func TestGenerateXMLPriceListHeaderAndRoot(t *testing.T) {
	parts := []domain.Part{
		{Name: "Тестовая запчасть", Status: true, Quantity: 1, Price: 1500, Photos: []string{"/p.jpg"}},
	}
	data, err := service.GenerateXMLPriceList(parts, nil, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList: %v", err)
	}

	xmlStr := string(data)
	if !strings.HasPrefix(xmlStr, xml.Header) {
		t.Fatalf("отсутствует корректный заголовок xml.Header, получено: %s", xmlStr[:38])
	}
	if !strings.Contains(xmlStr, "<!DOCTYPE yml_catalog SYSTEM \"shops.dtd\">") {
		t.Fatalf("отсутствует обязательный DOCTYPE yml_catalog shops.dtd")
	}

	parsed := parse(t, data)
	if len(parsed.Offers) != 1 {
		t.Fatalf("ожидалось 1 предложение в <offers>, получено: %d", len(parsed.Offers))
	}
	if parsed.Catalog.Shop.Name != "АВТОПЛАНЕТА" {
		t.Errorf("Shop Name: got %q, want 'АВТОПЛАНЕТА'", parsed.Catalog.Shop.Name)
	}
	if len(parsed.Catalog.Shop.Currencies) == 0 || parsed.Catalog.Shop.Currencies[0].ID != "РУБ" {
		t.Errorf("Currencies: got %+v, want РУБ", parsed.Catalog.Shop.Currencies)
	}
	if len(parsed.Catalog.Shop.Categories) == 0 || parsed.Catalog.Shop.Categories[0].ID != "1" {
		t.Errorf("Categories: got %+v, want category 1", parsed.Catalog.Shop.Categories)
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

// Проверка маппинга всех полей запчасти на структуру YML Offer площадки Drom
func TestGenerateXMLPriceListAllFieldsMapping(t *testing.T) {
	sellerINNs := map[int64]string{
		42: "7017123456",
	}

	part := domain.Part{
		ID:             554433,
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
		Transmission:   "АКПП",
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
	if o.ID != "554433" {
		t.Errorf("ID: got %q, want 554433", o.ID)
	}
	wantName := "Фара правая светодиодная, Toyota, Camry, XV70, 2AR-FE, 2019, перед, право"
	if o.Name != wantName {
		t.Errorf("Name: got %q, want %q", o.Name, wantName)
	}
	if !o.Available {
		t.Errorf("Available: got %v, want true", o.Available)
	}
	if o.Description == nil || o.Description.Text != "Оригинал в отличном состоянии, все крепления целые" {
		t.Errorf("Description: got %+v", o.Description)
	}
	if o.Price != 18500.50 {
		t.Errorf("Price: got %v, want 18500.50", o.Price)
	}
	if o.CurrencyId != "РУБ" {
		t.Errorf("CurrencyId: got %q, want РУБ", o.CurrencyId)
	}
	if o.CategoryId != "1" {
		t.Errorf("CategoryId: got %q, want 1", o.CategoryId)
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
	if o.Store != "true" || o.Pickup != "true" || o.Delivery != "true" {
		t.Errorf("Store/Pickup/Delivery flags: %v / %v / %v", o.Store, o.Pickup, o.Delivery)
	}

	// Проверка параметров YML
	if got := o.GetParam("Марка автомобиля"); got != "Toyota" {
		t.Errorf("Марка автомобиля: got %q, want Toyota", got)
	}
	if got := o.GetParam("Модель автомобиля"); got != "Camry" {
		t.Errorf("Модель автомобиля: got %q, want Camry", got)
	}
	if got := o.GetParam("Марка кузова"); got != "XV70" {
		t.Errorf("Марка кузова: got %q, want XV70", got)
	}
	if got := o.GetParam("Марка двигателя"); got != "2AR-FE" {
		t.Errorf("Марка двигателя: got %q, want 2AR-FE", got)
	}
	if got := o.GetParam("Дата выпуска автомобиля"); got != "2019" {
		t.Errorf("Дата выпуска: got %q, want 2019", got)
	}
	if got := o.GetParam("Перед/зад"); got != "перед" {
		t.Errorf("Перед/зад: got %q, want перед", got)
	}
	if got := o.GetParam("Право/лево"); got != "право" {
		t.Errorf("Право/лево: got %q, want право", got)
	}
	if got := o.GetParam("Верх/низ"); got != "верх" {
		t.Errorf("Верх/низ: got %q, want верх", got)
	}
	if got := o.GetParam("Б/у или новый(-ая)"); got != "Б/у" {
		t.Errorf("Б/у или новый(-ая): got %q, want Б/у", got)
	}
	if got := o.GetParam("Трансмиссия"); got != "АКПП" {
		t.Errorf("Трансмиссия: got %q, want АКПП", got)
	}
	if got := o.GetParam("Расположение"); got != "Склад №2, ряд 4, полка 12" {
		t.Errorf("Расположение: got %q, want 'Склад №2, ряд 4, полка 12'", got)
	}
	if got := o.GetParam("ОЕМ код"); got != "81110-33D30" {
		t.Errorf("ОЕМ код: got %q, want 81110-33D30", got)
	}
	if got := o.GetParam("Код производителя"); got != "AP-778899" {
		t.Errorf("Код производителя: got %q, want AP-778899", got)
	}
}

// Проверка строгого запрета на выгрузку VIN и серийных номеров кузова
func TestGenerateXMLPriceListStrictlyExcludesVINAndFrameNumber(t *testing.T) {
	part := domain.Part{
		ID:             998877,
		Name:           "Крыло переднее левое",
		Status:         true,
		Quantity:       1,
		Price:          7500,
		Photos:         []string{"/wing.jpg"},
		Brand:          "Toyota",
		Model:          "Corolla",
		BodyBrand:      "NZE121", // Разрешенная марка кузова (поколение)
		SupplierCode:   "WING-121",
	}

	data, err := service.GenerateXMLPriceList([]domain.Part{part}, nil, "https://backend-server.ru")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList failed: %v", err)
	}

	xmlStr := string(data)

	// 1. Проверяем, что нет тега <vin>
	if strings.Contains(strings.ToLower(xmlStr), "<vin>") || strings.Contains(strings.ToLower(xmlStr), "</vin>") {
		t.Errorf("XML содержит запрещенный тег <vin>")
	}

	// 2. Проверяем, что нет параметров с именем VIN или Номер кузова
	if strings.Contains(xmlStr, `name="VIN"`) || strings.Contains(xmlStr, `name="vin"`) {
		t.Errorf("XML содержит запрещенный параметр VIN")
	}
	if strings.Contains(xmlStr, `name="Номер кузова"`) || strings.Contains(xmlStr, `name="номер кузова"`) {
		t.Errorf("XML содержит запрещенный параметр 'Номер кузова'")
	}

	// 3. Проверяем, что параметр 'Марка кузова' присутствует и содержит поколение 'NZE121'
	if !strings.Contains(xmlStr, `<param name="Марка кузова">NZE121</param>`) {
		t.Errorf("XML обязан содержать <param name=\"Марка кузова\">NZE121</param>, получен XML:\n%s", xmlStr)
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
	if o.Description == nil || o.Description.Text != "Состояние: 100% OK & проверено > без трещин; цена < 5000" {
		t.Errorf("спецсимволы в Description искажены: got %+v", o.Description)
	}
	if got := o.GetParam("ОЕМ код"); got != "81150-12<34>&56" {
		t.Errorf("спецсимволы в OEMCode искажены: got %q", got)
	}
}

// Тест генерирует файл sample_drom.xml с типовыми деталями в каноническом YML формате
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

// Проверка формирования параметров шин и дисков для классифайда Drom в формате YML
func TestGenerateXMLPriceListTiresAndWheels(t *testing.T) {
	tirePart := domain.Part{
		ID:           328570,
		Name:         "Шина",
		Quantity:     4,
		Price:        9500,
		Status:       true,
		Photos:       []string{"/uploads/tires/michelin.jpg"},
		Season:       "Зимние",
		Diameter:     "17",
		Width:        "235",
		Profile:      "65",
		TireQuantity: "4",
		TireModel:    "Michelin X-Ice North",
	}

	wheelPart := domain.Part{
		ID:                 336214,
		Name:               "Диск литой",
		Quantity:           4,
		Price:              6000,
		Status:             true,
		Photos:             []string{"/uploads/wheels/bbs.jpg"},
		Diameter:           "16",
		Width:              "7.5",
		Drilling:           "5/120",
		Offset:             "20",
		CenterHoleDiameter: "72",
	}

	data, err := service.GenerateXMLPriceList([]domain.Part{tirePart, wheelPart}, nil, "https://backend-server.ru")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList failed: %v", err)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 2 {
		t.Fatalf("expected 2 offers, got %d", len(offers.Offers))
	}

	// Проверка шины
	tireOffer := offers.Offers[0]
	if tireOffer.Name != "Шина Michelin X-Ice North 235/65 R17" {
		t.Errorf("tire Name: got %q, want %q", tireOffer.Name, "Шина Michelin X-Ice North 235/65 R17")
	}
	if got := tireOffer.GetParam("Ширина"); got != "235" {
		t.Errorf("tire Ширина: got %q, want 235", got)
	}
	if got := tireOffer.GetParam("Профиль"); got != "65" {
		t.Errorf("tire Профиль: got %q, want 65", got)
	}
	if got := tireOffer.GetParam("Диаметр"); got != "17" {
		t.Errorf("tire Диаметр: got %q, want 17", got)
	}
	if got := tireOffer.GetParam("Сезон"); got != "Зимняя" {
		t.Errorf("tire Сезон: got %q, want Зимняя", got)
	}
	if got := tireOffer.GetParam("Модель шины"); got != "Michelin X-Ice North" {
		t.Errorf("tire Модель шины: got %q, want Michelin X-Ice North", got)
	}
	if got := tireOffer.GetParam("Количество"); got != "4" {
		t.Errorf("tire Количество: got %q, want 4", got)
	}
	if got := tireOffer.GetParam("Шипы"); got != "Нет" {
		t.Errorf("tire Шипы: got %q, want Нет", got)
	}

	// Проверка диска
	wheelOffer := offers.Offers[1]
	if wheelOffer.Name != "Диск литой R16 5x120" {
		t.Errorf("wheel Name: got %q, want %q", wheelOffer.Name, "Диск литой R16 5x120")
	}
	if got := wheelOffer.GetParam("Ширина"); got != "7.5" {
		t.Errorf("wheel Ширина: got %q, want 7.5", got)
	}
	if got := wheelOffer.GetParam("Диаметр"); got != "16" {
		t.Errorf("wheel Диаметр: got %q, want 16", got)
	}
	if got := wheelOffer.GetParam("Вылет"); got != "20" {
		t.Errorf("wheel Вылет: got %q, want 20", got)
	}
	if got := wheelOffer.GetParam("Сверловка"); got != "5x120" {
		t.Errorf("wheel Сверловка: got %q, want 5x120", got)
	}
	if got := wheelOffer.GetParam("Диаметр ЦО"); got != "72" {
		t.Errorf("wheel Диаметр ЦО: got %q, want 72", got)
	}
	if got := wheelOffer.GetParam("Количество"); got != "4" {
		t.Errorf("wheel Количество: got %q, want 4", got)
	}
}

// Проверка исправления опечаток лемматизатора Drom (no_lemma)
func TestGenerateXMLPriceListTypoCorrection(t *testing.T) {
	parts := []domain.Part{
		{
			ID:       1,
			Name:     "Фиксатор ремня безопастности",
			Quantity: 1,
			Price:    500,
			Status:   true,
			Photos:   []string{"/p1.jpg"},
		},
		{
			ID:       2,
			Name:     "Стеклоподьемник передний правый",
			Quantity: 1,
			Price:    1500,
			Status:   true,
			Photos:   []string{"/p2.jpg"},
		},
	}

	data, err := service.GenerateXMLPriceList(parts, nil, "")
	if err != nil {
		t.Fatalf("GenerateXMLPriceList failed: %v", err)
	}

	offers := parse(t, data)
	if len(offers.Offers) != 2 {
		t.Fatalf("expected 2 offers, got %d", len(offers.Offers))
	}

	if offers.Offers[0].Name != "Фиксатор ремня безопасности" {
		t.Errorf("Name 0: got %q, want %q", offers.Offers[0].Name, "Фиксатор ремня безопасности")
	}
	if offers.Offers[1].Name != "Стеклоподъемник передний правый" {
		t.Errorf("Name 1: got %q, want %q", offers.Offers[1].Name, "Стеклоподъемник передний правый")
	}
}
