package middleware

import (
	"context"
	"net/http"

	"inhumas-em-foco/internal/config"
)

type brandingContextKey struct{}

var BrandingContextKey = brandingContextKey{}

func InjectBranding(branding *config.TenantBrandingConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := context.WithValue(r.Context(), BrandingContextKey, branding)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func BrandingFromContext(ctx context.Context) *config.TenantBrandingConfig {
	b, _ := ctx.Value(BrandingContextKey).(*config.TenantBrandingConfig)
	return b
}
