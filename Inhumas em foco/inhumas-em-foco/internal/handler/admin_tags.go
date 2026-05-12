package handler

import (
	"net/http"
	"strconv"
	"strings"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/slug"
)

func (h *Handler) AdminTags(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermSettingsManage); !ok {
		return
	}
	tags, _ := h.repo.TagList(r.Context(), false)
	h.Render(w, r, "admin_tags.html", map[string]any{
		"Title":  "Tags",
		"Active": "tags",
		"Tags":   tags,
	})
}

func (h *Handler) AdminTagNew(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermSettingsManage); !ok {
		return
	}
	h.Render(w, r, "admin_tag_form.html", map[string]any{
		"Title":  "Nova Tag",
		"Active": "tags",
		"Tag":    nil,
	})
}

func (h *Handler) AdminTagCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermSettingsManage)
	if !ok {
		return
	}
	if err := r.ParseForm(); err != nil {
		h.renderTagForm(w, r, "Formulario invalido", nil)
		return
	}
	tag := h.tagFromRequest(r, nil)
	if msg := validateTag(tag); msg != "" {
		h.renderTagForm(w, r, msg, tag)
		return
	}
	tag.Slug = h.tagSlugFromRequest(r, tag.Name, 0)
	if tag.Slug == "" {
		h.renderTagForm(w, r, "Slug invalido ou ja utilizado", tag)
		return
	}
	if err := h.repo.TagCreate(r.Context(), tag); err != nil {
		h.renderTagForm(w, r, err.Error(), tag)
		return
	}
	h.auditAdminAction(r, user, "create", "tag", auditEntityID(tag.ID), map[string]any{
		"name":   tag.Name,
		"slug":   tag.Slug,
		"active": tag.Active,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/tags", http.StatusSeeOther)
}

func (h *Handler) AdminTagEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermSettingsManage); !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	tag, err := h.repo.TagGetByID(r.Context(), id)
	if err != nil || tag == nil {
		http.NotFound(w, r)
		return
	}
	h.Render(w, r, "admin_tag_form.html", map[string]any{
		"Title":  "Editar Tag",
		"Active": "tags",
		"Tag":    tag,
	})
}

func (h *Handler) AdminTagUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermSettingsManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	tag, err := h.repo.TagGetByID(r.Context(), id)
	if err != nil || tag == nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.renderTagForm(w, r, "Formulario invalido", tag)
		return
	}
	oldSlug := tag.Slug
	tag = h.tagFromRequest(r, tag)
	if msg := validateTag(tag); msg != "" {
		h.renderTagForm(w, r, msg, tag)
		return
	}
	tag.Slug = h.tagSlugFromRequest(r, tag.Name, tag.ID)
	if tag.Slug == "" {
		tag.Slug = oldSlug
		h.renderTagForm(w, r, "Slug invalido ou ja utilizado", tag)
		return
	}
	if err := h.repo.TagUpdate(r.Context(), tag); err != nil {
		h.renderTagForm(w, r, err.Error(), tag)
		return
	}
	h.auditAdminAction(r, user, "update", "tag", auditEntityID(tag.ID), map[string]any{
		"name":     tag.Name,
		"slug":     tag.Slug,
		"old_slug": oldSlug,
		"active":   tag.Active,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/tags", http.StatusSeeOther)
}

func (h *Handler) AdminTagDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermSettingsManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	tag, _ := h.repo.TagGetByID(r.Context(), id)
	count, _ := h.repo.TagPostCount(r.Context(), id)
	if count > 0 {
		tags, _ := h.repo.TagList(r.Context(), false)
		h.Render(w, r, "admin_tags.html", map[string]any{
			"Title":  "Tags",
			"Active": "tags",
			"Error":  "Nao e possivel excluir tag vinculada a noticias.",
			"Tags":   tags,
		})
		return
	}
	h.repo.TagDelete(r.Context(), id)
	h.auditAdminAction(r, user, "delete", "tag", auditEntityID(id), map[string]any{
		"name": tagName(tag),
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/tags", http.StatusSeeOther)
}

func (h *Handler) tagFromRequest(r *http.Request, tag *model.Tag) *model.Tag {
	if tag == nil {
		tag = &model.Tag{Active: true}
	}
	tag.Name = strings.TrimSpace(r.FormValue("name"))
	tag.Description = strings.TrimSpace(r.FormValue("description"))
	tag.MetaTitle = strings.TrimSpace(r.FormValue("meta_title"))
	tag.MetaDescription = strings.TrimSpace(r.FormValue("meta_description"))
	tag.Active = r.FormValue("active") == "on"
	return tag
}

func (h *Handler) tagSlugFromRequest(r *http.Request, name string, currentID int64) string {
	base := strings.TrimSpace(r.FormValue("slug"))
	if base == "" {
		base = name
	}
	generated := slug.Generate(base)
	if generated == "" {
		return ""
	}
	existing, _ := h.repo.TagGetBySlug(r.Context(), generated)
	if existing == nil || existing.ID == currentID {
		return generated
	}
	if strings.TrimSpace(r.FormValue("slug")) != "" {
		return ""
	}
	return slug.Unique(name, func(candidate string) bool {
		existing, _ := h.repo.TagGetBySlug(r.Context(), candidate)
		return existing != nil && existing.ID != currentID
	})
}

func validateTag(tag *model.Tag) string {
	if strings.TrimSpace(tag.Name) == "" {
		return "Nome e obrigatorio"
	}
	return ""
}

func (h *Handler) renderTagForm(w http.ResponseWriter, r *http.Request, errorMessage string, tag *model.Tag) {
	h.Render(w, r, "admin_tag_form.html", map[string]any{
		"Title":  "Tag",
		"Active": "tags",
		"Error":  errorMessage,
		"Tag":    tag,
	})
}

func tagName(tag *model.Tag) string {
	if tag == nil {
		return ""
	}
	return tag.Name
}
