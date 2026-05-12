package session

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManagerAcceptsPreviousSessionSecret(t *testing.T) {
	oldSecret := "old-secret-1234567890123456789012"
	newSecret := "new-secret-1234567890123456789012"

	oldManager := NewManager(oldSecret, false)
	loginReq := httptest.NewRequest(http.MethodGet, "/", nil)
	loginRec := httptest.NewRecorder()
	if err := oldManager.SetUserID(loginReq, loginRec, 42); err != nil {
		t.Fatalf("SetUserID old manager failed: %v", err)
	}

	cookies := loginRec.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatal("expected session cookie")
	}

	newManager := NewManagerWithPrevious(newSecret, oldSecret, false)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.AddCookie(cookies[0])

	if got := newManager.GetUserID(req); got != 42 {
		t.Fatalf("GetUserID with previous secret = %d, want 42", got)
	}

	rec := httptest.NewRecorder()
	if err := newManager.SetUserID(req, rec, 99); err != nil {
		t.Fatalf("SetUserID new manager failed: %v", err)
	}
	rotated := rec.Result().Cookies()
	if len(rotated) == 0 {
		t.Fatal("expected rotated session cookie")
	}

	rotatedReq := httptest.NewRequest(http.MethodGet, "/", nil)
	rotatedReq.AddCookie(rotated[0])
	if got := oldManager.GetUserID(rotatedReq); got != 0 {
		t.Fatalf("old manager decoded rotated cookie with user_id %d", got)
	}
}
