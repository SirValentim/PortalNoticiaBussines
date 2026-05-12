package handler

import (
	"context"
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

func (h *Handler) AdminBanners(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermBannersManage); !ok {
		return
	}
	banners, _ := h.repo.BannerList(r.Context())
	h.Render(w, r, "admin_banners.html", map[string]any{
		"Banners":     banners,
		"BannerSlots": bannerSlots(banners),
		"Summary":     bannerSummary(banners),
		"Active":      "banners",
	})
}

func (h *Handler) AdminBannerNew(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermBannersManage); !ok {
		return
	}
	h.Render(w, r, "admin_banner_form.html", map[string]any{"MediaAssets": h.mediaAssetsForForms(r.Context())})
}

func (h *Handler) AdminBannerEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermBannersManage); !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	banner, err := h.repo.BannerGetByID(r.Context(), id)
	if err != nil || banner == nil {
		http.NotFound(w, r)
		return
	}
	h.Render(w, r, "admin_banner_form.html", map[string]any{"Banner": banner, "MediaAssets": h.mediaAssetsForForms(r.Context())})
}

func (h *Handler) AdminBannerCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermBannersManage)
	if !ok {
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.Render(w, r, "admin_banner_form.html", map[string]any{"Error": "Tamanho maximo excedido", "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}

	banner := &model.Banner{
		Name:            r.FormValue("name"),
		AdvertiserName:  r.FormValue("advertiser_name"),
		ContactName:     r.FormValue("contact_name"),
		ContactPhone:    r.FormValue("contact_phone"),
		ContactWhatsapp: r.FormValue("contact_whatsapp"),
		PriceDisplay:    r.FormValue("price_display"),
		Notes:           r.FormValue("notes"),
		Position:        r.FormValue("position"),
		LinkURL:         r.FormValue("link_url"),
		Status:          normalizeBannerStatus(r.FormValue("status")),
	}
	if strings.TrimSpace(banner.AdvertiserName) == "" {
		banner.AdvertiserName = banner.Name
	}
	banner.Active = banner.Status == "active"

	priority, _ := strconv.Atoi(r.FormValue("priority"))
	banner.Priority = priority

	startDate, endDate, dateErr := h.comSvc.ParseDateRange(r.FormValue("start_date"), r.FormValue("end_date"))
	if dateErr != "" {
		h.Render(w, r, "admin_banner_form.html", map[string]any{"Error": dateErr, "Banner": banner})
		return
	}
	banner.StartDate = startDate
	banner.EndDate = endDate

	if banner.Active && h.bannerHasOverlap(r.Context(), banner.Position, startDate, endDate, 0) {
		h.Render(w, r, "admin_banner_form.html", map[string]any{"Error": "Ja existe banner ativo nessa posicao no periodo", "Banner": banner, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	if key := h.mediaKeyFromRequest(r, "image_media_key"); key != "" {
		banner.ImageKey = key
	}

	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			banner.ImageKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, banner.Name)
		}
	}
	if msg := h.comSvc.ValidateBanner(banner); msg != "" {
		h.Render(w, r, "admin_banner_form.html", map[string]any{"Error": msg, "Banner": banner, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}

	if err := h.repo.BannerCreate(r.Context(), banner); err != nil {
		h.Render(w, r, "admin_banner_form.html", map[string]any{"Error": err.Error(), "Banner": banner, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	h.auditAdminAction(r, user, "create", "banner", auditEntityID(banner.ID), map[string]any{
		"name":       banner.Name,
		"advertiser": banner.AdvertiserName,
		"position":   banner.Position,
		"status":     banner.Status,
		"start_date": banner.StartDate.Format("2006-01-02"),
		"end_date":   banner.EndDate.Format("2006-01-02"),
	})

	h.repo.JobCreate(r.Context(), &model.Job{
		Type:    model.JobExpireBanner,
		Payload: fmt.Sprintf(`{"banner_id":%d}`, banner.ID),
		Status:  model.JobPending,
		RunAt:   endDate.Add(24 * time.Hour),
	})

	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/banners", http.StatusSeeOther)
}

func (h *Handler) AdminBannerUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermBannersManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	banner, err := h.repo.BannerGetByID(r.Context(), id)
	if err != nil || banner == nil {
		http.NotFound(w, r)
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		h.Render(w, r, "admin_banner_form.html", map[string]any{"Error": "Tamanho maximo excedido", "Banner": banner, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}

	banner.Name = r.FormValue("name")
	banner.AdvertiserName = r.FormValue("advertiser_name")
	if strings.TrimSpace(banner.AdvertiserName) == "" {
		banner.AdvertiserName = banner.Name
	}
	banner.ContactName = r.FormValue("contact_name")
	banner.ContactPhone = r.FormValue("contact_phone")
	banner.ContactWhatsapp = r.FormValue("contact_whatsapp")
	banner.PriceDisplay = r.FormValue("price_display")
	banner.Notes = r.FormValue("notes")
	banner.Position = r.FormValue("position")
	banner.LinkURL = r.FormValue("link_url")
	banner.Status = normalizeBannerStatus(r.FormValue("status"))
	banner.Active = banner.Status == "active"
	banner.Priority, _ = strconv.Atoi(r.FormValue("priority"))
	startDate, endDate, dateErr := h.comSvc.ParseDateRange(r.FormValue("start_date"), r.FormValue("end_date"))
	if dateErr != "" {
		h.Render(w, r, "admin_banner_form.html", map[string]any{"Error": dateErr, "Banner": banner, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	banner.StartDate = startDate
	banner.EndDate = endDate

	if banner.Active && h.bannerHasOverlap(r.Context(), banner.Position, startDate, endDate, banner.ID) {
		h.Render(w, r, "admin_banner_form.html", map[string]any{"Error": "Ja existe banner ativo nessa posicao no periodo", "Banner": banner, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}

	if r.FormValue("remove_image") == "on" && banner.ImageKey != "" {
		h.storage.Delete(r.Context(), banner.ImageKey)
		banner.ImageKey = ""
	}
	if key := h.mediaKeyFromRequest(r, "image_media_key"); key != "" {
		banner.ImageKey = key
	}

	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			banner.ImageKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, banner.Name)
		}
	}
	if msg := h.comSvc.ValidateBanner(banner); msg != "" {
		h.Render(w, r, "admin_banner_form.html", map[string]any{"Error": msg, "Banner": banner, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}

	if err := h.repo.BannerUpdate(r.Context(), banner); err != nil {
		h.Render(w, r, "admin_banner_form.html", map[string]any{"Error": "Erro ao atualizar banner", "Banner": banner, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	h.auditAdminAction(r, user, "update", "banner", auditEntityID(banner.ID), map[string]any{
		"name":       banner.Name,
		"advertiser": banner.AdvertiserName,
		"position":   banner.Position,
		"status":     banner.Status,
		"active":     banner.Active,
		"start_date": banner.StartDate.Format("2006-01-02"),
		"end_date":   banner.EndDate.Format("2006-01-02"),
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/banners", http.StatusSeeOther)
}

func (h *Handler) bannerHasOverlap(ctx context.Context, position string, startDate, endDate time.Time, excludeID int64) bool {
	count, err := h.repo.BannerCountActiveInPeriodExcluding(ctx, position, startDate, endDate, excludeID)
	return err == nil && count > 0
}

func (h *Handler) AdminBannerDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermBannersManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	banner, err := h.repo.BannerGetByID(r.Context(), id)
	if err == nil && banner != nil && banner.ImageKey != "" {
		h.storage.Delete(r.Context(), banner.ImageKey)
	}
	h.repo.BannerDelete(r.Context(), id)
	h.auditAdminAction(r, user, "delete", "banner", auditEntityID(id), map[string]any{
		"name":     bannerName(banner),
		"position": bannerPosition(banner),
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/banners", http.StatusSeeOther)
}

type bannerCommercialSummary struct {
	Total        int
	ActiveNow    int
	Paused       int
	Draft        int
	Expired      int
	Expiring     int
	WithImage    int
	WithoutImage int
}

type bannerSlot struct {
	Position    string
	Label       string
	Spec        string
	Description string
	ActiveCount int
	TotalCount  int
}

func bannerSummary(banners []model.Banner) bannerCommercialSummary {
	now := time.Now()
	today := dateOnly(now)
	summary := bannerCommercialSummary{Total: len(banners)}
	for _, banner := range banners {
		status := normalizeBannerStatus(banner.Status)
		switch status {
		case "draft":
			summary.Draft++
		case "paused":
			summary.Paused++
		case "expired":
			summary.Expired++
		}
		if banner.ImageKey == "" {
			summary.WithoutImage++
		} else {
			summary.WithImage++
		}
		if banner.Active && status == "active" && !dateOnly(banner.StartDate).After(today) && !dateOnly(banner.EndDate).Before(today) {
			summary.ActiveNow++
			if days := int(dateOnly(banner.EndDate).Sub(today).Hours() / 24); days >= 0 && days <= 7 {
				summary.Expiring++
			}
		}
		if status == "active" && dateOnly(banner.EndDate).Before(today) {
			summary.Expired++
		}
	}
	return summary
}

func bannerSlots(banners []model.Banner) []bannerSlot {
	slots := []bannerSlot{
		{Position: "hero", Label: "Topo da home", Spec: "1200 x 220", Description: "Faixa principal logo abaixo do destaque editorial."},
		{Position: "in_feed", Label: "In-feed", Spec: "1200 x 180", Description: "Entre blocos de noticias, lojas e promocoes."},
		{Position: "sidebar_top", Label: "Sidebar topo", Spec: "320 x 250", Description: "Coluna lateral em materias e paginas de leitura."},
		{Position: "sidebar_bottom", Label: "Sidebar rodape", Spec: "320 x 250", Description: "Segundo ponto comercial lateral."},
		{Position: "sticky_footer", Label: "Rodape mobile", Spec: "720 x 120", Description: "Banner fixo inferior no celular."},
	}
	today := dateOnly(time.Now())
	for i := range slots {
		for _, banner := range banners {
			if banner.Position != slots[i].Position {
				continue
			}
			slots[i].TotalCount++
			if banner.Active && normalizeBannerStatus(banner.Status) == "active" && !dateOnly(banner.StartDate).After(today) && !dateOnly(banner.EndDate).Before(today) {
				slots[i].ActiveCount++
			}
		}
	}
	return slots
}

func (h *Handler) AdminPromotions(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermPromosManage); !ok {
		return
	}
	promos, _ := h.repo.PromotionListActive(r.Context(), 100)
	stores, _ := h.repo.StoreList(r.Context(), false, 1000)
	storeMap := make(map[int64]string)
	for _, s := range stores {
		storeMap[s.ID] = s.Name
	}
	h.Render(w, r, "admin_promotions.html", map[string]any{
		"Promos":   promos,
		"StoreMap": storeMap,
	})
}

func (h *Handler) AdminPromoNew(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermPromosManage); !ok {
		return
	}
	stores, _ := h.repo.StoreList(r.Context(), false, 1000)
	h.Render(w, r, "admin_promo_form.html", map[string]any{"Stores": stores, "MediaAssets": h.mediaAssetsForForms(r.Context())})
}

func (h *Handler) AdminPromoEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermPromosManage); !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	promo, err := h.repo.PromotionGetByID(r.Context(), id)
	if err != nil || promo == nil {
		http.NotFound(w, r)
		return
	}
	stores, _ := h.repo.StoreList(r.Context(), false, 1000)
	h.Render(w, r, "admin_promo_form.html", map[string]any{"Stores": stores, "Promo": promo, "MediaAssets": h.mediaAssetsForForms(r.Context())})
}

func (h *Handler) AdminPromoCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermPromosManage)
	if !ok {
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		stores, _ := h.repo.StoreList(r.Context(), false, 1000)
		h.Render(w, r, "admin_promo_form.html", map[string]any{"Error": "Tamanho maximo excedido", "Stores": stores, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}

	promo := &model.Promotion{
		Title:        r.FormValue("title"),
		Description:  r.FormValue("description"),
		PriceDisplay: r.FormValue("price_display"),
		Status:       "active",
		IsSponsored:  r.FormValue("is_sponsored") == "on",
	}

	if storeID, err := strconv.ParseInt(r.FormValue("store_id"), 10, 64); err == nil {
		promo.StoreID = storeID
	}

	startDate, endDate, dateErr := h.comSvc.ParseDateRange(r.FormValue("start_date"), r.FormValue("end_date"))
	if dateErr != "" {
		stores, _ := h.repo.StoreList(r.Context(), false, 1000)
		h.Render(w, r, "admin_promo_form.html", map[string]any{"Error": dateErr, "Stores": stores, "Promo": promo, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	promo.StartDate = startDate
	promo.EndDate = endDate
	if key := h.mediaKeyFromRequest(r, "image_media_key"); key != "" {
		promo.ImageKey = key
	}

	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			promo.ImageKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, promo.Title)
		}
	}

	promo.Slug = slug.Unique(promo.Title, func(s string) bool {
		return h.repo.PromotionSlugExists(r.Context(), s)
	})

	if err := h.repo.PromotionCreate(r.Context(), promo); err != nil {
		stores, _ := h.repo.StoreList(r.Context(), false, 1000)
		h.Render(w, r, "admin_promo_form.html", map[string]any{"Error": err.Error(), "Stores": stores, "Promo": promo, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	h.auditAdminAction(r, user, "create", "promotion", auditEntityID(promo.ID), map[string]any{
		"title":        promo.Title,
		"slug":         promo.Slug,
		"store_id":     promo.StoreID,
		"status":       promo.Status,
		"is_sponsored": promo.IsSponsored,
		"start_date":   promo.StartDate.Format("2006-01-02"),
		"end_date":     promo.EndDate.Format("2006-01-02"),
	})

	h.repo.JobCreate(r.Context(), &model.Job{
		Type:    model.JobExpirePromotion,
		Payload: fmt.Sprintf(`{"promotion_id":%d}`, promo.ID),
		Status:  model.JobPending,
		RunAt:   endDate.Add(24 * time.Hour),
	})

	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/promotions", http.StatusSeeOther)
}

func (h *Handler) AdminPromoUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermPromosManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	promo, err := h.repo.PromotionGetByID(r.Context(), id)
	if err != nil || promo == nil {
		http.NotFound(w, r)
		return
	}
	if err := h.parseMultipart(w, r); err != nil {
		stores, _ := h.repo.StoreList(r.Context(), false, 1000)
		h.Render(w, r, "admin_promo_form.html", map[string]any{"Error": "Tamanho maximo excedido", "Stores": stores, "Promo": promo, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}

	promo.Title = r.FormValue("title")
	promo.Description = r.FormValue("description")
	promo.PriceDisplay = r.FormValue("price_display")
	promo.Status = r.FormValue("status")
	if promo.Status == "" {
		promo.Status = "active"
	}
	promo.IsSponsored = r.FormValue("is_sponsored") == "on"
	if storeID, err := strconv.ParseInt(r.FormValue("store_id"), 10, 64); err == nil {
		promo.StoreID = storeID
	}
	startDate, endDate, dateErr := h.comSvc.ParseDateRange(r.FormValue("start_date"), r.FormValue("end_date"))
	if dateErr != "" {
		stores, _ := h.repo.StoreList(r.Context(), false, 1000)
		h.Render(w, r, "admin_promo_form.html", map[string]any{"Error": dateErr, "Stores": stores, "Promo": promo, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	promo.StartDate = startDate
	promo.EndDate = endDate

	if r.FormValue("remove_image") == "on" && promo.ImageKey != "" {
		h.storage.Delete(r.Context(), promo.ImageKey)
		promo.ImageKey = ""
	}
	if key := h.mediaKeyFromRequest(r, "image_media_key"); key != "" {
		promo.ImageKey = key
	}

	file, header, err := r.FormFile("image")
	if err == nil {
		defer file.Close()
		key := storage.GenerateKey(header.Filename)
		if info, err := h.storage.Upload(r.Context(), key, file, header.Header.Get("Content-Type")); err == nil {
			promo.ImageKey = info.Key
			h.recordMediaAssetFromUpload(r, user, header, info, promo.Title)
		}
	}

	newSlug := slug.Generate(promo.Title)
	if newSlug != promo.Slug && !h.repo.PromotionSlugExists(r.Context(), newSlug) {
		promo.Slug = newSlug
	}
	if err := h.repo.PromotionUpdate(r.Context(), promo); err != nil {
		stores, _ := h.repo.StoreList(r.Context(), false, 1000)
		h.Render(w, r, "admin_promo_form.html", map[string]any{"Error": "Erro ao atualizar promocao", "Stores": stores, "Promo": promo, "MediaAssets": h.mediaAssetsForForms(r.Context())})
		return
	}
	h.auditAdminAction(r, user, "update", "promotion", auditEntityID(promo.ID), map[string]any{
		"title":        promo.Title,
		"slug":         promo.Slug,
		"store_id":     promo.StoreID,
		"status":       promo.Status,
		"is_sponsored": promo.IsSponsored,
		"start_date":   promo.StartDate.Format("2006-01-02"),
		"end_date":     promo.EndDate.Format("2006-01-02"),
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/promotions", http.StatusSeeOther)
}

func (h *Handler) AdminPromoDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermPromosManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	promo, _ := h.repo.PromotionGetByID(r.Context(), id)
	h.repo.PromotionDelete(r.Context(), id)
	h.auditAdminAction(r, user, "delete", "promotion", auditEntityID(id), map[string]any{
		"title": promotionTitle(promo),
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/promotions", http.StatusSeeOther)
}
