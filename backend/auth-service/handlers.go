package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"github.com/sirupsen/logrus"
)

// Handler содержит все HTTP handlers для auth-service
type Handler struct {
	authService AuthService
}

// NewHandler создает новый handler с dependency injection
func NewHandler(authService AuthService) *Handler {
	return &Handler{
		authService: authService,
	}
}

// GoogleAuthHandler начинает процесс OAuth аутентификации через Google
func (h *Handler) GoogleAuthHandler(c *gin.Context) {
	// Проверяем, настроен ли Google OAuth
	if os.Getenv("GOOGLE_CLIENT_ID") == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google аутентификация не настроена"})
		return
	}

	// Начинаем процесс OAuth аутентификации
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// GoogleAuthCallbackHandler обрабатывает callback от Google OAuth
func (h *Handler) GoogleAuthCallbackHandler(c *gin.Context) {
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		logrus.WithError(err).WithField("ip", c.ClientIP()).Warn("Google OAuth callback failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось завершить аутентификацию"})
		return
	}

	// Создаем пользователя через сервис
	createdUser, err := h.authService.CreateUserFromGoogle(user.Email, user.Name)
	if err != nil {
		logrus.WithError(err).WithField("email", user.Email).Error("Failed to create user from Google")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя"})
		return
	}

	// Сохраняем в сессии
	session, _ := store.Get(c.Request, "auth-session")
	session.Values["user_id"] = createdUser.ID
	session.Values["login_time"] = time.Now().Unix()

	if err := session.Save(c.Request, c.Writer); err != nil {
		logrus.WithError(err).Error("Failed to save session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сессию"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"email": createdUser.Email,
		"ip":    c.ClientIP(),
	}).Info("Google login successful")

	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": createdUser})
}

// UserLoginHandler обрабатывает вход пользователя с email и паролем
func (h *Handler) UserLoginHandler(c *gin.Context) {
	var loginReq struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email и пароль обязательны"})
		return
	}

	// Аутентифицируем через сервис
	user, err := h.authService.AuthenticateUser(loginReq.Email, loginReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Сохраняем в сессии
	session, _ := store.Get(c.Request, "auth-session")
	session.Values["user_id"] = user.ID
	session.Values["login_time"] = time.Now().Unix()

	if err := session.Save(c.Request, c.Writer); err != nil {
		logrus.WithError(err).Error("Failed to save session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сессию"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Вход выполнен успешно", "user": user})
}

// LogoutHandler обрабатывает выход пользователя
func (h *Handler) LogoutHandler(c *gin.Context) {
	// Получаем информацию о пользователе для логирования
	var userEmail string
	if user, exists := c.Get("user"); exists {
		userEmail = user.(User).Email
	}

	// Получаем сессию
	session, err := store.Get(c.Request, "auth-session")
	if err != nil {
		logrus.WithError(err).WithField("ip", c.ClientIP()).Warn("Failed to get session during logout")
		// Продолжаем в любом случае
	}

	// Получаем user_id перед очисткой
	var userID uint
	if session != nil && session.Values["user_id"] != nil {
		userID = session.Values["user_id"].(uint)
	}

	// Очищаем сессию
	if session != nil {
		session.Values = make(map[interface{}]interface{})
		session.Options.MaxAge = -1

		if err := session.Save(c.Request, c.Writer); err != nil {
			logrus.WithError(err).WithField("ip", c.ClientIP()).Warn("Failed to save session during logout")
		}
	}

	// Выполняем logout через сервис
	if userID != 0 {
		if err := h.authService.Logout(userID); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Warn("Failed to logout user")
		}
	}

	logrus.WithField("email", userEmail).WithField("ip", c.ClientIP()).Info("User logged out")
	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен успешно"})
}

// GetCurrentUserHandler возвращает информацию о текущем пользователе
func (h *Handler) GetCurrentUserHandler(c *gin.Context) {
	session, err := store.Get(c.Request, "auth-session")
	if err != nil {
		logrus.WithError(err).Warn("Failed to get session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сеанс"})
		return
	}

	userID, ok := session.Values["user_id"]
	if !ok || userID == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
		return
	}

	// Проверяем возраст сессии
	if loginTime, ok := session.Values["login_time"].(int64); ok {
		if time.Now().Unix()-loginTime > 86400*30 { // 30 дней
			logrus.WithField("user_id", userID).Warn("Session expired")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Сеанс истек, пожалуйста, войдите снова"})
			return
		}
	}

	// Получаем пользователя через сервис
	user, err := h.authService.GetCurrentUser(userID.(uint))
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get current user")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	c.JSON(http.StatusOK, user)
}

// GetCSRFTokenHandler возвращает токен CSRF для текущего пользователя
func (h *Handler) GetCSRFTokenHandler(c *gin.Context) {
	session, _ := store.Get(c.Request, "auth-session")
	var userID *uint

	if session.Values["user_id"] != nil {
		uid := session.Values["user_id"].(uint)
		userID = &uid
	}

	// Генерируем токен через сервис
	token, err := h.authService.GenerateCSRFToken(userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate CSRF token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать токен CSRF"})
		return
	}

	// Сохраняем в сессии для неаутентифицированных пользователей
	if userID == nil {
		session.Values["csrf_token"] = token
		if err := session.Save(c.Request, c.Writer); err != nil {
			logrus.WithError(err).Error("Failed to save CSRF token in session")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить токен CSRF"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"csrf_token": token})
}

// CreateUserHandler создает нового пользователя (только для админов)
func (h *Handler) CreateUserHandler(c *gin.Context) {
	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	user, err := h.authService.CreateUser(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

// GetUsersHandler получает список пользователей (для менеджеров и выше)
func (h *Handler) GetUsersHandler(c *gin.Context) {
	users, err := h.authService.GetUsers()
	if err != nil {
		logrus.WithError(err).Error("Failed to get users")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить пользователей"})
		return
	}

	logrus.WithField("users_count", len(users)).Info("Returning users list")
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// UpdateUserHandler обновляет пользователя (только для админов)
func (h *Handler) UpdateUserHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	user, err := h.authService.UpdateUser(uint(userID), req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

// DeleteUserHandler удаляет пользователя (только для админов)
func (h *Handler) DeleteUserHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}

	if err := h.authService.DeleteUser(uint(userID)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пользователь удален"})
}

// GetServerStatusHandler получает статус сервера (только для админов)
func (h *Handler) GetServerStatusHandler(c *gin.Context) {
	// Получаем статистику пользователей из базы данных
	totalUsers, err := h.authService.GetTotalUsersCount()
	if err != nil {
		logrus.WithError(err).Error("Failed to get total users count")
		totalUsers = 0
	}

	// Получаем статистику запчастей через запрос к parts-service
	totalParts, err := h.getTotalPartsCount()
	if err != nil {
		logrus.WithError(err).Error("Failed to get total parts count")
		totalParts = 0
	}

	status := map[string]interface{}{
		"server": map[string]interface{}{
			"status":     "ok",
			"uptime":     "unknown", // TODO: реализовать получение uptime
			"go_version": "1.21",    // TODO: получить из runtime
			"os":         "linux",   // TODO: получить из runtime
			"arch":       "amd64",   // TODO: получить из runtime
		},
		"database": map[string]interface{}{
			"status":      "ok",
			"total_parts": totalParts,
			"total_users": totalUsers,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, status)
}

// GetServerLogsHandler получает логи сервера (только для админов)
func (h *Handler) GetServerLogsHandler(c *gin.Context) {
	logrus.Info("GetServerLogsHandler called - returning mock logs")

	// Заглушка - в реальности нужно реализовать чтение логов
	logs := []map[string]interface{}{
		{
			"timestamp": time.Now().Format(time.RFC3339),
			"level":     "INFO",
			"message":   "Server started successfully",
		},
		{
			"timestamp": time.Now().Add(-time.Minute).Format(time.RFC3339),
			"level":     "INFO",
			"message":   "All services are running",
		},
	}

	logrus.WithField("logs_count", len(logs)).Info("Returning server logs")
	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// GetUserActivityLogsHandler получает логи активности пользователей
func (h *Handler) GetUserActivityLogsHandler(c *gin.Context) {
	var filters ActivityLogFilters

	// Парсим query параметры
	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseUint(userIDStr, 10, 32); err == nil {
			userIDUint := uint(userID)
			filters.UserID = &userIDUint
		}
	}
	if action := c.Query("action"); action != "" {
		filters.Action = action
	}
	if resourceType := c.Query("resource_type"); resourceType != "" {
		filters.ResourceType = resourceType
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters.EndDate = &endDate
		}
	}
	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil && limit > 0 {
			filters.Limit = limit
		}
	} else {
		filters.Limit = 100 // default limit
	}
	if offsetStr := c.Query("offset"); offsetStr != "" {
		if offset, err := strconv.Atoi(offsetStr); err == nil && offset >= 0 {
			filters.Offset = offset
		}
	}

	logs, err := h.authService.GetUserActivityLogs(filters)
	if err != nil {
		logrus.WithError(err).Error("Failed to get user activity logs")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить логи"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// LogUserActivityHandler логирует активность пользователя
func (h *Handler) LogUserActivityHandler(c *gin.Context) {
	var req struct {
		Action       string `json:"action"`
		ResourceType string `json:"resource_type"`
		ResourceID   *uint  `json:"resource_id"`
		Details      string `json:"details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
		return
	}

	u := user.(User)
	if err := h.authService.LogUserActivity(&u, req.Action, req.ResourceType, req.ResourceID, req.Details, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		// Логирование не должно ломать пользовательский поток
		logrus.WithError(err).Warn("Failed to log user activity")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Активность залогирована"})
}

// InternalGetUsersHandler получает список пользователей для внутренних сервисов (без аутентификации)
func (h *Handler) InternalGetUsersHandler(c *gin.Context) {
	users, err := h.authService.GetUsers()
	if err != nil {
		logrus.WithError(err).Error("Failed to get users for internal request")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить пользователей"})
		return
	}

	logrus.WithField("users_count", len(users)).Info("Returning users list for internal request")
	c.JSON(http.StatusOK, gin.H{"users": users})
}

// InternalLogUserActivityHandler логирует активность пользователя для внутренних сервисов (без аутентификации)
func (h *Handler) InternalLogUserActivityHandler(c *gin.Context) {
	var req struct {
		Action       string `json:"action"`
		ResourceType string `json:"resource_type"`
		ResourceID   *uint  `json:"resource_id"`
		Details      string `json:"details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	// Получить userID из заголовка
	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Отсутствует X-User-ID"})
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный X-User-ID"})
		return
	}

	// Найти пользователя по ID
	user, err := h.authService.GetCurrentUser(uint(userID))
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Warn("Failed to get user for internal logging")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Логировать активность
	if err := h.authService.LogUserActivity(user, req.Action, req.ResourceType, req.ResourceID, req.Details, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		logrus.WithError(err).Warn("Failed to log user activity internally")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Активность залогирована"})
}

// getTotalPartsCount получает общее количество запчастей через запрос к parts-service
func (h *Handler) getTotalPartsCount() (int, error) {
	// Делаем запрос к parts-service для получения статистики
	req, err := http.NewRequest("GET", "http://localhost:8081/api/statistics", nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to create parts statistics request")
		return 0, err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		logrus.WithError(err).Error("Failed to fetch parts statistics")
		return 0, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			logrus.WithError(err).Warn("Failed to close parts statistics response body")
		}
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		logrus.WithField("status", resp.StatusCode).Error("Parts statistics request failed")
		return 0, fmt.Errorf("parts service returned status %d", resp.StatusCode)
	}

	var stats struct {
		TotalParts int `json:"total_parts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		logrus.WithError(err).Error("Failed to decode parts statistics response")
		return 0, err
	}

	logrus.WithField("total_parts", stats.TotalParts).Info("Fetched total parts count from parts-service")
	return stats.TotalParts, nil
}
