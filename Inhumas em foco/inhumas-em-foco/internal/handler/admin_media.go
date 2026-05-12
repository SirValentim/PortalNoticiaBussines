package handler

import (
	"context"
	"mime/multipart"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
	"inhumas-em-foco/internal/storage"
)

func (h *Handler) AdminMedia(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermMediaManage); !ok {
		return
	}
	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	filter := mediaAssetFilterFromRequest(r)
	limit := 24
	offset := (page - 1) * limit
	filter.Limit = limit
	filter.Offset = offset
	assets, _ := h.repo.MediaAssetListFiltered(r.Context(), filter)
	count, _ := h.repo.MediaAssetCountFiltered(r.Context(), filter)
	months, _ := h.repo.MediaAssetArchiveMonths(r.Context())

	h.Render(w, r, "admin_media.html", map[string]any{
		"Title":           "Biblioteca de Midia",
		"Active":          "media",
		"Assets":          assets,
		"ArchiveMonths":   months,
		"Page":            page,
		"TotalPages":      (count + limit - 1) / limit,
		"FilterQuery":     filter.Query,
		"FilterMonth":     strings.TrimSpace(r.URL.Query().Get("month")),
		"FilterDateFrom":  strings.TrimSpace(r.URL.Query().Get("date_from")),
		"FilterDateTo":    strings.TrimSpace(r.URL.Query().Get("date_to")),
		"FilterQueryPath": adminMediaFilterPath(r.URL.Query()),
		"Success":         mediaSuccessMessage(r.URL.Query().Get("success")),
	})
}

func (h *Handler) AdminMediaUpload(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermMediaManage)
	if !ok {
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.renderMediaWithError(w, r, "Tamanho maximo excedido")
		return
	}
	altText := strings.TrimSpace(r.FormValue("alt_text"))
	if altText == "" {
		h.renderMediaWithError(w, r, "Informe o texto alternativo da imagem")
		return
	}
	file, header, err := r.FormFile("file")
	if err != nil {
		h.renderMediaWithError(w, r, "Selecione uma imagem para enviar")
		return
	}
	defer file.Close()

	info, err := h.storage.Upload(r.Context(), storage.GenerateKey(header.Filename), file, header.Header.Get("Content-Type"))
	if err != nil {
		h.renderMediaWithError(w, r, "Upload invalido: "+err.Error())
		return
	}
	asset := &model.MediaAsset{
		Key:          info.Key,
		OriginalName: filepath.Base(header.Filename),
		Title:        strings.TrimSpace(r.FormValue("title")),
		AltText:      altText,
		ContentType:  info.ContentType,
		SizeBytes:    info.Size,
	}
	if user != nil {
		id := user.ID
		asset.UploadedBy = &id
	}
	if err := h.repo.MediaAssetCreate(r.Context(), asset); err != nil {
		_ = h.storage.Delete(r.Context(), info.Key)
		h.renderMediaWithError(w, r, "Nao foi possivel registrar a midia")
		return
	}
	h.auditAdminAction(r, user, "create", "media", auditEntityID(asset.ID), map[string]any{
		"key":           asset.Key,
		"original_name": asset.OriginalName,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/media?success=upload", http.StatusSeeOther)
}

func (h *Handler) AdminMediaUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermMediaManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	asset, err := h.repo.MediaAssetGetByID(r.Context(), id)
	if err != nil || asset == nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		h.renderMediaWithError(w, r, "Formulario invalido")
		return
	}
	asset.Title = strings.TrimSpace(r.FormValue("title"))
	asset.AltText = strings.TrimSpace(r.FormValue("alt_text"))
	if asset.AltText == "" {
		h.renderMediaWithError(w, r, "Texto alternativo e obrigatorio")
		return
	}
	if err := h.repo.MediaAssetUpdate(r.Context(), asset); err != nil {
		h.renderMediaWithError(w, r, "Nao foi possivel atualizar a midia")
		return
	}
	h.auditAdminAction(r, user, "update", "media", auditEntityID(asset.ID), map[string]any{
		"key":      asset.Key,
		"title":    asset.Title,
		"alt_text": asset.AltText,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/media?success=update", http.StatusSeeOther)
}

func (h *Handler) AdminMediaDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermMediaManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	asset, err := h.repo.MediaAssetGetByID(r.Context(), id)
	if err != nil || asset == nil {
		http.NotFound(w, r)
		return
	}
	if asset.UsageCount > 0 {
		h.renderMediaWithError(w, r, "Nao e possivel excluir uma midia em uso")
		return
	}
	if err := h.repo.MediaAssetDelete(r.Context(), asset.ID); err != nil {
		h.renderMediaWithError(w, r, "Nao foi possivel excluir a midia")
		return
	}
	_ = h.storage.Delete(r.Context(), asset.Key)
	h.auditAdminAction(r, user, "delete", "media", auditEntityID(asset.ID), map[string]any{
		"key":           asset.Key,
		"original_name": asset.OriginalName,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/media?success=delete", http.StatusSeeOther)
}

func (h *Handler) renderMediaWithError(w http.ResponseWriter, r *http.Request, errorMessage string) {
	page := 1
	filter := mediaAssetFilterFromRequest(r)
	filter.Limit = 24
	assets, _ := h.repo.MediaAssetListFiltered(r.Context(), filter)
	count, _ := h.repo.MediaAssetCountFiltered(r.Context(), filter)
	months, _ := h.repo.MediaAssetArchiveMonths(r.Context())
	h.Render(w, r, "admin_media.html", map[string]any{
		"Title":           "Biblioteca de Midia",
		"Active":          "media",
		"Assets":          assets,
		"ArchiveMonths":   months,
		"Page":            page,
		"TotalPages":      (count + 23) / 24,
		"FilterQuery":     filter.Query,
		"FilterMonth":     strings.TrimSpace(r.URL.Query().Get("month")),
		"FilterDateFrom":  strings.TrimSpace(r.URL.Query().Get("date_from")),
		"FilterDateTo":    strings.TrimSpace(r.URL.Query().Get("date_to")),
		"FilterQueryPath": adminMediaFilterPath(r.URL.Query()),
		"Error":           errorMessage,
	})
}

func adminMediaFilterPath(values url.Values) string {
	copyValues := url.Values{}
	for key, vals := range values {
		if key == "page" || key == "success" {
			continue
		}
		for _, value := range vals {
			if strings.TrimSpace(value) != "" {
				copyValues.Add(key, value)
			}
		}
	}
	encoded := copyValues.Encode()
	if encoded == "" {
		return ""
	}
	return "&" + encoded
}

func mediaAssetFilterFromRequest(r *http.Request) repository.MediaAssetFilter {
	values := r.URL.Query()
	filter := repository.MediaAssetFilter{
		Query: strings.TrimSpace(values.Get("q")),
	}
	month := strings.TrimSpace(values.Get("month"))
	if month != "" {
		if start, err := time.Parse("2006-01", month); err == nil {
			end := start.AddDate(0, 1, 0)
			filter.DateFrom = &start
			filter.DateTo = &end
			return filter
		}
	}
	if start := parseMediaDate(values.Get("date_from")); start != nil {
		filter.DateFrom = start
	}
	if end := parseMediaDate(values.Get("date_to")); end != nil {
		exclusive := end.AddDate(0, 0, 1)
		filter.DateTo = &exclusive
	}
	return filter
}

func parseMediaDate(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &t
}

func mediaSuccessMessage(code string) string {
	switch code {
	case "upload":
		return "Midia enviada com sucesso"
	case "update":
		return "Midia atualizada com sucesso"
	case "delete":
		return "Midia removida com sucesso"
	default:
		return ""
	}
}

func (h *Handler) mediaAssetsForForms(ctx context.Context) []model.MediaAsset {
	assets, err := h.repo.MediaAssetList(ctx, "", 80, 0)
	if err != nil {
		return nil
	}
	return assets
}

func (h *Handler) mediaKeyFromRequest(r *http.Request, field string) string {
	key := strings.TrimSpace(r.FormValue(field))
	if key == "" {
		return ""
	}
	asset, err := h.repo.MediaAssetGetByKey(r.Context(), key)
	if err != nil || asset == nil {
		return ""
	}
	return asset.Key
}

func (h *Handler) mediaKeysFromRequest(r *http.Request, field string, current []string) []string {
	seen := make(map[string]bool, len(current)+len(r.Form[field]))
	var keys []string
	for _, key := range current {
		if strings.TrimSpace(key) == "" || seen[key] {
			continue
		}
		seen[key] = true
		keys = append(keys, key)
	}
	for _, raw := range r.Form[field] {
		key := strings.TrimSpace(raw)
		if key == "" || seen[key] {
			continue
		}
		asset, err := h.repo.MediaAssetGetByKey(r.Context(), key)
		if err != nil || asset == nil {
			continue
		}
		seen[asset.Key] = true
		keys = append(keys, asset.Key)
	}
	return keys
}

func (h *Handler) recordMediaAssetFromUpload(r *http.Request, user *model.User, header *multipart.FileHeader, info *storage.FileInfo, title string) {
	if header == nil || info == nil || info.Key == "" {
		return
	}
	existing, err := h.repo.MediaAssetGetByKey(r.Context(), info.Key)
	if err == nil && existing != nil {
		return
	}
	originalName := filepath.Base(header.Filename)
	title = strings.TrimSpace(title)
	altText := title
	if altText == "" {
		altText = originalName
	}
	asset := &model.MediaAsset{
		Key:          info.Key,
		OriginalName: originalName,
		Title:        title,
		AltText:      altText,
		ContentType:  info.ContentType,
		SizeBytes:    info.Size,
	}
	if user != nil {
		id := user.ID
		asset.UploadedBy = &id
	}
	if err := h.repo.MediaAssetCreate(r.Context(), asset); err == nil {
		h.auditAdminAction(r, user, "create", "media", auditEntityID(asset.ID), map[string]any{
			"key":           asset.Key,
			"original_name": asset.OriginalName,
			"source":        "inline_upload",
		})
	}
}
