package domain

import (
	"bytes"
	"encoding/xml"
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
	Price         float64  `xml:"price,omitempty"`
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
}

// BuildOffer строит предложение Offer из Part и реквизитов продавца
func BuildOffer(part Part, sellerINN string, baseImageURL string) Offer {
	baseImageURL = strings.TrimRight(strings.TrimSpace(baseImageURL), "/")
	if baseImageURL == "" {
		baseImageURL = "https://backend-server.ru"
	}

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
			pictures = append(pictures, baseImageURL+photoPath)
		}
	}

	condition := strings.TrimSpace(part.Condition)
	if condition == "" {
		condition = "Б/у"
	}

	return Offer{
		Name:          strings.TrimSpace(part.Name),
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
		Supplier:      strings.TrimSpace(part.Salesman),
		SupplierInn:   sellerINN,
		Sklad:         strings.TrimSpace(part.Location),
		SupplierArt:   strings.TrimSpace(part.SupplierCode),
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
