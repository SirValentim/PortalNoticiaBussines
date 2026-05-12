package users

import (
	"context"
	"strings"

	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/model"
)

type AdminRepository interface {
	UserCreate(ctx context.Context, user *model.User) error
	UserUpdate(ctx context.Context, user *model.User) error
	UserUpdatePassword(ctx context.Context, id int64, hash string) error
}

type AdminService struct {
	repo   AdminRepository
	hasher PasswordHasher
	cfg    *config.Config
}

type UserCreateInput struct {
	Name     string
	Email    string
	Password string
	Role     model.UserRole
	Active   bool
}

type UserUpdateInput struct {
	Name   string
	Email  string
	Role   model.UserRole
	Active bool
}

func NewAdminService(repo AdminRepository, hasher PasswordHasher, cfg *config.Config) *AdminService {
	return &AdminService{repo: repo, hasher: hasher, cfg: cfg}
}

func (s *AdminService) Create(ctx context.Context, input UserCreateInput) (*model.User, string) {
	if !input.Role.IsValid() {
		return nil, "Perfil invalido"
	}
	name := strings.TrimSpace(input.Name)
	email := strings.TrimSpace(strings.ToLower(input.Email))
	if name == "" || email == "" {
		return nil, "Nome e email sao obrigatorios"
	}
	hash, err := s.hasher.HashPassword(input.Password, s.bcryptCost())
	if err != nil {
		return nil, "Senha invalida"
	}
	user := &model.User{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
		Role:         input.Role,
		Active:       input.Active,
	}
	if err := s.repo.UserCreate(ctx, user); err != nil {
		return nil, "Nao foi possivel criar o usuario"
	}
	return user, ""
}

func (s *AdminService) Update(ctx context.Context, currentUser, target *model.User, input UserUpdateInput) string {
	if target == nil {
		return "Usuario nao encontrado"
	}
	if !input.Role.IsValid() {
		return "Perfil invalido"
	}
	if currentUser != nil && currentUser.ID == target.ID {
		if !input.Active {
			return "Voce nao pode desativar sua propria conta"
		}
		if input.Role != model.RoleSuperAdmin && input.Role != model.RoleAdmin {
			return "Voce nao pode remover seu proprio perfil administrativo"
		}
	}

	target.Name = strings.TrimSpace(input.Name)
	target.Email = strings.TrimSpace(strings.ToLower(input.Email))
	target.Role = input.Role
	target.Active = input.Active
	if target.Name == "" || target.Email == "" {
		return "Nome e email sao obrigatorios"
	}
	if err := s.repo.UserUpdate(ctx, target); err != nil {
		return "Nao foi possivel atualizar o usuario"
	}
	return ""
}

func (s *AdminService) UpdatePassword(ctx context.Context, id int64, password, confirm string) string {
	if password == "" || password != confirm {
		return "Senha e confirmacao precisam ser iguais"
	}
	hash, err := s.hasher.HashPassword(password, s.bcryptCost())
	if err != nil {
		return "Senha invalida: use pelo menos 8 caracteres"
	}
	if err := s.repo.UserUpdatePassword(ctx, id, hash); err != nil {
		return "Erro ao atualizar senha"
	}
	return ""
}

func (s *AdminService) bcryptCost() int {
	if s.cfg == nil || s.cfg.DefaultBcryptCost == 0 {
		return 12
	}
	return s.cfg.DefaultBcryptCost
}
