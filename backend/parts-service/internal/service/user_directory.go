package service

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	authv1 "avtoplaneta/gen/auth/v1"
	"avtoplaneta/pkg/userdirectory"
)

// NewUserDirectory собирает справочник имён поверх gRPC-клиента в auth-service.
func NewUserDirectory(authClient authv1.AuthServiceClient) *userdirectory.Directory {
	return userdirectory.New(func(ctx context.Context) (map[int64]string, error) {
		if authClient == nil {
			return nil, status.Error(codes.Unavailable, "auth gRPC client not initialized")
		}

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		response, err := authClient.GetUsers(ctx, &authv1.GetUsersRequest{})
		if err != nil {
			return nil, err
		}

		names := make(map[int64]string, len(response.Users))
		for _, user := range response.Users {
			if user.Name != "" {
				names[int64(user.Id)] = user.Name
			}
		}
		return names, nil
	}, time.Minute)
}
