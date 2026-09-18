package main

import (
	"context"
	"time"

	authv1 "avtoplaneta/gen/auth/v1"
	"avtoplaneta/pkg/userdirectory"
)

// newUserDirectory собирает справочник имён поверх уже существующего gRPC-канала
// в auth-service — того же, по которому берут ИНН для выгрузки прайс-листа.
func newUserDirectory() *userdirectory.Directory {
	return userdirectory.New(func(ctx context.Context) (map[int64]string, error) {
		if authGRPCClient == nil {
			return nil, errGRPCNotInitialized()
		}

		ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		response, err := authGRPCClient.GetUsers(ctx, &authv1.GetUsersRequest{})
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
