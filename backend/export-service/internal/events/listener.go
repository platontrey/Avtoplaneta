package events

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/redis/go-redis/v9"

	"avtoplaneta/pkg/constants"
	"avtoplaneta/pkg/logger"
	"avtoplaneta/pkg/redisstream"
	"export-service/internal/service"
)

const (
	partsEventsStream = constants.StreamOrdersEvents
	exportEventsGroup = "export-catalog-sync-group"
)

// StartCatalogEventListener подписывается на поток изменений склада из parts-service
// через надежный инфраструктурный воркер pkg/redisstream с поддержкой ReadyChan.
func StartCatalogEventListener(ctx context.Context, client *redis.Client, svc *service.PriceListService, readyChan <-chan struct{}) {
	if client == nil {
		slog.InfoContext(ctx, "Redis не подключен: инкрементальный слушатель событий каталога отключен")
		return
	}

	handler := func(ctx context.Context, msg redis.XMessage) error {
		eventType, _ := msg.Values["type"].(string)
		partIDStr, _ := msg.Values["part_id"].(string)

		if partIDStr == "" {
			return nil
		}

		partID, parseErr := strconv.ParseInt(partIDStr, 10, 64)
		if parseErr != nil {
			return nil
		}

		switch eventType {
		case "part_index_requested":
			if updateErr := svc.UpdatePartInSnapshot(ctx, partID); updateErr != nil {
				slog.WarnContext(ctx, "Ошибка инкрементального обновления детали в слепке",
					logger.Err(updateErr),
					slog.Int64("part_id", partID),
				)
				return updateErr
			}
			slog.InfoContext(ctx, "Деталь точечно обновлена в слепке и XML", slog.Int64("part_id", partID))
		case "part_delete_requested":
			if delErr := svc.DeletePartFromSnapshot(ctx, partID); delErr != nil {
				slog.WarnContext(ctx, "Ошибка удаления детали из слепка",
					logger.Err(delErr),
					slog.Int64("part_id", partID),
				)
				return delErr
			}
			slog.InfoContext(ctx, "Деталь точечно удалена из слепка и XML", slog.Int64("part_id", partID))
		}

		return nil
	}

	worker := redisstream.NewConsumerGroupWorker(client, handler, redisstream.ConsumerConfig{
		Stream:    partsEventsStream,
		Group:     exportEventsGroup,
		ReadyChan: readyChan,
	})

	_ = worker.Start(ctx)
}
