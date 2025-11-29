package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// Offers представляет структуры XML для прайс-листа Drom
type Offers struct {
	XMLName xml.Name `xml:"offers"`
	Offers  []Offer  `xml:"offer"`
}

type Offer struct {
	Name          string `xml:"name"`
	Available     bool   `xml:"available"`
	OemNumber     string `xml:"oem_number,omitempty"`
	AnalogNumbers string `xml:"analog_numbers,omitempty"`
	Manufacturer  string `xml:"manufacturer,omitempty"`
	Ordercode     string `xml:"ordercode,omitempty"`
	Condition     string `xml:"condition,omitempty"`
	Brandcars     string `xml:"brandcars,omitempty"`
	Modelcars     string `xml:"modelcars,omitempty"`
	Bodycars      string `xml:"bodycars,omitempty"`
	Engine        string `xml:"engine,omitempty"`
	Year          string `xml:"year,omitempty"`
	Lr            string `xml:"lr,omitempty"`
	Fr            string `xml:"fr,omitempty"`
	Ud            string `xml:"ud,omitempty"`
	Color         string `xml:"color,omitempty"`
	Supplier      string `xml:"supplier,omitempty"`
	SupplierInn   string `xml:"supplier_inn,omitempty"`
	Sklad         string `xml:"sklad,omitempty"`
	SupplierArt   string `xml:"supplier_art,omitempty"`
}

// GenerateXMLPriceList генерирует XML прайс-лист для Drom из списка запчастей
func GenerateXMLPriceList(parts []Part) ([]byte, error) {
	offers := Offers{
		Offers: make([]Offer, 0, len(parts)),
	}

	for _, part := range parts {
		// Пропускаем запчасти с отрицательным количеством или помеченные для удаления
		if part.Quantity < 0 || part.ToDeleteAt != nil {
			continue
		}

		offer := Offer{
			Name:         strings.TrimSpace(part.Name),
			Available:    part.Status,
			Manufacturer: strings.TrimSpace(part.Brand),
			Brandcars:    strings.TrimSpace(part.Brand),
			Modelcars:    strings.TrimSpace(part.Model),
			Supplier:     strings.TrimSpace(part.Salesman),
			SupplierInn:  getUserINN(part.SellerID),
			Sklad:        strings.TrimSpace(part.Location),
			Condition:    "б/у", // По умолчанию б/у, можно расширить модель для хранения состояния
		}

		offers.Offers = append(offers.Offers, offer)
	}

	// Добавляем XML заголовок
	xmlData, err := xml.MarshalIndent(offers, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга XML: %v", err)
	}

	// Добавляем XML декларацию
	xmlWithHeader := []byte(xml.Header + string(xmlData))

	return xmlWithHeader, nil
}

// GetPartsForXML получает все доступные запчасти для экспорта в XML
func GetPartsForXML() ([]Part, error) {
	var parts []Part

	// Получаем все запчасти, которые не помечены для удаления и имеют quantity >= 0
	err := db.Where("to_delete_at IS NULL AND quantity >= 0").Find(&parts).Error
	if err != nil {
		return nil, fmt.Errorf("ошибка получения частей из базы данных: %v", err)
	}

	return parts, nil
}

// getUserINN получает ИНН пользователя по его ID из auth-service
func getUserINN(sellerID uint) string {
	if sellerID == 0 {
		return ""
	}

	// Создаем HTTP клиент для запроса к auth-service
	client := &http.Client{}

	// Формируем URL для получения данных пользователя
	url := fmt.Sprintf("http://localhost:8083/admin/users/%d", sellerID)

	// Создаем запрос
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Printf("Ошибка создания запроса для получения ИНН пользователя %d: %v\n", sellerID, err)
		return ""
	}

	// Выполняем запрос
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Ошибка выполнения запроса для получения ИНН пользователя %d: %v\n", sellerID, err)
		return ""
	}
	defer func() {
		if closeErr := resp.Body.Close(); closeErr != nil {
			fmt.Printf("Ошибка закрытия тела ответа для пользователя %d: %v\n", sellerID, closeErr)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		fmt.Printf("Ошибка получения данных пользователя %d: статус %d\n", sellerID, resp.StatusCode)
		return ""
	}

	// Читаем тело ответа
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Ошибка чтения ответа для пользователя %d: %v\n", sellerID, err)
		return ""
	}

	// Парсим JSON ответ
	var userData struct {
		INN string `json:"inn"`
	}

	if err := json.Unmarshal(body, &userData); err != nil {
		fmt.Printf("Ошибка парсинга JSON для пользователя %d: %v\n", sellerID, err)
		return ""
	}

	fmt.Printf("Получен ИНН для пользователя %d: %s\n", sellerID, userData.INN)
	return userData.INN
}
