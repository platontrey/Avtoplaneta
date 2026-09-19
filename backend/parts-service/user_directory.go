package main

import (
	"context"
	"time"

	authv1 "avtoplaneta/gen/auth/v1"
	"avtoplaneta/pkg/userdirectory"
)

// newUserDirectory собирает справочник имён поверх уже существующего gRPC-канала
// в auth-service.
//
// Склад спрашивает у auth-service ровно одно — как сейчас зовут продавца, чтобы
// имя в карточке запчасти не расходилось с админкой. Ни ИНН, ни прочие
// реквизиты сюда не ходят: они нужны только в прайс-листе, и знает их теперь
// только export-service.
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
