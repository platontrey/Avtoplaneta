package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartToESPart(t *testing.T) {
	p := Part{
		PartCore: PartCore{
			ID:          42,
			Name:        "Тормозной диск",
			Quantity:    2,
			Description: "Оригинальный диск",
			Category:    "Тормозная система",
			Price:       3500.50,
			Brand:       "Toyota",
			Model:       "Camry",
			Photos:      StringArray{"https://example.com/photo1.jpg"},
			Address:     "A1-05",
		},
		PartSpecifications: PartSpecifications{
			OEMCode: "43512-33140",
		},
	}

	es := partToESPart(&p)
	assert.Equal(t, p.ID, es.ID)
	assert.Equal(t, p.Name, es.Name)
	assert.Equal(t, p.Quantity, es.Quantity)
	assert.Equal(t, p.Description, es.Description)
	assert.Equal(t, p.Category, es.Category)
	assert.Equal(t, p.Price, es.Price)
	assert.Equal(t, p.Brand, es.Brand)
	assert.Equal(t, p.Model, es.Model)
	assert.Equal(t, []string(p.Photos), es.Photos)
	assert.Equal(t, p.OEMCode, es.OEMCode)
	assert.Equal(t, p.Address, es.Address)
}

func TestBulkIndexParts_NilClient(t *testing.T) {
	oldClient := esClient
	esClient = nil
	defer func() { esClient = oldClient }()

	err := BulkIndexParts(context.Background(), []Part{{PartCore: PartCore{ID: 1, Name: "Test"}}})
	assert.NoError(t, err)

	err = RefreshPartsIndex(context.Background())
	assert.NoError(t, err)
}

func TestBulkIndexParts_Empty(t *testing.T) {
	err := BulkIndexParts(context.Background(), nil)
	assert.NoError(t, err)

	err = BulkIndexParts(context.Background(), []Part{})
	assert.NoError(t, err)
}

func TestBulkIndexParts_MockServer(t *testing.T) {
	bulkCalled := false
	refreshCalled := false
	var receivedBody string

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-elastic-product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/_bulk") {
			bulkCalled = true
			bodyBytes, _ := io.ReadAll(r.Body)
			receivedBody = string(bodyBytes)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"took": 5, "errors": false, "items": []}`))
			return
		}
		if strings.HasSuffix(r.URL.Path, "/_refresh") {
			refreshCalled = true
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"_shards": {"total": 1, "successful": 1, "failed": 0}}`))
			return
		}
		http.NotFound(w, r)
	}))
	defer ts.Close()

	client, err := elasticsearch.NewClient(elasticsearch.Config{
		Addresses: []string{ts.URL},
	})
	require.NoError(t, err)

	oldClient := esClient
	esClient = client
	defer func() { esClient = oldClient }()

	parts := []Part{
		{PartCore: PartCore{ID: 101, Name: "Фара левая", Price: 5000}},
		{PartCore: PartCore{ID: 102, Name: "Фара правая", Price: 5200}},
	}

	err = BulkIndexParts(context.Background(), parts)
	assert.NoError(t, err)
	assert.True(t, bulkCalled, "Bulk API must be invoked")
	assert.True(t, refreshCalled, "IndicesRefresh must be invoked after bulk indexing")

	assert.Contains(t, receivedBody, `{"index":{"_index":"parts","_id":"101"}}`)
	assert.Contains(t, receivedBody, `"name":"Фара левая"`)
	assert.Contains(t, receivedBody, `{"index":{"_index":"parts","_id":"102"}}`)
	assert.Contains(t, receivedBody, `"name":"Фара правая"`)
}
