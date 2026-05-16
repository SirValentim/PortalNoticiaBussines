package handler

import (
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/storage"
)

func (h *Handler) AdminSettings(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermSettingsManage); !ok {
		return
	}
	h.renderSettings(w, r, h.portalSettings(r.Context()), settingsSuccessMessage(r.URL.Query().Get("success")), "")
}

func (h *Handler) AdminSettingsUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermSettingsManage)
	if !ok {
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.renderSettings(w, r, h.portalSettings(r.Context()), "", "Tamanho maximo excedido")
		return
	}

	current := h.portalSettings(r.Context())
	settings := model.PortalSettings{
		SiteName:                  strings.TrimSpace(r.FormValue("site_name")),
		Tagline:                   strings.TrimSpace(r.FormValue("tagline")),
		LogoKey:                   current.LogoKey,
		FaviconKey:                current.FaviconKey,
		ContactEmail:              strings.TrimSpace(r.FormValue("contact_email")),
		ContactWhatsapp:           strings.TrimSpace(r.FormValue("contact_whatsapp")),
		ContactPhone:              strings.TrimSpace(r.FormValue("contact_phone")),
		City:                      strings.TrimSpace(r.FormValue("city")),
		State:                     strings.TrimSpace(r.FormValue("state")),
		SEOTitle:                  strings.TrimSpace(r.FormValue("seo_title")),
		SEODescription:            strings.TrimSpace(r.FormValue("seo_description")),
		FacebookURL:               strings.TrimSpace(r.FormValue("facebook_url")),
		InstagramURL:              strings.TrimSpace(r.FormValue("instagram_url")),
		YoutubeURL:                strings.TrimSpace(r.FormValue("youtube_url")),
		TiktokURL:                 strings.TrimSpace(r.FormValue("tiktok_url")),
		UploadMaxMB:               intFromForm(r, "upload_max_mb", current.UploadMaxMB),
		AutomationEnabled:         r.FormValue("automation_enabled") == "on",
		AutomationIntervalMinutes: intFromForm(r, "automation_interval_minutes", current.AutomationIntervalMinutes),
	}
	if key := h.mediaKeyFromRequest(r, "logo_media_key"); key != "" {
		settings.LogoKey = key
	}
	if key := h.mediaKeyFromRequest(r, "favicon_media_key"); key != "" {
		settings.FaviconKey = key
	}
	if r.FormValue("remove_logo") == "on" {
		settings.LogoKey = ""
	}
	if r.FormValue("remove_favicon") == "on" {
		settings.FaviconKey = ""
	}
	if key, err := h.settingsImageUpload(r, user, "logo", "Logo do portal"); err != nil {
		h.renderSettings(w, r, settings, "", err.Error())
		return
	} else if key != "" {
		settings.LogoKey = key
	}
	if key, err := h.settingsImageUpload(r, user, "favicon", "Favicon do portal"); err != nil {
		h.renderSettings(w, r, settings, "", err.Error())
		return
	} else if key != "" {
		settings.FaviconKey = key
	}

	if settings.SiteName == "" {
		h.renderSettings(w, r, settings, "", "Informe o nome do portal")
		return
	}
	if settings.SEOTitle == "" || settings.SEODescription == "" {
		h.renderSettings(w, r, settings, "", "Informe titulo e descricao SEO globais")
		return
	}
	if settings.UploadMaxMB <= 0 || settings.UploadMaxMB > 50 {
		h.renderSettings(w, r, settings, "", "O limite de upload deve ficar entre 1 MB e 50 MB")
		return
	}
	if settings.AutomationIntervalMinutes < 5 {
		h.renderSettings(w, r, settings, "", "O intervalo de automacao deve ser de pelo menos 5 minutos")
		return
	}
	if err := h.repo.PortalSettingsUpdate(r.Context(), &settings); err != nil {
		h.renderSettings(w, r, settings, "", "Nao foi possivel salvar as configuracoes")
		return
	}
	h.auditAdminAction(r, user, "update", "settings", auditEntityID(1), map[string]any{
		"site_name":                   settings.SiteName,
		"logo_key":                    settings.LogoKey,
		"favicon_key":                 settings.FaviconKey,
		"contact_email":               settings.ContactEmail,
		"upload_max_mb":               settings.UploadMaxMB,
		"automation_enabled":          settings.AutomationEnabled,
		"automation_interval_minutes": settings.AutomationIntervalMinutes,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/settings?success=update", http.StatusSeeOther)
}

func (h *Handler) renderSettings(w http.ResponseWriter, r *http.Request, settings model.PortalSettings, success, errorMessage string) {
	h.Render(w, r, "admin_settings.html", map[string]any{
		"Title":       "Configuracoes",
		"Active":      "settings",
		"Settings":    settings,
		"MediaAssets": h.mediaAssetsForForms(r.Context()),
		"Success":     success,
		"Error":       errorMessage,
	})
}

func (h *Handler) settingsImageUpload(r *http.Request, user *model.User, field, title string) (string, error) {
	if r.MultipartForm == nil {
		return "", nil
	}
	file, header, err := r.FormFile(field)
	if err != nil {
		if errors.Is(err, http.ErrMissingFile) {
			return "", nil
		}
		return "", err
	}
	defer file.Close()
	return h.storeSettingsImage(r, user, file, header, title)
}

func (h *Handler) storeSettingsImage(r *http.Request, user *model.User, file multipart.File, header *multipart.FileHeader, title string) (string, error) {
	info, err := h.storage.Upload(r.Context(), storage.GenerateKey(header.Filename), file, header.Header.Get("Content-Type"))
	if err != nil {
		return "", err
	}
	h.recordMediaAssetFromUpload(r, user, header, info, title)
	return info.Key, nil
}

func intFromForm(r *http.Request, field string, fallback int) int {
	value := strings.TrimSpace(r.FormValue(field))
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

func settingsSuccessMessage(code string) string {
	if code == "update" {
		return "Configuracoes atualizadas com sucesso"
	}
	return ""
}

func socialLinks(settings model.PortalSettings) []string {
	links := []string{}
	for _, link := range []string{settings.FacebookURL, settings.InstagramURL, settings.YoutubeURL, settings.TiktokURL} {
		if strings.TrimSpace(link) != "" {
			links = append(links, strings.TrimSpace(link))
		}
	}
	return links
}
