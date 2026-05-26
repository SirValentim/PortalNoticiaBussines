package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestResolveTenantByDomain(t *testing.T) {
	repo, err := repository.New(":memory:")
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	defer repo.Close()

	tenant := &model.Tenant{Name: "LaMafia Music", Slug: "lamafia", PrimaryDomain: "lamafia.music"}
	if err := repo.TenantCreate(t.Context(), tenant); err != nil {
		t.Fatalf("TenantCreate failed: %v", err)
	}
	if err := repo.TenantDomainCreate(t.Context(), &model.TenantDomain{TenantID: tenant.ID, Domain: "lamafia.music", IsPrimary: true}); err != nil {
		t.Fatalf("TenantDomainCreate failed: %v", err)
	}

	var got *model.Tenant
	handler := ResolveTenant(repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = tenantctx.FromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "https://lamafia.music/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if got == nil || got.ID != tenant.ID {
		t.Fatalf("tenant from context = %#v, want ID %d", got, tenant.ID)
	}
}

func TestResolveTenantFallsBackToDefault(t *testing.T) {
	repo, err := repository.New(":memory:")
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	defer repo.Close()

	var got *model.Tenant
	handler := ResolveTenant(repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = tenantctx.FromContext(r.Context())
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "https://unknown.example/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d; body=%s", rec.Code, http.StatusNoContent, rec.Body.String())
	}
	if got == nil || got.Slug != "default" {
		t.Fatalf("tenant fallback = %#v, want default", got)
	}
}

func TestResolveTenantRejectsInactiveTenant(t *testing.T) {
	repo, err := repository.New(":memory:")
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	defer repo.Close()

	inactive := &model.Tenant{Name: "Paused Portal", Slug: "paused", Status: "inactive", PrimaryDomain: "paused.example"}
	if err := repo.TenantCreate(t.Context(), inactive); err != nil {
		t.Fatalf("TenantCreate failed: %v", err)
	}
	if err := repo.TenantDomainCreate(t.Context(), &model.TenantDomain{TenantID: inactive.ID, Domain: "paused.example", IsPrimary: true}); err != nil {
		t.Fatalf("TenantDomainCreate failed: %v", err)
	}

	handler := ResolveTenant(repo)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))

	req := httptest.NewRequest(http.MethodGet, "https://paused.example/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusServiceUnavailable)
	}
}
