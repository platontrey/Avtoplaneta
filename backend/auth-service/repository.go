package main

import (
	"context"
	"database/sql"
	"time"

	sq "github.com/Masterminds/squirrel"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"auth-service/db/sqlc"
)

type UpdateUserParams struct {
	Name     string
	Email    string
	Initials string
	INN      string
	Role     string
}

type UserRepository interface {
	Create(user *User) (*User, error)
	FindByID(id int64) (*User, error)
	FindByEmailOrName(identifier string) (*User, error)
	Update(id int64, params UpdateUserParams) (*User, error)
	Delete(id int64) error
	FindAll() ([]User, error)
	ExistsByEmailOrName(email, name string) (bool, error)
	CountAll() (int, error)
	FindByEmail(email string) (*User, error)
}

type userRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *userRepository) Create(user *User) (*User, error) {
	result, err := r.queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Email:    user.Email,
		Name:     user.Name,
		Initials: user.Initials,
		Inn:      user.INN,
		Provider: user.Provider,
		Role:     user.Role,
		Password: user.Password,
	})
	if err != nil {
		return nil, err
	}
	user.ID = result.ID
	return user, nil
}

func (r *userRepository) FindByID(id int64) (*User, error) {
	result, err := r.queries.GetUserByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return sqlcUserToDomain(result), nil
}

func (r *userRepository) FindByEmailOrName(identifier string) (*User, error) {
	result, err := r.queries.GetUserByEmailOrName(context.Background(), sqlc.GetUserByEmailOrNameParams{
		Email: identifier,
		Name:  identifier,
	})
	if err != nil {
		return nil, err
	}
	return sqlcUserToDomain(result), nil
}

func (r *userRepository) FindByEmail(email string) (*User, error) {
	result, err := r.queries.GetUserByEmailOrName(context.Background(), sqlc.GetUserByEmailOrNameParams{
		Email: email,
		Name:  "",
	})
	if err != nil {
		return nil, err
	}
	if result.Email != email {
		return nil, sql.ErrNoRows
	}
	return sqlcUserToDomain(result), nil
}

func (r *userRepository) Update(id int64, params UpdateUserParams) (*User, error) {
	err := r.queries.UpdateUser(context.Background(), sqlc.UpdateUserParams{
		ID:      id,
		Column2: params.Name,
		Column3: params.Email,
		Column4: params.Initials,
		Column5: params.INN,
		Column6: params.Role,
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *userRepository) Delete(id int64) error {
	return r.queries.DeleteUser(context.Background(), id)
}

func (r *userRepository) FindAll() ([]User, error) {
	results, err := r.queries.FindAllUsers(context.Background())
	if err != nil {
		return nil, err
	}
	users := make([]User, len(results))
	for i, u := range results {
		users[i] = *sqlcUserToDomain(u)
	}
	return users, nil
}

func (r *userRepository) ExistsByEmailOrName(email, name string) (bool, error) {
	if email != "" {
		exists, err := r.queries.ExistsByEmail(context.Background(), email)
		if err != nil {
			return false, err
		}
		if exists {
			return true, nil
		}
	}
	if name != "" {
		return r.queries.ExistsByName(context.Background(), name)
	}
	return false, nil
}

func (r *userRepository) CountAll() (int, error) {
	count, err := r.queries.CountAllUsers(context.Background())
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func sqlcUserToDomain(u sqlc.User) *User {
	return &User{
		ID:       u.ID,
		Email:    u.Email,
		Name:     u.Name,
		Initials: u.Initials,
		INN:      u.Inn,
		Provider: u.Provider,
		Role:     u.Role,
		Password: u.Password,
	}
}

// ─── Activity Logs ──────────────────────────────────────────────────────────

type ActivityLogFilters struct {
	UserID       *int64
	Action       string
	ResourceType string
	StartDate    *time.Time
	EndDate      *time.Time
	// UsefulOnly скрывает навигационный шум (просмотры, доступ к админке,
	// просмотр логов), оставляя только полезные действия — мутации и вход/выход.
	UsefulOnly bool
	Limit      int
	Offset     int
}

// nonUsefulActions — действия, генерирующие навигационный шум.
// Скрыты, когда фильтр "полезные данные" включён.
var nonUsefulActions = []string{
	"access_admin_panel",
	"view_activity_logs",
	"view_users",
	"view_part",
	"search_parts",
}

type ActivityLogRepository interface {
	Create(log *UserActivityLog) (*UserActivityLog, error)
	FindWithFilters(filters ActivityLogFilters) ([]UserActivityLog, error)
}

type activityLogRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

func NewActivityLogRepository(pool *pgxpool.Pool) ActivityLogRepository {
	return &activityLogRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *activityLogRepository) Create(log *UserActivityLog) (*UserActivityLog, error) {
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

func (r *activityLogRepository) FindWithFilters(filters ActivityLogFilters) ([]UserActivityLog, error) {
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
		builder = builder.Where(sq.NotEq{"action": nonUsefulActions})
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

	var logs []UserActivityLog
	for rows.Next() {
		var l UserActivityLog
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
		logs = []UserActivityLog{}
	}

	return logs, rows.Err()
}
