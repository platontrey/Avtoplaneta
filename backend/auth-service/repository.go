package main

import (
	"time"

	"gorm.io/gorm"
)

// UserRepository определяет интерфейс для работы с пользователями в базе данных
type UserRepository interface {
	// Create Основные операции CRUD
	Create(user *User) error
	FindByID(id uint) (*User, error)
	FindByEmailOrName(identifier string) (*User, error)
	Update(id uint, updates map[string]interface{}) error
	Delete(id uint) error

	// FindAll Специфические операции
	FindAll() ([]User, error)
	ExistsByEmailOrName(email, name string) (bool, error)
	CountAll() (int, error)
}

// userRepository реализует UserRepository
type userRepository struct {
	db *gorm.DB
}

// NewUserRepository создает новый экземпляр репозитория пользователей
func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

// Create создает нового пользователя
func (r *userRepository) Create(user *User) error {
	return r.db.Create(user).Error
}

// FindByID находит пользователя по ID
func (r *userRepository) FindByID(id uint) (*User, error) {
	var user User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmailOrName находит пользователя по email или имени
func (r *userRepository) FindByEmailOrName(identifier string) (*User, error) {
	var user User
	err := r.db.Where("email = ? OR name = ?", identifier, identifier).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update обновляет пользователя по ID
func (r *userRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&User{}).Where("id = ?", id).Updates(updates).Error
}

// Delete удаляет пользователя по ID
func (r *userRepository) Delete(id uint) error {
	return r.db.Delete(&User{}, id).Error
}

// FindAll находит всех пользователей
func (r *userRepository) FindAll() ([]User, error) {
	var users []User
	err := r.db.Find(&users).Error
	return users, err
}

// ExistsByEmailOrName проверяет, существует ли пользователь с таким email или именем
func (r *userRepository) ExistsByEmailOrName(email, name string) (bool, error) {
	var count int64
	err := r.db.Model(&User{}).Where("email = ? OR name = ?", email, name).Count(&count).Error
	return count > 0, err
}

// CountAll возвращает общее количество пользователей
func (r *userRepository) CountAll() (int, error) {
	var count int64
	err := r.db.Model(&User{}).Count(&count).Error
	return int(count), err
}

// ActivityLogRepository определяет интерфейс для работы с логами активности
type ActivityLogRepository interface {
	Create(log *UserActivityLog) error
	FindWithFilters(filters ActivityLogFilters) ([]UserActivityLog, error)
}

// ActivityLogFilters фильтры для поиска логов активности
type ActivityLogFilters struct {
	UserID       *uint
	Action       string
	ResourceType string
	StartDate    *time.Time
	EndDate      *time.Time
	Limit        int
	Offset       int
}

// activityLogRepository реализует ActivityLogRepository
type activityLogRepository struct {
	db *gorm.DB
}

// NewActivityLogRepository создает новый репозиторий логов активности
func NewActivityLogRepository(db *gorm.DB) ActivityLogRepository {
	return &activityLogRepository{db: db}
}

// Create создает запись лога активности
func (r *activityLogRepository) Create(log *UserActivityLog) error {
	return r.db.Create(log).Error
}

// FindWithFilters находит логи активности с фильтрами
func (r *activityLogRepository) FindWithFilters(filters ActivityLogFilters) ([]UserActivityLog, error) {
	query := r.db.Model(&UserActivityLog{}).Order("created_at DESC")

	if filters.UserID != nil {
		query = query.Where("user_id = ?", *filters.UserID)
	}
	if filters.Action != "" {
		query = query.Where("action = ?", filters.Action)
	}
	if filters.ResourceType != "" {
		query = query.Where("resource_type = ?", filters.ResourceType)
	}
	if filters.StartDate != nil {
		query = query.Where("created_at >= ?", *filters.StartDate)
	}
	if filters.EndDate != nil {
		query = query.Where("created_at <= ?", *filters.EndDate)
	}

	var logs []UserActivityLog
	err := query.Limit(filters.Limit).Offset(filters.Offset).Find(&logs).Error
	return logs, err
}
