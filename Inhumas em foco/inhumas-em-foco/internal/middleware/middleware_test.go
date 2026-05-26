package middleware

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/model"
)

func TestClientIP(t *testing.T) {
	tests := []struct {
		name       string
		remoteAddr string
		headers    map[string]string
		want       string
	}{
		{
			name:       "strips remote port",
			remoteAddr: "203.0.113.10:4567",
			want:       "203.0.113.10",
		},
		{
			name:       "trusts forwarded for from local proxy",
			remoteAddr: "127.0.0.1:4567",
			headers:    map[string]string{"X-Forwarded-For": "198.51.100.9, 10.0.0.2"},
			want:       "198.51.100.9",
		},
		{
			name:       "ignores forwarded for from untrusted remote",
			remoteAddr: "203.0.113.10:4567",
			headers:    map[string]string{"X-Forwarded-For": "198.51.100.9"},
			want:       "203.0.113.10",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.RemoteAddr = tt.remoteAddr
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			if got := ClientIP(req); got != tt.want {
				t.Fatalf("ClientIP() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSecurityHeadersSetsHSTSBehindHTTPSProxy(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-Proto", "https")
	rec := httptest.NewRecorder()

	SecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rec, req)

	if rec.Header().Get("Strict-Transport-Security") == "" {
		t.Fatal("expected HSTS header behind HTTPS proxy")
	}
}

func TestSecurityHeadersWithConfigCanEmitReportOnlyCSP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	SecurityHeadersWithConfig(&config.Config{CSPReportOnly: true, CSPReportURI: "/csp-report"})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {})).ServeHTTP(rec, req)

	enforced := rec.Header().Get("Content-Security-Policy")
	if !strings.Contains(enforced, "'unsafe-inline'") || !strings.Contains(enforced, "object-src 'none'") {
		t.Fatalf("unexpected enforced CSP: %s", enforced)
	}
	reportOnly := rec.Header().Get("Content-Security-Policy-Report-Only")
	if strings.Contains(reportOnly, "'unsafe-inline'") || !strings.Contains(reportOnly, "report-uri /csp-report") {
		t.Fatalf("unexpected report-only CSP: %s", reportOnly)
	}
}

func TestStructuredLoggerWritesJSONAndRequestID(t *testing.T) {
	var logs bytes.Buffer
	restore := SetStructuredLogWriter(&logs)
	defer restore()

	req := httptest.NewRequest(http.MethodPost, "/painel/posts?status=draft", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	req.Header.Set("User-Agent", "cms-test")
	rec := httptest.NewRecorder()

	StructuredLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if RequestID(r.Context()) == "" {
			t.Fatal("expected request id in context")
		}
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte("ok"))
	})).ServeHTTP(rec, req)

	if rec.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected X-Request-ID header")
	}
	var entry map[string]any
	if err := json.Unmarshal(bytes.TrimSpace(logs.Bytes()), &entry); err != nil {
		t.Fatalf("invalid json log: %v", err)
	}
	if entry["event"] != "http_request" || entry["method"] != http.MethodPost || entry["path"] != "/painel/posts" {
		t.Fatalf("unexpected log entry: %#v", entry)
	}
	if entry["status"].(float64) != http.StatusCreated {
		t.Fatalf("status = %v, want %d", entry["status"], http.StatusCreated)
	}
}

func TestCSRFProtection(t *testing.T) {
	t.Run("rejects missing token on login post", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("email=a"))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		CSRFProtection("/painel", false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})

	t.Run("accepts valid token header", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/painel/posts", nil)
		req.AddCookie(&http.Cookie{Name: "inhumas_csrf", Value: "token"})
		req.Header.Set("X-CSRF-Token", "token")
		rec := httptest.NewRecorder()

		CSRFProtection("/painel", false)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})).ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	})

	t.Run("rejects multipart token lookup above body limit", func(t *testing.T) {
		var body bytes.Buffer
		writer := multipart.NewWriter(&body)
		_ = writer.WriteField("csrf_token", "token")
		part, err := writer.CreateFormFile("cover", "large.jpg")
		if err != nil {
			t.Fatal(err)
		}
		part.Write(bytes.Repeat([]byte("x"), 256))
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}

		req := httptest.NewRequest(http.MethodPost, "/painel/posts", &body)
		req.AddCookie(&http.Cookie{Name: "inhumas_csrf", Value: "token"})
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rec := httptest.NewRecorder()

		CSRFProtection("/painel", false, 64)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		})).ServeHTTP(rec, req)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusForbidden)
		}
	})
}

func TestRequireAdminRBAC(t *testing.T) {
	tests := []struct {
		name string
		user *model.User
		path string
		want int
	}{
		{name: "anonymous redirects", path: "/painel/posts", want: http.StatusSeeOther},
		{name: "super admin can access users", user: &model.User{Role: model.RoleSuperAdmin}, path: "/painel/users", want: http.StatusNoContent},
		{name: "editor can access posts", user: &model.User{Role: model.RoleEditor}, path: "/painel/posts", want: http.StatusNoContent},
		{name: "editor cannot access stores", user: &model.User{Role: model.RoleEditor}, path: "/painel/stores", want: http.StatusForbidden},
		{name: "redator can access posts", user: &model.User{Role: model.RoleRedator}, path: "/painel/posts", want: http.StatusNoContent},
		{name: "redator cannot access users", user: &model.User{Role: model.RoleRedator}, path: "/painel/users", want: http.StatusForbidden},
		{name: "revisor can access posts", user: &model.User{Role: model.RoleRevisor}, path: "/painel/posts", want: http.StatusNoContent},
		{name: "comercial can access stores", user: &model.User{Role: model.RoleComercial}, path: "/painel/stores", want: http.StatusNoContent},
		{name: "comercial can access influencers", user: &model.User{Role: model.RoleComercial}, path: "/painel/influencers", want: http.StatusNoContent},
		{name: "comercial cannot access users", user: &model.User{Role: model.RoleComercial}, path: "/painel/users", want: http.StatusForbidden},
		{name: "admin can access users", user: &model.User{Role: model.RoleAdmin}, path: "/painel/users", want: http.StatusNoContent},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, tt.path, nil)
			if tt.user != nil {
				req = req.WithContext(auth.WithUser(req.Context(), tt.user))
			}
			rec := httptest.NewRecorder()

			RequireAdmin("/painel")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusNoContent)
			})).ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("status = %d, want %d", rec.Code, tt.want)
			}
		})
	}
}

func TestRequireAdminRBACMatrixByRoleAndSection(t *testing.T) {
	sections := []struct {
		name string
		path string
	}{
		{name: "dashboard", path: "/painel"},
		{name: "posts", path: "/painel/posts"},
		{name: "automation", path: "/painel/automation"},
		{name: "categories", path: "/painel/categories"},
		{name: "tags", path: "/painel/tags"},
		{name: "media", path: "/painel/media"},
		{name: "stores", path: "/painel/stores"},
		{name: "influencers", path: "/painel/influencers"},
		{name: "banners", path: "/painel/banners"},
		{name: "promotions", path: "/painel/promotions"},
		{name: "events", path: "/painel/events"},
		{name: "classifieds", path: "/painel/classifieds"},
		{name: "neighborhoods", path: "/painel/neighborhoods"},
		{name: "users", path: "/painel/users"},
		{name: "settings", path: "/painel/settings"},
	}
	tests := []struct {
		role    model.UserRole
		allowed map[string]bool
	}{
		{role: model.RoleSuperAdmin, allowed: allowAll(sections)},
		{role: model.RoleAdmin, allowed: allowAll(sections)},
		{role: model.RoleEditor, allowed: allowOnly("dashboard", "posts", "automation")},
		{role: model.RoleRedator, allowed: allowOnly("dashboard", "posts")},
		{role: model.RoleRevisor, allowed: allowOnly("dashboard", "posts")},
		{role: model.RoleComercial, allowed: allowOnly("dashboard", "stores", "influencers", "banners", "promotions")},
	}

	for _, tt := range tests {
		for _, section := range sections {
			t.Run(string(tt.role)+"_"+section.name, func(t *testing.T) {
				req := httptest.NewRequest(http.MethodGet, section.path, nil)
				req = req.WithContext(auth.WithUser(req.Context(), &model.User{Role: tt.role}))
				rec := httptest.NewRecorder()

				RequireAdmin("/painel")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})).ServeHTTP(rec, req)

				want := http.StatusForbidden
				if tt.allowed[section.name] {
					want = http.StatusNoContent
				}
				if rec.Code != want {
					t.Fatalf("role %s section %s status = %d, want %d", tt.role, section.name, rec.Code, want)
				}
			})
		}
	}
}

func allowAll(sections []struct {
	name string
	path string
}) map[string]bool {
	allowed := make(map[string]bool, len(sections))
	for _, section := range sections {
		allowed[section.name] = true
	}
	return allowed
}

func allowOnly(names ...string) map[string]bool {
	allowed := make(map[string]bool, len(names))
	for _, name := range names {
		allowed[name] = true
	}
	return allowed
}

func TestRateLimitByIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	handler := RateLimitByIP(1, time.Minute)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, req)
	if first.Code != http.StatusNoContent {
		t.Fatalf("first status = %d, want %d", first.Code, http.StatusNoContent)
	}

	second := httptest.NewRecorder()
	handler.ServeHTTP(second, req)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want %d", second.Code, http.StatusTooManyRequests)
	}
}

func TestRateLimitByIPWhenSkipsNonMatchingRequests(t *testing.T) {
	handler := RateLimitByIPWhen(1, time.Minute, func(r *http.Request) bool {
		return r.URL.Path == "/login"
	})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = "203.0.113.10:1234"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		if rec.Code != http.StatusNoContent {
			t.Fatalf("non matching status = %d, want %d", rec.Code, http.StatusNoContent)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/login", nil)
	req.RemoteAddr = "203.0.113.10:1234"
	first := httptest.NewRecorder()
	handler.ServeHTTP(first, req)
	if first.Code != http.StatusNoContent {
		t.Fatalf("first matching status = %d, want %d", first.Code, http.StatusNoContent)
	}
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, req)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second matching status = %d, want %d", second.Code, http.StatusTooManyRequests)
	}
}
