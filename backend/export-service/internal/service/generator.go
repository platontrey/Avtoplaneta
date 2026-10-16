package service

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
	"time"

	"export-service/internal/config"
	"export-service/internal/domain"
)

// dromHTTPClient переиспользуется для всех исходящих вызовов к API Drom
var dromHTTPClient = &http.Client{Timeout: 2 * time.Minute}

// GenerateXMLPriceList собирает XML прайс-лист Drom из среза запчастей
func GenerateXMLPriceList(parts []domain.Part, sellerINNs map[int64]string, baseImageURL string) ([]byte, error) {
	offers := domain.Offers{
		Offers: make([]domain.Offer, 0, len(parts)),
	}

	for _, part := range parts {
		// Пропускаем запчасти, помеченные к удалению
		if part.ToDeleteAt != nil {
			continue
		}
		offers.Offers = append(offers.Offers, domain.BuildOffer(part, sellerINNs[part.SellerID], baseImageURL))
	}

	// Оцениваем средний размер XML оффера (~500 байт на запчасть) для предварительного выделения буфера
	var buf bytes.Buffer
	buf.Grow(len(offers.Offers)*500 + len(xml.Header))
	buf.WriteString(xml.Header)

	enc := xml.NewEncoder(&buf)
	enc.Indent("", "  ")
	if err := enc.Encode(offers); err != nil {
		return nil, fmt.Errorf("ошибка маршалинга XML: %w", err)
	}

	return buf.Bytes(), nil
}

// SendToDromAPI отправляет XML прайс-лист на API Drom.ru
func SendToDromAPI(ctx context.Context, cfg *config.Config, xmlData []byte) error {
	if cfg.DromAPIKey == "" {
		return fmt.Errorf("DROM_API_KEY не установлен в переменных окружения")
	}
	if cfg.DromPacketID == "" {
		return fmt.Errorf("DROM_PACKET_ID не установлен в переменных окружения")
	}

	// Вычисляем auth хэш sha512
	hash := sha512.Sum512([]byte(cfg.DromAPIKey))
	auth := hex.EncodeToString(hash[:])

	// Создаем multipart/form-data
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	if err := writer.WriteField("packetId", cfg.DromPacketID); err != nil {
		return fmt.Errorf("ошибка добавления packetId: %w", err)
	}
	if err := writer.WriteField("auth", auth); err != nil {
		return fmt.Errorf("ошибка добавления auth: %w", err)
	}

	file, err := writer.CreateFormFile("data", PriceListFilename)
	if err != nil {
		return fmt.Errorf("ошибка создания form file: %w", err)
	}
	if _, err := file.Write(xmlData); err != nil {
		return fmt.Errorf("ошибка записи XML данных: %w", err)
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("ошибка закрытия multipart writer: %w", err)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 2*time.Minute)
	defer cancel()

	request, err := http.NewRequestWithContext(reqCtx, http.MethodPost, cfg.DromAPIURL, &body)
	if err != nil {
		return fmt.Errorf("ошибка создания HTTP запроса: %w", err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())

	response, err := dromHTTPClient.Do(request)
	if err != nil {
		return fmt.Errorf("ошибка выполнения HTTP запроса к Drom: %w", err)
	}
	defer response.Body.Close()

	responseBody, err := io.ReadAll(io.LimitReader(response.Body, 1<<20))
	if err != nil {
		return fmt.Errorf("ошибка чтения ответа Drom: %w", err)
	}

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("ошибка API Drom (HTTP %d): %s", response.StatusCode, string(responseBody))
	}

	return nil
}
