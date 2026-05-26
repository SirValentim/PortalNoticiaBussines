package tenant

import (
	"context"
	"net"
	"net/http"
	"strings"

	"inhumas-em-foco/internal/model"
)

type contextKey struct{}

func WithTenant(ctx context.Context, tenant *model.Tenant) context.Context {
	return context.WithValue(ctx, contextKey{}, tenant)
}

func FromContext(ctx context.Context) *model.Tenant {
	tenant, _ := ctx.Value(contextKey{}).(*model.Tenant)
	return tenant
}

func DomainFromRequest(r *http.Request) string {
	host := strings.TrimSpace(r.Host)
	if forwardedHost := strings.TrimSpace(r.Header.Get("X-Forwarded-Host")); forwardedHost != "" {
		host = strings.Split(forwardedHost, ",")[0]
	}
	if hostOnly, _, err := net.SplitHostPort(host); err == nil {
		host = hostOnly
	}
	return NormalizeDomain(host)
}

func NormalizeDomain(domain string) string {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimSuffix(domain, "/")
	return domain
}
