package main

import (
	"context"
	"fmt"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	authv1 "avtoplaneta/gen/auth/v1"
	partsv1 "avtoplaneta/gen/parts/v1"
)

// Clients — соединения с сервисами, из которых собирается прайс-лист.
// Своей базы у export-service нет намеренно: он ничем не владеет, он только
// собирает документ из чужих данных.
type Clients struct {
	parts partsv1.PartsServiceClient
	auth  authv1.AuthServiceClient

	partsConn *grpc.ClientConn
	authConn  *grpc.ClientConn
}

func NewClients(partsTarget, authTarget string) (*Clients, error) {
	partsConn, err := grpc.NewClient(partsTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("подключение к parts-service (%s): %w", partsTarget, err)
	}

	authConn, err := grpc.NewClient(authTarget, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		partsConn.Close()
		return nil, fmt.Errorf("подключение к auth-service (%s): %w", authTarget, err)
	}

	return &Clients{
		parts:     partsv1.NewPartsServiceClient(partsConn),
		auth:      authv1.NewAuthServiceClient(authConn),
		partsConn: partsConn,
		authConn:  authConn,
	}, nil
}

func (c *Clients) Close() {
	if c.partsConn != nil {
		c.partsConn.Close()
	}
	if c.authConn != nil {
		c.authConn.Close()
	}
}

// InventoryVersion спрашивает у parts-service отпечаток состояния склада.
// Дешёвый вопрос: одна агрегирующая строка вместо выкачивания всего склада.
func (c *Clients) InventoryVersion(ctx context.Context) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	response, err := c.parts.InventoryVersion(ctx, &partsv1.InventoryVersionRequest{})
	if err != nil {
		return "", fmt.Errorf("запросить отпечаток склада: %w", err)
	}
	return response.GetVersion(), nil
}

// PartsForExport вычитывает поток запчастей целиком.
func (c *Clients) PartsForExport(ctx context.Context) ([]Part, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	stream, err := c.parts.ListPartsForExport(ctx, &partsv1.ListPartsForExportRequest{})
	if err != nil {
		return nil, fmt.Errorf("запросить запчасти: %w", err)
	}

	var parts []Part
	for {
		item, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("читать поток запчастей: %w", err)
		}
		parts = append(parts, partFromProto(item))
	}
	return parts, nil
}

// SellerINNs возвращает ИНН продавцов: в прайс-листе каждое предложение
// привязано к конкретному ИП или юрлицу, это обязательный реквизит площадки.
//
// Реквизиты живут в auth-service и спрашиваются у него по идентификатору —
// копий здесь не заводим, иначе получим ту же проблему, что была с именем.
func (c *Clients) SellerINNs(ctx context.Context) (map[int64]string, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	response, err := c.auth.GetUsers(ctx, &authv1.GetUsersRequest{})
	if err != nil {
		return nil, fmt.Errorf("запросить пользователей: %w", err)
	}

	inns := make(map[int64]string, len(response.Users))
	for _, user := range response.Users {
		if user.Inn != "" {
			inns[int64(user.Id)] = user.Inn
		}
	}
	return inns, nil
}
