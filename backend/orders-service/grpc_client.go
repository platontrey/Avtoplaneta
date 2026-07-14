package main

import (
	"context"
	"fmt"

	partsv1 "avtoplaneta/gen/parts/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// PartsGRPCClient интерфейс для работы с запчастями по gRPC
type PartsGRPCClient interface {
	GetPartByID(ctx context.Context, id int64) (*Part, error)
	DecreaseQuantity(ctx context.Context, id int64, amount int) error
	IncreaseQuantity(ctx context.Context, id int64, amount int) error
	DeletePart(ctx context.Context, id int64) error
	Close() error
}

type partsGRPCClient struct {
	conn   *grpc.ClientConn
	client partsv1.PartsServiceClient
}

// NewPartsGRPCClient создает новый клиент к parts-service
func NewPartsGRPCClient(target string) (PartsGRPCClient, error) {
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to parts service at %s: %w", target, err)
	}

	return &partsGRPCClient{
		conn:   conn,
		client: partsv1.NewPartsServiceClient(conn),
	}, nil
}

func (c *partsGRPCClient) Close() error {
	return c.conn.Close()
}

func (c *partsGRPCClient) GetPartByID(ctx context.Context, id int64) (*Part, error) {
	req := &partsv1.GetPartRequest{Id: uint32(id)}
	resp, err := c.client.GetPart(ctx, req)
	if err != nil {
		return nil, err
	}
	
	photo := ""
	if len(resp.Photos) > 0 {
		photo = resp.Photos[0]
	}

	return &Part{
		ID:       int64(resp.Id),
		Quantity: int(resp.Quantity),
		Price:    resp.Price,
		Location: resp.Location,
		Photo:    photo,
	}, nil
}

func (c *partsGRPCClient) DecreaseQuantity(ctx context.Context, id int64, amount int) error {
	req := &partsv1.ChangePartQuantityRequest{
		Id:     uint32(id),
		Amount: int32(amount),
	}
	_, err := c.client.DecreasePartQuantity(ctx, req)
	return err
}

func (c *partsGRPCClient) IncreaseQuantity(ctx context.Context, id int64, amount int) error {
	req := &partsv1.ChangePartQuantityRequest{
		Id:     uint32(id),
		Amount: int32(amount),
	}
	_, err := c.client.IncreasePartQuantity(ctx, req)
	return err
}

func (c *partsGRPCClient) DeletePart(ctx context.Context, id int64) error {
	req := &partsv1.DeletePartRequest{
		Id: uint32(id),
	}
	_, err := c.client.DeletePart(ctx, req)
	return err
}
