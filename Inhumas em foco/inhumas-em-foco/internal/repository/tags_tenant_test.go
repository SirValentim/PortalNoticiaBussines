package repository

import (
	"context"
	"testing"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestTagsAreScopedByTenant(t *testing.T) {
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

	defaultTag := &model.Tag{Slug: "agenda", Name: "Agenda", Active: true}
	if err := repo.TagCreate(ctx, defaultTag); err != nil {
		t.Fatalf("default TagCreate failed: %v", err)
	}
	if defaultTag.TenantID != 1 {
		t.Fatalf("default tag tenant_id = %d, want 1", defaultTag.TenantID)
	}

	tenantTag := &model.Tag{Slug: "agenda", Name: "Agenda de Shows", Active: true}
	if err := repo.TagCreate(tenantCtx, tenantTag); err != nil {
		t.Fatalf("tenant TagCreate failed: %v", err)
	}
	if tenantTag.TenantID != tenant.ID {
		t.Fatalf("tenant tag tenant_id = %d, want %d", tenantTag.TenantID, tenant.ID)
	}

	gotDefault, err := repo.TagGetBySlug(ctx, "agenda")
	if err != nil || gotDefault == nil {
		t.Fatalf("default TagGetBySlug failed: tag=%#v err=%v", gotDefault, err)
	}
	if gotDefault.Name == "Agenda de Shows" {
		t.Fatalf("tenant tag leaked into default lookup: %#v", gotDefault)
	}

	gotTenant, err := repo.TagGetBySlug(tenantCtx, "agenda")
	if err != nil || gotTenant == nil {
		t.Fatalf("tenant TagGetBySlug failed: tag=%#v err=%v", gotTenant, err)
	}
	if gotTenant.Name != "Agenda de Shows" {
		t.Fatalf("tenant tag mismatch: %#v", gotTenant)
	}

	defaultList, err := repo.TagList(ctx, true)
	if err != nil {
		t.Fatalf("default TagList failed: %v", err)
	}
	tenantList, err := repo.TagList(tenantCtx, true)
	if err != nil {
		t.Fatalf("tenant TagList failed: %v", err)
	}
	if len(defaultList) != 1 || len(tenantList) != 1 {
		t.Fatalf("unexpected tag lists: default=%d tenant=%d", len(defaultList), len(tenantList))
	}
}

func TestTagGetByIDDoesNotCrossTenants(t *testing.T) {
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

	tag := &model.Tag{Slug: "reviews", Name: "Reviews", Active: true}
	if err := repo.TagCreate(tenantCtx, tag); err != nil {
		t.Fatalf("TagCreate failed: %v", err)
	}

	if got, err := repo.TagGetByID(ctx, tag.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant tag: tag=%#v err=%v", got, err)
	}
	if got, err := repo.TagGetByID(tenantCtx, tag.ID); err != nil || got == nil {
		t.Fatalf("tenant context could not access tag: tag=%#v err=%v", got, err)
	}
}
