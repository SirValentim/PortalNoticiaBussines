package auth

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"inhumas-em-foco/internal/model"
)

type fakeUserRepo struct {
	user *model.User
	err  error
}

func (r fakeUserRepo) UserGetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.user, r.err
}

func (r fakeUserRepo) UserGetByID(ctx context.Context, id int64) (*model.User, error) {
	return r.user, r.err
}

func (r fakeUserRepo) UserCreate(ctx context.Context, user *model.User) error {
	return nil
}

func (r fakeUserRepo) UserUpdatePassword(ctx context.Context, id int64, hash string) error {
	return nil
}

func TestAuthenticate(t *testing.T) {
	svc := NewService(fakeUserRepo{})
	hash, err := svc.HashPassword("senha-forte", 4)
	if err != nil {
		t.Fatalf("HashPassword failed: %v", err)
	}

	tests := []struct {
		name     string
		repo     fakeUserRepo
		password string
		wantErr  error
	}{
		{
			name:     "valid credentials",
			repo:     fakeUserRepo{user: &model.User{Email: "admin@example.com", PasswordHash: hash, Active: true}},
			password: "senha-forte",
		},
		{
			name:     "missing user",
			repo:     fakeUserRepo{err: errors.New("not found")},
			password: "senha-forte",
			wantErr:  ErrInvalidCredentials,
		},
		{
			name:     "inactive user",
			repo:     fakeUserRepo{user: &model.User{Email: "admin@example.com", PasswordHash: hash, Active: false}},
			password: "senha-forte",
			wantErr:  ErrUserInactive,
		},
		{
			name:     "wrong password",
			repo:     fakeUserRepo{user: &model.User{Email: "admin@example.com", PasswordHash: hash, Active: true}},
			password: "errada",
			wantErr:  ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := NewService(tt.repo)
			_, err := svc.Authenticate(context.Background(), "admin@example.com", tt.password, 4)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Authenticate error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestRolePermissions(t *testing.T) {
	svc := NewService(fakeUserRepo{})
	superAdmin := &model.User{Role: model.RoleSuperAdmin}
	admin := &model.User{Role: model.RoleAdmin}
	editor := &model.User{Role: model.RoleEditor}
	redator := &model.User{Role: model.RoleRedator}
	revisor := &model.User{Role: model.RoleRevisor}
	comercial := &model.User{Role: model.RoleComercial}

	if !svc.HasPermission(superAdmin, PermSettingsManage) {
		t.Fatal("super admin should have all permissions")
	}
	if !svc.HasPermission(admin, PermUsersManage) {
		t.Fatal("admin should have all permissions")
	}
	if !svc.HasPermission(editor, PermPostsPublish) {
		t.Fatal("editor should publish posts")
	}
	if !svc.HasPermission(editor, PermPostsApprove) {
		t.Fatal("editor should approve posts")
	}
	if svc.HasPermission(editor, PermUsersManage) {
		t.Fatal("editor should not manage users")
	}
	if !svc.HasPermission(redator, PermPostsCreate) || !svc.HasPermission(redator, PermPostsEditOwn) {
		t.Fatal("redator should create and edit own posts")
	}
	if svc.HasPermission(redator, PermPostsPublish) || svc.HasPermission(redator, PermPostsEditAny) {
		t.Fatal("redator should not publish or edit any post")
	}
	if !svc.HasPermission(revisor, PermPostsApprove) || !svc.HasPermission(revisor, PermPostsEditAny) {
		t.Fatal("revisor should review and approve posts")
	}
	if svc.HasPermission(revisor, PermPostsPublish) {
		t.Fatal("revisor should not publish posts")
	}
	if !svc.HasPermission(comercial, PermStoresManage) || !svc.HasPermission(comercial, PermInfluencersManage) {
		t.Fatal("comercial should manage stores and influencers")
	}
	if svc.HasPermission(comercial, PermPostsEditAny) {
		t.Fatal("comercial should not edit posts")
	}
}

func TestRequirePermissionWithoutAuthServiceDoesNotPanic(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(WithUser(req.Context(), &model.User{Role: model.RoleAdmin}))
	rec := httptest.NewRecorder()

	handler := RequirePermission(PermUsersManage)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}
