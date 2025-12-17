package main

import (
	"bytes"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"strings"
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

// GenerateXMLPriceList генерирует XML прайс-лист для Drom из списка запчастей
func GenerateXMLPriceList(parts []Part) ([]byte, error) {
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
			Available:     part.Status && part.Quantity > 0,
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
			SupplierInn:   getUserINN(part.SellerID),
			Sklad:         strings.TrimSpace(part.Location),
			SupplierArt:   strings.TrimSpace(part.SupplierCode),
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

	// Получаем все запчасти, которые не помечены для удаления и имеют quantity > 0
	err := db.Where("to_delete_at IS NULL AND quantity > 0").Find(&parts).Error
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

// sendToDromAPI отправляет XML прайс-лист на API Drom.ru
func sendToDromAPI(xmlData []byte) error {
	// Получаем настройки из переменных окружения
	apiKey := os.Getenv("DROM_API_KEY")
	packetID := os.Getenv("DROM_PACKET_ID")

	if apiKey == "" {
		return fmt.Errorf("DROM_API_KEY не установлен в переменных окружения")
	}
	if packetID == "" {
		return fmt.Errorf("DROM_PACKET_ID не установлен в переменных окружения")
	}

	// Вычисляем auth хэш
	hash := sha512.Sum512([]byte(apiKey))
	auth := hex.EncodeToString(hash[:])

	// Создаем multipart/form-data
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// Добавляем packetId
	if err := w.WriteField("packetId", packetID); err != nil {
		return fmt.Errorf("ошибка добавления packetId: %v", err)
	}

	// Добавляем auth
	if err := w.WriteField("auth", auth); err != nil {
		return fmt.Errorf("ошибка добавления auth: %v", err)
	}

	// Добавляем файл data
	fw, err := w.CreateFormFile("data", "pricelist.xml")
	if err != nil {
		return fmt.Errorf("ошибка создания form file: %v", err)
	}
	if _, err := fw.Write(xmlData); err != nil {
		return fmt.Errorf("ошибка записи XML данных: %v", err)
	}

	// Закрываем writer
	w.Close()

	// Создаем HTTP запрос
	req, err := http.NewRequest("POST", "https://baza.drom.ru/good/packet/api/sync", &b)
	if err != nil {
		return fmt.Errorf("ошибка создания HTTP запроса: %v", err)
	}

	// Устанавливаем Content-Type
	req.Header.Set("Content-Type", w.FormDataContentType())

	// Выполняем запрос
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("ошибка выполнения HTTP запроса: %v", err)
	}
	defer resp.Body.Close()

	// Читаем ответ
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа: %v", err)
	}

	// Проверяем статус
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка API Drom: статус %d, ответ: %s", resp.StatusCode, string(respBody))
	}

	fmt.Printf("Успешно отправлен прайс-лист на Drom API. Ответ: %s\n", string(respBody))
	return nil
}
