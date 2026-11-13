package main

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"

	"avtoplaneta/pkg/logger"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for dev
	},
}

type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

type BroadcastMessage struct {
	Participants []int64     `json:"participants"`
	Payload      interface{} `json:"payload"`
}

type Hub struct {
	// Registered clients: map[userID]map[*Client]bool
	clients map[int64]map[*Client]bool

	broadcast  chan BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

var globalHub = Hub{
	clients:    make(map[int64]map[*Client]bool),
	broadcast:  make(chan BroadcastMessage, 1024),
	register:   make(chan *Client, 256),
	unregister: make(chan *Client, 256),
}

func (h *Hub) Run() {
	slog.Info("WebSocket Hub is running...")
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			slog.Info("User connected via WebSocket", slog.Int64("user_id", client.UserID))

		case client := <-h.unregister:
			h.mu.Lock()
			if conns, ok := h.clients[client.UserID]; ok {
				if _, ok := conns[client]; ok {
					delete(conns, client)
					close(client.Send)
					if len(conns) == 0 {
						delete(h.clients, client.UserID)
					}
				}
			}
			h.mu.Unlock()
			slog.Info("User disconnected from WebSocket", slog.Int64("user_id", client.UserID))

		case msg := <-h.broadcast:
			payloadBytes, err := json.Marshal(msg.Payload)
			if err != nil {
				slog.Error("Failed to marshal broadcast message payload", logger.Err(err))
				continue
			}

			h.mu.Lock()
			for _, userID := range msg.Participants {
				if conns, ok := h.clients[userID]; ok {
					for client := range conns {
						select {
						case client.Send <- payloadBytes:
						default:
							// Вместо ручного изменения мапы и закрытия канала в RLock
							// просто закрываем соединение. Это приведет к завершению ReadPump/WritePump
							// и автоматическому безопасному удалению клиента через канал unregister.
							client.Conn.Close()
						}
					}
				}
			}
			h.mu.Unlock()
		}
	}
}

func (c *Client) WritePump() {
	defer func() {
		c.Conn.Close()
	}()
	for {
		select {
		case message, ok := <-c.Send:
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)
			if err := w.Close(); err != nil {
				return
			}
		}
	}
}

func (c *Client) ReadPump() {
	defer func() {
		globalHub.unregister <- c
		c.Conn.Close()
	}()
	for {
		_, _, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func ServeWebSocket(c *gin.Context) {
	userIDStr := c.Query("user_id")
	if userIDStr == "" {
		if u, exists := c.Get("user_id"); exists {
			userIDStr = fmt.Sprintf("%v", u)
		}
	}
	if userIDStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User ID required"})
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		slog.ErrorContext(c.Request.Context(), "Failed to upgrade HTTP connection to WebSocket", logger.Err(err))
		return
	}

	client := &Client{
		UserID: userID,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	globalHub.register <- client

	go client.WritePump()
	go client.ReadPump()
}
