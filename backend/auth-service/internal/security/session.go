package security

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

// SessionStore интерфейс для работы с HTTP сессиями
type SessionStore interface {
	Get(r *http.Request, name string) (*sessions.Session, error)
	Save(r *http.Request, w http.ResponseWriter, s *sessions.Session) error
}

// CookieSessionStore реализация SessionStore на базе gorilla/sessions
type CookieSessionStore struct {
	Store *sessions.CookieStore
}

// NewCookieSessionStore создает новое хранилище сессий на cookies
func NewCookieSessionStore(sessionSecret string) (*CookieSessionStore, error) {
	if len(sessionSecret) < 32 {
		return nil, fmt.Errorf("SESSION_SECRET должен быть не менее 32 символов")
	}

	cs := sessions.NewCookieStore([]byte(sessionSecret))
	cs.Options = &sessions.Options{
		Path:     "/",
		Domain:   "",
		MaxAge:   86400 * 7,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	}

	// Устанавливаем для gothic (Google OAuth)
	gothic.Store = cs

	return &CookieSessionStore{Store: cs}, nil
}

func (s *CookieSessionStore) Get(r *http.Request, name string) (*sessions.Session, error) {
	return s.Store.Get(r, name)
}

func (s *CookieSessionStore) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return s.Store.Save(r, w, session)
}

// GetUserIDFromSessionValue извлекает int64 ID пользователя из значения сессии
func GetUserIDFromSessionValue(v interface{}) (int64, bool) {
	switch id := v.(type) {
	case int64:
		return id, true
	case int:
		return int64(id), true
	case uint:
		return int64(id), true
	case float64:
		return int64(id), true
	}
	return 0, false
}

// ValidateSessionCookie валидирует сырой заголовок Cookie для gRPC запросов
func ValidateSessionCookie(store SessionStore, cookieHeader string) (int64, error) {
	fakeReq, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		return 0, err
	}
	fakeReq.Header.Set("Cookie", cookieHeader)

	session, err := store.Get(fakeReq, "auth-session")
	if err != nil {
		return 0, fmt.Errorf("недействительная сессия: %w", err)
	}

	val, ok := session.Values["user_id"]
	if !ok || val == nil {
		return 0, fmt.Errorf("user_id отсутствует в сессии")
	}

	userID, ok := GetUserIDFromSessionValue(val)
	if !ok {
		return 0, fmt.Errorf("неверный тип user_id в сессии")
	}

	if loginTime, ok := session.Values["login_time"].(int64); ok {
		if time.Now().Unix()-loginTime > 86400*30 {
			return 0, fmt.Errorf("сессия истекла")
		}
	}

	return userID, nil
}

// InitGoogleOAuth регистрирует провайдер Google в goth
func InitGoogleOAuth(callbackURL string) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	clientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if callbackURL == "" {
		callbackURL = os.Getenv("GOOGLE_CALLBACK_URL")
	}
	if callbackURL == "" {
		callbackURL = "http://localhost:8082/auth/google/callback"
	}

	if clientID != "" && clientSecret != "" {
		goth.UseProviders(
			google.New(clientID, clientSecret, callbackURL),
		)
	} else {
		log.Println("ПРЕДУПРЕЖДЕНИЕ: Google OAuth не настроен (отсутствует GOOGLE_CLIENT_ID или GOOGLE_CLIENT_SECRET)")
	}
}
