package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"auth-service/internal/domain"
	"auth-service/internal/service"
)

func TestProfileHandlers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("UpdateProfileHandler_Success", func(t *testing.T) {
		mockService := new(service.MockAuthService)
		h := &Handler{authService: mockService}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		user := domain.User{ID: 10, Name: "Old Name", Username: "olduser"}
		c.Set("user", user)

		reqBody := domain.UpdateProfileRequest{
			Name:     "New Name",
			Username: "newuser",
			INN:      "1234567890",
		}
		jsonBytes, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("PUT", "/auth/profile", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")

		updatedUser := &domain.User{
			ID:       10,
			Name:     "New Name",
			Username: "newuser",
			INN:      "1234567890",
		}
		mockService.On("UpdateProfile", int64(10), reqBody).Return(updatedUser, nil)

		h.UpdateProfileHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp map[string]interface{}
		_ = json.Unmarshal(w.Body.Bytes(), &resp)
		assert.Equal(t, "Профиль успешно обновлен", resp["message"])
		mockService.AssertExpectations(t)
	})

	t.Run("ChangePasswordHandler_Success", func(t *testing.T) {
		mockService := new(service.MockAuthService)
		h := &Handler{authService: mockService}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		user := domain.User{ID: 10, Email: "test@example.com"}
		c.Set("user", user)

		reqBody := domain.ChangePasswordRequest{
			OldPassword: "oldpassword123",
			NewPassword: "newpassword123",
		}
		jsonBytes, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")

		mockService.On("ChangePassword", int64(10), "oldpassword123", "newpassword123").Return(nil)

		h.ChangePasswordHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("ChangePasswordHandler_WrongOldPassword", func(t *testing.T) {
		mockService := new(service.MockAuthService)
		h := &Handler{authService: mockService}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		user := domain.User{ID: 10, Email: "test@example.com"}
		c.Set("user", user)

		reqBody := domain.ChangePasswordRequest{
			OldPassword: "wrongpassword",
			NewPassword: "newpassword123",
		}
		jsonBytes, _ := json.Marshal(reqBody)
		c.Request = httptest.NewRequest("POST", "/auth/change-password", bytes.NewBuffer(jsonBytes))
		c.Request.Header.Set("Content-Type", "application/json")

		mockService.On("ChangePassword", int64(10), "wrongpassword", "newpassword123").Return(errors.New("неверный текущий пароль"))

		h.ChangePasswordHandler(c)

		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("DeleteAvatarHandler_Success", func(t *testing.T) {
		mockService := new(service.MockAuthService)
		h := &Handler{authService: mockService}

		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		user := domain.User{ID: 10, AvatarURL: "/auth/avatars/avatar_10_123.jpg"}
		c.Set("user", user)

		c.Request = httptest.NewRequest("DELETE", "/auth/avatar", nil)

		updatedUser := &domain.User{ID: 10, AvatarURL: ""}
		mockService.On("DeleteAvatar", int64(10)).Return(updatedUser, nil)

		h.DeleteAvatarHandler(c)

		assert.Equal(t, http.StatusOK, w.Code)
		mockService.AssertExpectations(t)
	})
}
