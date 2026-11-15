package search

import "parts-service/internal/domain"

// ElasticsearchClient определяет интерфейс для работы с Elasticsearch
type ElasticsearchClient interface {
	IndexPart(part *domain.Part) error
	DeletePartFromIndex(partID int64) error
	SearchParts(query map[string]interface{}, from, size int) ([]ElasticsearchPart, int64, error)
}

// elasticsearchAdapter адаптер для Elasticsearch, реализующий ElasticsearchClient
type elasticsearchAdapter struct{}

// NewElasticsearchAdapter создает новый адаптер для Elasticsearch
func NewElasticsearchAdapter() ElasticsearchClient {
	return &elasticsearchAdapter{}
}

func (e *elasticsearchAdapter) IndexPart(part *domain.Part) error {
	return IndexPart(part)
}

func (e *elasticsearchAdapter) DeletePartFromIndex(partID int64) error {
	return DeletePartFromIndex(partID)
}

func (e *elasticsearchAdapter) SearchParts(query map[string]interface{}, from, size int) ([]ElasticsearchPart, int64, error) {
	return SearchParts(query, from, size)
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
	CarReleasePeriod   string   `json:"car_release_period,omitempty"`
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
