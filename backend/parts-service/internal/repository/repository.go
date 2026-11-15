package repository

import (
	"context"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgxpool"

	partsv1 "avtoplaneta/gen/parts/v1"
	"parts-service/db/sqlc"
	"parts-service/internal/domain"
)

// PartRepository определяет контракт для доступа к данным запчастей
type PartRepository interface {
	Create(ctx context.Context, part *domain.Part) error
	CreateBatch(ctx context.Context, parts []domain.Part) ([]domain.Part, error)
	FindByID(ctx context.Context, id int64) (*domain.Part, error)
	FindAll(ctx context.Context) ([]domain.Part, error)
	Update(ctx context.Context, id int64, updates map[string]interface{}) error
	Delete(ctx context.Context, id int64) error

	DecreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error
	IncreaseQuantity(ctx context.Context, id int64, amount int, operationID string) error

	FindWithFilters(ctx context.Context, filters map[string]interface{}, offset, limit int) ([]domain.Part, error)
	CountWithFilters(ctx context.Context, filters map[string]interface{}) (int64, error)

	MarkForDeletion(ctx context.Context, id int64, deleteAt time.Time) error
	DeleteExpiredParts(ctx context.Context, before time.Time) error
	GetStatistics(ctx context.Context) (domain.StatisticsResponse, error)
	InventoryVersion(ctx context.Context) (string, error)
	RenameSeller(ctx context.Context, sellerID int64, name string) ([]domain.Part, error)
	BulkDelete(ctx context.Context, ids []int64) error
	BulkUpdate(ctx context.Context, updates []map[string]interface{}) (int, error)
	BulkUpdateWithProto(ctx context.Context, req *partsv1.BulkUpdatePartsRequest, filterMap map[string]interface{}) ([]domain.Part, error)

	DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error)
	GetSupplierCodes(ctx context.Context) ([]string, error)
	GetNextSupplierCode(ctx context.Context) (string, error)
	PeekNextSupplierCode(ctx context.Context) (string, error)
	GetNextPartNumber(ctx context.Context) (string, error)

	GetTotalEarnings(ctx context.Context) (float64, error)
	UpdateTotalEarnings(ctx context.Context, amount float64) error

	UpdatePartPhotos(ctx context.Context, id int64, photos domain.StringArray) error
	GetPartsForXML(ctx context.Context) ([]domain.Part, error)
	GetLastCreatedPart(ctx context.Context) (*domain.Part, error)

	// Shelves
	GetShelves(ctx context.Context) ([]domain.Shelf, error)
	CreateShelf(ctx context.Context, name string) (*domain.Shelf, error)
	DeleteShelf(ctx context.Context, id int64) error

	// Outbox
	FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]sqlc.OutboxEvent, error)
	MarkOutboxEventProcessed(ctx context.Context, id int64) error
	MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error
	CleanupProcessedOutboxEvents(ctx context.Context) error
}

// partRepository реализует PartRepository
type partRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
	psq     squirrel.StatementBuilderType
}

// NewPartRepository создает новый экземпляр репозитория
func NewPartRepository(pool *pgxpool.Pool) PartRepository {
	return &partRepository{
		pool:    pool,
		queries: sqlc.New(pool),
		psq:     squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}
