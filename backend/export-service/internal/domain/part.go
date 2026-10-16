package domain

import "time"

// Part представляет доменную модель автозапчасти для экспорта в классифайды
type Part struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Quantity    int        `json:"quantity"`
	Description string     `json:"description,omitempty"`
	Price       float64    `json:"price,omitempty"`
	Salesman    string     `json:"salesman,omitempty"`
	Location    string     `json:"location,omitempty"`
	Status      bool       `json:"status,omitempty"`
	Brand       string     `json:"brand,omitempty"`
	Model       string     `json:"model,omitempty"`
	Photos      []string   `json:"photos,omitempty"`
	SellerID    int64      `json:"seller_id,omitempty"`
	ToDeleteAt  *time.Time `json:"to_delete_at,omitempty"`

	BodyBrand      string `json:"body_brand,omitempty"`
	EngineBrand    string `json:"engine_brand,omitempty"`
	CarReleaseDate string `json:"car_release_date,omitempty"`
	FrontRear      string `json:"front_rear,omitempty"`
	LeftRight      string `json:"left_right,omitempty"`
	TopBottom      string `json:"top_bottom,omitempty"`
	Manufacturer   string `json:"manufacturer,omitempty"`
	OEMCode        string `json:"oem_code,omitempty"`
	Color          string `json:"color,omitempty"`
	Condition      string `json:"condition,omitempty"`
	SupplierCode   string `json:"supplier_code,omitempty"`
}
