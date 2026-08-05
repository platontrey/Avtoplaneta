package main

import (
	"encoding/json"
	"testing"
)

func TestUserUsesSharedNameContract(t *testing.T) {
	var user User
	if err := json.Unmarshal([]byte(`{"id":7,"name":"Иван","email":"ivan@example.com","role":"operator"}`), &user); err != nil {
		t.Fatalf("unmarshal user: %v", err)
	}

	if user.Name != "Иван" {
		t.Fatalf("expected shared name field, got %q", user.Name)
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal user: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		t.Fatalf("unmarshal payload: %v", err)
	}
	if payload["name"] != "Иван" {
		t.Fatalf("expected name in JSON, got %v", payload)
	}
	if _, exists := payload["username"]; exists {
		t.Fatalf("unexpected legacy username field: %v", payload)
	}
}
