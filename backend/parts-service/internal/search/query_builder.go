package search

import (
	"strconv"
	"strings"

	"parts-service/internal/catalog"
	"parts-service/internal/domain"
)

// TransliterateLatinToCyrillic преобразует транслит латиницы в кириллицу (sirena -> сирена, bamper -> бампер)
func TransliterateLatinToCyrillic(text string) string {
	text = strings.ToLower(text)
	replacements := []struct {
		from string
		to   string
	}{
		{"shch", "щ"}, {"sh", "ш"}, {"ch", "ч"}, {"zh", "ж"},
		{"ya", "я"}, {"yu", "ю"}, {"yo", "ё"}, {"ts", "ц"},
		{"a", "а"}, {"b", "б"}, {"v", "в"}, {"g", "г"}, {"d", "д"},
		{"e", "е"}, {"z", "з"}, {"i", "и"}, {"j", "й"}, {"k", "к"},
		{"l", "л"}, {"m", "м"}, {"n", "н"}, {"o", "о"}, {"p", "п"},
		{"r", "р"}, {"s", "с"}, {"t", "т"}, {"u", "у"}, {"f", "ф"},
		{"h", "х"}, {"c", "к"}, {"y", "ы"}, {"w", "в"}, {"x", "кс"},
	}

	res := text
	for _, r := range replacements {
		res = strings.ReplaceAll(res, r.from, r.to)
	}
	return res
}

// ConvertQwertyToRussian переводит текст с неверной QWERTY раскладки на кириллицу (gthtlybq -> передний)
func ConvertQwertyToRussian(text string) string {
	qwertyMap := map[rune]rune{
		'q': 'й', 'w': 'ц', 'e': 'у', 'r': 'к', 't': 'е', 'y': 'н', 'u': 'г', 'i': 'ш', 'o': 'щ', 'p': 'з', '[': 'х', ']': 'ъ',
		'a': 'ф', 's': 'ы', 'd': 'в', 'f': 'а', 'g': 'п', 'h': 'р', 'j': 'о', 'k': 'л', 'l': 'д', ';': 'ж', '\'': 'э',
		'z': 'я', 'x': 'ч', 'c': 'с', 'v': 'м', 'b': 'и', 'n': 'т', 'm': 'ь', ',': 'б', '.': 'ю',
	}

	var builder strings.Builder
	for _, char := range strings.ToLower(text) {
		if ruChar, ok := qwertyMap[char]; ok {
			builder.WriteRune(ruChar)
		} else {
			builder.WriteRune(char)
		}
	}
	return builder.String()
}

// escapeESQuery экранирует спецсимволы синтаксиса Lucene для безопасного использования в query_string
func escapeESQuery(s string) string {
	var sb strings.Builder
	for _, r := range s {
		switch r {
		case '+', '-', '=', '!', '(', ')', '{', '}', '[', ']', '^', '"', '~', '*', '?', ':', '\\', '/', '&', '|', '<', '>':
			sb.WriteRune('\\')
			sb.WriteRune(r)
		default:
			sb.WriteRune(r)
		}
	}
	return sb.String()
}

// ExpandFrontRearSynonyms разворачивает любое обозначение положения перед/зад (русское, английское, F/R)
func ExpandFrontRearSynonyms(val string) []string {
	return catalog.ExpandFrontRearSynonyms(val)
}

// ExpandLeftRightSynonyms разворачивает любое обозначение стороны право/лево (русское, английское, L/R)
func ExpandLeftRightSynonyms(val string) []string {
	return catalog.ExpandLeftRightSynonyms(val)
}

// ExpandTopBottomSynonyms разворачивает любое обозначение вертикального положения верх/низ
func ExpandTopBottomSynonyms(val string) []string {
	return catalog.ExpandTopBottomSynonyms(val)
}

// ExpandCategorySynonyms разворачивает категорию в синонимы и подкатегории.
func ExpandCategorySynonyms(val string) []string {
	return catalog.ExpandCategorySynonyms(val)
}

// GetPositionTermQueries проверяет, является ли терм маркером расположения/стороны,
// и возвращает запросы к колонкам front_rear, left_right, top_bottom.
func GetPositionTermQueries(term string) []map[string]interface{} {
	clean := strings.ToLower(strings.TrimSpace(term))
	var queries []map[string]interface{}

	// Проверка на Перед: "перед", "передний", "передняя", "переднее", "передние", "front"
	if strings.HasPrefix(clean, "перед") || clean == "front" {
		synonyms := []interface{}{"F", "f", "Front", "front", "перед", "передний", "передняя", "переднее", "перед / зад"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"front_rear": synonyms,
			},
		})
	}

	// Проверка на Зад: "зад", "задний", "задняя", "заднее", "задние", "rear"
	if strings.HasPrefix(clean, "зад") || clean == "rear" {
		synonyms := []interface{}{"R", "r", "Rear", "rear", "зад", "задний", "задняя", "заднее", "перед / зад"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"front_rear": synonyms,
			},
		})
	}

	// Проверка на Право: "прав", "правый", "правая", "правое", "правые", "право", "right"
	if strings.HasPrefix(clean, "прав") || clean == "right" {
		synonyms := []interface{}{"R", "r", "Right", "right", "право", "правый", "правая", "правое", "лево / право"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"left_right": synonyms,
			},
		})
	}

	// Проверка на Лево: "лев", "левый", "левая", "левое", "левые", "лево", "left"
	if strings.HasPrefix(clean, "лев") || clean == "left" {
		synonyms := []interface{}{"L", "l", "Left", "left", "лево", "левый", "левая", "левое", "лево / право"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"left_right": synonyms,
			},
		})
	}

	// Проверка на Верх: "верх", "верхний", "верхняя", "верхнее", "top", "upper"
	if strings.HasPrefix(clean, "верх") || clean == "top" || clean == "upper" {
		synonyms := []interface{}{"T", "t", "U", "u", "Top", "top", "Upper", "upper", "верх", "верхний", "верхняя", "верх / низ"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"top_bottom": synonyms,
			},
		})
	}

	// Проверка на Низ: "низ", "нижний", "нижняя", "нижнее", "bottom", "lower"
	if strings.HasPrefix(clean, "низ") || clean == "bottom" || clean == "lower" {
		synonyms := []interface{}{"B", "b", "L", "l", "Bottom", "bottom", "Lower", "lower", "низ", "нижний", "нижняя", "верх / низ"}
		queries = append(queries, map[string]interface{}{
			"terms": map[string]interface{}{
				"top_bottom": synonyms,
			},
		})
	}

	return queries
}

// buildElasticsearchQuery строит запрос для Elasticsearch
func BuildElasticsearchQuery(params domain.InventoryQueryParams) map[string]interface{} {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{}

	should := []map[string]interface{}{}

	// Фильтр для отображения валидных запчастей (quantity >= 0)
	filter = append(filter, map[string]interface{}{
		"range": map[string]interface{}{
			"quantity": map[string]interface{}{
				"gte": 0,
			},
		},
	})

	if params.Search != "" {
		terms := strings.Fields(params.Search)
		transliteratedSearch := TransliterateLatinToCyrillic(params.Search)
		qwertySearch := ConvertQwertyToRussian(params.Search)

		searchableFields := []string{
			"name^10", "name.ngram^5",
			"brand^4", "brand.text^4", "brand.ngram^3",
			"model^4", "model.text^4", "model.ngram^3",
			"body_brand^3", "body_brand.text^3", "body_brand.ngram^2",
			"engine_brand^3", "engine_brand.text^3", "engine_brand.ngram^2",
			"number^4", "number.text^4",
			"oem_code^4", "oem_code.text^4",
			"manufacturer_code^3", "manufacturer_code.text^3",
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

		// 1. Обязательное совпадение: каждый терм поискового запроса должен присутствовать
		// в запчасти (по любому из полей или как префикс не завершенного слова).
		for _, term := range terms {
			if strings.TrimSpace(term) == "" {
				continue
			}

			variants := []string{term}
			tTrans := TransliterateLatinToCyrillic(term)
			if tTrans != strings.ToLower(term) {
				variants = append(variants, tTrans)
			}
			tQwerty := ConvertQwertyToRussian(term)
			if tQwerty != strings.ToLower(term) && tQwerty != tTrans {
				variants = append(variants, tQwerty)
			}

			termQueries := []map[string]interface{}{}
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

			must = append(must, map[string]interface{}{
				"bool": map[string]interface{}{
					"should":               termQueries,
					"minimum_should_match": 1,
				},
			})
		}

		// 2. Для ранжирования и релевантности добавляем should-запросы
		// с высокими весами для точного совпадения названия, фразового поиска и опечаток
		should = append(should,
			// Точное совпадение в названии (Максимальный приоритет 100.0)
			map[string]interface{}{
				"match": map[string]interface{}{
					"name": map[string]interface{}{
						"query": params.Search,
						"boost": 100.0,
					},
				},
			},
			// Фразовое совпадение с префиксом в названии
			map[string]interface{}{
				"match_phrase_prefix": map[string]interface{}{
					"name": map[string]interface{}{
						"query": params.Search,
						"boost": 50.0,
					},
				},
			},
			// Фразовое совпадение с префиксом по всей строке по всем полям
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  params.Search,
					"fields": searchableFields,
					"type":   "bool_prefix",
					"boost":  40.0,
				},
			},
			// Совпадение всей поисковой фразы целиком по всем полям
			map[string]interface{}{
				"multi_match": map[string]interface{}{
					"query":  params.Search,
					"fields": searchableFields,
					"type":   "best_fields",
					"boost":  15.0,
				},
			},
		)

		if transliteratedSearch != strings.ToLower(params.Search) {
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

		if qwertySearch != strings.ToLower(params.Search) && qwertySearch != transliteratedSearch {
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
				"query":                params.Search,
				"fields":               []string{"name^2", "brand.text^1", "model.text^1", "car_release_date.text^1", "car_release_period.text^1", "front_rear.text^1", "color.text^1", "transmission.text^1"},
				"type":                 "best_fields",
				"fuzziness":            "AUTO:4,7",
				"prefix_length":        2,
				"minimum_should_match": "75%",
				"boost":                0.1,
			},
		})
	}

	if params.Category != "" {
		variants := ExpandCategorySynonyms(params.Category)
		var catShould []map[string]interface{}
		for _, v := range variants {
			catShould = append(catShould,
				map[string]interface{}{
					"term": map[string]interface{}{
						"category": map[string]interface{}{
							"value":            v,
							"case_insensitive": true,
						},
					},
				},
				map[string]interface{}{
					"match": map[string]interface{}{
						"category.text": v,
					},
				},
				map[string]interface{}{
					"match": map[string]interface{}{
						"category.ngram": v,
					},
				},
				map[string]interface{}{
					"wildcard": map[string]interface{}{
						"category": map[string]interface{}{
							"value":            "*" + strings.ToLower(v) + "*",
							"case_insensitive": true,
						},
					},
				},
			)
		}
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               catShould,
				"minimum_should_match": 1,
			},
		})
	}

	if params.HasPhoto != "" && params.HasPhoto != "all" {
		if params.HasPhoto == "with" {
			filter = append(filter, map[string]interface{}{
				"exists": map[string]interface{}{
					"field": "photos",
				},
			})
		} else if params.HasPhoto == "without" {
			filter = append(filter, map[string]interface{}{
				"bool": map[string]interface{}{
					"must_not": []map[string]interface{}{
						{
							"exists": map[string]interface{}{
								"field": "photos",
							},
						},
					},
				},
			})
		}
	}

	if params.Brand != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"brand": params.Brand}},
					{"match": map[string]interface{}{"brand.text": params.Brand}},
					{"match": map[string]interface{}{"brand.ngram": params.Brand}},
					{"wildcard": map[string]interface{}{"brand": "*" + strings.ToLower(params.Brand) + "*"}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Model != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"model": params.Model}},
					{"match": map[string]interface{}{"model.text": params.Model}},
					{"match": map[string]interface{}{"model.ngram": params.Model}},
					{"wildcard": map[string]interface{}{"model": "*" + strings.ToLower(params.Model) + "*"}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Location != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"location.text": params.Location,
			},
		})
	}

	if params.Address != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"address.text": params.Address,
			},
		})
	}

	if params.Salesman != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"salesman.text": params.Salesman,
			},
		})
	}

	if params.Status != "" {
		statusBool := params.Status == "true" || params.Status == "active" || params.Status == "1"
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{
				"status": statusBool,
			},
		})
	}

	if params.Number != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"number": params.Number}},
					{"match": map[string]interface{}{"number.text": params.Number}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.OEMCode != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"oem_code": params.OEMCode}},
					{"match": map[string]interface{}{"oem_code.text": params.OEMCode}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.VIN != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"vin": params.VIN}},
					{"match": map[string]interface{}{"vin.text": params.VIN}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.BodyBrand != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"body_brand.text": params.BodyBrand,
			},
		})
	}

	if params.EngineBrand != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"engine_brand.text": params.EngineBrand,
			},
		})
	}

	if params.CarReleaseDate != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"car_release_date": params.CarReleaseDate}},
					{"match": map[string]interface{}{"car_release_date.text": params.CarReleaseDate}},
					{"match": map[string]interface{}{"car_release_date.ngram": params.CarReleaseDate}},
					{"wildcard": map[string]interface{}{"car_release_date": "*" + params.CarReleaseDate + "*"}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.CarReleasePeriod != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"car_release_period": params.CarReleasePeriod}},
					{"match": map[string]interface{}{"car_release_period.text": params.CarReleasePeriod}},
					{"match": map[string]interface{}{"car_release_period.ngram": params.CarReleasePeriod}},
					{"wildcard": map[string]interface{}{"car_release_period": "*" + params.CarReleasePeriod + "*"}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Transmission != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"transmission": params.Transmission}},
					{"match": map[string]interface{}{"transmission.text": params.Transmission}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Drive != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"drive": params.Drive}},
					{"match": map[string]interface{}{"drive.text": params.Drive}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Condition != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"condition": params.Condition,
			},
		})
	}

	if params.Manufacturer != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"manufacturer.text": params.Manufacturer,
			},
		})
	}

	if params.Defect != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"defect": params.Defect,
			},
		})
	}

	if params.Color != "" {
		filter = append(filter, map[string]interface{}{
			"match": map[string]interface{}{
				"color.text": params.Color,
			},
		})
	}

	if params.MinPrice != "" || params.MaxPrice != "" {
		priceRange := map[string]interface{}{}
		if params.MinPrice != "" {
			if minP, err := strconv.ParseFloat(params.MinPrice, 64); err == nil {
				priceRange["gte"] = minP
			}
		}
		if params.MaxPrice != "" {
			if maxP, err := strconv.ParseFloat(params.MaxPrice, 64); err == nil {
				priceRange["lte"] = maxP
			}
		}
		if len(priceRange) > 0 {
			filter = append(filter, map[string]interface{}{
				"range": map[string]interface{}{
					"price": priceRange,
				},
			})
		}
	}

	if params.MinQuantity != "" || params.MaxQuantity != "" {
		qtyRange := map[string]interface{}{}
		if params.MinQuantity != "" {
			if minQ, err := strconv.Atoi(params.MinQuantity); err == nil {
				qtyRange["gte"] = minQ
			}
		}
		if params.MaxQuantity != "" {
			if maxQ, err := strconv.Atoi(params.MaxQuantity); err == nil {
				qtyRange["lte"] = maxQ
			}
		}
		if len(qtyRange) > 0 {
			filter = append(filter, map[string]interface{}{
				"range": map[string]interface{}{
					"quantity": qtyRange,
				},
			})
		}
	}

	if params.FrontRear != "" {
		synonyms := ExpandFrontRearSynonyms(params.FrontRear)
		synInterfaces := make([]interface{}, len(synonyms))
		for i, v := range synonyms {
			synInterfaces[i] = v
		}
		shouldClauses := []map[string]interface{}{
			{"terms": map[string]interface{}{"front_rear": synInterfaces}},
		}
		for _, syn := range synonyms {
			shouldClauses = append(shouldClauses, map[string]interface{}{
				"match": map[string]interface{}{"front_rear.text": syn},
			})
		}
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		})
	}

	if params.LeftRight != "" {
		synonyms := ExpandLeftRightSynonyms(params.LeftRight)
		synInterfaces := make([]interface{}, len(synonyms))
		for i, v := range synonyms {
			synInterfaces[i] = v
		}
		shouldClauses := []map[string]interface{}{
			{"terms": map[string]interface{}{"left_right": synInterfaces}},
		}
		for _, syn := range synonyms {
			shouldClauses = append(shouldClauses, map[string]interface{}{
				"match": map[string]interface{}{"left_right.text": syn},
			})
		}
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		})
	}

	if params.TopBottom != "" {
		synonyms := ExpandTopBottomSynonyms(params.TopBottom)
		synInterfaces := make([]interface{}, len(synonyms))
		for i, v := range synonyms {
			synInterfaces[i] = v
		}
		shouldClauses := []map[string]interface{}{
			{"terms": map[string]interface{}{"top_bottom": synInterfaces}},
		}
		for _, syn := range synonyms {
			shouldClauses = append(shouldClauses, map[string]interface{}{
				"match": map[string]interface{}{"top_bottom.text": syn},
			})
		}
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		})
	}

	if params.ManufacturerCode != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"manufacturer_code": params.ManufacturerCode}},
					{"match": map[string]interface{}{"manufacturer_code.text": params.ManufacturerCode}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.SupplierCode != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"supplier_code": params.SupplierCode}},
					{"match": map[string]interface{}{"supplier_code.text": params.SupplierCode}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.TransmissionModel != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"transmission_model": params.TransmissionModel}},
					{"match": map[string]interface{}{"transmission_model.text": params.TransmissionModel}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.WearPercentage != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"wear_percentage": params.WearPercentage}},
					{"match": map[string]interface{}{"wear_percentage.text": params.WearPercentage}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Season != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"season": params.Season}},
					{"match": map[string]interface{}{"season.text": params.Season}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Diameter != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"diameter": params.Diameter}},
					{"match": map[string]interface{}{"diameter.text": params.Diameter}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Width != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"width": params.Width}},
					{"match": map[string]interface{}{"width.text": params.Width}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Profile != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"profile": params.Profile}},
					{"match": map[string]interface{}{"profile.text": params.Profile}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.TireQuantity != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"tire_quantity": params.TireQuantity}},
					{"match": map[string]interface{}{"tire_quantity.text": params.TireQuantity}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Drilling != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"drilling": params.Drilling}},
					{"match": map[string]interface{}{"drilling.text": params.Drilling}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.Offset != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"offset": params.Offset}},
					{"match": map[string]interface{}{"offset.text": params.Offset}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.CenterHoleDiameter != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"center_hole_diameter": params.CenterHoleDiameter}},
					{"match": map[string]interface{}{"center_hole_diameter.text": params.CenterHoleDiameter}},
				},
				"minimum_should_match": 1,
			},
		})
	}

	if params.TireModel != "" {
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should": []map[string]interface{}{
					{"term": map[string]interface{}{"tire_model": params.TireModel}},
					{"match": map[string]interface{}{"tire_model.text": params.TireModel}},
				},
				"minimum_should_match": 1,
			},
		})
	}

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
