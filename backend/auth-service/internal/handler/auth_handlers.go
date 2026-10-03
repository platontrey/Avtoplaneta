package handler

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

	"auth-service/internal/domain"
	"auth-service/internal/security"
)

var googleTokenInfoURL = "https://oauth2.googleapis.com/tokeninfo"

// GoogleAuthHandler инициирует поток веб-авторизации через Google OAuth
func (h *Handler) GoogleAuthHandler(c *gin.Context) {
	if os.Getenv("GOOGLE_CLIENT_ID") == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google аутентификация не настроена"})
		return
	}

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// GoogleAuthCallbackHandler обрабатывает возврат после Google OAuth
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

	if h.sessionStore != nil {
		session, _ := h.sessionStore.Get(c.Request, "auth-session")
		if session != nil {
			session.Values["user_id"] = createdUser.ID
			session.Values["login_time"] = time.Now().Unix()
			if err := h.sessionStore.Save(c.Request, c.Writer, session); err != nil {
				logrus.WithError(err).Error("Failed to save session")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сессию"})
				return
			}
		}
	}

	logrus.WithFields(logrus.Fields{
		"email": createdUser.Email,
		"ip":    c.ClientIP(),
	}).Info("Google login successful")

	c.JSON(http.StatusOK, createdUser)
}

// GoogleMobileAuthHandler обрабатывает авторизацию мобильного приложения по Google ID token
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

	if h.sessionStore != nil {
		session, _ := h.sessionStore.Get(c.Request, "auth-session")
		if session != nil {
			session.Values["user_id"] = createdUser.ID
			session.Values["login_time"] = time.Now().Unix()
			if err := h.sessionStore.Save(c.Request, c.Writer, session); err != nil {
				logrus.WithError(err).Error("Failed to save session (mobile)")
			}
		}
	}

	var secret string
	if h.config != nil {
		secret = h.config.JWTSecret
	}
	accessToken, refreshToken, err := security.GenerateJWTTokens(createdUser, secret)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate JWT tokens (mobile)")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать токены авторизации"})
		return
	}

	logrus.WithFields(logrus.Fields{
		"email": createdUser.Email,
		"ip":    c.ClientIP(),
	}).Info("Google mobile login successful")

	response := domain.LoginResponse{
		User:         *createdUser,
		Message:      "Login successful",
		Token:        accessToken,
		RefreshToken: refreshToken,
	}
	c.JSON(http.StatusOK, response)
}

// UserLoginHandler обрабатывает логин по email/username и паролю
func (h *Handler) UserLoginHandler(c *gin.Context) {
	var loginReq struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Логин (email или имя пользователя) и пароль обязательны"})
		return
	}

	loginReq.Email = strings.TrimSpace(loginReq.Email)
	if loginReq.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Логин не может быть пустым"})
		return
	}

	user, err := h.authService.AuthenticateUser(loginReq.Email, loginReq.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	if h.sessionStore != nil {
		session, _ := h.sessionStore.Get(c.Request, "auth-session")
		if session != nil {
			session.Values["user_id"] = user.ID
			session.Values["login_time"] = time.Now().Unix()
			if err := h.sessionStore.Save(c.Request, c.Writer, session); err != nil {
				logrus.WithError(err).Error("Failed to save session")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сессию"})
				return
			}
		}
	}

	var secret string
	if h.config != nil {
		secret = h.config.JWTSecret
	}
	accessToken, refreshToken, err := security.GenerateJWTTokens(user, secret)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate JWT tokens")
		accessToken = ""
		refreshToken = ""
	}

	response := domain.LoginResponse{
		User:         *user,
		Message:      "Login successful",
		Token:        accessToken,
		RefreshToken: refreshToken,
	}
	c.JSON(http.StatusOK, response)
}

// LogoutHandler завершает сессию пользователя
func (h *Handler) LogoutHandler(c *gin.Context) {
	var userEmail string
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(domain.User); ok {
			userEmail = u.Email
		}
	}

	var userID int64
	if h.sessionStore != nil {
		session, err := h.sessionStore.Get(c.Request, "auth-session")
		if err != nil {
			logrus.WithError(err).WithField("ip", c.ClientIP()).Warn("Failed to get session during logout")
		}
		if session != nil {
			if val, ok := session.Values["user_id"]; ok {
				userID, _ = security.GetUserIDFromSessionValue(val)
			}
			session.Values = make(map[interface{}]interface{})
			session.Options.MaxAge = -1
			if err := h.sessionStore.Save(c.Request, c.Writer, session); err != nil {
				logrus.WithError(err).WithField("ip", c.ClientIP()).Warn("Failed to save session during logout")
			}
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

// GetCurrentUserHandler возвращает данные текущего аутентифицированного пользователя
func (h *Handler) GetCurrentUserHandler(c *gin.Context) {
	c.Header("Cache-Control", "private, no-cache")

	authHeader := c.GetHeader("Authorization")
	if strings.HasPrefix(authHeader, "Bearer ") {
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		var secret string
		if h.config != nil {
			secret = h.config.JWTSecret
		}
		claims, err := security.ValidateJWTToken(tokenString, secret)
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

	if h.sessionStore == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
		return
	}

	session, err := h.sessionStore.Get(c.Request, "auth-session")
	if err != nil {
		logrus.WithError(err).Warn("Failed to get session")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сеанс"})
		return
	}

	userID, ok := security.GetUserIDFromSessionValue(session.Values["user_id"])
	if !ok {
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

// GetCSRFTokenHandler выдает или обновляет CSRF токен для клиента
func (h *Handler) GetCSRFTokenHandler(c *gin.Context) {
	var userID *int64

	if h.sessionStore != nil {
		session, _ := h.sessionStore.Get(c.Request, "auth-session")
		if session != nil {
			if val, ok := session.Values["user_id"]; ok && val != nil {
				if uid, ok := security.GetUserIDFromSessionValue(val); ok {
					userID = &uid
				}
			}
		}
	}

	token, err := h.authService.GenerateCSRFToken(userID)
	if err != nil {
		logrus.WithError(err).Error("Failed to generate CSRF token")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать токен CSRF"})
		return
	}

	if userID == nil && h.sessionStore != nil {
		session, _ := h.sessionStore.Get(c.Request, "auth-session")
		if session != nil {
			session.Values["csrf_token"] = token
			if err := h.sessionStore.Save(c.Request, c.Writer, session); err != nil {
				logrus.WithError(err).Error("Failed to save CSRF token in session")
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сохранить токен CSRF"})
				return
			}
		}
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, gin.H{"csrf_token": token})
}

// RefreshTokenHandler обновляет access/refresh пару по действующему refresh токену
func (h *Handler) RefreshTokenHandler(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh_token обязателен"})
		return
	}

	var secret string
	if h.config != nil {
		secret = h.config.JWTSecret
	}

	userID, err := security.ValidateRefreshToken(req.RefreshToken, secret)
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

	accessToken, refreshToken, err := security.GenerateJWTTokens(user, secret)
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

// VerifyAuthHandler используется Traefik ForwardAuth для верификации запросов на ingress шлюзе
func (h *Handler) VerifyAuthHandler(c *gin.Context) {
	val, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	user, ok := val.(domain.User)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	c.Header("X-User-ID", strconv.FormatInt(user.ID, 10))
	c.Header("X-User-Email", user.Email)
	c.Header("X-User-Name", user.Name)
	c.Header("X-User-Role", user.Role)
	c.Status(http.StatusOK)
}
