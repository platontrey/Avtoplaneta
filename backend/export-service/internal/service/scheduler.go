package service

import (
	"context"
	"log/slog"
	"time"

	"avtoplaneta/pkg/logger"
)

const priceListCheckInterval = time.Hour

// StartPriceListScheduler периодически проверяет актуальность XML прайс-листа и пересобирает его при изменениях.
// Вызов блокирует текущую горутину до отмены контекста.
func StartPriceListScheduler(ctx context.Context, svc *PriceListService, readyChan chan struct{}) {
	slog.InfoContext(ctx, "Планировщик прайс-листа запущен")

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
					slog.InfoContext(ctx, "XML прайс-лист успешно инициализирован при старте", slog.Int("parts_count", meta.PartsCount))
				}
				if readyChan != nil {
					select {
					case <-readyChan:
					default:
						close(readyChan)
					}
				}
				return
			}
			slog.WarnContext(ctx, "parts-service еще не готов для первичной сборки слепка, повтор через 5 секунд...", logger.Err(err))
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
			slog.InfoContext(ctx, "Планировщик прайс-листа остановлен")
			return
		case <-ticker.C:
			rebuildIfStale(ctx, svc)
		}
	}
}

func rebuildIfStale(ctx context.Context, svc *PriceListService) {
	meta, rebuilt, err := svc.Build(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Не удалось собрать XML прайс-лист по расписанию", logger.Err(err))
		return
	}
	if !rebuilt {
		slog.DebugContext(ctx, "XML прайс-лист актуален, плановая пересборка пропущена")
		return
	}

	slog.InfoContext(ctx, "XML прайс-лист пересобран по расписанию", slog.Int("parts_count", meta.PartsCount))
}
