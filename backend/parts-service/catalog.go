package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

//go:embed catalog/catalog.json
var embeddedPartCatalog []byte

//go:embed catalog/synonyms.txt
var embeddedSynonyms []byte

// LoadSynonyms загружает синонимы запчастей из файла catalog/synonyms.txt
func LoadSynonyms() []string {
	lines := strings.Split(string(embeddedSynonyms), "\n")
	var synonyms []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed != "" && !strings.HasPrefix(trimmed, "#") {
			synonyms = append(synonyms, trimmed)
		}
	}
	return synonyms
}

type CatalogAttribute struct {
	Code      string   `json:"code"`
	Label     string   `json:"label"`
	InputType string   `json:"input_type"`
	Options   []string `json:"options,omitempty"`
	SortOrder int      `json:"sort_order"`
}

type CatalogPartFormCategory struct {
	Code       string   `json:"code"`
	Name       string   `json:"name"`
	Attributes []string `json:"attributes"`
}

type CatalogReportBinding struct {
	Source             string   `json:"source"`
	Target             string   `json:"target"`
	Transform          string   `json:"transform,omitempty"`
	Categories         []string `json:"categories,omitempty"`
	ExcludedCategories []string `json:"excluded_categories,omitempty"`
	DefaultValue       string   `json:"default_value,omitempty"`
}

type CatalogPartTemplate struct {
	ID       string            `json:"id"`
	Name     string            `json:"name"`
	Category string            `json:"category"`
	Quantity int               `json:"quantity"`
	Price    float64           `json:"price"`
	Defaults map[string]string `json:"defaults,omitempty"`
}

type PartCatalog struct {
	Version            string                    `json:"version"`
	Attributes         []CatalogAttribute        `json:"attributes"`
	PartFormCategories []CatalogPartFormCategory `json:"part_form_categories"`
	ReportBindings     []CatalogReportBinding    `json:"report_bindings"`
	Parts              []CatalogPartTemplate     `json:"parts"`
}

type DefectReportPart struct {
	Name               string  `json:"name"`
	Category           string  `json:"category"`
	Description        string  `json:"description"`
	Quantity           int     `json:"quantity"`
	Price              float64 `json:"price"`
	BodyBrand          string  `json:"body_brand,omitempty"`
	EngineBrand        string  `json:"engine_brand,omitempty"`
	CarReleaseDate     string  `json:"car_release_date,omitempty"`
	CarReleasePeriod   string  `json:"car_release_period,omitempty"`
	FrontRear          string  `json:"front_rear,omitempty"`
	LeftRight          string  `json:"left_right,omitempty"`
	TopBottom          string  `json:"top_bottom,omitempty"`
	Number             string  `json:"number,omitempty"`
	Manufacturer       string  `json:"manufacturer,omitempty"`
	ManufacturerCode   string  `json:"manufacturer_code,omitempty"`
	OEMCode            string  `json:"oem_code,omitempty"`
	Color              string  `json:"color,omitempty"`
	Condition          string  `json:"condition,omitempty"`
	SupplierCode       string  `json:"supplier_code,omitempty"`
	Defect             string  `json:"defect,omitempty"`
	Transmission       string  `json:"transmission,omitempty"`
	TransmissionModel  string  `json:"transmission_model,omitempty"`
	Drive              string  `json:"drive,omitempty"`
	WearPercentage     string  `json:"wear_percentage,omitempty"`
	Season             string  `json:"season,omitempty"`
	Diameter           string  `json:"diameter,omitempty"`
	Width              string  `json:"width,omitempty"`
	Profile            string  `json:"profile,omitempty"`
	TireQuantity       string  `json:"tire_quantity,omitempty"`
	Drilling           string  `json:"drilling,omitempty"`
	Offset             string  `json:"offset,omitempty"`
	CenterHoleDiameter string  `json:"center_hole_diameter,omitempty"`
	TireModel          string  `json:"tire_model,omitempty"`
	VIN                string  `json:"vin,omitempty"`
}

type DefectReportRequest struct {
	Brand             string             `json:"brand"`
	Model             string             `json:"model"`
	Year              int                `json:"year"`
	CarReleasePeriod  string             `json:"car_release_period,omitempty"`
	VIN               string             `json:"vin"`
	Mileage           int                `json:"mileage"`
	Description       string             `json:"description"`
	EngineBrand       string             `json:"engine_brand"`
	BodyBrand         string             `json:"body_brand"`
	InteriorColor     string             `json:"interior_color"`
	BodyColor         string             `json:"body_color"`
	Transmission      string             `json:"transmission"`
	TransmissionModel string             `json:"transmission_model"`
	Drive             string             `json:"drive"`
	CatalogVersion    string             `json:"catalog_version,omitempty"`
	SellerID          int64              `json:"seller_id"`
	SellerName        string             `json:"seller_name"`
	SelectedParts     []DefectReportPart `json:"selectedParts,omitempty"`
}

func LoadPartCatalog() (*PartCatalog, error) {
	var catalog PartCatalog
	if err := json.Unmarshal(embeddedPartCatalog, &catalog); err != nil {
		return nil, fmt.Errorf("decode embedded part catalog: %w", err)
	}
	if err := validatePartCatalog(&catalog); err != nil {
		return nil, err
	}
	return &catalog, nil
}

func validatePartCatalog(catalog *PartCatalog) error {
	if strings.TrimSpace(catalog.Version) == "" {
		return fmt.Errorf("part catalog version is empty")
	}
	if len(catalog.Parts) == 0 {
		return fmt.Errorf("part catalog contains no templates")
	}

	attributeCodes := make(map[string]struct{}, len(catalog.Attributes))
	for _, attribute := range catalog.Attributes {
		if attribute.Code == "" {
			return fmt.Errorf("part catalog contains an attribute with an empty code")
		}
		if _, exists := attributeCodes[attribute.Code]; exists {
			return fmt.Errorf("part catalog contains duplicate attribute %q", attribute.Code)
		}
		if attribute.InputType == "select" && len(attribute.Options) == 0 {
			return fmt.Errorf("select attribute %q contains no options", attribute.Code)
		}
		attributeCodes[attribute.Code] = struct{}{}
	}

	formCategories := make(map[string]struct{}, len(catalog.PartFormCategories))
	for _, category := range catalog.PartFormCategories {
		if category.Name == "" {
			return fmt.Errorf("part catalog contains a form category with an empty name")
		}
		if _, exists := formCategories[category.Name]; exists {
			return fmt.Errorf("part catalog contains duplicate form category %q", category.Name)
		}
		formCategories[category.Name] = struct{}{}
		for _, attribute := range category.Attributes {
			if _, exists := attributeCodes[attribute]; !exists {
				return fmt.Errorf("form category %q references unknown attribute %q", category.Name, attribute)
			}
		}
	}

	partIDs := make(map[string]struct{}, len(catalog.Parts))
	partCategories := make(map[string]struct{})
	for _, part := range catalog.Parts {
		if part.ID == "" || part.Name == "" || part.Category == "" {
			return fmt.Errorf("part template must have id, name and category: %+v", part)
		}
		if _, exists := partIDs[part.ID]; exists {
			return fmt.Errorf("part catalog contains duplicate template id %q", part.ID)
		}
		partIDs[part.ID] = struct{}{}
		partCategories[part.Category] = struct{}{}
		for field := range part.Defaults {
			if !isDefectPartField(field) {
				return fmt.Errorf("part template %q contains unknown default field %q", part.ID, field)
			}
		}
	}

	for _, binding := range catalog.ReportBindings {
		if binding.Source == "" || !isDefectPartField(binding.Target) {
			return fmt.Errorf("invalid report binding %q -> %q", binding.Source, binding.Target)
		}
		for _, category := range append(binding.Categories, binding.ExcludedCategories...) {
			if _, exists := partCategories[category]; !exists {
				return fmt.Errorf("report binding %q references unknown part category %q", binding.Source, category)
			}
		}
	}

	return nil
}

func isDefectPartField(field string) bool {
	switch field {
	case "body_brand", "engine_brand", "car_release_date", "car_release_period", "front_rear", "left_right", "top_bottom",
		"number", "manufacturer", "manufacturer_code", "oem_code", "color", "condition", "supplier_code",
		"defect", "transmission", "transmission_model", "drive", "wear_percentage", "season", "diameter",
		"width", "profile", "tire_quantity", "drilling", "offset", "center_hole_diameter", "tire_model", "vin":
		return true
	default:
		return false
	}
}

func (catalog *PartCatalog) ExpandDefectReport(report DefectReportRequest) []DefectReportPart {
	yearStr := ""
	if report.Year > 0 {
		yearStr = strconv.Itoa(report.Year)
	}
	values := map[string]string{
		"body_brand":         strings.TrimSpace(report.BodyBrand),
		"engine_brand":       strings.TrimSpace(report.EngineBrand),
		"year":               yearStr,
		"car_release_period": strings.TrimSpace(report.CarReleasePeriod),
		"vin":                strings.TrimSpace(report.VIN),
		"transmission":       strings.TrimSpace(report.Transmission),
		"transmission_model": strings.TrimSpace(report.TransmissionModel),
		"drive":              strings.TrimSpace(report.Drive),
		"interior_color":     strings.TrimSpace(report.InteriorColor),
		"body_color":         strings.TrimSpace(report.BodyColor),
	}
	supplierCode := strconv.FormatInt(time.Now().UnixMilli(), 10)
	parts := make([]DefectReportPart, 0, len(catalog.Parts))

	for _, template := range catalog.Parts {
		part := DefectReportPart{
			Name:         template.Name,
			Category:     template.Category,
			Description:  "",
			Quantity:     template.Quantity,
			Price:        template.Price,
			SupplierCode: supplierCode,
		}

		for field, value := range template.Defaults {
			setDefectPartField(&part, field, value)
		}
		for _, binding := range catalog.ReportBindings {
			if !bindingApplies(binding, template.Category) {
				continue
			}
			value := values[binding.Source]
			if value == "" {
				value = binding.DefaultValue
			}
			if value != "" {
				setDefectPartField(&part, binding.Target, value)
			}
		}
		parts = append(parts, part)
	}

	return parts
}

func (catalog *PartCatalog) ApplyBindingsToParts(parts []DefectReportPart, report DefectReportRequest) {
	yearStr := ""
	if report.Year > 0 {
		yearStr = strconv.Itoa(report.Year)
	}
	values := map[string]string{
		"body_brand":         strings.TrimSpace(report.BodyBrand),
		"engine_brand":       strings.TrimSpace(report.EngineBrand),
		"year":               yearStr,
		"car_release_period": strings.TrimSpace(report.CarReleasePeriod),
		"vin":                strings.TrimSpace(report.VIN),
		"transmission":       strings.TrimSpace(report.Transmission),
		"transmission_model": strings.TrimSpace(report.TransmissionModel),
		"drive":              strings.TrimSpace(report.Drive),
		"interior_color":     strings.TrimSpace(report.InteriorColor),
		"body_color":         strings.TrimSpace(report.BodyColor),
	}

	for i := range parts {
		for _, binding := range catalog.ReportBindings {
			if !bindingApplies(binding, parts[i].Category) {
				continue
			}
			value := values[binding.Source]
			if value == "" {
				value = binding.DefaultValue
			}
			if value != "" && strings.TrimSpace(getDefectPartField(&parts[i], binding.Target)) == "" {
				setDefectPartField(&parts[i], binding.Target, value)
			}
		}
	}
}

func bindingApplies(binding CatalogReportBinding, category string) bool {
	if len(binding.Categories) > 0 && !containsString(binding.Categories, category) {
		return false
	}
	return !containsString(binding.ExcludedCategories, category)
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func setDefectPartField(part *DefectReportPart, field, value string) {
	switch field {
	case "body_brand":
		part.BodyBrand = value
	case "engine_brand":
		part.EngineBrand = value
	case "car_release_date":
		part.CarReleaseDate = value
	case "car_release_period":
		part.CarReleasePeriod = value
	case "front_rear":
		part.FrontRear = value
	case "left_right":
		part.LeftRight = value
	case "top_bottom":
		part.TopBottom = value
	case "number":
		part.Number = value
	case "manufacturer":
		part.Manufacturer = value
	case "manufacturer_code":
		part.ManufacturerCode = value
	case "oem_code":
		part.OEMCode = value
	case "color":
		part.Color = value
	case "condition":
		part.Condition = value
	case "supplier_code":
		part.SupplierCode = value
	case "defect":
		part.Defect = value
	case "transmission":
		part.Transmission = value
	case "transmission_model":
		part.TransmissionModel = value
	case "drive":
		part.Drive = value
	case "wear_percentage":
		part.WearPercentage = value
	case "season":
		part.Season = value
	case "diameter":
		part.Diameter = value
	case "width":
		part.Width = value
	case "profile":
		part.Profile = value
	case "tire_quantity":
		part.TireQuantity = value
	case "drilling":
		part.Drilling = value
	case "offset":
		part.Offset = value
	case "center_hole_diameter":
		part.CenterHoleDiameter = value
	case "tire_model":
		part.TireModel = value
	case "vin":
		part.VIN = value
	}
}

func getDefectPartField(part *DefectReportPart, field string) string {
	switch field {
	case "body_brand":
		return part.BodyBrand
	case "engine_brand":
		return part.EngineBrand
	case "car_release_date":
		return part.CarReleaseDate
	case "car_release_period":
		return part.CarReleasePeriod
	case "front_rear":
		return part.FrontRear
	case "left_right":
		return part.LeftRight
	case "top_bottom":
		return part.TopBottom
	case "number":
		return part.Number
	case "manufacturer":
		return part.Manufacturer
	case "manufacturer_code":
		return part.ManufacturerCode
	case "oem_code":
		return part.OEMCode
	case "color":
		return part.Color
	case "condition":
		return part.Condition
	case "supplier_code":
		return part.SupplierCode
	case "defect":
		return part.Defect
	case "transmission":
		return part.Transmission
	case "transmission_model":
		return part.TransmissionModel
	case "drive":
		return part.Drive
	case "wear_percentage":
		return part.WearPercentage
	case "season":
		return part.Season
	case "diameter":
		return part.Diameter
	case "width":
		return part.Width
	case "profile":
		return part.Profile
	case "tire_quantity":
		return part.TireQuantity
	case "drilling":
		return part.Drilling
	case "offset":
		return part.Offset
	case "center_hole_diameter":
		return part.CenterHoleDiameter
	case "tire_model":
		return part.TireModel
	case "vin":
		return part.VIN
	default:
		return ""
	}
}
