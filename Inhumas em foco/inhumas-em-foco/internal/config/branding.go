package config

import (
	"fmt"
	"os"
	"strings"
)

// TenantBrandingConfig encapsula todos os atributos de identidade visual
// e metadata de um portal. Todos os campos sao populados exclusivamente
// via variaveis de ambiente, sem fallback para valores hardcoded de portal especifico.
type TenantBrandingConfig struct {
	PortalName        string
	PortalTagline     string
	PortalDescription string
	PortalLocale      string
	PortalLanguage    string
	PortalCategory    string

	SiteURL         string
	AdminPathPrefix string

	LogoPath       string
	LogoAltText    string
	FaviconPath    string
	PrimaryColor   string
	SecondaryColor string
	AccentColor    string

	SEOTitleSuffix  string
	SEODefaultImage string
	TwitterHandle   string
	FacebookPageURL string
	InstagramHandle string

	ContactEmail   string
	ContactPhone   string
	ContactCity    string
	ContactState   string
	ContactCountry string

	ArticlesPerPage   int
	FeaturedTagSlug   string
	BreakingNewsLabel string
	CopyrightHolder   string
	FooterLegalText   string

	GoogleAnalyticsID  string
	GoogleTagManagerID string
	DisqusShortname    string
	RecaptchaSiteKey   string
}

// LoadTenantBrandingConfig le todas as variaveis de ambiente de branding
// e retorna um TenantBrandingConfig populado.
func LoadTenantBrandingConfig() (*TenantBrandingConfig, error) {
	cfg := &TenantBrandingConfig{}
	var missing []string

	cfg.PortalName = os.Getenv("PORTAL_NAME")
	if cfg.PortalName == "" {
		missing = append(missing, "PORTAL_NAME")
	}

	cfg.SiteURL = strings.TrimRight(os.Getenv("SITE_URL"), "/")
	if cfg.SiteURL == "" {
		missing = append(missing, "SITE_URL")
	}

	cfg.ContactEmail = os.Getenv("PORTAL_CONTACT_EMAIL")
	if cfg.ContactEmail == "" {
		missing = append(missing, "PORTAL_CONTACT_EMAIL")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("branding config: variaveis obrigatorias ausentes: %s", strings.Join(missing, ", "))
	}

	cfg.PortalTagline = envOrDefault("PORTAL_TAGLINE", "")
	cfg.PortalDescription = envOrDefault("PORTAL_DESCRIPTION", "")
	cfg.PortalLocale = envOrDefault("PORTAL_LOCALE", "pt_BR")
	cfg.PortalLanguage = envOrDefault("PORTAL_LANGUAGE", "pt-BR")
	cfg.PortalCategory = envOrDefault("PORTAL_CATEGORY", "news")
	cfg.AdminPathPrefix = envOrDefault("ADMIN_PATH_PREFIX", "/painel")

	cfg.LogoPath = envOrDefault("PORTAL_LOGO_PATH", "/static/branding/logo.svg")
	cfg.LogoAltText = envOrDefault("PORTAL_LOGO_ALT", cfg.PortalName)
	cfg.FaviconPath = envOrDefault("PORTAL_FAVICON_PATH", "/static/branding/favicon.ico")
	cfg.PrimaryColor = envOrDefault("PORTAL_PRIMARY_COLOR", "#1a4a3a")
	cfg.SecondaryColor = envOrDefault("PORTAL_SECONDARY_COLOR", "#f5c518")
	cfg.AccentColor = envOrDefault("PORTAL_ACCENT_COLOR", "#2d6a52")

	cfg.SEOTitleSuffix = envOrDefault("PORTAL_SEO_TITLE_SUFFIX", " | "+cfg.PortalName)
	cfg.SEODefaultImage = envOrDefault("PORTAL_SEO_DEFAULT_IMAGE", cfg.SiteURL+"/static/branding/og-default.jpg")
	cfg.TwitterHandle = envOrDefault("PORTAL_TWITTER_HANDLE", "")
	cfg.FacebookPageURL = envOrDefault("PORTAL_FACEBOOK_PAGE_URL", "")
	cfg.InstagramHandle = envOrDefault("PORTAL_INSTAGRAM_HANDLE", "")

	cfg.ContactPhone = envOrDefault("PORTAL_CONTACT_PHONE", "")
	cfg.ContactCity = envOrDefault("PORTAL_CONTACT_CITY", "")
	cfg.ContactState = envOrDefault("PORTAL_CONTACT_STATE", "")
	cfg.ContactCountry = envOrDefault("PORTAL_CONTACT_COUNTRY", "BR")

	cfg.ArticlesPerPage = envOrDefaultInt("PORTAL_ARTICLES_PER_PAGE", 12)
	cfg.FeaturedTagSlug = envOrDefault("PORTAL_FEATURED_TAG_SLUG", "destaque")
	cfg.BreakingNewsLabel = envOrDefault("PORTAL_BREAKING_NEWS_LABEL", "Ao vivo")
	cfg.CopyrightHolder = envOrDefault("PORTAL_COPYRIGHT_HOLDER", cfg.PortalName)
	cfg.FooterLegalText = envOrDefault("PORTAL_FOOTER_LEGAL_TEXT", "Todos os direitos reservados.")

	cfg.GoogleAnalyticsID = envOrDefault("PORTAL_GA_ID", "")
	cfg.GoogleTagManagerID = envOrDefault("PORTAL_GTM_ID", "")
	cfg.DisqusShortname = envOrDefault("PORTAL_DISQUS_SHORTNAME", "")
	cfg.RecaptchaSiteKey = envOrDefault("PORTAL_RECAPTCHA_SITE_KEY", "")

	return cfg, nil
}

func (b *TenantBrandingConfig) FullTitle(pageTitle string) string {
	if pageTitle == "" {
		return b.PortalName
	}
	if strings.HasSuffix(pageTitle, b.SEOTitleSuffix) || pageTitle == b.PortalName {
		return pageTitle
	}
	return pageTitle + b.SEOTitleSuffix
}

func (b *TenantBrandingConfig) AbsoluteURL(path string) string {
	base := strings.TrimRight(b.SiteURL, "/")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return base + path
}

func envOrDefault(key, defaultValue string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultValue
}

func envOrDefaultInt(key string, defaultValue int) int {
	v := os.Getenv(key)
	if v == "" {
		return defaultValue
	}
	var n int
	if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
		return defaultValue
	}
	return n
}
