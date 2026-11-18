package repository

import (
	"math"
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"avtoplaneta/pkg/logger"
	"parts-service/db/sqlc"
	"parts-service/internal/domain"
)

// ─── Specific operations ────────────────────────────────────────────────────

func (r *partRepository) MarkForDeletion(ctx context.Context, id int64, deleteAt time.Time) error {
	return r.queries.MarkForDeletion(ctx, sqlc.MarkForDeletionParams{
		ID:         id,
		ToDeleteAt: pgtype.Timestamptz{Time: deleteAt, Valid: true},
	})
}

func (r *partRepository) DeleteExpiredParts(ctx context.Context, before time.Time) error {
	_, err := r.queries.DeleteExpiredParts(ctx, pgtype.Timestamptz{Time: before, Valid: true})
	return err
}

func calculateGrowth(current, previous float64) float64 {
	if previous <= 0 {
		return 0
	}
	growth := ((current - previous) / previous) * 100.0
	return math.Round(growth*10) / 10
}

func (r *partRepository) GetStatistics(ctx context.Context) (domain.StatisticsResponse, error) {
	var stats domain.StatisticsResponse

	totals, err := r.queries.GetStatsTotals(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get totals", logger.Err(err))
		return stats, err
	}
	stats.TotalParts = int(totals.TotalParts)
	stats.TotalQuantity = int(totals.TotalQuantity)
	stats.TotalValue = totals.TotalValue

	now := time.Now()
	// Срез на 30 дней назад для стабильного расчёта динамики роста склада
	cutoff := now.AddDate(0, 0, -30)
	prev, err := r.queries.GetStatsTotalsBefore(ctx, pgtype.Timestamptz{Time: cutoff, Valid: true})

	if err == nil && prev.PrevParts > 0 {
		stats.PartsGrowth = calculateGrowth(float64(totals.TotalParts), float64(prev.PrevParts))
		stats.QuantityGrowth = calculateGrowth(float64(totals.TotalQuantity), float64(prev.PrevQuantity))
		stats.ValueGrowth = calculateGrowth(totals.TotalValue, prev.PrevValue)
	}

	categories, err := r.queries.GetStatsCategories(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Failed to get categories", logger.Err(err))
		return stats, err
	}

	catList := make([]domain.CategoryCount, len(categories))
	for i, c := range categories {
		catList[i] = domain.CategoryCount{
			Name:  c.Name,
			Count: int(c.Count),
		}
	}
	stats.Categories = catList

	slog.InfoContext(ctx, "Statistics retrieved successfully",
		slog.Int("total_parts", stats.TotalParts),
		slog.Float64("total_value", stats.TotalValue),
		slog.Float64("parts_growth", stats.PartsGrowth),
		slog.Float64("quantity_growth", stats.QuantityGrowth),
		slog.Float64("value_growth", stats.ValueGrowth),
		slog.Int("categories_count", len(stats.Categories)))

	return stats, nil
}



// ─── Earnings ───────────────────────────────────────────────────────────────

func (r *partRepository) GetTotalEarnings(ctx context.Context) (float64, error) {
	earning, err := r.queries.GetEarnings(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			created, err := r.queries.CreateEarnings(ctx, 0)
			if err != nil {
				return 0, err
			}
			return created.TotalAmount, nil
		}
		return 0, err
	}
	return earning.TotalAmount, nil
}

func (r *partRepository) UpdateTotalEarnings(ctx context.Context, amount float64) error {
	earning, err := r.queries.GetEarnings(ctx)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			_, err = r.queries.CreateEarnings(ctx, amount)
			return err
		}
		return err
	}
	return r.queries.UpdateEarnings(ctx, sqlc.UpdateEarningsParams{
		TotalAmount: amount,
		ID:          earning.ID,
	})
}
