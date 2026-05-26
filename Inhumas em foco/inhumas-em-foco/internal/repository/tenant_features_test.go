package repository

import (
	"context"
	"testing"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestTenantFeaturesAreScopedByTenant(t *testing.T) {
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

	defaultLimit := int64(3)
	defaultFeature := &model.TenantFeature{
		Feature: "automation",
		Enabled: true,
		Limit:   &defaultLimit,
	}
	if err := repo.TenantFeatureUpsert(ctx, defaultFeature); err != nil {
		t.Fatalf("default TenantFeatureUpsert failed: %v", err)
	}

	tenantLimit := int64(10)
	tenantFeature := &model.TenantFeature{
		Feature: "AUTOMATION",
		Enabled: false,
		Limit:   &tenantLimit,
	}
	if err := repo.TenantFeatureUpsert(tenantCtx, tenantFeature); err != nil {
		t.Fatalf("tenant TenantFeatureUpsert failed: %v", err)
	}

	gotDefault, err := repo.TenantFeatureGet(ctx, "automation")
	if err != nil || gotDefault == nil {
		t.Fatalf("default TenantFeatureGet failed: feature=%#v err=%v", gotDefault, err)
	}
	if gotDefault.TenantID != 1 || !gotDefault.Enabled || gotDefault.Limit == nil || *gotDefault.Limit != defaultLimit {
		t.Fatalf("unexpected default feature: %#v", gotDefault)
	}

	gotTenant, err := repo.TenantFeatureGet(tenantCtx, "automation")
	if err != nil || gotTenant == nil {
		t.Fatalf("tenant TenantFeatureGet failed: feature=%#v err=%v", gotTenant, err)
	}
	if gotTenant.TenantID != tenant.ID || gotTenant.Enabled || gotTenant.Limit == nil || *gotTenant.Limit != tenantLimit {
		t.Fatalf("unexpected tenant feature: %#v", gotTenant)
	}

	enabled, err := repo.TenantFeatureEnabled(tenantCtx, "automation")
	if err != nil {
		t.Fatalf("TenantFeatureEnabled failed: %v", err)
	}
	if enabled {
		t.Fatal("TenantFeatureEnabled returned true for disabled tenant feature")
	}

	features, err := repo.TenantFeatureList(tenantCtx)
	if err != nil {
		t.Fatalf("TenantFeatureList failed: %v", err)
	}
	if len(features) != 1 || features[0].ID != tenantFeature.ID {
		t.Fatalf("unexpected tenant feature list: %#v", features)
	}
}

func TestTenantFeatureUpsertUpdatesExistingFeature(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	limit := int64(1)
	feature := &model.TenantFeature{Feature: "media", Enabled: true, Limit: &limit}
	if err := repo.TenantFeatureUpsert(ctx, feature); err != nil {
		t.Fatalf("TenantFeatureUpsert create failed: %v", err)
	}
	createdID := feature.ID

	newLimit := int64(5)
	feature.Enabled = false
	feature.Limit = &newLimit
	if err := repo.TenantFeatureUpsert(ctx, feature); err != nil {
		t.Fatalf("TenantFeatureUpsert update failed: %v", err)
	}
	if feature.ID != createdID {
		t.Fatalf("feature id changed after upsert: got %d want %d", feature.ID, createdID)
	}

	got, err := repo.TenantFeatureGet(ctx, "media")
	if err != nil || got == nil {
		t.Fatalf("TenantFeatureGet failed: feature=%#v err=%v", got, err)
	}
	if got.Enabled || got.Limit == nil || *got.Limit != newLimit {
		t.Fatalf("feature was not updated: %#v", got)
	}
}
