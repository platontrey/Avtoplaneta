package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	authv1 "avtoplaneta/gen/auth/v1"
	"avtoplaneta/pkg/grpcclient"
)

var (
	authGRPCClient authv1.AuthServiceClient
	authClientOnce sync.Once
)

func getAuthGRPCClient() authv1.AuthServiceClient {
	authClientOnce.Do(func() {
		addr := os.Getenv("AUTH_GRPC_ADDR")
		if addr == "" {
			if u := os.Getenv("AUTH_SERVICE_GRPC_URL"); u != "" {
				addr = u
			} else {
				addr = "auth-service:9083"
			}
		}
		conn, err := grpcclient.NewClient(addr)
		if err != nil {
			log.Printf("Messaging: Failed to create gRPC connection to auth-service: %v", err)
			return
		}
		authGRPCClient = authv1.NewAuthServiceClient(conn)
		log.Printf("Messaging: Connected to auth-service via gRPC at %s", addr)
	})
	return authGRPCClient
}

func getAuthServiceURL() string {
	if url := os.Getenv("AUTH_SERVICE_URL"); url != "" {
		return url
	}
	if url := os.Getenv("GATEWAY_URL"); url != "" {
		return url
	}
	return "http://auth-service:8083"
}

func getUsers(c *gin.Context) {
	if client := getAuthGRPCClient(); client != nil {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
		defer cancel()

		resp, err := client.GetUsers(ctx, &authv1.GetUsersRequest{})
		if err == nil {
			users := make([]User, len(resp.Users))
			for i, u := range resp.Users {
				users[i] = User{
					ID:    int64(u.Id),
					Name:  u.Name,
					Email: u.Email,
					Role:  u.Role,
				}
			}
			c.JSON(http.StatusOK, gin.H{"users": users})
			return
		}
		log.Printf("Messaging getUsers: gRPC call failed, attempting HTTP fallback: %v", err)
	}

	authURL := getAuthServiceURL()
	req, err := http.NewRequestWithContext(c.Request.Context(), "GET", authURL+"/api/users", nil)
	if err != nil {
		log.Printf("Messaging getUsers: Failed to create request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
		return
	}

	authHeader := c.GetHeader("Authorization")
	userIDHeader := c.GetHeader("X-User-ID")
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("X-User-ID", userIDHeader)

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Messaging getUsers: Failed to connect to auth-service: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to auth-service"})
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Messaging getUsers: Auth service returned error status: %d", resp.StatusCode)
		c.JSON(resp.StatusCode, gin.H{"error": "Failed to fetch users"})
		return
	}

	var response struct {
		Users []User `json:"users"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		log.Printf("Messaging getUsers: Failed to parse response: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse response"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"users": response.Users})
}
