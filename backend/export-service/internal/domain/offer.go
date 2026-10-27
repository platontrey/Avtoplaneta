package domain

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// YMLCatalog представляет корневой элемент выгрузки в формате Yandex Market Language (YML) для Drom.ru
type YMLCatalog struct {
	XMLName xml.Name `xml:"yml_catalog"`
	Date    string   `xml:"date,attr"`
	Shop    YMLShop  `xml:"shop"`
}

// YMLShop содержит реквизиты магазина, валюты, категории и офферы
type YMLShop struct {
	XMLName    xml.Name      `xml:"shop"`
	Name       string        `xml:"name"`
	Company    string        `xml:"company"`
	URL        string        `xml:"url"`
	Platform   string        `xml:"platform,omitempty"`
	Email      string        `xml:"email,omitempty"`
	Currencies []YMLCurrency `xml:"currencies>currency"`
	Categories []YMLCategory `xml:"categories>category"`
	Offers     []YMLOffer    `xml:"offers>offer"`
}

// YMLCurrency описывает поддерживаемую валюту каталога
type YMLCurrency struct {
	ID   string `xml:"id,attr"`
	Rate string `xml:"rate,attr"`
}

// YMLCategory описывает товарную категорию YML
type YMLCategory struct {
	ID       string `xml:"id,attr"`
	ParentID string `xml:"parentId,attr,omitempty"`
	Name     string `xml:",chardata"`
}

// YMLParam представляет произвольный параметр детали в формате YML: <param name="...">значение</param>
type YMLParam struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

// CDATA оборачивает строку в XML CDATA секцию
type CDATA struct {
	Text string `xml:",cdata"`
}

// YMLOffer представляет товарное предложение в формате YML каталога Drom
type YMLOffer struct {
	XMLName     xml.Name   `xml:"offer"`
	ID          string     `xml:"id,attr"`
	Available   bool       `xml:"available,attr"`
	Bid         string     `xml:"bid,attr,omitempty"`
	URL         string     `xml:"url,omitempty"`
	Price       float64    `xml:"price"`
	CurrencyId  string     `xml:"currencyId"`
	CategoryId  string     `xml:"categoryId"`
	Pictures    []string   `xml:"picture,omitempty"`
	Store       string     `xml:"store,omitempty"`
	Pickup      string     `xml:"pickup,omitempty"`
	Delivery    string     `xml:"delivery,omitempty"`
	Name        string     `xml:"name"`
	Description *CDATA     `xml:"description,omitempty"`
	SalesNotes  string     `xml:"sales_notes,omitempty"`
	Barcode     string     `xml:"barcode,omitempty"`
	Cpa         string     `xml:"cpa,omitempty"`
	Params      []YMLParam `xml:"param,omitempty"`
}

// Offer является псевдонимом для YMLOffer для полной совместимости со всеми модулями
type Offer = YMLOffer

// Offers представляет коллекцию предложений (для частичного/фрагментарного разбора в тестах)
type Offers struct {
	XMLName xml.Name   `xml:"offers"`
	Offers  []YMLOffer `xml:"offer"`
}

// GetParam возвращает значение параметра по его имени
func (o *YMLOffer) GetParam(name string) string {
	for _, p := range o.Params {
		if p.Name == name {
			return p.Value
		}
	}
	return ""
}

// IsEligibleForExport определяет критерии включения детали в XML выгрузку на Drom:
// 1. Деталь не помечена на удаление (to_delete_at == nil)
// 2. Деталь активна (status == true)
// 3. Количество не менее 1 (quantity >= 1) — детали с quantity < 1 не выгружаются
// 4. Есть хотя бы одна валидная фотография — детали без фото не выгружаются
func IsEligibleForExport(part Part) bool {
	if part.ToDeleteAt != nil || !part.Status || part.Quantity < 1 {
		return false
	}
	return HasValidPhotos(part.Photos)
}

// HasValidPhotos проверяет наличие хотя бы одной непустой фотографии
func HasValidPhotos(photos []string) bool {
	for _, p := range photos {
		if strings.TrimSpace(p) != "" {
			return true
		}
	}
	return false
}

// BuildOffer строит предложение Offer из Part и реквизитов продавца
func BuildOffer(part Part, sellerINN string, baseImageURL string) YMLOffer {
	return BuildYMLOffer(part, sellerINN, baseImageURL)
}

// BuildYMLOffer строит валидное YML товарное предложение, соответствующее эталонному формату Multi_YML_Export
func BuildYMLOffer(part Part, sellerINN string, baseImageURL string) YMLOffer {
	baseImageURL = strings.TrimRight(strings.TrimSpace(baseImageURL), "/")

	pictures := make([]string, 0, len(part.Photos))
	for _, photo := range part.Photos {
		trimmed := strings.TrimSpace(photo)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "http://") || strings.HasPrefix(trimmed, "https://") {
			pictures = append(pictures, trimmed)
		} else {
			photoPath := "/" + strings.TrimPrefix(trimmed, "/")
			if baseImageURL != "" {
				pictures = append(pictures, baseImageURL+photoPath)
			} else {
				pictures = append(pictures, photoPath)
			}
		}
	}

	condition := strings.TrimSpace(part.Condition)
	if condition == "" {
		condition = "Б/У"
	}

	partIDStr := strconv.FormatInt(part.ID, 10)
	offerURL := ""
	if baseImageURL != "" {
		offerURL = fmt.Sprintf("%s/parts/%d", baseImageURL, part.ID)
	} else {
		offerURL = fmt.Sprintf("http://backend-server.ru/parts/%d", part.ID)
	}

	var desc *CDATA
	descText := strings.TrimSpace(part.Description)
	if descText != "" {
		desc = &CDATA{Text: descText}
	} else {
		desc = &CDATA{Text: "В связи с изменением цены конечную стоимость товара узнавать по WhatsApp 89138538227"}
	}

	offer := YMLOffer{
		ID:          partIDStr,
		Available:   part.Status && part.Quantity >= 1,
		Bid:         "3",
		URL:         offerURL,
		Price:       part.Price,
		CurrencyId:  "РУБ",
		CategoryId:  "1",
		Pictures:    pictures,
		Store:       "true",
		Pickup:      "true",
		Delivery:    "true",
		Name:        normalizePartName(part),
		Description: desc,
		SalesNotes:  "Способ оплаты уточняйте по телефону",
		Barcode:     partIDStr,
		Cpa:         "1",
	}

	params := make([]YMLParam, 0, 16)
	addParam := func(name, val string) {
		val = strings.TrimSpace(val)
		if val != "" {
			params = append(params, YMLParam{Name: name, Value: val})
		}
	}

	// 1. Основные характеристики автомобиля
	addParam("Марка автомобиля", part.Brand)
	addParam("Модель автомобиля", part.Model)

	// ВАЖНО: Марка кузова (поколение/шасси, например S50, XV70) разрешена.
	// КАТЕГОРИЧЕСКИ ЗАПРЕЩЕНО выгружать VIN и серийный номер кузова!
	addParam("Марка кузова", part.BodyBrand)

	addParam("Марка двигателя", part.EngineBrand)
	addParam("Дата выпуска автомобиля", part.CarReleaseDate)
	addParam("Перед/зад", part.FrontRear)
	addParam("Право/лево", part.LeftRight)
	addParam("Верх/низ", part.TopBottom)
	addParam("Б/у или новый(-ая)", condition)
	addParam("Трансмиссия", part.Transmission)

	// 2. Расположение и поставщик
	loc := strings.TrimSpace(part.Location)
	if loc == "" {
		loc = "г. Томск"
	}
	addParam("Расположение", loc)

	if sellerINN != "" {
		addParam("ИНН поставщика", sellerINN)
	}

	// 3. Коды деталей
	addParam("ОЕМ код", part.OEMCode)
	mfgCode := strings.TrimSpace(part.ManufacturerCode)
	if mfgCode == "" {
		mfgCode = strings.TrimSpace(part.SupplierCode)
	}
	addParam("Код производителя", mfgCode)

	// 4. Дефекты и состояние
	addParam("Состояние/дефект", part.Defect)
	addParam("Процент износа", part.WearPercentage)

	// 5. Параметры шин и дисков
	enrichWheelAndTireYMLParams(&params, part)

	offer.Params = params
	return offer
}

func enrichWheelAndTireYMLParams(params *[]YMLParam, part Part) {
	isTire := isTirePart(part)
	isDisc := isDiscPart(part)

	addParam := func(name, val string) {
		val = strings.TrimSpace(val)
		if val != "" {
			*params = append(*params, YMLParam{Name: name, Value: val})
		}
	}

	// Диаметр
	if d := cleanDiameter(part.Diameter); d != "" {
		addParam("Диаметр", d)
	}

	// Ширина
	if part.Width != "" {
		addParam("Ширина", part.Width)
	}

	// Профиль
	if part.Profile != "" {
		addParam("Профиль", part.Profile)
	}

	// Сезон
	if isTire || part.Season != "" {
		seasonLower := strings.ToLower(strings.TrimSpace(part.Season))
		var season string
		switch {
		case strings.Contains(seasonLower, "зим"):
			season = "Зимняя"
		case strings.Contains(seasonLower, "лет"):
			season = "Летняя"
		case strings.Contains(seasonLower, "всесез"):
			season = "Всесезонная"
		case part.Season != "":
			season = strings.TrimSpace(part.Season)
		}
		addParam("Сезон", season)
	}

	// Модель шины
	addParam("Модель шины", part.TireModel)

	// Параметры дисков
	if isDisc || part.Drilling != "" || part.Offset != "" || part.CenterHoleDiameter != "" {
		pcd := strings.TrimSpace(part.Drilling)
		pcd = strings.ReplaceAll(pcd, "/", "x")
		pcd = strings.ReplaceAll(pcd, "*", "x")
		addParam("Сверловка", pcd)
		addParam("Вылет", strings.TrimSpace(part.Offset))
		addParam("Диаметр ЦО", strings.TrimSpace(part.CenterHoleDiameter))
	}

	// Количество в комплекте
	inSet := strings.TrimSpace(part.TireQuantity)
	if inSet == "" && (isTire || isDisc) && part.Quantity > 0 {
		inSet = strconv.Itoa(part.Quantity)
	}
	addParam("Количество", inSet)
}

// normalizePartName устраняет опечатки, которые ломают лемматизатор Drom (no_lemma),
// и делает заголовки для шин и дисков информативными.
func normalizePartName(part Part) string {
	name := strings.TrimSpace(part.Name)
	if name == "" {
		return ""
	}

	// 1. Исправление типичных опечаток в базе данных, из-за которых Drom ставит no_lemma:
	name = strings.ReplaceAll(name, "безопастност", "безопасност")
	name = strings.ReplaceAll(name, "Безопастност", "Безопасност")
	name = strings.ReplaceAll(name, "БЕЗОПАСТНОСТ", "БЕЗОПАСНОСТ")
	name = strings.ReplaceAll(name, "подьем", "подъем")
	name = strings.ReplaceAll(name, "Подьем", "Подъем")
	name = strings.ReplaceAll(name, "ПОДЬЕМ", "ПОДЪЕМ")
	name = strings.ReplaceAll(name, "подьём", "подъём")
	name = strings.ReplaceAll(name, "Подьём", "Подъём")
	name = strings.ReplaceAll(name, "ПОДЬЁМ", "ПОДЪЁМ")

	if strings.HasSuffix(name, "двери дверью") {
		name = strings.TrimSpace(strings.TrimSuffix(name, "дверью"))
	}

	// 2. Для шин с неинформативным заголовком ("Шина", "Шины", "Покрышка")
	lower := strings.ToLower(name)
	if lower == "шина" || lower == "шины" || lower == "покрышка" || lower == "покрышки" || lower == "резина" {
		var parts []string
		parts = append(parts, "Шина")
		if tm := strings.TrimSpace(part.TireModel); tm != "" {
			parts = append(parts, tm)
		}
		w := strings.TrimSpace(part.Width)
		p := strings.TrimSpace(part.Profile)
		d := cleanDiameter(part.Diameter)
		if w != "" && p != "" && d != "" {
			parts = append(parts, fmt.Sprintf("%s/%s R%s", w, p, d))
		} else if d != "" {
			parts = append(parts, "R"+d)
		}
		if len(parts) > 1 {
			return strings.Join(parts, " ")
		}
	}

	// 3. Для дисков с неинформативным заголовком ("Диск", "Диски", "Диск литой")
	if lower == "диск" || lower == "диски" || lower == "диск литой" || lower == "диски литые" {
		var parts []string
		if strings.Contains(lower, "лит") {
			parts = append(parts, "Диск литой")
		} else {
			parts = append(parts, "Диск")
		}
		if d := cleanDiameter(part.Diameter); d != "" {
			parts = append(parts, "R"+d)
		}
		if dr := strings.TrimSpace(part.Drilling); dr != "" {
			dr = strings.ReplaceAll(dr, "/", "x")
			parts = append(parts, dr)
		}
		if len(parts) > 1 {
			return strings.Join(parts, " ")
		}
	}

	return name
}

func cleanDiameter(d string) string {
	d = strings.TrimSpace(d)
	d = strings.TrimPrefix(d, "R")
	d = strings.TrimPrefix(d, "r")
	d = strings.TrimPrefix(d, "Р")
	d = strings.TrimPrefix(d, "р")
	return strings.TrimSpace(d)
}

func isTirePart(part Part) bool {
	if part.Season != "" || part.Profile != "" || part.TireModel != "" {
		return true
	}
	lower := strings.ToLower(part.Name)
	return strings.Contains(lower, "шина") || strings.Contains(lower, "покрышк") || strings.Contains(lower, "резина")
}

func isDiscPart(part Part) bool {
	if part.Drilling != "" || part.Offset != "" || part.CenterHoleDiameter != "" {
		return true
	}
	lower := strings.ToLower(part.Name)
	return strings.Contains(lower, "диск") || strings.Contains(lower, "колес")
}

// RenderOfferXML сериализует одиночный оффер в аккуратный XML-фрагмент
func RenderOfferXML(offer YMLOffer) ([]byte, error) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	enc.Indent("    ", "  ")
	if err := enc.Encode(offer); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}
