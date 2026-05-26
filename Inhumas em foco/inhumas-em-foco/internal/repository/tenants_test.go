package repository

import (
	"context"
	"testing"

	"inhumas-em-foco/internal/model"
)

func TestTenantRepositoryCreatesAndResolvesByDomain(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	tenant := &model.Tenant{
		Name:          "LaMafia Music",
		Slug:          "lamafia",
		PrimaryDomain: "https://LaMafia.Music/",
	}
	if err := repo.TenantCreate(ctx, tenant); err != nil {
		t.Fatalf("TenantCreate failed: %v", err)
	}
	if tenant.ID == 0 {
		t.Fatal("TenantCreate did not set ID")
	}

	domain := &model.TenantDomain{TenantID: tenant.ID, Domain: "LaMafia.Music", IsPrimary: true}
	if err := repo.TenantDomainCreate(ctx, domain); err != nil {
		t.Fatalf("TenantDomainCreate failed: %v", err)
	}

	bySlug, err := repo.TenantGetBySlug(ctx, "lamafia")
	if err != nil || bySlug == nil {
		t.Fatalf("TenantGetBySlug failed: tenant=%#v err=%v", bySlug, err)
	}
	if bySlug.PrimaryDomain != "lamafia.music" || bySlug.Status != "active" {
		t.Fatalf("unexpected tenant normalization: %#v", bySlug)
	}

	byDomain, err := repo.TenantGetByDomain(ctx, "https://lamafia.music/")
	if err != nil || byDomain == nil {
		t.Fatalf("TenantGetByDomain failed: tenant=%#v err=%v", byDomain, err)
	}
	if byDomain.ID != tenant.ID {
		t.Fatalf("TenantGetByDomain ID = %d, want %d", byDomain.ID, tenant.ID)
	}
}

func TestTenantRepositorySeedsDefaultTenant(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	defer repo.Close()

	tenant, err := repo.TenantGetBySlug(context.Background(), "default")
	if err != nil {
		t.Fatalf("TenantGetBySlug default failed: %v", err)
	}
	if tenant == nil || tenant.Status != "active" {
		t.Fatalf("default tenant not seeded: %#v", tenant)
	}
}
