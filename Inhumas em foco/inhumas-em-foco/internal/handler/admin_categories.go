package handler

import (
	"net/http"
	"strconv"
	"strings"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/slug"
	"inhumas-em-foco/internal/storage"
)

func (h *Handler) AdminCategories(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermSettingsManage); !ok {
		return
	}
	categories, _ := h.repo.CategoryList(r.Context())
	h.Render(w, r, "admin_categories.html", map[string]any{
		"Title":      "Categorias",
		"Active":     "categories",
		"Categories": categories,
	})
}

func (h *Handler) AdminCategoryNew(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermSettingsManage); !ok {
		return
	}
	h.Render(w, r, "admin_category_form.html", map[string]any{
		"Title":       "Nova Categoria",
		"Active":      "categories",
		"Category":    nil,
		"MediaAssets": h.mediaAssetsForForms(r.Context()),
	})
}

func (h *Handler) AdminCategoryCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermSettingsManage)
	if !ok {
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.renderCategoryForm(w, r, "Tamanho maximo excedido", nil)
		return
	}

	category := h.categoryFromRequest(r, nil)
	if msg := h.validateCategory(category); msg != "" {
		h.renderCategoryForm(w, r, msg, category)
		return
	}
	category.Slug = h.categorySlugFromRequest(r, category.Name, 0)
	if category.Slug == "" {
		h.renderCategoryForm(w, r, "Slug invalido", category)
		return
	}
	if key := h.mediaKeyFromRequest(r, "image_media_key"); key != "" {
		category.ImageKey = key
	}
	h.applyCategoryUpload(r, user, category)

	if err := h.repo.CategoryCreate(r.Context(), category); err != nil {
		h.renderCategoryForm(w, r, err.Error(), category)
		return
	}
	h.auditAdminAction(r, user, "create", "category", auditEntityID(category.ID), map[string]any{
		"name":       category.Name,
		"slug":       category.Slug,
		"active":     category.Active,
		"sort_order": category.SortOrder,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/categories", http.StatusSeeOther)
}

func (h *Handler) AdminCategoryEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermSettingsManage); !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	category, err := h.repo.CategoryGetByID(r.Context(), id)
	if err != nil || category == nil {
		http.NotFound(w, r)
		return
	}
	h.Render(w, r, "admin_category_form.html", map[string]any{
		"Title":       "Editar Categoria",
		"Active":      "categories",
		"Category":    category,
		"MediaAssets": h.mediaAssetsForForms(r.Context()),
	})
}

func (h *Handler) AdminCategoryUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermSettingsManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	category, err := h.repo.CategoryGetByID(r.Context(), id)
	if err != nil || category == nil {
		http.NotFound(w, r)
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.renderCategoryForm(w, r, "Tamanho maximo excedido", category)
		return
	}

	oldSlug := category.Slug
	category = h.categoryFromRequest(r, category)
	if msg := h.validateCategory(category); msg != "" {
		h.renderCategoryForm(w, r, msg, category)
		return
	}
	category.Slug = h.categorySlugFromRequest(r, category.Name, category.ID)
	if category.Slug == "" {
		category.Slug = oldSlug
		h.renderCategoryForm(w, r, "Slug invalido ou ja utilizado", category)
		return
	}

	if r.FormValue("remove_image") == "on" && category.ImageKey != "" {
		h.storage.Delete(r.Context(), category.ImageKey)
		category.ImageKey = ""
	}
	if key := h.mediaKeyFromRequest(r, "image_media_key"); key != "" {
		category.ImageKey = key
	}
	h.applyCategoryUpload(r, user, category)

	if err := h.repo.CategoryUpdate(r.Context(), category); err != nil {
		h.renderCategoryForm(w, r, err.Error(), category)
		return
	}
	h.auditAdminAction(r, user, "update", "category", auditEntityID(category.ID), map[string]any{
		"name":       category.Name,
		"slug":       category.Slug,
		"old_slug":   oldSlug,
		"active":     category.Active,
		"sort_order": category.SortOrder,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/categories", http.StatusSeeOther)
}

func (h *Handler) AdminCategoryDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermSettingsManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	category, _ := h.repo.CategoryGetByID(r.Context(), id)
	count, _ := h.repo.CategoryPostCount(r.Context(), id)
	if count > 0 {
		categories, _ := h.repo.CategoryList(r.Context())
		h.Render(w, r, "admin_categories.html", map[string]any{
			"Title":      "Categorias",
			"Active":     "categories",
			"Error":      "Nao e possivel excluir categoria com materias vinculadas.",
			"Categories": categories,
		})
		return
	}
	if category != nil && category.ImageKey != "" {
		h.storage.Delete(r.Context(), category.ImageKey)
	}
	h.repo.CategoryDelete(r.Context(), id)
	h.auditAdminAction(r, user, "delete", "category", auditEntityID(id), map[string]any{
		"name": categoryName(category),
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/categories", http.StatusSeeOther)
}

func (h *Handler) categoryFromRequest(r *http.Request, category *model.Category) *model.Category {
	if category == nil {
		category = &model.Category{Active: true}
	}
	category.Name = strings.TrimSpace(r.FormValue("name"))
	category.Description = strings.TrimSpace(r.FormValue("description"))
	category.MetaTitle = strings.TrimSpace(r.FormValue("meta_title"))
	category.MetaDescription = strings.TrimSpace(r.FormValue("meta_description"))
	category.SortOrder, _ = strconv.Atoi(r.FormValue("sort_order"))
	category.Active = r.FormValue("active") == "on"
	category.RequiresEditorialNotes = r.FormValue("requires_editorial_notes") == "on"
	return category
}

func (h *Handler) categorySlugFromRequest(r *http.Request, name string, currentID int64) string {
	base := strings.TrimSpace(r.FormValue("slug"))
	if base == "" {
		base = name
	}
	generated := slug.Generate(base)
	if generated == "" {
		return ""
	}
	existing, _ := h.repo.CategoryGetBySlug(r.Context(), generated)
	if existing == nil || existing.ID == currentID {
		return generated
	}
	if strings.TrimSpace(r.FormValue("slug")) != "" {
		return ""
	}
	return slug.Unique(name, func(candidate string) bool {
		existing, _ := h.repo.CategoryGetBySlug(r.Context(), candidate)
		return existing != nil && existing.ID != currentID
	})
}

func (h *Handler) applyCategoryUpload(r *http.Request, user *model.User, category *model.Category) {
	file, header, err := r.FormFile("image")
	if err != nil {
		return
	}
	defer file.Close()
	key := storage.GenerateKey(header.Filename)
	if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
		category.ImageKey = info.Key
		h.recordMediaAssetFromUpload(r, user, header, info, category.Name)
	}
}

func (h *Handler) validateCategory(category *model.Category) string {
	if strings.TrimSpace(category.Name) == "" {
		return "Nome e obrigatorio"
	}
	return ""
}

func (h *Handler) renderCategoryForm(w http.ResponseWriter, r *http.Request, errorMessage string, category *model.Category) {
	h.Render(w, r, "admin_category_form.html", map[string]any{
		"Title":       "Categoria",
		"Active":      "categories",
		"Error":       errorMessage,
		"Category":    category,
		"MediaAssets": h.mediaAssetsForForms(r.Context()),
	})
}

func categoryName(category *model.Category) string {
	if category == nil {
		return ""
	}
	return category.Name
}
