package service

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

const priceListCheckInterval = time.Hour

// StartPriceListScheduler периодически проверяет актуальность XML прайс-листа и пересобирает его при изменениях.
// Вызов блокирует текущую горутину до отмены контекста.
func StartPriceListScheduler(ctx context.Context, svc *PriceListService) {
	logrus.Info("Планировщик прайс-листа запущен")

	// Асинхронно добиваемся первичной сборки слепка, если parts-service был временно недоступен
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
			}
			meta, rebuilt, err := svc.Build(ctx)
			if err == nil {
				if rebuilt {
					logrus.WithField("parts_count", meta.PartsCount).Info("XML прайс-лист успешно инициализирован при старте")
				}
				return
			}
			logrus.WithError(err).Warn("parts-service еще не готов для первичной сборки слепка, повтор через 5 секунд...")
			select {
			case <-ctx.Done():
				return
			case <-time.After(5 * time.Second):
			}
		}
	}()

	ticker := time.NewTicker(priceListCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Планировщик прайс-листа остановлен")
			return
		case <-ticker.C:
			rebuildIfStale(ctx, svc)
		}
	}
}

func rebuildIfStale(ctx context.Context, svc *PriceListService) {
	meta, rebuilt, err := svc.Build(ctx)
	if err != nil {
		logrus.WithError(err).Error("Не удалось собрать XML прайс-лист по расписанию")
		return
	}
	if !rebuilt {
		logrus.Debug("XML прайс-лист актуален, плановая пересборка пропущена")
		return
	}

	logrus.WithField("parts_count", meta.PartsCount).Info("XML прайс-лист пересобран по расписанию")
}
