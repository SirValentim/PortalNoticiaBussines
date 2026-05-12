package users

import (
	"context"
	"strings"
	"testing"
	"time"

	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/model"
)

type fakeResetRepo struct {
	user  *model.User
	token *model.PasswordResetToken
	hash  string
}

func (r *fakeResetRepo) UserGetByEmail(ctx context.Context, email string) (*model.User, error) {
	return r.user, nil
}

func (r *fakeResetRepo) UserGetByID(ctx context.Context, id int64) (*model.User, error) {
	return r.user, nil
}

func (r *fakeResetRepo) UserUpdatePassword(ctx context.Context, id int64, hash string) error {
	r.hash = hash
	return nil
}

func (r *fakeResetRepo) PasswordResetInvalidateUser(ctx context.Context, userID int64) error {
	return nil
}

func (r *fakeResetRepo) PasswordResetCreate(ctx context.Context, token *model.PasswordResetToken) error {
	r.token = token
	token.ID = 1
	return nil
}

func (r *fakeResetRepo) PasswordResetGetActive(ctx context.Context, tokenHash string, now time.Time) (*model.PasswordResetToken, error) {
	if r.token == nil || r.token.TokenHash != tokenHash || r.token.UsedAt != nil || !r.token.ExpiresAt.After(now) {
		return nil, nil
	}
	return r.token, nil
}

func (r *fakeResetRepo) PasswordResetMarkUsed(ctx context.Context, id int64) error {
	now := time.Now()
	r.token.UsedAt = &now
	return nil
}

type fakeHasher struct{}

func (fakeHasher) HashPassword(password string, cost int) (string, error) {
	return "hash:" + password, nil
}

func TestPasswordResetServiceRequestCreatesTokenWithoutSMTP(t *testing.T) {
	repo := &fakeResetRepo{user: &model.User{ID: 7, Email: "editor@example.com", Active: true}}
	svc := NewPasswordResetService(repo, fakeHasher{}, &config.Config{SiteURL: "https://example.com"})

	result := svc.Request(context.Background(), "EDITOR@example.com", "127.0.0.1", "test")

	if result.User == nil || result.EmailStatus != PasswordResetEmailNotConfigured {
		t.Fatalf("unexpected result: %#v", result)
	}
	if repo.token == nil || repo.token.UserID != 7 || repo.token.TokenHash == "" {
		t.Fatalf("token was not created: %#v", repo.token)
	}
	if !strings.HasPrefix(result.ResetURL, "https://example.com/redefinir-senha/") {
		t.Fatalf("unexpected reset URL: %q", result.ResetURL)
	}
}

func TestPasswordResetServiceCompleteConsumesToken(t *testing.T) {
	plain := "token-de-teste"
	repo := &fakeResetRepo{
		user: &model.User{ID: 7, Email: "editor@example.com", Active: true},
		token: &model.PasswordResetToken{
			ID:        1,
			UserID:    7,
			TokenHash: PasswordResetHash(plain),
			ExpiresAt: time.Now().Add(time.Hour),
		},
	}
	svc := NewPasswordResetService(repo, fakeHasher{}, &config.Config{DefaultBcryptCost: 4})

	user, msg, ok := svc.Complete(context.Background(), plain, "senha-nova", "senha-nova")

	if !ok || msg != "" || user == nil {
		t.Fatalf("complete = user=%#v msg=%q ok=%v", user, msg, ok)
	}
	if repo.hash != "hash:senha-nova" {
		t.Fatalf("password hash not updated: %q", repo.hash)
	}
	if repo.token.UsedAt == nil {
		t.Fatal("token was not marked as used")
	}
}
