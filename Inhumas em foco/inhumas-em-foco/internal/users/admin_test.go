package users

import (
	"context"
	"testing"

	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/model"
)

type fakeAdminRepo struct {
	created         *model.User
	updated         *model.User
	deletedID       int64
	superAdminCount int
	passwordUpdates map[int64]string
}

func (r *fakeAdminRepo) UserCreate(ctx context.Context, user *model.User) error {
	user.ID = 10
	copyUser := *user
	r.created = &copyUser
	return nil
}

func (r *fakeAdminRepo) UserUpdate(ctx context.Context, user *model.User) error {
	copyUser := *user
	r.updated = &copyUser
	return nil
}

func (r *fakeAdminRepo) UserUpdatePassword(ctx context.Context, id int64, hash string) error {
	if r.passwordUpdates == nil {
		r.passwordUpdates = make(map[int64]string)
	}
	r.passwordUpdates[id] = hash
	return nil
}

func (r *fakeAdminRepo) UserSoftDelete(ctx context.Context, id int64) error {
	r.deletedID = id
	return nil
}

func (r *fakeAdminRepo) UserActiveSuperAdminCount(ctx context.Context) (int, error) {
	if r.superAdminCount == 0 {
		return 2, nil
	}
	return r.superAdminCount, nil
}

func TestAdminServiceCreateNormalizesAndHashesUser(t *testing.T) {
	repo := &fakeAdminRepo{}
	svc := NewAdminService(repo, fakeHasher{}, &config.Config{DefaultBcryptCost: 4})

	user, msg := svc.Create(context.Background(), UserCreateInput{
		Name:     "  Editor Local  ",
		Email:    "EDITOR@EXAMPLE.COM ",
		Password: "senha-forte",
		Role:     model.RoleEditor,
		Active:   true,
	})

	if msg != "" || user == nil {
		t.Fatalf("Create msg=%q user=%#v", msg, user)
	}
	if repo.created.Name != "Editor Local" || repo.created.Email != "editor@example.com" || repo.created.PasswordHash != "hash:senha-forte" {
		t.Fatalf("created user not normalized: %#v", repo.created)
	}
}

func TestAdminServiceUpdateBlocksSelfDeactivationAndRoleRemoval(t *testing.T) {
	repo := &fakeAdminRepo{}
	svc := NewAdminService(repo, fakeHasher{}, &config.Config{DefaultBcryptCost: 4})
	current := &model.User{ID: 1, Role: model.RoleAdmin, Active: true}
	target := &model.User{ID: 1, Role: model.RoleAdmin, Active: true}

	msg := svc.Update(context.Background(), current, target, UserUpdateInput{
		Name:   "Admin",
		Email:  "admin@example.com",
		Role:   model.RoleAdmin,
		Active: false,
	})
	if msg != "Voce nao pode desativar sua propria conta" {
		t.Fatalf("self deactivate msg = %q", msg)
	}

	msg = svc.Update(context.Background(), current, target, UserUpdateInput{
		Name:   "Admin",
		Email:  "admin@example.com",
		Role:   model.RoleEditor,
		Active: true,
	})
	if msg != "Voce nao pode remover seu proprio perfil administrativo" {
		t.Fatalf("self role removal msg = %q", msg)
	}
	if repo.updated != nil {
		t.Fatalf("blocked update should not persist: %#v", repo.updated)
	}
}

func TestAdminServiceDeleteBlocksSelfAndLastSuperAdmin(t *testing.T) {
	repo := &fakeAdminRepo{superAdminCount: 1}
	svc := NewAdminService(repo, fakeHasher{}, &config.Config{DefaultBcryptCost: 4})
	current := &model.User{ID: 1, Role: model.RoleSuperAdmin, Active: true}

	if msg := svc.Delete(context.Background(), current, current); msg != "Voce nao pode excluir sua propria conta" {
		t.Fatalf("self delete msg = %q", msg)
	}

	target := &model.User{ID: 2, Role: model.RoleSuperAdmin, Active: true}
	if msg := svc.Delete(context.Background(), current, target); msg != "Esta conta nao pode ser excluida porque e o ultimo Super Admin." {
		t.Fatalf("last super admin msg = %q", msg)
	}
	if repo.deletedID != 0 {
		t.Fatalf("blocked delete should not persist: %d", repo.deletedID)
	}
}

func TestAdminServiceDeleteSoftDeletesAllowedUser(t *testing.T) {
	repo := &fakeAdminRepo{superAdminCount: 2}
	svc := NewAdminService(repo, fakeHasher{}, &config.Config{DefaultBcryptCost: 4})
	current := &model.User{ID: 1, Role: model.RoleSuperAdmin, Active: true}
	target := &model.User{ID: 2, Role: model.RoleEditor, Active: true}

	if msg := svc.Delete(context.Background(), current, target); msg != "" {
		t.Fatalf("Delete msg = %q", msg)
	}
	if repo.deletedID != 2 || target.Active {
		t.Fatalf("delete not persisted correctly: id=%d target=%#v", repo.deletedID, target)
	}
}

func TestAdminServiceUpdatePasswordValidatesAndHashes(t *testing.T) {
	repo := &fakeAdminRepo{}
	svc := NewAdminService(repo, fakeHasher{}, &config.Config{DefaultBcryptCost: 4})

	if msg := svc.UpdatePassword(context.Background(), 9, "nova-senha", "nova-senha"); msg != "" {
		t.Fatalf("UpdatePassword msg = %q", msg)
	}
	if got := repo.passwordUpdates[9]; got != "hash:nova-senha" {
		t.Fatalf("password update hash = %q", got)
	}
	if msg := svc.UpdatePassword(context.Background(), 9, "a", "b"); msg != "Senha e confirmacao precisam ser iguais" {
		t.Fatalf("mismatch msg = %q", msg)
	}
}
