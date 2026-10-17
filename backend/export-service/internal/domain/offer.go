package domain

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"strconv"
	"strings"
)

// Offers представляет корневой элемент XML прайс-листа площадки Drom
type Offers struct {
	XMLName xml.Name `xml:"offers"`
	Offers  []Offer  `xml:"offer"`
}

// Offer представляет структуру товарного предложения в XML-формате Drom
type Offer struct {
	XMLName       xml.Name `xml:"offer"`
	Name          string   `xml:"name"`
	Available     bool     `xml:"available"`
	Description   string   `xml:"description,omitempty"`
	Price         float64  `xml:"price"`
	CurrencyId    string   `xml:"currencyId,omitempty"`
	Pictures      []string `xml:"picture,omitempty"`
	Store         string   `xml:"store,omitempty"`
	Quantity      int      `xml:"quantity,omitempty"`
	OemNumber     string   `xml:"oem_number,omitempty"`
	AnalogNumbers string   `xml:"analog_numbers,omitempty"`
	Manufacturer  string   `xml:"manufacturer,omitempty"`
	Ordercode     string   `xml:"ordercode,omitempty"`
	Condition     string   `xml:"condition,omitempty"`
	Brandcars     string   `xml:"brandcars,omitempty"`
	Modelcars     string   `xml:"modelcars,omitempty"`
	Bodycars      string   `xml:"bodycars,omitempty"`
	Engine        string   `xml:"engine,omitempty"`
	Year          string   `xml:"year,omitempty"`
	Lr            string   `xml:"lr,omitempty"`
	Fr            string   `xml:"fr,omitempty"`
	Ud            string   `xml:"ud,omitempty"`
	Color         string   `xml:"color,omitempty"`
	Supplier      string   `xml:"supplier,omitempty"`
	SupplierInn   string   `xml:"supplier_inn,omitempty"`
	Sklad         string   `xml:"sklad,omitempty"`
	SupplierArt   string   `xml:"supplier_art,omitempty"`

	// Параметры шин и дисков для Дром
	Marking          string `xml:"marking,omitempty"`
	InSet            string `xml:"inSet,omitempty"`
	InSetQuantity    string `xml:"in_set_quantity,omitempty"`
	WheelWidth       string `xml:"wheel_width,omitempty"`
	WheelDiameter    string `xml:"wheel_diameter,omitempty"`
	DiscET           string `xml:"disc_et,omitempty"`
	DiscPCD          string `xml:"disc_pcd,omitempty"`
	DiscHoleDiameter string `xml:"disc_hole_diameter,omitempty"`
	DiscDia          string `xml:"disc_dia,omitempty"`
	TireWidth        string `xml:"tire_width,omitempty"`
	TireSectionRatio string `xml:"tire_section_ratio,omitempty"`
	TireSpike        string `xml:"tire_spike,omitempty"`
	Season           string `xml:"season,omitempty"`
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
func BuildOffer(part Part, sellerINN string, baseImageURL string) Offer {
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
		condition = "Б/у"
	}

	offer := Offer{
		Name:          normalizePartName(part),
		Available:     part.Status && part.Quantity >= 0,
		Description:   strings.TrimSpace(part.Description),
		Price:         part.Price,
		CurrencyId:    "RUR",
		Pictures:      pictures,
		Store:         strings.TrimSpace(part.Salesman),
		Quantity:      part.Quantity,
		OemNumber:     strings.TrimSpace(part.OEMCode),
		AnalogNumbers: "",
		Manufacturer:  strings.TrimSpace(part.Manufacturer),
		Ordercode:     strings.TrimSpace(part.SupplierCode),
		Condition:     condition,
		Brandcars:     strings.TrimSpace(part.Brand),
		Modelcars:     strings.TrimSpace(part.Model),
		Bodycars:      strings.TrimSpace(part.BodyBrand),
		Engine:        strings.TrimSpace(part.EngineBrand),
		Year:          strings.TrimSpace(part.CarReleaseDate),
		Lr:            strings.TrimSpace(part.LeftRight),
		Fr:            strings.TrimSpace(part.FrontRear),
		Ud:            strings.TrimSpace(part.TopBottom),
		Color:         strings.TrimSpace(part.Color),
		Supplier:      "Автопланета",
		SupplierInn:   sellerINN,
		Sklad:         strings.TrimSpace(part.Location),
		SupplierArt:   strings.TrimSpace(part.SupplierCode),
	}

	enrichWheelAndTireParams(&offer, part)

	return offer
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

func enrichWheelAndTireParams(offer *Offer, part Part) {
	isTire := isTirePart(part)
	isDisc := isDiscPart(part)

	// Комплектность (inSet / in_set_quantity)
	inSet := strings.TrimSpace(part.TireQuantity)
	if inSet == "" && (isTire || isDisc) && part.Quantity > 0 {
		inSet = strconv.Itoa(part.Quantity)
	}
	if inSet != "" {
		offer.InSet = inSet
		offer.InSetQuantity = inSet
	}

	// Параметры дисков
	if isDisc || part.Drilling != "" || part.Offset != "" || part.CenterHoleDiameter != "" {
		offer.WheelDiameter = cleanDiameter(part.Diameter)
		offer.WheelWidth = strings.TrimSpace(part.Width)

		pcd := strings.TrimSpace(part.Drilling)
		pcd = strings.ReplaceAll(pcd, "/", "x")
		pcd = strings.ReplaceAll(pcd, "*", "x")
		offer.DiscPCD = pcd
		offer.DiscET = strings.TrimSpace(part.Offset)

		dia := strings.TrimSpace(part.CenterHoleDiameter)
		offer.DiscHoleDiameter = dia
		offer.DiscDia = dia

		var markingParts []string
		if offer.WheelWidth != "" {
			w := offer.WheelWidth
			if !strings.HasSuffix(strings.ToUpper(w), "J") {
				w += "J"
			}
			markingParts = append(markingParts, w)
		}
		if offer.WheelDiameter != "" {
			markingParts = append(markingParts, "R"+offer.WheelDiameter)
		}
		if offer.DiscET != "" {
			et := offer.DiscET
			if !strings.HasPrefix(strings.ToUpper(et), "ET") {
				et = "ET" + et
			}
			markingParts = append(markingParts, et)
		}
		if offer.DiscPCD != "" {
			markingParts = append(markingParts, offer.DiscPCD)
		}
		if offer.DiscDia != "" {
			diaVal := offer.DiscDia
			if !strings.HasPrefix(strings.ToUpper(diaVal), "DIA") {
				diaVal = "DIA" + diaVal
			}
			markingParts = append(markingParts, diaVal)
		}
		if len(markingParts) > 0 {
			offer.Marking = strings.Join(markingParts, " ")
		}
	}

	// Параметры шин
	if isTire || part.Season != "" || part.Profile != "" {
		offer.TireWidth = strings.TrimSpace(part.Width)
		offer.TireSectionRatio = strings.TrimSpace(part.Profile)
		if offer.WheelDiameter == "" {
			offer.WheelDiameter = cleanDiameter(part.Diameter)
		}

		seasonLower := strings.ToLower(strings.TrimSpace(part.Season))
		switch {
		case strings.Contains(seasonLower, "зим"):
			offer.Season = "Зимняя"
		case strings.Contains(seasonLower, "лет"):
			offer.Season = "Летняя"
		case strings.Contains(seasonLower, "всесез"):
			offer.Season = "Всесезонная"
		case part.Season != "":
			offer.Season = strings.TrimSpace(part.Season)
		}

		text := strings.ToLower(part.Name + " " + part.Description + " " + part.TireModel)
		if strings.Contains(text, "нешип") || strings.Contains(text, "липуч") || strings.Contains(text, "friction") {
			offer.TireSpike = "Нешипованные"
		} else if strings.Contains(text, "шип") || strings.Contains(text, "spike") || strings.Contains(text, "stud") || strings.Contains(text, "ice north") {
			offer.TireSpike = "Шипованные"
		} else if offer.Season == "Зимняя" {
			offer.TireSpike = "Нешипованные"
		}

		if offer.TireWidth != "" && offer.TireSectionRatio != "" && offer.WheelDiameter != "" {
			offer.Marking = fmt.Sprintf("%s/%sR%s", offer.TireWidth, offer.TireSectionRatio, offer.WheelDiameter)
		}
	}

	// Fallback диаметра для прочих деталей
	if offer.WheelDiameter == "" && part.Diameter != "" {
		offer.WheelDiameter = cleanDiameter(part.Diameter)
	}
}

// RenderOfferXML сериализует одиночный оффер в аккуратный XML-фрагмент
func RenderOfferXML(offer Offer) ([]byte, error) {
	var buf bytes.Buffer
	enc := xml.NewEncoder(&buf)
	enc.Indent("  ", "  ")
	if err := enc.Encode(offer); err != nil {
		return nil, err
	}
	buf.WriteByte('\n')
	return buf.Bytes(), nil
}
