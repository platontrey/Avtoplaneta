package repository

import (
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5"

	"parts-service/db/sqlc"
	"parts-service/internal/domain"
)

// ─── Helpers ────────────────────────────────────────────────────────────────

var partColumns = `id, name, quantity, description, category, price, salesman, location, address, status,
	brand, model, photos, seller_id, to_delete_at, vin,
	body_brand, engine_brand, car_release_date, car_release_period, front_rear, left_right, top_bottom,
	number, manufacturer, manufacturer_code, oem_code, color, condition,
	supplier_code, defect, transmission, transmission_model, drive, wear_percentage,
	season, diameter, width, profile, tire_quantity, drilling, "offset",
	center_hole_diameter, tire_model, created_at, updated_at, deleted_at`

func scanPart(row pgx.Row) (*domain.Part, error) {
	var p domain.Part
	var photos []byte
	var toDeleteAt *time.Time
	var createdAt, updatedAt time.Time
	var deletedAt *time.Time

	err := row.Scan(
		&p.ID, &p.Name, &p.Quantity, &p.Description, &p.Category, &p.Price,
		&p.Salesman, &p.Location, &p.Address, &p.Status, &p.Brand, &p.Model, &photos,
		&p.SellerID, &toDeleteAt, &p.VIN,
		&p.BodyBrand, &p.EngineBrand, &p.CarReleaseDate, &p.CarReleasePeriod, &p.FrontRear, &p.LeftRight, &p.TopBottom,
		&p.Number, &p.Manufacturer, &p.ManufacturerCode, &p.OEMCode, &p.Color, &p.Condition,
		&p.SupplierCode, &p.Defect, &p.Transmission, &p.TransmissionModel, &p.Drive, &p.WearPercentage,
		&p.Season, &p.Diameter, &p.Width, &p.Profile, &p.TireQuantity, &p.Drilling, &p.Offset,
		&p.CenterHoleDiameter, &p.TireModel,
		&createdAt, &updatedAt, &deletedAt,
	)
	if err != nil {
		return nil, err
	}

	p.ToDeleteAt = toDeleteAt
	if photos != nil {
		_ = json.Unmarshal(photos, &p.Photos)
	}
	if len(p.Photos) > 0 {
		p.Photo = p.Photos[0]
	}
	return &p, nil
}

func scanParts(rows pgx.Rows) ([]domain.Part, error) {
	var parts []domain.Part
	for rows.Next() {
		var p domain.Part
		var photos []byte
		var toDeleteAt *time.Time
		var createdAt, updatedAt time.Time
		var deletedAt *time.Time

		err := rows.Scan(
			&p.ID, &p.Name, &p.Quantity, &p.Description, &p.Category, &p.Price,
			&p.Salesman, &p.Location, &p.Address, &p.Status, &p.Brand, &p.Model, &photos,
			&p.SellerID, &toDeleteAt, &p.VIN,
			&p.BodyBrand, &p.EngineBrand, &p.CarReleaseDate, &p.CarReleasePeriod, &p.FrontRear, &p.LeftRight, &p.TopBottom,
			&p.Number, &p.Manufacturer, &p.ManufacturerCode, &p.OEMCode, &p.Color, &p.Condition,
			&p.SupplierCode, &p.Defect, &p.Transmission, &p.TransmissionModel, &p.Drive, &p.WearPercentage,
			&p.Season, &p.Diameter, &p.Width, &p.Profile, &p.TireQuantity, &p.Drilling, &p.Offset,
			&p.CenterHoleDiameter, &p.TireModel,
			&createdAt, &updatedAt, &deletedAt,
		)
		if err != nil {
			return nil, err
		}

		p.ToDeleteAt = toDeleteAt
		if photos != nil {
			_ = json.Unmarshal(photos, &p.Photos)
		}
		if len(p.Photos) > 0 {
			p.Photo = p.Photos[0]
		}
		parts = append(parts, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if parts == nil {
		parts = []domain.Part{}
	}
	return parts, nil
}

func sqlcPartToDomain(p sqlc.Part) *domain.Part {
	res := &domain.Part{
		PartCore: domain.PartCore{
			ID:          p.ID,
			Name:        p.Name,
			Quantity:    int(p.Quantity),
			Description: p.Description,
			Category:    p.Category,
			Price:       p.Price,
			Salesman:    p.Salesman,
			Location:    p.Location,
			Address:     p.Address,
			Status:      p.Status,
			Brand:       p.Brand,
			Model:       p.Model,
			SellerID:    p.SellerID,
			VIN:         p.Vin,
		},
		PartSpecifications: domain.PartSpecifications{
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
			OEMCode:           p.OemCode,
			Color:             p.Color,
			Condition:         p.Condition,
			SupplierCode:      p.SupplierCode,
			Defect:            p.Defect,
			Transmission:      p.Transmission,
			TransmissionModel: p.TransmissionModel,
			Drive:             p.Drive,
			WearPercentage:    p.WearPercentage,
		},
		PartTireSpecifications: domain.PartTireSpecifications{
			Season:             p.Season,
			Diameter:           p.Diameter,
			Width:              p.Width,
			Profile:            p.Profile,
			TireQuantity:       p.TireQuantity,
			Drilling:           p.Drilling,
			Offset:             p.Offset,
			CenterHoleDiameter: p.CenterHoleDiameter,
			TireModel:          p.TireModel,
		},
	}
	if p.ToDeleteAt.Valid {
		t := p.ToDeleteAt.Time
		res.ToDeleteAt = &t
	}
	if p.Photos != nil {
		_ = json.Unmarshal(p.Photos, &res.Photos)
	}
	if len(res.Photos) > 0 {
		res.Photo = res.Photos[0]
	}
	return res
}
