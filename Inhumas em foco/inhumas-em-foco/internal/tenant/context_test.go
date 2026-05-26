package tenant

import (
	"context"
	"net/http/httptest"
	"testing"

	"inhumas-em-foco/internal/model"
)

func TestTenantContext(t *testing.T) {
	expected := &model.Tenant{ID: 10, Slug: "lamafia"}
	ctx := WithTenant(context.Background(), expected)

	if got := FromContext(ctx); got != expected {
		t.Fatalf("FromContext() = %#v, want %#v", got, expected)
	}
}

func TestDomainFromRequest(t *testing.T) {
	req := httptest.NewRequest("GET", "https://example.test/path", nil)
	req.Host = "LaMafia.Music:443"

	if got := DomainFromRequest(req); got != "lamafia.music" {
		t.Fatalf("DomainFromRequest() = %q", got)
	}

	req.Header.Set("X-Forwarded-Host", "GoiasNews.com.br, proxy.local")
	if got := DomainFromRequest(req); got != "goiasnews.com.br" {
		t.Fatalf("DomainFromRequest() forwarded = %q", got)
	}
}
