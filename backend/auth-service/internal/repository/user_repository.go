package repository

import (
	"context"
	"database/sql"
	"strings"

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
	CountAll() (int, error)
	FindByEmail(email string) (*domain.User, error)
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

func (r *userRepository) FindByID(id int64) (*domain.User, error) {
	result, err := r.queries.GetUserByID(context.Background(), id)
	if err != nil {
		return nil, err
	}
	return sqlcUserToDomain(result), nil
}

func (r *userRepository) FindByEmailOrName(identifier string) (*domain.User, error) {
	trimmed := strings.TrimSpace(identifier)
	result, err := r.queries.GetUserByEmailOrName(context.Background(), sqlc.GetUserByEmailOrNameParams{
		Email: trimmed,
		Name:  trimmed,
	})
	if err != nil {
		return nil, err
	}
	return sqlcUserToDomain(result), nil
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	trimmed := strings.TrimSpace(email)
	result, err := r.queries.GetUserByEmailOrName(context.Background(), sqlc.GetUserByEmailOrNameParams{
		Email: trimmed,
		Name:  "",
	})
	if err != nil {
		return nil, err
	}
	if !strings.EqualFold(strings.TrimSpace(result.Email), trimmed) {
		return nil, sql.ErrNoRows
	}
	return sqlcUserToDomain(result), nil
}

func (r *userRepository) Update(id int64, params domain.UpdateUserParams) (*domain.User, error) {
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

func (r *userRepository) FindAll() ([]domain.User, error) {
	results, err := r.queries.FindAllUsers(context.Background())
	if err != nil {
		return nil, err
	}
	users := make([]domain.User, len(results))
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

func sqlcUserToDomain(u sqlc.User) *domain.User {
	return &domain.User{
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
