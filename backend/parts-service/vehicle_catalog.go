package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed vehicles/vehicles.json
var embeddedVehicleCatalog []byte

// VehicleModel — модель в рамках марки. Slug хранится, чтобы парсер мог
// сопоставлять записи между запусками, а клиенты — строить ссылки.
type VehicleModel struct {
	Name string `json:"name"`
	Slug string `json:"slug,omitempty"`
}

type VehicleBrand struct {
	Name   string         `json:"name"`
	Slug   string         `json:"slug,omitempty"`
	Models []VehicleModel `json:"models"`
}

// VehicleCatalog — справочник марок и моделей автомобилей.
//
// Это отдельный от catalog.json справочник: у него другой жизненный цикл
// (его перезаписывает cmd/vehicle-catalog-sync, а не человек) и своя версия,
// поэтому обновление марок не сбрасывает клиентский кэш каталога запчастей.
type VehicleCatalog struct {
	Version     string         `json:"version"`
	Source      string         `json:"source,omitempty"`
	GeneratedAt string         `json:"generated_at,omitempty"`
	Brands      []VehicleBrand `json:"brands"`
}

func LoadVehicleCatalog() (*VehicleCatalog, error) {
	var catalog VehicleCatalog
	if err := json.Unmarshal(embeddedVehicleCatalog, &catalog); err != nil {
		return nil, fmt.Errorf("decode embedded vehicle catalog: %w", err)
	}
	if err := validateVehicleCatalog(&catalog); err != nil {
		return nil, err
	}
	return &catalog, nil
}

func validateVehicleCatalog(catalog *VehicleCatalog) error {
	if strings.TrimSpace(catalog.Version) == "" {
		return fmt.Errorf("vehicle catalog version is empty")
	}
	if len(catalog.Brands) == 0 {
		return fmt.Errorf("vehicle catalog contains no brands")
	}

	seenBrands := make(map[string]struct{}, len(catalog.Brands))
	for _, brand := range catalog.Brands {
		name := strings.TrimSpace(brand.Name)
		if name == "" {
			return fmt.Errorf("vehicle catalog contains a brand without a name")
		}
		key := strings.ToLower(name)
		if _, exists := seenBrands[key]; exists {
			return fmt.Errorf("vehicle catalog contains duplicate brand %q", name)
		}
		seenBrands[key] = struct{}{}

		seenModels := make(map[string]struct{}, len(brand.Models))
		for _, model := range brand.Models {
			modelName := strings.TrimSpace(model.Name)
			if modelName == "" {
				return fmt.Errorf("brand %q contains a model without a name", name)
			}
			modelKey := strings.ToLower(modelName)
			if _, exists := seenModels[modelKey]; exists {
				return fmt.Errorf("brand %q contains duplicate model %q", name, modelName)
			}
			seenModels[modelKey] = struct{}{}
		}

		if !sort.SliceIsSorted(brand.Models, func(i, j int) bool {
			return vehicleSortKey(brand.Models[i].Name) < vehicleSortKey(brand.Models[j].Name)
		}) {
			return fmt.Errorf("models of brand %q are not sorted", name)
		}
	}

	if !sort.SliceIsSorted(catalog.Brands, func(i, j int) bool {
		return vehicleSortKey(catalog.Brands[i].Name) < vehicleSortKey(catalog.Brands[j].Name)
	}) {
		return fmt.Errorf("vehicle catalog brands are not sorted")
	}

	return nil
}

// vehicleSortKey задаёт один порядок сортировки для парсера, валидатора и
// клиентов, чтобы дифф vehicles.json между запусками оставался пустым,
// когда данные не поменялись.
func vehicleSortKey(name string) string {
	return strings.ToLower(strings.TrimSpace(name))
}

// ModelsOf возвращает модели марки. Марка ищется без учёта регистра, потому что
// в уже заведённых запчастях марка лежит как свободный текст.
func (c *VehicleCatalog) ModelsOf(brand string) []VehicleModel {
	if c == nil {
		return nil
	}
	needle := vehicleSortKey(brand)
	for _, item := range c.Brands {
		if vehicleSortKey(item.Name) == needle {
			return item.Models
		}
	}
	return nil
}
