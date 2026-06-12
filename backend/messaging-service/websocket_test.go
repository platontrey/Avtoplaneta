package main

import (
	"encoding/json"
	"testing"
	"time"
)

func TestWebSocketHubRegisterUnregister(t *testing.T) {
	// Start hub in background
	go globalHub.Run()

	// Wait for hub initialization
	time.Sleep(50 * time.Millisecond)

	// Create mock client
	client := &Client{
		UserID: 42,
		Conn:   nil, // nil is fine since we won't start read/write pumps in this test
		Send:   make(chan []byte, 10),
	}

	// Register client
	globalHub.register <- client
	time.Sleep(50 * time.Millisecond)

	globalHub.mu.RLock()
	conns, ok := globalHub.clients[42]
	globalHub.mu.RUnlock()

	if !ok || !conns[client] {
		t.Error("Client was not registered in the Hub")
	}

	// Unregister client
	globalHub.unregister <- client
	time.Sleep(50 * time.Millisecond)

	globalHub.mu.RLock()
	_, ok = globalHub.clients[42]
	globalHub.mu.RUnlock()

	if ok {
		t.Error("Client was not removed from the Hub")
	}
}

func TestWebSocketBroadcast(t *testing.T) {
	// Create client 1
	c1 := &Client{
		UserID: 101,
		Conn:   nil,
		Send:   make(chan []byte, 10),
	}
	// Create client 2
	c2 := &Client{
		UserID: 102,
		Conn:   nil,
		Send:   make(chan []byte, 10),
	}

	globalHub.register <- c1
	globalHub.register <- c2
	time.Sleep(50 * time.Millisecond)

	// Broadcast payload
	msg := BroadcastMessage{
		Participants: []int64{101, 102, 103}, // 103 is offline
		Payload: map[string]string{
			"message": "hello",
		},
	}

	globalHub.broadcast <- msg
	time.Sleep(50 * time.Millisecond)

	// Verify client 1 received the message
	select {
	case payloadBytes := <-c1.Send:
		var payload map[string]string
		err := json.Unmarshal(payloadBytes, &payload)
		if err != nil || payload["message"] != "hello" {
			t.Errorf("Client 1 did not receive the correct broadcast payload: %v", err)
		}
	default:
		t.Error("Client 1 did not receive the broadcast message")
	}

	// Verify client 2 received the message
	select {
	case payloadBytes := <-c2.Send:
		var payload map[string]string
		err := json.Unmarshal(payloadBytes, &payload)
		if err != nil || payload["message"] != "hello" {
			t.Errorf("Client 2 did not receive the correct broadcast payload: %v", err)
		}
	default:
		t.Error("Client 2 did not receive the broadcast message")
	}

	// Clean up
	globalHub.unregister <- c1
	globalHub.unregister <- c2
	time.Sleep(50 * time.Millisecond)
}
