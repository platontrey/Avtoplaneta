package grpc

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authv1 "avtoplaneta/gen/auth/v1"
	partsv1 "avtoplaneta/gen/parts/v1"
	"export-service/internal/domain"
)

// Clients инкапсулирует gRPC соединения с parts-service и auth-service
type Clients struct {
	parts     partsv1.PartsServiceClient
	auth      authv1.AuthServiceClient
	partsConn *grpc.ClientConn
	authConn  *grpc.ClientConn
}

// NewClients создает клиентские gRPC соединения
func NewClients(partsTarget, authTarget string) (*Clients, error) {
	partsConn, err := grpc.NewClient(partsTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("подключение к parts-service (%s): %w", partsTarget, err)
	}

	authConn, err := grpc.NewClient(authTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		_ = partsConn.Close()
		return nil, fmt.Errorf("подключение к auth-service (%s): %w", authTarget, err)
	}

	return &Clients{
		parts:     partsv1.NewPartsServiceClient(partsConn),
		auth:      authv1.NewAuthServiceClient(authConn),
		partsConn: partsConn,
		authConn:  authConn,
	}, nil
}

// Close закрывает сетевые соединения
func (c *Clients) Close() {
	if c.partsConn != nil {
		_ = c.partsConn.Close()
	}
	if c.authConn != nil {
		_ = c.authConn.Close()
	}
}

// InventoryVersion запрашивает легковесный отпечаток склада
func (c *Clients) InventoryVersion(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.parts.InventoryVersion(ctx, &partsv1.InventoryVersionRequest{})
	if err != nil {
		return "", fmt.Errorf("запросить отпечаток склада: %w", err)
	}
	return resp.GetVersion(), nil
}

// PartsForExport вычитывает все запчасти склада потоком
func (c *Clients) PartsForExport(ctx context.Context) ([]domain.Part, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	stream, err := c.parts.ListPartsForExport(ctx, &partsv1.ListPartsForExportRequest{})
	if err != nil {
		return nil, fmt.Errorf("запросить поток запчастей: %w", err)
	}

	var parts []domain.Part
	for {
		item, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("читать запчасть из потока: %w", err)
		}
		if item.GetPart() != nil {
			parts = append(parts, partFromProto(item.GetPart()))
		}
	}
	return parts, nil
}

// GetPart запрашивает одну деталь по первичному ключу для точечного обновления
func (c *Clients) GetPart(ctx context.Context, partID int64) (*domain.Part, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	resp, err := c.parts.GetPart(ctx, &partsv1.GetPartRequest{Id: partID})
	if err != nil {
		return nil, fmt.Errorf("запросить деталь %d: %w", partID, err)
	}
	part := partFromProto(resp.GetPart())
	return &part, nil
}

// SellerINNs возвращает мапу ИНН продавцов из auth-service
func (c *Clients) SellerINNs(ctx context.Context) (map[int64]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	resp, err := c.auth.GetUsers(ctx, &authv1.GetUsersRequest{})
	if err != nil {
		return nil, fmt.Errorf("запросить продавцов: %w", err)
	}

	inns := make(map[int64]string, len(resp.GetUsers()))
	for _, user := range resp.GetUsers() {
		if user.GetInn() != "" {
			inns[user.GetId()] = user.GetInn()
		}
	}
	return inns, nil
}

func partFromProto(p *partsv1.Part) domain.Part {
	part := domain.Part{
		ID:             p.GetId(),
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
		SellerID:       p.GetSellerId(),
		BodyBrand:      p.GetBodyBrand(),
		EngineBrand:    p.GetEngineBrand(),
		CarReleaseDate: p.GetCarReleaseDate(),
		FrontRear:      p.GetFrontRear(),
		LeftRight:      p.GetLeftRight(),
		TopBottom:      p.GetTopBottom(),
		Manufacturer:   p.GetManufacturer(),
		OEMCode:        p.GetOemCode(),
		Color:          p.GetColor(),
		Condition:          p.GetCondition(),
		SupplierCode:       p.GetSupplierCode(),

		// Параметры шин и дисков
		Season:             p.GetSeason(),
		Diameter:           p.GetDiameter(),
		Width:              p.GetWidth(),
		Profile:            p.GetProfile(),
		TireQuantity:       p.GetTireQuantity(),
		Drilling:           p.GetDrilling(),
		Offset:             p.GetOffset(),
		CenterHoleDiameter: p.GetCenterHoleDiameter(),
		TireModel:          p.GetTireModel(),
	}

	if ts := p.GetToDeleteAt(); ts != nil {
		deleteAt := ts.AsTime()
		part.ToDeleteAt = &deleteAt
	}

	return part
}
