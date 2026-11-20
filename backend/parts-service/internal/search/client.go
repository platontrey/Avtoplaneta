package search

import (
	"errors"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"

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

// InitElasticsearch initializes the Elasticsearch client
func InitElasticsearch(url string) error {
	cfg := elasticsearch.Config{
		Addresses: []string{url},
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		RecordESError()
		return fmt.Errorf("error creating the client: %w", err)
	}

	res, err := client.Info()
	if err != nil {
		RecordESError()
		return fmt.Errorf("error getting response: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		RecordESError()
		return fmt.Errorf("error response: %s", res.String())
	}

	esClient = client
	log.Println("Elasticsearch client initialized successfully")

	return CreatePartsIndex()
}

// SearchParts выполняет поиск запчастей по запросу
func SearchParts(query map[string]interface{}, from, size int) ([]ElasticsearchPart, int64, error) {
	if esClient == nil {
		return nil, 0, errors.New("elasticsearch client is not initialized")
	}

	queryBody := map[string]interface{}{
		"query": query,
		"from":  from,
		"size":  size,
	}

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(queryBody); err != nil {
		RecordESError()
		return nil, 0, fmt.Errorf("error encoding query: %w", err)
	}

	req := esapi.SearchRequest{
		Index: []string{"parts"},
		Body:  &buf,
	}

	res, err := req.Do(context.Background(), esClient)
	if err != nil {
		RecordESError()
		return nil, 0, fmt.Errorf("error performing search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		RecordESError()
		return nil, 0, fmt.Errorf("search error response: %s", res.String())
	}

	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		RecordESError()
		return nil, 0, fmt.Errorf("error parsing search response: %w", err)
	}

	hitsMap, ok := result["hits"].(map[string]interface{})
	if !ok {
		return []ElasticsearchPart{}, 0, nil
	}

	var total int64
	if totalMap, ok := hitsMap["total"].(map[string]interface{}); ok {
		if val, ok := totalMap["value"].(float64); ok {
			total = int64(val)
		}
	}

	hitsList, ok := hitsMap["hits"].([]interface{})
	if !ok {
		return []ElasticsearchPart{}, total, nil
	}

	parts := make([]ElasticsearchPart, 0, len(hitsList))
	for _, hit := range hitsList {
		hitMap, ok := hit.(map[string]interface{})
		if !ok {
			continue
		}

		sourceMap, ok := hitMap["_source"].(map[string]interface{})
		if !ok {
			continue
		}

		sourceData, err := json.Marshal(sourceMap)
		if err != nil {
			continue
		}

		var part ElasticsearchPart
		if err := json.Unmarshal(sourceData, &part); err != nil {
			continue
		}

		parts = append(parts, part)
	}

	return parts, total, nil
}
