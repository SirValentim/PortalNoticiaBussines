package repository

import (
	"context"
	"testing"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestTenantUsersAreScopedByTenant(t *testing.T) {
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

	user := &model.User{
		Name:         "Portal User",
		Email:        "portal@example.com",
		PasswordHash: "hash",
		Role:         model.RoleEditor,
		Active:       true,
	}
	if err := repo.UserCreate(ctx, user); err != nil {
		t.Fatalf("UserCreate failed: %v", err)
	}

	defaultLink := &model.TenantUser{
		UserID: user.ID,
		Role:   model.RoleEditor,
		Active: true,
	}
	if err := repo.TenantUserUpsert(ctx, defaultLink); err != nil {
		t.Fatalf("default TenantUserUpsert failed: %v", err)
	}
	tenantLink := &model.TenantUser{
		UserID: user.ID,
		Role:   model.RoleAdmin,
		Active: true,
	}
	if err := repo.TenantUserUpsert(tenantCtx, tenantLink); err != nil {
		t.Fatalf("tenant TenantUserUpsert failed: %v", err)
	}

	gotDefault, err := repo.TenantUserGet(ctx, user.ID)
	if err != nil || gotDefault == nil {
		t.Fatalf("default TenantUserGet failed: link=%#v err=%v", gotDefault, err)
	}
	if gotDefault.TenantID != 1 || gotDefault.Role != model.RoleEditor || gotDefault.UserEmail != user.Email {
		t.Fatalf("unexpected default link: %#v", gotDefault)
	}

	gotTenant, err := repo.TenantUserGet(tenantCtx, user.ID)
	if err != nil || gotTenant == nil {
		t.Fatalf("tenant TenantUserGet failed: link=%#v err=%v", gotTenant, err)
	}
	if gotTenant.TenantID != tenant.ID || gotTenant.Role != model.RoleAdmin || gotTenant.UserEmail != user.Email {
		t.Fatalf("unexpected tenant link: %#v", gotTenant)
	}

	defaultUsers, err := repo.TenantUserList(ctx)
	if err != nil {
		t.Fatalf("default TenantUserList failed: %v", err)
	}
	tenantUsers, err := repo.TenantUserList(tenantCtx)
	if err != nil {
		t.Fatalf("tenant TenantUserList failed: %v", err)
	}
	if len(defaultUsers) != 1 || defaultUsers[0].ID != defaultLink.ID {
		t.Fatalf("unexpected default tenant users: %#v", defaultUsers)
	}
	if len(tenantUsers) != 1 || tenantUsers[0].ID != tenantLink.ID {
		t.Fatalf("unexpected tenant users: %#v", tenantUsers)
	}

	allLinks, err := repo.TenantUserListByUser(ctx, user.ID)
	if err != nil {
		t.Fatalf("TenantUserListByUser failed: %v", err)
	}
	if len(allLinks) != 2 {
		t.Fatalf("unexpected link count for user: %#v", allLinks)
	}
}

func TestTenantUserUpsertAndDelete(t *testing.T) {
	repo, err := New(":memory:")
	if err != nil {
		t.Fatalf("repository.New failed: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	user := &model.User{
		Name:         "Editor",
		Email:        "editor@example.com",
		PasswordHash: "hash",
		Role:         model.RoleEditor,
		Active:       true,
	}
	if err := repo.UserCreate(ctx, user); err != nil {
		t.Fatalf("UserCreate failed: %v", err)
	}

	link := &model.TenantUser{UserID: user.ID, Role: model.RoleEditor, Active: true}
	if err := repo.TenantUserUpsert(ctx, link); err != nil {
		t.Fatalf("TenantUserUpsert create failed: %v", err)
	}
	linkID := link.ID

	link.Role = model.RoleComercial
	link.Active = false
	if err := repo.TenantUserUpsert(ctx, link); err != nil {
		t.Fatalf("TenantUserUpsert update failed: %v", err)
	}
	if link.ID != linkID {
		t.Fatalf("link id changed after upsert: got %d want %d", link.ID, linkID)
	}

	got, err := repo.TenantUserGet(ctx, user.ID)
	if err != nil || got == nil {
		t.Fatalf("TenantUserGet failed: link=%#v err=%v", got, err)
	}
	if got.Role != model.RoleComercial || got.Active {
		t.Fatalf("tenant user was not updated: %#v", got)
	}

	if err := repo.TenantUserDelete(ctx, user.ID); err != nil {
		t.Fatalf("TenantUserDelete failed: %v", err)
	}
	got, err = repo.TenantUserGet(ctx, user.ID)
	if err != nil {
		t.Fatalf("TenantUserGet after delete failed: %v", err)
	}
	if got != nil {
		t.Fatalf("tenant user was not deleted: %#v", got)
	}
}
