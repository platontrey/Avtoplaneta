// Package authcontext предоставляет легковесные и безопасные утилиты для
// извлечения контекста аутентифицированного пользователя из HTTP-заголовков,
// проставляемых Traefik ForwardAuth (auth-service).
package authcontext

import (
	"net/http"
	"strconv"
	"strings"
)

// User представляет данные пользователя, проброшенные через Traefik ForwardAuth
type User struct {
	ID    int64  `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
	Role  string `json:"role"`
}

// IsAuthenticated проверяет, валиден ли пользователь
func (u User) IsAuthenticated() bool {
	return u.ID > 0
}

// IsAdmin проверяет, является ли пользователь администратором
func (u User) IsAdmin() bool {
	return strings.EqualFold(u.Role, "admin")
}

// IsManagerOrAdmin проверяет, является ли пользователь менеджером или администратором
func (u User) IsManagerOrAdmin() bool {
	r := strings.ToLower(u.Role)
	return r == "admin" || r == "manager"
}

// FromRequest извлекает данные пользователя из HTTP-заголовков входящего запроса:
// X-User-ID, X-User-Email, X-User-Name, X-User-Role
func FromRequest(r *http.Request) (User, bool) {
	if r == nil {
		return User{}, false
	}
	return FromHeader(r.Header)
}

// FromHeader извлекает данные пользователя из http.Header
func FromHeader(h http.Header) (User, bool) {
	idStr := h.Get("X-User-ID")
	if idStr == "" {
		// fallback на X-User-Id
		idStr = h.Get("X-User-Id")
	}
	if idStr == "" {
		return User{}, false
	}

	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil || id <= 0 {
		return User{}, false
	}

	email := h.Get("X-User-Email")
	name := h.Get("X-User-Name")
	role := h.Get("X-User-Role")

	return User{
		ID:    id,
		Email: email,
		Name:  name,
		Role:  role,
	}, true
}
