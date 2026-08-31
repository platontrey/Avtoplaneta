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
	"github.com/prometheus/client_golang/prometheus"
)

var (
	esErrorsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "avtoplaneta_elasticsearch_errors_total",
			Help: "Total number of Elasticsearch errors",
		},
	)
)

func RecordESError() {
	esErrorsTotal.Inc()
}
func init() {
	prometheus.MustRegister(esErrorsTotal)
}

var esClient *elasticsearch.Client

// elasticsearchAdapter адаптер для Elasticsearch, реализующий ElasticsearchClient
type elasticsearchAdapter struct{}

// NewElasticsearchAdapter создает новый адаптер для Elasticsearch
func NewElasticsearchAdapter() ElasticsearchClient {
	return &elasticsearchAdapter{}
}

type ElasticsearchPart struct {
	ID                 int64    `json:"id"`
	Name               string   `json:"name"`
	Quantity           int      `json:"quantity"`
	Description        string   `json:"description"`
	Category           string   `json:"category"`
	Price              float64  `json:"price"`
	Salesman           string   `json:"salesman"`
	Location           string   `json:"location"`
	Address            string   `json:"address"`
	Status             bool     `json:"status"`
	Brand              string   `json:"brand"`
	Model              string   `json:"model"`
	Photos             []string `json:"photos"`
	FrontRear          string   `json:"front_rear,omitempty"`
	LeftRight          string   `json:"left_right,omitempty"`
	TopBottom          string   `json:"top_bottom,omitempty"`
	OEMCode            string   `json:"oem_code,omitempty"`
	ManufacturerCode   string   `json:"manufacturer_code,omitempty"`
	Manufacturer       string   `json:"manufacturer,omitempty"`
	Color              string   `json:"color,omitempty"`
	Condition          string   `json:"condition,omitempty"`
	BodyBrand          string   `json:"body_brand,omitempty"`
	EngineBrand        string   `json:"engine_brand,omitempty"`
	CarReleaseDate     string   `json:"car_release_date,omitempty"`
	Transmission       string   `json:"transmission,omitempty"`
	TransmissionModel  string   `json:"transmission_model,omitempty"`
	Drive              string   `json:"drive,omitempty"`
	Defect             string   `json:"defect,omitempty"`
	Number             string   `json:"number,omitempty"`
	SupplierCode       string   `json:"supplier_code,omitempty"`
	WearPercentage     string   `json:"wear_percentage,omitempty"`
	VIN                string   `json:"vin,omitempty"`
	Season             string   `json:"season,omitempty"`
	Diameter           string   `json:"diameter,omitempty"`
	Width              string   `json:"width,omitempty"`
	Profile            string   `json:"profile,omitempty"`
	TireQuantity       string   `json:"tire_quantity,omitempty"`
	Drilling           string   `json:"drilling,omitempty"`
	Offset             string   `json:"offset,omitempty"`
	CenterHoleDiameter string   `json:"center_hole_diameter,omitempty"`
	TireModel          string   `json:"tire_model,omitempty"`
}

// InitElasticsearch initializes the Elasticsearch client
func InitElasticsearch(url string) error {
	if url == "" {
		url = "http://localhost:9200"
	}
	cfg := elasticsearch.Config{
		Addresses: []string{url},
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

	synonymsList := LoadSynonyms()
	synonymsJSON, _ := json.Marshal(synonymsList)

	mapping := fmt.Sprintf(`{
		"settings": {
			"analysis": {
				"filter": {
					"russian_stop": {
						"type": "stop",
						"stopwords": "_russian_"
					},
					"russian_stemmer": {
						"type": "stemmer",
						"language": "russian"
					},
					"auto_synonyms": {
						"type": "synonym_graph",
						"synonyms": %s
					},
					"code_ngram": {
						"type": "edge_ngram",
						"min_gram": 2,
						"max_gram": 15
					},
					"oem_delimiter": {
						"type": "word_delimiter_graph",
						"generate_word_parts": true,
						"generate_number_parts": true,
						"catenate_words": true,
						"catenate_numbers": true,
						"catenate_all": true,
						"split_on_case_change": true,
						"preserve_original": true
					}
				},
				"analyzer": {
					"custom_russian_index": {
						"tokenizer": "standard",
						"filter": [
							"lowercase",
							"russian_stop",
							"russian_stemmer"
						]
					},
					"custom_russian_search": {
						"tokenizer": "standard",
						"filter": [
							"lowercase",
							"russian_stop",
							"auto_synonyms",
							"russian_stemmer"
						]
					},
					"part_number_analyzer": {
						"tokenizer": "standard",
						"filter": [
							"lowercase",
							"oem_delimiter",
							"code_ngram"
						]
					},
					"part_number_search_analyzer": {
						"tokenizer": "standard",
						"filter": [
							"lowercase",
							"oem_delimiter"
						]
					}
				}
			}
		},
		"mappings": {
			"properties": {
				"id": {
					"type": "integer"
				},
				"name": {
					"type": "text",
					"analyzer": "custom_russian_index",
					"search_analyzer": "custom_russian_search",
					"fields": {
						"keyword": {
							"type": "keyword"
						},
						"ngram": {
							"type": "text",
							"analyzer": "part_number_analyzer",
							"search_analyzer": "part_number_search_analyzer"
						}
					}
				},
				"quantity": {
					"type": "integer"
				},
				"description": {
					"type": "text",
					"analyzer": "custom_russian_index",
					"search_analyzer": "custom_russian_search"
				},
				"category": {
					"type": "keyword",
					"fields": {
						"text": {
							"type": "text",
							"analyzer": "custom_russian_index",
							"search_analyzer": "custom_russian_search"
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
							"analyzer": "custom_russian_index",
							"search_analyzer": "custom_russian_search"
						}
					}
				},
				"location": {
					"type": "keyword",
					"fields": {
						"text": {
							"type": "text",
							"analyzer": "custom_russian_index",
							"search_analyzer": "custom_russian_search"
						}
					}
				},
				"address": {
					"type": "keyword",
					"fields": {
						"text": {
							"type": "text",
							"analyzer": "custom_russian_index",
							"search_analyzer": "custom_russian_search"
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
				"photos": {
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
	}`, string(synonymsJSON))

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

// IndexPart indexes a single part in Elasticsearch
func IndexPart(part *Part) error {
	esPart := ElasticsearchPart{
		ID:                 part.ID,
		Name:               part.Name,
		Quantity:           part.Quantity,
		Description:        part.Description,
		Category:           part.Category,
		Price:              part.Price,
		Salesman:           part.Salesman,
		Location:           part.Location,
		Address:            part.Address,
		Status:             part.Status,
		Brand:              part.Brand,
		Model:              part.Model,
		Photos:             part.Photos,
		FrontRear:          part.FrontRear,
		LeftRight:          part.LeftRight,
		TopBottom:          part.TopBottom,
		OEMCode:            part.OEMCode,
		ManufacturerCode:   part.ManufacturerCode,
		Manufacturer:       part.Manufacturer,
		Color:              part.Color,
		Condition:          part.Condition,
		BodyBrand:          part.BodyBrand,
		EngineBrand:        part.EngineBrand,
		CarReleaseDate:     part.CarReleaseDate,
		Transmission:       part.Transmission,
		TransmissionModel:  part.TransmissionModel,
		Drive:              part.Drive,
		Defect:             part.Defect,
		Number:             part.Number,
		SupplierCode:       part.SupplierCode,
		WearPercentage:     part.WearPercentage,
		VIN:                part.VIN,
		Season:             part.Season,
		Diameter:           part.Diameter,
		Width:              part.Width,
		Profile:            part.Profile,
		TireQuantity:       part.TireQuantity,
		Drilling:           part.Drilling,
		Offset:             part.Offset,
		CenterHoleDiameter: part.CenterHoleDiameter,
		TireModel:          part.TireModel,
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
		RecordESError()
		return fmt.Errorf("error response: %s", res.String())
	}

	return nil
}

// DeletePartFromIndex removes a part from the Elasticsearch index
func DeletePartFromIndex(partID int64) error {
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
func (e *elasticsearchAdapter) DeletePartFromIndex(partID int64) error {
	return DeletePartFromIndex(partID)
}

// SearchParts выполняет поиск в Elasticsearch
func (e *elasticsearchAdapter) SearchParts(query map[string]interface{}, from, size int) ([]ElasticsearchPart, int64, error) {
	return SearchParts(query, from, size)
}
