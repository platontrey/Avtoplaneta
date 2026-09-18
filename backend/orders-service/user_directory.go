package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"avtoplaneta/pkg/userdirectory"
)

// newUserDirectory собирает справочник имён поверх внутреннего эндпоинта
// auth-service. Отдельного gRPC-клиента здесь не заводим: orders-service уже
// ходит в auth по HTTP, а /internal/users существует ровно для таких запросов.
func newUserDirectory(authServiceURL string) *userdirectory.Directory {
	client := &http.Client{Timeout: 5 * time.Second}

	return userdirectory.New(func(ctx context.Context) (map[int64]string, error) {
		request, err := http.NewRequestWithContext(ctx, http.MethodGet, authServiceURL+"/internal/users", nil)
		if err != nil {
			return nil, err
		}

		response, err := client.Do(request)
		if err != nil {
			return nil, err
		}
		defer response.Body.Close()

		if response.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("auth-service ответил %s", response.Status)
		}

		var payload struct {
			Users []struct {
				ID   int64  `json:"id"`
				Name string `json:"name"`
			} `json:"users"`
		}
		if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
			return nil, err
		}

		names := make(map[int64]string, len(payload.Users))
		for _, user := range payload.Users {
			if user.Name != "" {
				names[user.ID] = user.Name
			}
		}
		return names, nil
	}, time.Minute)
}
