package repository

import (
	"context"
	"testing"
	"time"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestAutomationSourcesAreScopedByTenant(t *testing.T) {
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

	defaultSource := &model.AutomationSource{
		Name:       "Default Feed",
		SourceType: string(model.AutomationSourceRSS),
		URL:        "https://example.com/default.xml",
		Active:     true,
	}
	if err := repo.AutomationSourceCreate(ctx, defaultSource); err != nil {
		t.Fatalf("default AutomationSourceCreate failed: %v", err)
	}

	tenantSource := &model.AutomationSource{
		Name:       "Tenant Feed",
		SourceType: string(model.AutomationSourceRSS),
		URL:        "https://example.com/tenant.xml",
		Active:     true,
	}
	if err := repo.AutomationSourceCreate(tenantCtx, tenantSource); err != nil {
		t.Fatalf("tenant AutomationSourceCreate failed: %v", err)
	}

	if defaultSource.TenantID != 1 {
		t.Fatalf("default source tenant_id = %d, want 1", defaultSource.TenantID)
	}
	if tenantSource.TenantID != tenant.ID {
		t.Fatalf("tenant source tenant_id = %d, want %d", tenantSource.TenantID, tenant.ID)
	}

	defaultSources, err := repo.AutomationSourceList(ctx, false, 20)
	if err != nil {
		t.Fatalf("default AutomationSourceList failed: %v", err)
	}
	tenantSources, err := repo.AutomationSourceList(tenantCtx, false, 20)
	if err != nil {
		t.Fatalf("tenant AutomationSourceList failed: %v", err)
	}
	if len(defaultSources) != 1 || defaultSources[0].ID != defaultSource.ID {
		t.Fatalf("unexpected default sources: %#v", defaultSources)
	}
	if len(tenantSources) != 1 || tenantSources[0].ID != tenantSource.ID {
		t.Fatalf("unexpected tenant sources: %#v", tenantSources)
	}

	if got, err := repo.AutomationSourceGetByID(ctx, tenantSource.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant source: source=%#v err=%v", got, err)
	}
	if got, err := repo.AutomationSourceGetByID(tenantCtx, tenantSource.ID); err != nil || got == nil {
		t.Fatalf("tenant context could not access source: source=%#v err=%v", got, err)
	}

	runAt := time.Now()
	if err := repo.AutomationSourceMarkRun(ctx, tenantSource.ID, runAt); err != nil {
		t.Fatalf("default AutomationSourceMarkRun failed: %v", err)
	}
	got, err := repo.AutomationSourceGetByID(tenantCtx, tenantSource.ID)
	if err != nil || got == nil {
		t.Fatalf("tenant AutomationSourceGetByID failed: source=%#v err=%v", got, err)
	}
	if got.LastRunAt != nil {
		t.Fatalf("default context changed tenant source last_run_at: %#v", got.LastRunAt)
	}
}

func TestAutomationRunsAreScopedByTenant(t *testing.T) {
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

	defaultSource := &model.AutomationSource{Name: "Default Feed", SourceType: string(model.AutomationSourceRSS), URL: "https://example.com/default.xml", Active: true}
	if err := repo.AutomationSourceCreate(ctx, defaultSource); err != nil {
		t.Fatalf("default AutomationSourceCreate failed: %v", err)
	}
	tenantSource := &model.AutomationSource{Name: "Tenant Feed", SourceType: string(model.AutomationSourceRSS), URL: "https://example.com/tenant.xml", Active: true}
	if err := repo.AutomationSourceCreate(tenantCtx, tenantSource); err != nil {
		t.Fatalf("tenant AutomationSourceCreate failed: %v", err)
	}

	defaultRun := &model.AutomationRun{SourceID: &defaultSource.ID, Status: string(model.AutomationRunSuccess), ItemsFound: 1}
	if err := repo.AutomationRunCreate(ctx, defaultRun); err != nil {
		t.Fatalf("default AutomationRunCreate failed: %v", err)
	}
	tenantRun := &model.AutomationRun{SourceID: &tenantSource.ID, Status: string(model.AutomationRunSuccess), ItemsFound: 2}
	if err := repo.AutomationRunCreate(tenantCtx, tenantRun); err != nil {
		t.Fatalf("tenant AutomationRunCreate failed: %v", err)
	}

	defaultRuns, err := repo.AutomationRunList(ctx, 20)
	if err != nil {
		t.Fatalf("default AutomationRunList failed: %v", err)
	}
	tenantRuns, err := repo.AutomationRunList(tenantCtx, 20)
	if err != nil {
		t.Fatalf("tenant AutomationRunList failed: %v", err)
	}
	if len(defaultRuns) != 1 || defaultRuns[0].ID != defaultRun.ID || defaultRuns[0].SourceName != "Default Feed" {
		t.Fatalf("unexpected default runs: %#v", defaultRuns)
	}
	if len(tenantRuns) != 1 || tenantRuns[0].ID != tenantRun.ID || tenantRuns[0].SourceName != "Tenant Feed" {
		t.Fatalf("unexpected tenant runs: %#v", tenantRuns)
	}

	tenantRun.Status = string(model.AutomationRunPartial)
	if err := repo.AutomationRunUpdate(ctx, tenantRun); err != nil {
		t.Fatalf("default AutomationRunUpdate failed: %v", err)
	}
	tenantRuns, err = repo.AutomationRunList(tenantCtx, 20)
	if err != nil {
		t.Fatalf("tenant AutomationRunList after cross update failed: %v", err)
	}
	if tenantRuns[0].Status != string(model.AutomationRunSuccess) {
		t.Fatalf("default context updated tenant run: %#v", tenantRuns[0])
	}
}
