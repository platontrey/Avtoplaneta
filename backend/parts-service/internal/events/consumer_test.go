package events

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"

	"parts-service/internal/domain"
)

type mockInventoryService struct {
	updatedEarnings float64
	renamedSellerID int64
	renamedName     string
}

func (m *mockInventoryService) GetInventory(context.Context, domain.InventoryQueryParams) ([]domain.Part, error) {
	return nil, nil
}
func (m *mockInventoryService) AddPart(context.Context, *domain.Part) (*domain.Part, error) {
	return nil, nil
}
func (m *mockInventoryService) AddPartsBatch(context.Context, []domain.Part) ([]domain.Part, error) {
	return nil, nil
}
func (m *mockInventoryService) UpdatePart(context.Context, int64, map[string]interface{}) error {
	return nil
}
func (m *mockInventoryService) DeletePart(context.Context, int64) error         { return nil }
func (m *mockInventoryService) MarkPartForDeletion(context.Context, int64) error { return nil }
func (m *mockInventoryService) GetStatistics(context.Context) (domain.StatisticsResponse, error) {
	return domain.StatisticsResponse{}, nil
}
func (m *mockInventoryService) InventoryVersion(context.Context) (string, error) { return "", nil }
func (m *mockInventoryService) RenameSeller(_ context.Context, id int64, name string) ([]domain.Part, error) {
	m.renamedSellerID = id
	m.renamedName = name
	return nil, nil
}
func (m *mockInventoryService) PartsForExport(context.Context) ([]domain.Part, error) {
	return nil, nil
}
func (m *mockInventoryService) BulkDeleteParts(context.Context, []int64) error { return nil }
func (m *mockInventoryService) BulkUpdateParts(context.Context, []map[string]interface{}) (int, error) {
	return 0, nil
}
func (m *mockInventoryService) DeleteZeroQuantityPartsBySupplier(context.Context, string) (int64, error) {
	return 0, nil
}
func (m *mockInventoryService) GetSupplierCodes(context.Context) ([]string, error) { return nil, nil }
func (m *mockInventoryService) GetNextSupplierCode(context.Context) (string, error) { return "1", nil }
func (m *mockInventoryService) PeekNextSupplierCode(context.Context) (string, error) { return "1", nil }
func (m *mockInventoryService) UploadPartPhoto(context.Context, int64, *gin.Context) (string, error) {
	return "", nil
}
func (m *mockInventoryService) DeletePartPhoto(context.Context, int64, string) error { return nil }
func (m *mockInventoryService) SavePhotoFromBytes(context.Context, int64, []byte) (string, error) {
	return "", nil
}
func (m *mockInventoryService) GetPartByID(context.Context, int64) (*domain.Part, error) {
	return nil, nil
}
func (m *mockInventoryService) DecreasePartQuantity(context.Context, int64, int, string) error {
	return nil
}
func (m *mockInventoryService) IncreasePartQuantity(context.Context, int64, int, string) error {
	return nil
}
func (m *mockInventoryService) UpdateEarnings(_ context.Context, amount float64) error {
	m.updatedEarnings += amount
	return nil
}
func (m *mockInventoryService) ReindexAllParts(context.Context) error { return nil }
func (m *mockInventoryService) SetMonthlySalesProvider(func(context.Context) ([]domain.MonthlySales, error)) {
}

func TestProcessMessageOrderCompleted(t *testing.T) {
	svc := &mockInventoryService{}
	consumer := &RedisEventConsumer{
		service: svc,
	}

	msg := redis.XMessage{
		ID: "1-0",
		Values: map[string]interface{}{
			"type":     "order_completed",
			"order_id": "123",
			"amount":   "450.50",
		},
	}

	err := consumer.processMessage(context.Background(), msg)
	require.NoError(t, err)
	require.Equal(t, 450.50, svc.updatedEarnings)
}

func TestProcessMessageSellerRenamed(t *testing.T) {
	svc := &mockInventoryService{}
	consumer := &RedisEventConsumer{
		service: svc,
	}

	msg := redis.XMessage{
		ID: "2-0",
		Values: map[string]interface{}{
			"type":      "seller_renamed",
			"seller_id": "42",
			"name":      "New Seller Name",
		},
	}

	err := consumer.processMessage(context.Background(), msg)
	require.NoError(t, err)
	require.Equal(t, int64(42), svc.renamedSellerID)
	require.Equal(t, "New Seller Name", svc.renamedName)
}

func TestProcessMessageUnknownAndIgnored(t *testing.T) {
	consumer := &RedisEventConsumer{}

	err := consumer.processMessage(context.Background(), redis.XMessage{
		ID: "3-0",
		Values: map[string]interface{}{
			"type": "unknown_type",
		},
	})
	require.NoError(t, err)

	err = consumer.processMessage(context.Background(), redis.XMessage{
		ID: "4-0",
		Values: map[string]interface{}{
			"type": "defect_report_created",
		},
	})
	require.NoError(t, err)
}
