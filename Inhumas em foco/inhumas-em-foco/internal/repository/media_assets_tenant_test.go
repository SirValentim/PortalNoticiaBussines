package repository

import (
	"context"
	"testing"
	"time"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestMediaAssetsAreScopedByTenant(t *testing.T) {
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

	defaultAsset := &model.MediaAsset{
		Key:          "webp/shared.webp",
		OriginalName: "shared-default.png",
		Title:        "Imagem Default",
		AltText:      "Default",
		ContentType:  "image/webp",
	}
	if err := repo.MediaAssetCreate(ctx, defaultAsset); err != nil {
		t.Fatalf("default MediaAssetCreate failed: %v", err)
	}
	if defaultAsset.TenantID != 1 {
		t.Fatalf("default asset tenant_id = %d, want 1", defaultAsset.TenantID)
	}

	tenantAsset := &model.MediaAsset{
		Key:          "webp/shared.webp",
		OriginalName: "shared-tenant.png",
		Title:        "Imagem Tenant",
		AltText:      "Tenant",
		ContentType:  "image/webp",
	}
	if err := repo.MediaAssetCreate(tenantCtx, tenantAsset); err != nil {
		t.Fatalf("tenant MediaAssetCreate failed: %v", err)
	}
	if tenantAsset.TenantID != tenant.ID {
		t.Fatalf("tenant asset tenant_id = %d, want %d", tenantAsset.TenantID, tenant.ID)
	}

	gotDefault, err := repo.MediaAssetGetByKey(ctx, "webp/shared.webp")
	if err != nil || gotDefault == nil {
		t.Fatalf("default MediaAssetGetByKey failed: asset=%#v err=%v", gotDefault, err)
	}
	if gotDefault.Title == "Imagem Tenant" {
		t.Fatalf("tenant asset leaked into default lookup: %#v", gotDefault)
	}

	gotTenant, err := repo.MediaAssetGetByKey(tenantCtx, "webp/shared.webp")
	if err != nil || gotTenant == nil {
		t.Fatalf("tenant MediaAssetGetByKey failed: asset=%#v err=%v", gotTenant, err)
	}
	if gotTenant.Title != "Imagem Tenant" {
		t.Fatalf("tenant asset mismatch: %#v", gotTenant)
	}

	if got, err := repo.MediaAssetGetByID(ctx, tenantAsset.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant asset: asset=%#v err=%v", got, err)
	}
	if got, err := repo.MediaAssetGetByID(tenantCtx, tenantAsset.ID); err != nil || got == nil {
		t.Fatalf("tenant context could not access asset: asset=%#v err=%v", got, err)
	}

	defaultList, err := repo.MediaAssetList(ctx, "", 10, 0)
	if err != nil {
		t.Fatalf("default MediaAssetList failed: %v", err)
	}
	tenantList, err := repo.MediaAssetList(tenantCtx, "", 10, 0)
	if err != nil {
		t.Fatalf("tenant MediaAssetList failed: %v", err)
	}
	if len(defaultList) != 1 || len(tenantList) != 1 {
		t.Fatalf("unexpected asset lists: default=%d tenant=%d", len(defaultList), len(tenantList))
	}
	if defaultList[0].ID == tenantList[0].ID {
		t.Fatalf("asset lists returned same asset across tenants")
	}
}

func TestMediaAssetCountsAndArchiveMonthsAreScopedByTenant(t *testing.T) {
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

	defaultAsset := &model.MediaAsset{Key: "webp/default.webp", OriginalName: "default.png", Title: "Default", ContentType: "image/webp"}
	tenantAsset := &model.MediaAsset{Key: "webp/tenant.webp", OriginalName: "tenant.png", Title: "Tenant", ContentType: "image/webp"}
	if err := repo.MediaAssetCreate(ctx, defaultAsset); err != nil {
		t.Fatalf("default MediaAssetCreate failed: %v", err)
	}
	if err := repo.MediaAssetCreate(tenantCtx, tenantAsset); err != nil {
		t.Fatalf("tenant MediaAssetCreate failed: %v", err)
	}

	jan := time.Date(2026, 1, 15, 10, 0, 0, 0, time.UTC)
	may := time.Date(2026, 5, 12, 10, 0, 0, 0, time.UTC)
	if _, err := repo.DB().ExecContext(ctx, `UPDATE media_assets SET created_at = $1, updated_at = $1 WHERE id = $2`, jan, defaultAsset.ID); err != nil {
		t.Fatalf("default media date update failed: %v", err)
	}
	if _, err := repo.DB().ExecContext(ctx, `UPDATE media_assets SET created_at = $1, updated_at = $1 WHERE id = $2`, may, tenantAsset.ID); err != nil {
		t.Fatalf("tenant media date update failed: %v", err)
	}

	defaultCount, err := repo.MediaAssetCount(ctx, "")
	if err != nil {
		t.Fatalf("default MediaAssetCount failed: %v", err)
	}
	tenantCount, err := repo.MediaAssetCount(tenantCtx, "")
	if err != nil {
		t.Fatalf("tenant MediaAssetCount failed: %v", err)
	}
	if defaultCount != 1 || tenantCount != 1 {
		t.Fatalf("unexpected counts: default=%d tenant=%d", defaultCount, tenantCount)
	}

	defaultMonths, err := repo.MediaAssetArchiveMonths(ctx)
	if err != nil {
		t.Fatalf("default MediaAssetArchiveMonths failed: %v", err)
	}
	tenantMonths, err := repo.MediaAssetArchiveMonths(tenantCtx)
	if err != nil {
		t.Fatalf("tenant MediaAssetArchiveMonths failed: %v", err)
	}
	if len(defaultMonths) != 1 || defaultMonths[0].Month != "2026-01" {
		t.Fatalf("default months = %#v, want only 2026-01", defaultMonths)
	}
	if len(tenantMonths) != 1 || tenantMonths[0].Month != "2026-05" {
		t.Fatalf("tenant months = %#v, want only 2026-05", tenantMonths)
	}
}
