package main

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"time"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	partsv1 "avtoplaneta/gen/parts/v1"
)

// partsGRPCServer реализует gRPC-сервер для PartsService
type partsGRPCServer struct {
	partsv1.UnimplementedPartsServiceServer
	service       InventoryService
	defectReports *DefectReportWorkflow
}

// NewPartsGRPCServer создаёт новый gRPC-сервер.
// Workflow передаётся явно — адаптер не знает ни про каталог, ни про Redis.
func NewPartsGRPCServer(service InventoryService, defectReports *DefectReportWorkflow) *partsGRPCServer {
	return &partsGRPCServer{
		service:       service,
		defectReports: defectReports,
	}
}

// GetInventory возвращает список запчастей с фильтрами
func (s *partsGRPCServer) GetInventory(ctx context.Context, req *partsv1.GetInventoryRequest) (*partsv1.InventoryResponse, error) {
	params := InventoryQueryParams{
		Search:             req.Search,
		Category:           req.Category,
		Brand:              req.Brand,
		Model:              req.Model,
		Location:           req.Location,
		Address:            req.Address,
		Salesman:           req.Salesman,
		Status:             req.Status,
		HasPhoto:           req.HasPhoto,
		Number:             req.Number,
		OEMCode:            req.OemCode,
		VIN:                req.Vin,
		BodyBrand:          req.BodyBrand,
		EngineBrand:        req.EngineBrand,
		CarReleaseDate:     req.CarReleaseDate,
		CarReleasePeriod:   req.CarReleasePeriod,
		Transmission:       req.Transmission,
		Drive:              req.Drive,
		Condition:          req.Condition,
		Manufacturer:       req.Manufacturer,
		Defect:             req.Defect,
		Color:              req.Color,
		MinPrice:           req.MinPrice,
		MaxPrice:           req.MaxPrice,
		MinQuantity:        req.MinQuantity,
		MaxQuantity:        req.MaxQuantity,
		FrontRear:          req.FrontRear,
		LeftRight:          req.LeftRight,
		TopBottom:          req.TopBottom,
		ManufacturerCode:   req.ManufacturerCode,
		SupplierCode:       req.SupplierCode,
		TransmissionModel:  req.TransmissionModel,
		WearPercentage:     req.WearPercentage,
		Season:             req.Season,
		Diameter:           req.Diameter,
		Width:              req.Width,
		Profile:            req.Profile,
		TireQuantity:       req.TireQuantity,
		Drilling:           req.Drilling,
		Offset:             req.Offset,
		CenterHoleDiameter: req.CenterHoleDiameter,
		TireModel:          req.TireModel,
		Page:               int(req.Page),
		Limit:              int(req.Limit),
	}

	if params.Page == 0 {
		params.Page = 1
	}
	if params.Limit == 0 {
		params.Limit = 50
	}

	parts, err := s.service.GetInventory(ctx, params)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить инвентарь: %v", err)
	}

	protoParts := make([]*partsv1.Part, len(parts))
	for i, p := range parts {
		protoParts[i] = partToProto(&p)
	}

	return &partsv1.InventoryResponse{
		Parts: protoParts,
		Total: int32(len(parts)),
		Page:  req.Page,
		Limit: req.Limit,
	}, nil
}

// GetPart возвращает запчасть по ID
func (s *partsGRPCServer) GetPart(ctx context.Context, req *partsv1.GetPartRequest) (*partsv1.Part, error) {
	part, err := s.service.GetPartByID(ctx, int64(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "запчасть не найдена: %v", err)
	}
	return partToProto(part), nil
}

// DecreasePartQuantity уменьшает количество запчастей
func (s *partsGRPCServer) DecreasePartQuantity(ctx context.Context, req *partsv1.ChangePartQuantityRequest) (*partsv1.ChangePartQuantityResponse, error) {
	err := s.service.DecreasePartQuantity(ctx, int64(req.Id), int(req.Amount))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка уменьшения количества: %v", err)
	}

	part, err := s.service.GetPartByID(ctx, int64(req.Id))
	if err != nil {
		return &partsv1.ChangePartQuantityResponse{Success: true}, nil
	}
	return &partsv1.ChangePartQuantityResponse{Success: true, NewQuantity: int32(part.Quantity)}, nil
}

// IncreasePartQuantity увеличивает количество запчастей
func (s *partsGRPCServer) IncreasePartQuantity(ctx context.Context, req *partsv1.ChangePartQuantityRequest) (*partsv1.ChangePartQuantityResponse, error) {
	err := s.service.IncreasePartQuantity(ctx, int64(req.Id), int(req.Amount))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "ошибка увеличения количества: %v", err)
	}

	part, err := s.service.GetPartByID(ctx, int64(req.Id))
	if err != nil {
		return &partsv1.ChangePartQuantityResponse{Success: true}, nil
	}
	return &partsv1.ChangePartQuantityResponse{Success: true, NewQuantity: int32(part.Quantity)}, nil
}

// AddPart добавляет новую запчасть
func (s *partsGRPCServer) AddPart(ctx context.Context, req *partsv1.AddPartRequest) (*partsv1.Part, error) {
	part := &Part{
		PartCore: PartCore{
			Name:        req.Name,
			Quantity:    int(req.Quantity),
			Description: req.Description,
			Category:    req.Category,
			Price:       req.Price,
			Salesman:    req.Salesman,
			Location:    req.Location,
			Address:     req.Address,
			Status:      req.Status,
			Brand:       req.Brand,
			Model:       req.Model,
			SellerID:    int64(req.SellerId),
			VIN:         req.Vin,
		},
		PartSpecifications: PartSpecifications{
			BodyBrand:         req.BodyBrand,
			EngineBrand:       req.EngineBrand,
			CarReleaseDate:    req.CarReleaseDate,
			CarReleasePeriod:  req.CarReleasePeriod,
			FrontRear:         req.FrontRear,
			LeftRight:         req.LeftRight,
			TopBottom:         req.TopBottom,
			Number:            req.Number,
			Manufacturer:      req.Manufacturer,
			ManufacturerCode:  req.ManufacturerCode,
			OEMCode:           req.OemCode,
			Color:             req.Color,
			Condition:         req.Condition,
			SupplierCode:      req.SupplierCode,
			Defect:            req.Defect,
			Transmission:      req.Transmission,
			TransmissionModel: req.TransmissionModel,
			Drive:             req.Drive,
			WearPercentage:    req.WearPercentage,
		},
		PartTireSpecifications: PartTireSpecifications{
			Season:             req.Season,
			Diameter:           req.Diameter,
			Width:              req.Width,
			Profile:            req.Profile,
			TireQuantity:       req.TireQuantity,
			Drilling:           req.Drilling,
			Offset:             req.Offset,
			CenterHoleDiameter: req.CenterHoleDiameter,
			TireModel:          req.TireModel,
		},
	}

	created, err := s.service.AddPart(ctx, part)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "не удалось добавить запчасть: %v", err)
	}

	return partToProto(created), nil
}

// UpdatePart обновляет запчасть
func (s *partsGRPCServer) UpdatePart(ctx context.Context, req *partsv1.UpdatePartRequest) (*partsv1.Part, error) {
	updates := make(map[string]interface{})

	if req.Name != nil {
		updates["name"] = *req.Name
	}
	if req.Quantity != nil {
		updates["quantity"] = int(*req.Quantity)
	}
	if req.Description != nil {
		updates["description"] = *req.Description
	}
	if req.Category != nil {
		updates["category"] = *req.Category
	}
	if req.Price != nil {
		updates["price"] = *req.Price
	}
	if req.Salesman != nil {
		updates["salesman"] = *req.Salesman
	}
	if req.Location != nil {
		updates["location"] = *req.Location
	}
	if req.Address != nil {
		updates["address"] = *req.Address
	}
	if req.Status != nil {
		updates["status"] = *req.Status
	}
	if req.Brand != nil {
		updates["brand"] = *req.Brand
	}
	if req.Model != nil {
		updates["model"] = *req.Model
	}
	if req.SellerId != nil {
		updates["seller_id"] = uint(*req.SellerId)
	}
	if req.Vin != nil {
		updates["vin"] = *req.Vin
	}
	// Specifications
	if req.BodyBrand != nil {
		updates["body_brand"] = *req.BodyBrand
	}
	if req.EngineBrand != nil {
		updates["engine_brand"] = *req.EngineBrand
	}
	if req.CarReleaseDate != nil {
		updates["car_release_date"] = *req.CarReleaseDate
	}
	if req.CarReleasePeriod != nil {
		updates["car_release_period"] = *req.CarReleasePeriod
	}
	if req.FrontRear != nil {
		updates["front_rear"] = *req.FrontRear
	}
	if req.LeftRight != nil {
		updates["left_right"] = *req.LeftRight
	}
	if req.TopBottom != nil {
		updates["top_bottom"] = *req.TopBottom
	}
	if req.Number != nil {
		updates["number"] = *req.Number
	}
	if req.Manufacturer != nil {
		updates["manufacturer"] = *req.Manufacturer
	}
	if req.ManufacturerCode != nil {
		updates["manufacturer_code"] = *req.ManufacturerCode
	}
	if req.OemCode != nil {
		updates["oem_code"] = *req.OemCode
	}
	if req.Color != nil {
		updates["color"] = *req.Color
	}
	if req.Condition != nil {
		updates["condition"] = *req.Condition
	}
	if req.SupplierCode != nil {
		updates["supplier_code"] = *req.SupplierCode
	}
	if req.Defect != nil {
		updates["defect"] = *req.Defect
	}
	if req.Transmission != nil {
		updates["transmission"] = *req.Transmission
	}
	if req.TransmissionModel != nil {
		updates["transmission_model"] = *req.TransmissionModel
	}
	if req.Drive != nil {
		updates["drive"] = *req.Drive
	}
	if req.WearPercentage != nil {
		updates["wear_percentage"] = *req.WearPercentage
	}
	// Tire specifications
	if req.Season != nil {
		updates["season"] = *req.Season
	}
	if req.Diameter != nil {
		updates["diameter"] = *req.Diameter
	}
	if req.Width != nil {
		updates["width"] = *req.Width
	}
	if req.Profile != nil {
		updates["profile"] = *req.Profile
	}
	if req.TireQuantity != nil {
		updates["tire_quantity"] = *req.TireQuantity
	}
	if req.Drilling != nil {
		updates["drilling"] = *req.Drilling
	}
	if req.Offset != nil {
		updates["offset"] = *req.Offset
	}
	if req.CenterHoleDiameter != nil {
		updates["center_hole_diameter"] = *req.CenterHoleDiameter
	}
	if req.TireModel != nil {
		updates["tire_model"] = *req.TireModel
	}

	if len(updates) == 0 {
		return nil, status.Errorf(codes.InvalidArgument, "нет полей для обновления")
	}

	if err := s.service.UpdatePart(ctx, int64(req.Id), updates); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось обновить запчасть: %v", err)
	}

	// Возвращаем обновлённую запчасть
	updated, err := s.service.GetPartByID(ctx, int64(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "запчасть обновлена, но не удалось получить результат: %v", err)
	}

	return partToProto(updated), nil
}

// DeletePart удаляет запчасть
func (s *partsGRPCServer) DeletePart(ctx context.Context, req *partsv1.DeletePartRequest) (*partsv1.DeletePartResponse, error) {
	if err := s.service.DeletePart(ctx, int64(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить запчасть: %v", err)
	}
	return &partsv1.DeletePartResponse{}, nil
}

// MarkPartForDeletion отмечает запчасть для отложенного удаления
func (s *partsGRPCServer) MarkPartForDeletion(ctx context.Context, req *partsv1.MarkPartForDeletionRequest) (*partsv1.Part, error) {
	if err := s.service.MarkPartForDeletion(ctx, int64(req.Id)); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось отметить запчасть для удаления: %v", err)
	}

	part, err := s.service.GetPartByID(ctx, int64(req.Id))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить запчасть: %v", err)
	}

	return partToProto(part), nil
}

// UploadPartPhoto — streaming upload фото
func (s *partsGRPCServer) UploadPartPhoto(stream partsv1.PartsService_UploadPartPhotoServer) error {
	var partID uint32
	var fileData []byte

	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return status.Errorf(codes.Internal, "ошибка чтения stream: %v", err)
		}
		partID = chunk.PartId
		fileData = append(fileData, chunk.Data...)
	}

	if partID == 0 {
		return status.Errorf(codes.InvalidArgument, "part_id обязателен")
	}

	// Сохраняем файл
	photoPath, err := SavePhotoFromBytes(int64(partID), fileData)
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось сохранить фото: %v", err)
	}

	// Получаем обновлённую запчасть
	part, err := s.service.GetPartByID(stream.Context(), int64(partID))
	if err != nil {
		return status.Errorf(codes.Internal, "фото сохранено, но не удалось получить запчасть: %v", err)
	}

	return stream.SendAndClose(&partsv1.UploadPartPhotoResponse{
		PhotoUrl: photoPath,
		Part:     partToProto(part),
	})
}

// DeletePartPhoto удаляет фото запчасти
func (s *partsGRPCServer) DeletePartPhoto(ctx context.Context, req *partsv1.DeletePartPhotoRequest) (*partsv1.DeletePartPhotoResponse, error) {
	if err := s.service.DeletePartPhoto(ctx, int64(req.PartId), req.PhotoUrl); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить фото: %v", err)
	}
	return &partsv1.DeletePartPhotoResponse{}, nil
}

// GetStatistics возвращает статистику
func (s *partsGRPCServer) GetStatistics(ctx context.Context, req *partsv1.GetStatisticsRequest) (*partsv1.StatisticsResponse, error) {
	stats, err := s.service.GetStatistics(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить статистику: %v", err)
	}

	categories := make([]*partsv1.CategoryCount, len(stats.Categories))
	for i, c := range stats.Categories {
		categories[i] = &partsv1.CategoryCount{
			Name:  c.Name,
			Count: int32(c.Count),
		}
	}

	monthlySales := make([]*partsv1.MonthlySales, len(stats.MonthlySales))
	for i, ms := range stats.MonthlySales {
		monthlySales[i] = &partsv1.MonthlySales{
			Month: ms.Month,
			Sales: ms.Sales,
		}
	}

	return &partsv1.StatisticsResponse{
		TotalParts:    int32(stats.TotalParts),
		TotalQuantity: int32(stats.TotalQuantity),
		TotalValue:    stats.TotalValue,
		TotalEarnings: stats.TotalEarnings,
		Categories:    categories,
		MonthlySales:  monthlySales,
	}, nil
}

// UpdateEarnings обновляет заработок
func (s *partsGRPCServer) UpdateEarnings(ctx context.Context, req *partsv1.UpdateEarningsRequest) (*partsv1.UpdateEarningsResponse, error) {
	if err := s.service.UpdateEarnings(ctx, req.Amount); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось обновить заработок: %v", err)
	}
	return &partsv1.UpdateEarningsResponse{TotalEarnings: req.Amount}, nil
}

func defectReportRequestFromProto(req *partsv1.CreateDefectReportRequest) DefectReportRequest {
	year, _ := strconv.Atoi(req.Year)
	mileage, _ := strconv.Atoi(req.Mileage)
	report := DefectReportRequest{
		Brand:             req.Brand,
		Model:             req.Model,
		Year:              year,
		CarReleasePeriod:  req.CarReleasePeriod,
		VIN:               req.Vin,
		Mileage:           mileage,
		Description:       req.Description,
		EngineBrand:       req.EngineBrand,
		BodyBrand:         req.BodyBrand,
		InteriorColor:     req.InteriorColor,
		BodyColor:         req.BodyColor,
		Transmission:      req.Transmission,
		TransmissionModel: req.TransmissionModel,
		Drive:             req.Drive,
		CatalogVersion:    req.CatalogVersion,
		SellerName:        req.Salesman,
		SellerID:          int64(req.SellerId),
		SelectedParts:     make([]DefectReportPart, 0, len(req.SelectedParts)),
	}

	for _, part := range req.SelectedParts {
		report.SelectedParts = append(report.SelectedParts, DefectReportPart{
			Name:               part.Name,
			Category:           part.Category,
			Description:        part.Description,
			Quantity:           int(part.Quantity),
			Price:              part.Price,
			Location:           part.Location,
			Address:            part.Address,
			BodyBrand:          part.BodyBrand,
			EngineBrand:        part.EngineBrand,
			CarReleaseDate:     part.CarReleaseDate,
			CarReleasePeriod:   part.CarReleasePeriod,
			FrontRear:          part.FrontRear,
			LeftRight:          part.LeftRight,
			TopBottom:          part.TopBottom,
			Number:             part.Number,
			Manufacturer:       part.Manufacturer,
			ManufacturerCode:   part.ManufacturerCode,
			OEMCode:            part.OemCode,
			Color:              part.Color,
			Condition:          part.Condition,
			SupplierCode:       part.SupplierCode,
			Defect:             part.Defect,
			Transmission:       part.Transmission,
			TransmissionModel:  part.TransmissionModel,
			Drive:              part.Drive,
			WearPercentage:     part.WearPercentage,
			Season:             part.Season,
			Diameter:           part.Diameter,
			Width:              part.Width,
			Profile:            part.Profile,
			TireQuantity:       part.TireQuantity,
			Drilling:           part.Drilling,
			Offset:             part.Offset,
			CenterHoleDiameter: part.CenterHoleDiameter,
			TireModel:          part.TireModel,
			VIN:                part.Vin,
		})
	}

	return report
}

// defectReportPartToProto — обратное преобразование для preview.
func defectReportPartToProto(part DefectReportPart) *partsv1.DefectReportPart {
	return &partsv1.DefectReportPart{
		Name:               part.Name,
		Category:           part.Category,
		Description:        part.Description,
		Quantity:           int32(part.Quantity),
		Price:              part.Price,
		Location:           part.Location,
		Address:            part.Address,
		BodyBrand:          part.BodyBrand,
		EngineBrand:        part.EngineBrand,
		CarReleaseDate:     part.CarReleaseDate,
		CarReleasePeriod:   part.CarReleasePeriod,
		FrontRear:          part.FrontRear,
		LeftRight:          part.LeftRight,
		TopBottom:          part.TopBottom,
		Number:             part.Number,
		Manufacturer:       part.Manufacturer,
		ManufacturerCode:   part.ManufacturerCode,
		OemCode:            part.OEMCode,
		Color:              part.Color,
		Condition:          part.Condition,
		SupplierCode:       part.SupplierCode,
		Defect:             part.Defect,
		Transmission:       part.Transmission,
		TransmissionModel:  part.TransmissionModel,
		Drive:              part.Drive,
		WearPercentage:     part.WearPercentage,
		Season:             part.Season,
		Diameter:           part.Diameter,
		Width:              part.Width,
		Profile:            part.Profile,
		TireQuantity:       part.TireQuantity,
		Drilling:           part.Drilling,
		Offset:             part.Offset,
		CenterHoleDiameter: part.CenterHoleDiameter,
		TireModel:          part.TireModel,
		Vin:                part.VIN,
	}
}

// PreviewDefectReport возвращает набор запчастей, который создаст CreateDefectReport,
// ничего не записывая и не публикуя. Набор строит сервер по каталогу.
func (s *partsGRPCServer) PreviewDefectReport(ctx context.Context, req *partsv1.PreviewDefectReportRequest) (*partsv1.PreviewDefectReportResponse, error) {
	report := DefectReportRequest{
		Brand:             req.Brand,
		Model:             req.Model,
		CarReleasePeriod:  req.CarReleasePeriod,
		VIN:               req.Vin,
		Description:       req.Description,
		EngineBrand:       req.EngineBrand,
		BodyBrand:         req.BodyBrand,
		InteriorColor:     req.InteriorColor,
		BodyColor:         req.BodyColor,
		Transmission:      req.Transmission,
		TransmissionModel: req.TransmissionModel,
		Drive:             req.Drive,
		CatalogVersion:    req.CatalogVersion,
	}
	report.Year, _ = strconv.Atoi(req.Year)
	report.Mileage, _ = strconv.Atoi(req.Mileage)

	prepared, err := s.defectReports.Preview(report)
	if err != nil {
		return nil, status.Errorf(codes.Unavailable, "не удалось построить дефектную ведомость: %v", err)
	}

	parts := make([]*partsv1.DefectReportPart, 0, len(prepared.SelectedParts))
	for _, part := range prepared.SelectedParts {
		parts = append(parts, defectReportPartToProto(part))
	}

	return &partsv1.PreviewDefectReportResponse{
		CatalogVersion: prepared.CatalogVersion,
		Total:          int32(len(parts)),
		Parts:          parts,
	}, nil
}

// CreateDefectReport использует тот же асинхронный workflow, что и HTTP endpoint.
func (s *partsGRPCServer) CreateDefectReport(ctx context.Context, req *partsv1.CreateDefectReportRequest) (*partsv1.CreateDefectReportResponse, error) {
	report := defectReportRequestFromProto(req)
	if err := s.defectReports.Enqueue(ctx, &report, true); err != nil {
		if defectReportUnavailable(err) {
			return nil, status.Errorf(codes.Unavailable, "дефектные ведомости временно недоступны: %v", err)
		}
		return nil, status.Errorf(codes.Internal, "не удалось поставить дефектную ведомость в очередь: %v", err)
	}

	return &partsv1.CreateDefectReportResponse{
		PartsQueued: int32(len(report.SelectedParts)),
		Message:     "Дефектная ведомость отправлена в очередь обработки",
	}, nil
}

// BulkDeleteParts массовое удаление
func (s *partsGRPCServer) BulkDeleteParts(ctx context.Context, req *partsv1.BulkDeletePartsRequest) (*partsv1.BulkDeletePartsResponse, error) {
	ids := make([]int64, len(req.Ids))
	for i, id := range req.Ids {
		ids[i] = int64(id)
	}

	if err := s.service.BulkDeleteParts(ctx, ids); err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось массово удалить запчасти: %v", err)
	}

	return &partsv1.BulkDeletePartsResponse{DeletedCount: int32(len(ids))}, nil
}

// BulkUpdateParts массовое обновление
func (s *partsGRPCServer) BulkUpdateParts(ctx context.Context, req *partsv1.BulkUpdatePartsRequest) (*partsv1.BulkUpdatePartsResponse, error) {
	updates := make([]map[string]interface{}, len(req.Parts))
	for i, p := range req.Parts {
		m := make(map[string]interface{})
		m["id"] = p.Id
		for k, v := range p.Fields {
			m[k] = v
		}
		updates[i] = m
	}

	count, err := s.service.BulkUpdateParts(ctx, updates)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось массово обновить запчасти: %v", err)
	}

	return &partsv1.BulkUpdatePartsResponse{UpdatedCount: int32(count)}, nil
}

// GetSupplierCodes возвращает коды поставщиков
func (s *partsGRPCServer) GetSupplierCodes(ctx context.Context, req *partsv1.GetSupplierCodesRequest) (*partsv1.SupplierCodesResponse, error) {
	supplierCodes, err := s.service.GetSupplierCodes(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить коды поставщиков: %v", err)
	}
	return &partsv1.SupplierCodesResponse{Codes: supplierCodes}, nil
}

// DeleteZeroQuantityParts удаляет запчасти с нулевым количеством
func (s *partsGRPCServer) DeleteZeroQuantityParts(ctx context.Context, req *partsv1.DeleteZeroQuantityPartsRequest) (*partsv1.DeleteZeroQuantityPartsResponse, error) {
	count, err := s.service.DeleteZeroQuantityPartsBySupplier(ctx, req.SupplierCode)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось удалить запчасти: %v", err)
	}
	return &partsv1.DeleteZeroQuantityPartsResponse{DeletedCount: int32(count)}, nil
}

// ─── Helpers ────────────────────────────────────────────────────────────────

func partToProto(p *Part) *partsv1.Part {
	proto := &partsv1.Part{
		Id:          uint32(p.ID),
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
		SellerId:    uint32(p.SellerID),
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

// ─── gRPC Server Startup ────────────────────────────────────────────────────

// StartGRPCServer запускает gRPC-сервер на указанном порту
func StartGRPCServer(service InventoryService, defectReports *DefectReportWorkflow, port string) error {
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", port, err)
	}

	srv := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			loggingUnaryInterceptor,
			recoveryUnaryInterceptor,
		),
	)

	// Регистрируем parts-сервис
	partsv1.RegisterPartsServiceServer(srv, NewPartsGRPCServer(service, defectReports))

	// Health check
	healthSrv := health.NewServer()
	healthpb.RegisterHealthServer(srv, healthSrv)
	healthSrv.SetServingStatus("parts.v1.PartsService", healthpb.HealthCheckResponse_SERVING)

	// Reflection для отладки
	reflection.Register(srv)

	logrus.WithField("port", port).Info("gRPC server listening")
	return srv.Serve(lis)
}

// ─── Interceptors ───────────────────────────────────────────────────────────

func loggingUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	start := time.Now()
	resp, err := handler(ctx, req)
	duration := time.Since(start)

	fields := logrus.Fields{
		"method":   info.FullMethod,
		"duration": duration.String(),
	}
	if err != nil {
		fields["error"] = err.Error()
		logrus.WithFields(fields).Warn("gRPC call failed")
	} else if duration > 100*time.Millisecond {
		logrus.WithFields(fields).Info("gRPC call slow")
	}

	return resp, err
}

func recoveryUnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp interface{}, err error) {
	defer func() {
		if r := recover(); r != nil {
			logrus.WithField("panic", r).WithField("method", info.FullMethod).Error("gRPC panic recovered")
			err = status.Errorf(codes.Internal, "internal server error")
		}
	}()
	return handler(ctx, req)
}

// ListPartsForExport отдаёт запчасти потоком: их больше ста тысяч, и в один
// gRPC-ответ они не помещаются. Потребитель — export-service, который собирает
// из них прайс-лист для Drom.
func (s *partsGRPCServer) ListPartsForExport(_ *partsv1.ListPartsForExportRequest, stream partsv1.PartsService_ListPartsForExportServer) error {
	parts, err := s.service.PartsForExport(stream.Context())
	if err != nil {
		return status.Errorf(codes.Internal, "не удалось получить запчасти для выгрузки: %v", err)
	}

	for i := range parts {
		if err := stream.Send(partToProto(&parts[i])); err != nil {
			return err
		}
	}
	return nil
}

// InventoryVersion отдаёт отпечаток состояния склада — тот же, что уходит в
// ETag обычных HTTP-ответов. Нужен, чтобы export-service мог спросить «менялось
// ли что-нибудь» одной дешёвой строкой, не вычитывая весь склад потоком.
func (s *partsGRPCServer) InventoryVersion(ctx context.Context, _ *partsv1.InventoryVersionRequest) (*partsv1.InventoryVersionResponse, error) {
	version, err := s.service.InventoryVersion(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "не удалось получить отпечаток склада: %v", err)
	}
	return &partsv1.InventoryVersionResponse{Version: version}, nil
}
