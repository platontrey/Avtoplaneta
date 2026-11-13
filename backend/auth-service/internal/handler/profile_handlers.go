package handler

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"auth-service/internal/domain"
	"avtoplaneta/pkg/logger"
)

const (
	maxAvatarFileSize = 5 * 1024 * 1024 // 5 MB
	avatarStorageDir  = "./uploads/avatars"
)

// getUserFromContext извлекает текущего пользователя из контекста Gin
func (h *Handler) getUserFromContext(c *gin.Context) (*domain.User, bool) {
	val, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Требуется авторизация"})
		return nil, false
	}

	user, ok := val.(domain.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка авторизации"})
		return nil, false
	}

	return &user, true
}

// UpdateProfileHandler обновляет имя, логин и ИНН текущего пользователя
func (h *Handler) UpdateProfileHandler(c *gin.Context) {
	user, ok := h.getUserFromContext(c)
	if !ok {
		return
	}

	var req domain.UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса: " + err.Error()})
		return
	}

	req.Username = strings.ToLower(strings.TrimSpace(req.Username))
	req.Name = strings.TrimSpace(req.Name)
	req.INN = strings.TrimSpace(req.INN)

	if req.Username == "" && req.Name == "" && req.INN == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Необходимо указать хотя бы одно поле для обновления"})
		return
	}

	updatedUser, err := h.authService.UpdateProfile(user.ID, req)
	if err != nil {
		slog.WarnContext(c.Request.Context(), "Failed to update user profile", logger.Err(err), slog.Int64("user_id", user.ID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Обновляем объект пользователя в контексте Gin
	c.Set("user", *updatedUser)

	c.JSON(http.StatusOK, gin.H{
		"message": "Профиль успешно обновлен",
		"user":    updatedUser,
	})
}

// ChangePasswordHandler изменяет пароль текущего пользователя
func (h *Handler) ChangePasswordHandler(c *gin.Context) {
	user, ok := h.getUserFromContext(c)
	if !ok {
		return
	}

	var req domain.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неверные данные запроса: " + err.Error()})
		return
	}

	if err := h.authService.ChangePassword(user.ID, req.OldPassword, req.NewPassword); err != nil {
		slog.WarnContext(c.Request.Context(), "Failed to change password", logger.Err(err), slog.Int64("user_id", user.ID))
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Пароль успешно изменен",
	})
}

// UploadAvatarHandler загружает и сохраняет аватар текущего пользователя
func (h *Handler) UploadAvatarHandler(c *gin.Context) {
	user, ok := h.getUserFromContext(c)
	if !ok {
		return
	}

	file, header, err := c.Request.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Файл аватара обязателен"})
		return
	}
	defer file.Close()

	if header.Size > maxAvatarFileSize {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Размер файла не должен превышать 5 МБ"})
		return
	}

	// Валидация MIME-типа по первым 512 байтам
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка чтения файла"})
		return
	}
	contentType := http.DetectContentType(buffer[:n])

	// Возвращаем указатель в начало файла
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Ошибка перемотки файла"})
		return
	}

	var ext string
	switch contentType {
	case "image/jpeg":
		ext = ".jpg"
	case "image/png":
		ext = ".png"
	case "image/webp":
		ext = ".webp"
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Неподдерживаемый формат изображения. Допустимы: JPEG, PNG, WebP"})
		return
	}

	if err := os.MkdirAll(avatarStorageDir, 0755); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to create avatars directory", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Внутренняя ошибка файлового хранилища"})
		return
	}

	filename := fmt.Sprintf("avatar_%d_%d%s", user.ID, time.Now().Unix(), ext)
	dstPath := filepath.Join(avatarStorageDir, filename)

	out, err := os.Create(dstPath)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to create avatar file", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить файл аватара"})
		return
	}
	defer out.Close()

	if _, err := io.Copy(out, file); err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to write avatar file", logger.Err(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось записать файл аватара"})
		return
	}

	// Удаляем старый аватар, если он был сохранен локально
	if strings.HasPrefix(user.AvatarURL, "/auth/avatars/") {
		oldFilename := filepath.Base(user.AvatarURL)
		if oldFilename != "." && oldFilename != "/" {
			_ = os.Remove(filepath.Join(avatarStorageDir, oldFilename))
		}
	}

	avatarURL := "/auth/avatars/" + filename
	updatedUser, err := h.authService.UpdateAvatar(user.ID, avatarURL)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to save avatar URL to DB", logger.Err(err), slog.Int64("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить профиль пользователя"})
		return
	}

	c.Set("user", *updatedUser)

	c.JSON(http.StatusOK, gin.H{
		"message":    "Аватар успешно обновлен",
		"avatar_url": avatarURL,
		"user":       updatedUser,
	})
}

// DeleteAvatarHandler удаляет текущий аватар пользователя
func (h *Handler) DeleteAvatarHandler(c *gin.Context) {
	user, ok := h.getUserFromContext(c)
	if !ok {
		return
	}

	if strings.HasPrefix(user.AvatarURL, "/auth/avatars/") {
		oldFilename := filepath.Base(user.AvatarURL)
		if oldFilename != "." && oldFilename != "/" {
			_ = os.Remove(filepath.Join(avatarStorageDir, oldFilename))
		}
	}

	updatedUser, err := h.authService.DeleteAvatar(user.ID)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to delete avatar from DB", logger.Err(err), slog.Int64("user_id", user.ID))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить аватар"})
		return
	}

	c.Set("user", *updatedUser)

	c.JSON(http.StatusOK, gin.H{
		"message": "Аватар успешно удален",
		"user":    updatedUser,
	})
}
