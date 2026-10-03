package security

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"auth-service/internal/domain"
)

func TestPasswordHashingAndVerification(t *testing.T) {
	password := "SecretP@ssw0rd123"
	hashed, err := HashPassword(password)
	require.NoError(t, err)
	require.NotEmpty(t, hashed)

	require.NoError(t, CheckPassword(hashed, password))
	require.Error(t, CheckPassword(hashed, "WrongPassword"))
}

func TestJWTTokensLifecycle(t *testing.T) {
	secret := "test-jwt-secret-very-long-and-secure"
	user := &domain.User{
		ID:    42,
		Email: "user@example.com",
		Role:  "admin",
		Name:  "Test Admin",
	}

	access, refresh, err := GenerateJWTTokens(user, secret)
	require.NoError(t, err)
	require.NotEmpty(t, access)
	require.NotEmpty(t, refresh)

	claims, err := ValidateJWTToken(access, secret)
	require.NoError(t, err)
	require.Equal(t, int64(42), claims.UserID)
	require.Equal(t, "user@example.com", claims.Email)
	require.Equal(t, "admin", claims.Role)
	require.Equal(t, "Test Admin", claims.Name)

	userID, err := ValidateRefreshToken(refresh, secret)
	require.NoError(t, err)
	require.Equal(t, int64(42), userID)

	_, err = ValidateJWTToken(access, "wrong-secret")
	require.Error(t, err)
}

func TestCSRFManager(t *testing.T) {
	mgr := NewCSRFManager()
	var userID int64 = 100

	tok, err := mgr.GenerateToken(&userID)
	require.NoError(t, err)
	require.NotEmpty(t, tok)

	require.True(t, mgr.ValidateToken(&userID, "", tok))
	require.False(t, mgr.ValidateToken(&userID, "", "invalid-token"))

	mgr.Invalidate(userID)
	require.False(t, mgr.ValidateToken(&userID, "", tok))
}

func TestRateLimiter(t *testing.T) {
	limiter := NewMemoryRateLimiter()
	key := "test-action"

	for i := 1; i <= 3; i++ {
		limited, err := limiter.IsLimited(key, 3, time.Minute)
		require.NoError(t, err)
		require.False(t, limited)
	}

	limited, err := limiter.IsLimited(key, 3, time.Minute)
	require.NoError(t, err)
	require.True(t, limited)
}

func TestSessionCookieValidation(t *testing.T) {
	secret := "this-is-a-very-long-secret-key-at-least-32-bytes"
	store, err := NewCookieSessionStore(secret)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	r, _ := http.NewRequest("GET", "/", nil)
	session, err := store.Get(r, "auth-session")
	require.NoError(t, err)

	session.Values["user_id"] = int64(77)
	session.Values["login_time"] = time.Now().Unix()
	require.NoError(t, store.Save(r, w, session))

	cookies := w.Result().Cookies()
	require.NotEmpty(t, cookies)

	cookieHeader := cookies[0].String()
	userID, err := ValidateSessionCookie(store, cookieHeader)
	require.NoError(t, err)
	require.Equal(t, int64(77), userID)
}
