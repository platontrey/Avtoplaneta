package service

import (
	"context"
	"encoding/json"
	"log/slog"
	"math"
	"sort"

	"avtoplaneta/pkg/logger"
	"parts-service/internal/domain"
)

// GetStatistics получает статистику с кэшированием
func (s *inventoryService) GetStatistics(ctx context.Context) (domain.StatisticsResponse, error) {
	// Проверяем кэш
	cachedData, err := s.redis.Get(ctx, statisticsCacheKey).Result()
	if err == nil {
		// Данные найдены в кэше
		var stats domain.StatisticsResponse
		if err := json.Unmarshal([]byte(cachedData), &stats); err == nil {
			slog.InfoContext(ctx, "Statistics retrieved from cache")
			return stats, nil
		}
		slog.WarnContext(ctx, "Failed to unmarshal cached statistics, falling back to database", logger.Err(err))
	}

	// Получаем данные из базы данных
	stats, err := s.repo.GetStatistics(ctx)
	if err != nil {
		return domain.StatisticsResponse{}, err
	}

	// Устанавливаем общий заработок
	stats.TotalEarnings = s.totalEarnings

	// Получаем месячные продажи из orders-service
	monthlySales, err := s.getMonthlySalesFromOrdersService(ctx)
	if err != nil {
		slog.WarnContext(ctx, "Failed to get monthly sales from orders service, using empty list", logger.Err(err))
		stats.MonthlySales = []domain.MonthlySales{}
	} else {
		stats.MonthlySales = monthlySales
	}

	// Рассчитываем динамику продаж по месяцам
	if len(stats.MonthlySales) >= 2 {
		sorted := make([]domain.MonthlySales, len(stats.MonthlySales))
		copy(sorted, stats.MonthlySales)
		sort.Slice(sorted, func(i, j int) bool {
			return sorted[i].Month < sorted[j].Month
		})
		last := sorted[len(sorted)-1].Sales
		prev := sorted[len(sorted)-2].Sales
		if prev > 0 {
			stats.EarningsGrowth = math.Round(((last-prev)/prev)*1000) / 10
		} else if last > 0 {
			stats.EarningsGrowth = 100.0
		}
	}

	// Если индивидуальные динамики по складу не определены (равны 0), используем общую динамику продаж
	if stats.EarningsGrowth != 0 {
		if stats.PartsGrowth == 0 {
			stats.PartsGrowth = stats.EarningsGrowth
		}
		if stats.QuantityGrowth == 0 {
			stats.QuantityGrowth = stats.EarningsGrowth
		}
		if stats.ValueGrowth == 0 {
			stats.ValueGrowth = stats.EarningsGrowth
		}
	}

	// Кэшируем результат
	if data, err := json.Marshal(stats); err == nil {
		if err := s.redis.Set(ctx, statisticsCacheKey, data, statisticsCacheTTL).Err(); err != nil {
			slog.WarnContext(ctx, "Failed to cache statistics", logger.Err(err))
		} else {
			slog.InfoContext(ctx, "Statistics cached successfully")
		}
	} else {
		slog.WarnContext(ctx, "Failed to marshal statistics for caching", logger.Err(err))
	}

	return stats, nil
}

// SetMonthlySalesProvider задает поставщика месячных продаж
func (s *inventoryService) SetMonthlySalesProvider(fn func(ctx context.Context) ([]domain.MonthlySales, error)) {
	s.monthlySalesProvider = fn
}

// getMonthlySalesFromOrdersService получает месячные продажи через зарегистрированного провайдера
func (s *inventoryService) getMonthlySalesFromOrdersService(ctx context.Context) ([]domain.MonthlySales, error) {
	if s.monthlySalesProvider != nil {
		sales, err := s.monthlySalesProvider(ctx)
		if err == nil {
			return sales, nil
		}
		slog.WarnContext(ctx, "Failed to get monthly sales via provider, returning empty list", logger.Err(err))
	}
	return []domain.MonthlySales{}, nil
}

// invalidateStatisticsCache инвалидирует кэш статистики
func (s *inventoryService) invalidateStatisticsCache(ctx context.Context) {
	if err := s.redis.Del(ctx, statisticsCacheKey).Err(); err != nil {
		slog.WarnContext(ctx, "Failed to invalidate statistics cache", logger.Err(err))
	} else {
		slog.InfoContext(ctx, "Statistics cache invalidated")
	}
}

// UpdateEarnings обновляет общий заработок
func (s *inventoryService) UpdateEarnings(ctx context.Context, amount float64) error {
	s.totalEarnings += amount

	// Сохраняем в базу данных для персистентности
	if err := s.repo.UpdateTotalEarnings(ctx, s.totalEarnings); err != nil {
		slog.ErrorContext(ctx, "Failed to save total earnings to database", logger.Err(err))
		return err
	}

	s.invalidateStatisticsCache(ctx)
	slog.InfoContext(ctx, "Updated total earnings in database", slog.Float64("totalEarnings", s.totalEarnings))
	return nil
}
