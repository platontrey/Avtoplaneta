package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

var esClient *elasticsearch.Client

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
func ParseSearchQuery(query string) map[string]string {
	filters := make(map[string]string)

	if query == "" {
		return filters
	}

	// Split by comma first, then by space if no commas
	var tokens []string
	if strings.Contains(query, ",") {
		// Split by comma and clean
		parts := strings.Split(query, ",")
		for _, part := range parts {
			token := strings.TrimSpace(part)
			if token != "" {
				tokens = append(tokens, token)
			}
		}
	} else {
		// Split by space
		tokens = strings.Fields(query)
	}

	// Known brands
	brands := map[string]bool{
		"bmw": true, "mercedes": true, "audi": true, "volkswagen": true, "vw": true,
		"toyota": true, "nissan": true, "honda": true, "mazda": true, "mitsubishi": true,
		"ford": true, "chevrolet": true, "opel": true, "renault": true, "peugeot": true,
		"citroen": true, "fiat": true, "alfa": true, "lancia": true, "ferrari": true,
		"lamborghini": true, "maserati": true, "bentley": true, "rolls": true, "royce": true,
		"aston": true, "martin": true, "jaguar": true, "land": true, "rover": true,
		"volvo": true, "saab": true, "skoda": true, "seat": true, "porsche": true,
		"lada": true, "vaz": true, "gaz": true, "uaz": true, "kamaz": true,
		"zil": true, "moskvich": true, "izh": true,
	}

	// Known model patterns
	modelPatterns := []*regexp.Regexp{
		regexp.MustCompile(`^[A-Z]\d+$`),       // E90, A4, X5
		regexp.MustCompile(`^[A-Z]\d+[A-Z]?$`), // E90, A4, X5, C180
		regexp.MustCompile(`^\d+[A-Z]+$`),      // 316i, 528i
		regexp.MustCompile(`^[A-Z]-[A-Z]`),     // C-Class, E-Class
	}

	// Characteristics (positions, sides)
	characteristics := map[string]bool{
		"левый": true, "левая": true, "лев": true, "l": true, "left": true,
		"правый": true, "правая": true, "прав": true, "r": true, "right": true,
		"передний": true, "передняя": true, "перед": true, "front": true, "f": true,
		"задний": true, "задняя": true, "зад": true, "rear": true, "back": true,
		"верхний": true, "верхняя": true, "верх": true, "top": true, "up": true,
		"нижний": true, "нижняя": true, "низ": true, "bottom": true, "down": true,
		"тормоз": true, "тормозной": true, "тормозные": true, "brake": true,
		"амортизатор": true, "аморт": true, "стойка": true, "shock": true,
		"фильтр": true, "фильт": true, "filter": true,
		"масло": true, "масляный": true, "масляного": true, "oil": true,
		"воздух": true, "воздушный": true, "воздушного": true, "air": true,
	}

	var generalSearch []string

	for _, token := range tokens {
		tokenLower := strings.ToLower(token)

		// Check if it's a brand
		if brands[tokenLower] {
			if filters["brand"] == "" {
				filters["brand"] = token
			} else {
				generalSearch = append(generalSearch, token)
			}
			continue
		}

		// Check if it's a model
		isModel := false
		for _, pattern := range modelPatterns {
			if pattern.MatchString(token) {
				if filters["model"] == "" {
					filters["model"] = token
				} else {
					generalSearch = append(generalSearch, token)
				}
				isModel = true
				break
			}
		}
		if isModel {
			continue
		}

		// Check if it's a characteristic
		if characteristics[tokenLower] {
			generalSearch = append(generalSearch, token)
			continue
		}

		// Everything else goes to general search
		generalSearch = append(generalSearch, token)
	}

	// Combine general search terms
	if len(generalSearch) > 0 {
		filters["general"] = strings.Join(generalSearch, " ")
	}

	return filters
}

// BuildSearchQuery builds an Elasticsearch query from search parameters
func BuildSearchQuery(search, category, brand, model, location, salesman, status, hasPhoto string) map[string]interface{} {
	var must []map[string]interface{}

	// Parse search query to extract filters
	parsedFilters := ParseSearchQuery(search)

	// Add general search query (searches in all text fields)
	if generalSearch := parsedFilters["general"]; generalSearch != "" {
		must = append(must, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					// Exact match with higher boost on name
					{
						"multi_match": map[string]interface{}{
							"query":  generalSearch,
							"fields": []string{"name^4", "description^3", "category^2", "brand^2", "model^2", "salesman", "location"},
							"type":   "best_fields",
						},
					},
					// Prefix match for partial words
					{
						"multi_match": map[string]interface{}{
							"query":  generalSearch,
							"fields": []string{"name", "description", "category", "brand", "model", "salesman", "location"},
							"type":   "phrase_prefix",
							"boost":  2.0,
						},
					},
					// Fuzzy match for typos
					{
						"multi_match": map[string]interface{}{
							"query":     generalSearch,
							"fields":    []string{"name", "description", "category", "brand", "model", "salesman", "location"},
							"fuzziness": "AUTO",
							"boost":     0.5,
						},
					},
				},
				"minimum_should_match": 1,
			},
		})
	}

	// Add term filters
	if category != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"category": category,
			},
		})
	}

	// Use explicit brand parameter or parsed from search
	searchBrand := brand
	if searchBrand == "" && parsedFilters["brand"] != "" {
		searchBrand = parsedFilters["brand"]
	}
	if searchBrand != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"brand": searchBrand,
			},
		})
	}

	// Use explicit model parameter or parsed from search
	searchModel := model
	if searchModel == "" && parsedFilters["model"] != "" {
		searchModel = parsedFilters["model"]
	}
	if searchModel != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"model": searchModel,
			},
		})
	}

	if location != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"location": location,
			},
		})
	}

	if salesman != "" {
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"salesman": salesman,
			},
		})
	}

	if status != "" {
		statusBool := status == "true"
		must = append(must, map[string]interface{}{
			"term": map[string]interface{}{
				"status": statusBool,
			},
		})
	}

	if hasPhoto != "" {
		if hasPhoto == "with" {
			must = append(must, map[string]interface{}{
				"exists": map[string]interface{}{
					"field": "photo",
				},
			})
			must = append(must, map[string]interface{}{
				"bool": map[string]interface{}{
					"must_not": []map[string]interface{}{
						{
							"term": map[string]interface{}{
								"photo": "",
							},
						},
					},
				},
			})
		} else if hasPhoto == "without" {
			must = append(must, map[string]interface{}{
				"bool": map[string]interface{}{
					"should": []map[string]interface{}{
						{
							"bool": map[string]interface{}{
								"must_not": []map[string]interface{}{
									{
										"exists": map[string]interface{}{
											"field": "photo",
										},
									},
								},
							},
						},
						{
							"term": map[string]interface{}{
								"photo": "",
							},
						},
					},
					"minimum_should_match": 1,
				},
			})
		}
		// "all" - не добавляем фильтр, показываем все
	}

	if len(must) == 0 {
		return map[string]interface{}{
			"match_all": map[string]interface{}{},
		}
	}

	return map[string]interface{}{
		"bool": map[string]interface{}{
			"must": must,
		},
	}
}
