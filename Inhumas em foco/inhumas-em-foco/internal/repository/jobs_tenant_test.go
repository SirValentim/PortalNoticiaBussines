package repository

import (
	"context"
	"testing"
	"time"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestJobsCarryTenantContext(t *testing.T) {
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

	job := &model.Job{
		Type:        model.JobCollectNews,
		Payload:     "{}",
		Status:      model.JobPending,
		RunAt:       time.Now().Add(-time.Minute),
		MaxAttempts: 1,
	}
	if err := repo.JobCreate(tenantCtx, job); err != nil {
		t.Fatalf("JobCreate failed: %v", err)
	}
	if job.TenantID != tenant.ID {
		t.Fatalf("job tenant_id = %d, want %d", job.TenantID, tenant.ID)
	}

	defaultHas, err := repo.JobHasActiveType(ctx, model.JobCollectNews)
	if err != nil {
		t.Fatalf("default JobHasActiveType failed: %v", err)
	}
	tenantHas, err := repo.JobHasActiveType(tenantCtx, model.JobCollectNews)
	if err != nil {
		t.Fatalf("tenant JobHasActiveType failed: %v", err)
	}
	if defaultHas || !tenantHas {
		t.Fatalf("unexpected active job scope: default=%v tenant=%v", defaultHas, tenantHas)
	}

	claimed, err := repo.JobClaimPending(ctx, 10)
	if err != nil {
		t.Fatalf("JobClaimPending failed: %v", err)
	}
	if len(claimed) != 1 || claimed[0].TenantID != tenant.ID {
		t.Fatalf("claimed job did not preserve tenant: %#v", claimed)
	}
}

func TestDeadJobsCarryTenantContext(t *testing.T) {
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

	job := &model.Job{
		Type:        model.JobCollectNews,
		Payload:     "{}",
		Status:      model.JobPending,
		RunAt:       time.Now().Add(-time.Minute),
		MaxAttempts: 1,
	}
	if err := repo.JobCreate(tenantCtx, job); err != nil {
		t.Fatalf("JobCreate failed: %v", err)
	}
	moved, err := repo.JobRecordFailure(ctx, model.Job{ID: job.ID, TenantID: job.TenantID, Attempts: 0, MaxAttempts: 1}, "falha permanente")
	if err != nil {
		t.Fatalf("JobRecordFailure failed: %v", err)
	}
	if !moved {
		t.Fatal("job was not moved to dead_jobs")
	}

	deadJobs, err := repo.DeadJobList(ctx, 10)
	if err != nil {
		t.Fatalf("DeadJobList failed: %v", err)
	}
	if len(deadJobs) != 1 || deadJobs[0].TenantID != tenant.ID {
		t.Fatalf("dead job did not preserve tenant: %#v", deadJobs)
	}
}
