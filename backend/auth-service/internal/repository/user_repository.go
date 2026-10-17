package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"strings"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"auth-service/db/sqlc"
	"auth-service/internal/domain"
)

// UserRepository определяет контракт доступа к данным пользователей
type UserRepository interface {
	Create(user *domain.User) (*domain.User, error)
	FindByID(id int64) (*domain.User, error)
	FindByEmailOrName(identifier string) (*domain.User, error)
	Update(id int64, params domain.UpdateUserParams) (*domain.User, error)
	Delete(id int64) error
	FindAll() ([]domain.User, error)
	ExistsByEmailOrName(email, name string) (bool, error)
	ExistsByUsername(username string) (bool, error)
	CountAll() (int, error)
	FindByEmail(email string) (*domain.User, error)
	UpdateProfile(id int64, name, username, inn string) (*domain.User, error)
	UpdatePassword(id int64, passwordHash string) error
	UpdateAvatar(id int64, avatarURL string) (*domain.User, error)

	// Outbox
	CreateOutboxEvent(ctx context.Context, aggregateType string, aggregateID int64, eventType string, payload []byte) error
	FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]sqlc.OutboxEvent, error)
	MarkOutboxEventProcessed(ctx context.Context, id int64) error
	MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error
	CleanupProcessedOutboxEvents(ctx context.Context) error
}

type userRepository struct {
	pool    *pgxpool.Pool
	queries *sqlc.Queries
}

// NewUserRepository создает новый репозиторий пользователей
func NewUserRepository(pool *pgxpool.Pool) UserRepository {
	return &userRepository{
		pool:    pool,
		queries: sqlc.New(pool),
	}
}

func (r *userRepository) Create(user *domain.User) (*domain.User, error) {
	result, err := r.queries.CreateUser(context.Background(), sqlc.CreateUserParams{
		Username:  strings.TrimSpace(user.Username),
		Email:     strings.TrimSpace(user.Email),
		Name:      user.Name,
		Initials:  user.Initials,
		Inn:       user.INN,
		Provider:  user.Provider,
		Role:      user.Role,
		Password:  user.Password,
		AvatarUrl: pgtype.Text{String: user.AvatarURL, Valid: user.AvatarURL != ""},
	})
	if err != nil {
		return nil, err
	}
	user.ID = result.ID
	user.Username = result.Username
	user.AvatarURL = result.AvatarUrl
	return user, nil
}

func (r *userRepository) FindByID(id int64) (*domain.User, error) {
	result, err := r.queries.GetUserByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	var emailStr string
	if result.Email.Valid {
		emailStr = result.Email.String
	}
	return &domain.User{
		ID:        result.ID,
		Username:  result.Username,
		Email:     emailStr,
		Name:      result.Name,
		Initials:  result.Initials,
		INN:       result.Inn,
		Provider:  result.Provider,
		Role:      result.Role,
		Password:  result.Password,
		AvatarURL: result.AvatarUrl,
	}, nil
}

func (r *userRepository) FindByEmailOrName(identifier string) (*domain.User, error) {
	trimmed := strings.TrimSpace(identifier)
	result, err := r.queries.GetUserByLoginOrEmail(context.Background(), trimmed)
	if err != nil {
		return nil, err
	}
	var emailStr string
	if result.Email.Valid {
		emailStr = result.Email.String
	}
	return &domain.User{
		ID:        result.ID,
		Username:  result.Username,
		Email:     emailStr,
		Name:      result.Name,
		Initials:  result.Initials,
		INN:       result.Inn,
		Provider:  result.Provider,
		Role:      result.Role,
		Password:  result.Password,
		AvatarURL: result.AvatarUrl,
	}, nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	trimmed := strings.TrimSpace(email)
	result, err := r.queries.GetUserByLoginOrEmail(context.Background(), trimmed)
	if err != nil {
		return nil, err
	}
	var emailStr string
	if result.Email.Valid {
		emailStr = result.Email.String
	}
	if !strings.EqualFold(strings.TrimSpace(emailStr), trimmed) {
		return nil, sql.ErrNoRows
	}
	return &domain.User{
		ID:        result.ID,
		Username:  result.Username,
		Email:     emailStr,
		Name:      result.Name,
		Initials:  result.Initials,
		INN:       result.Inn,
		Provider:  result.Provider,
		Role:      result.Role,
		Password:  result.Password,
		AvatarURL: result.AvatarUrl,
	}, nil
}

func (r *userRepository) Update(id int64, params domain.UpdateUserParams) (*domain.User, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var avatarText pgtype.Text
	if params.AvatarURL != nil {
		avatarText = pgtype.Text{String: *params.AvatarURL, Valid: true}
	}

	qtx := r.queries.WithTx(tx)
	err = qtx.UpdateUser(ctx, sqlc.UpdateUserParams{
		ID:        id,
		Username:  params.Username,
		Name:      params.Name,
		Email:     params.Email,
		Initials:  params.Initials,
		Inn:       params.INN,
		Role:      params.Role,
		AvatarUrl: avatarText,
	})
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(params.Name) != "" {
		payload, _ := json.Marshal(map[string]interface{}{
			"seller_id": id,
			"name":      strings.TrimSpace(params.Name),
		})
		_, err = qtx.CreateAuthOutboxEvent(ctx, sqlc.CreateAuthOutboxEventParams{
			AggregateType: "user",
			AggregateID:   id,
			EventType:     "seller_renamed",
			Payload:       payload,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.FindByID(id)
}

func (r *userRepository) UpdateProfile(id int64, name, username, inn string) (*domain.User, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	qtx := r.queries.WithTx(tx)
	err = qtx.UpdateUserProfile(ctx, sqlc.UpdateUserProfileParams{
		ID:       id,
		Name:     strings.TrimSpace(name),
		Username: strings.TrimSpace(username),
		Inn:      strings.TrimSpace(inn),
	})
	if err != nil {
		return nil, err
	}

	if strings.TrimSpace(name) != "" {
		payload, _ := json.Marshal(map[string]interface{}{
			"seller_id": id,
			"name":      strings.TrimSpace(name),
		})
		_, err = qtx.CreateAuthOutboxEvent(ctx, sqlc.CreateAuthOutboxEventParams{
			AggregateType: "user",
			AggregateID:   id,
			EventType:     "seller_renamed",
			Payload:       payload,
		})
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.FindByID(id)
}

func (r *userRepository) UpdatePassword(id int64, passwordHash string) error {
	return r.queries.UpdateUserPassword(context.Background(), sqlc.UpdateUserPasswordParams{
		ID:       id,
		Password: passwordHash,
	})
}

func (r *userRepository) UpdateAvatar(id int64, avatarURL string) (*domain.User, error) {
	err := r.queries.UpdateUserAvatar(context.Background(), sqlc.UpdateUserAvatarParams{
		ID:        id,
		AvatarUrl: avatarURL,
	})
	if err != nil {
		return nil, err
	}
	return r.FindByID(id)
}

func (r *userRepository) Delete(id int64) error {
	return r.queries.DeleteUser(context.Background(), id)
}

func (r *userRepository) FindAll() ([]domain.User, error) {
	results, err := r.queries.FindAllUsers(context.Background())
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, len(results))
	for i, u := range results {
		var emailStr string
		if u.Email.Valid {
			emailStr = u.Email.String
		}
		users[i] = domain.User{
			ID:        u.ID,
			Username:  u.Username,
			Email:     emailStr,
			Name:      u.Name,
			Initials:  u.Initials,
			INN:       u.Inn,
			Provider:  u.Provider,
			Role:      u.Role,
			Password:  u.Password,
			AvatarURL: u.AvatarUrl,
		}
	}
	return users, nil
}

func (r *userRepository) ExistsByUsername(username string) (bool, error) {
	if strings.TrimSpace(username) != "" {
		return r.queries.ExistsByUsername(context.Background(), strings.TrimSpace(username))
	}
	return false, nil
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

func sqlcUserToDomain(u sqlc.User) *domain.User {
	var emailStr string
	if u.Email.Valid {
		emailStr = u.Email.String
	}
	return &domain.User{
		ID:       u.ID,
		Username: u.Username,
		Email:    emailStr,
		Name:     u.Name,
		Initials: u.Initials,
		INN:      u.Inn,
		Provider: u.Provider,
		Role:     u.Role,
		Password: u.Password,
	}
}

func (r *userRepository) CreateOutboxEvent(ctx context.Context, aggregateType string, aggregateID int64, eventType string, payload []byte) error {
	_, err := r.queries.CreateAuthOutboxEvent(ctx, sqlc.CreateAuthOutboxEventParams{
		AggregateType: aggregateType,
		AggregateID:   aggregateID,
		EventType:     eventType,
		Payload:       payload,
	})
	return err
}

func (r *userRepository) FetchPendingOutboxEvents(ctx context.Context, limit int32) ([]sqlc.OutboxEvent, error) {
	return r.queries.FetchPendingAuthOutboxEvents(ctx, limit)
}

func (r *userRepository) MarkOutboxEventProcessed(ctx context.Context, id int64) error {
	return r.queries.MarkAuthOutboxEventProcessed(ctx, id)
}

func (r *userRepository) MarkOutboxEventFailed(ctx context.Context, id int64, lastErr string) error {
	return r.queries.MarkAuthOutboxEventFailed(ctx, sqlc.MarkAuthOutboxEventFailedParams{
		ID:        id,
		LastError: pgtype.Text{String: lastErr, Valid: lastErr != ""},
	})
}

func (r *userRepository) CleanupProcessedOutboxEvents(ctx context.Context) error {
	return r.queries.CleanupProcessedAuthOutboxEvents(ctx)
}
