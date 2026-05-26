package repository

import (
	"context"
	"testing"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestMetricsAreScopedByTenant(t *testing.T) {
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

	defaultMetrics := []model.Metric{
		{MetricType: "post_view", EntityType: "post", EntityID: 10},
		{MetricType: "post_view", EntityType: "post", EntityID: 10},
		{MetricType: "banner_click", EntityType: "banner", EntityID: 1},
	}
	for i := range defaultMetrics {
		if err := repo.MetricTrack(ctx, &defaultMetrics[i]); err != nil {
			t.Fatalf("default MetricTrack failed: %v", err)
		}
	}
	tenantMetrics := []model.Metric{
		{MetricType: "post_view", EntityType: "post", EntityID: 20},
		{MetricType: "store_view", EntityType: "store", EntityID: 2},
	}
	for i := range tenantMetrics {
		if err := repo.MetricTrack(tenantCtx, &tenantMetrics[i]); err != nil {
			t.Fatalf("tenant MetricTrack failed: %v", err)
		}
	}

	if defaultMetrics[0].TenantID != 1 || tenantMetrics[0].TenantID != tenant.ID {
		t.Fatalf("unexpected metric tenant ids: default=%d tenant=%d want default=1 tenant=%d", defaultMetrics[0].TenantID, tenantMetrics[0].TenantID, tenant.ID)
	}

	defaultPostViews, err := repo.MetricCountByType(ctx, "post_view")
	if err != nil {
		t.Fatalf("default MetricCountByType failed: %v", err)
	}
	tenantPostViews, err := repo.MetricCountByType(tenantCtx, "post_view")
	if err != nil {
		t.Fatalf("tenant MetricCountByType failed: %v", err)
	}
	if defaultPostViews != 2 || tenantPostViews != 1 {
		t.Fatalf("unexpected post view counts: default=%d tenant=%d", defaultPostViews, tenantPostViews)
	}

	defaultTop, err := repo.MetricTopEntities(ctx, "post_view", 10)
	if err != nil {
		t.Fatalf("default MetricTopEntities failed: %v", err)
	}
	tenantTop, err := repo.MetricTopEntities(tenantCtx, "post_view", 10)
	if err != nil {
		t.Fatalf("tenant MetricTopEntities failed: %v", err)
	}
	if len(defaultTop) != 1 || defaultTop[0].EntityID != 10 || defaultTop[0].Total != 2 {
		t.Fatalf("default top = %#v, want post 10 with total 2", defaultTop)
	}
	if len(tenantTop) != 1 || tenantTop[0].EntityID != 20 || tenantTop[0].Total != 1 {
		t.Fatalf("tenant top = %#v, want post 20 with total 1", tenantTop)
	}

	defaultTotals, err := repo.MetricTotals(ctx, 10)
	if err != nil {
		t.Fatalf("default MetricTotals failed: %v", err)
	}
	tenantTotals, err := repo.MetricTotals(tenantCtx, 10)
	if err != nil {
		t.Fatalf("tenant MetricTotals failed: %v", err)
	}
	if len(defaultTotals) != 2 || len(tenantTotals) != 2 {
		t.Fatalf("unexpected totals lengths: default=%#v tenant=%#v", defaultTotals, tenantTotals)
	}
}
