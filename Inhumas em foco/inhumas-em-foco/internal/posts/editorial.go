package posts

import (
	"strings"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
)

type EditorialService struct {
	authSvc *auth.Service
}

func NewEditorialService(authSvc *auth.Service) *EditorialService {
	return &EditorialService{authSvc: authSvc}
}

func (s *EditorialService) CanEdit(user *model.User, post *model.Post) bool {
	if user == nil || post == nil || s.authSvc == nil {
		return false
	}
	if s.authSvc.HasPermission(user, auth.PermPostsEditAny) {
		return true
	}
	if !s.authSvc.HasPermission(user, auth.PermPostsEditOwn) || post.AuthorID == nil || *post.AuthorID != user.ID {
		return false
	}
	return post.Status == model.StatusDraft || post.Status == model.StatusReview
}

func (s *EditorialService) ValidateStatusPermission(user *model.User, status model.PostStatus) string {
	if s.authSvc == nil {
		return "Servico de permissao indisponivel"
	}
	switch status {
	case model.StatusApproved:
		if !s.authSvc.HasPermission(user, auth.PermPostsApprove) {
			return "Seu perfil nao pode aprovar noticias."
		}
	case model.StatusPublished, model.StatusScheduled:
		if !s.authSvc.HasPermission(user, auth.PermPostsPublish) {
			return "Seu perfil pode salvar rascunhos ou enviar para revisao, mas nao publicar/agendar noticias."
		}
	case model.StatusArchived:
		if !s.authSvc.HasPermission(user, auth.PermPostsDelete) {
			return "Seu perfil nao pode arquivar noticias."
		}
	}
	return ""
}

func (s *EditorialService) ValidateForm(post *model.Post) string {
	if post == nil {
		return "Noticia invalida"
	}
	if strings.TrimSpace(post.Title) == "" {
		return "Titulo e obrigatorio"
	}
	if strings.TrimSpace(post.Content) == "" {
		return "Conteudo e obrigatorio"
	}
	switch post.Status {
	case model.StatusDraft, model.StatusReview, model.StatusApproved, model.StatusScheduled, model.StatusPublished, model.StatusArchived:
		return ""
	default:
		return "Status invalido"
	}
}

func (s *EditorialService) CanSubmitReview(user *model.User, post *model.Post) bool {
	if !s.CanEdit(user, post) {
		return false
	}
	return post.Status == model.StatusDraft
}

func (s *EditorialService) CanApprove(user *model.User, post *model.Post) bool {
	if user == nil || post == nil || s.authSvc == nil {
		return false
	}
	return post.Status == model.StatusReview && s.authSvc.HasPermission(user, auth.PermPostsApprove)
}

func (s *EditorialService) CanReject(user *model.User, post *model.Post) bool {
	return s.CanApprove(user, post)
}

func (s *EditorialService) ValidateRequiredEditorialNotes(post *model.Post, required bool) string {
	if !required || post == nil {
		return ""
	}
	if strings.TrimSpace(post.EditorialNotes) == "" || strings.TrimSpace(post.EditorResponsible) == "" {
		return "Apuracao e responsavel editorial sao obrigatorios para Politica & Bastidores"
	}
	return ""
}
