package users

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"strings"
	"time"

	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/mailer"
	"inhumas-em-foco/internal/model"
)

const (
	PasswordResetEmailSent          = "sent"
	PasswordResetEmailError         = "error"
	PasswordResetEmailNotConfigured = "not_configured"
)

type Repository interface {
	UserGetByEmail(ctx context.Context, email string) (*model.User, error)
	UserGetByID(ctx context.Context, id int64) (*model.User, error)
	UserUpdatePassword(ctx context.Context, id int64, hash string) error
	PasswordResetInvalidateUser(ctx context.Context, userID int64) error
	PasswordResetCreate(ctx context.Context, token *model.PasswordResetToken) error
	PasswordResetGetActive(ctx context.Context, tokenHash string, now time.Time) (*model.PasswordResetToken, error)
	PasswordResetMarkUsed(ctx context.Context, id int64) error
}

type PasswordHasher interface {
	HashPassword(password string, cost int) (string, error)
}

type PasswordResetService struct {
	repo   Repository
	hasher PasswordHasher
	cfg    *config.Config
}

type PasswordResetRequestResult struct {
	User        *model.User
	ResetURL    string
	EmailStatus string
	EmailError  string
}

func NewPasswordResetService(repo Repository, hasher PasswordHasher, cfg *config.Config) *PasswordResetService {
	return &PasswordResetService{repo: repo, hasher: hasher, cfg: cfg}
}

func (s *PasswordResetService) Request(ctx context.Context, email, ip, userAgent string) PasswordResetRequestResult {
	email = strings.TrimSpace(strings.ToLower(email))
	user, err := s.repo.UserGetByEmail(ctx, email)
	if err != nil || user == nil || !user.Active {
		return PasswordResetRequestResult{}
	}

	plainToken, tokenHash, err := NewPasswordResetToken()
	if err != nil {
		return PasswordResetRequestResult{User: user}
	}
	_ = s.repo.PasswordResetInvalidateUser(ctx, user.ID)
	reset := &model.PasswordResetToken{
		UserID:      user.ID,
		TokenHash:   tokenHash,
		RequestedIP: ip,
		UserAgent:   userAgent,
		ExpiresAt:   time.Now().Add(30 * time.Minute),
	}
	if err := s.repo.PasswordResetCreate(ctx, reset); err != nil {
		return PasswordResetRequestResult{User: user}
	}

	resetURL := s.ResetURL(plainToken)
	result := PasswordResetRequestResult{
		User:     user,
		ResetURL: resetURL,
	}
	if s.cfg != nil && s.cfg.SMTPEnabled() {
		if err := mailer.SendPasswordReset(ctx, s.cfg, user.Email, resetURL); err != nil {
			result.EmailStatus = PasswordResetEmailError
			result.EmailError = err.Error()
		} else {
			result.EmailStatus = PasswordResetEmailSent
		}
	} else {
		result.EmailStatus = PasswordResetEmailNotConfigured
	}
	return result
}

func (s *PasswordResetService) Context(ctx context.Context, tokenValue string) (*model.PasswordResetToken, *model.User, bool) {
	if strings.TrimSpace(tokenValue) == "" {
		return nil, nil, false
	}
	reset, err := s.repo.PasswordResetGetActive(ctx, PasswordResetHash(tokenValue), time.Now())
	if err != nil || reset == nil {
		return nil, nil, false
	}
	user, err := s.repo.UserGetByID(ctx, reset.UserID)
	if err != nil || user == nil || !user.Active {
		return nil, nil, false
	}
	return reset, user, true
}

func (s *PasswordResetService) Complete(ctx context.Context, tokenValue, password, confirm string) (*model.User, string, bool) {
	reset, user, ok := s.Context(ctx, tokenValue)
	if !ok || reset == nil || user == nil {
		return nil, "Link invalido ou expirado. Solicite uma nova recuperacao de senha.", false
	}
	if password == "" || password != confirm {
		return user, "Senha e confirmacao precisam ser iguais.", false
	}
	hash, err := s.hasher.HashPassword(password, s.cfg.DefaultBcryptCost)
	if err != nil {
		return user, "Senha invalida: use pelo menos 8 caracteres.", false
	}
	if err := s.repo.UserUpdatePassword(ctx, user.ID, hash); err != nil {
		return user, "Nao foi possivel atualizar a senha.", false
	}
	_ = s.repo.PasswordResetMarkUsed(ctx, reset.ID)
	return user, "", true
}

func (s *PasswordResetService) ResetURL(token string) string {
	siteURL := ""
	if s.cfg != nil {
		siteURL = strings.TrimRight(s.cfg.SiteURL, "/")
	}
	if siteURL == "" {
		return "/redefinir-senha/" + token
	}
	return siteURL + "/redefinir-senha/" + token
}

func NewPasswordResetToken() (string, string, error) {
	var b [32]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", "", err
	}
	token := base64.RawURLEncoding.EncodeToString(b[:])
	return token, PasswordResetHash(token), nil
}

func PasswordResetHash(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
