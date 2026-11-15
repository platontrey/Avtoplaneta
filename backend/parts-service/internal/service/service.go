package service

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"

	partsv1 "avtoplaneta/gen/parts/v1"
	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/redisclient"
	"avtoplaneta/pkg/userdirectory"
	"parts-service/internal/config"
	"parts-service/internal/domain"
	"parts-service/internal/repository"
	"parts-service/internal/search"
)

const (
	statisticsCacheKey             = "parts:statistics"
	statisticsCacheTTL             = 5 * time.Minute
	inventoryCacheKey              = "parts:inventory"
	inventoryCacheKeyWithoutPhotos = "parts:inventory:without_photos"
	inventoryCacheTTL              = 10 * time.Minute
)

// InventoryService определяет интерфейс для бизнес-логики управления инвентарем запчастей
// Содержит всю логику валидации, обработки и координации между репозиторием и внешними сервисами
type InventoryService interface {
	// GetInventory Основные операции с запчастями
	GetInventory(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, error)                  // Получает список запчастей с фильтрами
	GetInventoryWithCount(ctx context.Context, params domain.InventoryQueryParams) ([]domain.Part, int64, error) // Получает запчасти и общее количество
	AddPart(ctx context.Context, part *domain.Part) (*domain.Part, error)                                         // Добавляет новую запчасть
	DuplicatePart(ctx context.Context, id int64) (*domain.Part, error)                                           // Дублирует существующую запчасть без фото
	UpdatePart(ctx context.Context, id int64, updates map[string]interface{}) error                              // Обновляет существующую запчасть
	DeletePart(ctx context.Context, id int64) error                                                                 // Удаляет запчасть
	MarkPartForDeletion(ctx context.Context, id int64) error                                                        // Отмечает запчасть для отложенного удаления
	GetStatistics(ctx context.Context) (domain.StatisticsResponse, error)                                          // Получает статистику по инвентарю
	InventoryVersion(ctx context.Context) (string, error)                                                           // Отпечаток состояния склада для условных запросов
	RenameSeller(ctx context.Context, sellerID int64, name string) ([]domain.Part, error)                          // Приводит копию имени продавца в строках к справочнику
	PartsForExport(ctx context.Context) ([]domain.Part, error)                                                     // Запчасти для выгрузки прайс-листа

	// BulkDeleteParts Админ операции
	BulkDeleteParts(ctx context.Context, ids []int64) error                                         // Массовое удаление запчастей
	BulkDeletePartsWithProto(ctx context.Context, req *partsv1.BulkDeletePartsRequest) (int, error) // Массовое удаление запчастей по Protobuf спецификации
	BulkUpdateParts(ctx context.Context, req *partsv1.BulkUpdatePartsRequest) (int, error)          // Массовое обновление запчастей
	DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error)      // Удаление по поставщику
	GetSupplierCodes(ctx context.Context) ([]string, error)                                         // Получение кодов поставщиков
	GetNextSupplierCode(ctx context.Context) (string, error)                                        // Получение следующего номера поставки
	PeekNextSupplierCode(ctx context.Context) (string, error)                                       // Просмотр следующего номера поставки без инкремента

	// Полки склада
	GetShelves(ctx context.Context) ([]domain.Shelf, error)
	CreateShelf(ctx context.Context, name string) (*domain.Shelf, error)
	DeleteShelf(ctx context.Context, id int64) error

	// UploadPartPhoto Фото операции
	UploadPartPhoto(ctx context.Context, id int64, c *gin.Context) (string, error) // Загрузка фото запчасти
	DeletePartPhoto(ctx context.Context, id int64, photoPath string) error         // Удаление фото запчасти (если photoPath пустой - удаляет все)
	SavePhotoFromBytes(ctx context.Context, id int64, data []byte) (string, error)

	// GetPartByID Получение запчасти по ID
	GetPartByID(ctx context.Context, id int64) (*domain.Part, error)
	AddPartsBatch(ctx context.Context, parts []domain.Part) ([]domain.Part, error)

	DecreasePartQuantity(ctx context.Context, id int64, amount int, operationID string) error
	IncreasePartQuantity(ctx context.Context, id int64, amount int, operationID string) error

	// UpdateEarnings Обновление общего заработка
	UpdateEarnings(ctx context.Context, amount float64) error
	ReindexAllParts(ctx context.Context) error
	SetMonthlySalesProvider(fn func(ctx context.Context) ([]domain.MonthlySales, error))
}

// inventoryService реализует InventoryService
type inventoryService struct {
	repo                 repository.PartRepository
	es                   search.ElasticsearchClient
	redis                *redis.Client
	users                *userdirectory.Directory
	totalEarnings        float64
	queryPool            sync.Pool // Pool для повторного использования map для Elasticsearch queries
	monthlySalesProvider func(ctx context.Context) ([]domain.MonthlySales, error)
}

// NewInventoryService создает новый сервис инвентаря
func NewInventoryService(repo repository.PartRepository, es search.ElasticsearchClient, config *config.Config, users *userdirectory.Directory) InventoryService {
	// Инициализация Redis клиента с настройками для IPv4
	rdb := redis.NewClient(&redis.Options{
		Addr:         config.RedisURL,
		Network:      "tcp", // Явно указываем TCP для IPv4
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	})
	redisclient.Instrument(rdb)

	service := &inventoryService{
		repo:  repo,
		es:    es,
		redis: rdb,
		users: users,
		queryPool: sync.Pool{
			New: func() interface{} {
				return make(map[string]interface{})
			},
		},
	}

	// Инициализируем totalEarnings из базы данных
	ctx := context.Background()
	if earnings, err := repo.GetTotalEarnings(ctx); err == nil {
		service.totalEarnings = earnings
	} else {
		slog.WarnContext(ctx, "Failed to load total earnings from database, starting with 0", logger.Err(err))
		service.totalEarnings = 0
	}

	return service
}

// syncPartIndex переиндексирует запчасть в Elasticsearch, отправляет событие в Redis и сбрасывает кэш
func (s *inventoryService) syncPartIndex(ctx context.Context, id int64) {
	updatedPart, err := s.repo.FindByID(ctx, id)
	if err != nil || updatedPart == nil {
		slog.WarnContext(ctx, "syncPartIndex: failed to find part for reindexing", logger.Err(err), slog.Int64("part_id", id))
		return
	}

	// Синхронная прямая индексация в Elasticsearch (Refresh: "true") для мгновенной видимости в фильтрах "с фото"
	if err := search.IndexPart(updatedPart); err != nil {
		slog.WarnContext(ctx, "syncPartIndex: failed to index part in Elasticsearch", logger.Err(err), slog.Int64("part_id", id))
	}

	// Отправка события в Redis Stream events:orders
	if s.redis != nil {
		err := s.redis.XAdd(ctx, &redis.XAddArgs{
			Stream: constants.StreamOrdersEvents,
			Values: map[string]interface{}{
				"type":    "part_index_requested",
				"part_id": fmt.Sprintf("%d", id),
			},
		}).Err()
		if err != nil {
			slog.WarnContext(ctx, "syncPartIndex: failed to publish part_index_requested to Redis", logger.Err(err))
		}

		s.redis.Del(ctx, inventoryCacheKey)
		s.redis.Del(ctx, inventoryCacheKeyWithoutPhotos)
		s.invalidateStatisticsCache(ctx)
	}
}

// UploadPartPhoto загружает фото и синхронизирует поисковый индекс
func (s *inventoryService) UploadPartPhoto(ctx context.Context, id int64, c *gin.Context) (string, error) {
	photoPath, err := HandlePhotoUpload(ctx, c, id, s.repo)
	if err != nil {
		return "", err
	}
	s.syncPartIndex(ctx, id)
	return photoPath, nil
}

// DeletePartPhoto удаляет фото и синхронизирует поисковый индекс
func (s *inventoryService) DeletePartPhoto(ctx context.Context, id int64, photoPath string) error {
	if err := DeletePhoto(ctx, id, photoPath, s.repo); err != nil {
		return err
	}
	s.syncPartIndex(ctx, id)
	return nil
}

// SavePhotoFromBytes сохраняет фото из байтов (для gRPC streaming upload) и синхронизирует поисковый индекс
func (s *inventoryService) SavePhotoFromBytes(ctx context.Context, id int64, data []byte) (string, error) {
	photoPath, err := SavePhotoFromBytes(ctx, id, data, s.repo)
	if err != nil {
		return "", err
	}
	s.syncPartIndex(ctx, id)
	return photoPath, nil
}

// InventoryVersion пробрасывает отпечаток склада из репозитория. Кэшировать его
// в Redis не нужно: запрос и так дешёвый, а лишний слой добавил бы окно, в
// котором клиент получал бы 304 на уже изменившиеся данные.
func (s *inventoryService) InventoryVersion(ctx context.Context) (string, error) {
	return s.repo.InventoryVersion(ctx)
}

// RenameSeller пробрасывает починку копии имени продавца в репозиторий.
func (s *inventoryService) RenameSeller(ctx context.Context, sellerID int64, name string) ([]domain.Part, error) {
	return s.repo.RenameSeller(ctx, sellerID, name)
}

// PartsForExport отдаёт запчасти, пригодные для выгрузки прайс-листа.
// Сам прайс-лист собирает export-service; сервис запчастей только отдаёт данные.
func (s *inventoryService) PartsForExport(ctx context.Context) ([]domain.Part, error) {
	return s.repo.GetPartsForXML(ctx)
}

// ReindexAllParts переиндексирует все запчасти в Elasticsearch
func (s *inventoryService) ReindexAllParts(ctx context.Context) error {
	parts, err := s.repo.FindAll(ctx)
	if err != nil {
		return fmt.Errorf("не удалось получить запчасти: %w", err)
	}
	if s.es == nil {
		return nil
	}
	return search.BulkIndexParts(ctx, parts)
}
