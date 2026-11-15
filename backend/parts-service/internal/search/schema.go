package search

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8/esapi"

	"parts-service/internal/catalog"
)

//go:embed schema.json.tmpl
var partsIndexSchemaTemplate string

func buildPartsIndexMapping() string {
	synonymsList := catalog.LoadSynonyms()
	synonymsJSON, _ := json.Marshal(synonymsList)
	return fmt.Sprintf(partsIndexSchemaTemplate, string(synonymsJSON))
}

// CreatePartsIndex creates the parts index with proper mappings and analyzers
func CreatePartsIndex() error {
	// Сначала проверяем, существует ли индекс parts. Если существует, удаляем его, чтобы полностью обновить маппинги и анализаторы
	existsReq := esapi.IndicesExistsRequest{
		Index: []string{"parts"},
	}
	existsRes, err := existsReq.Do(context.Background(), esClient)
	if err == nil && existsRes.StatusCode == 200 {
		_ = existsRes.Body.Close()
		log.Println("Обнаружен старый индекс parts, удаляем для обновления маппинга...")
		deleteReq := esapi.IndicesDeleteRequest{
			Index: []string{"parts"},
		}
		delRes, delErr := deleteReq.Do(context.Background(), esClient)
		if delErr == nil {
			_ = delRes.Body.Close()
		}
	} else if existsRes != nil {
		_ = existsRes.Body.Close()
	}

	mapping := buildPartsIndexMapping()

	req := esapi.IndicesCreateRequest{
		Index: "parts",
		Body:  strings.NewReader(mapping),
	}

	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		return fmt.Errorf("error creating index: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	if res.IsError() && res.StatusCode != 400 { // 400 means index already exists
		return fmt.Errorf("error response: %s", res.String())
	}

	log.Println("Parts index created successfully")
	return nil
}

// RefreshPartsIndex выполняет refresh индекса parts для немедленной доступности изменений в поиске
func RefreshPartsIndex(ctx context.Context) error {
	if esClient == nil {
		return nil
	}

	req := esapi.IndicesRefreshRequest{
		Index: []string{"parts"},
	}

	res, err := req.Do(ctx, esClient)
	if err != nil {
		RecordESError()
		return fmt.Errorf("ошибка при выполнении refresh индекса: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		RecordESError()
		return fmt.Errorf("ошибка при выполнении refresh индекса: статус %s", res.Status())
	}

	return nil
}
