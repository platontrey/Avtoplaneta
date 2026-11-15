package search

import (
	"strconv"
	"strings"

	"parts-service/internal/domain"
)

// buildQueryFilters строит структурированные фильтры по параметрам инвентаря
func buildQueryFilters(params domain.InventoryQueryParams) []map[string]interface{} {
	filter := make([]map[string]interface{}, 0, 16)

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
		clean := CleanPartNumber(params.Number)
		shouldClauses := []map[string]interface{}{
			{"term": map[string]interface{}{"number": params.Number}},
			{"term": map[string]interface{}{"number.normalized": params.Number}},
		}
		if clean != "" && clean != params.Number {
			shouldClauses = append(shouldClauses,
				map[string]interface{}{"term": map[string]interface{}{"number": clean}},
				map[string]interface{}{"term": map[string]interface{}{"number.normalized": clean}},
			)
		}
		shouldClauses = append(shouldClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"number.text": map[string]interface{}{
					"query":    params.Number,
					"operator": "AND",
				},
			},
		})
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               shouldClauses,
				"minimum_should_match": 1,
			},
		})
	}

	if params.OEMCode != "" {
		clean := CleanPartNumber(params.OEMCode)
		shouldClauses := []map[string]interface{}{
			{"term": map[string]interface{}{"oem_code": params.OEMCode}},
			{"term": map[string]interface{}{"oem_code.normalized": params.OEMCode}},
		}
		if clean != "" && clean != params.OEMCode {
			shouldClauses = append(shouldClauses,
				map[string]interface{}{"term": map[string]interface{}{"oem_code": clean}},
				map[string]interface{}{"term": map[string]interface{}{"oem_code.normalized": clean}},
			)
		}
		shouldClauses = append(shouldClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"oem_code.text": map[string]interface{}{
					"query":    params.OEMCode,
					"operator": "AND",
				},
			},
		})
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               shouldClauses,
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
		clean := CleanPartNumber(params.ManufacturerCode)
		shouldClauses := []map[string]interface{}{
			{"term": map[string]interface{}{"manufacturer_code": params.ManufacturerCode}},
			{"term": map[string]interface{}{"manufacturer_code.normalized": params.ManufacturerCode}},
		}
		if clean != "" && clean != params.ManufacturerCode {
			shouldClauses = append(shouldClauses,
				map[string]interface{}{"term": map[string]interface{}{"manufacturer_code": clean}},
				map[string]interface{}{"term": map[string]interface{}{"manufacturer_code.normalized": clean}},
			)
		}
		shouldClauses = append(shouldClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"manufacturer_code.text": map[string]interface{}{
					"query":    params.ManufacturerCode,
					"operator": "AND",
				},
			},
		})
		filter = append(filter, map[string]interface{}{
			"bool": map[string]interface{}{
				"should":               shouldClauses,
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

	return filter
}
