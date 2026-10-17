package domain

import (
	partsv1 "avtoplaneta/gen/parts/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// ToProto конвертирует модель Part в Protobuf сообщение partsv1.Part.
func (p *Part) ToProto() *partsv1.Part {
	if p == nil {
		return nil
	}
	proto := &partsv1.Part{
		Id:          p.ID,
		Name:        p.Name,
		Quantity:    int32(p.Quantity),
		Description: p.Description,
		Category:    p.Category,
		Price:       p.Price,
		Salesman:    p.Salesman,
		Location:    p.Location,
		Address:     p.Address,
		Status:      p.Status,
		Brand:       p.Brand,
		Model:       p.Model,
		Photos:      []string(p.Photos),
		SellerId:    p.SellerID,
		Vin:         p.VIN,
		// Specifications
		BodyBrand:         p.BodyBrand,
		EngineBrand:       p.EngineBrand,
		CarReleaseDate:    p.CarReleaseDate,
		CarReleasePeriod:  p.CarReleasePeriod,
		FrontRear:         p.FrontRear,
		LeftRight:         p.LeftRight,
		TopBottom:         p.TopBottom,
		Number:            p.Number,
		Manufacturer:      p.Manufacturer,
		ManufacturerCode:  p.ManufacturerCode,
		OemCode:           p.OEMCode,
		Color:             p.Color,
		Condition:         p.Condition,
		SupplierCode:      p.SupplierCode,
		Defect:            p.Defect,
		Transmission:      p.Transmission,
		TransmissionModel: p.TransmissionModel,
		Drive:             p.Drive,
		WearPercentage:    p.WearPercentage,
		// Tire
		Season:             p.Season,
		Diameter:           p.Diameter,
		Width:              p.Width,
		Profile:            p.Profile,
		TireQuantity:       p.TireQuantity,
		Drilling:           p.Drilling,
		Offset:             p.Offset,
		CenterHoleDiameter: p.CenterHoleDiameter,
		TireModel:          p.TireModel,
		// Display
		ToDeleteAtFormatted: p.ToDeleteAtFormatted,
		TimeUntilDeletion:   p.TimeUntilDeletion,
	}

	if p.ToDeleteAt != nil {
		proto.ToDeleteAt = timestamppb.New(*p.ToDeleteAt)
	}

	return proto
}
