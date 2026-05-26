package repository

import (
	"context"
	"testing"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestCategoriesAreScopedByTenant(t *testing.T) {
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

	defaultCategory, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || defaultCategory == nil {
		t.Fatalf("default CategoryGetBySlug failed: category=%#v err=%v", defaultCategory, err)
	}
	if defaultCategory.TenantID != 1 {
		t.Fatalf("default category tenant_id = %d, want 1", defaultCategory.TenantID)
	}

	tenantCategory := &model.Category{
		Slug:   "noticias",
		Name:   "Noticias da Musica",
		Active: true,
	}
	if err := repo.CategoryCreate(tenantCtx, tenantCategory); err != nil {
		t.Fatalf("tenant CategoryCreate failed: %v", err)
	}
	if tenantCategory.TenantID != tenant.ID {
		t.Fatalf("tenant category tenant_id = %d, want %d", tenantCategory.TenantID, tenant.ID)
	}

	gotDefault, err := repo.CategoryGetBySlug(ctx, "noticias")
	if err != nil || gotDefault == nil {
		t.Fatalf("default category lookup failed: category=%#v err=%v", gotDefault, err)
	}
	if gotDefault.Name == "Noticias da Musica" {
		t.Fatalf("tenant category leaked into default lookup: %#v", gotDefault)
	}

	gotTenant, err := repo.CategoryGetBySlug(tenantCtx, "noticias")
	if err != nil || gotTenant == nil {
		t.Fatalf("tenant category lookup failed: category=%#v err=%v", gotTenant, err)
	}
	if gotTenant.Name != "Noticias da Musica" {
		t.Fatalf("tenant category mismatch: %#v", gotTenant)
	}

	defaultList, err := repo.CategoryList(ctx)
	if err != nil {
		t.Fatalf("default CategoryList failed: %v", err)
	}
	tenantList, err := repo.CategoryList(tenantCtx)
	if err != nil {
		t.Fatalf("tenant CategoryList failed: %v", err)
	}
	if len(defaultList) == 0 || len(tenantList) != 1 {
		t.Fatalf("unexpected category lists: default=%d tenant=%d", len(defaultList), len(tenantList))
	}
}

func TestCategoryGetByIDDoesNotCrossTenants(t *testing.T) {
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

	category := &model.Category{Slug: "reviews", Name: "Reviews", Active: true}
	if err := repo.CategoryCreate(tenantCtx, category); err != nil {
		t.Fatalf("CategoryCreate failed: %v", err)
	}

	if got, err := repo.CategoryGetByID(ctx, category.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant category: category=%#v err=%v", got, err)
	}
	if got, err := repo.CategoryGetByID(tenantCtx, category.ID); err != nil || got == nil {
		t.Fatalf("tenant context could not access category: category=%#v err=%v", got, err)
	}
}
