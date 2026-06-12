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

type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

type RateLimiter interface {
	IsLimited(key string, limit int, window time.Duration) (bool, error)
}

type AuthService interface {
	AuthenticateUser(email, password string) (*User, error)
	CreateUserFromGoogle(email, name string) (*User, error)
	GetCurrentUser(userID int64) (*User, error)
	Logout(userID int64) error

	CreateUser(req CreateUserRequest) (*User, error)
	GetUsers() ([]User, error)
	UpdateUser(userID int64, req UpdateUserRequest) (*User, error)
	DeleteUser(userID int64) error
	GetTotalUsersCount() (int, error)

	LogUserActivity(user *User, action, resourceType string, resourceID *int64, details, ip, userAgent string) error
	GetUserActivityLogs(filters ActivityLogFilters) ([]UserActivityLog, error)

	GenerateCSRFToken(userID *int64) (string, error)
}

type CreateUserRequest struct {
	Email    string
	Name     string
	Initials string
	INN      string
	Password string
	Role     string
}

type UpdateUserRequest struct {
	Name     string
	Email    string
	Initials string
	INN      string
	Role     string
}

type authService struct {
	userRepo     UserRepository
	activityRepo ActivityLogRepository
	sessionStore SessionStore
	rateLimiter  RateLimiter
}

func NewAuthService(userRepo UserRepository, activityRepo ActivityLogRepository, sessionStore SessionStore, rateLimiter RateLimiter) AuthService {
	return &authService{
		userRepo:     userRepo,
		activityRepo: activityRepo,
		sessionStore: sessionStore,
		rateLimiter:  rateLimiter,
	}
}

func (s *authService) AuthenticateUser(email, password string) (*User, error) {
	clientIP := "unknown"
	if s.rateLimiter != nil {
		if limited, _ := s.rateLimiter.IsLimited(clientIP+":user", 10, time.Minute); limited {
			logrus.WithFields(logrus.Fields{
				"email": email,
				"ip":    clientIP,
			}).Warn("Rate limit exceeded for user login")
			return nil, fmt.Errorf("слишком много попыток входа")
		}
	}

	user, err := s.userRepo.FindByEmailOrName(email)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"email": email,
			"ip":    clientIP,
		}).Warn("User not found during login")
		time.Sleep(time.Second)
		return nil, fmt.Errorf("неверные учетные данные")
	}

	if user.Provider != "local" || user.Password == "" {
		logrus.WithFields(logrus.Fields{
			"email": email,
			"ip":    clientIP,
		}).Warn("Attempt to login with non-local user")
		time.Sleep(time.Second)
		return nil, fmt.Errorf("неверные учетные данные")
	}

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

func (s *authService) CreateUserFromGoogle(email, name string) (*User, error) {
	existing, err := s.userRepo.FindByEmailOrName(email)
	if err == nil {
		logrus.WithField("email", email).Info("Existing Google user found")
		return existing, nil
	}

	user := &User{
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

func (s *authService) GetCurrentUser(userID int64) (*User, error) {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get current user")
		return nil, err
	}

	user.Password = ""
	return user, nil
}

func (s *authService) Logout(userID int64) error {
	userIDStr := strconv.FormatInt(userID, 10)
	csrfMutex.Lock()
	delete(csrfTokens, userIDStr)
	csrfMutex.Unlock()

	logrus.WithField("user_id", userID).Info("User logged out")
	return nil
}

func (s *authService) CreateUser(req CreateUserRequest) (*User, error) {
	exists, err := s.userRepo.ExistsByEmailOrName(req.Email, req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, fmt.Errorf("пользователь с таким email или именем уже существует")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		logrus.WithError(err).Error("Failed to hash password")
		return nil, fmt.Errorf("не удалось создать пользователя")
	}

	user := &User{
		Email:    req.Email,
		Name:     req.Name,
		Initials: req.Initials,
		INN:      req.INN,
		Provider: "local",
		Role:     req.Role,
		Password: string(hashedPassword),
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

func (s *authService) GetUsers() ([]User, error) {
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

func (s *authService) UpdateUser(userID int64, req UpdateUserRequest) (*User, error) {
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

	updateParams := UpdateUserParams{
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

	updatedUser, err := s.userRepo.Update(userID, updateParams)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to update user")
		return nil, err
	}

	updatedUser.Password = ""

	logrus.WithField("user_id", userID).Info("User updated successfully")
	return updatedUser, nil
}

func (s *authService) DeleteUser(userID int64) error {
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

func (s *authService) LogUserActivity(user *User, action, resourceType string, resourceID *int64, details, ip, userAgent string) error {
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

	if _, err := s.activityRepo.Create(logEntry); err != nil {
		logrus.WithError(err).WithField("user_id", user.ID).Error("Failed to log user activity")
	}

	return nil
}

func (s *authService) GetUserActivityLogs(filters ActivityLogFilters) ([]UserActivityLog, error) {
	logs, err := s.activityRepo.FindWithFilters(filters)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user activity logs")
		return nil, err
	}

	return logs, nil
}

func (s *authService) GetTotalUsersCount() (int, error) {
	return s.userRepo.CountAll()
}

func (s *authService) GenerateCSRFToken(userID *int64) (string, error) {
	if userID == nil {
		return generateCsrfToken(), nil
	}

	userIDStr := strconv.FormatInt(*userID, 10)

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
