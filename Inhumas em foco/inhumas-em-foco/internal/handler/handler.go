package handler

import (
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"

	"inhumas-em-foco/internal/auth"
	commercialsvc "inhumas-em-foco/internal/commercial"
	"inhumas-em-foco/internal/config"
	"inhumas-em-foco/internal/editorialai"
	"inhumas-em-foco/internal/middleware"
	"inhumas-em-foco/internal/model"
	postsvc "inhumas-em-foco/internal/posts"
	"inhumas-em-foco/internal/repository"
	"inhumas-em-foco/internal/session"
	"inhumas-em-foco/internal/sitemap"
	"inhumas-em-foco/internal/storage"
	tenantctx "inhumas-em-foco/internal/tenant"
	usersvc "inhumas-em-foco/internal/users"
)

type Handler struct {
	repo      *repository.Repository
	cfg       *config.Config
	session   *session.Manager
	authSvc   *auth.Service
	resetSvc  *usersvc.PasswordResetService
	userSvc   *usersvc.AdminService
	postSvc   *postsvc.EditorialService
	comSvc    *commercialsvc.Service
	aiSvc     editorialai.Provider
	storage   storage.Provider
	sanitizer *bluemonday.Policy
	templates *template.Template
}

func (h *Handler) metricContext(r *http.Request) context.Context {
	if tenant := tenantctx.FromContext(r.Context()); tenant != nil {
		return tenantctx.WithTenant(context.Background(), tenant)
	}
	return context.Background()
}

func New(repo *repository.Repository, cfg *config.Config, session *session.Manager, authSvc *auth.Service, storage storage.Provider) (*Handler, error) {
	h := &Handler{
		repo:      repo,
		cfg:       cfg,
		session:   session,
		authSvc:   authSvc,
		resetSvc:  usersvc.NewPasswordResetService(repo, authSvc, cfg),
		userSvc:   usersvc.NewAdminService(repo, authSvc, cfg),
		postSvc:   postsvc.NewEditorialService(authSvc),
		comSvc:    commercialsvc.NewService(),
		aiSvc:     editorialai.NewMockProvider(),
		storage:   storage,
		sanitizer: bluemonday.UGCPolicy(),
	}
	if err := h.loadTemplates(); err != nil {
		return nil, err
	}
	return h, nil
}

func (h *Handler) loadTemplates() error {
	h.templates = template.New("").Funcs(h.funcMap())
	return nil
}

func (h *Handler) funcMap() template.FuncMap {
	return template.FuncMap{
		"friendlyID": friendlyID,
		"formatDate": func(value any) string {
			t := templateTime(value)
			if t.IsZero() {
				return ""
			}
			return t.Format("02/01/2006")
		},
		"formatDateTime": func(value any) string {
			t := templateTime(value)
			if t.IsZero() {
				return ""
			}
			return t.Format("02/01/2006 15:04")
		},
		"datetimeLocal": func(value any) string {
			t := templateTime(value)
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02T15:04")
		},
		"dateInput": func(value any) string {
			t := templateTime(value)
			if t.IsZero() {
				return ""
			}
			return t.Format("2006-01-02")
		},
		"htmlSafe": func(s string) template.HTML {
			return template.HTML(h.sanitizer.Sanitize(s))
		},
		"add": func(a, b int) int {
			return a + b
		},
		"sub": func(a, b int) int {
			return a - b
		},
		"imageURL": func(key string) string {
			if key == "" {
				if h.cfg != nil && h.cfg.Branding != nil {
					return h.cfg.Branding.SEODefaultImage
				}
				return "/static/branding/og-default.jpg"
			}
			return h.storage.URL(context.Background(), key)
		},
		"formatBytes": func(size int64) string {
			if size < 1024 {
				return fmt.Sprintf("%d B", size)
			}
			if size < 1024*1024 {
				return fmt.Sprintf("%.1f KB", float64(size)/1024)
			}
			return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
		},
		"metricLabel":         metricLabel,
		"bannerStatusLabel":   bannerStatusLabel,
		"bannerPositionLabel": bannerPositionLabel,
		"bannerDaysLeft":      bannerDaysLeft,
		"bannerCTR":           bannerCTR,
		"postStatusLabel":     postStatusLabel,
		"selected": func(a, b string) bool {
			return a == b
		},
		"idPtrEq": func(ptr *int64, id int64) bool {
			return ptr != nil && *ptr == id
		},
		"roleLabel": func(role model.UserRole) string {
			return role.Label()
		},
		"hasPermission": func(user *model.User, perm auth.Permission) bool {
			return h.authSvc.HasPermission(user, perm)
		},
		"auditEntityIDLabel":         auditEntityIDLabel,
		"auditUserLabel":             auditUserLabel,
		"eventStatusLabel":           eventStatusLabel,
		"classifiedStatusLabel":      classifiedStatusLabel,
		"storeCommercialStatusLabel": storeCommercialStatusLabel,
		"promoStatusLabel":           promoStatusLabel,
		"promoValidityLabel":         promoValidityLabel,
		"aiActionLabel":              aiActionLabel,
	}
}

func templateTime(value any) time.Time {
	switch v := value.(type) {
	case time.Time:
		return v
	case *time.Time:
		if v != nil {
			return *v
		}
	}
	return time.Time{}
}

func (h *Handler) Render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	user := auth.UserFromContext(r.Context())
	if data == nil {
		data = make(map[string]any)
	}
	branding := h.branding(r.Context())
	if _, ok := data["Branding"]; !ok {
		data["Branding"] = branding
	}
	settings := h.portalSettings(r.Context())
	if _, ok := data["PortalSettings"]; !ok {
		data["PortalSettings"] = settings
	}
	if seo, ok := data["SEO"].(model.SEOData); ok {
		h.normalizeSEO(r, &seo)
		data["SEO"] = seo
	}
	data["User"] = user
	data["AdminSystemName"] = "NewsCore CMS"
	data["AdminSystemDescriptor"] = "CMS Premium Multiportal"
	data["AdminSystemTagline"] = "Gestao profissional de portais de noticias"
	data["CurrentPortalName"] = firstNonEmpty(branding.PortalName, settings.SiteName, "Portal")
	data["CanManageTenants"] = h.authSvc.HasPermission(user, auth.PermTenantsManage)
	data["CanManagePosts"] = h.authSvc.HasPermission(user, auth.PermPostsCreate) ||
		h.authSvc.HasPermission(user, auth.PermPostsEditAny) ||
		h.authSvc.HasPermission(user, auth.PermPostsEditOwn) ||
		h.authSvc.HasPermission(user, auth.PermPostsApprove)
	data["CanCreatePosts"] = h.authSvc.HasPermission(user, auth.PermPostsCreate)
	data["CanManageEditorialTaxonomy"] = h.authSvc.HasPermission(user, auth.PermSettingsManage)
	data["CanManageMedia"] = h.authSvc.HasPermission(user, auth.PermMediaManage)
	data["CanManageUsers"] = h.authSvc.HasPermission(user, auth.PermUsersManage)
	data["CanManageSettings"] = h.authSvc.HasPermission(user, auth.PermSettingsManage)
	data["CanManageAutomation"] = h.authSvc.HasPermission(user, auth.PermAutomationManage)
	data["CanManageCommercial"] = h.authSvc.HasPermission(user, auth.PermBannersManage) ||
		h.authSvc.HasPermission(user, auth.PermStoresManage) ||
		h.authSvc.HasPermission(user, auth.PermPromosManage) ||
		h.authSvc.HasPermission(user, auth.PermEventsManage) ||
		h.authSvc.HasPermission(user, auth.PermClassifiedsManage) ||
		h.authSvc.HasPermission(user, auth.PermInfluencersManage)
	data["MediaFeatureEnabled"] = h.tenantFeatureEnabled(r, "media", true)
	data["CommercialFeatureEnabled"] = h.tenantFeatureEnabled(r, "commercial", true)
	data["AdminPath"] = branding.AdminPathPrefix
	data["Year"] = time.Now().Year()
	data["CSRFToken"] = middleware.CSRFToken(r.Context())
	if _, ok := data["Active"]; !ok {
		data["Active"] = ""
	}
	if _, ok := data["Title"]; !ok {
		data["Title"] = "Painel"
	}
	if _, ok := data["JSONLD"]; !ok {
		if seo, ok := data["SEO"].(model.SEOData); ok {
			data["JSONLD"] = h.organizationJSONLD(seo)
		}
	}

	tmpl, layout, err := h.templateFor(name)
	if err != nil {
		http.Error(w, "Erro ao carregar template", http.StatusInternalServerError)
		return
	}

	var buf strings.Builder
	if err := tmpl.ExecuteTemplate(&buf, layout, data); err != nil {
		log.Printf("erro ao renderizar template %s com layout %s: %v", name, layout, err)
		http.Error(w, "Erro ao renderizar", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(buf.String()))
}

func (h *Handler) normalizeSEO(r *http.Request, seo *model.SEOData) {
	branding := h.branding(r.Context())
	siteURL := strings.TrimRight(branding.SiteURL, "/")
	if siteURL == "" {
		scheme := "http"
		if r.TLS != nil || strings.EqualFold(r.Header.Get("X-Forwarded-Proto"), "https") {
			scheme = "https"
		}
		siteURL = scheme + "://" + r.Host
	}
	settings := h.portalSettings(r.Context())
	if seo.Title == "" {
		seo.Title = settings.SEOTitle
	}
	if seo.Description == "" {
		seo.Description = settings.SEODescription
	}
	if seo.Type == "" {
		seo.Type = "website"
	}
	if seo.URL == "" {
		seo.URL = siteURL + r.URL.Path
	}
	if seo.CanonicalURL == "" {
		seo.CanonicalURL = seo.URL
	}
	if seo.Image == "" {
		if settings.LogoKey != "" {
			seo.Image = h.storage.URL(r.Context(), settings.LogoKey)
		} else {
			seo.Image = branding.SEODefaultImage
		}
	}
}

func (h *Handler) branding(ctx context.Context) *config.TenantBrandingConfig {
	if branding := middleware.BrandingFromContext(ctx); branding != nil {
		return branding
	}
	if h.cfg != nil && h.cfg.Branding != nil {
		return h.cfg.Branding
	}
	siteURL := ""
	adminPath := "/painel"
	if h.cfg != nil {
		siteURL = strings.TrimRight(h.cfg.SiteURL, "/")
		adminPath = h.cfg.AdminPathPrefix
	}
	return &config.TenantBrandingConfig{
		PortalName:        "Portal",
		PortalLocale:      "pt_BR",
		PortalLanguage:    "pt-BR",
		PortalCategory:    "news",
		SiteURL:           siteURL,
		AdminPathPrefix:   adminPath,
		LogoPath:          "/static/branding/logo.svg",
		LogoAltText:       "Portal",
		FaviconPath:       "/static/branding/favicon.ico",
		PrimaryColor:      "#1a4a3a",
		SecondaryColor:    "#f5c518",
		AccentColor:       "#2d6a52",
		SEOTitleSuffix:    " | Portal",
		SEODefaultImage:   strings.TrimRight(siteURL, "/") + "/static/branding/og-default.jpg",
		ContactCountry:    "BR",
		ArticlesPerPage:   12,
		FeaturedTagSlug:   "destaque",
		BreakingNewsLabel: "Ao vivo",
		CopyrightHolder:   "Portal",
		FooterLegalText:   "Todos os direitos reservados.",
	}
}

func (h *Handler) siteURL(ctx context.Context) string {
	return strings.TrimRight(h.branding(ctx).SiteURL, "/")
}

func (h *Handler) pageTitle(ctx context.Context, title string) string {
	return h.branding(ctx).FullTitle(title)
}

func (h *Handler) portalSettings(ctx context.Context) model.PortalSettings {
	settings, err := h.repo.PortalSettingsGet(ctx)
	if err != nil {
		return repository.DefaultPortalSettings()
	}
	return settings
}

func (h *Handler) templateFor(name string) (*template.Template, string, error) {
	layout := "base.html"
	if name == "login.html" {
		layout = "auth.html"
	} else if strings.HasPrefix(name, "admin_") {
		layout = "admin.html"
	}

	files := []string{
		filepath.Join(h.cfg.ProjectRoot, "internal", "view", "layouts", layout),
		filepath.Join(h.cfg.ProjectRoot, "internal", "view", "pages", name),
	}
	components, err := filepath.Glob(filepath.Join(h.cfg.ProjectRoot, "internal", "view", "components", "*.html"))
	if err != nil {
		return nil, "", err
	}
	files = append(files, components...)

	tmpl, err := template.New(layout).Funcs(h.funcMap()).ParseFiles(files...)
	if err != nil {
		return nil, "", fmt.Errorf("parse %s: %w", name, err)
	}
	return tmpl, layout, nil
}

func friendlyID(value any) string {
	var id int64
	switch v := value.(type) {
	case int:
		id = int64(v)
	case int64:
		id = v
	case *int64:
		if v != nil {
			id = *v
		}
	case model.Post:
		id = v.ID
	case *model.Post:
		if v != nil {
			id = v.ID
		}
	case model.User:
		id = v.ID
	case *model.User:
		if v != nil {
			id = v.ID
		}
	case model.Tenant:
		id = v.ID
	case *model.Tenant:
		if v != nil {
			id = v.ID
		}
	case model.Category:
		id = v.ID
	case *model.Category:
		if v != nil {
			id = v.ID
		}
	case model.Tag:
		id = v.ID
	case *model.Tag:
		if v != nil {
			id = v.ID
		}
	case model.MetricEntityTotal:
		id = v.EntityID
	case *model.MetricEntityTotal:
		if v != nil {
			id = v.EntityID
		}
	}
	if id <= 0 {
		return "-"
	}
	return fmt.Sprintf("#%03d", id)
}

func (h *Handler) RenderError(w http.ResponseWriter, status int, msg string) {
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprintf(w, "<h1>Erro %d</h1><p>%s</p>", status, msg)
}

func (h *Handler) parseMultipart(w http.ResponseWriter, r *http.Request) error {
	limit := h.effectiveMaxUploadSize(r.Context())
	r.Body = http.MaxBytesReader(w, r.Body, limit)
	if !strings.HasPrefix(strings.ToLower(r.Header.Get("Content-Type")), "multipart/form-data") {
		return r.ParseForm()
	}
	return r.ParseMultipartForm(limit)
}

func (h *Handler) effectiveMaxUploadSize(ctx context.Context) int64 {
	settings := h.portalSettings(ctx)
	if settings.UploadMaxMB > 0 {
		return int64(settings.UploadMaxMB) * 1024 * 1024
	}
	if h.cfg.MaxUploadSize > 0 {
		return h.cfg.MaxUploadSize
	}
	return 2 * 1024 * 1024
}

// Routes

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	// Public
	mux.HandleFunc("GET /", h.Home)
	mux.HandleFunc("GET /noticias", h.NewsList)
	mux.HandleFunc("GET /noticias/mais", h.NewsListMore)
	mux.HandleFunc("GET /noticia/{slug}", h.PostDetail)
	mux.HandleFunc("GET /post/{slug}", h.PostAlias)
	mux.HandleFunc("GET /categoria/{slug}", h.CategoryPosts)
	mux.HandleFunc("GET /tag/{slug}", h.TagPosts)
	mux.HandleFunc("GET /eventos", h.EventList)
	mux.HandleFunc("GET /evento/{slug}", h.EventDetail)
	mux.HandleFunc("GET /loja/{slug}", h.StoreDetail)
	mux.HandleFunc("GET /lojas", h.StoreList)
	mux.HandleFunc("GET /influencer/{slug}", h.InfluencerDetail)
	mux.HandleFunc("GET /influencers", h.InfluencerList)
	mux.HandleFunc("GET /promocao/{slug}", h.PromoDetail)
	mux.HandleFunc("GET /promocoes", h.PromoList)
	mux.HandleFunc("GET /classificados", h.Classifieds)
	mux.HandleFunc("GET /classificado/{slug}", h.ClassifiedDetail)
	mux.HandleFunc("GET /bairro/{slug}", h.NeighborhoodDetail)
	mux.HandleFunc("GET /busca", h.Search)
	mux.HandleFunc("GET /login", h.LoginPage)
	mux.HandleFunc("POST /login", h.LoginPost)
	mux.HandleFunc("GET /logout", h.Logout)
	mux.HandleFunc("GET /recuperar-senha", h.PasswordResetRequestPage)
	mux.HandleFunc("POST /recuperar-senha", h.PasswordResetRequestPost)
	mux.HandleFunc("GET /redefinir-senha/{token}", h.PasswordResetPage)
	mux.HandleFunc("POST /redefinir-senha/{token}", h.PasswordResetPost)
	mux.HandleFunc("GET /health", h.Health)
	mux.HandleFunc("GET /metrics", h.OperationalMetrics)
	mux.HandleFunc("GET /sobre", h.About)
	mux.HandleFunc("GET /contato", h.Contact)
	mux.HandleFunc("GET /robots.txt", h.Robots)
	mux.HandleFunc("GET /sitemap.xml", h.Sitemap)
	mux.HandleFunc("GET /rss.xml", h.RSS)
	mux.HandleFunc("GET /manifest.json", h.Manifest)

	// Static
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir(h.cfg.StaticDir))))
	mux.Handle("GET /uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(h.cfg.UploadDir))))

	// Admin routes
	admin := h.cfg.AdminPathPrefix
	mux.HandleFunc("GET "+admin+"", h.AdminDashboard)
	mux.HandleFunc("GET "+admin+"/posts", h.AdminPosts)
	mux.HandleFunc("GET "+admin+"/posts/new", h.AdminPostNew)
	mux.HandleFunc("POST "+admin+"/posts", h.AdminPostCreate)
	mux.HandleFunc("GET "+admin+"/posts/{id}", h.AdminPostDetail)
	mux.HandleFunc("GET "+admin+"/posts/{id}/edit", h.AdminPostEdit)
	mux.HandleFunc("GET "+admin+"/posts/{id}/preview", h.AdminPostPreview)
	mux.HandleFunc("POST "+admin+"/posts/{id}", h.AdminPostUpdate)
	mux.HandleFunc("POST "+admin+"/posts/{id}/autosave", h.AdminPostAutosave)
	mux.HandleFunc("POST "+admin+"/posts/{id}/lock", h.AdminPostLockHeartbeat)
	mux.HandleFunc("POST "+admin+"/posts/{id}/duplicate", h.AdminPostDuplicate)
	mux.HandleFunc("POST "+admin+"/posts/{id}/archive", h.AdminPostArchive)
	mux.HandleFunc("POST "+admin+"/posts/{id}/delete", h.AdminPostDelete)
	mux.HandleFunc("POST "+admin+"/posts/{id}/submit-review", h.AdminPostSubmitReview)
	mux.HandleFunc("POST "+admin+"/posts/{id}/approve", h.AdminPostApprove)
	mux.HandleFunc("POST "+admin+"/posts/{id}/reject", h.AdminPostReject)
	mux.HandleFunc("POST "+admin+"/posts/{id}/publish", h.AdminPostPublish)
	mux.HandleFunc("POST "+admin+"/posts/{id}/ai/{action}", h.AdminPostAIAction)

	mux.HandleFunc("GET "+admin+"/automation", h.AdminAutomation)
	mux.HandleFunc("GET "+admin+"/automation/sources/new", h.AdminAutomationSourceNew)
	mux.HandleFunc("POST "+admin+"/automation/sources", h.AdminAutomationSourceCreate)
	mux.HandleFunc("GET "+admin+"/automation/sources/{id}/edit", h.AdminAutomationSourceEdit)
	mux.HandleFunc("POST "+admin+"/automation/sources/{id}", h.AdminAutomationSourceUpdate)
	mux.HandleFunc("POST "+admin+"/automation/sources/{id}/delete", h.AdminAutomationSourceDelete)
	mux.HandleFunc("POST "+admin+"/automation/sources/{id}/run", h.AdminAutomationSourceRun)
	mux.HandleFunc("POST "+admin+"/automation/run-all", h.AdminAutomationRunAll)

	mux.HandleFunc("GET "+admin+"/categories", h.AdminCategories)
	mux.HandleFunc("GET "+admin+"/categories/new", h.AdminCategoryNew)
	mux.HandleFunc("POST "+admin+"/categories", h.AdminCategoryCreate)
	mux.HandleFunc("GET "+admin+"/categories/{id}/edit", h.AdminCategoryEdit)
	mux.HandleFunc("POST "+admin+"/categories/{id}", h.AdminCategoryUpdate)
	mux.HandleFunc("POST "+admin+"/categories/{id}/delete", h.AdminCategoryDelete)

	mux.HandleFunc("GET "+admin+"/tags", h.AdminTags)
	mux.HandleFunc("GET "+admin+"/tags/new", h.AdminTagNew)
	mux.HandleFunc("POST "+admin+"/tags", h.AdminTagCreate)
	mux.HandleFunc("GET "+admin+"/tags/{id}/edit", h.AdminTagEdit)
	mux.HandleFunc("POST "+admin+"/tags/{id}", h.AdminTagUpdate)
	mux.HandleFunc("POST "+admin+"/tags/{id}/delete", h.AdminTagDelete)

	mux.HandleFunc("GET "+admin+"/media", h.AdminMedia)
	mux.HandleFunc("POST "+admin+"/media", h.AdminMediaUpload)
	mux.HandleFunc("POST "+admin+"/media/{id}", h.AdminMediaUpdate)
	mux.HandleFunc("POST "+admin+"/media/{id}/delete", h.AdminMediaDelete)

	mux.HandleFunc("GET "+admin+"/stores", h.AdminStores)
	mux.HandleFunc("GET "+admin+"/stores/new", h.AdminStoreNew)
	mux.HandleFunc("POST "+admin+"/stores", h.AdminStoreCreate)
	mux.HandleFunc("GET "+admin+"/stores/{id}/edit", h.AdminStoreEdit)
	mux.HandleFunc("POST "+admin+"/stores/{id}", h.AdminStoreUpdate)
	mux.HandleFunc("POST "+admin+"/stores/{id}/delete", h.AdminStoreDelete)

	mux.HandleFunc("GET "+admin+"/influencers", h.AdminInfluencers)
	mux.HandleFunc("GET "+admin+"/influencers/new", h.AdminInfluencerNew)
	mux.HandleFunc("POST "+admin+"/influencers", h.AdminInfluencerCreate)
	mux.HandleFunc("GET "+admin+"/influencers/{id}/edit", h.AdminInfluencerEdit)
	mux.HandleFunc("POST "+admin+"/influencers/{id}", h.AdminInfluencerUpdate)
	mux.HandleFunc("POST "+admin+"/influencers/{id}/delete", h.AdminInfluencerDelete)

	mux.HandleFunc("GET "+admin+"/banners", h.AdminBanners)
	mux.HandleFunc("GET "+admin+"/banners/export.csv", h.AdminBannersExportCSV)
	mux.HandleFunc("GET "+admin+"/banners/new", h.AdminBannerNew)
	mux.HandleFunc("POST "+admin+"/banners", h.AdminBannerCreate)
	mux.HandleFunc("GET "+admin+"/banners/{id}/edit", h.AdminBannerEdit)
	mux.HandleFunc("POST "+admin+"/banners/{id}", h.AdminBannerUpdate)
	mux.HandleFunc("POST "+admin+"/banners/{id}/delete", h.AdminBannerDelete)

	mux.HandleFunc("GET "+admin+"/promotions", h.AdminPromotions)
	mux.HandleFunc("GET "+admin+"/promotions/new", h.AdminPromoNew)
	mux.HandleFunc("POST "+admin+"/promotions", h.AdminPromoCreate)
	mux.HandleFunc("GET "+admin+"/promotions/{id}/edit", h.AdminPromoEdit)
	mux.HandleFunc("POST "+admin+"/promotions/{id}", h.AdminPromoUpdate)
	mux.HandleFunc("POST "+admin+"/promotions/{id}/delete", h.AdminPromoDelete)

	mux.HandleFunc("GET "+admin+"/events", h.AdminEvents)
	mux.HandleFunc("GET "+admin+"/events/new", h.AdminEventNew)
	mux.HandleFunc("POST "+admin+"/events", h.AdminEventCreate)
	mux.HandleFunc("GET "+admin+"/events/{id}/edit", h.AdminEventEdit)
	mux.HandleFunc("POST "+admin+"/events/{id}", h.AdminEventUpdate)
	mux.HandleFunc("POST "+admin+"/events/{id}/delete", h.AdminEventDelete)

	mux.HandleFunc("GET "+admin+"/classifieds", h.AdminClassifieds)
	mux.HandleFunc("GET "+admin+"/classifieds/new", h.AdminClassifiedNew)
	mux.HandleFunc("POST "+admin+"/classifieds", h.AdminClassifiedCreate)
	mux.HandleFunc("GET "+admin+"/classifieds/{id}/edit", h.AdminClassifiedEdit)
	mux.HandleFunc("POST "+admin+"/classifieds/{id}", h.AdminClassifiedUpdate)
	mux.HandleFunc("POST "+admin+"/classifieds/{id}/delete", h.AdminClassifiedDelete)

	mux.HandleFunc("GET "+admin+"/neighborhoods", h.AdminNeighborhoods)
	mux.HandleFunc("POST "+admin+"/neighborhoods", h.AdminNeighborhoodCreate)
	mux.HandleFunc("POST "+admin+"/neighborhoods/{id}/delete", h.AdminNeighborhoodDelete)

	mux.HandleFunc("GET "+admin+"/users", h.AdminUsers)
	mux.HandleFunc("POST "+admin+"/users", h.AdminUserCreate)
	mux.HandleFunc("GET "+admin+"/users/{id}", h.AdminUserDetail)
	mux.HandleFunc("GET "+admin+"/users/{id}/edit", h.AdminUserEdit)
	mux.HandleFunc("POST "+admin+"/users/{id}/edit", h.AdminUserUpdate)
	mux.HandleFunc("POST "+admin+"/users/{id}/password", h.AdminUserUpdatePassword)
	mux.HandleFunc("POST "+admin+"/users/{id}/activate", h.AdminUserActivate)
	mux.HandleFunc("POST "+admin+"/users/{id}/deactivate", h.AdminUserDeactivate)
	mux.HandleFunc("POST "+admin+"/users/{id}/delete", h.AdminUserDelete)
	mux.HandleFunc("GET "+admin+"/profile", h.AdminProfile)
	mux.HandleFunc("POST "+admin+"/profile/password", h.AdminProfilePasswordUpdate)

	mux.HandleFunc("GET "+admin+"/tenants", h.AdminTenants)
	mux.HandleFunc("POST "+admin+"/tenants", h.AdminTenantCreate)
	mux.HandleFunc("GET "+admin+"/tenants/{id}", h.AdminTenantDetail)
	mux.HandleFunc("GET "+admin+"/tenants/{id}/edit", h.AdminTenantEdit)
	mux.HandleFunc("POST "+admin+"/tenants/{id}", h.AdminTenantUpdate)
	mux.HandleFunc("POST "+admin+"/tenants/{id}/deactivate", h.AdminTenantDeactivate)
	mux.HandleFunc("POST "+admin+"/tenants/{id}/domains", h.AdminTenantDomainCreate)
	mux.HandleFunc("POST "+admin+"/tenants/{id}/domains/{domainID}/delete", h.AdminTenantDomainDelete)
	mux.HandleFunc("POST "+admin+"/tenants/{id}/features", h.AdminTenantFeatureUpsert)
	mux.HandleFunc("POST "+admin+"/tenants/{id}/users", h.AdminTenantUserUpsert)
	mux.HandleFunc("POST "+admin+"/tenants/{id}/users/{userID}/delete", h.AdminTenantUserDelete)

	mux.HandleFunc("GET "+admin+"/metrics", h.AdminMetrics)
	mux.HandleFunc("GET "+admin+"/dead-jobs", h.AdminDeadJobs)
	mux.HandleFunc("GET "+admin+"/audit", h.AdminAuditLogs)
	mux.HandleFunc("GET "+admin+"/settings", h.AdminSettings)
	mux.HandleFunc("POST "+admin+"/settings", h.AdminSettingsUpdate)

	// API
	mux.HandleFunc("POST /api/metrics/{type}", h.APITrackMetric)
}

// Home Handler
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	ctx := r.Context()

	// Get data
	headlinePost, _ := h.repo.PostGetFeaturedPublished(ctx)
	latestOffset := 0
	if headlinePost == nil {
		headline, _ := h.repo.PostListPublished(ctx, 1, 0)
		if len(headline) > 0 {
			headlinePost = &headline[0]
			latestOffset = 1
		}
	}
	latest, _ := h.repo.PostListPublished(ctx, 6, latestOffset)

	catBastidores, _ := h.repo.CategoryGetBySlug(ctx, "politica-bastidores")
	var bastidores []model.Post
	if catBastidores != nil {
		bastidores, _ = h.repo.PostListByCategory(ctx, catBastidores.ID, 4)
	}

	heroBanner, _ := h.repo.BannerGetActiveByPosition(ctx, "hero")
	inFeedBanner, _ := h.repo.BannerGetActiveByPosition(ctx, "in_feed")
	stickyBanner, _ := h.repo.BannerGetActiveByPosition(ctx, "sticky_footer")

	featuredStores, _ := h.repo.StoreListFeatured(ctx, 6)
	if len(featuredStores) == 0 {
		featuredStores, _ = h.repo.StoreList(ctx, true, 6)
	}

	featuredPromos, _ := h.repo.PromotionListActive(ctx, 3)

	neighborhoods, _ := h.repo.NeighborhoodList(ctx)
	influencers, _ := h.repo.InfluencerList(ctx, true, 6)

	seo := model.SEOData{
		Title:       h.pageTitle(ctx, ""),
		Description: firstNonEmpty(h.branding(ctx).PortalDescription, "Portal de noticias, comercio e eventos locais."),
		URL:         h.siteURL(ctx) + "/",
		Type:        "website",
		Tags:        []string{h.branding(ctx).PortalName, h.branding(ctx).PortalCategory},
	}

	data := map[string]any{
		"SEO":            seo,
		"JSONLD":         h.homeJSONLD(seo),
		"Headline":       headlinePost,
		"LatestNews":     latest,
		"Bastidores":     bastidores,
		"HeroBanner":     heroBanner,
		"InFeedBanner":   inFeedBanner,
		"StickyBanner":   stickyBanner,
		"FeaturedStores": featuredStores,
		"FeaturedPromos": featuredPromos,
		"Neighborhoods":  neighborhoods,
		"Influencers":    influencers,
		"HomeData":       true,
	}

	h.Render(w, r, "home.html", data)
}

// Post Detail
func (h *Handler) PostDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	post, err := h.repo.PostGetBySlug(r.Context(), slug)
	if err != nil || post == nil {
		// Check redirect
		redirect, _ := h.repo.SlugRedirectGet(r.Context(), slug)
		if redirect != nil {
			http.Redirect(w, r, "/noticia/"+redirect.NewSlug, http.StatusMovedPermanently)
			return
		}
		http.NotFound(w, r)
		return
	}

	if post.Status != "published" {
		user := auth.UserFromContext(r.Context())
		canEditAny := h.authSvc.HasPermission(user, auth.PermPostsEditAny)
		canEditOwn := h.authSvc.HasPermission(user, auth.PermPostsEditOwn) && post.AuthorID != nil && user.ID == *post.AuthorID
		if user == nil || (!canEditAny && !canEditOwn) {
			http.NotFound(w, r)
			return
		}
	}

	// Track view
	go h.repo.MetricTrack(h.metricContext(r), &model.Metric{
		MetricType: "post_view",
		EntityType: "post",
		EntityID:   post.ID,
		IPAddress:  middleware.ClientIP(r),
		UserAgent:  r.UserAgent(),
		Referrer:   r.Referer(),
	})

	// Related posts
	related, _ := h.repo.PostListByCategory(r.Context(), *post.CategoryID, 3)
	sidebarTopBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "sidebar_top")
	sidebarBottomBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "sidebar_bottom")

	seo := model.SEOData{
		Title:        firstNonEmpty(post.MetaTitle, h.pageTitle(r.Context(), post.Title)),
		Description:  firstNonEmpty(post.MetaDescription, post.Excerpt),
		URL:          h.siteURL(r.Context()) + "/noticia/" + post.Slug,
		CanonicalURL: firstNonEmpty(post.CanonicalURL, h.siteURL(r.Context())+"/noticia/"+post.Slug),
		Type:         "article",
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
	})
}

func (h *Handler) Robots(w http.ResponseWriter, r *http.Request) {
	siteURL := h.siteURL(r.Context())
	if siteURL == "" {
		siteURL = "https://" + r.Host
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprintf(w, "User-agent: *\nAllow: /\nDisallow: %s/\nDisallow: /login\nDisallow: /busca\n\nSitemap: %s/sitemap.xml\n", strings.TrimRight(h.branding(r.Context()).AdminPathPrefix, "/"), siteURL)
}

func (h *Handler) Sitemap(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/xml; charset=utf-8")
	path := filepath.Join(h.cfg.StaticDir, "sitemap.xml")
	if _, err := os.Stat(path); err == nil {
		http.ServeFile(w, r, path)
		return
	} else if err != nil && !os.IsNotExist(err) {
		http.Error(w, "Erro ao carregar sitemap", http.StatusInternalServerError)
		return
	}

	entries, err := h.repo.SitemapEntries(r.Context())
	if err != nil {
		http.Error(w, "Erro ao gerar sitemap", http.StatusInternalServerError)
		return
	}
	data, err := sitemap.Build(h.siteURL(r.Context()), entries)
	if err != nil {
		http.Error(w, "Erro ao gerar sitemap", http.StatusInternalServerError)
		return
	}
	w.Write(data)
}

func (h *Handler) PostAlias(w http.ResponseWriter, r *http.Request) {
	http.Redirect(w, r, "/noticia/"+r.PathValue("slug"), http.StatusMovedPermanently)
}

// Category Posts
func (h *Handler) CategoryPosts(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	cat, err := h.repo.CategoryGetBySlug(r.Context(), slug)
	if err != nil || cat == nil {
		http.NotFound(w, r)
		return
	}

	posts, _ := h.repo.PostListByCategory(r.Context(), cat.ID, 21)
	hasMore := len(posts) > 20
	if hasMore {
		posts = posts[:20]
	}
	listBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "in_feed")

	seo := model.SEOData{
		Title:       firstNonEmpty(cat.MetaTitle, h.pageTitle(r.Context(), cat.Name)),
		Description: firstNonEmpty(cat.MetaDescription, cat.Description, "Noticias sobre "+cat.Name+" em Inhumas, Goias"),
		URL:         h.siteURL(r.Context()) + "/categoria/" + cat.Slug,
		Tags:        []string{"Inhumas", "Inhumas GO", cat.Name},
	}

	h.Render(w, r, "category.html", map[string]any{
		"SEO":        seo,
		"JSONLD":     h.collectionJSONLD(seo, cat.Name, "/categoria/"+cat.Slug),
		"Category":   cat,
		"Posts":      posts,
		"HasMore":    hasMore,
		"NextPage":   2,
		"ListBanner": listBanner,
	})
}

func (h *Handler) TagPosts(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	tag, err := h.repo.TagGetBySlug(r.Context(), slug)
	if err != nil || tag == nil || !tag.Active {
		http.NotFound(w, r)
		return
	}

	posts, _ := h.repo.PostListByTag(r.Context(), tag.ID, 21)
	hasMore := len(posts) > 20
	if hasMore {
		posts = posts[:20]
	}
	listBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "in_feed")
	description := tag.Description
	if description == "" {
		description = "Noticias marcadas com " + tag.Name + " em Inhumas, Goias"
	}

	seo := model.SEOData{
		Title:       firstNonEmpty(tag.MetaTitle, h.pageTitle(r.Context(), tag.Name)),
		Description: firstNonEmpty(tag.MetaDescription, description),
		URL:         h.siteURL(r.Context()) + "/tag/" + tag.Slug,
		Tags:        []string{"Inhumas", "Inhumas GO", tag.Name},
	}

	h.Render(w, r, "tag.html", map[string]any{
		"SEO":        seo,
		"JSONLD":     h.collectionJSONLD(seo, tag.Name, "/tag/"+tag.Slug),
		"Tag":        tag,
		"Posts":      posts,
		"HasMore":    hasMore,
		"NextPage":   2,
		"ListBanner": listBanner,
	})
}

// NewsList renders the general news archive.
func (h *Handler) NewsList(w http.ResponseWriter, r *http.Request) {
	posts, _ := h.repo.PostListPublished(r.Context(), 21, 0)
	hasMore := len(posts) > 20
	if hasMore {
		posts = posts[:20]
	}
	listBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "in_feed")
	cat := &model.Category{
		Name:        "Noticias",
		Description: "Tudo que acontece em Inhumas e regiao.",
	}
	seo := model.SEOData{
		Title:       h.pageTitle(r.Context(), "Noticias"),
		Description: "Ultimas noticias de Inhumas GO, politica local, cidade, eventos, comercio e servicos para moradores.",
		URL:         h.siteURL(r.Context()) + "/noticias",
		Tags:        []string{"noticias de Inhumas", "Inhumas GO", "portal de noticias Inhumas"},
	}
	h.Render(w, r, "category.html", map[string]any{
		"SEO":        seo,
		"JSONLD":     h.collectionJSONLD(seo, "Noticias", "/noticias"),
		"Category":   cat,
		"Posts":      posts,
		"HasMore":    hasMore,
		"NextPage":   2,
		"ListBanner": listBanner,
	})
}

func (h *Handler) NewsListMore(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 2 {
		page = 2
	}
	const pageSize = 10
	posts, _ := h.repo.PostListPublished(r.Context(), pageSize+1, (page-1)*pageSize)
	hasMore := len(posts) > pageSize
	if hasMore {
		posts = posts[:pageSize]
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.renderNewsRows(w, posts, hasMore, page+1); err != nil {
		log.Printf("erro ao renderizar noticias parciais: %v", err)
	}
}

func (h *Handler) renderNewsRows(w http.ResponseWriter, posts []model.Post, hasMore bool, nextPage int) error {
	const partial = `{{range .Posts}}
<article class="news-row">
	<a href="/noticia/{{.Slug}}">
		<img src="{{if .CoverImageKey}}{{imageURL .CoverImageKey}}{{else}}{{imageURL ""}}{{end}}" alt="{{.Title}}" loading="lazy">
	</a>
	<div>
		<span class="badge">{{.CategoryName}}</span>
		<h2><a href="/noticia/{{.Slug}}">{{.Title}}</a></h2>
		<p>{{.Excerpt}}</p>
		<small>{{.PublishedAt | formatDate}}</small>
	</div>
</article>
{{end}}
{{if .HasMore}}
<div id="news-load-more" class="load-more-wrap">
	<button class="btn btn-primary" hx-get="/noticias/mais?page={{.NextPage}}" hx-target="#news-load-more" hx-swap="outerHTML">Carregar mais</button>
</div>
{{end}}`
	tmpl, err := template.New("news-partial").Funcs(h.funcMap()).Parse(partial)
	if err != nil {
		return err
	}
	return tmpl.Execute(w, map[string]any{
		"Posts":    posts,
		"HasMore":  hasMore,
		"NextPage": nextPage,
	})
}

func (h *Handler) EventList(w http.ResponseWriter, r *http.Request) {
	events, _ := h.repo.EventList(r.Context(), true, 100)
	featuredEvents, regularEvents := splitCommercialEvents(events, 6)
	listBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "in_feed")
	seo := model.SEOData{
		Title:       h.pageTitle(r.Context(), "Eventos"),
		Description: "Eventos e agenda local de Inhumas, Goias.",
		URL:         h.siteURL(r.Context()) + "/eventos",
		Tags:        []string{"eventos em Inhumas", "agenda Inhumas", "Inhumas GO"},
	}
	h.Render(w, r, "event_list.html", map[string]any{
		"SEO":            seo,
		"JSONLD":         h.collectionJSONLD(seo, "Eventos", "/eventos"),
		"Events":         regularEvents,
		"FeaturedEvents": featuredEvents,
		"HasEvents":      len(events) > 0,
		"ListBanner":     listBanner,
	})
}

func (h *Handler) EventDetail(w http.ResponseWriter, r *http.Request) {
	event, err := h.repo.EventGetBySlug(r.Context(), r.PathValue("slug"))
	if err != nil || event == nil || event.Status != "active" {
		http.NotFound(w, r)
		return
	}
	seo := model.SEOData{
		Title:       firstNonEmpty(event.MetaTitle, h.pageTitle(r.Context(), event.Title)),
		Description: firstNonEmpty(event.MetaDescription, event.Description),
		URL:         h.siteURL(r.Context()) + "/evento/" + event.Slug,
		Type:        "event",
		Tags:        []string{"eventos em Inhumas", event.Title, event.Location},
	}
	if event.ImageKey != "" {
		seo.Image = h.storage.URL(r.Context(), event.ImageKey)
	}
	h.Render(w, r, "event_detail.html", map[string]any{
		"SEO":    seo,
		"JSONLD": h.eventJSONLD(event, seo),
		"Event":  event,
	})
}

// Store Detail
func (h *Handler) StoreDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	store, err := h.repo.StoreGetBySlug(r.Context(), slug)
	if err != nil || store == nil {
		http.NotFound(w, r)
		return
	}

	go h.repo.MetricTrack(h.metricContext(r), &model.Metric{
		MetricType: "store_view",
		EntityType: "store",
		EntityID:   store.ID,
		IPAddress:  middleware.ClientIP(r),
		UserAgent:  r.UserAgent(),
	})

	promos, _ := h.repo.PromotionListActive(r.Context(), 5)
	// Filter by store
	var storePromos []model.Promotion
	for _, p := range promos {
		if p.StoreID == store.ID {
			storePromos = append(storePromos, p)
		}
	}

	seo := model.SEOData{
		Title:       firstNonEmpty(store.MetaTitle, h.pageTitle(r.Context(), store.Name)),
		Description: firstNonEmpty(store.MetaDescription, store.Description),
		URL:         h.siteURL(r.Context()) + "/loja/" + store.Slug,
		Type:        "profile",
		Tags:        []string{store.Name, store.Category, "lojas em Inhumas", "Inhumas GO"},
	}
	if store.CoverImageKey != "" {
		seo.Image = h.storage.URL(r.Context(), store.CoverImageKey)
	} else if store.LogoKey != "" {
		seo.Image = h.storage.URL(r.Context(), store.LogoKey)
	}

	h.Render(w, r, "store_detail.html", map[string]any{
		"SEO":    seo,
		"JSONLD": h.storeJSONLD(store, seo),
		"Store":  store,
		"Promos": storePromos,
	})
}

// Store List
func (h *Handler) StoreList(w http.ResponseWriter, r *http.Request) {
	stores, _ := h.repo.StoreList(r.Context(), true, 50)
	category := strings.TrimSpace(r.URL.Query().Get("categoria"))
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if category != "" || query != "" {
		filtered := make([]model.Store, 0, len(stores))
		for _, store := range stores {
			if category != "" && !strings.EqualFold(store.Category, category) {
				continue
			}
			if query != "" && !strings.Contains(strings.ToLower(store.Name+" "+store.Description+" "+store.Address), strings.ToLower(query)) {
				continue
			}
			filtered = append(filtered, store)
		}
		stores = filtered
	}
	featuredStores, regularStores := splitCommercialStores(stores, 6)

	seo := model.SEOData{
		Title:       h.pageTitle(r.Context(), "Lojas e Comercios"),
		Description: "Guia comercial de Inhumas GO com lojas, servicos, contato, bairros e empresas locais.",
		URL:         h.siteURL(r.Context()) + "/lojas",
		Tags:        []string{"lojas em Inhumas", "guia comercial Inhumas", "empresas em Inhumas GO"},
	}
	listBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "in_feed")

	h.Render(w, r, "store_list.html", map[string]any{
		"SEO":            seo,
		"JSONLD":         h.collectionJSONLD(seo, "Lojas", "/lojas"),
		"Stores":         regularStores,
		"FeaturedStores": featuredStores,
		"HasStores":      len(stores) > 0,
		"Query":          query,
		"ListBanner":     listBanner,
	})
}

func (h *Handler) InfluencerList(w http.ResponseWriter, r *http.Request) {
	influencers, _ := h.repo.InfluencerList(r.Context(), true, 100)
	niches := influencerNiches(influencers)
	selectedNiche := strings.TrimSpace(r.URL.Query().Get("niche"))
	if selectedNiche != "" {
		filtered := make([]model.Influencer, 0, len(influencers))
		for _, influencer := range influencers {
			if strings.EqualFold(strings.TrimSpace(influencer.Niche), selectedNiche) {
				filtered = append(filtered, influencer)
			}
		}
		influencers = filtered
	}
	seo := model.SEOData{
		Title:       h.pageTitle(r.Context(), "Influencers da Cidade"),
		Description: "Conheca criadores, comunicadores e personalidades de Inhumas.",
		URL:         h.siteURL(r.Context()) + "/influencers",
		Tags:        []string{"influencers de Inhumas", "criadores de Inhumas", "Inhumas GO"},
	}
	listBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "in_feed")
	h.Render(w, r, "influencer_list.html", map[string]any{
		"SEO":         seo,
		"JSONLD":      h.collectionJSONLD(seo, "Influencers da Cidade", "/influencers"),
		"Influencers": influencers,
		"Niches":      niches,
		"ActiveNiche": selectedNiche,
		"ListBanner":  listBanner,
	})
}

func (h *Handler) InfluencerDetail(w http.ResponseWriter, r *http.Request) {
	influencer, err := h.repo.InfluencerGetBySlug(r.Context(), r.PathValue("slug"))
	if err != nil || influencer == nil || !influencer.Active {
		http.NotFound(w, r)
		return
	}
	go h.repo.MetricTrack(h.metricContext(r), &model.Metric{
		MetricType: "influencer_view",
		EntityType: "influencer",
		EntityID:   influencer.ID,
		IPAddress:  middleware.ClientIP(r),
		UserAgent:  r.UserAgent(),
	})
	seo := model.SEOData{
		Title:       firstNonEmpty(influencer.MetaTitle, influencer.Name+" | Influencers da Cidade"),
		Description: firstNonEmpty(influencer.MetaDescription, influencer.Bio),
		URL:         h.siteURL(r.Context()) + "/influencer/" + influencer.Slug,
		Type:        "profile",
		Tags:        influencerSEOTags(influencer),
	}
	if influencer.CoverImageKey != "" {
		seo.Image = h.storage.URL(r.Context(), influencer.CoverImageKey)
	} else if influencer.AvatarKey != "" {
		seo.Image = h.storage.URL(r.Context(), influencer.AvatarKey)
	}
	h.Render(w, r, "influencer_detail.html", map[string]any{
		"SEO":        seo,
		"Influencer": influencer,
	})
}

func influencerNiches(influencers []model.Influencer) []string {
	seen := make(map[string]bool)
	var niches []string
	for _, influencer := range influencers {
		niche := strings.TrimSpace(influencer.Niche)
		if niche == "" || seen[strings.ToLower(niche)] {
			continue
		}
		seen[strings.ToLower(niche)] = true
		niches = append(niches, niche)
	}
	return niches
}

func influencerSEOTags(influencer *model.Influencer) []string {
	tags := []string{influencer.Name, "influencer de Inhumas", "Inhumas GO"}
	if niche := strings.TrimSpace(influencer.Niche); niche != "" {
		tags = append(tags, niche+" em Inhumas")
	}
	if area := strings.TrimSpace(influencer.CityArea); area != "" {
		tags = append(tags, area)
	}
	return tags
}

func splitCommercialStores(stores []model.Store, limit int) ([]model.Store, []model.Store) {
	featured := make([]model.Store, 0)
	regular := make([]model.Store, 0, len(stores))
	for _, store := range stores {
		if (store.IsFeatured || store.IsSponsored) && (limit <= 0 || len(featured) < limit) {
			featured = append(featured, store)
			continue
		}
		regular = append(regular, store)
	}
	return featured, regular
}

func splitCommercialEvents(events []model.Event, limit int) ([]model.Event, []model.Event) {
	featured := make([]model.Event, 0)
	regular := make([]model.Event, 0, len(events))
	for _, event := range events {
		if (event.IsFeatured || event.IsSponsored) && (limit <= 0 || len(featured) < limit) {
			featured = append(featured, event)
			continue
		}
		regular = append(regular, event)
	}
	return featured, regular
}

func splitCommercialClassifieds(classifieds []model.Classified, limit int) ([]model.Classified, []model.Classified) {
	featured := make([]model.Classified, 0)
	regular := make([]model.Classified, 0, len(classifieds))
	for _, classified := range classifieds {
		if (classified.IsFeatured || classified.IsSponsored) && (limit <= 0 || len(featured) < limit) {
			featured = append(featured, classified)
			continue
		}
		regular = append(regular, classified)
	}
	return featured, regular
}

type localCommercialSummary struct {
	Total     int
	Active    int
	Featured  int
	Sponsored int
}

func localCommercialSummaryStores(stores []model.Store) localCommercialSummary {
	var summary localCommercialSummary
	summary.Total = len(stores)
	for _, store := range stores {
		if store.Active {
			summary.Active++
		}
		if store.IsFeatured {
			summary.Featured++
		}
		if store.IsSponsored {
			summary.Sponsored++
		}
	}
	return summary
}

func localCommercialSummaryEvents(events []model.Event) localCommercialSummary {
	var summary localCommercialSummary
	summary.Total = len(events)
	for _, event := range events {
		if event.Status == "active" {
			summary.Active++
		}
		if event.IsFeatured {
			summary.Featured++
		}
		if event.IsSponsored {
			summary.Sponsored++
		}
	}
	return summary
}

func localCommercialSummaryClassifieds(classifieds []model.Classified) localCommercialSummary {
	var summary localCommercialSummary
	summary.Total = len(classifieds)
	for _, classified := range classifieds {
		if classified.Status == "active" {
			summary.Active++
		}
		if classified.IsFeatured {
			summary.Featured++
		}
		if classified.IsSponsored {
			summary.Sponsored++
		}
	}
	return summary
}

// Promo Detail
func (h *Handler) PromoDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	promo, err := h.repo.PromotionGetBySlug(r.Context(), slug)
	if err != nil || promo == nil || !promotionPubliclyVisible(promo) {
		http.NotFound(w, r)
		return
	}

	go h.repo.MetricTrack(h.metricContext(r), &model.Metric{
		MetricType: "promo_click",
		EntityType: "promotion",
		EntityID:   promo.ID,
		IPAddress:  middleware.ClientIP(r),
		UserAgent:  r.UserAgent(),
	})

	store, _ := h.repo.StoreGetBySlug(r.Context(), promo.StoreSlug)

	seo := model.SEOData{
		Title:       firstNonEmpty(promo.MetaTitle, h.pageTitle(r.Context(), promo.Title)),
		Description: firstNonEmpty(promo.MetaDescription, promo.Description),
		URL:         h.siteURL(r.Context()) + "/promocao/" + promo.Slug,
		Type:        "article",
		Tags:        []string{"promocoes em Inhumas", promo.StoreName, "Inhumas GO"},
	}
	if promo.ImageKey != "" {
		seo.Image = h.storage.URL(r.Context(), promo.ImageKey)
	}

	h.Render(w, r, "promo_detail.html", map[string]any{
		"SEO":   seo,
		"Promo": promo,
		"Store": store,
	})
}

// Promo List
func (h *Handler) PromoList(w http.ResponseWriter, r *http.Request) {
	promos, _ := h.repo.PromotionListActive(r.Context(), 20)
	category := strings.TrimSpace(r.URL.Query().Get("categoria"))
	if category != "" {
		filtered := make([]model.Promotion, 0, len(promos))
		for _, promo := range promos {
			if strings.EqualFold(promo.StoreName, category) || strings.Contains(strings.ToLower(promo.Title+" "+promo.Description), strings.ToLower(category)) {
				filtered = append(filtered, promo)
			}
		}
		promos = filtered
	}

	seo := model.SEOData{
		Title:       h.pageTitle(r.Context(), "Promocoes do Dia"),
		Description: "Promocoes em Inhumas GO, ofertas de lojas locais, servicos, alimentacao e comercio da cidade.",
		URL:         h.siteURL(r.Context()) + "/promocoes",
		Tags:        []string{"promocoes em Inhumas", "ofertas Inhumas", "comercio local Inhumas"},
	}
	listBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "in_feed")

	h.Render(w, r, "promo_list.html", map[string]any{
		"SEO":        seo,
		"JSONLD":     h.collectionJSONLD(seo, "Promocoes", "/promocoes"),
		"Promos":     promos,
		"ListBanner": listBanner,
	})
}

func (h *Handler) Classifieds(w http.ResponseWriter, r *http.Request) {
	filter := repository.ClassifiedFilter{
		Query:      strings.TrimSpace(r.URL.Query().Get("q")),
		Category:   strings.TrimSpace(r.URL.Query().Get("categoria")),
		ActiveOnly: true,
		Limit:      100,
	}
	classifieds, _ := h.repo.ClassifiedList(r.Context(), filter)
	featuredClassifieds, regularClassifieds := splitCommercialClassifieds(classifieds, 6)
	listBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "in_feed")
	seo := model.SEOData{
		Title:       h.pageTitle(r.Context(), "Classificados"),
		Description: "Anuncios locais de imoveis, veiculos, empregos, servicos e oportunidades em Inhumas.",
		URL:         h.siteURL(r.Context()) + "/classificados",
		Tags:        []string{"classificados Inhumas", "empregos Inhumas", "imoveis Inhumas"},
	}
	h.Render(w, r, "classifieds.html", map[string]any{
		"SEO":                 seo,
		"JSONLD":              h.collectionJSONLD(seo, "Classificados", "/classificados"),
		"Classifieds":         regularClassifieds,
		"FeaturedClassifieds": featuredClassifieds,
		"HasClassifieds":      len(classifieds) > 0,
		"Query":               filter.Query,
		"Category":            filter.Category,
		"ListBanner":          listBanner,
	})
}

func (h *Handler) ClassifiedDetail(w http.ResponseWriter, r *http.Request) {
	classified, err := h.repo.ClassifiedGetBySlug(r.Context(), r.PathValue("slug"))
	if err != nil || classified == nil || classified.Status != "active" || classifiedExpired(classified) {
		http.NotFound(w, r)
		return
	}
	seo := model.SEOData{
		Title:       firstNonEmpty(classified.MetaTitle, h.pageTitle(r.Context(), classified.Title)),
		Description: firstNonEmpty(classified.MetaDescription, classified.Description),
		URL:         h.siteURL(r.Context()) + "/classificado/" + classified.Slug,
		Type:        "product",
		Tags:        []string{"classificados Inhumas", classified.Category, classified.Title},
	}
	if classified.ImageKey != "" {
		seo.Image = h.storage.URL(r.Context(), classified.ImageKey)
	}
	h.Render(w, r, "classified_detail.html", map[string]any{
		"SEO":        seo,
		"JSONLD":     h.classifiedJSONLD(classified, seo),
		"Classified": classified,
	})
}

// Neighborhood Detail
func (h *Handler) NeighborhoodDetail(w http.ResponseWriter, r *http.Request) {
	slug := r.PathValue("slug")
	neighborhood, err := h.repo.NeighborhoodGetBySlug(r.Context(), slug)
	if err != nil || neighborhood == nil {
		http.NotFound(w, r)
		return
	}

	seo := model.SEOData{
		Title:       neighborhood.MetaTitle,
		Description: neighborhood.MetaDescription,
	}
	if seo.Title == "" {
		seo.Title = h.pageTitle(r.Context(), neighborhood.Name)
	}
	if seo.Description == "" {
		seo.Description = "Noticias e comercio do bairro " + neighborhood.Name + " em Inhumas, GO"
	}
	seo.URL = h.siteURL(r.Context()) + "/bairro/" + neighborhood.Slug
	seo.Tags = []string{neighborhood.Name, "bairro em Inhumas", "Inhumas GO"}

	stores, _ := h.repo.StoreList(r.Context(), true, 20)
	var neighborhoodStores []model.Store
	for _, s := range stores {
		if s.NeighborhoodID != nil && *s.NeighborhoodID == neighborhood.ID {
			neighborhoodStores = append(neighborhoodStores, s)
		}
	}

	h.Render(w, r, "neighborhood.html", map[string]any{
		"SEO":          seo,
		"Neighborhood": neighborhood,
		"Stores":       neighborhoodStores,
	})
}

// Search
func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	var posts []model.Post
	if query != "" {
		posts, _ = h.repo.PostSearch(r.Context(), query, 20)
	}

	seo := model.SEOData{
		Title:       h.pageTitle(r.Context(), "Busca: "+query),
		Description: "Resultados da busca por " + query,
		URL:         h.siteURL(r.Context()) + "/busca",
		NoIndex:     true,
	}
	listBanner, _ := h.repo.BannerGetActiveByPosition(r.Context(), "in_feed")

	h.Render(w, r, "search.html", map[string]any{
		"SEO":        seo,
		"Query":      query,
		"Posts":      posts,
		"ListBanner": listBanner,
	})
}

// Auth
func (h *Handler) LoginPage(w http.ResponseWriter, r *http.Request) {
	if auth.UserFromContext(r.Context()) != nil {
		http.Redirect(w, r, h.cfg.AdminPathPrefix, http.StatusSeeOther)
		return
	}
	h.Render(w, r, "login.html", map[string]any{"Title": "Login"})
}

func (h *Handler) LoginPost(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	password := r.FormValue("password")
	clientIP := middleware.ClientIP(r)

	// Check login attempts
	attempts, _ := h.repo.LoginAttemptCountRecent(r.Context(), clientIP, 30)
	if attempts >= 5 {
		h.Render(w, r, "login.html", map[string]any{"Title": "Login", "Error": "Muitas tentativas. Aguarde 30 minutos."})
		return
	}

	user, err := h.authSvc.Authenticate(r.Context(), email, password, h.cfg.DefaultBcryptCost)
	if err != nil {
		h.repo.LoginAttemptCreate(r.Context(), &model.LoginAttempt{
			IPAddress: clientIP,
			Email:     email,
			Success:   false,
		})
		h.Render(w, r, "login.html", map[string]any{"Title": "Login", "Error": "Credenciais invalidas"})
		return
	}

	h.repo.LoginAttemptCreate(r.Context(), &model.LoginAttempt{
		IPAddress: clientIP,
		Email:     email,
		Success:   true,
	})

	h.session.SetUserID(r, w, user.ID)
	http.Redirect(w, r, h.cfg.AdminPathPrefix, http.StatusSeeOther)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	h.session.Clear(r, w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) PasswordResetRequestPage(w http.ResponseWriter, r *http.Request) {
	if auth.UserFromContext(r.Context()) != nil {
		http.Redirect(w, r, h.cfg.AdminPathPrefix, http.StatusSeeOther)
		return
	}
	h.Render(w, r, "password_reset_request.html", map[string]any{
		"SEO": model.SEOData{
			Title:       h.pageTitle(r.Context(), "Recuperar senha"),
			Description: "Solicite a redefinicao de senha do painel " + h.branding(r.Context()).PortalName + ".",
			NoIndex:     true,
		},
	})
}

func (h *Handler) PasswordResetRequestPost(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimSpace(strings.ToLower(r.FormValue("email")))
	data := map[string]any{
		"SEO": model.SEOData{
			Title:       h.pageTitle(r.Context(), "Recuperar senha"),
			Description: "Solicite a redefinicao de senha do painel " + h.branding(r.Context()).PortalName + ".",
			NoIndex:     true,
		},
		"Success": "Se o e-mail estiver cadastrado e ativo, enviaremos instrucoes para redefinir a senha.",
	}

	result := h.resetSvc.Request(r.Context(), email, middleware.ClientIP(r), r.UserAgent())
	if result.User != nil {
		h.auditAdminAction(r, nil, "password_reset_requested", "user", auditEntityID(result.User.ID), map[string]any{
			"email": result.User.Email,
		})
		switch result.EmailStatus {
		case usersvc.PasswordResetEmailSent:
			h.auditAdminAction(r, nil, "password_reset_email_sent", "user", auditEntityID(result.User.ID), map[string]any{
				"email": result.User.Email,
			})
		case usersvc.PasswordResetEmailError:
			h.auditAdminAction(r, nil, "password_reset_email_error", "user", auditEntityID(result.User.ID), map[string]any{
				"email": result.User.Email,
				"error": result.EmailError,
			})
		case usersvc.PasswordResetEmailNotConfigured:
			h.auditAdminAction(r, nil, "password_reset_email_not_configured", "user", auditEntityID(result.User.ID), map[string]any{
				"email": result.User.Email,
			})
		}
		if h.cfg.IsLocal() && result.ResetURL != "" {
			data["ResetURL"] = result.ResetURL
		}
	}

	h.Render(w, r, "password_reset_request.html", data)
}

func (h *Handler) PasswordResetPage(w http.ResponseWriter, r *http.Request) {
	tokenValue := r.PathValue("token")
	reset, user, ok := h.resetSvc.Context(r.Context(), tokenValue)
	data := map[string]any{
		"SEO": model.SEOData{
			Title:       h.pageTitle(r.Context(), "Redefinir senha"),
			Description: "Defina uma nova senha de acesso ao painel " + h.branding(r.Context()).PortalName + ".",
			NoIndex:     true,
		},
		"Token": tokenValue,
	}
	if !ok || reset == nil || user == nil {
		data["Error"] = "Link invalido ou expirado. Solicite uma nova recuperacao de senha."
		h.Render(w, r, "password_reset_form.html", data)
		return
	}
	data["CanReset"] = true
	h.Render(w, r, "password_reset_form.html", data)
}

func (h *Handler) PasswordResetPost(w http.ResponseWriter, r *http.Request) {
	tokenValue := r.PathValue("token")
	reset, user, ok := h.resetSvc.Context(r.Context(), tokenValue)
	data := map[string]any{
		"SEO": model.SEOData{
			Title:       h.pageTitle(r.Context(), "Redefinir senha"),
			Description: "Defina uma nova senha de acesso ao painel " + h.branding(r.Context()).PortalName + ".",
			NoIndex:     true,
		},
		"Token": tokenValue,
	}
	if !ok || reset == nil || user == nil {
		data["Error"] = "Link invalido ou expirado. Solicite uma nova recuperacao de senha."
		h.Render(w, r, "password_reset_form.html", data)
		return
	}
	data["CanReset"] = true

	user, msg, completed := h.resetSvc.Complete(r.Context(), tokenValue, r.FormValue("password"), r.FormValue("password_confirm"))
	if !completed {
		data["Error"] = msg
		h.Render(w, r, "password_reset_form.html", data)
		return
	}
	h.auditAdminAction(r, nil, "password_reset_completed", "user", auditEntityID(user.ID), map[string]any{
		"email": user.Email,
	})
	h.Render(w, r, "password_reset_form.html", map[string]any{
		"SEO":     data["SEO"],
		"Success": "Senha atualizada com sucesso. Voce ja pode entrar no painel.",
	})
}

// Health
func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	status := "ok"
	checks := map[string]string{}

	dbStatus := "ok"
	if err := h.repo.DB().PingContext(r.Context()); err != nil {
		dbStatus = "error"
		status = "error"
	}
	checks["database"] = dbStatus

	uploadsStatus := "ok"
	if err := os.MkdirAll(h.cfg.UploadDir, 0755); err != nil {
		uploadsStatus = "error"
		status = "error"
	} else {
		probe, err := os.CreateTemp(h.cfg.UploadDir, ".health-*")
		if err != nil {
			uploadsStatus = "error"
			status = "error"
		} else {
			name := probe.Name()
			_ = probe.Close()
			_ = os.Remove(name)
		}
	}
	checks["uploads"] = uploadsStatus

	jobsStatus := "ok"
	if _, err := h.repo.JobPendingCount(r.Context()); err != nil {
		jobsStatus = "error"
		status = "error"
	}
	checks["jobs"] = jobsStatus

	checks["backup"] = h.backupHealthStatus()

	response := map[string]any{
		"status":    status,
		"timestamp": time.Now().Unix(),
		"version":   "1.0.0",
		"checks":    checks,
	}

	if status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}

	json.NewEncoder(w).Encode(response)
}

func (h *Handler) backupHealthStatus() string {
	if h.cfg.BackupDir == "" {
		return "not_configured"
	}
	entries, err := os.ReadDir(h.cfg.BackupDir)
	if err != nil {
		return "missing"
	}
	cutoff := time.Now().Add(-26 * time.Hour)
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".gz") {
			continue
		}
		info, err := entry.Info()
		if err == nil && info.ModTime().After(cutoff) {
			return "ok"
		}
	}
	return "stale"
}

func (h *Handler) OperationalMetrics(w http.ResponseWriter, r *http.Request) {
	if !h.metricsAuthorized(r) {
		http.Error(w, "Acesso negado", http.StatusUnauthorized)
		return
	}

	pendingJobs, pendingErr := h.repo.JobPendingCount(r.Context())
	deadJobs, deadErr := h.repo.DeadJobCount(r.Context())
	dbStats := h.repo.DB().Stats()

	status := "ok"
	checks := map[string]string{
		"database": "ok",
		"uploads":  "ok",
		"jobs":     "ok",
	}
	if err := h.repo.DB().PingContext(r.Context()); err != nil {
		checks["database"] = "error"
		status = "error"
	}
	if _, err := os.Stat(h.cfg.UploadDir); err != nil {
		checks["uploads"] = "error"
		status = "error"
	}
	if pendingErr != nil || deadErr != nil {
		checks["jobs"] = "error"
		status = "error"
	}

	snapshot := middleware.MetricsSnapshot()
	response := map[string]any{
		"status":         status,
		"timestamp":      time.Now().Unix(),
		"started_at":     snapshot.StartedAt.Format(time.RFC3339),
		"uptime_seconds": snapshot.UptimeSeconds,
		"requests_total": snapshot.RequestsTotal,
		"database": map[string]any{
			"open_connections": dbStats.OpenConnections,
			"in_use":           dbStats.InUse,
			"idle":             dbStats.Idle,
			"wait_count":       dbStats.WaitCount,
		},
		"jobs": map[string]any{
			"pending": pendingJobs,
			"dead":    deadJobs,
		},
		"disk": map[string]any{
			"upload_bytes": dirSize(h.cfg.UploadDir),
		},
		"checks": checks,
	}

	w.Header().Set("Content-Type", "application/json")
	if status != "ok" {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(response)
}

func (h *Handler) metricsAuthorized(r *http.Request) bool {
	if h.cfg.MetricsToken != "" {
		token := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			token = r.Header.Get("X-Metrics-Token")
		}
		if token == "" {
			token = r.URL.Query().Get("token")
		}
		return token == h.cfg.MetricsToken
	}
	return auth.UserFromContext(r.Context()) != nil
}

func dirSize(root string) int64 {
	var total int64
	_ = filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err == nil {
			total += info.Size()
		}
		return nil
	})
	return total
}

// RSS exposes the latest published posts for readers and aggregators.
func (h *Handler) RSS(w http.ResponseWriter, r *http.Request) {
	posts, err := h.repo.PostListPublished(r.Context(), 30, 0)
	if err != nil {
		http.Error(w, "Erro ao gerar RSS", http.StatusInternalServerError)
		return
	}

	branding := h.branding(r.Context())
	feed := rssFeed{
		Version: "2.0",
		AtomNS:  "http://www.w3.org/2005/Atom",
		Channel: rssChannel{
			Title:       branding.PortalName,
			Link:        h.siteURL(r.Context()) + "/",
			Description: branding.PortalDescription,
			Language:    branding.PortalLanguage,
			LastBuild:   time.Now().Format(time.RFC1123Z),
			AtomLink: rssAtomLink{
				Href: h.siteURL(r.Context()) + "/rss.xml",
				Rel:  "self",
				Type: "application/rss+xml",
			},
		},
	}

	for _, post := range posts {
		publishedAt := post.CreatedAt
		if post.PublishedAt != nil {
			publishedAt = *post.PublishedAt
		}
		item := rssItem{
			Title:       post.Title,
			Link:        h.siteURL(r.Context()) + "/noticia/" + post.Slug,
			GUID:        h.siteURL(r.Context()) + "/noticia/" + post.Slug,
			Description: post.Excerpt,
			PubDate:     publishedAt.Format(time.RFC1123Z),
		}
		if post.CategoryName != "" {
			item.Category = post.CategoryName
		}
		feed.Channel.Items = append(feed.Channel.Items, item)
	}

	w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(xml.Header))
	if err := xml.NewEncoder(w).Encode(feed); err != nil {
		return
	}
}

func (h *Handler) Manifest(w http.ResponseWriter, r *http.Request) {
	branding := h.branding(r.Context())
	manifest := map[string]any{
		"name":             branding.PortalName,
		"short_name":       branding.PortalName,
		"description":      branding.PortalDescription,
		"start_url":        "/",
		"display":          "standalone",
		"background_color": branding.PrimaryColor,
		"theme_color":      branding.PrimaryColor,
		"icons": []map[string]string{
			{"src": "/static/branding/icon-192.png", "sizes": "192x192", "type": "image/png"},
			{"src": "/static/branding/icon-512.png", "sizes": "512x512", "type": "image/png"},
		},
	}

	w.Header().Set("Content-Type", "application/manifest+json")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	if err := json.NewEncoder(w).Encode(manifest); err != nil {
		http.Error(w, "erro ao serializar manifest", http.StatusInternalServerError)
	}
}

func (h *Handler) organizationJSONLD(seo model.SEOData) template.JS {
	return mustJSONScript(h.organizationPayload(seo))
}

func (h *Handler) homeJSONLD(seo model.SEOData) template.JS {
	branding := h.branding(context.Background())
	return graphJSONScript(h.organizationPayload(seo), map[string]any{
		"@type":       "WebSite",
		"name":        branding.PortalName,
		"url":         h.siteURL(context.Background()) + "/",
		"inLanguage":  branding.PortalLanguage,
		"description": seo.Description,
		"potentialAction": map[string]any{
			"@type":       "SearchAction",
			"target":      h.siteURL(context.Background()) + "/busca?q={search_term_string}",
			"query-input": "required name=search_term_string",
		},
	})
}

func (h *Handler) collectionJSONLD(seo model.SEOData, name string, path string) template.JS {
	branding := h.branding(context.Background())
	return graphJSONScript(
		h.organizationPayload(seo),
		map[string]any{
			"@type":       "CollectionPage",
			"name":        name,
			"description": seo.Description,
			"url":         seo.URL,
			"inLanguage":  branding.PortalLanguage,
			"isPartOf": map[string]any{
				"@type": "WebSite",
				"name":  branding.PortalName,
				"url":   h.siteURL(context.Background()) + "/",
			},
		},
		h.breadcrumbPayload([]breadcrumbItem{
			{Name: "Home", Path: "/"},
			{Name: name, Path: path},
		}),
	)
}

func (h *Handler) storeJSONLD(store *model.Store, seo model.SEOData) template.JS {
	settings := h.portalSettings(context.Background())
	branding := h.branding(context.Background())
	payload := map[string]any{
		"@type":       "LocalBusiness",
		"name":        store.Name,
		"description": firstNonEmpty(store.Description, seo.Description),
		"url":         seo.URL,
		"inLanguage":  branding.PortalLanguage,
		"address": map[string]any{
			"@type":           "PostalAddress",
			"streetAddress":   store.Address,
			"addressLocality": settings.City,
			"addressRegion":   settings.State,
			"addressCountry":  "BR",
		},
		"areaServed": map[string]any{
			"@type": "City",
			"name":  settings.City,
		},
	}
	if store.Phone != "" {
		payload["telephone"] = store.Phone
	}
	if seo.Image != "" {
		payload["image"] = seo.Image
	} else if store.LogoKey != "" {
		payload["image"] = h.storage.URL(context.Background(), store.LogoKey)
	}
	return graphJSONScript(h.organizationPayload(seo), payload, h.breadcrumbPayload([]breadcrumbItem{
		{Name: "Home", Path: "/"},
		{Name: "Lojas", Path: "/lojas"},
		{Name: store.Name, Path: "/loja/" + store.Slug},
	}))
}

func (h *Handler) eventJSONLD(event *model.Event, seo model.SEOData) template.JS {
	settings := h.portalSettings(context.Background())
	payload := map[string]any{
		"@type":               "Event",
		"name":                event.Title,
		"description":         firstNonEmpty(event.MetaDescription, event.Description),
		"url":                 seo.URL,
		"startDate":           event.StartAt.Format(time.RFC3339),
		"eventStatus":         "https://schema.org/EventScheduled",
		"eventAttendanceMode": "https://schema.org/OfflineEventAttendanceMode",
		"location": map[string]any{
			"@type": "Place",
			"name":  firstNonEmpty(event.Location, settings.City),
			"address": map[string]any{
				"@type":           "PostalAddress",
				"addressLocality": settings.City,
				"addressRegion":   settings.State,
				"addressCountry":  "BR",
			},
		},
		"organizer": map[string]any{
			"@type": "Organization",
			"name":  firstNonEmpty(event.Organizer, settings.SiteName),
		},
	}
	if event.EndAt != nil {
		payload["endDate"] = event.EndAt.Format(time.RFC3339)
	}
	if seo.Image != "" {
		payload["image"] = []string{seo.Image}
	}
	if event.PriceDisplay != "" || event.TicketURL != "" {
		offer := map[string]any{
			"@type":        "Offer",
			"availability": "https://schema.org/InStock",
		}
		if event.TicketURL != "" {
			offer["url"] = event.TicketURL
		}
		if event.PriceDisplay != "" {
			offer["priceSpecification"] = event.PriceDisplay
		}
		payload["offers"] = offer
	}
	return graphJSONScript(h.organizationPayload(seo), payload, h.breadcrumbPayload([]breadcrumbItem{
		{Name: "Home", Path: "/"},
		{Name: "Eventos", Path: "/eventos"},
		{Name: event.Title, Path: "/evento/" + event.Slug},
	}))
}

func (h *Handler) classifiedJSONLD(classified *model.Classified, seo model.SEOData) template.JS {
	settings := h.portalSettings(context.Background())
	payload := map[string]any{
		"@type":       "Product",
		"name":        classified.Title,
		"description": firstNonEmpty(classified.MetaDescription, classified.Description),
		"url":         seo.URL,
		"category":    classified.Category,
		"areaServed": map[string]any{
			"@type": "City",
			"name":  settings.City,
		},
		"offers": map[string]any{
			"@type":         "Offer",
			"availability":  "https://schema.org/InStock",
			"priceCurrency": "BRL",
			"seller": map[string]any{
				"@type": "Organization",
				"name":  firstNonEmpty(classified.ContactName, settings.SiteName),
			},
		},
	}
	if seo.Image != "" {
		payload["image"] = []string{seo.Image}
	}
	if classified.PriceDisplay != "" {
		payload["offers"].(map[string]any)["priceSpecification"] = classified.PriceDisplay
	}
	return graphJSONScript(h.organizationPayload(seo), payload, h.breadcrumbPayload([]breadcrumbItem{
		{Name: "Home", Path: "/"},
		{Name: "Classificados", Path: "/classificados"},
		{Name: classified.Title, Path: "/classificado/" + classified.Slug},
	}))
}

func (h *Handler) organizationPayload(seo model.SEOData) map[string]any {
	settings := h.portalSettings(context.Background())
	branding := h.branding(context.Background())
	logoURL := branding.AbsoluteURL(branding.LogoPath)
	if settings.LogoKey != "" {
		logoURL = h.storage.URL(context.Background(), settings.LogoKey)
	}
	payload := map[string]any{
		"@type": "NewsMediaOrganization",
		"name":  branding.PortalName,
		"url":   h.siteURL(context.Background()) + "/",
		"logo": map[string]any{
			"@type": "ImageObject",
			"url":   logoURL,
		},
		"areaServed": map[string]any{
			"@type": "City",
			"name":  settings.City,
		},
		"address": map[string]any{
			"@type":           "PostalAddress",
			"addressLocality": settings.City,
			"addressRegion":   settings.State,
			"addressCountry":  "BR",
		},
		"sameAs": []string{},
	}
	sameAs := socialLinks(settings)
	if len(sameAs) > 0 {
		payload["sameAs"] = sameAs
	}
	if seo.Description != "" {
		payload["description"] = seo.Description
	}
	return payload
}

func (h *Handler) articleJSONLD(post *model.Post, seo model.SEOData) template.JS {
	branding := h.branding(context.Background())
	logoURL := branding.AbsoluteURL(branding.LogoPath)
	settings := h.portalSettings(context.Background())
	if settings.LogoKey != "" {
		logoURL = h.storage.URL(context.Background(), settings.LogoKey)
	}
	payload := map[string]any{
		"@type":            "NewsArticle",
		"headline":         post.Title,
		"description":      firstNonEmpty(post.MetaDescription, post.Excerpt),
		"mainEntityOfPage": seo.URL,
		"url":              seo.URL,
		"inLanguage":       branding.PortalLanguage,
		"articleSection":   post.CategoryName,
		"publisher": map[string]any{
			"@type": "NewsMediaOrganization",
			"name":  branding.PortalName,
			"url":   h.siteURL(context.Background()) + "/",
			"logo": map[string]any{
				"@type": "ImageObject",
				"url":   logoURL,
			},
		},
	}
	if post.PublishedAt != nil {
		payload["datePublished"] = post.PublishedAt.Format(time.RFC3339)
	}
	payload["dateModified"] = post.UpdatedAt.Format(time.RFC3339)
	if seo.Author != "" {
		payload["author"] = map[string]any{
			"@type": "Person",
			"name":  seo.Author,
		}
	}
	if seo.Image != "" {
		payload["image"] = []string{seo.Image}
	}
	if post.ReadingTimeMinutes > 0 {
		payload["timeRequired"] = fmt.Sprintf("PT%dM", post.ReadingTimeMinutes)
	}
	if post.SourceURL != "" {
		payload["citation"] = post.SourceURL
	}
	if post.IsSponsored {
		payload["isAccessibleForFree"] = true
		payload["sponsor"] = map[string]any{
			"@type": "Organization",
			"name":  "Conteudo patrocinado",
		}
	}
	return graphJSONScript(payload, h.breadcrumbPayload([]breadcrumbItem{
		{Name: "Home", Path: "/"},
		{Name: "Noticias", Path: "/noticias"},
		{Name: post.Title, Path: "/noticia/" + post.Slug},
	}))
}

type breadcrumbItem struct {
	Name string
	Path string
}

func (h *Handler) breadcrumbPayload(items []breadcrumbItem) map[string]any {
	elements := make([]map[string]any, 0, len(items))
	for i, item := range items {
		elements = append(elements, map[string]any{
			"@type":    "ListItem",
			"position": i + 1,
			"name":     item.Name,
			"item":     h.siteURL(context.Background()) + item.Path,
		})
	}
	return map[string]any{
		"@type":           "BreadcrumbList",
		"itemListElement": elements,
	}
}

func graphJSONScript(payloads ...map[string]any) template.JS {
	graph := make([]map[string]any, 0, len(payloads))
	for _, payload := range payloads {
		if payload != nil {
			graph = append(graph, payload)
		}
	}
	return mustJSONScript(map[string]any{
		"@context": "https://schema.org",
		"@graph":   graph,
	})
}

func mustJSONScript(payload any) template.JS {
	b, err := json.Marshal(payload)
	if err != nil {
		return "{}"
	}
	return template.JS(b)
}

type rssFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	AtomNS  string     `xml:"xmlns:atom,attr"`
	Channel rssChannel `xml:"channel"`
}

type rssChannel struct {
	Title       string      `xml:"title"`
	Link        string      `xml:"link"`
	Description string      `xml:"description"`
	Language    string      `xml:"language"`
	LastBuild   string      `xml:"lastBuildDate"`
	AtomLink    rssAtomLink `xml:"atom:link"`
	Items       []rssItem   `xml:"item"`
}

type rssAtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

type rssItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	GUID        string `xml:"guid"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
	Category    string `xml:"category,omitempty"`
}

// Static pages
func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	seo := model.SEOData{Title: h.pageTitle(r.Context(), "Sobre"), Description: "Sobre o portal " + h.branding(r.Context()).PortalName, URL: h.siteURL(r.Context()) + "/sobre"}
	h.Render(w, r, "about.html", map[string]any{"SEO": seo})
}

func (h *Handler) Contact(w http.ResponseWriter, r *http.Request) {
	branding := h.branding(r.Context())
	settings := h.portalSettings(r.Context())
	seo := model.SEOData{Title: h.pageTitle(r.Context(), "Contato"), Description: "Entre em contato com " + firstNonEmpty(settings.SiteName, branding.PortalName), URL: h.siteURL(r.Context()) + "/contato", NoIndex: true}
	h.Render(w, r, "contact.html", map[string]any{"SEO": seo})
}

// Admin Dashboard
func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	ctx := r.Context()
	period := adminPeriodFromRequest(r)
	allTenants := h.authSvc.HasPermission(user, auth.PermTenantsManage)
	selectedTenantID := int64(0)
	if allTenants {
		selectedTenantID = selectedTenantIDFromRequest(r)
	}

	summary, _ := h.repo.DashboardSummary(ctx, allTenants, selectedTenantID, period.From, period.To)
	topPosts, _ := h.repo.TopPostMetrics(ctx, allTenants, selectedTenantID, period.From, period.To, 6)
	latestPosts, _ := h.repo.LatestPublishedPostMetrics(ctx, allTenants, selectedTenantID, 6)
	tenantSummaries, _ := h.repo.TenantMetricSummaries(ctx, allTenants, selectedTenantID, 6)
	deadJobCount, _ := h.repo.DeadJobCount(ctx)
	prevFrom, prevTo := previousPeriod(period.From, period.To)
	previousViews, _ := h.repo.MetricCountFiltered(ctx, "post_view", allTenants, selectedTenantID, prevFrom, prevTo)
	tenants, _ := h.repo.TenantList(ctx)

	h.Render(w, r, "admin_dashboard.html", map[string]any{
		"Title":            "Dashboard",
		"Summary":          summary,
		"Period":           period,
		"SelectedTenantID": selectedTenantID,
		"Tenants":          tenants,
		"TopPosts":         topPosts,
		"HasTopPostViews":  !noPostViews(topPosts),
		"LatestPosts":      latestPosts,
		"TenantSummaries":  tenantSummaries,
		"ViewTrend":        trendLabel(summary.ViewsSelectedRange, previousViews),
		"DeadJobCount":     deadJobCount,
		"Alerts":           dashboardAlerts(summary, deadJobCount, topPosts),
		"Active":           "dashboard",
	})
}

func (h *Handler) requireCurrentUser(w http.ResponseWriter, r *http.Request) (*model.User, bool) {
	user := auth.UserFromContext(r.Context())
	if user == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return nil, false
	}
	return user, true
}

func (h *Handler) requirePermission(w http.ResponseWriter, r *http.Request, perm auth.Permission) (*model.User, bool) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return nil, false
	}
	if !h.authSvc.HasPermission(user, perm) {
		http.Error(w, "Acesso negado", http.StatusForbidden)
		return nil, false
	}
	if feature := featureForPermission(perm); feature != "" && !h.tenantFeatureEnabled(r, feature, true) {
		http.Error(w, "Feature desabilitada para este portal", http.StatusForbidden)
		return nil, false
	}
	return user, true
}

func (h *Handler) tenantFeatureEnabled(r *http.Request, feature string, fallback bool) bool {
	enabled, err := h.repo.TenantFeatureEnabledOrDefault(r.Context(), feature, fallback)
	return err == nil && enabled
}

func featureForPermission(perm auth.Permission) string {
	switch perm {
	case auth.PermMediaManage:
		return "media"
	case auth.PermStoresManage, auth.PermInfluencersManage, auth.PermBannersManage, auth.PermPromosManage, auth.PermEventsManage, auth.PermClassifiedsManage:
		return "commercial"
	default:
		return ""
	}
}

func (h *Handler) auditAdminAction(r *http.Request, user *model.User, action, entityType string, entityID *int64, changes map[string]any) {
	var userID *int64
	if user != nil {
		id := user.ID
		userID = &id
	}
	changesJSON := "{}"
	if len(changes) > 0 {
		if b, err := json.Marshal(changes); err == nil {
			changesJSON = string(b)
		}
	}
	_ = h.repo.AuditLogCreate(r.Context(), &model.AuditLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Changes:    changesJSON,
		IPAddress:  middleware.ClientIP(r),
		UserAgent:  r.UserAgent(),
	})
}

func auditEntityID(id int64) *int64 {
	return &id
}

func postTitle(post *model.Post) string {
	if post == nil {
		return ""
	}
	return post.Title
}

func storeName(store *model.Store) string {
	if store == nil {
		return ""
	}
	return store.Name
}

func storeCommercialStatusLabel(status string) string {
	switch status {
	case "lead":
		return "Prospect"
	case "paused":
		return "Pausado"
	case "inactive":
		return "Inativo comercial"
	default:
		return "Cliente ativo"
	}
}

func influencerName(influencer *model.Influencer) string {
	if influencer == nil {
		return ""
	}
	return influencer.Name
}

func bannerName(banner *model.Banner) string {
	if banner == nil {
		return ""
	}
	return banner.Name
}

func bannerPosition(banner *model.Banner) string {
	if banner == nil {
		return ""
	}
	return banner.Position
}

func promotionTitle(promo *model.Promotion) string {
	if promo == nil {
		return ""
	}
	return promo.Title
}

func promoStatusLabel(status string) string {
	switch status {
	case "draft":
		return "Rascunho"
	case "expired":
		return "Expirada"
	default:
		return "Ativa"
	}
}

func promoValidityLabel(promo model.Promotion) string {
	today := dateOnly(time.Now())
	start := dateOnly(promo.StartDate)
	end := dateOnly(promo.EndDate)
	if promo.Status == "draft" {
		return "Rascunho"
	}
	if promo.Status == "expired" || end.Before(today) {
		return "Expirada"
	}
	if start.After(today) {
		return "Agendada"
	}
	return "No ar"
}

func promotionPubliclyVisible(promo *model.Promotion) bool {
	if promo == nil || promo.Status != "active" {
		return false
	}
	today := dateOnly(time.Now())
	return !dateOnly(promo.StartDate).After(today) && !dateOnly(promo.EndDate).Before(today)
}

func eventTitle(event *model.Event) string {
	if event == nil {
		return ""
	}
	return event.Title
}

func classifiedTitle(classified *model.Classified) string {
	if classified == nil {
		return ""
	}
	return classified.Title
}

func classifiedExpired(classified *model.Classified) bool {
	if classified == nil || classified.ExpiresAt == nil {
		return false
	}
	return classified.ExpiresAt.Before(time.Now().AddDate(0, 0, -1))
}

func classifiedStatusLabel(status string) string {
	switch status {
	case "draft":
		return "Rascunho"
	case "archived":
		return "Arquivado"
	case "sold":
		return "Vendido"
	default:
		return "Ativo"
	}
}

func eventStatusLabel(status string) string {
	switch status {
	case "draft":
		return "Rascunho"
	case "archived":
		return "Arquivado"
	default:
		return "Ativo"
	}
}

func postStatusLabel(status model.PostStatus) string {
	switch status {
	case model.StatusPublished:
		return "Publicado"
	case model.StatusReview:
		return "Em revisao"
	case model.StatusApproved:
		return "Aprovado"
	case model.StatusScheduled:
		return "Agendado"
	case model.StatusArchived:
		return "Arquivado"
	default:
		return "Rascunho"
	}
}

func postSEOTags(post *model.Post) []string {
	tags := []string{"Inhumas", "Inhumas GO", post.CategoryName}
	if strings.TrimSpace(post.SEOKeyword) != "" {
		tags = append(tags, strings.TrimSpace(post.SEOKeyword))
	}
	for _, tag := range post.Tags {
		if tag.Active && strings.TrimSpace(tag.Name) != "" {
			tags = append(tags, strings.TrimSpace(tag.Name))
		}
	}
	return tags
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func adminPostFilterPath(status string, categoryID int64, tagID int64, query string) string {
	values := make(url.Values)
	if status != "" {
		values.Set("status", status)
	}
	if categoryID > 0 {
		values.Set("category_id", strconv.FormatInt(categoryID, 10))
	}
	if tagID > 0 {
		values.Set("tag_id", strconv.FormatInt(tagID, 10))
	}
	if query != "" {
		values.Set("q", query)
	}
	encoded := values.Encode()
	if encoded == "" {
		return ""
	}
	return "&" + encoded
}

func normalizeBannerStatus(status string) string {
	return commercialsvc.NormalizeBannerStatus(status)
}

func bannerStatusLabel(status string) string {
	labels := map[string]string{
		"active":  "Ativo",
		"paused":  "Pausado",
		"draft":   "Rascunho",
		"expired": "Expirado",
	}
	status = normalizeBannerStatus(status)
	if label, ok := labels[status]; ok {
		return label
	}
	return "Ativo"
}

func bannerPositionLabel(position string) string {
	labels := map[string]string{
		"hero":           "Topo da home",
		"in_feed":        "In-feed",
		"sidebar_top":    "Sidebar topo",
		"sidebar_bottom": "Sidebar rodape",
		"sticky_footer":  "Rodape mobile",
	}
	if label, ok := labels[position]; ok {
		return label
	}
	return position
}

func bannerDaysLeft(endDate time.Time) int {
	today := dateOnly(time.Now())
	end := dateOnly(endDate)
	return int(end.Sub(today).Hours() / 24)
}

func bannerCTR(banner model.Banner) string {
	return formatCTR(banner.ClickCount, banner.ImpressionCount)
}

func formatCTR(clicks, impressions int) string {
	if impressions <= 0 {
		return "0.00%"
	}
	return fmt.Sprintf("%.2f%%", (float64(clicks)/float64(impressions))*100)
}

func dateOnly(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, time.Local)
}
