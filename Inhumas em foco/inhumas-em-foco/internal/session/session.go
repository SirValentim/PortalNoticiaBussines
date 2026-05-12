package session

import (
	"net/http"

	"github.com/gorilla/sessions"
)

type Manager struct {
	store *sessions.CookieStore
	name  string
}

func NewManager(secret string, secure bool) *Manager {
	return NewManagerWithPrevious(secret, "", secure)
}

func NewManagerWithPrevious(secret, previousSecret string, secure bool) *Manager {
	keyPairs := [][]byte{[]byte(secret), nil}
	if previousSecret != "" {
		keyPairs = append(keyPairs, []byte(previousSecret), nil)
	}
	store := sessions.NewCookieStore(keyPairs...)
	store.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 7 days
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
	}
	return &Manager{
		store: store,
		name:  "inhumas_session",
	}
}

func (m *Manager) Get(r *http.Request) (*sessions.Session, error) {
	return m.store.Get(r, m.name)
}

func (m *Manager) Save(r *http.Request, w http.ResponseWriter, session *sessions.Session) error {
	return m.store.Save(r, w, session)
}

func (m *Manager) GetUserID(r *http.Request) int64 {
	session, _ := m.Get(r)
	if val, ok := session.Values["user_id"]; ok {
		switch id := val.(type) {
		case int64:
			return id
		case int:
			return int64(id)
		case int32:
			return int64(id)
		case float64:
			return int64(id)
		}
	}
	return 0
}

func (m *Manager) SetUserID(r *http.Request, w http.ResponseWriter, userID int64) error {
	session, _ := m.Get(r)
	session.Values["user_id"] = userID
	return m.Save(r, w, session)
}

func (m *Manager) Clear(r *http.Request, w http.ResponseWriter) error {
	session, _ := m.Get(r)
	delete(session.Values, "user_id")
	return m.Save(r, w, session)
}
