package repository

import (
	"context"
	"testing"
	"time"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestPostsAreScopedByTenant(t *testing.T) {
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

	publishedAt := time.Now()
	defaultPost := &model.Post{
		Title:       "Agenda da cidade",
		Slug:        "agenda-da-cidade",
		Content:     "Programacao local em Inhumas.",
		Status:      model.StatusPublished,
		PublishedAt: &publishedAt,
	}
	if err := repo.PostCreate(ctx, defaultPost); err != nil {
		t.Fatalf("default PostCreate failed: %v", err)
	}
	if defaultPost.TenantID != 1 {
		t.Fatalf("default post tenant_id = %d, want 1", defaultPost.TenantID)
	}

	tenantPost := &model.Post{
		Title:       "Agenda de shows",
		Slug:        "agenda-da-cidade",
		Content:     "Programacao de shows e bastidores da musica.",
		Status:      model.StatusPublished,
		PublishedAt: &publishedAt,
	}
	if err := repo.PostCreate(tenantCtx, tenantPost); err != nil {
		t.Fatalf("tenant PostCreate failed: %v", err)
	}
	if tenantPost.TenantID != tenant.ID {
		t.Fatalf("tenant post tenant_id = %d, want %d", tenantPost.TenantID, tenant.ID)
	}

	gotDefault, err := repo.PostGetBySlug(ctx, "agenda-da-cidade")
	if err != nil || gotDefault == nil {
		t.Fatalf("default PostGetBySlug failed: post=%#v err=%v", gotDefault, err)
	}
	if gotDefault.Title == "Agenda de shows" {
		t.Fatalf("tenant post leaked into default lookup: %#v", gotDefault)
	}

	gotTenant, err := repo.PostGetBySlug(tenantCtx, "agenda-da-cidade")
	if err != nil || gotTenant == nil {
		t.Fatalf("tenant PostGetBySlug failed: post=%#v err=%v", gotTenant, err)
	}
	if gotTenant.Title != "Agenda de shows" {
		t.Fatalf("tenant post mismatch: %#v", gotTenant)
	}

	defaultList, err := repo.PostListPublished(ctx, 10, 0)
	if err != nil {
		t.Fatalf("default PostListPublished failed: %v", err)
	}
	tenantList, err := repo.PostListPublished(tenantCtx, 10, 0)
	if err != nil {
		t.Fatalf("tenant PostListPublished failed: %v", err)
	}
	if len(defaultList) != 1 || len(tenantList) != 1 {
		t.Fatalf("unexpected published lists: default=%d tenant=%d", len(defaultList), len(tenantList))
	}
	if defaultList[0].ID == tenantList[0].ID {
		t.Fatalf("published lists returned same post across tenants")
	}
}

func TestPostGetByIDAndSearchDoNotCrossTenants(t *testing.T) {
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

	publishedAt := time.Now()
	post := &model.Post{
		Title:       "Festival exclusivo da musica",
		Slug:        "festival-exclusivo-da-musica",
		Excerpt:     "Agenda musical",
		Content:     "Palco exclusivo com bandas independentes.",
		Status:      model.StatusPublished,
		PublishedAt: &publishedAt,
	}
	if err := repo.PostCreate(tenantCtx, post); err != nil {
		t.Fatalf("PostCreate failed: %v", err)
	}

	if got, err := repo.PostGetByID(ctx, post.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant post: post=%#v err=%v", got, err)
	}
	if got, err := repo.PostGetByID(tenantCtx, post.ID); err != nil || got == nil {
		t.Fatalf("tenant context could not access post: post=%#v err=%v", got, err)
	}
	if repo.PostSlugExists(ctx, "festival-exclusivo-da-musica") {
		t.Fatalf("default context saw tenant slug")
	}
	if !repo.PostSlugExists(tenantCtx, "festival-exclusivo-da-musica") {
		t.Fatalf("tenant context did not see tenant slug")
	}

	defaultResults, err := repo.PostSearch(ctx, "bandas independentes", 10)
	if err != nil {
		t.Fatalf("default PostSearch failed: %v", err)
	}
	tenantResults, err := repo.PostSearch(tenantCtx, "bandas independentes", 10)
	if err != nil {
		t.Fatalf("tenant PostSearch failed: %v", err)
	}
	if len(defaultResults) != 0 {
		t.Fatalf("default search leaked tenant post: %#v", defaultResults)
	}
	if len(tenantResults) != 1 || tenantResults[0].ID != post.ID {
		t.Fatalf("tenant search results = %#v, want tenant post", tenantResults)
	}
}
