package main

import (
	"context"
	"errors"
	"strconv"
	"strings"
)

var (
	errPartCatalogUnavailable   = errors.New("каталог шаблонов запчастей недоступен")
	errDefectServiceUnavailable = errors.New("сервис инвентаря для создания запчастей недоступен")
)

const (
	defectReportEventVersion = 1
)

// defectReportUnavailable отличает временную недоступность инфраструктуры
// (каталог не загружен, сервис не сконфигурирован) от других ошибок.
func defectReportUnavailable(err error) bool {
	return errors.Is(err, errPartCatalogUnavailable) || errors.Is(err, errDefectServiceUnavailable)
}

// DefectReportWorkflow — единственное место, где входные данные ведомости
// превращаются в запчасти и пакетно создаются в БД и Elasticsearch.
type DefectReportWorkflow struct {
	catalog *PartCatalog
	service InventoryService
}

func NewDefectReportWorkflow(catalog *PartCatalog, service InventoryService) *DefectReportWorkflow {
	return &DefectReportWorkflow{catalog: catalog, service: service}
}

// prepare разворачивает набор запчастей по каталогу.
func (w *DefectReportWorkflow) prepare(report *DefectReportRequest, allowLegacyClientParts bool) error {
	if w == nil || w.catalog == nil {
		return errPartCatalogUnavailable
	}

	if !allowLegacyClientParts || len(report.SelectedParts) == 0 {
		report.SelectedParts = w.catalog.ExpandDefectReport(*report)
	} else {
		w.catalog.ApplyBindingsToParts(report.SelectedParts, *report)
	}
	report.CatalogVersion = w.catalog.Version
	report.EventVersion = defectReportEventVersion
	return nil
}

func (w *DefectReportWorkflow) Preview(report DefectReportRequest) (DefectReportRequest, error) {
	report.SelectedParts = nil
	if err := w.prepare(&report, false); err != nil {
		return DefectReportRequest{}, err
	}
	return report, nil
}

// ConvertDefectReportToParts преобразует подготовленную ведомость в срез запчастей
func ConvertDefectReportToParts(report *DefectReportRequest) []Part {
	var defaultUserID int64 = 1
	var defaultUserName string = "System"

	if report.SellerID > 0 {
		defaultUserID = report.SellerID
	}
	if report.SellerName != "" {
		defaultUserName = report.SellerName
	}

	parts := make([]Part, 0, len(report.SelectedParts))
	for _, selectedPart := range report.SelectedParts {
		vin := strings.TrimSpace(selectedPart.VIN)
		carReleaseDate := strings.TrimSpace(selectedPart.CarReleaseDate)
		carReleasePeriod := strings.TrimSpace(selectedPart.CarReleasePeriod)
		bodyBrand := strings.TrimSpace(selectedPart.BodyBrand)
		engineBrand := strings.TrimSpace(selectedPart.EngineBrand)
		transmission := strings.TrimSpace(selectedPart.Transmission)
		transmissionModel := strings.TrimSpace(selectedPart.TransmissionModel)
		drive := strings.TrimSpace(selectedPart.Drive)
		color := strings.TrimSpace(selectedPart.Color)

		if vin == "" {
			vin = strings.TrimSpace(report.VIN)
		}
		if carReleaseDate == "" && report.Year > 0 {
			carReleaseDate = strconv.Itoa(report.Year)
		}
		if carReleasePeriod == "" {
			carReleasePeriod = strings.TrimSpace(report.CarReleasePeriod)
		}
		if bodyBrand == "" {
			bodyBrand = strings.TrimSpace(report.BodyBrand)
		}
		if engineBrand == "" {
			engineBrand = strings.TrimSpace(report.EngineBrand)
		}
		if transmission == "" {
			transmission = strings.TrimSpace(report.Transmission)
		}
		if transmissionModel == "" {
			transmissionModel = strings.TrimSpace(report.TransmissionModel)
		}
		if drive == "" {
			drive = strings.TrimSpace(report.Drive)
		}
		if color == "" {
			switch selectedPart.Category {
			case "Электрооснащение", "Система кондиционирования", "Сопутствующие товары":
				color = strings.TrimSpace(report.InteriorColor)
				if color == "" {
					color = "Черный"
				}
			default:
				color = strings.TrimSpace(report.BodyColor)
				if color == "" {
					color = "Белый"
				}
			}
		}

		part := Part{
			PartCore: PartCore{
				Name:        strings.TrimSpace(selectedPart.Name),
				Quantity:    selectedPart.Quantity,
				Description: strings.TrimSpace(selectedPart.Description),
				Category:    strings.TrimSpace(selectedPart.Category),
				Price:       selectedPart.Price,
				Salesman:    strings.TrimSpace(defaultUserName),
				Location:    strings.TrimSpace(selectedPart.Location),
				Address:     selectedPart.Address,
				Status:      true,
				Brand:       strings.TrimSpace(report.Brand),
				Model:       strings.TrimSpace(report.Model),
				Photo:       "",
				SellerID:    defaultUserID,
				VIN:         vin,
			},
			PartSpecifications: PartSpecifications{
				BodyBrand:         bodyBrand,
				EngineBrand:       engineBrand,
				CarReleaseDate:    carReleaseDate,
				CarReleasePeriod:  carReleasePeriod,
				FrontRear:         selectedPart.FrontRear,
				LeftRight:         selectedPart.LeftRight,
				TopBottom:         selectedPart.TopBottom,
				Number:            selectedPart.Number,
				Manufacturer:      selectedPart.Manufacturer,
				ManufacturerCode:  selectedPart.ManufacturerCode,
				OEMCode:           selectedPart.OEMCode,
				Color:             color,
				Condition:         selectedPart.Condition,
				SupplierCode:      selectedPart.SupplierCode,
				Defect:            selectedPart.Defect,
				Transmission:      transmission,
				TransmissionModel: transmissionModel,
				Drive:             drive,
				WearPercentage:    selectedPart.WearPercentage,
			},
			PartTireSpecifications: PartTireSpecifications{
				Season:             selectedPart.Season,
				Diameter:           selectedPart.Diameter,
				Width:              selectedPart.Width,
				Profile:            selectedPart.Profile,
				TireQuantity:       selectedPart.TireQuantity,
				Drilling:           selectedPart.Drilling,
				Offset:             selectedPart.Offset,
				CenterHoleDiameter: selectedPart.CenterHoleDiameter,
				TireModel:          selectedPart.TireModel,
			},
		}

		parts = append(parts, part)
	}

	return parts
}

// Create синхронно и транзакционно создает все запчасти из ведомости за один батч
func (w *DefectReportWorkflow) Create(ctx context.Context, report *DefectReportRequest, allowLegacyClientParts bool) ([]Part, error) {
	if w == nil || w.service == nil {
		return nil, errDefectServiceUnavailable
	}
	if err := w.prepare(report, allowLegacyClientParts); err != nil {
		return nil, err
	}

	parts := ConvertDefectReportToParts(report)
	return w.service.AddPartsBatch(ctx, parts)
}

// Enqueue оставлен для обратной совместимости, выполняет прямое батч-создание
func (w *DefectReportWorkflow) Enqueue(ctx context.Context, report *DefectReportRequest, allowLegacyClientParts bool) error {
	_, err := w.Create(ctx, report, allowLegacyClientParts)
	return err
}
