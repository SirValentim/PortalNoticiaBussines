package repository

import (
	"context"
	"testing"
	"time"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestBannersAreScopedByTenant(t *testing.T) {
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
	tenantCtx := tenantctx.WithTenant(ctx, tenant)

	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	defaultBanner := &model.Banner{
		Name:           "Default Hero",
		AdvertiserName: "Cliente Default",
		Position:       "hero",
		ImageKey:       "webp/default-hero.webp",
		LinkURL:        "https://example.com/default",
		StartDate:      start,
		EndDate:        end,
		Status:         "active",
		Active:         true,
		Priority:       1,
	}
	if err := repo.BannerCreate(ctx, defaultBanner); err != nil {
		t.Fatalf("default BannerCreate failed: %v", err)
	}
	if defaultBanner.TenantID != 1 {
		t.Fatalf("default banner tenant_id = %d, want 1", defaultBanner.TenantID)
	}

	tenantBanner := &model.Banner{
		Name:           "Tenant Hero",
		AdvertiserName: "Cliente Tenant",
		Position:       "hero",
		ImageKey:       "webp/tenant-hero.webp",
		LinkURL:        "https://example.com/tenant",
		StartDate:      start,
		EndDate:        end,
		Status:         "active",
		Active:         true,
		Priority:       10,
	}
	if err := repo.BannerCreate(tenantCtx, tenantBanner); err != nil {
		t.Fatalf("tenant BannerCreate failed: %v", err)
	}
	if tenantBanner.TenantID != tenant.ID {
		t.Fatalf("tenant banner tenant_id = %d, want %d", tenantBanner.TenantID, tenant.ID)
	}

	gotDefault, err := repo.BannerGetActiveByPosition(ctx, "hero")
	if err != nil || gotDefault == nil {
		t.Fatalf("default BannerGetActiveByPosition failed: banner=%#v err=%v", gotDefault, err)
	}
	if gotDefault.Name == "Tenant Hero" {
		t.Fatalf("tenant banner leaked into default active lookup: %#v", gotDefault)
	}

	gotTenant, err := repo.BannerGetActiveByPosition(tenantCtx, "hero")
	if err != nil || gotTenant == nil {
		t.Fatalf("tenant BannerGetActiveByPosition failed: banner=%#v err=%v", gotTenant, err)
	}
	if gotTenant.Name != "Tenant Hero" {
		t.Fatalf("tenant banner mismatch: %#v", gotTenant)
	}

	if got, err := repo.BannerGetByID(ctx, tenantBanner.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant banner: banner=%#v err=%v", got, err)
	}
	if got, err := repo.BannerGetByID(tenantCtx, tenantBanner.ID); err != nil || got == nil {
		t.Fatalf("tenant context could not access banner: banner=%#v err=%v", got, err)
	}

	defaultList, err := repo.BannerList(ctx)
	if err != nil {
		t.Fatalf("default BannerList failed: %v", err)
	}
	tenantList, err := repo.BannerList(tenantCtx)
	if err != nil {
		t.Fatalf("tenant BannerList failed: %v", err)
	}
	if len(defaultList) != 1 || len(tenantList) != 1 {
		t.Fatalf("unexpected banner lists: default=%d tenant=%d", len(defaultList), len(tenantList))
	}
}

func TestBannerOverlapIsScopedByTenant(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	tenant := &model.Tenant{Name: "LaMafia Music", Slug: "lamafia"}
	if err := repo.TenantCreate(ctx, tenant); err != nil {
		t.Fatalf("TenantCreate failed: %v", err)
	}
	tenantCtx := tenantctx.WithTenant(ctx, tenant)

	start := time.Date(2026, 5, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 5, 31, 0, 0, 0, 0, time.UTC)
	if err := repo.BannerCreate(ctx, &model.Banner{
		Name:           "Default",
		AdvertiserName: "Cliente",
		Position:       "hero",
		ImageKey:       "webp/default.webp",
		LinkURL:        "https://example.com",
		StartDate:      start,
		EndDate:        end,
		Status:         "active",
		Active:         true,
	}); err != nil {
		t.Fatalf("BannerCreate failed: %v", err)
	}

	defaultCount, err := repo.BannerCountActiveInPeriod(ctx, "hero", start, end)
	if err != nil {
		t.Fatalf("default BannerCountActiveInPeriod failed: %v", err)
	}
	tenantCount, err := repo.BannerCountActiveInPeriod(tenantCtx, "hero", start, end)
	if err != nil {
		t.Fatalf("tenant BannerCountActiveInPeriod failed: %v", err)
	}
	if defaultCount != 1 || tenantCount != 0 {
		t.Fatalf("unexpected overlap counts: default=%d tenant=%d", defaultCount, tenantCount)
	}
}
