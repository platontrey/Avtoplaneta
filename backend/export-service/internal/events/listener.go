package events

import (
	"context"
	"strconv"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"avtoplaneta/pkg/redisstream"
	"export-service/internal/service"
)

const (
	partsEventsStream = "events:orders"
	exportEventsGroup = "export-catalog-sync-group"
)

// StartCatalogEventListener подписывается на поток изменений склада из parts-service
// через надежный инфраструктурный воркер pkg/redisstream с поддержкой ReadyChan.
func StartCatalogEventListener(ctx context.Context, client *redis.Client, svc *service.PriceListService, readyChan <-chan struct{}) {
	if client == nil {
		logrus.Info("Redis не подключен: инкрементальный слушатель событий каталога отключен")
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
				logrus.WithError(updateErr).WithField("part_id", partID).Warn("Ошибка инкрементального обновления детали в слепке")
				return updateErr
			}
			logrus.WithField("part_id", partID).Info("Деталь точечно обновлена в слепке и XML")
		case "part_delete_requested":
			if delErr := svc.DeletePartFromSnapshot(ctx, partID); delErr != nil {
				logrus.WithError(delErr).WithField("part_id", partID).Warn("Ошибка удаления детали из слепка")
				return delErr
			}
			logrus.WithField("part_id", partID).Info("Деталь точечно удалена из слепка и XML")
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
