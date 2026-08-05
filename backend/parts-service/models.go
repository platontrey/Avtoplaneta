package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"
)

// StringArray - кастомный тип для работы с JSONB массивами строк в PostgreSQL
type StringArray []string

// Scan реализует sql.Scanner интерфейс для чтения из БД
func (a *StringArray) Scan(value interface{}) error {
	if value == nil {
		*a = StringArray{}
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("failed to scan StringArray: value is not []byte")
	}

	var arr []string
	if err := json.Unmarshal(bytes, &arr); err != nil {
		return err
	}

	*a = arr
	return nil
}

// Value реализует driver.Valuer интерфейс для записи в БД
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal(a)
}

// PartCore содержит основные поля запчасти
type PartCore struct {
	ID          int64       `json:"id"`
	Name        string      `json:"name"`
	Quantity    int         `json:"quantity"`
	Description string      `json:"description,omitempty"`
	Category    string      `json:"category,omitempty"`
	Price       float64     `json:"price,omitempty"`
	Salesman    string      `json:"salesman,omitempty"`
	Location    string      `json:"location,omitempty"`
	Status      bool        `json:"status,omitempty"`
	Brand       string      `json:"brand,omitempty"`
	Model       string      `json:"model,omitempty"`
	Photos      StringArray `json:"photos,omitempty"`
	Photo       string      `json:"photo,omitempty"` // Для обратной совместимости
	SellerID    int64       `json:"seller_id,omitempty"`
	ToDeleteAt  *time.Time  `json:"to_delete_at,omitempty"`
	VIN         string      `json:"vin,omitempty"`
}

// PartSpecifications содержит технические характеристики запчасти
type PartSpecifications struct {
	BodyBrand         string `json:"body_brand,omitempty"`
	EngineBrand       string `json:"engine_brand,omitempty"`
	CarReleaseDate    string `json:"car_release_date,omitempty"`
	FrontRear         string `json:"front_rear,omitempty"`
	LeftRight         string `json:"left_right,omitempty"`
	TopBottom         string `json:"top_bottom,omitempty"`
	Number            string `json:"number,omitempty"`
	Manufacturer      string `json:"manufacturer,omitempty"`
	ManufacturerCode  string `json:"manufacturer_code,omitempty"`
	OEMCode           string `json:"oem_code,omitempty"`
	Color             string `json:"color,omitempty"`
	Condition         string `json:"condition,omitempty"`
	SupplierCode      string `json:"supplier_code,omitempty"`
	Defect            string `json:"defect,omitempty"`
	Transmission      string `json:"transmission,omitempty"`
	TransmissionModel string `json:"transmission_model,omitempty"`
	Drive             string `json:"drive,omitempty"`
	WearPercentage    string `json:"wear_percentage,omitempty"`
}

// PartTireSpecifications содержит характеристики шин
type PartTireSpecifications struct {
	Season             string `json:"season,omitempty"`
	Diameter           string `json:"diameter,omitempty"`
	Width              string `json:"width,omitempty"`
	Profile            string `json:"profile,omitempty"`
	TireQuantity       string `json:"tire_quantity,omitempty"`
	Drilling           string `json:"drilling,omitempty"`
	Offset             string `json:"offset,omitempty"`
	CenterHoleDiameter string `json:"center_hole_diameter,omitempty"`
	TireModel          string `json:"tire_model,omitempty"`
}

// PartDisplay содержит поля для отображения на фронтенде
type PartDisplay struct {
	ToDeleteAtFormatted string `json:"to_delete_at_formatted,omitempty"`
	TimeUntilDeletion   string `json:"time_until_deletion,omitempty"`
}

// Part объединяет все части модели запчасти
type Part struct {
	PartCore
	PartSpecifications
	PartTireSpecifications
	PartDisplay
}

// IsTire проверяет, является ли запчасть шиной
func (p *Part) IsTire() bool {
	return p.Category == "Шины" || p.Category == "Tires" || p.Category == "Шины и диски"
}

// GetFullSpecifications возвращает все характеристики в виде карты
func (p *Part) GetFullSpecifications() map[string]interface{} {
	specs := make(map[string]interface{})

	specFields := []struct {
		key   string
		value string
	}{
		{"body_brand", p.BodyBrand},
		{"engine_brand", p.EngineBrand},
		{"car_release_date", p.CarReleaseDate},
		{"front_rear", p.FrontRear},
		{"left_right", p.LeftRight},
		{"top_bottom", p.TopBottom},
		{"number", p.Number},
		{"manufacturer", p.Manufacturer},
		{"manufacturer_code", p.ManufacturerCode},
		{"oem_code", p.OEMCode},
		{"color", p.Color},
		{"condition", p.Condition},
		{"supplier_code", p.SupplierCode},
		{"defect", p.Defect},
		{"transmission", p.Transmission},
		{"transmission_model", p.TransmissionModel},
		{"drive", p.Drive},
		{"wear_percentage", p.WearPercentage},
	}
	for _, field := range specFields {
		if field.value != "" {
			specs[field.key] = field.value
		}
	}

	if p.IsTire() {
		tireFields := []struct {
			key   string
			value string
		}{
			{"season", p.Season},
			{"diameter", p.Diameter},
			{"width", p.Width},
			{"profile", p.Profile},
			{"tire_quantity", p.TireQuantity},
			{"drilling", p.Drilling},
			{"offset", p.Offset},
			{"center_hole_diameter", p.CenterHoleDiameter},
			{"tire_model", p.TireModel},
		}
		for _, field := range tireFields {
			if field.value != "" {
				specs[field.key] = field.value
			}
		}
	}

	return specs
}

// SetSpecifications устанавливает характеристики из карты
func (p *Part) SetSpecifications(specs map[string]interface{}) {
	for key, value := range specs {
		strValue := value.(string)
		switch key {
		case "body_brand":
			p.BodyBrand = strValue
		case "engine_brand":
			p.EngineBrand = strValue
		case "car_release_date":
			p.CarReleaseDate = strValue
		case "front_rear":
			p.FrontRear = strValue
		case "left_right":
			p.LeftRight = strValue
		case "top_bottom":
			p.TopBottom = strValue
		case "number":
			p.Number = strValue
		case "manufacturer":
			p.Manufacturer = strValue
		case "manufacturer_code":
			p.ManufacturerCode = strValue
		case "oem_code":
			p.OEMCode = strValue
		case "color":
			p.Color = strValue
		case "condition":
			p.Condition = strValue
		case "supplier_code":
			p.SupplierCode = strValue
		case "defect":
			p.Defect = strValue
		case "transmission":
			p.Transmission = strValue
		case "transmission_model":
			p.TransmissionModel = strValue
		case "drive":
			p.Drive = strValue
		case "wear_percentage":
			p.WearPercentage = strValue
		case "season":
			p.Season = strValue
		case "diameter":
			p.Diameter = strValue
		case "width":
			p.Width = strValue
		case "profile":
			p.Profile = strValue
		case "tire_quantity":
			p.TireQuantity = strValue
		case "drilling":
			p.Drilling = strValue
		case "offset":
			p.Offset = strValue
		case "center_hole_diameter":
			p.CenterHoleDiameter = strValue
		case "tire_model":
			p.TireModel = strValue
		}
	}
}

type StatisticsResponse struct {
	TotalParts    int             `json:"total_parts"`
	TotalQuantity int             `json:"total_quantity"`
	TotalValue    float64         `json:"total_value"`
	TotalEarnings float64         `json:"total_earnings"`
	Categories    []CategoryCount `json:"categories"`
	MonthlySales  []MonthlySales  `json:"monthly_sales"`
}

type CategoryCount struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

type MonthlySales struct {
	Month string  `json:"month"`
	Sales float64 `json:"sales"`
}

// Earnings хранит общий заработок системы
type Earnings struct {
	ID          int64     `json:"id"`
	TotalAmount float64   `json:"total_amount"`
	UpdatedAt   time.Time `json:"updated_at"`
}
