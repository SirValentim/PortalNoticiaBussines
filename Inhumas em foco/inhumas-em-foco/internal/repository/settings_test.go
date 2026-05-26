package repository

import (
	"context"
	"testing"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestPortalSettingsAreScopedByTenant(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	tenant := &model.Tenant{Name: "LaMafia Music", Slug: "lamafia", PrimaryDomain: "lamafia.music"}
	if err := repo.TenantCreate(ctx, tenant); err != nil {
		t.Fatalf("TenantCreate failed: %v", err)
	}

	defaultSettings := DefaultPortalSettings()
	defaultSettings.SiteName = "Default Portal"
	defaultSettings.SEOTitle = "Default SEO"
	defaultSettings.SEODescription = "Default description"
	if err := repo.PortalSettingsUpdate(ctx, &defaultSettings); err != nil {
		t.Fatalf("PortalSettingsUpdate default failed: %v", err)
	}

	tenantSettings := DefaultPortalSettings()
	tenantSettings.TenantID = tenant.ID
	tenantSettings.SiteName = "LaMafia Music"
	tenantSettings.SEOTitle = "LaMafia SEO"
	tenantSettings.SEODescription = "LaMafia description"
	tenantSettings.ContactEmail = "contato@lamafia.music"
	if err := repo.PortalSettingsUpdate(tenantctx.WithTenant(ctx, tenant), &tenantSettings); err != nil {
		t.Fatalf("PortalSettingsUpdate tenant failed: %v", err)
	}

	gotDefault, err := repo.PortalSettingsGet(ctx)
	if err != nil {
		t.Fatalf("PortalSettingsGet default failed: %v", err)
	}
	if gotDefault.SiteName != "Default Portal" {
		t.Fatalf("default settings leaked: %#v", gotDefault)
	}

	gotTenant, err := repo.PortalSettingsGet(tenantctx.WithTenant(ctx, tenant))
	if err != nil {
		t.Fatalf("PortalSettingsGet tenant failed: %v", err)
	}
	if gotTenant.SiteName != "LaMafia Music" || gotTenant.ContactEmail != "contato@lamafia.music" {
		t.Fatalf("tenant settings not isolated: %#v", gotTenant)
	}
}
