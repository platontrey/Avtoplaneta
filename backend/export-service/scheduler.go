package main

import (
	"context"
	"time"

	"github.com/sirupsen/logrus"
)

// Раньше расписание жило в Redis: ключ xml_last_generated_at помнил, когда была
// последняя генерация, и раз в две недели прайс-лист пересобирался целиком —
// даже если склад не менялся, и не пересобирался, если менялся.
//
// Теперь состояние хранит сам прайс-лист: рядом с ним лежит отпечаток склада,
// по которому видно, актуален файл или нет. Планировщику остаётся только
// регулярно спрашивать, и отдельное хранилище ему больше не нужно.
const priceListCheckInterval = time.Hour

// StartPriceListScheduler держит прайс-лист в актуальном состоянии.
// Вызов блокирующий, до отмены контекста.
func StartPriceListScheduler(ctx context.Context, builder *PriceListBuilder) {
	logrus.Info("Планировщик прайс-листа запущен")

	rebuildIfStale(ctx, builder)

	ticker := time.NewTicker(priceListCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logrus.Info("Планировщик прайс-листа остановлен")
			return
		case <-ticker.C:
			rebuildIfStale(ctx, builder)
		}
	}
}

func rebuildIfStale(ctx context.Context, builder *PriceListBuilder) {
	meta, rebuilt, err := builder.Build(ctx)
	if err != nil {
		logrus.WithError(err).Error("Не удалось собрать XML прайс-лист")
		return
	}
	if !rebuilt {
		logrus.Debug("XML прайс-лист актуален, пересборка пропущена")
		return
	}

	logrus.WithField("parts_count", meta.PartsCount).Info("XML прайс-лист пересобран")
}
