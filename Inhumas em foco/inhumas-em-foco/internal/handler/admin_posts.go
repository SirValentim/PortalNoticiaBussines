package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/slug"
	"inhumas-em-foco/internal/storage"
)

func (h *Handler) adminPostFormData(ctx context.Context, user *model.User, post *model.Post, errorMsg string) map[string]any {
	cats, _ := h.repo.CategoryList(ctx)
	tags, _ := h.repo.TagList(ctx, true)
	var revisions []model.PostRevision
	var aiLogs []model.AIUsageLog
	isEditing := post != nil && post.ID > 0
	if isEditing {
		revisions, _ = h.repo.PostRevisionList(ctx, post.ID, 12)
		aiLogs, _ = h.repo.AIUsageLogListForPost(ctx, post.ID, 8)
	}
	data := map[string]any{
		"Categories":     cats,
		"Tags":           tags,
		"SelectedTagIDs": selectedTagIDs(post),
		"SEOChecklist":   seoChecklist(post),
		"Revisions":      revisions,
		"AIUsageLogs":    aiLogs,
		"Post":           post,
		"IsEditing":      isEditing,
		"MediaAssets":    h.mediaAssetsForForms(ctx),
		"CanPublish":     h.authSvc.HasPermission(user, auth.PermPostsPublish),
		"CanApprove":     h.authSvc.HasPermission(user, auth.PermPostsApprove),
		"CanArchive":     h.authSvc.HasPermission(user, auth.PermPostsDelete),
	}
	if errorMsg != "" {
		data["Error"] = errorMsg
	}
	return data
}

func (h *Handler) AdminPosts(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		page, _ = strconv.Atoi(p)
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	categoryID, _ := strconv.ParseInt(r.URL.Query().Get("category_id"), 10, 64)
	tagID, _ := strconv.ParseInt(r.URL.Query().Get("tag_id"), 10, 64)
	limit := 20
	offset := (page - 1) * limit

	posts, _ := h.repo.PostListAdmin(r.Context(), status, categoryID, tagID, query, limit, offset)
	count, _ := h.repo.PostCountAdmin(r.Context(), status, categoryID, tagID, query)

	cats, _ := h.repo.CategoryList(r.Context())
	tags, _ := h.repo.TagList(r.Context(), true)
	catMap := make(map[int64]string)
	for _, c := range cats {
		catMap[c.ID] = c.Name
	}

	h.Render(w, r, "admin_posts.html", map[string]any{
		"Posts":            posts,
		"Page":             page,
		"TotalPages":       (count + limit - 1) / limit,
		"CatMap":           catMap,
		"Categories":       cats,
		"Tags":             tags,
		"FilterStatus":     status,
		"FilterCategoryID": categoryID,
		"FilterTagID":      tagID,
		"FilterQuery":      query,
		"FilterQueryPath":  adminPostFilterPath(status, categoryID, tagID, query),
		"CanCreate":        h.authSvc.HasPermission(user, auth.PermPostsCreate),
		"CanPublish":       h.authSvc.HasPermission(user, auth.PermPostsPublish),
		"CanApprove":       h.authSvc.HasPermission(user, auth.PermPostsApprove),
		"CanDelete":        h.authSvc.HasPermission(user, auth.PermPostsDelete),
		"Active":           "posts",
	})
}

func (h *Handler) AdminPostNew(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermPostsCreate)
	if !ok {
		return
	}
	h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, nil, ""))
}

func (h *Handler) AdminPostDetail(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if !h.canViewPostDetail(user, post) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}
	revisions, _ := h.repo.PostRevisionList(r.Context(), post.ID, 12)
	auditLogs, _ := h.repo.AuditLogList(r.Context(), "post", post.ID, 12)
	views, _ := h.repo.MetricCountByEntity(r.Context(), "post_view", "post", post.ID)
	clicks, _ := h.repo.MetricCountByEntity(r.Context(), "post_click", "post", post.ID)
	shareClicks, _ := h.repo.MetricCountByEntity(r.Context(), "share_click", "post", post.ID)
	tenant, _ := h.repo.TenantGetByID(r.Context(), post.TenantID)
	tenantName := h.branding(r.Context()).PortalName
	if tenant != nil && tenant.Name != "" {
		tenantName = tenant.Name
	}
	h.Render(w, r, "admin_post_detail.html", map[string]any{
		"Title":      "Materia",
		"Active":     "posts",
		"Post":       post,
		"TenantName": tenantName,
		"Views":      views,
		"Clicks":     clicks + shareClicks,
		"Revisions":  revisions,
		"AuditLogs":  auditLogs,
		"PublicURL":  h.siteURL(r.Context()) + "/noticia/" + post.Slug,
		"CanEdit":    h.postSvc.CanEdit(user, post),
		"CanPublish": h.authSvc.HasPermission(user, auth.PermPostsPublish),
		"CanDelete":  h.authSvc.HasPermission(user, auth.PermPostsDelete),
		"CanCreate":  h.authSvc.HasPermission(user, auth.PermPostsCreate),
	})
}

func (h *Handler) canViewPostDetail(user *model.User, post *model.Post) bool {
	if h.postSvc.CanEdit(user, post) {
		return true
	}
	return h.authSvc.HasPermission(user, auth.PermPostsApprove) ||
		h.authSvc.HasPermission(user, auth.PermPostsPublish) ||
		h.authSvc.HasPermission(user, auth.PermPostsDelete)
}

func (h *Handler) AdminPostCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermPostsCreate)
	if !ok {
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, nil, "Tamanho maximo excedido"))
		return
	}

	post := &model.Post{
		Title:             r.FormValue("title"),
		Excerpt:           r.FormValue("excerpt"),
		Content:           r.FormValue("content"),
		MetaTitle:         r.FormValue("meta_title"),
		MetaDescription:   r.FormValue("meta_description"),
		SEOKeyword:        r.FormValue("seo_keyword"),
		CanonicalURL:      r.FormValue("canonical_url"),
		SourceName:        r.FormValue("source_name"),
		SourceURL:         r.FormValue("source_url"),
		Status:            model.PostStatus(r.FormValue("status")),
		IsSponsored:       r.FormValue("is_sponsored") == "on",
		IsFeatured:        r.FormValue("is_featured") == "on",
		IsPinned:          r.FormValue("is_pinned") == "on",
		EditorialNotes:    r.FormValue("editorial_notes"),
		EditorResponsible: r.FormValue("editor_responsible"),
	}
	post.Tags = h.tagsFromRequest(r)
	post.ReadingTimeMinutes = estimateReadingTime(post.Content)
	if key := h.mediaKeyFromRequest(r, "cover_image_media_key"); key != "" {
		post.CoverImageKey = key
	}
	post.GalleryImageKeys = h.mediaKeysFromRequest(r, "gallery_media_keys", post.GalleryImageKeys)

	if catID, err := strconv.ParseInt(r.FormValue("category_id"), 10, 64); err == nil {
		post.CategoryID = &catID
	}
	post.AuthorID = &user.ID
	if msg := h.postSvc.ValidateStatusPermission(user, post.Status); msg != "" {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, msg))
		return
	}
	if msg := h.postSvc.ValidateForm(post); msg != "" {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, msg))
		return
	}

	cat, _ := h.repo.CategoryGetBySlug(r.Context(), "politica-bastidores")
	if cat != nil && post.CategoryID != nil && *post.CategoryID == cat.ID {
		if msg := h.postSvc.ValidateRequiredEditorialNotes(post, true); msg != "" {
			h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, msg))
			return
		}
	}

	file, header, err := r.FormFile("cover_image")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			post.CoverImageKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, post.Title)
		}
	}
	h.applyPostGalleryUploads(r, user, post)

	post.Slug = slug.Unique(post.Title, func(s string) bool {
		return h.repo.PostSlugExists(r.Context(), s)
	})

	if post.Status == "published" {
		now := time.Now()
		post.PublishedAt = &now
	}
	if publishAt := r.FormValue("publish_at"); publishAt != "" && post.Status == "scheduled" {
		if t, err := time.Parse("2006-01-02T15:04", publishAt); err == nil {
			post.PublishAt = &t
		}
	}
	if post.Status == "scheduled" && post.PublishAt == nil {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, "Informe uma data de publicacao valida para posts agendados"))
		return
	}
	if msg := validatePostPublishReadiness(post); msg != "" {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, msg))
		return
	}

	if err := h.repo.PostCreate(r.Context(), post); err != nil {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, "Erro ao criar post: "+err.Error()))
		return
	}
	_ = h.repo.PostSetTags(r.Context(), post.ID, tagIDs(post.Tags))
	h.createPostRevision(r, user, post, "create")
	if post.IsFeatured {
		_ = h.repo.PostSetFeatured(r.Context(), post.ID, true)
	}

	if post.Status == "scheduled" && post.PublishAt != nil {
		h.repo.JobCreate(r.Context(), &model.Job{
			Type:    model.JobPublishPost,
			Payload: fmt.Sprintf(`{"post_id":%d}`, post.ID),
			Status:  model.JobPending,
			RunAt:   *post.PublishAt,
		})
	}

	h.auditAdminAction(r, user, "create", "post", auditEntityID(post.ID), map[string]any{
		"title":  post.Title,
		"slug":   post.Slug,
		"status": post.Status,
	})

	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/posts", http.StatusSeeOther)
}

func (h *Handler) AdminPostEdit(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if !h.postSvc.CanEdit(user, post) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	lockWarning := h.acquireEditLock(r, "post", id)
	data := h.adminPostFormData(r.Context(), user, post, "")
	data["LockWarning"] = lockWarning
	data["LockHeartbeat"] = h.cfg.AdminPathPrefix + "/posts/" + strconv.FormatInt(id, 10) + "/lock"
	h.Render(w, r, "admin_post_form.html", data)
}

func (h *Handler) AdminPostPreview(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if !h.postSvc.CanEdit(user, post) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}
	related := []model.Post{}
	if post.CategoryID != nil {
		related, _ = h.repo.PostListByCategory(r.Context(), *post.CategoryID, 3)
	}
	sidebarTopBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "sidebar_top")
	sidebarBottomBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "sidebar_bottom")
	seo := model.SEOData{
		Title:        "[Preview] " + firstNonEmpty(post.MetaTitle, h.pageTitle(r.Context(), post.Title)),
		Description:  firstNonEmpty(post.MetaDescription, post.Excerpt),
		URL:          h.siteURL(r.Context()) + "/noticia/" + post.Slug,
		CanonicalURL: firstNonEmpty(post.CanonicalURL, h.siteURL(r.Context())+"/noticia/"+post.Slug),
		Type:         "article",
		NoIndex:      true,
		PublishedAt:  post.PublishedAt,
		ModifiedAt:   &post.UpdatedAt,
		Author:       post.AuthorName,
		Tags:         postSEOTags(post),
	}
	if post.CoverImageKey != "" {
		seo.Image = h.storage.URL(r.Context(), post.CoverImageKey)
	}
	h.Render(w, r, "post_detail.html", map[string]any{
		"SEO":                 seo,
		"JSONLD":              h.articleJSONLD(post, seo),
		"Post":                post,
		"Related":             related,
		"Category":            post.CategoryName,
		"SidebarTopBanner":    sidebarTopBanner,
		"SidebarBottomBanner": sidebarBottomBanner,
		"PreviewMode":         true,
	})
}

func (h *Handler) AdminPostLockHeartbeat(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if !h.postSvc.CanEdit(user, post) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}
	if warning := h.acquireEditLock(r, "post", id); warning != "" {
		http.Error(w, warning, http.StatusConflict)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) AdminPostUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if !h.postSvc.CanEdit(user, post) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}

	oldSlug := post.Slug

	if err := h.parseMultipart(w, r); err != nil {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, "Tamanho maximo excedido"))
		return
	}

	post.Title = r.FormValue("title")
	post.Excerpt = r.FormValue("excerpt")
	post.Content = r.FormValue("content")
	post.MetaTitle = r.FormValue("meta_title")
	post.MetaDescription = r.FormValue("meta_description")
	post.SEOKeyword = r.FormValue("seo_keyword")
	post.CanonicalURL = r.FormValue("canonical_url")
	post.SourceName = r.FormValue("source_name")
	post.SourceURL = r.FormValue("source_url")
	post.Status = model.PostStatus(r.FormValue("status"))
	post.IsSponsored = r.FormValue("is_sponsored") == "on"
	post.IsFeatured = r.FormValue("is_featured") == "on"
	post.IsPinned = r.FormValue("is_pinned") == "on"
	post.EditorialNotes = r.FormValue("editorial_notes")
	post.EditorResponsible = r.FormValue("editor_responsible")
	post.Tags = h.tagsFromRequest(r)
	post.ReadingTimeMinutes = estimateReadingTime(post.Content)
	if key := h.mediaKeyFromRequest(r, "cover_image_media_key"); key != "" {
		post.CoverImageKey = key
	}
	post.GalleryImageKeys = h.mediaKeysFromRequest(r, "gallery_media_keys", post.GalleryImageKeys)

	if catID, err := strconv.ParseInt(r.FormValue("category_id"), 10, 64); err == nil {
		post.CategoryID = &catID
	}
	if msg := h.postSvc.ValidateStatusPermission(user, post.Status); msg != "" {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, msg))
		return
	}
	if msg := h.postSvc.ValidateForm(post); msg != "" {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, msg))
		return
	}

	cat, _ := h.repo.CategoryGetBySlug(r.Context(), "politica-bastidores")
	if cat != nil && post.CategoryID != nil && *post.CategoryID == cat.ID {
		if msg := h.postSvc.ValidateRequiredEditorialNotes(post, true); msg != "" {
			h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, msg))
			return
		}
	}

	if r.FormValue("remove_cover_image") == "on" && post.CoverImageKey != "" {
		h.storage.Delete(r.Context(), post.CoverImageKey)
		post.CoverImageKey = ""
	}

	file, header, err := r.FormFile("cover_image")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			if post.CoverImageKey != "" {
				h.storage.Delete(r.Context(), post.CoverImageKey)
			}
			post.CoverImageKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, post.Title)
		}
	}
	if r.FormValue("remove_gallery") != "" {
		removeSet := map[string]bool{}
		for _, key := range r.Form["remove_gallery"] {
			removeSet[key] = true
		}
		var keep []string
		for _, key := range post.GalleryImageKeys {
			if removeSet[key] {
				h.storage.Delete(r.Context(), key)
				continue
			}
			keep = append(keep, key)
		}
		post.GalleryImageKeys = keep
	}
	h.applyPostGalleryUploads(r, user, post)

	newSlug := slug.Generate(post.Title)
	if newSlug != oldSlug && !h.repo.PostSlugExists(r.Context(), newSlug) {
		h.repo.SlugRedirectCreate(r.Context(), oldSlug, newSlug, "post")
		post.Slug = newSlug
	}

	if post.Status == "published" && post.PublishedAt == nil {
		now := time.Now()
		post.PublishedAt = &now
	}
	post.PublishAt = nil
	if publishAt := r.FormValue("publish_at"); publishAt != "" && post.Status == "scheduled" {
		if t, err := time.Parse("2006-01-02T15:04", publishAt); err == nil {
			post.PublishAt = &t
		}
	}
	if post.Status == "scheduled" && post.PublishAt == nil {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, "Informe uma data de publicacao valida para posts agendados"))
		return
	}
	if msg := validatePostPublishReadiness(post); msg != "" {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, msg))
		return
	}

	if err := h.repo.PostUpdate(r.Context(), post); err != nil {
		h.Render(w, r, "admin_post_form.html", h.adminPostFormData(r.Context(), user, post, "Erro ao atualizar post"))
		return
	}
	_ = h.repo.PostSetTags(r.Context(), post.ID, tagIDs(post.Tags))
	h.createPostRevision(r, user, post, "update")
	if post.IsFeatured {
		_ = h.repo.PostSetFeatured(r.Context(), post.ID, true)
	}
	if post.Status == "scheduled" && post.PublishAt != nil {
		h.repo.JobCreate(r.Context(), &model.Job{
			Type:    model.JobPublishPost,
			Payload: fmt.Sprintf(`{"post_id":%d}`, post.ID),
			Status:  model.JobPending,
			RunAt:   *post.PublishAt,
		})
	}
	h.repo.EditLockDelete(r.Context(), "post", id)

	h.auditAdminAction(r, user, "update", "post", auditEntityID(post.ID), map[string]any{
		"title":       post.Title,
		"slug":        post.Slug,
		"status":      post.Status,
		"is_featured": post.IsFeatured,
	})

	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/posts", http.StatusSeeOther)
}

func (h *Handler) tagsFromRequest(r *http.Request) []model.Tag {
	values := r.Form["tag_ids"]
	tags := make([]model.Tag, 0, len(values))
	seen := map[int64]bool{}
	for _, value := range values {
		id, err := strconv.ParseInt(value, 10, 64)
		if err != nil || id <= 0 || seen[id] {
			continue
		}
		seen[id] = true
		tags = append(tags, model.Tag{ID: id})
	}
	return tags
}

func selectedTagIDs(post *model.Post) map[int64]bool {
	selected := map[int64]bool{}
	if post == nil {
		return selected
	}
	for _, tag := range post.Tags {
		selected[tag.ID] = true
	}
	return selected
}

func tagIDs(tags []model.Tag) []int64 {
	ids := make([]int64, 0, len(tags))
	for _, tag := range tags {
		ids = append(ids, tag.ID)
	}
	return ids
}

func (h *Handler) AdminPostAutosave(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if !h.postSvc.CanEdit(user, post) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		http.Error(w, "Tamanho maximo excedido", http.StatusRequestEntityTooLarge)
		return
	}
	post.Title = r.FormValue("title")
	post.Excerpt = r.FormValue("excerpt")
	post.Content = r.FormValue("content")
	post.MetaTitle = r.FormValue("meta_title")
	post.MetaDescription = r.FormValue("meta_description")
	post.SEOKeyword = r.FormValue("seo_keyword")
	post.CanonicalURL = r.FormValue("canonical_url")
	post.SourceName = r.FormValue("source_name")
	post.SourceURL = r.FormValue("source_url")
	post.EditorialNotes = r.FormValue("editorial_notes")
	post.EditorResponsible = r.FormValue("editor_responsible")
	post.IsSponsored = r.FormValue("is_sponsored") == "on"
	post.IsFeatured = r.FormValue("is_featured") == "on"
	post.IsPinned = r.FormValue("is_pinned") == "on"
	post.ReadingTimeMinutes = estimateReadingTime(post.Content)
	if catID, err := strconv.ParseInt(r.FormValue("category_id"), 10, 64); err == nil {
		post.CategoryID = &catID
	}
	if tags := h.tagsFromRequest(r); len(tags) > 0 || len(r.Form["tag_ids"]) > 0 {
		post.Tags = tags
	}
	if msg := h.postSvc.ValidateForm(post); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if err := h.repo.PostUpdate(r.Context(), post); err != nil {
		http.Error(w, "Erro ao autosalvar", http.StatusInternalServerError)
		return
	}
	_ = h.repo.PostSetTags(r.Context(), post.ID, tagIDs(post.Tags))
	h.createPostRevision(r, user, post, "autosave")
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"ok"}`))
}

func (h *Handler) acquireEditLock(r *http.Request, entityType string, entityID int64) string {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		return ""
	}
	now := time.Now()
	current, _ := h.repo.EditLockGet(r.Context(), entityType, entityID)
	if current != nil && current.UserID != user.ID && current.ExpiresAt.After(now) {
		return "Este conteudo esta sendo editado por outro usuario. Aguarde o lock expirar antes de salvar."
	}
	_ = h.repo.EditLockCreate(r.Context(), &model.EditLock{
		EntityType: entityType,
		EntityID:   entityID,
		UserID:     user.ID,
		LockedAt:   now,
		ExpiresAt:  now.Add(2 * time.Minute),
	})
	return ""
}

func (h *Handler) AdminPostDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermPostsDelete)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, _ := h.repo.PostGetByID(r.Context(), id)
	if post != nil && post.CoverImageKey != "" {
		h.storage.Delete(r.Context(), post.CoverImageKey)
	}
	h.repo.PostDelete(r.Context(), id)

	h.auditAdminAction(r, user, "delete", "post", auditEntityID(id), map[string]any{
		"title": postTitle(post),
	})

	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/posts", http.StatusSeeOther)
}

func (h *Handler) AdminPostArchive(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermPostsDelete)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if err := h.repo.PostUpdateStatus(r.Context(), id, model.StatusArchived); err != nil {
		http.Error(w, "Nao foi possivel arquivar a materia", http.StatusInternalServerError)
		return
	}
	h.auditAdminAction(r, user, "archive", "post", auditEntityID(id), map[string]any{
		"title": post.Title,
		"from":  post.Status,
		"to":    model.StatusArchived,
	})
	post.Status = model.StatusArchived
	h.createPostRevision(r, user, post, "archive")
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/posts/"+strconv.FormatInt(id, 10), http.StatusSeeOther)
}

func (h *Handler) AdminPostDuplicate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermPostsCreate)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	nowTitle := strings.TrimSpace(post.Title + " (copia)")
	duplicate := *post
	duplicate.ID = 0
	duplicate.Title = nowTitle
	duplicate.Status = model.StatusDraft
	duplicate.PublishedAt = nil
	duplicate.PublishAt = nil
	duplicate.AuthorID = &user.ID
	duplicate.IsFeatured = false
	duplicate.IsPinned = false
	duplicate.Slug = slug.Unique(nowTitle, func(candidate string) bool {
		return h.repo.PostSlugExists(r.Context(), candidate)
	})
	if err := h.repo.PostCreate(r.Context(), &duplicate); err != nil {
		http.Error(w, "Nao foi possivel duplicar a materia", http.StatusInternalServerError)
		return
	}
	_ = h.repo.PostSetTags(r.Context(), duplicate.ID, tagIDs(post.Tags))
	h.auditAdminAction(r, user, "duplicate", "post", auditEntityID(duplicate.ID), map[string]any{
		"source_id": post.ID,
		"title":     duplicate.Title,
	})
	h.createPostRevision(r, user, &duplicate, "duplicate")
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/posts/"+strconv.FormatInt(duplicate.ID, 10), http.StatusSeeOther)
}

func (h *Handler) AdminPostSubmitReview(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if !h.postSvc.CanSubmitReview(user, post) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}
	if err := h.repo.PostUpdateStatus(r.Context(), id, model.StatusReview); err != nil {
		h.Render(w, r, "admin_posts.html", map[string]any{"Error": "Erro ao enviar materia para revisao"})
		return
	}
	h.auditAdminAction(r, user, "submit_review", "post", auditEntityID(id), map[string]any{
		"title": post.Title,
		"from":  post.Status,
		"to":    model.StatusReview,
	})
	post.Status = model.StatusReview
	h.createPostRevision(r, user, post, "submit_review")
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/posts?status=review", http.StatusSeeOther)
}

func (h *Handler) AdminPostApprove(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if !h.postSvc.CanApprove(user, post) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}
	if msg := validatePostPublishReadiness(post); msg != "" {
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	if err := h.repo.PostUpdateStatus(r.Context(), id, model.StatusApproved); err != nil {
		h.Render(w, r, "admin_posts.html", map[string]any{"Error": "Erro ao aprovar materia"})
		return
	}
	h.auditAdminAction(r, user, "approve", "post", auditEntityID(id), map[string]any{
		"title": post.Title,
		"from":  post.Status,
		"to":    model.StatusApproved,
	})
	post.Status = model.StatusApproved
	h.createPostRevision(r, user, post, "approve")
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/posts?status=approved", http.StatusSeeOther)
}

func (h *Handler) AdminPostReject(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, err := h.repo.PostGetByID(r.Context(), id)
	if err != nil || post == nil {
		http.NotFound(w, r)
		return
	}
	if !h.postSvc.CanReject(user, post) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Formulario invalido", http.StatusBadRequest)
		return
	}
	comment := strings.TrimSpace(r.FormValue("comment"))
	if comment == "" {
		http.Error(w, "Informe o motivo da reprovacao", http.StatusBadRequest)
		return
	}
	notes := appendEditorialReviewNote(post.EditorialNotes, user.Name, comment, time.Now())
	responsible := post.EditorResponsible
	if strings.TrimSpace(responsible) == "" {
		responsible = user.Name
	}
	if err := h.repo.PostUpdateStatusAndEditorialNotes(r.Context(), id, model.StatusDraft, notes, responsible); err != nil {
		h.Render(w, r, "admin_posts.html", map[string]any{"Error": "Erro ao reprovar materia"})
		return
	}
	h.auditAdminAction(r, user, "reject", "post", auditEntityID(id), map[string]any{
		"title":   post.Title,
		"from":    post.Status,
		"to":      model.StatusDraft,
		"comment": comment,
	})
	post.Status = model.StatusDraft
	post.EditorialNotes = notes
	post.EditorResponsible = responsible
	h.createPostRevision(r, user, post, "reject")
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/posts?status=draft", http.StatusSeeOther)
}

func (h *Handler) AdminPostPublish(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermPostsPublish)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	post, _ := h.repo.PostGetByID(r.Context(), id)
	if post != nil {
		if msg := validatePostPublishReadiness(post); msg != "" {
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
	}
	now := time.Now()
	h.repo.PostSetPublished(r.Context(), id, now)
	h.auditAdminAction(r, user, "publish", "post", auditEntityID(id), map[string]any{
		"published_at": now.Format(time.RFC3339),
	})
	if post != nil {
		post.Status = model.StatusPublished
		post.PublishedAt = &now
		h.createPostRevision(r, user, post, "publish")
	}
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/posts", http.StatusSeeOther)
}

func appendEditorialReviewNote(current string, reviewerName string, comment string, now time.Time) string {
	reviewerName = strings.TrimSpace(reviewerName)
	if reviewerName == "" {
		reviewerName = "Revisor"
	}
	entry := fmt.Sprintf("Reprovado em %s por %s: %s", now.Format("02/01/2006 15:04"), reviewerName, strings.TrimSpace(comment))
	current = strings.TrimSpace(current)
	if current == "" {
		return entry
	}
	return current + "\n\n" + entry
}

func (h *Handler) applyPostGalleryUploads(r *http.Request, user *model.User, post *model.Post) {
	if r.MultipartForm == nil || r.MultipartForm.File == nil {
		return
	}
	for _, header := range r.MultipartForm.File["gallery_images"] {
		file, err := header.Open()
		if err != nil {
			continue
		}
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			post.GalleryImageKeys = append(post.GalleryImageKeys, info.Key)
			h.recordMediaAssetFromUpload(r, user, header, info, post.Title)
		}
		file.Close()
	}
}

func estimateReadingTime(content string) int {
	words := strings.Fields(stripSimpleHTML(content))
	minutes := (len(words) + 199) / 200
	if minutes < 1 {
		return 1
	}
	return minutes
}

func stripSimpleHTML(content string) string {
	replacer := strings.NewReplacer("<", " <", ">", "> ")
	parts := strings.Fields(replacer.Replace(content))
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if strings.HasPrefix(part, "<") && strings.HasSuffix(part, ">") {
			continue
		}
		out = append(out, part)
	}
	return strings.Join(out, " ")
}

func seoChecklist(post *model.Post) []string {
	if post == nil {
		return nil
	}
	var missing []string
	if strings.TrimSpace(post.Title) == "" {
		missing = append(missing, "Titulo")
	}
	if strings.TrimSpace(post.MetaDescription) == "" {
		missing = append(missing, "Meta description")
	}
	if post.CategoryID == nil {
		missing = append(missing, "Categoria")
	}
	if strings.TrimSpace(post.Content) == "" {
		missing = append(missing, "Conteudo")
	}
	if strings.TrimSpace(post.SourceName) == "" && strings.TrimSpace(post.SourceURL) == "" {
		missing = append(missing, "Fonte original")
	}
	return missing
}

func validatePostPublishReadiness(post *model.Post) string {
	if post == nil {
		return "Noticia invalida"
	}
	switch post.Status {
	case model.StatusApproved, model.StatusScheduled, model.StatusPublished:
	default:
		return ""
	}
	missing := seoChecklist(post)
	filtered := missing[:0]
	for _, item := range missing {
		if item == "Fonte original" {
			continue
		}
		filtered = append(filtered, item)
	}
	missing = filtered
	if len(missing) > 0 {
		return "Checklist SEO pendente: " + strings.Join(missing, ", ")
	}
	return ""
}

func (h *Handler) createPostRevision(r *http.Request, user *model.User, post *model.Post, action string) {
	if post == nil {
		return
	}
	var userID *int64
	if user != nil {
		id := user.ID
		userID = &id
	}
	snapshot, _ := json.Marshal(map[string]any{
		"title":                post.Title,
		"excerpt":              post.Excerpt,
		"status":               post.Status,
		"category_id":          post.CategoryID,
		"meta_title":           post.MetaTitle,
		"meta_description":     post.MetaDescription,
		"seo_keyword":          post.SEOKeyword,
		"canonical_url":        post.CanonicalURL,
		"source_name":          post.SourceName,
		"source_url":           post.SourceURL,
		"reading_time_minutes": post.ReadingTimeMinutes,
		"is_featured":          post.IsFeatured,
		"is_pinned":            post.IsPinned,
		"editorial_notes":      post.EditorialNotes,
		"gallery_image_keys":   post.GalleryImageKeys,
	})
	_ = h.repo.PostRevisionCreate(r.Context(), &model.PostRevision{
		PostID:   post.ID,
		UserID:   userID,
		Action:   action,
		Title:    post.Title,
		Status:   post.Status,
		Snapshot: string(snapshot),
	})
}
