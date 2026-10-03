package repository

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"auth-service/db/sqlc"
	"auth-service/internal/domain"
)

// ActivityLogRepository определяет контракт доступа к логам активности
type ActivityLogRepository interface {
	Create(log *domain.UserActivityLog) (*domain.UserActivityLog, error)
	FindWithFilters(filters domain.ActivityLogFilters) ([]domain.UserActivityLog, error)
}

type activityLogRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewActivityLogRepository создает новый репозиторий логов активности
func NewActivityLogRepository(pool *pgxpool.Pool) ActivityLogRepository {
	return &activityLogRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *activityLogRepository) Create(log *domain.UserActivityLog) (*domain.UserActivityLog, error) {
	var resourceID pgtype.Int8
	if log.ResourceID != nil {
		resourceID.Int64 = *log.ResourceID
		resourceID.Valid = true
	}

	var details pgtype.Text
	if log.Details != "" {
		details.String = log.Details
		details.Valid = true
	}

	var ipAddress pgtype.Text
	if log.IPAddress != "" {
		ipAddress.String = log.IPAddress
		ipAddress.Valid = true
	}

	var userAgent pgtype.Text
	if log.UserAgent != "" {
		userAgent.String = log.UserAgent
		userAgent.Valid = true
	}

	result, err := r.queries.CreateActivityLog(context.Background(), sqlc.CreateActivityLogParams{
		UserID:       log.UserID,
		UserName:     log.UserName,
		UserEmail:    log.UserEmail,
		Action:       log.Action,
		ResourceType: log.ResourceType,
		ResourceID:   resourceID,
		Details:      details,
		IpAddress:    ipAddress,
		UserAgent:    userAgent,
	})
	if err != nil {
		return nil, err
	}

	log.ID = result.ID
	if result.CreatedAt.Valid {
		log.CreatedAt = result.CreatedAt.Time
	}
	if result.ResourceID.Valid {
		log.ResourceID = &result.ResourceID.Int64
	}
	if result.Details.Valid {
		log.Details = result.Details.String
	}
	if result.IpAddress.Valid {
		log.IPAddress = result.IpAddress.String
	}
	if result.UserAgent.Valid {
		log.UserAgent = result.UserAgent.String
	}

	return log, nil
}

func (r *activityLogRepository) FindWithFilters(filters domain.ActivityLogFilters) ([]domain.UserActivityLog, error) {
	if filters.Limit == 0 {
		filters.Limit = 100
	}

	psql := sq.StatementBuilder.PlaceholderFormat(sq.Dollar)

	builder := psql.Select(
		"id", "user_id", "user_name", "user_email", "action",
		"resource_type", "resource_id", "details", "ip_address", "user_agent", "created_at",
	).From("user_activity_logs")

	if filters.UserID != nil {
		builder = builder.Where(sq.Eq{"user_id": *filters.UserID})
	}
	if filters.Action != "" {
		builder = builder.Where(sq.Eq{"action": filters.Action})
	}
	if filters.ResourceType != "" {
		builder = builder.Where(sq.Eq{"resource_type": filters.ResourceType})
	}
	if filters.UsefulOnly {
		builder = builder.Where(sq.NotEq{"action": domain.NonUsefulActions})
	}
	if filters.StartDate != nil {
		builder = builder.Where(sq.GtOrEq{"created_at": *filters.StartDate})
	}
	if filters.EndDate != nil {
		builder = builder.Where(sq.LtOrEq{"created_at": *filters.EndDate})
	}

	query, args, err := builder.
		OrderBy("created_at DESC").
		Limit(uint64(filters.Limit)).
		Offset(uint64(filters.Offset)).
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.pool.Query(context.Background(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []domain.UserActivityLog
	for rows.Next() {
		var l domain.UserActivityLog
		var resourceID pgtype.Int8
		var details, ipAddress, userAgent pgtype.Text
		var createdAt pgtype.Timestamptz

		err := rows.Scan(
			&l.ID, &l.UserID, &l.UserName, &l.UserEmail,
			&l.Action, &l.ResourceType, &resourceID,
			&details, &ipAddress, &userAgent, &createdAt,
		)
		if err != nil {
			return nil, err
		}

		if resourceID.Valid {
			l.ResourceID = &resourceID.Int64
		}
		if details.Valid {
			l.Details = details.String
		}
		if ipAddress.Valid {
			l.IPAddress = ipAddress.String
		}
		if userAgent.Valid {
			l.UserAgent = userAgent.String
		}
		if createdAt.Valid {
			l.CreatedAt = createdAt.Time
		}

		logs = append(logs, l)
	}

	if logs == nil {
		logs = []domain.UserActivityLog{}
	}

	return logs, rows.Err()
}
