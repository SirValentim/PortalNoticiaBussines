package repository

import (
	"context"
	"testing"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestNeighborhoodsAreScopedByTenant(t *testing.T) {
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

	defaultNeighborhood := &model.Neighborhood{Slug: "centro", Name: "Centro"}
	if err := repo.NeighborhoodCreate(ctx, defaultNeighborhood); err != nil {
		t.Fatalf("default NeighborhoodCreate failed: %v", err)
	}
	tenantNeighborhood := &model.Neighborhood{Slug: "centro", Name: "Centro Musical"}
	if err := repo.NeighborhoodCreate(tenantCtx, tenantNeighborhood); err != nil {
		t.Fatalf("tenant NeighborhoodCreate failed: %v", err)
	}

	if defaultNeighborhood.TenantID != 1 || tenantNeighborhood.TenantID != tenant.ID {
		t.Fatalf("unexpected tenant ids: default=%d tenant=%d want default=1 tenant=%d", defaultNeighborhood.TenantID, tenantNeighborhood.TenantID, tenant.ID)
	}

	gotDefault, err := repo.NeighborhoodGetBySlug(ctx, "centro")
	if err != nil || gotDefault == nil {
		t.Fatalf("default NeighborhoodGetBySlug failed: neighborhood=%#v err=%v", gotDefault, err)
	}
	if gotDefault.Name == "Centro Musical" {
		t.Fatalf("tenant neighborhood leaked into default lookup: %#v", gotDefault)
	}

	gotTenant, err := repo.NeighborhoodGetBySlug(tenantCtx, "centro")
	if err != nil || gotTenant == nil {
		t.Fatalf("tenant NeighborhoodGetBySlug failed: neighborhood=%#v err=%v", gotTenant, err)
	}
	if gotTenant.Name != "Centro Musical" {
		t.Fatalf("tenant neighborhood mismatch: %#v", gotTenant)
	}

	defaultList, err := repo.NeighborhoodList(ctx)
	if err != nil {
		t.Fatalf("default NeighborhoodList failed: %v", err)
	}
	tenantList, err := repo.NeighborhoodList(tenantCtx)
	if err != nil {
		t.Fatalf("tenant NeighborhoodList failed: %v", err)
	}
	if len(defaultList) != 1 || len(tenantList) != 1 {
		t.Fatalf("unexpected neighborhood lists: default=%d tenant=%d", len(defaultList), len(tenantList))
	}
}

func TestStoresAreScopedByTenant(t *testing.T) {
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

	defaultStore := &model.Store{Slug: "loja-central", Name: "Loja Central", Active: true, IsFeatured: true}
	if err := repo.StoreCreate(ctx, defaultStore); err != nil {
		t.Fatalf("default StoreCreate failed: %v", err)
	}
	tenantStore := &model.Store{Slug: "loja-central", Name: "Loja Central Music", Active: true, IsFeatured: true}
	if err := repo.StoreCreate(tenantCtx, tenantStore); err != nil {
		t.Fatalf("tenant StoreCreate failed: %v", err)
	}

	if defaultStore.TenantID != 1 || tenantStore.TenantID != tenant.ID {
		t.Fatalf("unexpected tenant ids: default=%d tenant=%d want default=1 tenant=%d", defaultStore.TenantID, tenantStore.TenantID, tenant.ID)
	}

	gotDefault, err := repo.StoreGetBySlug(ctx, "loja-central")
	if err != nil || gotDefault == nil {
		t.Fatalf("default StoreGetBySlug failed: store=%#v err=%v", gotDefault, err)
	}
	if gotDefault.Name == "Loja Central Music" {
		t.Fatalf("tenant store leaked into default lookup: %#v", gotDefault)
	}

	gotTenant, err := repo.StoreGetBySlug(tenantCtx, "loja-central")
	if err != nil || gotTenant == nil {
		t.Fatalf("tenant StoreGetBySlug failed: store=%#v err=%v", gotTenant, err)
	}
	if gotTenant.Name != "Loja Central Music" {
		t.Fatalf("tenant store mismatch: %#v", gotTenant)
	}

	if got, err := repo.StoreGetByID(ctx, tenantStore.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant store: store=%#v err=%v", got, err)
	}
	if got, err := repo.StoreGetByID(tenantCtx, tenantStore.ID); err != nil || got == nil {
		t.Fatalf("tenant context could not access store: store=%#v err=%v", got, err)
	}

	defaultFeatured, err := repo.StoreListFeatured(ctx, 10)
	if err != nil {
		t.Fatalf("default StoreListFeatured failed: %v", err)
	}
	tenantFeatured, err := repo.StoreListFeatured(tenantCtx, 10)
	if err != nil {
		t.Fatalf("tenant StoreListFeatured failed: %v", err)
	}
	if len(defaultFeatured) != 1 || len(tenantFeatured) != 1 {
		t.Fatalf("unexpected featured lists: default=%d tenant=%d", len(defaultFeatured), len(tenantFeatured))
	}
	if defaultFeatured[0].ID == tenantFeatured[0].ID {
		t.Fatalf("featured lists returned same store across tenants")
	}
}
