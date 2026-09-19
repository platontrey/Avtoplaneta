package main

import (
	"bytes"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strings"
	"time"
)

// Offers представляет структуры XML для прайс-листа Drom
type Offers struct {
	XMLName xml.Name `xml:"offers"`
	Offers  []Offer  `xml:"offer"`
}

type Offer struct {
	Name          string  `xml:"name"`
	Available     bool    `xml:"available"`
	Description   string  `xml:"description,omitempty"`
	Price         float64 `xml:"price,omitempty"`
	CurrencyId    string  `xml:"currencyId,omitempty"`
	Picture       string  `xml:"picture,omitempty"`
	Store         string  `xml:"store,omitempty"`
	Quantity      int     `xml:"quantity,omitempty"`
	OemNumber     string  `xml:"oem_number,omitempty"`
	AnalogNumbers string  `xml:"analog_numbers,omitempty"`
	Manufacturer  string  `xml:"manufacturer,omitempty"`
	Ordercode     string  `xml:"ordercode,omitempty"`
	Condition     string  `xml:"condition,omitempty"`
	Brandcars     string  `xml:"brandcars,omitempty"`
	Modelcars     string  `xml:"modelcars,omitempty"`
	Bodycars      string  `xml:"bodycars,omitempty"`
	Engine        string  `xml:"engine,omitempty"`
	Year          string  `xml:"year,omitempty"`
	Lr            string  `xml:"lr,omitempty"`
	Fr            string  `xml:"fr,omitempty"`
	Ud            string  `xml:"ud,omitempty"`
	Color         string  `xml:"color,omitempty"`
	Supplier      string  `xml:"supplier,omitempty"`
	SupplierInn   string  `xml:"supplier_inn,omitempty"`
	Sklad         string  `xml:"sklad,omitempty"`
	SupplierArt   string  `xml:"supplier_art,omitempty"`
}

// GenerateXMLPriceList собирает XML прайс-лист Drom из уже полученных данных.
//
// Функция ничего не запрашивает сама: и запчасти, и ИНН продавцов передаются
// аргументами. Так её можно проверить тестом без сети, а поход за данными
// остаётся в одном месте — в Clients.
func GenerateXMLPriceList(parts []Part, sellerINNs map[int64]string) ([]byte, error) {
	offers := Offers{
		Offers: make([]Offer, 0, len(parts)),
	}

	for _, part := range parts {
		// Пропускаем только запчасти помеченные для удаления
		if part.ToDeleteAt != nil {
			continue
		}

		// Формируем полный URL для фотографии (используем первое фото из массива)
		pictureURL := ""
		if len(part.Photos) > 0 && part.Photos[0] != "" {
			pictureURL = "https://avtoplaneta70.ru" + part.Photos[0]
		}

		// Устанавливаем condition по умолчанию "Б/у" если не указано
		condition := strings.TrimSpace(part.Condition)
		if condition == "" {
			condition = "Б/у"
		}

		offer := Offer{
			Name:          strings.TrimSpace(part.Name),
			Available:     part.Status && part.Quantity >= 0,
			Description:   strings.TrimSpace(part.Description),
			Price:         part.Price,
			CurrencyId:    "RUR",
			Picture:       pictureURL,
			Store:         strings.TrimSpace(part.Salesman),
			Quantity:      part.Quantity,
			OemNumber:     strings.TrimSpace(part.OEMCode),
			AnalogNumbers: "", // Поле для аналогов не реализовано, оставить пустым
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
			SupplierInn:   sellerINNs[part.SellerID],
			Sklad:         strings.TrimSpace(part.Location),
			SupplierArt:   strings.TrimSpace(part.SupplierCode),
		}

		offers.Offers = append(offers.Offers, offer)
	}

	xmlData, err := xml.MarshalIndent(offers, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("ошибка маршалинга XML: %v", err)
	}

	return []byte(xml.Header + string(xmlData)), nil
}

// sendToDromAPI отправляет XML прайс-лист на API Drom.ru
func sendToDromAPI(ctx context.Context, config *Config, xmlData []byte) error {
	if config.DromAPIKey == "" {
		return fmt.Errorf("DROM_API_KEY не установлен в переменных окружения")
	}
	if config.DromPacketID == "" {
		return fmt.Errorf("DROM_PACKET_ID не установлен в переменных окружения")
	}

	// Вычисляем auth хэш
	hash := sha512.Sum512([]byte(config.DromAPIKey))
	auth := hex.EncodeToString(hash[:])

	// Создаем multipart/form-data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("packetId", config.DromPacketID); err != nil {
		return fmt.Errorf("ошибка добавления packetId: %v", err)
	}
	if err := writer.WriteField("auth", auth); err != nil {
		return fmt.Errorf("ошибка добавления auth: %v", err)
	}

	file, err := writer.CreateFormFile("data", priceListFilename)
	if err != nil {
		return fmt.Errorf("ошибка создания form file: %v", err)
	}
	if _, err := file.Write(xmlData); err != nil {
		return fmt.Errorf("ошибка записи XML данных: %v", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия multipart writer: %v", err)
	}

	ctx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, config.DromAPIURL, &body)
	if err != nil {
		return fmt.Errorf("ошибка создания HTTP запроса: %v", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	response, err := dromHTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("ошибка выполнения HTTP запроса: %v", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка API Drom: статус %d, ответ: %s", response.StatusCode, string(responseBody))
	}

	return nil
}

// Один клиент на сервис: соединения переиспользуются, таймаут задан явно.
var dromHTTPClient = &http.Client{Timeout: 2 * time.Minute}
