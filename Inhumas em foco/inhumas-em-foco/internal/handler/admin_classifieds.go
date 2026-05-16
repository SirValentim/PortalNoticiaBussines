package handler

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
	"inhumas-em-foco/internal/slug"
	"inhumas-em-foco/internal/storage"
)

func (h *Handler) AdminClassifieds(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermClassifiedsManage); !ok {
		return
	}
	classifieds, _ := h.repo.ClassifiedList(r.Context(), repository.ClassifiedFilter{Limit: 300})
	h.Render(w, r, "admin_classifieds.html", map[string]any{
		"Title":       "Classificados",
		"Active":      "classifieds",
		"Classifieds": classifieds,
		"Summary":     localCommercialSummaryClassifieds(classifieds),
	})
}

func (h *Handler) AdminClassifiedNew(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermClassifiedsManage); !ok {
		return
	}
	h.renderClassifiedForm(w, r, nil, "")
}

func (h *Handler) AdminClassifiedCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermClassifiedsManage)
	if !ok {
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.renderClassifiedForm(w, r, nil, "Tamanho maximo excedido")
		return
	}
	classified := h.classifiedFromRequest(r, nil)
	if strings.TrimSpace(classified.Title) == "" {
		h.renderClassifiedForm(w, r, classified, "Titulo e obrigatorio")
		return
	}
	if strings.TrimSpace(classified.Category) == "" {
		h.renderClassifiedForm(w, r, classified, "Categoria e obrigatoria")
		return
	}
	classified.Slug = slug.Unique(classified.Title, func(s string) bool {
		return h.repo.ClassifiedSlugExists(r.Context(), s)
	})
	h.applyClassifiedMediaSelection(r, classified)
	h.applyClassifiedUpload(r, user, classified)
	if err := h.repo.ClassifiedCreate(r.Context(), classified); err != nil {
		h.renderClassifiedForm(w, r, classified, "Nao foi possivel criar o classificado")
		return
	}
	h.auditAdminAction(r, user, "create", "classified", auditEntityID(classified.ID), map[string]any{
		"title":       classified.Title,
		"slug":        classified.Slug,
		"category":    classified.Category,
		"status":      classified.Status,
		"is_featured": classified.IsFeatured,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/classifieds", http.StatusSeeOther)
}

func (h *Handler) AdminClassifiedEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermClassifiedsManage); !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	classified, err := h.repo.ClassifiedGetByID(r.Context(), id)
	if err != nil || classified == nil {
		http.NotFound(w, r)
		return
	}
	h.renderClassifiedForm(w, r, classified, "")
}

func (h *Handler) AdminClassifiedUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermClassifiedsManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	classified, err := h.repo.ClassifiedGetByID(r.Context(), id)
	if err != nil || classified == nil {
		http.NotFound(w, r)
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.renderClassifiedForm(w, r, classified, "Tamanho maximo excedido")
		return
	}
	oldSlug := classified.Slug
	classified = h.classifiedFromRequest(r, classified)
	if strings.TrimSpace(classified.Title) == "" {
		h.renderClassifiedForm(w, r, classified, "Titulo e obrigatorio")
		return
	}
	if strings.TrimSpace(classified.Category) == "" {
		h.renderClassifiedForm(w, r, classified, "Categoria e obrigatoria")
		return
	}
	newSlug := slug.Generate(classified.Title)
	if newSlug != oldSlug {
		classified.Slug = slug.Unique(classified.Title, func(s string) bool {
			return h.repo.ClassifiedSlugExists(r.Context(), s)
		})
	}
	if r.FormValue("remove_image") == "on" {
		classified.ImageKey = ""
	}
	h.applyClassifiedMediaSelection(r, classified)
	h.applyClassifiedUpload(r, user, classified)
	if err := h.repo.ClassifiedUpdate(r.Context(), classified); err != nil {
		h.renderClassifiedForm(w, r, classified, "Nao foi possivel atualizar o classificado")
		return
	}
	h.auditAdminAction(r, user, "update", "classified", auditEntityID(classified.ID), map[string]any{
		"title":       classified.Title,
		"slug":        classified.Slug,
		"category":    classified.Category,
		"status":      classified.Status,
		"is_featured": classified.IsFeatured,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/classifieds", http.StatusSeeOther)
}

func (h *Handler) AdminClassifiedDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermClassifiedsManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	classified, _ := h.repo.ClassifiedGetByID(r.Context(), id)
	if err := h.repo.ClassifiedDelete(r.Context(), id); err != nil {
		h.Render(w, r, "admin_classifieds.html", map[string]any{"Title": "Classificados", "Active": "classifieds", "Error": "Nao foi possivel excluir o classificado"})
		return
	}
	h.auditAdminAction(r, user, "delete", "classified", auditEntityID(id), map[string]any{
		"title": classifiedTitle(classified),
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/classifieds", http.StatusSeeOther)
}

func (h *Handler) renderClassifiedForm(w http.ResponseWriter, r *http.Request, classified *model.Classified, errorMessage string) {
	title := "Novo Classificado"
	if classified != nil && classified.ID > 0 {
		title = "Editar Classificado"
	}
	h.Render(w, r, "admin_classified_form.html", map[string]any{
		"Title":       title,
		"Active":      "classifieds",
		"Classified":  classified,
		"MediaAssets": h.mediaAssetsForForms(r.Context()),
		"Error":       errorMessage,
	})
}

func (h *Handler) classifiedFromRequest(r *http.Request, classified *model.Classified) *model.Classified {
	if classified == nil {
		classified = &model.Classified{}
	}
	classified.Title = strings.TrimSpace(r.FormValue("title"))
	classified.Description = strings.TrimSpace(r.FormValue("description"))
	classified.Category = strings.TrimSpace(r.FormValue("category"))
	classified.PriceDisplay = strings.TrimSpace(r.FormValue("price_display"))
	classified.ContactName = strings.TrimSpace(r.FormValue("contact_name"))
	classified.ContactPhone = strings.TrimSpace(r.FormValue("contact_phone"))
	classified.ContactWhatsapp = strings.TrimSpace(r.FormValue("contact_whatsapp"))
	classified.Location = strings.TrimSpace(r.FormValue("location"))
	classified.Status = normalizeClassifiedStatusForHandler(r.FormValue("status"))
	classified.IsFeatured = r.FormValue("is_featured") == "on"
	classified.IsSponsored = r.FormValue("is_sponsored") == "on"
	classified.MetaTitle = strings.TrimSpace(r.FormValue("meta_title"))
	classified.MetaDescription = strings.TrimSpace(r.FormValue("meta_description"))
	expiresAt := parseDateTimeLocal(r.FormValue("expires_at"))
	if expiresAt.IsZero() {
		classified.ExpiresAt = nil
	} else {
		classified.ExpiresAt = &expiresAt
	}
	return classified
}

func (h *Handler) applyClassifiedMediaSelection(r *http.Request, classified *model.Classified) {
	if key := h.mediaKeyFromRequest(r, "image_media_key"); key != "" {
		classified.ImageKey = key
	}
}

func (h *Handler) applyClassifiedUpload(r *http.Request, user *model.User, classified *model.Classified) {
	file, header, err := r.FormFile("image")
	if err != nil {
		return
	}
	defer file.Close()
	key := storage.GenerateKey(header.Filename)
	if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
		classified.ImageKey = info.Key
		h.recordMediaAssetFromUpload(r, user, header, info, classified.Title)
	}
}

func normalizeClassifiedStatusForHandler(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "draft", "archived", "sold":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "active"
	}
}

func classifiedExpiryLabel(classified *model.Classified) string {
	if classified == nil || classified.ExpiresAt == nil {
		return ""
	}
	return classified.ExpiresAt.Format(time.RFC3339)
}
