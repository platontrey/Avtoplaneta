package authcontext

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFromRequest(t *testing.T) {
	t.Run("ValidHeaders", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-ID", "42")
		req.Header.Set("X-User-Email", "ivan@avtoplaneta.ru")
		req.Header.Set("X-User-Name", "Иван Иванов")
		req.Header.Set("X-User-Role", "admin")

		user, ok := FromRequest(req)
		if !ok {
			t.Fatal("expected ok=true, got false")
		}
		if user.ID != 42 {
			t.Errorf("expected ID=42, got %d", user.ID)
		}
		if user.Email != "ivan@avtoplaneta.ru" {
			t.Errorf("expected email=ivan@avtoplaneta.ru, got %s", user.Email)
		}
		if !user.IsAdmin() {
			t.Error("expected IsAdmin()=true")
		}
		if !user.IsManagerOrAdmin() {
			t.Error("expected IsManagerOrAdmin()=true")
		}
	})

	t.Run("MissingUserID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		_, ok := FromRequest(req)
		if ok {
			t.Fatal("expected ok=false for missing user ID")
		}
	})

	t.Run("InvalidUserID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-ID", "not-a-number")
		_, ok := FromRequest(req)
		if ok {
			t.Fatal("expected ok=false for invalid user ID")
		}
	})

	t.Run("ManagerRole", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Set("X-User-ID", "10")
		req.Header.Set("X-User-Role", "manager")

		user, ok := FromRequest(req)
		if !ok {
			t.Fatal("expected ok=true")
		}
		if user.IsAdmin() {
			t.Error("expected IsAdmin()=false for manager")
		}
		if !user.IsManagerOrAdmin() {
			t.Error("expected IsManagerOrAdmin()=true for manager")
		}
	})
}
