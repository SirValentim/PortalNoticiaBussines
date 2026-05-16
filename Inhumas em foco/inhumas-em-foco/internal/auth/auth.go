package auth

import (
	"context"
	"errors"
	"net/http"

	"inhumas-em-foco/internal/model"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("credenciais invalidas")
	ErrUserInactive       = errors.New("usuario inativo")
	ErrUnauthorized       = errors.New("nao autorizado")
	ErrPasswordTooWeak    = errors.New("senha muito fraca: minimo 8 caracteres")
)

type Service struct {
	userRepo UserRepository
}

type UserRepository interface {
	UserGetByEmail(ctx context.Context, email string) (*model.User, error)
	UserGetByID(ctx context.Context, id int64) (*model.User, error)
	UserCreate(ctx context.Context, user *model.User) error
	UserUpdatePassword(ctx context.Context, id int64, hash string) error
}

func NewService(repo UserRepository) *Service {
	return &Service{userRepo: repo}
}

func (s *Service) Authenticate(ctx context.Context, email, password string, bcryptCost int) (*model.User, error) {
	user, err := s.userRepo.UserGetByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.Active {
		return nil, ErrUserInactive
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (s *Service) HashPassword(password string, cost int) (string, error) {
	if len(password) < 8 {
		return "", ErrPasswordTooWeak
	}
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), cost)
	return string(bytes), err
}

type Permission string

const (
	PermPostsCreate       Permission = "posts:create"
	PermPostsEditOwn      Permission = "posts:edit_own"
	PermPostsEditAny      Permission = "posts:edit_any"
	PermPostsDelete       Permission = "posts:delete"
	PermPostsPublish      Permission = "posts:publish"
	PermPostsApprove      Permission = "posts:approve"
	PermStoresManage      Permission = "stores:manage"
	PermInfluencersManage Permission = "influencers:manage"
	PermBannersManage     Permission = "banners:manage"
	PermPromosManage      Permission = "promos:manage"
	PermEventsManage      Permission = "events:manage"
	PermClassifiedsManage Permission = "classifieds:manage"
	PermMediaManage       Permission = "media:manage"
	PermUsersManage       Permission = "users:manage"
	PermSettingsManage    Permission = "settings:manage"
	PermAutomationManage  Permission = "automation:manage"
)

func (s *Service) HasPermission(user *model.User, perm Permission) bool {
	if user == nil {
		return false
	}
	perms := RolePermissions(user.Role)
	for _, p := range perms {
		if p == perm || p == "*" {
			return true
		}
	}
	return false
}

func RolePermissions(role model.UserRole) []Permission {
	switch role {
	case model.RoleSuperAdmin:
		return []Permission{"*"}
	case model.RoleAdmin:
		return []Permission{"*"}
	case model.RoleEditor:
		return []Permission{
			PermPostsCreate, PermPostsEditAny, PermPostsDelete, PermPostsPublish, PermPostsApprove, PermEventsManage, PermMediaManage, PermAutomationManage,
		}
	case model.RoleRedator:
		return []Permission{
			PermPostsCreate, PermPostsEditOwn, PermMediaManage,
		}
	case model.RoleRevisor:
		return []Permission{
			PermPostsEditAny, PermPostsApprove,
		}
	case model.RoleComercial:
		return []Permission{
			PermStoresManage, PermInfluencersManage, PermBannersManage, PermPromosManage, PermEventsManage, PermClassifiedsManage, PermMediaManage,
		}
	default:
		return []Permission{}
	}
}

func UserFromContext(ctx context.Context) *model.User {
	if u, ok := ctx.Value(userContextKey).(*model.User); ok {
		return u
	}
	return nil
}

func RequirePermission(perm Permission) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := UserFromContext(r.Context())
			if user == nil {
				http.Error(w, "Autenticacao necessaria", http.StatusUnauthorized)
				return
			}
			authSvc, ok := r.Context().Value("authService").(*Service)
			if !ok || authSvc == nil {
				http.Error(w, "Servico de autenticacao indisponivel", http.StatusInternalServerError)
				return
			}
			if !authSvc.HasPermission(user, perm) {
				http.Error(w, "Acesso negado", http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

type contextKey string

const userContextKey contextKey = "user"

func WithUser(ctx context.Context, user *model.User) context.Context {
	return context.WithValue(ctx, userContextKey, user)
}
