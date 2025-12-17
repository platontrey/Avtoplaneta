package main

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"gorm.io/gorm"
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

	*a = StringArray(arr)
	return nil
}

// Value реализует driver.Valuer интерфейс для записи в БД
func (a StringArray) Value() (driver.Value, error) {
	if len(a) == 0 {
		return json.Marshal([]string{})
	}
	return json.Marshal([]string(a))
}

// PartCore содержит основные поля запчасти
type PartCore struct {
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
	Photos      StringArray `json:"photos,omitempty" gorm:"type:jsonb"`
	Photo       string     `json:"photo,omitempty" gorm:"-"` // Для обратной совместимости
	SellerID    uint       `json:"seller_id,omitempty"` // ID продавца из auth-service
	ToDeleteAt  *time.Time `json:"to_delete_at,omitempty" gorm:"default:null"`
	VIN         string     `json:"vin,omitempty"` // VIN автомобиля
}

// PartSpecifications содержит технические характеристики запчасти
type PartSpecifications struct {
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
}

// PartTireSpecifications содержит характеристики шин
type PartTireSpecifications struct {
	Season             string `json:"season,omitempty"`               // Сезон
	Diameter           string `json:"diameter,omitempty"`             // Диаметр
	Width              string `json:"width,omitempty"`                // Ширина
	Profile            string `json:"profile,omitempty"`              // Профиль
	TireQuantity       string `json:"tire_quantity,omitempty"`        // Количество
	Drilling           string `json:"drilling,omitempty"`             // Сверловка
	Offset             string `json:"offset,omitempty"`               // Вылет
	CenterHoleDiameter string `json:"center_hole_diameter,omitempty"` // Диаметр ЦО
	TireModel          string `json:"tire_model,omitempty"`           // Модель шины
}

// PartDisplay содержит поля для отображения на фронтенде
type PartDisplay struct {
	ToDeleteAtFormatted string `json:"to_delete_at_formatted,omitempty"` // Форматированная дата удаления
	TimeUntilDeletion   string `json:"time_until_deletion,omitempty"`    // Время до удаления (например, "через 3 дня")
}

// Part объединяет все части модели запчасти
type Part struct {
	PartCore                   `gorm:"embedded"`
	PartSpecifications         `gorm:"embedded"`
	PartTireSpecifications     `gorm:"embedded"`
	PartDisplay                `gorm:"-"` // Не сохраняется в БД
}

// NewPart создает новую запчасть с базовыми полями
func NewPart(name string, quantity int) *Part {
	return &Part{
		PartCore: PartCore{
			Name:     name,
			Quantity: quantity,
		},
	}
}

// IsTire проверяет, является ли запчасть шиной
func (p *Part) IsTire() bool {
	return p.Category == "Шины" || p.Category == "Tires"
}

// AfterFind синхронизирует поле Photo с первым элементом массива Photos для обратной совместимости
func (p *Part) AfterFind(tx *gorm.DB) error {
	if len(p.Photos) > 0 {
		p.Photo = string(p.Photos[0])
	}
	return nil
}

// BeforeSave синхронизирует массив Photos с полем Photo для обратной совместимости
func (p *Part) BeforeSave(tx *gorm.DB) error {
	if p.Photo != "" && len(p.Photos) == 0 {
		p.Photos = StringArray{p.Photo}
	} else if len(p.Photos) > 0 {
		p.Photo = string(p.Photos[0])
	}
	return nil
}

// GetFullSpecifications возвращает все характеристики в виде карты
func (p *Part) GetFullSpecifications() map[string]interface{} {
	specs := make(map[string]interface{})

	// Основные характеристики
	if p.BodyBrand != "" {
		specs["body_brand"] = p.BodyBrand
	}
	if p.EngineBrand != "" {
		specs["engine_brand"] = p.EngineBrand
	}
	if p.CarReleaseDate != "" {
		specs["car_release_date"] = p.CarReleaseDate
	}
	if p.FrontRear != "" {
		specs["front_rear"] = p.FrontRear
	}
	if p.LeftRight != "" {
		specs["left_right"] = p.LeftRight
	}
	if p.TopBottom != "" {
		specs["top_bottom"] = p.TopBottom
	}
	if p.Number != "" {
		specs["number"] = p.Number
	}
	if p.Manufacturer != "" {
		specs["manufacturer"] = p.Manufacturer
	}
	if p.ManufacturerCode != "" {
		specs["manufacturer_code"] = p.ManufacturerCode
	}
	if p.OEMCode != "" {
		specs["oem_code"] = p.OEMCode
	}
	if p.Color != "" {
		specs["color"] = p.Color
	}
	if p.Condition != "" {
		specs["condition"] = p.Condition
	}
	if p.SupplierCode != "" {
		specs["supplier_code"] = p.SupplierCode
	}
	if p.Defect != "" {
		specs["defect"] = p.Defect
	}
	if p.Transmission != "" {
		specs["transmission"] = p.Transmission
	}
	if p.Drive != "" {
		specs["drive"] = p.Drive
	}
	if p.WearPercentage != "" {
		specs["wear_percentage"] = p.WearPercentage
	}

	// Характеристики шин (если применимо)
	if p.IsTire() {
		if p.Season != "" {
			specs["season"] = p.Season
		}
		if p.Diameter != "" {
			specs["diameter"] = p.Diameter
		}
		if p.Width != "" {
			specs["width"] = p.Width
		}
		if p.Profile != "" {
			specs["profile"] = p.Profile
		}
		if p.TireQuantity != "" {
			specs["tire_quantity"] = p.TireQuantity
		}
		if p.Drilling != "" {
			specs["drilling"] = p.Drilling
		}
		if p.Offset != "" {
			specs["offset"] = p.Offset
		}
		if p.CenterHoleDiameter != "" {
			specs["center_hole_diameter"] = p.CenterHoleDiameter
		}
		if p.TireModel != "" {
			specs["tire_model"] = p.TireModel
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
