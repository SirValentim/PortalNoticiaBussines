package repository

import (
	"context"
	"testing"
	"time"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestPromotionsEventsClassifiedsAndInfluencersAreScopedByTenant(t *testing.T) {
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

	defaultStore := &model.Store{Slug: "loja", Name: "Loja Default", Active: true}
	tenantStore := &model.Store{Slug: "loja", Name: "Loja Tenant", Active: true}
	if err := repo.StoreCreate(ctx, defaultStore); err != nil {
		t.Fatalf("default StoreCreate failed: %v", err)
	}
	if err := repo.StoreCreate(tenantCtx, tenantStore); err != nil {
		t.Fatalf("tenant StoreCreate failed: %v", err)
	}

	start := time.Now().Add(-24 * time.Hour)
	end := time.Now().Add(24 * time.Hour)
	defaultPromo := &model.Promotion{StoreID: defaultStore.ID, Slug: "oferta", Title: "Oferta Default", StartDate: start, EndDate: end, Status: "active"}
	tenantPromo := &model.Promotion{StoreID: tenantStore.ID, Slug: "oferta", Title: "Oferta Tenant", StartDate: start, EndDate: end, Status: "active"}
	if err := repo.PromotionCreate(ctx, defaultPromo); err != nil {
		t.Fatalf("default PromotionCreate failed: %v", err)
	}
	if err := repo.PromotionCreate(tenantCtx, tenantPromo); err != nil {
		t.Fatalf("tenant PromotionCreate failed: %v", err)
	}

	defaultEvent := &model.Event{Slug: "show", Title: "Show Default", Status: "active", StartAt: start}
	tenantEvent := &model.Event{Slug: "show", Title: "Show Tenant", Status: "active", StartAt: start}
	if err := repo.EventCreate(ctx, defaultEvent); err != nil {
		t.Fatalf("default EventCreate failed: %v", err)
	}
	if err := repo.EventCreate(tenantCtx, tenantEvent); err != nil {
		t.Fatalf("tenant EventCreate failed: %v", err)
	}

	defaultClassified := &model.Classified{Slug: "vaga", Title: "Vaga Default", Status: "active"}
	tenantClassified := &model.Classified{Slug: "vaga", Title: "Vaga Tenant", Status: "active"}
	if err := repo.ClassifiedCreate(ctx, defaultClassified); err != nil {
		t.Fatalf("default ClassifiedCreate failed: %v", err)
	}
	if err := repo.ClassifiedCreate(tenantCtx, tenantClassified); err != nil {
		t.Fatalf("tenant ClassifiedCreate failed: %v", err)
	}

	defaultInfluencer := &model.Influencer{Slug: "perfil", Name: "Perfil Default", Active: true}
	tenantInfluencer := &model.Influencer{Slug: "perfil", Name: "Perfil Tenant", Active: true}
	if err := repo.InfluencerCreate(ctx, defaultInfluencer); err != nil {
		t.Fatalf("default InfluencerCreate failed: %v", err)
	}
	if err := repo.InfluencerCreate(tenantCtx, tenantInfluencer); err != nil {
		t.Fatalf("tenant InfluencerCreate failed: %v", err)
	}

	assertLookup := func(name string, gotDefault string, gotTenant string) {
		t.Helper()
		if gotDefault == gotTenant {
			t.Fatalf("%s lookup returned same title/name across tenants: %q", name, gotDefault)
		}
	}

	gotDefaultPromo, err := repo.PromotionGetBySlug(ctx, "oferta")
	if err != nil || gotDefaultPromo == nil {
		t.Fatalf("default PromotionGetBySlug failed: promo=%#v err=%v", gotDefaultPromo, err)
	}
	gotTenantPromo, err := repo.PromotionGetBySlug(tenantCtx, "oferta")
	if err != nil || gotTenantPromo == nil {
		t.Fatalf("tenant PromotionGetBySlug failed: promo=%#v err=%v", gotTenantPromo, err)
	}
	assertLookup("promotion", gotDefaultPromo.Title, gotTenantPromo.Title)

	gotDefaultEvent, err := repo.EventGetBySlug(ctx, "show")
	if err != nil || gotDefaultEvent == nil {
		t.Fatalf("default EventGetBySlug failed: event=%#v err=%v", gotDefaultEvent, err)
	}
	gotTenantEvent, err := repo.EventGetBySlug(tenantCtx, "show")
	if err != nil || gotTenantEvent == nil {
		t.Fatalf("tenant EventGetBySlug failed: event=%#v err=%v", gotTenantEvent, err)
	}
	assertLookup("event", gotDefaultEvent.Title, gotTenantEvent.Title)

	gotDefaultClassified, err := repo.ClassifiedGetBySlug(ctx, "vaga")
	if err != nil || gotDefaultClassified == nil {
		t.Fatalf("default ClassifiedGetBySlug failed: classified=%#v err=%v", gotDefaultClassified, err)
	}
	gotTenantClassified, err := repo.ClassifiedGetBySlug(tenantCtx, "vaga")
	if err != nil || gotTenantClassified == nil {
		t.Fatalf("tenant ClassifiedGetBySlug failed: classified=%#v err=%v", gotTenantClassified, err)
	}
	assertLookup("classified", gotDefaultClassified.Title, gotTenantClassified.Title)

	gotDefaultInfluencer, err := repo.InfluencerGetBySlug(ctx, "perfil")
	if err != nil || gotDefaultInfluencer == nil {
		t.Fatalf("default InfluencerGetBySlug failed: influencer=%#v err=%v", gotDefaultInfluencer, err)
	}
	gotTenantInfluencer, err := repo.InfluencerGetBySlug(tenantCtx, "perfil")
	if err != nil || gotTenantInfluencer == nil {
		t.Fatalf("tenant InfluencerGetBySlug failed: influencer=%#v err=%v", gotTenantInfluencer, err)
	}
	assertLookup("influencer", gotDefaultInfluencer.Name, gotTenantInfluencer.Name)

	if got, err := repo.PromotionGetByID(ctx, tenantPromo.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant promotion: promo=%#v err=%v", got, err)
	}
	if got, err := repo.EventGetByID(ctx, tenantEvent.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant event: event=%#v err=%v", got, err)
	}
	if got, err := repo.ClassifiedGetByID(ctx, tenantClassified.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant classified: classified=%#v err=%v", got, err)
	}
	if got, err := repo.InfluencerGetByID(ctx, tenantInfluencer.ID); err != nil || got != nil {
		t.Fatalf("default context accessed tenant influencer: influencer=%#v err=%v", got, err)
	}
}
