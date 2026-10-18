package grpc

import (
	"context"
	"fmt"
	"time"

	partsv1 "avtoplaneta/gen/parts/v1"
	"avtoplaneta/pkg/grpcclient"
	"orders-service/internal/domain"

	"google.golang.org/grpc"
)

// PartsGRPCClient интерфейс для работы с запчастями по gRPC
type PartsGRPCClient interface {
	GetPartByID(ctx context.Context, id int64) (*domain.Part, error)
	DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	DeletePart(ctx context.Context, id int64) error
	Close() error
}

type partsGRPCClient struct {
	conn   *grpc.ClientConn
	client partsv1.PartsServiceClient
}

// NewPartsGRPCClient создает новый клиент к parts-service со стандартным keepalive
func NewPartsGRPCClient(target string) (PartsGRPCClient, error) {
	conn, err := grpcclient.NewClient(target)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to parts service at %s: %w", target, err)
	}

	return &partsGRPCClient{
		conn:   conn,
		client: partsv1.NewPartsServiceClient(conn),
	}, nil
}

func (c *partsGRPCClient) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

func (c *partsGRPCClient) GetPartByID(ctx context.Context, id int64) (*domain.Part, error) {
	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &partsv1.GetPartRequest{Id: id}
	resp, err := c.client.GetPart(callCtx, req)
	if err != nil {
		return nil, err
	}

	p := resp.GetPart()
	photo := ""
	if len(p.GetPhotos()) > 0 {
		photo = p.GetPhotos()[0]
	}

	return &domain.Part{
		ID:       p.GetId(),
		Name:     p.GetName(),
		Quantity: int(p.GetQuantity()),
		Price:    p.GetPrice(),
		Location: p.GetLocation(),
		Photo:    photo,
	}, nil
}

func (c *partsGRPCClient) DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &partsv1.DecreasePartQuantityRequest{
		Id:          id,
		Amount:      int32(amount),
		OperationId: operationID,
	}
	_, err := c.client.DecreasePartQuantity(callCtx, req)
	return err
}

func (c *partsGRPCClient) IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error {
	callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req := &partsv1.IncreasePartQuantityRequest{
		Id:          id,
		Amount:      int32(amount),
		OperationId: operationID,
	}
	_, err := c.client.IncreasePartQuantity(callCtx, req)
	return err
}

func (c *partsGRPCClient) DeletePart(ctx context.Context, id int64) error {
	callCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	req := &partsv1.DeletePartRequest{
		Id: id,
	}
	_, err := c.client.DeletePart(callCtx, req)
	return err
}
