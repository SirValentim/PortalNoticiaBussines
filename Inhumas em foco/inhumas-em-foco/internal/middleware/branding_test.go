package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"inhumas-em-foco/internal/config"
)

func TestInjectBranding(t *testing.T) {
	branding := &config.TenantBrandingConfig{PortalName: "Test Portal"}
	var got *config.TenantBrandingConfig

	handler := InjectBranding(branding)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = BrandingFromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got != branding {
		t.Fatalf("branding nao foi injetado no contexto")
	}
}
