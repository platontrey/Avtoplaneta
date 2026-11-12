package handler

import (
	"github.com/gin-gonic/gin"

	"avtoplaneta/pkg/ginmw"
)

// CORSMiddleware делегирует настройку CORS в единый пакет ginmw.CORS
func CORSMiddleware() gin.HandlerFunc {
	return ginmw.CORS()
}

// AuthMiddleware проверяет аутентификацию пользователя через единый пакет ginmw.RequireAuth
func AuthMiddleware() gin.HandlerFunc {
	return ginmw.RequireAuth()
}

// AdminMiddleware проверяет права администратора пользователя через единый пакет ginmw.RequireAdmin
func AdminMiddleware() gin.HandlerFunc {
	return ginmw.RequireAdmin()
}
