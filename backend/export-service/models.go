package main

import (
	"time"

	partsv1 "avtoplaneta/gen/parts/v1"
)

// Part — не копия запчасти из parts-service, а ровно тот набор полей, который
// попадает в прайс-лист. Сервис выгрузки не хранит запчасти и ничего о них не
// решает: он читает поток, перекладывает его в offer и забывает.
//
// Поэтому здесь нет ни id склада, ни шин, ни сроков удаления сверх нужного —
// добавлять поле сюда имеет смысл только тогда, когда его требует площадка.
type Part struct {
	Name        string
	Quantity    int
	Description string
	Price       float64
	Salesman    string
	Location    string
	Status      bool
	Brand       string
	Model       string
	Photos      []string
	SellerID    int64
	ToDeleteAt  *time.Time

	BodyBrand      string
	EngineBrand    string
	CarReleaseDate string
	FrontRear      string
	LeftRight      string
	TopBottom      string
	Manufacturer   string
	OEMCode        string
	Color          string
	Condition      string
	SupplierCode   string
}

func partFromProto(p *partsv1.Part) Part {
	part := Part{
		Name:           p.GetName(),
		Quantity:       int(p.GetQuantity()),
		Description:    p.GetDescription(),
		Price:          p.GetPrice(),
		Salesman:       p.GetSalesman(),
		Location:       p.GetLocation(),
		Status:         p.GetStatus(),
		Brand:          p.GetBrand(),
		Model:          p.GetModel(),
		Photos:         p.GetPhotos(),
		SellerID:       int64(p.GetSellerId()),
		BodyBrand:      p.GetBodyBrand(),
		EngineBrand:    p.GetEngineBrand(),
		CarReleaseDate: p.GetCarReleaseDate(),
		FrontRear:      p.GetFrontRear(),
		LeftRight:      p.GetLeftRight(),
		TopBottom:      p.GetTopBottom(),
		Manufacturer:   p.GetManufacturer(),
		OEMCode:        p.GetOemCode(),
		Color:          p.GetColor(),
		Condition:      p.GetCondition(),
		SupplierCode:   p.GetSupplierCode(),
	}

	if ts := p.GetToDeleteAt(); ts != nil {
		deleteAt := ts.AsTime()
		part.ToDeleteAt = &deleteAt
	}

	return part
}
