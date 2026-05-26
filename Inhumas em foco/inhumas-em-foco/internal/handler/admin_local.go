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

func (h *Handler) AdminStores(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermStoresManage); !ok {
		return
	}
	stores, _ := h.repo.StoreList(r.Context(), false, 100)
	h.Render(w, r, "admin_stores.html", map[string]any{
		"Stores":  stores,
		"Summary": localCommercialSummaryStores(stores),
	})
}

func (h *Handler) AdminStoreNew(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermStoresManage); !ok {
		return
	}
	neighborhoods, _ := h.repo.NeighborhoodList(r.Context())
	h.Render(w, r, "admin_store_form.html", map[string]any{
		"Neighborhoods": neighborhoods,
		"Store":         nil,
		"MediaAssets":   h.mediaAssetsForForms(r.Context()),
	})
}

func (h *Handler) AdminStoreCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermStoresManage)
	if !ok {
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.Render(w, r, "admin_store_form.html", map[string]any{"Error": "Tamanho maximo excedido"})
		return
	}

	store := &model.Store{
		Name:             r.FormValue("name"),
		Description:      r.FormValue("description"),
		Category:         r.FormValue("category"),
		Address:          r.FormValue("address"),
		Phone:            r.FormValue("phone"),
		Whatsapp:         r.FormValue("whatsapp"),
		WebsiteURL:       r.FormValue("website_url"),
		CommercialStatus: normalizeStoreCommercialStatusForHandler(r.FormValue("commercial_status")),
		MetaTitle:        r.FormValue("meta_title"),
		MetaDescription:  r.FormValue("meta_description"),
		IsSponsored:      r.FormValue("is_sponsored") == "on",
		IsFeatured:       r.FormValue("is_featured") == "on",
		Active:           true,
	}

	if nID, err := strconv.ParseInt(r.FormValue("neighborhood_id"), 10, 64); err == nil {
		store.NeighborhoodID = &nID
	}
	if key := h.mediaKeyFromRequest(r, "logo_media_key"); key != "" {
		store.LogoKey = key
	}
	if key := h.mediaKeyFromRequest(r, "cover_media_key"); key != "" {
		store.CoverImageKey = key
	}

	file, header, err := r.FormFile("logo")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			store.LogoKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, store.Name)
		}
	}

	file, header, err = r.FormFile("cover")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			store.CoverImageKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, store.Name)
		}
	}

	store.Slug = slug.Unique(store.Name, func(s string) bool {
		return h.repo.StoreSlugExists(r.Context(), s)
	})

	if err := h.repo.StoreCreate(r.Context(), store); err != nil {
		neighborhoods, _ := h.repo.NeighborhoodList(r.Context())
		h.Render(w, r, "admin_store_form.html", map[string]any{
			"Error":         err.Error(),
			"Neighborhoods": neighborhoods,
			"Store":         store,
			"MediaAssets":   h.mediaAssetsForForms(r.Context()),
		})
		return
	}
	h.auditAdminAction(r, user, "create", "store", auditEntityID(store.ID), map[string]any{
		"name":              store.Name,
		"slug":              store.Slug,
		"commercial_status": store.CommercialStatus,
		"is_featured":       store.IsFeatured,
		"is_sponsored":      store.IsSponsored,
	})

	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/stores", http.StatusSeeOther)
}

func (h *Handler) AdminStoreEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermStoresManage); !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	store, err := h.repo.StoreGetByID(r.Context(), id)
	if err != nil || store == nil {
		http.NotFound(w, r)
		return
	}
	neighborhoods, _ := h.repo.NeighborhoodList(r.Context())
	h.Render(w, r, "admin_store_form.html", map[string]any{
		"Neighborhoods": neighborhoods,
		"Store":         store,
		"MediaAssets":   h.mediaAssetsForForms(r.Context()),
	})
}

func (h *Handler) AdminStoreUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermStoresManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	store, err := h.repo.StoreGetByID(r.Context(), id)
	if err != nil || store == nil {
		http.NotFound(w, r)
		return
	}

	if err := h.parseMultipart(w, r); err != nil {
		h.Render(w, r, "admin_store_form.html", map[string]any{"Error": "Tamanho maximo excedido", "Store": store, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}

	store.Name = r.FormValue("name")
	store.Description = r.FormValue("description")
	store.Category = r.FormValue("category")
	store.Address = r.FormValue("address")
	store.Phone = r.FormValue("phone")
	store.Whatsapp = r.FormValue("whatsapp")
	store.WebsiteURL = r.FormValue("website_url")
	store.CommercialStatus = normalizeStoreCommercialStatusForHandler(r.FormValue("commercial_status"))
	store.MetaTitle = r.FormValue("meta_title")
	store.MetaDescription = r.FormValue("meta_description")
	store.IsSponsored = r.FormValue("is_sponsored") == "on"
	store.IsFeatured = r.FormValue("is_featured") == "on"
	store.Active = r.FormValue("active") == "on"

	if nID, err := strconv.ParseInt(r.FormValue("neighborhood_id"), 10, 64); err == nil {
		store.NeighborhoodID = &nID
	}

	if r.FormValue("remove_logo") == "on" && store.LogoKey != "" {
		h.storage.Delete(r.Context(), store.LogoKey)
		store.LogoKey = ""
	}
	if r.FormValue("remove_cover") == "on" && store.CoverImageKey != "" {
		h.storage.Delete(r.Context(), store.CoverImageKey)
		store.CoverImageKey = ""
	}
	if key := h.mediaKeyFromRequest(r, "logo_media_key"); key != "" {
		store.LogoKey = key
	}
	if key := h.mediaKeyFromRequest(r, "cover_media_key"); key != "" {
		store.CoverImageKey = key
	}

	file, header, err := r.FormFile("logo")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			store.LogoKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, store.Name)
		}
	}

	file, header, err = r.FormFile("cover")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			store.CoverImageKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, store.Name)
		}
	}

	if err := h.repo.StoreUpdate(r.Context(), store); err != nil {
		neighborhoods, _ := h.repo.NeighborhoodList(r.Context())
		h.Render(w, r, "admin_store_form.html", map[string]any{
			"Error":         err.Error(),
			"Neighborhoods": neighborhoods,
			"Store":         store,
			"MediaAssets":   h.mediaAssetsForForms(r.Context()),
		})
		return
	}
	h.auditAdminAction(r, user, "update", "store", auditEntityID(store.ID), map[string]any{
		"name":              store.Name,
		"active":            store.Active,
		"commercial_status": store.CommercialStatus,
		"is_featured":       store.IsFeatured,
		"is_sponsored":      store.IsSponsored,
	})

	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/stores", http.StatusSeeOther)
}

func (h *Handler) AdminStoreDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermStoresManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	store, _ := h.repo.StoreGetByID(r.Context(), id)
	h.repo.StoreDelete(r.Context(), id)
	h.auditAdminAction(r, user, "delete", "store", auditEntityID(id), map[string]any{
		"name": storeName(store),
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/stores", http.StatusSeeOther)
}

func normalizeStoreCommercialStatusForHandler(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "lead", "paused", "inactive":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "active"
	}
}

func (h *Handler) AdminInfluencers(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermInfluencersManage); !ok {
		return
	}
	influencers, _ := h.repo.InfluencerList(r.Context(), false, 200)
	h.Render(w, r, "admin_influencers.html", map[string]any{
		"Title":       "Influencers",
		"Active":      "influencers",
		"Influencers": influencers,
		"Summary":     influencerAdminSummary(influencers),
	})
}

func (h *Handler) AdminInfluencerNew(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermInfluencersManage); !ok {
		return
	}
	h.Render(w, r, "admin_influencer_form.html", map[string]any{
		"Title":       "Novo Influencer",
		"Active":      "influencers",
		"Influencer":  nil,
		"MediaAssets": h.mediaAssetsForForms(r.Context()),
	})
}

func (h *Handler) AdminInfluencerCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermInfluencersManage)
	if !ok {
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.Render(w, r, "admin_influencer_form.html", map[string]any{"Error": "Tamanho maximo excedido"})
		return
	}
	influencer := h.influencerFromRequest(r, nil)
	influencer.Slug = slug.Unique(influencer.Name, func(s string) bool {
		return h.repo.InfluencerSlugExists(r.Context(), s)
	})
	h.applyInfluencerMediaSelection(r, influencer)
	h.applyInfluencerUploads(r, user, influencer)
	if strings.TrimSpace(influencer.Name) == "" {
		h.Render(w, r, "admin_influencer_form.html", map[string]any{"Error": "Nome e obrigatorio", "Influencer": influencer, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	if err := h.repo.InfluencerCreate(r.Context(), influencer); err != nil {
		h.Render(w, r, "admin_influencer_form.html", map[string]any{"Error": err.Error(), "Influencer": influencer, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	h.auditAdminAction(r, user, "create", "influencer", auditEntityID(influencer.ID), map[string]any{
		"name":        influencer.Name,
		"slug":        influencer.Slug,
		"niche":       influencer.Niche,
		"is_featured": influencer.IsFeatured,
		"active":      influencer.Active,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/influencers", http.StatusSeeOther)
}

func (h *Handler) AdminInfluencerEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermInfluencersManage); !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	influencer, err := h.repo.InfluencerGetByID(r.Context(), id)
	if err != nil || influencer == nil {
		http.NotFound(w, r)
		return
	}
	h.Render(w, r, "admin_influencer_form.html", map[string]any{
		"Title":       "Editar Influencer",
		"Active":      "influencers",
		"Influencer":  influencer,
		"MediaAssets": h.mediaAssetsForForms(r.Context()),
	})
}

func (h *Handler) AdminInfluencerUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermInfluencersManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	influencer, err := h.repo.InfluencerGetByID(r.Context(), id)
	if err != nil || influencer == nil {
		http.NotFound(w, r)
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.Render(w, r, "admin_influencer_form.html", map[string]any{"Error": "Tamanho maximo excedido", "Influencer": influencer, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	oldSlug := influencer.Slug
	influencer = h.influencerFromRequest(r, influencer)
	if strings.TrimSpace(influencer.Name) == "" {
		h.Render(w, r, "admin_influencer_form.html", map[string]any{"Error": "Nome e obrigatorio", "Influencer": influencer, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	newSlug := slug.Generate(influencer.Name)
	if newSlug != oldSlug {
		influencer.Slug = slug.Unique(influencer.Name, func(s string) bool {
			return h.repo.InfluencerSlugExists(r.Context(), s)
		})
	}
	if r.FormValue("remove_avatar") == "on" && influencer.AvatarKey != "" {
		h.storage.Delete(r.Context(), influencer.AvatarKey)
		influencer.AvatarKey = ""
	}
	if r.FormValue("remove_cover") == "on" && influencer.CoverImageKey != "" {
		h.storage.Delete(r.Context(), influencer.CoverImageKey)
		influencer.CoverImageKey = ""
	}
	h.applyInfluencerMediaSelection(r, influencer)
	h.applyInfluencerUploads(r, user, influencer)
	if err := h.repo.InfluencerUpdate(r.Context(), influencer); err != nil {
		h.Render(w, r, "admin_influencer_form.html", map[string]any{"Error": err.Error(), "Influencer": influencer, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	h.auditAdminAction(r, user, "update", "influencer", auditEntityID(influencer.ID), map[string]any{
		"name":        influencer.Name,
		"slug":        influencer.Slug,
		"niche":       influencer.Niche,
		"is_featured": influencer.IsFeatured,
		"active":      influencer.Active,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/influencers", http.StatusSeeOther)
}

func (h *Handler) AdminInfluencerDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermInfluencersManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	influencer, _ := h.repo.InfluencerGetByID(r.Context(), id)
	h.repo.InfluencerDelete(r.Context(), id)
	h.auditAdminAction(r, user, "delete", "influencer", auditEntityID(id), map[string]any{
		"name": influencerName(influencer),
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/influencers", http.StatusSeeOther)
}

func (h *Handler) influencerFromRequest(r *http.Request, influencer *model.Influencer) *model.Influencer {
	if influencer == nil {
		influencer = &model.Influencer{}
	}
	influencer.Name = r.FormValue("name")
	influencer.Bio = r.FormValue("bio")
	influencer.CityArea = r.FormValue("city_area")
	influencer.Niche = r.FormValue("niche")
	influencer.Instagram = r.FormValue("instagram")
	influencer.TikTok = r.FormValue("tiktok")
	influencer.YouTube = r.FormValue("youtube")
	influencer.Whatsapp = r.FormValue("whatsapp")
	influencer.MetaTitle = r.FormValue("meta_title")
	influencer.MetaDescription = r.FormValue("meta_description")
	influencer.IsFeatured = r.FormValue("is_featured") == "on"
	influencer.IsSponsored = r.FormValue("is_sponsored") == "on"
	influencer.Active = r.FormValue("active") == "on"
	return influencer
}

type influencerSummary struct {
	Total      int
	Active     int
	Featured   int
	Sponsored  int
	TotalViews int
}

func influencerAdminSummary(influencers []model.Influencer) influencerSummary {
	var summary influencerSummary
	summary.Total = len(influencers)
	for _, influencer := range influencers {
		if influencer.Active {
			summary.Active++
		}
		if influencer.IsFeatured {
			summary.Featured++
		}
		if influencer.IsSponsored {
			summary.Sponsored++
		}
		summary.TotalViews += influencer.ViewCount
	}
	return summary
}

func (h *Handler) applyInfluencerMediaSelection(r *http.Request, influencer *model.Influencer) {
	if key := h.mediaKeyFromRequest(r, "avatar_media_key"); key != "" {
		influencer.AvatarKey = key
	}
	if key := h.mediaKeyFromRequest(r, "cover_media_key"); key != "" {
		influencer.CoverImageKey = key
	}
}

func (h *Handler) applyInfluencerUploads(r *http.Request, user *model.User, influencer *model.Influencer) {
	if file, header, err := r.FormFile("avatar"); err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			influencer.AvatarKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, influencer.Name)
		}
	}
	if file, header, err := r.FormFile("cover"); err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			influencer.CoverImageKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, influencer.Name)
		}
	}
}

func (h *Handler) AdminNeighborhoods(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermSettingsManage); !ok {
		return
	}
	if !h.tenantFeatureEnabled(r, "commercial", true) {
		http.Error(w, "Feature desabilitada para este portal", http.StatusForbidden)
		return
	}
	neighborhoods, _ := h.repo.NeighborhoodList(r.Context())
	h.Render(w, r, "admin_neighborhoods.html", map[string]any{"Neighborhoods": neighborhoods})
}

func (h *Handler) AdminNeighborhoodCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermSettingsManage)
	if !ok {
		return
	}
	if !h.tenantFeatureEnabled(r, "commercial", true) {
		http.Error(w, "Feature desabilitada para este portal", http.StatusForbidden)
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.Render(w, r, "admin_neighborhoods.html", map[string]any{"Error": "Tamanho maximo excedido"})
		return
	}

	n := &model.Neighborhood{
		Name:            r.FormValue("name"),
		Description:     r.FormValue("description"),
		MetaTitle:       r.FormValue("meta_title"),
		MetaDescription: r.FormValue("meta_description"),
	}

	file, header, err := r.FormFile("cover")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			n.CoverImageKey = info.Key
		}
	}

	n.Slug = slug.Unique(n.Name, func(s string) bool {
		return h.repo.NeighborhoodSlugExists(r.Context(), s)
	})

	h.repo.NeighborhoodCreate(r.Context(), n)
	h.auditAdminAction(r, user, "create", "neighborhood", auditEntityID(n.ID), map[string]any{
		"name": n.Name,
		"slug": n.Slug,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/neighborhoods", http.StatusSeeOther)
}

func (h *Handler) AdminNeighborhoodDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermSettingsManage)
	if !ok {
		return
	}
	if !h.tenantFeatureEnabled(r, "commercial", true) {
		http.Error(w, "Feature desabilitada para este portal", http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	h.repo.NeighborhoodDelete(r.Context(), id)
	h.auditAdminAction(r, user, "delete", "neighborhood", auditEntityID(id), nil)
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/neighborhoods", http.StatusSeeOther)
}
