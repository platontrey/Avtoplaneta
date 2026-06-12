package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for testing
	},
}

// Client represents a connected WebSocket user
type Client struct {
	UserID int64
	Conn   *websocket.Conn
	Send   chan []byte
}

// Hub manages active connections and broadcasts messages
type Hub struct {
	clients    map[int64]map[*Client]bool
	broadcast  chan BroadcastMessage
	register   chan *Client
	unregister chan *Client
	mu         sync.RWMutex
}

type BroadcastMessage struct {
	Participants []int64     `json:"participants"`
	Payload      interface{} `json:"payload"`
}

var globalHub = &Hub{
	clients:    make(map[int64]map[*Client]bool),
	broadcast:  make(chan BroadcastMessage, 1024),
	register:   make(chan *Client, 256),
	unregister: make(chan *Client, 256),
}

func (h *Hub) Run() {
	logrus.Info("WebSocket Hub is running...")
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			if _, ok := h.clients[client.UserID]; !ok {
				h.clients[client.UserID] = make(map[*Client]bool)
			}
			h.clients[client.UserID][client] = true
			h.mu.Unlock()
			logrus.Infof("User %d connected via WebSocket", client.UserID)

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
			logrus.Infof("User %d disconnected from WebSocket", client.UserID)

		case msg := <-h.broadcast:
			payloadBytes, err := json.Marshal(msg.Payload)
			if err != nil {
				logrus.WithError(err).Error("Failed to marshal broadcast message payload")
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

// ServeWebSocket handles websocket requests from clients
func ServeWebSocket(c *gin.Context) {
	userIDStr := c.Query("userId")
	if userIDStr == "" {
		userIDStr = c.GetHeader("X-User-ID")
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
		logrus.WithError(err).Error("Failed to upgrade HTTP connection to WebSocket")
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
