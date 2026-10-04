package service

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"parts-service/internal/catalog"
	"parts-service/internal/domain"
)

var (
	errPartCatalogUnavailable   = errors.New("каталог шаблонов запчастей недоступен")
	errDefectServiceUnavailable = errors.New("сервис инвентаря для создания запчастей недоступен")
)

const (
	defectReportEventVersion = 1
	defectIdempotencyTTL     = 2 * time.Minute
)

type defectIdempotencyEntry struct {
	createdAt time.Time
	parts     []domain.Part
	err       error
	done      chan struct{}
}

// DefectReportUnavailable отличает временную недоступность инфраструктуры
// (каталог не загружен, сервис не сконфигурирован) от других ошибок.
func DefectReportUnavailable(err error) bool {
	return errors.Is(err, errPartCatalogUnavailable) || errors.Is(err, errDefectServiceUnavailable)
}

// DefectReportWorkflow — единственное место, где входные данные ведомости
// превращаются в запчасти и пакетно создаются в БД и Elasticsearch.
type DefectReportWorkflow struct {
	catalog *catalog.PartCatalog
	service InventoryService
	mu      sync.Mutex
	recent  map[string]*defectIdempotencyEntry
}

func NewDefectReportWorkflow(catalog *catalog.PartCatalog, service InventoryService) *DefectReportWorkflow {
	return &DefectReportWorkflow{
		catalog: catalog,
		service: service,
		recent:  make(map[string]*defectIdempotencyEntry),
	}
}

func defectReportDedupKey(report *catalog.DefectReportRequest) string {
	if key := strings.TrimSpace(report.IdempotencyKey); key != "" {
		sum := sha256.Sum256([]byte(fmt.Sprintf("key:%d:%s", report.SellerID, key)))
		return hex.EncodeToString(sum[:])
	}
	raw := strings.Join([]string{
		strconv.FormatInt(report.SellerID, 10),
		strings.ToLower(strings.TrimSpace(report.Brand)),
		strings.ToLower(strings.TrimSpace(report.Model)),
		strconv.Itoa(report.Year),
		strings.ToLower(strings.TrimSpace(report.CarReleasePeriod)),
		strings.ToUpper(strings.TrimSpace(report.VIN)),
		strings.ToLower(strings.TrimSpace(report.EngineBrand)),
		strings.ToLower(strings.TrimSpace(report.BodyBrand)),
		strings.ToLower(strings.TrimSpace(report.InteriorColor)),
		strings.ToLower(strings.TrimSpace(report.BodyColor)),
		strings.ToLower(strings.TrimSpace(report.Transmission)),
		strings.ToLower(strings.TrimSpace(report.TransmissionModel)),
		strings.ToLower(strings.TrimSpace(report.Drive)),
		strings.ToLower(strings.TrimSpace(report.SupplierCode)),
	}, "|")
	sum := sha256.Sum256([]byte("payload:" + raw))
	return hex.EncodeToString(sum[:])
}

// prepare разворачивает набор запчастей по каталогу.
func (w *DefectReportWorkflow) prepare(ctx context.Context, report *catalog.DefectReportRequest, allowLegacyClientParts bool, isCreate bool) error {
	if w == nil || w.catalog == nil {
		return errPartCatalogUnavailable
	}

	supplierCode := strings.TrimSpace(report.SupplierCode)
	if supplierCode == "" {
		if w.service != nil {
			var err error
			if isCreate {
				supplierCode, err = w.service.GetNextSupplierCode(ctx)
			} else {
				supplierCode, err = w.service.PeekNextSupplierCode(ctx)
			}
			if err != nil || supplierCode == "" {
				supplierCode = "1"
			}
		} else {
			supplierCode = "1"
		}
		report.SupplierCode = supplierCode
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

func (w *DefectReportWorkflow) Preview(ctx context.Context, report catalog.DefectReportRequest) (catalog.DefectReportRequest, error) {
	report.SelectedParts = nil
	if err := w.prepare(ctx, &report, false, false); err != nil {
		return catalog.DefectReportRequest{}, err
	}
	return report, nil
}

// ConvertDefectReportToParts преобразует подготовленную ведомость в срез запчастей
func ConvertDefectReportToParts(report *catalog.DefectReportRequest) []domain.Part {
	var defaultUserID int64 = 1
	var defaultUserName string = "System"

	if report.SellerID > 0 {
		defaultUserID = report.SellerID
	}
	if report.SellerName != "" {
		defaultUserName = report.SellerName
	}

	parts := make([]domain.Part, 0, len(report.SelectedParts))
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

		part := domain.Part{
			PartCore: domain.PartCore{
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
			PartSpecifications: domain.PartSpecifications{
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
			PartTireSpecifications: domain.PartTireSpecifications{
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

// Create синхронно, транзакционно и идемпотентно создает все запчасти из ведомости за один батч
func (w *DefectReportWorkflow) Create(ctx context.Context, report *catalog.DefectReportRequest, allowLegacyClientParts bool) ([]domain.Part, error) {
	if w == nil || w.service == nil {
		return nil, errDefectServiceUnavailable
	}

	dedupKey := defectReportDedupKey(report)
	now := time.Now()

	w.mu.Lock()
	if w.recent == nil {
		w.recent = make(map[string]*defectIdempotencyEntry)
	}
	for k, v := range w.recent {
		select {
		case <-v.done:
			if now.Sub(v.createdAt) > defectIdempotencyTTL {
				delete(w.recent, k)
			}
		default:
		}
	}
	if existing, ok := w.recent[dedupKey]; ok {
		w.mu.Unlock()
		select {
		case <-existing.done:
			return existing.parts, existing.err
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	entry := &defectIdempotencyEntry{
		createdAt: now,
		done:      make(chan struct{}),
	}
	w.recent[dedupKey] = entry
	w.mu.Unlock()

	defer func() {
		if entry.err != nil {
			w.mu.Lock()
			delete(w.recent, dedupKey)
			w.mu.Unlock()
		}
		close(entry.done)
	}()

	if err := w.prepare(ctx, report, allowLegacyClientParts, true); err != nil {
		entry.err = err
		return nil, err
	}

	parts := ConvertDefectReportToParts(report)
	created, err := w.service.AddPartsBatch(ctx, parts)
	entry.parts = created
	entry.err = err
	return created, err
}

// Enqueue оставлен для обратной совместимости, выполняет прямое батч-создание
func (w *DefectReportWorkflow) Enqueue(ctx context.Context, report *catalog.DefectReportRequest, allowLegacyClientParts bool) error {
	_, err := w.Create(ctx, report, allowLegacyClientParts)
	return err
}
