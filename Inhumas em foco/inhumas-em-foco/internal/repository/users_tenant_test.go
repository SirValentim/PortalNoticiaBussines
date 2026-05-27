package repository

import (
	"context"
	"testing"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func TestUsersAreScopedByTenant(t *testing.T) {
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

	defaultUser := &model.User{
		Name:         "Default Admin",
		Email:        "admin@example.com",
		PasswordHash: "default-hash",
		Role:         model.RoleAdmin,
		Active:       true,
	}
	if err := repo.UserCreate(ctx, defaultUser); err != nil {
		t.Fatalf("default UserCreate failed: %v", err)
	}

	tenantUser := &model.User{
		Name:         "Tenant Admin",
		Email:        "admin@example.com",
		PasswordHash: "tenant-hash",
		Role:         model.RoleAdmin,
		Active:       true,
	}
	if err := repo.UserCreate(tenantCtx, tenantUser); err != nil {
		t.Fatalf("tenant UserCreate failed: %v", err)
	}

	if defaultUser.TenantID != 1 {
		t.Fatalf("default user tenant_id = %d, want 1", defaultUser.TenantID)
	}
	if tenantUser.TenantID != tenant.ID {
		t.Fatalf("tenant user tenant_id = %d, want %d", tenantUser.TenantID, tenant.ID)
	}

	gotDefault, err := repo.UserGetByEmail(ctx, "ADMIN@example.com")
	if err != nil || gotDefault == nil {
		t.Fatalf("default UserGetByEmail failed: user=%#v err=%v", gotDefault, err)
	}
	if gotDefault.ID != defaultUser.ID || gotDefault.PasswordHash != "default-hash" {
		t.Fatalf("default lookup returned wrong user: %#v", gotDefault)
	}

	gotTenant, err := repo.UserGetByEmail(tenantCtx, "admin@example.com")
	if err != nil || gotTenant == nil {
		t.Fatalf("tenant UserGetByEmail failed: user=%#v err=%v", gotTenant, err)
	}
	if gotTenant.ID != tenantUser.ID || gotTenant.PasswordHash != "tenant-hash" {
		t.Fatalf("tenant lookup returned wrong user: %#v", gotTenant)
	}

	if got, err := repo.UserGetByID(ctx, tenantUser.ID); err == nil || got != nil {
		t.Fatalf("default context accessed tenant user: user=%#v err=%v", got, err)
	}
	if got, err := repo.UserGetByID(tenantCtx, tenantUser.ID); err != nil || got == nil {
		t.Fatalf("tenant context could not access tenant user: user=%#v err=%v", got, err)
	}

	defaultUsers, err := repo.UserList(ctx)
	if err != nil {
		t.Fatalf("default UserList failed: %v", err)
	}
	tenantUsers, err := repo.UserList(tenantCtx)
	if err != nil {
		t.Fatalf("tenant UserList failed: %v", err)
	}
	if len(defaultUsers) != 1 || defaultUsers[0].ID != defaultUser.ID {
		t.Fatalf("unexpected default users: %#v", defaultUsers)
	}
	if len(tenantUsers) != 1 || tenantUsers[0].ID != tenantUser.ID {
		t.Fatalf("unexpected tenant users: %#v", tenantUsers)
	}
}

func TestUserPasswordUpdateDoesNotCrossTenants(t *testing.T) {
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
		Name:         "Tenant Editor",
		Email:        "editor@example.com",
		PasswordHash: "old-hash",
		Role:         model.RoleEditor,
		Active:       true,
	}
	if err := repo.UserCreate(tenantCtx, user); err != nil {
		t.Fatalf("UserCreate failed: %v", err)
	}

	if err := repo.UserUpdatePassword(ctx, user.ID, "wrong-tenant-hash"); err != nil {
		t.Fatalf("default UserUpdatePassword failed: %v", err)
	}
	got, err := repo.UserGetByID(tenantCtx, user.ID)
	if err != nil {
		t.Fatalf("UserGetByID failed: %v", err)
	}
	if got.PasswordHash != "old-hash" {
		t.Fatalf("password changed across tenants: %q", got.PasswordHash)
	}

	if err := repo.UserUpdatePassword(tenantCtx, user.ID, "new-hash"); err != nil {
		t.Fatalf("tenant UserUpdatePassword failed: %v", err)
	}
	got, err = repo.UserGetByID(tenantCtx, user.ID)
	if err != nil {
		t.Fatalf("UserGetByID after update failed: %v", err)
	}
	if got.PasswordHash != "new-hash" {
		t.Fatalf("password was not updated in tenant context: %q", got.PasswordHash)
	}
}

func TestUserLookupUsesTenantUserRole(t *testing.T) {
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
		Name:         "Multi Portal",
		Email:        "multi@example.com",
		PasswordHash: "hash",
		Role:         model.RoleEditor,
		Active:       true,
	}
	if err := repo.UserCreate(ctx, user); err != nil {
		t.Fatalf("UserCreate failed: %v", err)
	}
	if err := repo.TenantUserUpsert(tenantCtx, &model.TenantUser{
		UserID: user.ID,
		Role:   model.RoleAdmin,
		Active: true,
	}); err != nil {
		t.Fatalf("TenantUserUpsert failed: %v", err)
	}

	defaultUser, err := repo.UserGetByEmail(ctx, "multi@example.com")
	if err != nil || defaultUser == nil {
		t.Fatalf("default UserGetByEmail failed: user=%#v err=%v", defaultUser, err)
	}
	if defaultUser.TenantID != 1 || defaultUser.Role != model.RoleEditor {
		t.Fatalf("unexpected default effective user: %#v", defaultUser)
	}

	tenantUser, err := repo.UserGetByEmail(tenantCtx, "multi@example.com")
	if err != nil || tenantUser == nil {
		t.Fatalf("tenant UserGetByEmail failed: user=%#v err=%v", tenantUser, err)
	}
	if tenantUser.TenantID != tenant.ID || tenantUser.ID != user.ID || tenantUser.Role != model.RoleAdmin {
		t.Fatalf("unexpected tenant effective user: %#v", tenantUser)
	}

	list, err := repo.UserList(tenantCtx)
	if err != nil {
		t.Fatalf("tenant UserList failed: %v", err)
	}
	if len(list) != 1 || list[0].ID != user.ID || list[0].Role != model.RoleAdmin {
		t.Fatalf("unexpected tenant user list: %#v", list)
	}
}

func TestUserSoftDeleteIsScopedToTenantLink(t *testing.T) {
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
		Name:         "Multi Portal",
		Email:        "multi-delete@example.com",
		PasswordHash: "hash",
		Role:         model.RoleSuperAdmin,
		Active:       true,
	}
	if err := repo.UserCreate(ctx, user); err != nil {
		t.Fatalf("UserCreate failed: %v", err)
	}
	if err := repo.TenantUserUpsert(tenantCtx, &model.TenantUser{
		UserID: user.ID,
		Role:   model.RoleAdmin,
		Active: true,
	}); err != nil {
		t.Fatalf("TenantUserUpsert failed: %v", err)
	}

	count, err := repo.UserActiveSuperAdminCount(ctx)
	if err != nil || count != 1 {
		t.Fatalf("UserActiveSuperAdminCount = %d err=%v, want 1", count, err)
	}
	if err := repo.UserSoftDelete(tenantCtx, user.ID); err != nil {
		t.Fatalf("UserSoftDelete failed: %v", err)
	}

	defaultUser, err := repo.UserGetByID(ctx, user.ID)
	if err != nil || defaultUser == nil || !defaultUser.Active {
		t.Fatalf("default user should remain active: user=%#v err=%v", defaultUser, err)
	}
	tenantUser, err := repo.UserGetByID(tenantCtx, user.ID)
	if err != nil || tenantUser == nil || tenantUser.Active {
		t.Fatalf("tenant link should be inactive: user=%#v err=%v", tenantUser, err)
	}
}
