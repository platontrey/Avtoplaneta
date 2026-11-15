package repository

import (
	"parts-service/db/sqlc"
	"fmt"
	"context"
	"log/slog"
	"strconv"
	"strings"

	"avtoplaneta/pkg/logger"
	"parts-service/internal/domain"
)

// ─── Supplier operations ────────────────────────────────────────────────────

func (r *partRepository) DeleteZeroQuantityPartsWithIDs(ctx context.Context, supplierCode string) ([]int64, error) {
	supplierCode = strings.TrimSpace(supplierCode)
	slog.InfoContext(ctx, "Repository: DeleteZeroQuantityPartsWithIDs", slog.String("supplier_code", supplierCode))

	var (
		deletedIDs []int64
		err        error
	)

	if strings.HasPrefix(supplierCode, "car:") {
		raw := strings.TrimPrefix(supplierCode, "car:")
		parts := strings.SplitN(raw, "|", 4)
		for len(parts) < 4 {
			parts = append(parts, "")
		}
		deletedIDs, err = r.queries.DeleteZeroQuantityPartsByCar(ctx, sqlc.DeleteZeroQuantityPartsByCarParams{
			Brand:          parts[0],
			Model:          parts[1],
			CarReleaseDate: parts[2],
			Vin:            parts[3],
		})
	} else {
		deletedIDs, err = r.queries.DeleteZeroQuantityPartsBySupplierCode(ctx, supplierCode)
	}
	if err != nil {
		return nil, err
	}

	slog.InfoContext(ctx, "Repository: Successfully deleted zero-quantity parts", slog.Int("deleted", len(deletedIDs)))
	return deletedIDs, nil
}

func (r *partRepository) DeleteZeroQuantityPartsBySupplier(ctx context.Context, supplierCode string) (int64, error) {
	ids, err := r.DeleteZeroQuantityPartsWithIDs(ctx, supplierCode)
	if err != nil {
		return 0, err
	}
	return int64(len(ids)), nil
}

func (r *partRepository) GetSupplierBatches(ctx context.Context) ([]domain.SupplierBatchInfo, error) {
	rows, err := r.queries.GetSupplierBatches(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Repository: GetSupplierBatches query failed", logger.Err(err))
		return nil, err
	}

	var batches []domain.SupplierBatchInfo
	for _, row := range rows {
		var code string
		switch v := row.BatchCode.(type) {
		case string:
			code = v
		case []byte:
			code = string(v)
		default:
			code = fmt.Sprintf("%v", v)
		}

		b := domain.SupplierBatchInfo{
			Code:      code,
			Brand:     row.BrandName,
			Model:     row.ModelName,
			Year:      row.ReleaseDate,
			VIN:       row.CarVin,
			ZeroCount: row.ZeroCount,
		}

		carTitle := strings.TrimSpace(strings.Join([]string{b.Brand, b.Model, b.Year}, " "))
		if carTitle == "" {
			carTitle = "Без марки/модели"
		}
		if b.VIN != "" {
			carTitle += fmt.Sprintf(" (VIN: %s)", b.VIN)
		}
		if strings.HasPrefix(b.Code, "car:") {
			b.Label = fmt.Sprintf("%s — %d шт. с 0 [без кода поставки]", carTitle, b.ZeroCount)
		} else {
			b.Label = fmt.Sprintf("%s — %d шт. с 0 [код: %s]", carTitle, b.ZeroCount, b.Code)
		}
		batches = append(batches, b)
	}

	slog.InfoContext(ctx, "Repository: GetSupplierBatches loaded batches", slog.Int("count", len(batches)))
	return batches, nil
}

func (r *partRepository) GetSupplierCodes(ctx context.Context) ([]string, error) {
	batches, err := r.GetSupplierBatches(ctx)
	if err == nil {
		codes := make([]string, 0, len(batches))
		for _, b := range batches {
			codes = append(codes, b.Code)
		}
		slog.InfoContext(ctx, "Repository: GetSupplierCodes from batches", slog.Int("count", len(codes)))
		return codes, nil
	}

	slog.WarnContext(ctx, "Repository: GetSupplierBatches failed, falling back to GetSupplierCodes query", logger.Err(err))
	codes, err := r.queries.GetSupplierCodes(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "Repository: GetSupplierCodes query failed", logger.Err(err))
		return nil, err
	}
	slog.InfoContext(ctx, "Repository: GetSupplierCodes", slog.Int("count", len(codes)))
	return codes, nil
}

func (r *partRepository) GetNextSupplierCode(ctx context.Context) (string, error) {
	seqVal, err := r.queries.GetNextDefectBatchSequence(ctx)
	if err != nil {
		slog.WarnContext(ctx, "Repository: nextval(defect_batch_seq) failed, falling back to max supplier_code query", logger.Err(err))
		maxVal, fallbackErr := r.queries.GetMaxSupplierCodeFallback(ctx)
		if fallbackErr == nil && maxVal > 0 {
			return strconv.FormatInt(int64(maxVal), 10), nil
		}
		return "1", nil
	}
	return strconv.FormatInt(seqVal, 10), nil
}

func (r *partRepository) PeekNextSupplierCode(ctx context.Context) (string, error) {
	var lastVal int64
	var isCalled bool
	err := r.pool.QueryRow(ctx, "SELECT last_value, is_called FROM defect_batch_seq").Scan(&lastVal, &isCalled)
	if err != nil {
		maxVal, fallbackErr := r.queries.GetMaxSupplierCodeFallback(ctx)
		if fallbackErr == nil && maxVal > 0 {
			return strconv.FormatInt(int64(maxVal), 10), nil
		}
		return "1", nil
	}
	if !isCalled {
		return strconv.FormatInt(lastVal, 10), nil
	}
	return strconv.FormatInt(lastVal+1, 10), nil
}

func (r *partRepository) GetNextPartNumber(ctx context.Context) (string, error) {
	seqVal, err := r.queries.GetNextPartNumberSequence(ctx)
	if err != nil {
		slog.WarnContext(ctx, "Repository: nextval(part_number_seq) failed, falling back to max number query", logger.Err(err))
		maxVal, fallbackErr := r.queries.GetMaxPartNumberFallback(ctx)
		if fallbackErr == nil && maxVal > 0 {
			return strconv.FormatInt(int64(maxVal), 10), nil
		}
		return "1", nil
	}
	return strconv.FormatInt(seqVal, 10), nil
}
