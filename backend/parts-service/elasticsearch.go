package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var esClient *elasticsearch.Client

// elasticsearchAdapter адаптер для Elasticsearch, реализующий ElasticsearchClient
type elasticsearchAdapter struct{}

// NewElasticsearchAdapter создает новый адаптер для Elasticsearch
func NewElasticsearchAdapter() ElasticsearchClient {
	return &elasticsearchAdapter{}
}

type ElasticsearchPart struct {
	ID          uint    `json:"id"`
	Name        string  `json:"name"`
	Quantity    int     `json:"quantity"`
	Description string  `json:"description"`
	Category    string  `json:"category"`
	Price       float64 `json:"price"`
	Salesman    string  `json:"salesman"`
	Location    string  `json:"location"`
	Status      bool    `json:"status"`
	Brand       string  `json:"brand"`
	Model       string  `json:"model"`
	Photo       string  `json:"photo"`
}

// InitElasticsearch initializes the Elasticsearch client
func InitElasticsearch() error {
	cfg := elasticsearch.Config{
		Addresses: []string{
			"http://localhost:9200", // Default Elasticsearch address
		},
	}

	var err error
	esClient, err = elasticsearch.NewClient(cfg)
	if err != nil {
		return fmt.Errorf("error creating Elasticsearch client: %w", err)
	}

	// Test the connection
	res, err := esClient.Info()
	if err != nil {
		return fmt.Errorf("error getting Elasticsearch info: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("error response from Elasticsearch: %s", res.String())
	}

	log.Println("Successfully connected to Elasticsearch")
	return nil
}

// CreatePartsIndex creates the parts index with proper mappings
func CreatePartsIndex() error {
	mapping := `{
		"mappings": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "text",
					"analyzer": "standard",
					"fields": {
						"keyword": {
							"type": "keyword"
						}
					}
				},
				"quantity": {
					"type": "integer"
				},
				"description": {
					"type": "text",
					"analyzer": "standard"
				},
				"category": {
					"type": "keyword",
					"fields": {
						"text": {
							"type": "text",
							"analyzer": "standard"
						}
					}
				},
				"price": {
					"type": "float"
				},
				"salesman": {
					"type": "keyword",
					"fields": {
						"text": {
							"type": "text",
							"analyzer": "standard"
						}
					}
				},
				"location": {
					"type": "keyword",
					"fields": {
						"text": {
							"type": "text",
							"analyzer": "standard"
						}
					}
				},
				"status": {
					"type": "boolean"
				},
				"brand": {
					"type": "keyword",
					"fields": {
						"text": {
							"type": "text",
							"analyzer": "standard"
						}
					}
				},
				"model": {
					"type": "keyword",
					"fields": {
						"text": {
							"type": "text",
							"analyzer": "standard"
						}
					}
				},
				"photo": {
					"type": "keyword"
				},
				"inn": {
					"type": "keyword",
					"fields": {
						"text": {
							"type": "text",
							"analyzer": "standard"
						}
					}
				}
			}
		}
	}`

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

	log.Println("Parts index created or already exists")
	return nil
}

// IndexPart indexes a single part in Elasticsearch
func IndexPart(part *Part) error {
	esPart := ElasticsearchPart{
		ID:          part.ID,
		Name:        part.Name,
		Quantity:    part.Quantity,
		Description: part.Description,
		Category:    part.Category,
		Price:       part.Price,
		Salesman:    part.Salesman,
		Location:    part.Location,
		Status:      part.Status,
		Brand:       part.Brand,
		Model:       part.Model,
		Photo:       part.Photo,
	}

	body, err := json.Marshal(esPart)
	if err != nil {
		return fmt.Errorf("error marshaling part: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      "parts",
		DocumentID: fmt.Sprintf("%d", part.ID),
		Body:       bytes.NewReader(body),
		Refresh:    "true", // Make it immediately available for search
	}

	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		return fmt.Errorf("error indexing part: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	if res.IsError() {
		return fmt.Errorf("error response: %s", res.String())
	}

	return nil
}

// DeletePartFromIndex removes a part from the Elasticsearch index
func DeletePartFromIndex(partID uint) error {
	req := esapi.DeleteRequest{
		Index:      "parts",
		DocumentID: fmt.Sprintf("%d", partID),
	}

	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		return fmt.Errorf("error deleting part from index: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	if res.IsError() && res.StatusCode != 404 { // 404 means document doesn't exist
		return fmt.Errorf("error response: %s", res.String())
	}

	return nil
}

// SearchParts performs a search query against the parts index
func SearchParts(query map[string]interface{}, from, size int) ([]ElasticsearchPart, int64, error) {
	searchBody := map[string]interface{}{
		"from":  from,
		"size":  size,
		"query": query,
		"sort": []map[string]interface{}{
			{"_score": map[string]string{"order": "desc"}},
			{"name.keyword": map[string]string{"order": "asc"}},
		},
	}

	body, err := json.Marshal(searchBody)
	if err != nil {
		return nil, 0, fmt.Errorf("error marshaling search query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{"parts"},
		Body:  bytes.NewReader(body),
	}

	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		return nil, 0, fmt.Errorf("error performing search: %w", err)
	}
	defer func() {
		if closeErr := res.Body.Close(); closeErr != nil {
			log.Printf("Error closing response body: %v", closeErr)
		}
	}()

	if res.IsError() {
		return nil, 0, fmt.Errorf("search error response: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("error parsing search response: %w", err)
	}

	hits := result["hits"].(map[string]interface{})
	total := hits["total"].(map[string]interface{})["value"].(float64)
	hitsArray := hits["hits"].([]interface{})

	parts := make([]ElasticsearchPart, 0, len(hitsArray))
	for _, hit := range hitsArray {
		source := hit.(map[string]interface{})["_source"]
		sourceBytes, _ := json.Marshal(source)

		var part ElasticsearchPart
		if err := json.Unmarshal(sourceBytes, &part); err != nil {
			log.Printf("Error unmarshaling part: %v", err)
			continue
		}
		parts = append(parts, part)
	}

	return parts, int64(total), nil
}

// ParseSearchQuery parses a search query and returns filters for different fields

// IndexPart индексирует запчасть в Elasticsearch
func (e *elasticsearchAdapter) IndexPart(part *Part) error {
	return IndexPart(part)
}

// DeletePartFromIndex удаляет запчасть из индекса Elasticsearch
func (e *elasticsearchAdapter) DeletePartFromIndex(partID uint) error {
	return DeletePartFromIndex(partID)
}

// SearchParts выполняет поиск в Elasticsearch
func (e *elasticsearchAdapter) SearchParts(query map[string]interface{}, from, size int) ([]ElasticsearchPart, int64, error) {
	return SearchParts(query, from, size)
}
