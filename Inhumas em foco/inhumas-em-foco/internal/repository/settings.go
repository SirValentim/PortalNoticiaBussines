package repository

import (
	"context"
	"database/sql"
	"strings"

	"inhumas-em-foco/internal/model"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func DefaultPortalSettings() model.PortalSettings {
	return model.PortalSettings{
		TenantID:                  1,
		SiteName:                  "Inhumas em Foco",
		Tagline:                   "O portal de noticias e comercio local que conecta Inhumas em um so lugar.",
		ContactEmail:              "contato@inhumasemfoco.online",
		ContactWhatsapp:           "(62) 99999-9999",
		ContactPhone:              "(62) 99999-9999",
		City:                      "Inhumas",
		State:                     "GO",
		SEOTitle:                  "Inhumas em Foco - Noticias, comercio e eventos de Inhumas GO",
		SEODescription:            "Noticias de Inhumas GO, guia comercial, promocoes, eventos e informacoes locais atualizadas.",
		UploadMaxMB:               2,
		AutomationIntervalMinutes: 60,
	}
}

func (r *Repository) PortalSettingsGet(ctx context.Context) (model.PortalSettings, error) {
	return r.PortalSettingsGetForTenant(ctx, tenantIDFromContext(ctx))
}

func (r *Repository) PortalSettingsGetForTenant(ctx context.Context, tenantID int64) (model.PortalSettings, error) {
	settings := DefaultPortalSettings()
	settings.TenantID = tenantID
	row := r.db.QueryRowContext(ctx, `
		SELECT id, tenant_id, site_name, COALESCE(tagline, ''), COALESCE(logo_key, ''), COALESCE(favicon_key, ''),
		       COALESCE(contact_email, ''), COALESCE(contact_whatsapp, ''), COALESCE(contact_phone, ''),
		       COALESCE(city, ''), COALESCE(state, ''), COALESCE(seo_title, ''), COALESCE(seo_description, ''),
		       COALESCE(facebook_url, ''), COALESCE(instagram_url, ''), COALESCE(youtube_url, ''), COALESCE(tiktok_url, ''),
		       COALESCE(upload_max_mb, 2), COALESCE(automation_enabled, false), COALESCE(automation_interval_minutes, 60), updated_at
		FROM portal_settings
		WHERE tenant_id = $1`, tenantID)
	err := row.Scan(
		&settings.ID,
		&settings.TenantID,
		&settings.SiteName,
		&settings.Tagline,
		&settings.LogoKey,
		&settings.FaviconKey,
		&settings.ContactEmail,
		&settings.ContactWhatsapp,
		&settings.ContactPhone,
		&settings.City,
		&settings.State,
		&settings.SEOTitle,
		&settings.SEODescription,
		&settings.FacebookURL,
		&settings.InstagramURL,
		&settings.YoutubeURL,
		&settings.TiktokURL,
		&settings.UploadMaxMB,
		&settings.AutomationEnabled,
		&settings.AutomationIntervalMinutes,
		&settings.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return settings, nil
	}
	return settings, err
}

func (r *Repository) PortalSettingsUpdate(ctx context.Context, settings *model.PortalSettings) error {
	if settings.TenantID <= 0 {
		settings.TenantID = tenantIDFromContext(ctx)
	}
	normalizePortalSettings(settings)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO portal_settings (
			tenant_id, site_name, tagline, logo_key, favicon_key, contact_email, contact_whatsapp, contact_phone,
			city, state, seo_title, seo_description, facebook_url, instagram_url, youtube_url, tiktok_url,
			upload_max_mb, automation_enabled, automation_interval_minutes, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, CURRENT_TIMESTAMP
		)
		ON CONFLICT(tenant_id) DO UPDATE SET
			site_name=excluded.site_name,
			tagline=excluded.tagline,
			logo_key=excluded.logo_key,
			favicon_key=excluded.favicon_key,
			contact_email=excluded.contact_email,
			contact_whatsapp=excluded.contact_whatsapp,
			contact_phone=excluded.contact_phone,
			city=excluded.city,
			state=excluded.state,
			seo_title=excluded.seo_title,
			seo_description=excluded.seo_description,
			facebook_url=excluded.facebook_url,
			instagram_url=excluded.instagram_url,
			youtube_url=excluded.youtube_url,
			tiktok_url=excluded.tiktok_url,
			upload_max_mb=excluded.upload_max_mb,
			automation_enabled=excluded.automation_enabled,
			automation_interval_minutes=excluded.automation_interval_minutes,
			updated_at=CURRENT_TIMESTAMP`,
		settings.TenantID,
		settings.SiteName,
		settings.Tagline,
		settings.LogoKey,
		settings.FaviconKey,
		settings.ContactEmail,
		settings.ContactWhatsapp,
		settings.ContactPhone,
		settings.City,
		settings.State,
		settings.SEOTitle,
		settings.SEODescription,
		settings.FacebookURL,
		settings.InstagramURL,
		settings.YoutubeURL,
		settings.TiktokURL,
		settings.UploadMaxMB,
		settings.AutomationEnabled,
		settings.AutomationIntervalMinutes,
	)
	return err
}

func tenantIDFromContext(ctx context.Context) int64 {
	if tenant := tenantctx.FromContext(ctx); tenant != nil && tenant.ID > 0 {
		return tenant.ID
	}
	return 1
}

func normalizePortalSettings(settings *model.PortalSettings) {
	defaults := DefaultPortalSettings()
	settings.SiteName = firstSetting(settings.SiteName, defaults.SiteName)
	settings.Tagline = firstSetting(settings.Tagline, defaults.Tagline)
	settings.ContactEmail = strings.TrimSpace(settings.ContactEmail)
	settings.ContactWhatsapp = strings.TrimSpace(settings.ContactWhatsapp)
	settings.ContactPhone = strings.TrimSpace(settings.ContactPhone)
	settings.City = firstSetting(settings.City, defaults.City)
	settings.State = firstSetting(strings.ToUpper(settings.State), defaults.State)
	settings.SEOTitle = firstSetting(settings.SEOTitle, defaults.SEOTitle)
	settings.SEODescription = firstSetting(settings.SEODescription, defaults.SEODescription)
	settings.FacebookURL = strings.TrimSpace(settings.FacebookURL)
	settings.InstagramURL = strings.TrimSpace(settings.InstagramURL)
	settings.YoutubeURL = strings.TrimSpace(settings.YoutubeURL)
	settings.TiktokURL = strings.TrimSpace(settings.TiktokURL)
	if settings.UploadMaxMB <= 0 {
		settings.UploadMaxMB = defaults.UploadMaxMB
	}
	if settings.UploadMaxMB > 50 {
		settings.UploadMaxMB = 50
	}
	if settings.AutomationIntervalMinutes <= 0 {
		settings.AutomationIntervalMinutes = defaults.AutomationIntervalMinutes
	}
}

func firstSetting(value, fallback string) string {
	if strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}
