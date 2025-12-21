package main

import (
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize database
	InitDatabase()

	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Routes will be added here
	setupRoutes(r)

	log.Println("Messaging service starting on port 8084")
	r.Run(":8084")
}