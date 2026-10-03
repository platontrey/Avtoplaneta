package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/sirupsen/logrus"

	"auth-service/internal/domain"
	"auth-service/internal/events"
	"auth-service/internal/repository"
	"auth-service/internal/security"
)

// AuthService определяет интерфейс бизнес-логики аутентификации и пользователей
type AuthService interface {
	AuthenticateUser(email, password string) (*domain.User, error)
	CreateUserFromGoogle(email, name string) (*domain.User, error)
	GetCurrentUser(userID int64) (*domain.User, error)
	Logout(userID int64) error

	CreateUser(req domain.CreateUserRequest) (*domain.User, error)
	GetUsers() ([]domain.User, error)
	UpdateUser(userID int64, req domain.UpdateUserRequest) (*domain.User, error)
	DeleteUser(userID int64) error
	GetTotalUsersCount() (int, error)

	LogUserActivity(user *domain.User, action, resourceType string, resourceID *int64, details, ip, userAgent string) error
	GetUserActivityLogs(filters domain.ActivityLogFilters) ([]domain.UserActivityLog, error)

	GenerateCSRFToken(userID *int64) (string, error)
}

type authService struct {
	userRepo       repository.UserRepository
	activityRepo   repository.ActivityLogRepository
	sessionStore   security.SessionStore
	rateLimiter    security.RateLimiter
	csrfManager    security.CSRFManager
	eventPublisher events.UserEventPublisher
}

// NewAuthService создает новый экземпляр AuthService с dependency injection
func NewAuthService(
	userRepo repository.UserRepository,
	activityRepo repository.ActivityLogRepository,
	sessionStore security.SessionStore,
	rateLimiter security.RateLimiter,
	csrfManager security.CSRFManager,
	publisher ...events.UserEventPublisher,
) AuthService {
	var pub events.UserEventPublisher
	if len(publisher) > 0 {
		pub = publisher[0]
	}
	return &authService{
		userRepo:       userRepo,
		activityRepo:   activityRepo,
		sessionStore:   sessionStore,
		rateLimiter:    rateLimiter,
		csrfManager:    csrfManager,
		eventPublisher: pub,
	}
}

func (s *authService) AuthenticateUser(email, password string) (*domain.User, error) {
	trimmedIdentifier := strings.TrimSpace(email)
	if trimmedIdentifier == "" {
		return nil, fmt.Errorf("логин не может быть пустым")
	}

	if s.rateLimiter != nil {
		if limited, _ := s.rateLimiter.IsLimited(strings.ToLower(trimmedIdentifier)+":user", 10, time.Minute); limited {
			logrus.WithFields(logrus.Fields{
				"identifier": trimmedIdentifier,
			}).Warn("Rate limit exceeded for user login")
			return nil, fmt.Errorf("слишком много попыток входа")
		}
	}

	user, err := s.userRepo.FindByEmailOrName(trimmedIdentifier)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"identifier": trimmedIdentifier,
		}).Warn("User not found during login")
		time.Sleep(time.Second)
		return nil, fmt.Errorf("неверные учетные данные")
	}

	if user.Provider != "local" || user.Password == "" {
		logrus.WithFields(logrus.Fields{
			"identifier": trimmedIdentifier,
		}).Warn("Attempt to login with non-local user")
		time.Sleep(time.Second)
		return nil, fmt.Errorf("неверные учетные данные")
	}

	if err := security.CheckPassword(user.Password, password); err != nil {
		logrus.WithFields(logrus.Fields{
			"identifier": trimmedIdentifier,
		}).Warn("Invalid password during login")
		time.Sleep(time.Second)
		return nil, fmt.Errorf("неверные учетные данные")
	}

	logrus.WithFields(logrus.Fields{
		"email": user.Email,
		"name":  user.Name,
	}).Info("User authenticated successfully")

	return user, nil
}

func (s *authService) CreateUserFromGoogle(email, name string) (*domain.User, error) {
	existing, err := s.userRepo.FindByEmailOrName(email)
	if err == nil {
		logrus.WithField("email", email).Info("Existing Google user found")
		return existing, nil
	}

	user := &domain.User{
		Email:    email,
		Name:     name,
		Provider: "google",
		Role:     "operator",
	}

	if _, err := s.userRepo.Create(user); err != nil {
		logrus.WithError(err).WithField("email", email).Error("Failed to create Google user")
		return nil, err
	}

	logrus.WithField("email", email).Info("New Google user created")
	return user, nil
}

func (s *authService) GetCurrentUser(userID int64) (*domain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get current user")
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *authService) Logout(userID int64) error {
	if s.csrfManager != nil {
		s.csrfManager.Invalidate(userID)
	}

	logrus.WithField("user_id", userID).Info("User logged out")
	return nil
}

func (s *authService) CreateUser(req domain.CreateUserRequest) (*domain.User, error) {
	exists, err := s.userRepo.ExistsByEmailOrName(req.Email, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("пользователь с таким email или именем уже существует")
	}

	hashedPassword, err := security.HashPassword(req.Password)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		return nil, fmt.Errorf("не удалось создать пользователя")
	}

	user := &domain.User{
		Email:    req.Email,
		Name:     req.Name,
		Initials: req.Initials,
		INN:      req.INN,
		Provider: "local",
		Role:     req.Role,
		Password: hashedPassword,
	}

	if _, err := s.userRepo.Create(user); err != nil {
		logrus.WithError(err).WithField("email", req.Email).Error("Failed to create user")
		return nil, err
	}

	user.Password = ""

	logrus.WithFields(logrus.Fields{
		"email": req.Email,
		"role":  req.Role,
	}).Info("User created successfully")

	return user, nil
}

func (s *authService) GetUsers() ([]domain.User, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		logrus.WithError(err).Error("Failed to get users")
		return nil, err
	}

	for i := range users {
		users[i].Password = ""
	}

	return users, nil
}

func (s *authService) UpdateUser(userID int64, req domain.UpdateUserRequest) (*domain.User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден")
	}

	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmailOrName(req.Email, "")
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("пользователь с таким email уже существует")
		}
	}

	updateParams := domain.UpdateUserParams{
		Name:     req.Name,
		Email:    req.Email,
		Initials: req.Initials,
		INN:      req.INN,
		Role:     req.Role,
	}

	if updateParams.Name == "" && updateParams.Email == "" &&
		updateParams.Initials == "" && updateParams.INN == "" &&
		updateParams.Role == "" {
		return nil, fmt.Errorf("нет полей для обновления")
	}

	oldName := user.Name
	updatedUser, err := s.userRepo.Update(userID, updateParams)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to update user")
		return nil, err
	}

	updatedUser.Password = ""

	if s.eventPublisher != nil && updatedUser.Name != "" && updatedUser.Name != oldName {
		if err := s.eventPublisher.PublishUserRenamed(context.Background(), userID, updatedUser.Name); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Warn("Failed to publish seller_renamed event")
		}
	}

	logrus.WithField("user_id", userID).Info("User updated successfully")
	return updatedUser, nil
}

func (s *authService) DeleteUser(userID int64) error {
	_, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("пользователь не найден")
	}

	if err := s.userRepo.Delete(userID); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to delete user")
		return err
	}

	logrus.WithField("user_id", userID).Info("User deleted successfully")
	return nil
}

func (s *authService) GetTotalUsersCount() (int, error) {
	count, err := s.userRepo.CountAll()
	if err != nil {
		logrus.WithError(err).Error("Failed to get users count")
		return 0, err
	}
	return count, nil
}

func (s *authService) LogUserActivity(user *domain.User, action, resourceType string, resourceID *int64, details, ip, userAgent string) error {
	log := &domain.UserActivityLog{
		UserID:       user.ID,
		UserName:     user.Name,
		UserEmail:    user.Email,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ip,
		UserAgent:    userAgent,
	}

	if _, err := s.activityRepo.Create(log); err != nil {
		logrus.WithError(err).WithFields(logrus.Fields{
			"user_id": user.ID,
			"action":  action,
		}).Error("Failed to log user activity")
		return err
	}

	return nil
}

func (s *authService) GetUserActivityLogs(filters domain.ActivityLogFilters) ([]domain.UserActivityLog, error) {
	logs, err := s.activityRepo.FindWithFilters(filters)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user activity logs")
		return nil, err
	}
	return logs, nil
}

func (s *authService) GenerateCSRFToken(userID *int64) (string, error) {
	if s.csrfManager != nil {
		return s.csrfManager.GenerateToken(userID)
	}
	return "", nil
}
