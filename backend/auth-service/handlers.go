package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	"golang.org/x/crypto/bcrypt"
)

// googleAuth начинает процесс OAuth аутентификации через Google
func googleAuth(c *gin.Context) {
	// Проверяем, настроен ли Google OAuth
	if os.Getenv("GOOGLE_CLIENT_ID") == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "Google аутентификация не настроена"})
		return
	}

	// Начинаем процесс OAuth аутентификации
	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// googleAuthCallback обрабатывает callback от Google OAuth
func googleAuthCallback(c *gin.Context) {
	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Ошибка авторизации от %s: %v", c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"ошибка": "Не удалось завершить аутентификацию"})
		return
	}

	// Проверить, существует ли пользователь, создать, если нет
	var dbUser User
	result := db.Where("email = ?", user.Email).First(&dbUser)
	if result.Error != nil {
		// Создаём нового пользователя
		dbUser = User{
			Email:    user.Email,
			Name:     user.Name,
			Provider: "google",
			Role:     "operator",
		}
		if err := db.Create(&dbUser).Error; err != nil {
			log.Printf("БЕЗОПАСНОСТЬ: Не удалось создать пользователя %s: %v", user.Email, err)
			c.JSON(http.StatusInternalServerError, gin.H{"ошибка": "Не удалось создать пользователя"})
			return
		}
		log.Printf("БЕЗОПАСНОСТЬ: Новый пользователь создан через Google: %s от %s", user.Email, c.ClientIP())
	}

	// Сохранить пользователя в сеансе
	session, _ := store.Get(c.Request, "auth-session")
	session.Values["user_id"] = dbUser.ID
	session.Values["login_time"] = time.Now().Unix()

	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось сохранить сеанс: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"ошибка": "Не удалось создать сеанс"})
		return
	}

	log.Printf("SECURITY: Successful Google login for %s from %s", dbUser.Email, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": dbUser})
}

// userLogin обрабатывает вход пользователя с email и паролем
func userLogin(c *gin.Context) {
	// Ограничение скорости: 10 попыток в минуту на IP
	clientIP := c.ClientIP()
	if limited, _ := isRateLimited(clientIP+":user", 10, time.Minute); limited {
		log.Printf("БЕЗОПАСНОСТЬ: Превышен лимит скорости для входа пользователя от %s", clientIP)
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "Слишком много попыток входа. Попробуйте позже."})
		return
	}

	var loginReq struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&loginReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email и пароль обязательны"})
		return
	}

	// Базовая проверка ввода
	if len(loginReq.Email) == 0 || len(loginReq.Password) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Email и пароль не могут быть пустыми"})
		return
	}

	log.Printf("БЕЗОПАСНОСТЬ: Попытка входа пользователя для '%s' от %s", loginReq.Email, clientIP)

	// Найти пользователя по email или имени
	var user User
	if err := db.Where("email = ? OR name = ?", loginReq.Email, loginReq.Email).First(&user).Error; err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Неудачный вход - пользователь не найден: '%s' от %s", loginReq.Email, clientIP)
		// Добавить искусственную задержку для замедления перечисления пользователей
		time.Sleep(time.Second)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
		return
	}

	// Проверить, является ли пользователь локальным (имеет пароль)
	if user.Provider != "local" || user.Password == "" {
		log.Printf("БЕЗОПАСНОСТЬ: Попытка входа для нелокального пользователя: '%s' от %s", loginReq.Email, clientIP)
		time.Sleep(time.Second)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
		return
	}

	// Проверить пароль
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(loginReq.Password)); err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Неудачный вход - неправильный пароль для '%s' от %s", loginReq.Email, clientIP)
		time.Sleep(time.Second)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверные учетные данные"})
		return
	}

	// Сохранить пользователя в сессии
	session, _ := store.Get(c.Request, "auth-session")
	session.Values["user_id"] = user.ID
	session.Values["login_time"] = time.Now().Unix()

	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось сохранить сессию пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать сессию"})
		return
	}

	log.Printf("БЕЗОПАСНОСТЬ: Успешный вход для '%s' от %s", user.Email, clientIP)
	c.JSON(http.StatusOK, gin.H{"message": "Вход выполнен успешно", "user": user})
}

// createUser создает нового пользователя
func createUser(c *gin.Context) {
	var createReq struct {
		Email    string `json:"email" binding:"required,email"`
		Name     string `json:"name" binding:"required"`
		Initials string `json:"initials,omitempty"`
		INN      string `json:"inn,omitempty"`
		Password string `json:"password" binding:"required,min=8"`
		Role     string `json:"role" binding:"required,oneof=admin manager operator"`
	}

	if err := c.ShouldBindJSON(&createReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("БЕЗОПАСНОСТЬ: Попытка создания пользователя для '%s' (роль: %s) пользователем %s от %s",
		createReq.Email, createReq.Role, c.GetString("user_email"), c.ClientIP())

	// Проверить, существует ли пользователь уже
	var existingUser User
	if err := db.Where("email = ? OR name = ?", createReq.Email, createReq.Name).First(&existingUser).Error; err == nil {
		log.Printf("БЕЗОПАСНОСТЬ: Создание пользователя не удалось - пользователь уже существует: '%s'", createReq.Email)
		c.JSON(http.StatusConflict, gin.H{"error": "Пользователь с таким email или именем уже существует"})
		return
	}

	// Хэшировать пароль с более высокой стоимостью для лучшей безопасности
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(createReq.Password), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось хэшировать пароль: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя"})
		return
	}

	// Создать пользователя
	user := User{
		Email:    createReq.Email,
		Name:     createReq.Name,
		Initials: createReq.Initials,
		INN:      createReq.INN,
		Provider: "local",
		Role:     createReq.Role,
		Password: string(hashedPassword),
	}

	if err := db.Create(&user).Error; err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось создать пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось создать пользователя"})
		return
	}

	// Не возвращать пароль в ответе
	user.Password = ""

	log.Printf("БЕЗОПАСНОСТЬ: Пользователь создан успешно: '%s' (роль: %s)", user.Email, user.Role)
	c.JSON(http.StatusCreated, gin.H{"message": "Пользователь создан успешно", "user": user})
}

// logout обрабатывает выход пользователя
func logout(c *gin.Context) {
	// Получить информацию о пользователе для логирования перед очисткой сессии
	var userEmail string
	if user, exists := c.Get("user"); exists {
		userEmail = user.(User).Email
	}

	// Получить сессию
	session, err := store.Get(c.Request, "auth-session")
	if err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось получить сессию при выходе от %s: %v", c.ClientIP(), err)
		// Продолжить в любом случае, чтобы попытаться очистить cookie
	}

	// Получить user_id перед очисткой для очистки CSRF
	userID := session.Values["user_id"]

	// Очистить все значения сессии
	session.Values = make(map[interface{}]interface{})

	// Установить MaxAge в -1 для удаления cookie
	session.Options.MaxAge = -1

	// Сохранить сессию (это отправит заголовок Set-Cookie для его удаления)
	if err := session.Save(c.Request, c.Writer); err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось сохранить сессию при выходе от %s: %v", c.ClientIP(), err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Выход не удался"})
		return
	}

	// Очистить токен CSRF
	if userID != nil {
		csrfMutex.Lock()
		delete(csrfTokens, strconv.FormatUint(uint64(userID.(uint)), 10))
		csrfMutex.Unlock()
	}

	log.Printf("БЕЗОПАСНОСТЬ: Пользователь %s вышел от %s", userEmail, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Выход выполнен успешно"})
}

// getUsers возвращает список всех пользователей
func getUsers(c *gin.Context) {
	var users []User
	if err := db.Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить пользователей"})
		return
	}

	// Не возвращать пароли
	for i := range users {
		users[i].Password = ""
	}

	c.JSON(http.StatusOK, gin.H{"users": users})
}

// updateUser обновляет данные пользователя
func updateUser(c *gin.Context) {
	userID := c.Param("id")
	log.Printf("ОТЛАДКА: updateUser вызван для ID: %s от %s", userID, c.ClientIP())

	var updateReq struct {
		Name     string `json:"name,omitempty"`
		Initials string `json:"initials,omitempty"`
		INN      string `json:"inn,omitempty"`
		Role     string `json:"role,omitempty" binding:"oneof=admin manager operator"`
	}

	if err := c.ShouldBindJSON(&updateReq); err != nil {
		log.Printf("ОТЛАДКА: Ошибка привязки JSON в updateUser: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("ОТЛАДКА: Данные для обновления: name=%s, initials=%s, inn=%s, role=%s", updateReq.Name, updateReq.Initials, updateReq.INN, updateReq.Role)

	// Проверить, существует ли пользователь
	var user User
	if err := db.First(&user, userID).Error; err != nil {
		log.Printf("ОТЛАДКА: Пользователь с ID %s не найден: %v", userID, err)
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	log.Printf("ОТЛАДКА: Найден пользователь: ID=%d, Email=%s, Role=%s", user.ID, user.Email, user.Role)

	// Обновить поля, если они предоставлены
	updates := make(map[string]interface{})
	if updateReq.Name != "" {
		updates["name"] = updateReq.Name
	}
	if updateReq.Initials != "" {
		updates["initials"] = updateReq.Initials
	}
	if updateReq.INN != "" {
		updates["inn"] = updateReq.INN
	}
	if updateReq.Role != "" {
		updates["role"] = updateReq.Role
	}

	log.Printf("ОТЛАДКА: Поля для обновления: %v", updates)

	if len(updates) == 0 {
		log.Printf("ОТЛАДКА: Нет полей для обновления")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Нет полей для обновления"})
		return
	}

	if err := db.Model(&user).Updates(updates).Error; err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось обновить пользователя %s: %v", user.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось обновить пользователя"})
		return
	}

	log.Printf("ОТЛАДКА: Пользователь %s успешно обновлен", user.Email)

	// Получить обновленного пользователя
	if err := db.First(&user, userID).Error; err != nil {
		log.Printf("ОТЛАДКА: Не удалось получить обновленного пользователя: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось получить обновленного пользователя"})
		return
	}

	// Не возвращать пароль
	user.Password = ""

	log.Printf("БЕЗОПАСНОСТЬ: Пользователь %s обновлен от %s", user.Email, c.ClientIP())
	log.Printf("ОТЛАДКА: Возвращаю обновленного пользователя: ID=%d, Email=%s, Name=%s, Role=%s", user.ID, user.Email, user.Name, user.Role)
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь обновлен успешно", "user": user})
}

// deleteUser удаляет пользователя
func deleteUser(c *gin.Context) {
	userID := c.Param("id")

	// Проверить, существует ли пользователь и получить его информацию для логирования
	var userToDelete User
	if err := db.First(&userToDelete, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Пользователь не найден"})
		return
	}

	if err := db.Delete(&User{}, userID).Error; err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Не удалось удалить пользователя %s: %v", userToDelete.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось удалить пользователя"})
		return
	}

	log.Printf("БЕЗОПАСНОСТЬ: Пользователь %s удален от %s", userToDelete.Email, c.ClientIP())
	c.JSON(http.StatusOK, gin.H{"message": "Пользователь удален успешно"})
}

// getCurrentUser возвращает информацию о текущем пользователе
func getCurrentUser(c *gin.Context) {
	log.Printf("getCurrentUser: Origin: %s, Method: %s", c.GetHeader("Origin"), c.Request.Method)
	session, err := store.Get(c.Request, "auth-session")
	if err != nil {
		log.Printf("getCurrentUser: Session error: %v", err)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Неверный сеанс"})
		return
	}

	userID, ok := session.Values["user_id"]
	log.Printf("getCurrentUser: userID present: %v, value: %v", ok, userID)
	if !ok || userID == nil {
		log.Printf("getCurrentUser: userID missing or nil")
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Не аутентифицирован"})
		return
	}

	// Проверить возраст сеанса (опционально: принудительный пере-логин через X дней)
	if loginTime, ok := session.Values["login_time"].(int64); ok {
		if time.Now().Unix()-loginTime > 86400*30 { // 30 дней
			log.Printf("БЕЗОПАСНОСТЬ: Сеанс истек для пользователя ID %v", userID)
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Сеанс истек, пожалуйста, войдите снова"})
			return
		}
	}

	var user User
	if err := db.First(&user, userID).Error; err != nil {
		log.Printf("БЕЗОПАСНОСТЬ: Пользователь не найден для сеанса: %v", userID)
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Пользователь не найден"})
		return
	}

	// Не возвращать пароль
	user.Password = ""

	c.JSON(http.StatusOK, user)
}

// getCsrfToken возвращает токен CSRF для текущего пользователя
func getCsrfToken(c *gin.Context) {
	session, _ := store.Get(c.Request, "auth-session")
	userID, ok := session.Values["user_id"]

	var token string
	if ok && userID != nil {
		// Специфичный для пользователя токен для аутентифицированных пользователей
		userIDStr := strconv.FormatUint(uint64(userID.(uint)), 10)

		csrfMutex.Lock()
		csrfTokenEntry, exists := csrfTokens[userIDStr]

		// Сгенерировать новый токен, если не существует или истек
		if !exists || time.Now().After(csrfTokenEntry.expiresAt) {
			token = generateCsrfToken()
			csrfTokens[userIDStr] = csrfToken{
				token:     token,
				expiresAt: time.Now().Add(24 * time.Hour), // Токен действителен 24 часа
			}
		} else {
			token = csrfTokenEntry.token
		}
		csrfMutex.Unlock()
	} else {
		// Токен на основе сеанса для не аутентифицированных пользователей
		token = generateCsrfToken()
		session.Values["csrf_token"] = token
		if err := session.Save(c.Request, c.Writer); err != nil {
			log.Printf("БЕЗОПАСНОСТЬ: Не удалось сохранить токен CSRF в сеансе: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Не удалось сгенерировать токен CSRF"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"csrf_token": token})
}

// getServerStatus возвращает информацию о статусе сервера
func getServerStatus(c *gin.Context) {
	// Get database stats
	var userCount int64
	var partCount int64

	// Count users from auth service database
	db.Model(&User{}).Count(&userCount)

	// For parts count, we'd need to query the parts service database
	// For now, return a placeholder
	partCount = 0 // This should be queried from parts service

	status := gin.H{
		"server": gin.H{
			"status":     "running",
			"uptime":     "unknown", // Would need to track this
			"go_version": "1.21+",
			"os":         "windows",
			"arch":       "amd64",
		},
		"database": gin.H{
			"status":      "connected",
			"total_parts": partCount,
			"total_users": userCount,
		},
		"timestamp": time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, status)
}

// getServerLogs возвращает логи сервера (заглушка)
func getServerLogs(c *gin.Context) {
	// This is a placeholder - in a real implementation you'd read from log files
	// or have a logging system that stores logs in database
	logs := []gin.H{
		{
			"timestamp": time.Now().Add(-time.Hour).Format(time.RFC3339),
			"level":     "INFO",
			"message":   "Server started successfully",
		},
		{
			"timestamp": time.Now().Add(-30 * time.Minute).Format(time.RFC3339),
			"level":     "INFO",
			"message":   "Database connection established",
		},
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// getUserActivityLogs возвращает логи активности пользователей с фильтрами
func getUserActivityLogs(c *gin.Context) {
	userID := c.Query("user_id")
	action := c.Query("action")
	resourceType := c.Query("resource_type")
	limitStr := c.DefaultQuery("limit", "50")
	offsetStr := c.DefaultQuery("offset", "0")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	limit, _ := strconv.Atoi(limitStr)
	offset, _ := strconv.Atoi(offsetStr)

	query := db.Model(&UserActivityLog{}).Order("created_at DESC")

	if userID != "" {
		query = query.Where("user_id = ?", userID)
	}
	if action != "" {
		query = query.Where("action = ?", action)
	}
	if resourceType != "" {
		query = query.Where("resource_type = ?", resourceType)
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	var logs []UserActivityLog
	if err := query.Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		log.Printf("Failed to fetch user activity logs: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch activity logs"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"logs": logs})
}

// logUserActivity логирует активность пользователя
func logUserActivity(c *gin.Context) {
	var req struct {
		Action       string `json:"action" binding:"required"`
		ResourceType string `json:"resource_type" binding:"required"`
		ResourceID   *uint  `json:"resource_id,omitempty"`
		Details      string `json:"details,omitempty"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not authenticated"})
		return
	}

	userObj := user.(User)

	logEntry := UserActivityLog{
		UserID:       userObj.ID,
		UserName:     userObj.Name,
		UserEmail:    userObj.Email,
		Action:       req.Action,
		ResourceType: req.ResourceType,
		ResourceID:   req.ResourceID,
		Details:      req.Details,
		IPAddress:    c.ClientIP(),
		UserAgent:    c.GetHeader("User-Agent"),
	}

	if err := db.Create(&logEntry).Error; err != nil {
		log.Printf("Failed to log user activity: %v", err)
		// Don't return error to avoid breaking user flow
		c.JSON(http.StatusOK, gin.H{"message": "Activity logged"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Activity logged successfully"})
}
