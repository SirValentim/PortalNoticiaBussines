package middleware

import (
	"context"
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
	"inhumas-em-foco/internal/session"
)

var (
	startedAt           = time.Now()
	requestsTotal       uint64
	structuredLogWriter io.Writer = os.Stdout
)

type requestContextKey string

const requestIDKey requestContextKey = "request_id"

type OperationalMetrics struct {
	StartedAt     time.Time `json:"started_at"`
	UptimeSeconds int64     `json:"uptime_seconds"`
	RequestsTotal uint64    `json:"requests_total"`
}

func MetricsSnapshot() OperationalMetrics {
	return OperationalMetrics{
		StartedAt:     startedAt,
		UptimeSeconds: int64(time.Since(startedAt).Seconds()),
		RequestsTotal: atomic.LoadUint64(&requestsTotal),
	}
}

func SecurityHeaders(next http.Handler) http.Handler {
	return SecurityHeadersWithConfig(nil)(next)
}

func SecurityHeadersWithConfig(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-Frame-Options", "SAMEORIGIN")
			w.Header().Set("X-Content-Type-Options", "nosniff")
			w.Header().Set("X-XSS-Protection", "1; mode=block")
			w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
			w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=(), payment=()")
			w.Header().Set("Content-Security-Policy", cspHeader(true, isHTTPSRequest(r), ""))
			if cfg != nil && cfg.CSPReportOnly {
				w.Header().Set("Content-Security-Policy-Report-Only", cspHeader(false, false, cfg.CSPReportURI))
			}
			if isHTTPSRequest(r) {
				w.Header().Set("Strict-Transport-Security", "max-age=63072000; includeSubDomains; preload")
			}
			next.ServeHTTP(w, r)
		})
	}
}

func cspHeader(allowUnsafeInline bool, upgradeInsecure bool, reportURI string) string {
	scriptSrc := "script-src 'self'"
	styleSrc := "style-src 'self'"
	if allowUnsafeInline {
		scriptSrc += " 'unsafe-inline'"
		styleSrc += " 'unsafe-inline'"
	}
	directives := []string{
		"default-src 'self'",
		scriptSrc,
		styleSrc,
		"img-src 'self' data: https:",
		"font-src 'self'",
		"connect-src 'self'",
		"object-src 'none'",
		"frame-ancestors 'self'",
		"base-uri 'self'",
		"form-action 'self'",
	}
	if upgradeInsecure {
		directives = append(directives, "upgrade-insecure-requests")
	}
	if reportURI != "" {
		directives = append(directives, "report-uri "+reportURI)
	}
	return strings.Join(directives, "; ")
}

func isHTTPSRequest(r *http.Request) bool {
	return r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https")
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	bytes  int
}

func (r *responseRecorder) WriteHeader(status int) {
	if r.status != 0 {
		return
	}
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *responseRecorder) Write(body []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	n, err := r.ResponseWriter.Write(body)
	r.bytes += n
	return n, err
}

func (r *responseRecorder) Flush() {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if !ok {
		return
	}
	flusher.Flush()
}

func StructuredLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		requestID := strings.TrimSpace(r.Header.Get("X-Request-ID"))
		if requestID == "" {
			requestID = newRequestID()
		}
		w.Header().Set("X-Request-ID", requestID)

		rec := &responseRecorder{ResponseWriter: w}
		ctx := context.WithValue(r.Context(), requestIDKey, requestID)
		next.ServeHTTP(rec, r.WithContext(ctx))
		if rec.status == 0 {
			rec.status = http.StatusOK
		}

		level := "info"
		if rec.status >= 500 {
			level = "error"
		} else if rec.status >= 400 {
			level = "warn"
		}
		entry := map[string]any{
			"ts":          time.Now().UTC().Format(time.RFC3339Nano),
			"level":       level,
			"event":       "http_request",
			"request_id":  requestID,
			"method":      r.Method,
			"path":        r.URL.Path,
			"query":       r.URL.RawQuery,
			"status":      rec.status,
			"bytes":       rec.bytes,
			"duration_ms": time.Since(started).Milliseconds(),
			"ip":          ClientIP(r),
			"user_agent":  r.UserAgent(),
		}
		if data, err := json.Marshal(entry); err == nil {
			fmt.Fprintln(structuredLogWriter, string(data))
		}
	})
}

func RequestID(ctx context.Context) string {
	if requestID, ok := ctx.Value(requestIDKey).(string); ok {
		return requestID
	}
	return ""
}

func SetStructuredLogWriter(w io.Writer) func() {
	previous := structuredLogWriter
	if w == nil {
		structuredLogWriter = io.Discard
	} else {
		structuredLogWriter = w
	}
	return func() {
		structuredLogWriter = previous
	}
}

func ClientIP(r *http.Request) string {
	remote := hostOnly(r.RemoteAddr)
	remoteIP := net.ParseIP(remote)
	if remoteIP == nil {
		return remote
	}

	if isTrustedProxy(remoteIP) {
		forwardedFor := firstHeaderIP(r.Header.Get("X-Forwarded-For"))
		if forwardedFor != "" {
			return forwardedFor
		}
		realIP := firstHeaderIP(r.Header.Get("X-Real-IP"))
		if realIP != "" {
			return realIP
		}
		cfIP := firstHeaderIP(r.Header.Get("CF-Connecting-IP"))
		if cfIP != "" {
			return cfIP
		}
	}

	return remoteIP.String()
}

func hostOnly(addr string) string {
	host, _, err := net.SplitHostPort(addr)
	if err == nil {
		return host
	}
	return strings.Trim(addr, "[]")
}

func firstHeaderIP(header string) string {
	for _, part := range strings.Split(header, ",") {
		ip := net.ParseIP(strings.TrimSpace(part))
		if ip != nil {
			return ip.String()
		}
	}
	return ""
}

func isTrustedProxy(ip net.IP) bool {
	return ip.IsLoopback() || ip.IsPrivate()
}

func RequestTimeout(timeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), timeout)
			defer cancel()
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(w, "Erro interno")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

func AuthMiddleware(sessionMgr *session.Manager, repo *repository.Repository, authSvc *auth.Service) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := sessionMgr.GetUserID(r)
			if userID > 0 {
				user, err := repo.UserGetByID(r.Context(), userID)
				if err == nil && user != nil && user.Active {
					ctx := auth.WithUser(r.Context(), user)
					r = r.WithContext(ctx)
				}
			}
			next.ServeHTTP(w, r)
		})
	}
}

func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := auth.UserFromContext(r.Context())
		if user == nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func RequireAdmin(adminPrefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !strings.HasPrefix(r.URL.Path, adminPrefix) {
				next.ServeHTTP(w, r)
				return
			}
			user := auth.UserFromContext(r.Context())
			if user == nil {
				http.Redirect(w, r, "/login", http.StatusSeeOther)
				return
			}
			if !user.Role.IsValid() {
				http.Error(w, "Acesso negado", http.StatusForbidden)
				return
			}
			if !adminRouteAllowed(user.Role, r.URL.Path, adminPrefix) {
				http.Error(w, "Acesso negado", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func adminRouteAllowed(role model.UserRole, path, adminPrefix string) bool {
	if role == model.RoleSuperAdmin || role == model.RoleAdmin {
		return true
	}
	section := strings.TrimPrefix(path, adminPrefix)
	if section == "" || section == "/" {
		return true
	}
	if strings.HasPrefix(section, "/profile") {
		return true
	}
	switch role {
	case model.RoleEditor, model.RoleRedator, model.RoleRevisor:
		if role == model.RoleEditor && strings.HasPrefix(section, "/automation") {
			return true
		}
		return strings.HasPrefix(section, "/posts")
	case model.RoleComercial:
		return strings.HasPrefix(section, "/stores") ||
			strings.HasPrefix(section, "/influencers") ||
			strings.HasPrefix(section, "/banners") ||
			strings.HasPrefix(section, "/promotions")
	default:
		return false
	}
}

type csrfContextKey string

const csrfTokenKey csrfContextKey = "csrf_token"

func CSRFProtection(adminPrefix string, secure bool, maxBodyBytes ...int64) func(http.Handler) http.Handler {
	bodyLimit := int64(2 * 1024 * 1024)
	if len(maxBodyBytes) > 0 && maxBodyBytes[0] > 0 {
		bodyLimit = maxBodyBytes[0]
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := csrfTokenFromCookie(r)
			if token == "" {
				token = newCSRFToken()
				http.SetCookie(w, &http.Cookie{
					Name:     "inhumas_csrf",
					Value:    token,
					Path:     "/",
					MaxAge:   86400,
					HttpOnly: true,
					Secure:   secure,
					SameSite: http.SameSiteStrictMode,
				})
			}

			r = r.WithContext(context.WithValue(r.Context(), csrfTokenKey, token))

			if requiresCSRF(r, adminPrefix) {
				submitted := r.Header.Get("X-CSRF-Token")
				if submitted == "" {
					r.Body = http.MaxBytesReader(w, r.Body, bodyLimit)
					if err := r.ParseMultipartForm(32 << 20); err != nil {
						_ = r.ParseForm()
					}
					submitted = r.FormValue("csrf_token")
				}
				if subtle.ConstantTimeCompare([]byte(token), []byte(submitted)) != 1 {
					http.Error(w, "CSRF token invalido", http.StatusForbidden)
					return
				}
			}

			next.ServeHTTP(w, r)
		})
	}
}

func CSRFToken(ctx context.Context) string {
	if token, ok := ctx.Value(csrfTokenKey).(string); ok {
		return token
	}
	return ""
}

func requiresCSRF(r *http.Request, adminPrefix string) bool {
	if r.Method == http.MethodGet || r.Method == http.MethodHead || r.Method == http.MethodOptions {
		return false
	}
	return r.URL.Path == "/login" ||
		r.URL.Path == "/recuperar-senha" ||
		strings.HasPrefix(r.URL.Path, "/redefinir-senha/") ||
		strings.HasPrefix(r.URL.Path, adminPrefix)
}

func csrfTokenFromCookie(r *http.Request) string {
	c, err := r.Cookie("inhumas_csrf")
	if err != nil {
		return ""
	}
	return c.Value
}

func newCSRFToken() string {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}

func newRequestID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.RawURLEncoding.EncodeToString(b[:])
}

func MaintenanceMode(cfg *config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.MaintenanceMode && !strings.HasPrefix(r.URL.Path, cfg.AdminPathPrefix) && r.URL.Path != "/health" {
				w.WriteHeader(http.StatusServiceUnavailable)
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(`<!DOCTYPE html>
<html><head><title>Manutencao</title></head>
<body style="font-family:sans-serif;text-align:center;padding:50px;">
<h1>Manutencao Programada</h1>
<p>Voltamos em breve!</p>
</body></html>`))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func MetricsMiddleware(repo *repository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			atomic.AddUint64(&requestsTotal, 1)
			next.ServeHTTP(w, r)
		})
	}
}

func RateLimitByIP(limit int, window time.Duration) func(http.Handler) http.Handler {
	return RateLimitByIPWhen(limit, window, func(r *http.Request) bool { return true })
}

func RateLimitByIPWhen(limit int, window time.Duration, applies func(*http.Request) bool) func(http.Handler) http.Handler {
	requests := make(map[string][]time.Time)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if applies != nil && !applies(r) {
				next.ServeHTTP(w, r)
				return
			}
			ip := ClientIP(r)
			now := time.Now()
			cutoff := now.Add(-window)
			var recent []time.Time
			for _, t := range requests[ip] {
				if t.After(cutoff) {
					recent = append(recent, t)
				}
			}
			requests[ip] = recent
			if len(recent) >= limit {
				http.Error(w, "Muitas requisicoes", http.StatusTooManyRequests)
				return
			}
			requests[ip] = append(requests[ip], now)
			next.ServeHTTP(w, r)
		})
	}
}
