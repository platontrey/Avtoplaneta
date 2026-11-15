package search

import (
	"avtoplaneta/pkg/constants"
	"parts-service/internal/domain"
)

// BuildElasticsearchQuery строит запрос для Elasticsearch
func BuildElasticsearchQuery(params domain.InventoryQueryParams) map[string]interface{} {
	must := make([]map[string]interface{}, 0, 8)
	filter := make([]map[string]interface{}, 0, 8)
	should := make([]map[string]interface{}, 0, 8)

	// Фильтр для отображения валидных запчастей (quantity >= 0)
	filter = append(filter, map[string]interface{}{
		"range": map[string]interface{}{
			"quantity": map[string]interface{}{
				"gte": constants.PartQuantityDefective,
			},
		},
	})

	if params.Search != "" {
		fragments := SplitMultiSearchQuery(params.Search)
		if len(fragments) > 1 {
			if q := buildMultiSearchQuery(fragments); q != nil {
				must = append(must, q)
			}
		} else if len(fragments) == 1 {
			searchMust, searchShould := buildSingleSearchQuery(fragments[0])
			must = append(must, searchMust...)
			should = append(should, searchShould...)
		}
	}

	filter = append(filter, buildQueryFilters(params)...)

	boolQuery := map[string]interface{}{}
	if len(must) > 0 {
		boolQuery["must"] = must
	}
	if len(filter) > 0 {
		boolQuery["filter"] = filter
	}
	if len(should) > 0 {
		boolQuery["should"] = should
	}

	return map[string]interface{}{
		"bool": boolQuery,
	}
}
