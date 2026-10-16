package events

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"

	"export-service/internal/service"
)

const (
	partsEventsStream = "events:orders"
	exportEventsGroup = "export-catalog-sync-group"
)

// StartCatalogEventListener подписывается на поток изменений склада из parts-service
// и точечно обновляет инкрементальный слепок bbolt и XML на диске.
func StartCatalogEventListener(ctx context.Context, client *redis.Client, svc *service.PriceListService) {
	if client == nil {
		logrus.Info("Redis не подключен: инкрементальный слушатель событий каталога отключен")
		return
	}

	err := client.XGroupCreateMkStream(ctx, partsEventsStream, exportEventsGroup, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		logrus.WithError(err).Warn("Не удалось создать Consumer Group для каталога в Redis Streams")
	}

	hostname, _ := os.Hostname()
	if hostname == "" {
		hostname = "export-events"
	}
	consumerName := fmt.Sprintf("%s-%d", hostname, os.Getpid())

	logrus.WithFields(logrus.Fields{
		"stream":   partsEventsStream,
		"group":    exportEventsGroup,
		"consumer": consumerName,
	}).Info("Запущен слушатель событий склада для инкрементального обновления XML")

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Слушатель событий каталога остановлен")
			return
		default:
			streams, err := client.XReadGroup(ctx, &redis.XReadGroupArgs{
				Group:    exportEventsGroup,
				Consumer: consumerName,
				Streams:  []string{partsEventsStream, ">"},
				Count:    10,
				Block:    3 * time.Second,
			}).Result()

			if err != nil && !errors.Is(err, redis.Nil) {
				time.Sleep(time.Second)
				continue
			}

			if len(streams) == 0 {
				continue
			}

			for _, stream := range streams {
				for _, msg := range stream.Messages {
					eventType, _ := msg.Values["type"].(string)
					partIDStr, _ := msg.Values["part_id"].(string)

					if partIDStr != "" {
						partID, parseErr := strconv.ParseInt(partIDStr, 10, 64)
						if parseErr == nil {
							switch eventType {
							case "part_index_requested":
								if updateErr := svc.UpdatePartInSnapshot(ctx, partID); updateErr != nil {
									logrus.WithError(updateErr).WithField("part_id", partID).Warn("Ошибка инкрементального обновления детали в слепке")
								} else {
									logrus.WithField("part_id", partID).Info("Деталь точечно обновлена в слепке и XML")
								}
							case "part_delete_requested":
								if delErr := svc.DeletePartFromSnapshot(ctx, partID); delErr != nil {
									logrus.WithError(delErr).WithField("part_id", partID).Warn("Ошибка удаления детали из слепка")
								} else {
									logrus.WithField("part_id", partID).Info("Деталь точечно удалена из слепка и XML")
								}
							}
						}
					}

					_ = client.XAck(ctx, partsEventsStream, exportEventsGroup, msg.ID).Err()
				}
			}
		}
	}
}
