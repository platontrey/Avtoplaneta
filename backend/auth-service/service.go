package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

// SessionStore определяет интерфейс для работы с сессиями
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

// RateLimiter определяет интерфейс для ограничения скорости
type RateLimiter interface {
	IsLimited(key string, limit int, window time.Duration) (bool, error)
}

// AuthService определяет интерфейс для бизнес-логики аутентификации
type AuthService interface {
	// AuthenticateUser Аутентификация
	AuthenticateUser(email, password string) (*User, error)
	CreateUserFromGoogle(email, name string) (*User, error)
	GetCurrentUser(userID uint) (*User, error)
	Logout(userID uint) error

	// CreateUser Управление пользователями
	CreateUser(req CreateUserRequest) (*User, error)
	GetUsers() ([]User, error)
	UpdateUser(userID uint, req UpdateUserRequest) (*User, error)
	DeleteUser(userID uint) error
	GetTotalUsersCount() (int, error)

	// LogUserActivity Логи активности
	LogUserActivity(user *User, action, resourceType string, resourceID *uint, details, ip, userAgent string) error
	GetUserActivityLogs(filters ActivityLogFilters) ([]UserActivityLog, error)

	// GenerateCSRFToken CSRF токены
	GenerateCSRFToken(userID *uint) (string, error)
}

// CreateUserRequest запрос на создание пользователя
type CreateUserRequest struct {
	Email    string
	Name     string
	Initials string
	INN      string
	Password string
	Role     string
}

// UpdateUserRequest запрос на обновление пользователя
type UpdateUserRequest struct {
	Name     string
	Email    string
	Initials string
	INN      string
	Role     string
}

// authService реализует AuthService
type authService struct {
	userRepo     UserRepository
	activityRepo ActivityLogRepository
	sessionStore SessionStore
	rateLimiter  RateLimiter
}

// NewAuthService создает новый сервис аутентификации
func NewAuthService(userRepo UserRepository, activityRepo ActivityLogRepository, sessionStore SessionStore, rateLimiter RateLimiter) AuthService {
	return &authService{
		userRepo:     userRepo,
		activityRepo: activityRepo,
		sessionStore: sessionStore,
		rateLimiter:  rateLimiter,
	}
}

// AuthenticateUser аутентифицирует пользователя
func (s *authService) AuthenticateUser(email, password string) (*User, error) {
	// Проверка rate limiting (если rate limiter реализован)
	clientIP := "unknown" // В реальности передавать из контекста
	if s.rateLimiter != nil {
		if limited, _ := s.rateLimiter.IsLimited(clientIP+":user", 10, time.Minute); limited {
			logrus.WithFields(logrus.Fields{
				"email": email,
				"ip":    clientIP,
			}).Warn("Rate limit exceeded for user login")
			return nil, fmt.Errorf("слишком много попыток входа")
		}
	}

	// Поиск пользователя
	user, err := s.userRepo.FindByEmailOrName(email)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"email": email,
			"ip":    clientIP,
		}).Warn("User not found during login")
		time.Sleep(time.Second) // Искусственная задержка
		return nil, fmt.Errorf("неверные учетные данные")
	}

	// Проверка провайдера и пароля
	if user.Provider != "local" || user.Password == "" {
		logrus.WithFields(logrus.Fields{
			"email": email,
			"ip":    clientIP,
		}).Warn("Attempt to login with non-local user")
		time.Sleep(time.Second)
		return nil, fmt.Errorf("неверные учетные данные")
	}

	// Проверка пароля
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		logrus.WithFields(logrus.Fields{
			"email": email,
			"ip":    clientIP,
		}).Warn("Invalid password during login")
		time.Sleep(time.Second)
		return nil, fmt.Errorf("неверные учетные данные")
	}

	logrus.WithFields(logrus.Fields{
		"email": user.Email,
		"ip":    clientIP,
	}).Info("User authenticated successfully")

	return user, nil
}

// CreateUserFromGoogle создает пользователя из Google OAuth
func (s *authService) CreateUserFromGoogle(email, name string) (*User, error) {
	// Проверяем, существует ли пользователь
	existing, err := s.userRepo.FindByEmailOrName(email)
	if err == nil {
		logrus.WithField("email", email).Info("Existing Google user found")
		return existing, nil
	}

	// Создаем нового пользователя
	user := &User{
		Email:    email,
		Name:     name,
		Provider: "google",
		Role:     "operator",
	}

	if err := s.userRepo.Create(user); err != nil {
		logrus.WithError(err).WithField("email", email).Error("Failed to create Google user")
		return nil, err
	}

	logrus.WithField("email", email).Info("New Google user created")
	return user, nil
}

// GetCurrentUser получает текущего пользователя
func (s *authService) GetCurrentUser(userID uint) (*User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get current user")
		return nil, err
	}

	// Не возвращаем пароль
	user.Password = ""
	return user, nil
}

// Logout выполняет выход пользователя
func (s *authService) Logout(userID uint) error {
	// Очистка CSRF токена
	userIDStr := strconv.FormatUint(uint64(userID), 10)
	csrfMutex.Lock()
	delete(csrfTokens, userIDStr)
	csrfMutex.Unlock()

	logrus.WithField("user_id", userID).Info("User logged out")
	return nil
}

// CreateUser создает нового пользователя
func (s *authService) CreateUser(req CreateUserRequest) (*User, error) {
	// Проверяем, существует ли пользователь
	exists, err := s.userRepo.ExistsByEmailOrName(req.Email, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("пользователь с таким email или именем уже существует")
	}

	// Хэшируем пароль
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		return nil, fmt.Errorf("не удалось создать пользователя")
	}

	// Создаем пользователя
	user := &User{
		Email:    req.Email,
		Name:     req.Name,
		Initials: req.Initials,
		INN:      req.INN,
		Provider: "local",
		Role:     req.Role,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		logrus.WithError(err).WithField("email", req.Email).Error("Failed to create user")
		return nil, err
	}

	// Не возвращаем пароль в ответе
	user.Password = ""

	logrus.WithFields(logrus.Fields{
		"email": req.Email,
		"role":  req.Role,
	}).Info("User created successfully")

	return user, nil
}

// GetUsers получает список всех пользователей
func (s *authService) GetUsers() ([]User, error) {
	users, err := s.userRepo.FindAll()
	if err != nil {
		logrus.WithError(err).Error("Failed to get users")
		return nil, err
	}

	// Убираем пароли
	for i := range users {
		users[i].Password = ""
	}

	return users, nil
}

// UpdateUser обновляет пользователя
func (s *authService) UpdateUser(userID uint, req UpdateUserRequest) (*User, error) {
	// Проверяем существование
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, fmt.Errorf("пользователь не найден")
	}

	// Проверяем уникальность email, если он изменяется
	if req.Email != "" && req.Email != user.Email {
		exists, err := s.userRepo.ExistsByEmailOrName(req.Email, "")
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, fmt.Errorf("пользователь с таким email уже существует")
		}
	}

	// Собираем обновления
	updates := make(map[string]interface{})
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Email != "" {
		updates["email"] = req.Email
	}
	if req.Initials != "" {
		updates["initials"] = req.Initials
	}
	if req.INN != "" {
		updates["inn"] = req.INN
	}
	if req.Role != "" {
		updates["role"] = req.Role
	}

	if len(updates) == 0 {
		return nil, fmt.Errorf("нет полей для обновления")
	}

	if err := s.userRepo.Update(userID, updates); err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to update user")
		return nil, err
	}

	// Получаем обновленного пользователя
	updatedUser, err := s.userRepo.FindByID(userID)
	if err != nil {
		return nil, err
	}

	updatedUser.Password = ""

	logrus.WithField("user_id", userID).Info("User updated successfully")
	return updatedUser, nil
}

// DeleteUser удаляет пользователя
func (s *authService) DeleteUser(userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		return fmt.Errorf("пользователь не найден")
	}

	if err := s.userRepo.Delete(userID); err != nil {
		logrus.WithError(err).WithField("email", user.Email).Error("Failed to delete user")
		return err
	}

	logrus.WithField("email", user.Email).Info("User deleted successfully")
	return nil
}

// LogUserActivity логирует активность пользователя
func (s *authService) LogUserActivity(user *User, action, resourceType string, resourceID *uint, details, ip, userAgent string) error {
	logEntry := &UserActivityLog{
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

	if err := s.activityRepo.Create(logEntry); err != nil {
		logrus.WithError(err).WithField("user_id", user.ID).Error("Failed to log user activity")
		// Не возвращаем ошибку, чтобы не ломать пользовательский поток
	}

	return nil
}

// GetUserActivityLogs получает логи активности
func (s *authService) GetUserActivityLogs(filters ActivityLogFilters) ([]UserActivityLog, error) {
	logs, err := s.activityRepo.FindWithFilters(filters)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user activity logs")
		return nil, err
	}

	return logs, nil
}

// GetTotalUsersCount получает общее количество пользователей
func (s *authService) GetTotalUsersCount() (int, error) {
	return s.userRepo.CountAll()
}

// GenerateCSRFToken генерирует CSRF токен
func (s *authService) GenerateCSRFToken(userID *uint) (string, error) {
	if userID == nil {
		return generateCsrfToken(), nil
	}

	userIDStr := strconv.FormatUint(uint64(*userID), 10)

	csrfMutex.Lock()
	defer csrfMutex.Unlock()

	csrfTokenEntry, exists := csrfTokens[userIDStr]
	if !exists || time.Now().After(csrfTokenEntry.expiresAt) {
		token := generateCsrfToken()
		csrfTokens[userIDStr] = csrfToken{
			token:     token,
			expiresAt: time.Now().Add(24 * time.Hour),
		}
		return token, nil
	}

	return csrfTokenEntry.token, nil
}
