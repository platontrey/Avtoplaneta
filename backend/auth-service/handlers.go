package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	authService AuthService
	config      *Config
}

var googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"


func NewHandler(authService AuthService, config *Config) *Handler {
	return &Handler{
		authService: authService,
		config:      config,
	}
}

func (h *Handler) GoogleAuthHandler(c *gin.Context) {
	if os.Getenv("GOOGLE_CLIENT_ID") == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google аутентификация не настроена"})
		return
	}

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

func (h *Handler) GoogleAuthCallbackHandler(c *gin.Context) {
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		logrus.WithError(err).WithField("ip", c.ClientIP()).Warn("Google OAuth callback failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось завершить аутентификацию"})
		return
	}

	createdUser, err := h.authService.CreateUserFromGoogle(user.Email, user.Name)
	if err != nil {
		logrus.WithError(err).WithField("email", user.Email).Error("Failed to create user from Google")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя"})
		return
	}

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

	c.JSON(http.StatusOK, createdUser)
}

func (h *Handler) GoogleMobileAuthHandler(c *gin.Context) {
	var req struct {
		IDToken string `json:"id_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id_token обязателен"})
		return
	}

	client := &http.Client{Timeout: 10 * time.Second}
	reqURL := fmt.Sprintf("%s?id_token=%s", googleTokenInfoURL, req.IDToken)

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), "GET", reqURL, nil)
	if err != nil {
		logrus.WithError(err).Error("Failed to create HTTP request for Google tokeninfo")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка сервера"})
		return
	}

	resp, err := client.Do(httpReq)
	if err != nil {
		logrus.WithError(err).Error("Failed to call Google tokeninfo API")
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Не удалось связаться с сервером Google"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		logrus.WithFields(logrus.Fields{
			"status": resp.StatusCode,
			"body":   string(bodyBytes),
		}).Warn("Google tokeninfo returned non-200 status")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Невалидный id_token"})
		return
	}

	var tokenInfo struct {
		Email         string `json:"email"`
		EmailVerified string `json:"email_verified"`
		Name          string `json:"name"`
		Error         string `json:"error"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenInfo); err != nil {
		logrus.WithError(err).Error("Failed to decode Google tokeninfo response")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка обработки ответа от Google"})
		return
	}

	if tokenInfo.Error != "" {
		logrus.WithField("google_error", tokenInfo.Error).Warn("Google tokeninfo returned error")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Невалидный id_token"})
		return
	}

	if tokenInfo.EmailVerified != "true" || tokenInfo.Email == "" {
		logrus.WithField("email_verified", tokenInfo.EmailVerified).Warn("Google email not verified or empty")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Email не подтвержден в Google"})
		return
	}

	createdUser, err := h.authService.CreateUserFromGoogle(tokenInfo.Email, tokenInfo.Name)
	if err != nil {
		logrus.WithError(err).WithField("email", tokenInfo.Email).Error("Failed to create user from Google (mobile)")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось войти через Google"})
		return
	}

	session, _ := store.Get(c.Request, "auth-session")
	session.Values["user_id"] = createdUser.ID
	session.Values["login_time"] = time.Now().Unix()
	if err := session.Save(c.Request, c.Writer); err != nil {
		logrus.WithError(err).Error("Failed to save session (mobile)")
	}

	accessToken, refreshToken, err := GenerateJWTTokens(createdUser, h.config)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate JWT tokens (mobile)")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать токены авторизации"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"email": createdUser.Email,
		"ip":    c.ClientIP(),
	}).Info("Google mobile login successful")

	response := LoginResponse{
		User:         *createdUser,
		Message:      "Login successful",
		Token:        accessToken,
		RefreshToken: refreshToken,
	}
	c.JSON(http.StatusOK, response)
}


func (h *Handler) UserLoginHandler(c *gin.Context) {
	var loginReq struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email и пароль обязательны"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"email":        loginReq.Email,
		"password_len": len(loginReq.Password),
	}).Info("Login attempt")

	user, err := h.authService.AuthenticateUser(loginReq.Email, loginReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	session, _ := store.Get(c.Request, "auth-session")
	session.Values["user_id"] = user.ID
	session.Values["login_time"] = time.Now().Unix()

	if err := session.Save(c.Request, c.Writer); err != nil {
		logrus.WithError(err).Error("Failed to save session")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сессию"})
		return
	}

	accessToken, refreshToken, err := GenerateJWTTokens(user, h.config)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate JWT tokens")
		accessToken = ""
		refreshToken = ""
	}

	response := LoginResponse{
		User:         *user,
		Message:      "Login successful",
		Token:        accessToken,
		RefreshToken: refreshToken,
	}
	c.JSON(http.StatusOK, response)
}

func (h *Handler) LogoutHandler(c *gin.Context) {
	var userEmail string
	if user, exists := c.Get("user"); exists {
		userEmail = user.(User).Email
	}

	session, err := store.Get(c.Request, "auth-session")
	if err != nil {
		logrus.WithError(err).WithField("ip", c.ClientIP()).Warn("Failed to get session during logout")
	}

	var userID int64
	if session != nil {
		val, ok := session.Values["user_id"]
		if ok {
			userID, _ = getUserIDFromSessionValue(val)
		}
	}

	if session != nil {
		session.Values = make(map[interface{}]interface{})
		session.Options.MaxAge = -1

		if err := session.Save(c.Request, c.Writer); err != nil {
			logrus.WithError(err).WithField("ip", c.ClientIP()).Warn("Failed to save session during logout")
		}
	}

	if userID != 0 {
		if err := h.authService.Logout(userID); err != nil {
			logrus.WithError(err).WithField("user_id", userID).Warn("Failed to logout user")
		}
	}

	logrus.WithField("email", userEmail).WithField("ip", c.ClientIP()).Info("User logged out")
	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен успешно"})
}

func (h *Handler) GetCurrentUserHandler(c *gin.Context) {
	logrus.WithFields(logrus.Fields{
		"cookies_count": len(c.Request.Cookies()),
	}).Info("GetCurrentUserHandler: received request")

	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := ValidateJWTToken(tokenString, h.config.JWTSecret)
		if err != nil {
			logrus.WithError(err).Warn("GetCurrentUserHandler: invalid JWT token")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный токен"})
			return
		}
		user, err := h.authService.GetCurrentUser(claims.UserID)
		if err != nil {
			logrus.WithError(err).WithField("user_id", claims.UserID).Error("GetCurrentUserHandler: user not found by JWT claims")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
			return
		}
		c.JSON(http.StatusOK, user)
		return
	}

	session, err := store.Get(c.Request, "auth-session")
	if err != nil {
		logrus.WithError(err).Warn("Failed to get session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сеанс"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"session_values": session.Values,
		"session_id":     session.ID,
	}).Info("GetCurrentUserHandler: session details")

	userID, ok := getUserIDFromSessionValue(session.Values["user_id"])
	if !ok {
		logrus.Warn("GetCurrentUserHandler: user_id not found in session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
		return
	}

	if loginTime, ok := session.Values["login_time"].(int64); ok {
		if time.Now().Unix()-loginTime > 86400*30 {
			logrus.WithField("user_id", userID).Warn("Session expired")
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Сеанс истек, пожалуйста, войдите снова"})
			return
		}
	}

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("Failed to get current user")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) GetCSRFTokenHandler(c *gin.Context) {
	session, _ := store.Get(c.Request, "auth-session")
	var userID *int64

	if val, ok := session.Values["user_id"]; ok && val != nil {
		uid, _ := getUserIDFromSessionValue(val)
		userID = &uid
	}

	token, err := h.authService.GenerateCSRFToken(userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate CSRF token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать токен CSRF"})
		return
	}

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

func (h *Handler) UpdateUserHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}

	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	user, err := h.authService.UpdateUser(userID, req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *Handler) DeleteUserHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}

	if err := h.authService.DeleteUser(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Пользователь удален"})
}

func (h *Handler) GetServerStatusHandler(c *gin.Context) {
	totalUsers, err := h.authService.GetTotalUsersCount()
	if err != nil {
		logrus.WithError(err).Error("Failed to get total users count")
		totalUsers = 0
	}

	totalParts, err := h.getTotalPartsCount()
	if err != nil {
		logrus.WithError(err).Error("Failed to get total parts count")
		totalParts = 0
	}

	status := map[string]interface{}{
		"server": map[string]interface{}{
			"status":     "ok",
			"uptime":     "unknown",
			"go_version": "1.21",
			"os":         "linux",
			"arch":       "amd64",
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

func (h *Handler) GetServerLogsHandler(c *gin.Context) {
	logrus.Info("GetServerLogsHandler called - returning mock logs")

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

func (h *Handler) GetUserActivityLogsHandler(c *gin.Context) {
	var filters ActivityLogFilters

	if userIDStr := c.Query("user_id"); userIDStr != "" {
		if userID, err := strconv.ParseInt(userIDStr, 10, 64); err == nil {
			filters.UserID = &userID
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
		filters.Limit = 100
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

func (h *Handler) LogUserActivityHandler(c *gin.Context) {
	var req struct {
		Action       string `json:"action"`
		ResourceType string `json:"resource_type"`
		ResourceID   *int64 `json:"resource_id"`
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
		logrus.WithError(err).Warn("Failed to log user activity")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Активность залогирована"})
}

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

func (h *Handler) InternalLogUserActivityHandler(c *gin.Context) {
	var req struct {
		Action       string `json:"action"`
		ResourceType string `json:"resource_type"`
		ResourceID   *int64 `json:"resource_id"`
		Details      string `json:"details"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные"})
		return
	}

	userIDStr := c.GetHeader("X-User-ID")
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Отсутствует X-User-ID"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный X-User-ID"})
		return
	}

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Warn("Failed to get user for internal logging")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Пользователь не найден"})
		return
	}

	if err := h.authService.LogUserActivity(user, req.Action, req.ResourceType, req.ResourceID, req.Details, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		logrus.WithError(err).Warn("Failed to log user activity internally")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Активность залогирована"})
}

func (h *Handler) RefreshTokenHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token обязателен"})
		return
	}

	userID, err := ValidateRefreshToken(req.RefreshToken, h.config.JWTSecret)
	if err != nil {
		logrus.WithError(err).Warn("RefreshTokenHandler: invalid refresh token")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Недействительный refresh token"})
		return
	}

	user, err := h.authService.GetCurrentUser(userID)
	if err != nil {
		logrus.WithError(err).WithField("user_id", userID).Error("RefreshTokenHandler: user not found")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	accessToken, refreshToken, err := GenerateJWTTokens(user, h.config)
	if err != nil {
		logrus.WithError(err).Error("RefreshTokenHandler: failed to generate tokens")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить токен"})
		return
	}

	logrus.WithField("user_id", userID).Info("JWT tokens refreshed")
	c.JSON(http.StatusOK, gin.H{
		"token":         accessToken,
		"refresh_token": refreshToken,
		"user":          user,
	})
}

func (h *Handler) getTotalPartsCount() (int, error) {
	url := fmt.Sprintf("%s/api/statistics", h.config.PartsServiceURL)
	req, err := http.NewRequest("GET", url, nil)
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
