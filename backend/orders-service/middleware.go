package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

// CORSMiddleware добавляет CORS заголовки для кросс-доменных запросов
func CORSMiddleware() gin.HandlerFunc {
	return gin.HandlerFunc(func(c *gin.Context) {
		log.Printf("CORS: Получен %s запрос к %s от %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())
		c.Header("Access-Control-Allow-Origin", "http://localhost:5173")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")

		if c.Request.Method == "OPTIONS" {
			log.Printf("CORS: Обработка предварительного OPTIONS запроса к %s", c.Request.URL.Path)
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	})
}