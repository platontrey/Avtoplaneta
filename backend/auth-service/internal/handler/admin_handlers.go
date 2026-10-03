package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"

	partsv1 "avtoplaneta/gen/parts/v1"

	"auth-service/internal/domain"
)

// CreateUserHandler создает нового пользователя (admin-only)
func (h *Handler) CreateUserHandler(c *gin.Context) {
	var req domain.CreateUserRequest
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

// GetUsersHandler возвращает список всех пользователей
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

// UpdateUserHandler обновляет данные пользователя (admin-only)
func (h *Handler) UpdateUserHandler(c *gin.Context) {
	userIDStr := c.Param("id")
	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверный ID пользователя"})
		return
	}

	var req domain.UpdateUserRequest
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

// DeleteUserHandler удаляет пользователя (admin-only)
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

// GetServerStatusHandler возвращает диагностический статус системы
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

// GetServerLogsHandler возвращает логи сервера
func (h *Handler) GetServerLogsHandler(c *gin.Context) {
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

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// GetUserActivityLogsHandler возвращает список логов активности с фильтрацией
func (h *Handler) GetUserActivityLogsHandler(c *gin.Context) {
	var filters domain.ActivityLogFilters

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
	if usefulOnlyStr := c.Query("useful_only"); usefulOnlyStr != "" {
		if usefulOnly, err := strconv.ParseBool(usefulOnlyStr); err == nil {
			filters.UsefulOnly = usefulOnly
		}
	}
	if startDateStr := c.Query("start_date"); startDateStr != "" {
		if startDate, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			filters.StartDate = &startDate
		} else if startDate, err := time.Parse("2006-01-02", startDateStr); err == nil {
			filters.StartDate = &startDate
		}
	}
	if endDateStr := c.Query("end_date"); endDateStr != "" {
		if endDate, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			filters.EndDate = &endDate
		} else if endDate, err := time.Parse("2006-01-02", endDateStr); err == nil {
			endOfDay := endDate.Add(24*time.Hour - time.Nanosecond)
			filters.EndDate = &endOfDay
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

// LogUserActivityHandler записывает активность текущего пользователя
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

	u := user.(domain.User)
	if err := h.authService.LogUserActivity(&u, req.Action, req.ResourceType, req.ResourceID, req.Details, c.ClientIP(), c.GetHeader("User-Agent")); err != nil {
		logrus.WithError(err).Warn("Failed to log user activity")
	}

	c.JSON(http.StatusOK, gin.H{"message": "Активность залогирована"})
}

// InternalGetUsersHandler возвращает список пользователей для межсервисных вызовов
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

// InternalLogUserActivityHandler записывает активность от лица других внутренних сервисов
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

func (h *Handler) getTotalPartsCount() (int, error) {
	if h.partsClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		resp, err := h.partsClient.GetStatistics(ctx, &partsv1.GetStatisticsRequest{})
		if err == nil {
			logrus.WithField("total_parts", resp.TotalParts).Info("Fetched total parts count from parts-service via gRPC")
			return int(resp.TotalParts), nil
		}
		logrus.WithError(err).Warn("Failed to fetch parts statistics via gRPC, attempting HTTP fallback")
	}

	if h.config == nil || h.config.PartsServiceURL == "" {
		return 0, nil
	}

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
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("parts service returned status %d", resp.StatusCode)
	}

	var stats struct {
		TotalParts int `json:"total_parts"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&stats); err != nil {
		logrus.WithError(err).Error("Failed to decode parts statistics response")
		return 0, err
	}

	return stats.TotalParts, nil
}
