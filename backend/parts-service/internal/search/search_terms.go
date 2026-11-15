package search

import (
	"regexp"
	"strings"
)

var multiSearchSplitRegex = regexp.MustCompile(`(?i)(?:\s+(?:или|or)\s+|[,\n\r;|]+)`)

// SplitMultiSearchQuery разбивает строку поиска на отдельные поисковые подзапросы,
// если используются разделители (запятая, точка с запятой, перенос строки, |, ключевые слова "или" / "or").
func SplitMultiSearchQuery(raw string) []string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return nil
	}

	parts := multiSearchSplitRegex.Split(trimmed, -1)
	var result []string
	for _, p := range parts {
		clean := strings.TrimSpace(p)
		if clean != "" {
			result = append(result, clean)
		}
	}
	return result
}

var defaultSearchableFields = []string{
	"name^10", "name.ngram^5",
	"brand^4", "brand.text^4", "brand.ngram^3",
	"model^4", "model.text^4", "model.ngram^3",
	"body_brand^3", "body_brand.text^3", "body_brand.ngram^2",
	"engine_brand^3", "engine_brand.text^3", "engine_brand.ngram^2",
	"number^4", "number.normalized^5", "number.text^4",
	"oem_code^4", "oem_code.normalized^6", "oem_code.text^4",
	"manufacturer_code^3", "manufacturer_code.normalized^5", "manufacturer_code.text^3",
	"supplier_code^2", "supplier_code.text^2",
	"vin^3", "vin.text^3",
	"category^3", "category.text^3", "category.ngram^2",
	"car_release_date^3", "car_release_date.text^3", "car_release_date.ngram^2",
	"car_release_period^3", "car_release_period.text^3", "car_release_period.ngram^2",
	"front_rear^3", "front_rear.text^3", "front_rear.ngram^2",
	"left_right^3", "left_right.text^3", "left_right.ngram^2",
	"top_bottom^3", "top_bottom.text^3", "top_bottom.ngram^2",
	"color^3", "color.text^3", "color.ngram^2",
	"condition^2", "condition.text^2", "condition.ngram^1",
	"transmission^3", "transmission.text^3", "transmission.ngram^2",
	"transmission_model^3", "transmission_model.text^3", "transmission_model.ngram^2",
	"drive^3", "drive.text^3", "drive.ngram^2",
	"manufacturer^2", "manufacturer.text^2", "manufacturer.ngram^1",
	"defect^2", "defect.text^2",
	"season^3", "season.text^3", "season.ngram^2",
	"diameter^2", "diameter.text^2",
	"width^2", "width.text^2",
	"profile^2", "profile.text^2",
	"drilling^2", "drilling.text^2",
	"offset^2", "offset.text^2",
	"center_hole_diameter^2", "center_hole_diameter.text^2",
	"tire_model^3", "tire_model.text^3", "tire_model.ngram^2",
	"tire_quantity^1", "tire_quantity.text^1",
	"wear_percentage^1", "wear_percentage.text^1",
	"location^1", "location.text^1",
	"address^1", "address.text^1",
	"salesman^1", "salesman.text^1",
	"description^1",
}

func buildTermQueries(term string, searchableFields []string) []map[string]interface{} {
	variants := make([]string, 1, 3)
	variants[0] = term
	termLower := strings.ToLower(term)
	tTrans := TransliterateLatinToCyrillic(termLower)
	if tTrans != termLower {
		variants = append(variants, tTrans)
	}
	tQwerty := ConvertQwertyToRussian(termLower)
	if tQwerty != termLower && tQwerty != tTrans {
		variants = append(variants, tQwerty)
	}

	termQueries := make([]map[string]interface{}, 0, len(variants)*5)
	for _, variant := range variants {
		escapedVar := escapeESQuery(variant)

		// 1. Точное / ngram / стеммированное совпадение по всем полям
		termQueries = append(termQueries, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  variant,
				"fields": searchableFields,
				"type":   "best_fields",
			},
		})

		// 2. Префиксный поиск через match_bool_prefix (когда слово не дописано)
		termQueries = append(termQueries, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":  variant,
				"fields": searchableFields,
				"type":   "bool_prefix",
			},
		})

		// 3. Префиксный wildcard (слово*) через query_string по всем полям
		termQueries = append(termQueries, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query":            escapedVar + "*",
				"fields":           searchableFields,
				"default_operator": "OR",
				"analyze_wildcard": true,
				"boost":            2.0,
			},
		})

		// 4. Подстрочный wildcard (*слово*) для фрагментов от 3 символов
		if len([]rune(variant)) >= 3 {
			termQueries = append(termQueries, map[string]interface{}{
				"query_string": map[string]interface{}{
					"query":            "*" + escapedVar + "*",
					"fields":           searchableFields,
					"default_operator": "OR",
					"analyze_wildcard": true,
					"boost":            1.0,
				},
			})
		}

		// 5. Позиционные синонимы (F/R/L и перед/зад/право/лево/верх/низ)
		if posQueries := GetPositionTermQueries(variant); len(posQueries) > 0 {
			termQueries = append(termQueries, posQueries...)
		}
	}

	cleanTerm := CleanPartNumber(term)
	if cleanTerm != "" && len(cleanTerm) >= 3 {
		termQueries = append(termQueries,
			map[string]interface{}{
				"term": map[string]interface{}{
					"oem_code.normalized": map[string]interface{}{
						"value": cleanTerm,
						"boost": 25.0,
					},
				},
			},
			map[string]interface{}{
				"term": map[string]interface{}{
					"manufacturer_code.normalized": map[string]interface{}{
						"value": cleanTerm,
						"boost": 20.0,
					},
				},
			},
			map[string]interface{}{
				"term": map[string]interface{}{
					"number.normalized": map[string]interface{}{
						"value": cleanTerm,
						"boost": 20.0,
					},
				},
			},
		)
	}

	return termQueries
}

func buildSearchRelevanceBoosts(searchStr string, searchableFields []string) []map[string]interface{} {
	transliteratedSearch := TransliterateLatinToCyrillic(searchStr)
	qwertySearch := ConvertQwertyToRussian(searchStr)

	should := []map[string]interface{}{
		// Точное совпадение в названии (Максимальный приоритет 100.0)
		{
			"match": map[string]interface{}{
				"name": map[string]interface{}{
					"query": searchStr,
					"boost": 100.0,
				},
			},
		},
		// Фразовое совпадение с префиксом в названии
		{
			"match_phrase_prefix": map[string]interface{}{
				"name": map[string]interface{}{
					"query": searchStr,
					"boost": 50.0,
				},
			},
		},
		// Фразовое совпадение с префиксом по всей строке по всем полям
		{
			"multi_match": map[string]interface{}{
				"query":  searchStr,
				"fields": searchableFields,
				"type":   "bool_prefix",
				"boost":  40.0,
			},
		},
		// Совпадение всей поисковой фразы целиком по всем полям
		{
			"multi_match": map[string]interface{}{
				"query":  searchStr,
				"fields": searchableFields,
				"type":   "best_fields",
				"boost":  15.0,
			},
		},
	}

	cleanSearch := CleanPartNumber(searchStr)
	if cleanSearch != "" && len(cleanSearch) >= 3 {
		should = append(should,
			map[string]interface{}{
				"term": map[string]interface{}{
					"oem_code.normalized": map[string]interface{}{
						"value": cleanSearch,
						"boost": 150.0,
					},
				},
			},
			map[string]interface{}{
				"term": map[string]interface{}{
					"manufacturer_code.normalized": map[string]interface{}{
						"value": cleanSearch,
						"boost": 140.0,
					},
				},
			},
			map[string]interface{}{
				"term": map[string]interface{}{
					"number.normalized": map[string]interface{}{
						"value": cleanSearch,
						"boost": 130.0,
					},
				},
			},
		)
	}

	if transliteratedSearch != strings.ToLower(searchStr) {
		should = append(should,
			map[string]interface{}{
				"match": map[string]interface{}{
					"name": map[string]interface{}{
						"query": transliteratedSearch,
						"boost": 80.0,
					},
				},
			},
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  transliteratedSearch,
					"fields": searchableFields,
					"type":   "best_fields",
					"boost":  10.0,
				},
			},
		)
	}

	if qwertySearch != strings.ToLower(searchStr) && qwertySearch != transliteratedSearch {
		should = append(should,
			map[string]interface{}{
				"match": map[string]interface{}{
					"name": map[string]interface{}{
						"query": qwertySearch,
						"boost": 70.0,
					},
				},
			},
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  qwertySearch,
					"fields": searchableFields,
					"type":   "best_fields",
					"boost":  8.0,
				},
			},
		)
	}

	// Нечёткий поиск для компенсации опечаток
	should = append(should, map[string]interface{}{
		"multi_match": map[string]interface{}{
			"query":                searchStr,
			"fields":               []string{"name^2", "brand.text^1", "model.text^1", "car_release_date.text^1", "car_release_period.text^1", "front_rear.text^1", "color.text^1", "transmission.text^1"},
			"type":                 "best_fields",
			"fuzziness":            "AUTO:4,7",
			"prefix_length":        2,
			"minimum_should_match": "75%",
			"boost":                0.1,
		},
	})

	return should
}

// buildSingleFragmentQuery строит bool запрос для отдельного фрагмента мультипоиска
func buildSingleFragmentQuery(frag string, searchableFields []string) map[string]interface{} {
	terms := strings.Fields(frag)
	if len(terms) == 0 {
		return nil
	}

	fragMust := make([]map[string]interface{}, 0, len(terms))
	for _, term := range terms {
		if strings.TrimSpace(term) == "" {
			continue
		}
		termQueries := buildTermQueries(term, searchableFields)
		fragMust = append(fragMust, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               termQueries,
				"minimum_should_match": 1,
			},
		})
	}

	fragShould := buildSearchRelevanceBoosts(frag, searchableFields)

	fragBool := map[string]interface{}{
		"must": fragMust,
	}
	if len(fragShould) > 0 {
		fragBool["should"] = fragShould
	}

	return map[string]interface{}{
		"bool": fragBool,
	}
}

func buildMultiSearchQuery(fragments []string) map[string]interface{} {
	multiShould := make([]map[string]interface{}, 0, len(fragments))
	for _, frag := range fragments {
		if fragQuery := buildSingleFragmentQuery(frag, defaultSearchableFields); fragQuery != nil {
			multiShould = append(multiShould, fragQuery)
		}
	}
	if len(multiShould) == 0 {
		return nil
	}
	return map[string]interface{}{
		"bool": map[string]interface{}{
			"should":               multiShould,
			"minimum_should_match": 1,
		},
	}
}

func buildSingleSearchQuery(frag string) ([]map[string]interface{}, []map[string]interface{}) {
	var must []map[string]interface{}
	for _, term := range strings.Fields(frag) {
		if strings.TrimSpace(term) == "" {
			continue
		}
		termQueries := buildTermQueries(term, defaultSearchableFields)
		must = append(must, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               termQueries,
				"minimum_should_match": 1,
			},
		})
	}
	should := buildSearchRelevanceBoosts(frag, defaultSearchableFields)
	return must, should
}
