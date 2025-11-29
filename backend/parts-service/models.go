package main

import (
	"time"

	"gorm.io/gorm"
)

type Part struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Name        string     `json:"name" gorm:"not null"`
	Quantity    int        `json:"quantity" gorm:"not null"`
	Description string     `json:"description,omitempty"`
	Category    string     `json:"category,omitempty"`
	Price       float64    `json:"price,omitempty"`
	Salesman    string     `json:"salesman,omitempty"`
	Location    string     `json:"location,omitempty"`
	Status      bool       `json:"status,omitempty"`
	Brand       string     `json:"brand,omitempty"`
	Model       string     `json:"model,omitempty"`
	Photo       string     `json:"photo,omitempty"`
	SellerID    uint       `json:"seller_id,omitempty"` // ID продавца из auth-service
	ToDeleteAt  *time.Time `json:"to_delete_at,omitempty" gorm:"default:null"`
	VIN         string     `json:"vin,omitempty"` // VIN автомобиля
	// Характеристики запчасти
	BodyBrand          string `json:"body_brand,omitempty"`           // Марка кузова
	EngineBrand        string `json:"engine_brand,omitempty"`         // Марка двигателя
	CarReleaseDate     string `json:"car_release_date,omitempty"`     // Дата выпуска автомобиля
	FrontRear          string `json:"front_rear,omitempty"`           // Перед/зад
	LeftRight          string `json:"left_right,omitempty"`           // Право/лево
	TopBottom          string `json:"top_bottom,omitempty"`           // Верх/низ
	Number             string `json:"number,omitempty"`               // Номер
	Manufacturer       string `json:"manufacturer,omitempty"`         // Производитель
	ManufacturerCode   string `json:"manufacturer_code,omitempty"`    // Код производителя
	OEMCode            string `json:"oem_code,omitempty"`             // OEM код
	Color              string `json:"color,omitempty"`                // Цвет
	Condition          string `json:"condition,omitempty"`            // Состояние (Б/у или новый(-ая))
	SupplierCode       string `json:"supplier_code,omitempty"`        // Код поставки
	Defect             string `json:"defect,omitempty"`               // Дефект
	Transmission       string `json:"transmission,omitempty"`         // Трансмиссия
	Drive              string `json:"drive,omitempty"`                // Привод
	WearPercentage     string `json:"wear_percentage,omitempty"`      // Процент износа (%)
	Season             string `json:"season,omitempty"`               // Сезон
	Diameter           string `json:"diameter,omitempty"`             // Диаметр
	Width              string `json:"width,omitempty"`                // Ширина
	Profile            string `json:"profile,omitempty"`              // Профиль
	TireQuantity       string `json:"tire_quantity,omitempty"`        // Количество
	Drilling           string `json:"drilling,omitempty"`             // Сверловка
	Offset             string `json:"offset,omitempty"`               // Вылет
	CenterHoleDiameter string `json:"center_hole_diameter,omitempty"` // Диаметр ЦО
	TireModel          string `json:"tire_model,omitempty"`           // Модель шины
	// Форматированные поля для фронтенда (не сохраняются в БД)
	ToDeleteAtFormatted string `json:"to_delete_at_formatted,omitempty"` // Форматированная дата удаления
	TimeUntilDeletion   string `json:"time_until_deletion,omitempty"`    // Время до удаления (например, "через 3 дня")
}

type StatisticsResponse struct {
	TotalParts    int             `json:"total_parts"`
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

// Модели PartSpecification и SpecificationTemplate больше не нужны - характеристики хранятся в основной таблице Part

var db *gorm.DB
