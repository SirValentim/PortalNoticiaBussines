package model

import (
	"time"
)

type UserRole string

const (
	RoleSuperAdmin UserRole = "super_admin"
	RoleAdmin      UserRole = "admin"
	RoleEditor     UserRole = "editor"
	RoleRedator    UserRole = "redator"
	RoleRevisor    UserRole = "revisor"
	RoleComercial  UserRole = "comercial"
)

func (r UserRole) IsValid() bool {
	switch r {
	case RoleSuperAdmin, RoleAdmin, RoleEditor, RoleRedator, RoleRevisor, RoleComercial:
		return true
	default:
		return false
	}
}

func (r UserRole) Label() string {
	switch r {
	case RoleSuperAdmin:
		return "Super Admin"
	case RoleAdmin:
		return "Admin"
	case RoleEditor:
		return "Editor"
	case RoleRedator:
		return "Redator"
	case RoleRevisor:
		return "Revisor"
	case RoleComercial:
		return "Comercial"
	default:
		return string(r)
	}
}

type PostStatus string

const (
	StatusDraft     PostStatus = "draft"
	StatusReview    PostStatus = "review"
	StatusApproved  PostStatus = "approved"
	StatusScheduled PostStatus = "scheduled"
	StatusPublished PostStatus = "published"
	StatusArchived  PostStatus = "archived"
)

type JobStatus string

const (
	JobPending   JobStatus = "pending"
	JobRunning   JobStatus = "running"
	JobCompleted JobStatus = "completed"
	JobFailed    JobStatus = "failed"
)

type JobType string

const (
	JobPublishPost        JobType = "publish_post"
	JobExpirePromotion    JobType = "expire_promotion"
	JobExpireBanner       JobType = "expire_banner"
	JobBackupDatabase     JobType = "backup_database"
	JobVacuumDB           JobType = "vacuum_db"
	JobGenerateSitemap    JobType = "generate_sitemap"
	JobCleanupOldJobs     JobType = "cleanup_old_jobs"
	JobCompressOldUploads JobType = "compress_old_uploads"
	JobCollectNews        JobType = "collect_news"
)

type AutomationSourceType string

const (
	AutomationSourceRSS      AutomationSourceType = "rss"
	AutomationSourceOfficial AutomationSourceType = "official"
)

type AutomationRunStatus string

const (
	AutomationRunSuccess AutomationRunStatus = "success"
	AutomationRunError   AutomationRunStatus = "error"
	AutomationRunPartial AutomationRunStatus = "partial"
)

type User struct {
	TenantID     int64     `db:"tenant_id" json:"tenant_id"`
	ID           int64     `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         UserRole  `db:"role" json:"role"`
	Active       bool      `db:"active" json:"active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
}

type Tenant struct {
	ID            int64     `db:"id" json:"id"`
	Name          string    `db:"name" json:"name"`
	Slug          string    `db:"slug" json:"slug"`
	Status        string    `db:"status" json:"status"`
	PrimaryDomain string    `db:"primary_domain" json:"primary_domain"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
	UpdatedAt     time.Time `db:"updated_at" json:"updated_at"`
}

type TenantDomain struct {
	ID        int64     `db:"id" json:"id"`
	TenantID  int64     `db:"tenant_id" json:"tenant_id"`
	Domain    string    `db:"domain" json:"domain"`
	IsPrimary bool      `db:"is_primary" json:"is_primary"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type TenantFeature struct {
	ID        int64     `db:"id" json:"id"`
	TenantID  int64     `db:"tenant_id" json:"tenant_id"`
	Feature   string    `db:"feature" json:"feature"`
	Enabled   bool      `db:"enabled" json:"enabled"`
	Limit     *int64    `db:"limit_value" json:"limit,omitempty"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type TenantUser struct {
	ID        int64     `db:"id" json:"id"`
	TenantID  int64     `db:"tenant_id" json:"tenant_id"`
	UserID    int64     `db:"user_id" json:"user_id"`
	Role      UserRole  `db:"role" json:"role"`
	Active    bool      `db:"active" json:"active"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
	UserName  string    `json:"user_name,omitempty"`
	UserEmail string    `json:"user_email,omitempty"`
}

type PasswordResetToken struct {
	ID          int64      `db:"id" json:"id"`
	UserID      int64      `db:"user_id" json:"user_id"`
	TokenHash   string     `db:"token_hash" json:"-"`
	RequestedIP string     `db:"requested_ip" json:"requested_ip"`
	UserAgent   string     `db:"user_agent" json:"user_agent"`
	ExpiresAt   time.Time  `db:"expires_at" json:"expires_at"`
	UsedAt      *time.Time `db:"used_at" json:"used_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
}

type Category struct {
	TenantID               int64  `db:"tenant_id" json:"tenant_id"`
	ID                     int64  `db:"id" json:"id"`
	Slug                   string `db:"slug" json:"slug"`
	Name                   string `db:"name" json:"name"`
	Description            string `db:"description" json:"description"`
	MetaTitle              string `db:"meta_title" json:"meta_title"`
	MetaDescription        string `db:"meta_description" json:"meta_description"`
	ImageKey               string `db:"image_key" json:"image_key"`
	SortOrder              int    `db:"sort_order" json:"sort_order"`
	Active                 bool   `db:"active" json:"active"`
	RequiresEditorialNotes bool   `db:"requires_editorial_notes" json:"requires_editorial_notes"`
}

type Tag struct {
	TenantID        int64     `db:"tenant_id" json:"tenant_id"`
	ID              int64     `db:"id" json:"id"`
	Slug            string    `db:"slug" json:"slug"`
	Name            string    `db:"name" json:"name"`
	Description     string    `db:"description" json:"description"`
	MetaTitle       string    `db:"meta_title" json:"meta_title"`
	MetaDescription string    `db:"meta_description" json:"meta_description"`
	Active          bool      `db:"active" json:"active"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type Post struct {
	TenantID           int64      `db:"tenant_id" json:"tenant_id"`
	ID                 int64      `db:"id" json:"id"`
	Title              string     `db:"title" json:"title"`
	Slug               string     `db:"slug" json:"slug"`
	Excerpt            string     `db:"excerpt" json:"excerpt"`
	Content            string     `db:"content" json:"content"`
	CoverImageKey      string     `db:"cover_image_key" json:"cover_image_key"`
	GalleryImageKeys   []string   `json:"gallery_image_keys,omitempty"`
	MetaTitle          string     `db:"meta_title" json:"meta_title"`
	MetaDescription    string     `db:"meta_description" json:"meta_description"`
	SEOKeyword         string     `db:"seo_keyword" json:"seo_keyword"`
	CanonicalURL       string     `db:"canonical_url" json:"canonical_url"`
	SourceName         string     `db:"source_name" json:"source_name"`
	SourceURL          string     `db:"source_url" json:"source_url"`
	ReadingTimeMinutes int        `db:"reading_time_minutes" json:"reading_time_minutes"`
	CategoryID         *int64     `db:"category_id" json:"category_id"`
	AuthorID           *int64     `db:"author_id" json:"author_id"`
	Status             PostStatus `db:"status" json:"status"`
	IsSponsored        bool       `db:"is_sponsored" json:"is_sponsored"`
	IsFeatured         bool       `db:"is_featured" json:"is_featured"`
	IsPinned           bool       `db:"is_pinned" json:"is_pinned"`
	EditorialNotes     string     `db:"editorial_notes" json:"editorial_notes"`
	EditorResponsible  string     `db:"editor_responsible" json:"editor_responsible"`
	PublishedAt        *time.Time `db:"published_at" json:"published_at"`
	PublishAt          *time.Time `db:"publish_at" json:"publish_at"`
	CreatedAt          time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt          time.Time  `db:"updated_at" json:"updated_at"`
	CategoryName       string     `db:"category_name" json:"category_name,omitempty"`
	AuthorName         string     `db:"author_name" json:"author_name,omitempty"`
	Tags               []Tag      `json:"tags,omitempty"`
}

type PostRevision struct {
	ID        int64      `db:"id" json:"id"`
	PostID    int64      `db:"post_id" json:"post_id"`
	UserID    *int64     `db:"user_id" json:"user_id"`
	Action    string     `db:"action" json:"action"`
	Title     string     `db:"title" json:"title"`
	Status    PostStatus `db:"status" json:"status"`
	Snapshot  string     `db:"snapshot" json:"snapshot"`
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UserName  string     `json:"user_name,omitempty"`
}

type MediaAsset struct {
	TenantID       int64     `db:"tenant_id" json:"tenant_id"`
	ID             int64     `db:"id" json:"id"`
	Key            string    `db:"key" json:"key"`
	OriginalName   string    `db:"original_name" json:"original_name"`
	Title          string    `db:"title" json:"title"`
	AltText        string    `db:"alt_text" json:"alt_text"`
	ContentType    string    `db:"content_type" json:"content_type"`
	SizeBytes      int64     `db:"size_bytes" json:"size_bytes"`
	UploadedBy     *int64    `db:"uploaded_by" json:"uploaded_by"`
	UploadedByName string    `json:"uploaded_by_name,omitempty"`
	UsageCount     int       `json:"usage_count,omitempty"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

type MediaArchiveMonth struct {
	Month string `json:"month"`
	Label string `json:"label"`
	Count int    `json:"count"`
}

type PortalSettings struct {
	ID                        int64     `db:"id" json:"id"`
	TenantID                  int64     `db:"tenant_id" json:"tenant_id"`
	SiteName                  string    `db:"site_name" json:"site_name"`
	Tagline                   string    `db:"tagline" json:"tagline"`
	LogoKey                   string    `db:"logo_key" json:"logo_key"`
	FaviconKey                string    `db:"favicon_key" json:"favicon_key"`
	ContactEmail              string    `db:"contact_email" json:"contact_email"`
	ContactWhatsapp           string    `db:"contact_whatsapp" json:"contact_whatsapp"`
	ContactPhone              string    `db:"contact_phone" json:"contact_phone"`
	City                      string    `db:"city" json:"city"`
	State                     string    `db:"state" json:"state"`
	SEOTitle                  string    `db:"seo_title" json:"seo_title"`
	SEODescription            string    `db:"seo_description" json:"seo_description"`
	FacebookURL               string    `db:"facebook_url" json:"facebook_url"`
	InstagramURL              string    `db:"instagram_url" json:"instagram_url"`
	YoutubeURL                string    `db:"youtube_url" json:"youtube_url"`
	TiktokURL                 string    `db:"tiktok_url" json:"tiktok_url"`
	UploadMaxMB               int       `db:"upload_max_mb" json:"upload_max_mb"`
	AutomationEnabled         bool      `db:"automation_enabled" json:"automation_enabled"`
	AutomationIntervalMinutes int       `db:"automation_interval_minutes" json:"automation_interval_minutes"`
	UpdatedAt                 time.Time `db:"updated_at" json:"updated_at"`
}

type AutomationSource struct {
	TenantID          int64      `db:"tenant_id" json:"tenant_id"`
	ID                int64      `db:"id" json:"id"`
	Name              string     `db:"name" json:"name"`
	SourceType        string     `db:"source_type" json:"source_type"`
	URL               string     `db:"url" json:"url"`
	DefaultCategoryID *int64     `db:"default_category_id" json:"default_category_id"`
	Active            bool       `db:"active" json:"active"`
	LastRunAt         *time.Time `db:"last_run_at" json:"last_run_at"`
	CreatedAt         time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt         time.Time  `db:"updated_at" json:"updated_at"`
	CategoryName      string     `json:"category_name,omitempty"`
}

type AutomationRun struct {
	TenantID      int64      `db:"tenant_id" json:"tenant_id"`
	ID            int64      `db:"id" json:"id"`
	SourceID      *int64     `db:"source_id" json:"source_id"`
	SourceName    string     `json:"source_name,omitempty"`
	Status        string     `db:"status" json:"status"`
	ItemsFound    int        `db:"items_found" json:"items_found"`
	DraftsCreated int        `db:"drafts_created" json:"drafts_created"`
	Duplicates    int        `db:"duplicates" json:"duplicates"`
	Error         string     `db:"error" json:"error"`
	Log           string     `db:"log" json:"log"`
	StartedAt     time.Time  `db:"started_at" json:"started_at"`
	FinishedAt    *time.Time `db:"finished_at" json:"finished_at"`
}

type AIUsageLog struct {
	ID         int64     `db:"id" json:"id"`
	PostID     *int64    `db:"post_id" json:"post_id"`
	UserID     *int64    `db:"user_id" json:"user_id"`
	Action     string    `db:"action" json:"action"`
	Provider   string    `db:"provider" json:"provider"`
	InputTitle string    `db:"input_title" json:"input_title"`
	Output     string    `db:"output" json:"output"`
	SourceName string    `db:"source_name" json:"source_name"`
	SourceURL  string    `db:"source_url" json:"source_url"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
	UserName   string    `json:"user_name,omitempty"`
	UserEmail  string    `json:"user_email,omitempty"`
}

type SlugRedirect struct {
	OldSlug    string    `db:"old_slug" json:"old_slug"`
	NewSlug    string    `db:"new_slug" json:"new_slug"`
	EntityType string    `db:"entity_type" json:"entity_type"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type Store struct {
	TenantID         int64     `db:"tenant_id" json:"tenant_id"`
	ID               int64     `db:"id" json:"id"`
	Slug             string    `db:"slug" json:"slug"`
	Name             string    `db:"name" json:"name"`
	Description      string    `db:"description" json:"description"`
	Category         string    `db:"category" json:"category"`
	Address          string    `db:"address" json:"address"`
	Phone            string    `db:"phone" json:"phone"`
	Whatsapp         string    `db:"whatsapp" json:"whatsapp"`
	WebsiteURL       string    `db:"website_url" json:"website_url"`
	LogoKey          string    `db:"logo_key" json:"logo_key"`
	CoverImageKey    string    `db:"cover_image_key" json:"cover_image_key"`
	CommercialStatus string    `db:"commercial_status" json:"commercial_status"`
	MetaTitle        string    `db:"meta_title" json:"meta_title"`
	MetaDescription  string    `db:"meta_description" json:"meta_description"`
	IsSponsored      bool      `db:"is_sponsored" json:"is_sponsored"`
	IsFeatured       bool      `db:"is_featured" json:"is_featured"`
	NeighborhoodID   *int64    `db:"neighborhood_id" json:"neighborhood_id"`
	Active           bool      `db:"active" json:"active"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

type Promotion struct {
	TenantID        int64     `db:"tenant_id" json:"tenant_id"`
	ID              int64     `db:"id" json:"id"`
	StoreID         int64     `db:"store_id" json:"store_id"`
	Title           string    `db:"title" json:"title"`
	Slug            string    `db:"slug" json:"slug"`
	Description     string    `db:"description" json:"description"`
	PriceDisplay    string    `db:"price_display" json:"price_display"`
	CouponCode      string    `db:"coupon_code" json:"coupon_code"`
	ImageKey        string    `db:"image_key" json:"image_key"`
	StartDate       time.Time `db:"start_date" json:"start_date"`
	EndDate         time.Time `db:"end_date" json:"end_date"`
	Status          string    `db:"status" json:"status"`
	IsSponsored     bool      `db:"is_sponsored" json:"is_sponsored"`
	MetaTitle       string    `db:"meta_title" json:"meta_title"`
	MetaDescription string    `db:"meta_description" json:"meta_description"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
	StoreName       string    `db:"store_name" json:"store_name,omitempty"`
	StoreSlug       string    `db:"store_slug" json:"store_slug,omitempty"`
	ClickCount      int       `json:"click_count,omitempty"`
}

type Event struct {
	TenantID        int64      `db:"tenant_id" json:"tenant_id"`
	ID              int64      `db:"id" json:"id"`
	Slug            string     `db:"slug" json:"slug"`
	Title           string     `db:"title" json:"title"`
	Description     string     `db:"description" json:"description"`
	Location        string     `db:"location" json:"location"`
	Organizer       string     `db:"organizer" json:"organizer"`
	TicketURL       string     `db:"ticket_url" json:"ticket_url"`
	PriceDisplay    string     `db:"price_display" json:"price_display"`
	ImageKey        string     `db:"image_key" json:"image_key"`
	Status          string     `db:"status" json:"status"`
	IsFeatured      bool       `db:"is_featured" json:"is_featured"`
	IsSponsored     bool       `db:"is_sponsored" json:"is_sponsored"`
	MetaTitle       string     `db:"meta_title" json:"meta_title"`
	MetaDescription string     `db:"meta_description" json:"meta_description"`
	StartAt         time.Time  `db:"start_at" json:"start_at"`
	EndAt           *time.Time `db:"end_at" json:"end_at"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

type Classified struct {
	TenantID        int64      `db:"tenant_id" json:"tenant_id"`
	ID              int64      `db:"id" json:"id"`
	Slug            string     `db:"slug" json:"slug"`
	Title           string     `db:"title" json:"title"`
	Description     string     `db:"description" json:"description"`
	Category        string     `db:"category" json:"category"`
	PriceDisplay    string     `db:"price_display" json:"price_display"`
	ContactName     string     `db:"contact_name" json:"contact_name"`
	ContactPhone    string     `db:"contact_phone" json:"contact_phone"`
	ContactWhatsapp string     `db:"contact_whatsapp" json:"contact_whatsapp"`
	Location        string     `db:"location" json:"location"`
	ImageKey        string     `db:"image_key" json:"image_key"`
	Status          string     `db:"status" json:"status"`
	IsFeatured      bool       `db:"is_featured" json:"is_featured"`
	IsSponsored     bool       `db:"is_sponsored" json:"is_sponsored"`
	MetaTitle       string     `db:"meta_title" json:"meta_title"`
	MetaDescription string     `db:"meta_description" json:"meta_description"`
	ExpiresAt       *time.Time `db:"expires_at" json:"expires_at"`
	CreatedAt       time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

type Banner struct {
	TenantID        int64     `db:"tenant_id" json:"tenant_id"`
	ID              int64     `db:"id" json:"id"`
	Name            string    `db:"name" json:"name"`
	AdvertiserName  string    `db:"advertiser_name" json:"advertiser_name"`
	ContactName     string    `db:"contact_name" json:"contact_name"`
	ContactPhone    string    `db:"contact_phone" json:"contact_phone"`
	ContactWhatsapp string    `db:"contact_whatsapp" json:"contact_whatsapp"`
	PriceDisplay    string    `db:"price_display" json:"price_display"`
	Notes           string    `db:"notes" json:"notes"`
	Position        string    `db:"position" json:"position"`
	ImageKey        string    `db:"image_key" json:"image_key"`
	LinkURL         string    `db:"link_url" json:"link_url"`
	StartDate       time.Time `db:"start_date" json:"start_date"`
	EndDate         time.Time `db:"end_date" json:"end_date"`
	Status          string    `db:"status" json:"status"`
	Active          bool      `db:"active" json:"active"`
	Priority        int       `db:"priority" json:"priority"`
	ImpressionCount int       `db:"-" json:"impression_count"`
	ClickCount      int       `db:"-" json:"click_count"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type Neighborhood struct {
	TenantID        int64     `db:"tenant_id" json:"tenant_id"`
	ID              int64     `db:"id" json:"id"`
	Slug            string    `db:"slug" json:"slug"`
	Name            string    `db:"name" json:"name"`
	Description     string    `db:"description" json:"description"`
	MetaTitle       string    `db:"meta_title" json:"meta_title"`
	MetaDescription string    `db:"meta_description" json:"meta_description"`
	CoverImageKey   string    `db:"cover_image_key" json:"cover_image_key"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type Influencer struct {
	TenantID        int64     `db:"tenant_id" json:"tenant_id"`
	ID              int64     `db:"id" json:"id"`
	Slug            string    `db:"slug" json:"slug"`
	Name            string    `db:"name" json:"name"`
	Bio             string    `db:"bio" json:"bio"`
	CityArea        string    `db:"city_area" json:"city_area"`
	Niche           string    `db:"niche" json:"niche"`
	Instagram       string    `db:"instagram" json:"instagram"`
	TikTok          string    `db:"tiktok" json:"tiktok"`
	YouTube         string    `db:"youtube" json:"youtube"`
	Whatsapp        string    `db:"whatsapp" json:"whatsapp"`
	AvatarKey       string    `db:"avatar_key" json:"avatar_key"`
	CoverImageKey   string    `db:"cover_image_key" json:"cover_image_key"`
	MetaTitle       string    `db:"meta_title" json:"meta_title"`
	MetaDescription string    `db:"meta_description" json:"meta_description"`
	IsFeatured      bool      `db:"is_featured" json:"is_featured"`
	IsSponsored     bool      `db:"is_sponsored" json:"is_sponsored"`
	Active          bool      `db:"active" json:"active"`
	ViewCount       int       `db:"-" json:"view_count"`
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type Job struct {
	TenantID    int64      `db:"tenant_id" json:"tenant_id"`
	ID          int64      `db:"id" json:"id"`
	Type        JobType    `db:"type" json:"type"`
	Payload     string     `db:"payload" json:"payload"`
	Status      JobStatus  `db:"status" json:"status"`
	RunAt       time.Time  `db:"run_at" json:"run_at"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	Attempts    int        `db:"attempts" json:"attempts"`
	MaxAttempts int        `db:"max_attempts" json:"max_attempts"`
	Error       string     `db:"error" json:"error"`
	ProcessedAt *time.Time `db:"processed_at" json:"processed_at"`
}

type Metric struct {
	TenantID   int64     `db:"tenant_id" json:"tenant_id"`
	ID         int64     `db:"id" json:"id"`
	MetricType string    `db:"metric_type" json:"metric_type"`
	EntityType string    `db:"entity_type" json:"entity_type"`
	EntityID   int64     `db:"entity_id" json:"entity_id"`
	UserID     *int64    `db:"user_id" json:"user_id"`
	IPAddress  string    `db:"ip_address" json:"ip_address"`
	UserAgent  string    `db:"user_agent" json:"user_agent"`
	Referrer   string    `db:"referrer" json:"referrer"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type AuditLog struct {
	ID         int64     `db:"id" json:"id"`
	UserID     *int64    `db:"user_id" json:"user_id"`
	Action     string    `db:"action" json:"action"`
	EntityType string    `db:"entity_type" json:"entity_type"`
	EntityID   *int64    `db:"entity_id" json:"entity_id"`
	Changes    string    `db:"changes" json:"changes"`
	IPAddress  string    `db:"ip_address" json:"ip_address"`
	UserAgent  string    `db:"user_agent" json:"user_agent"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type AuditLogEntry struct {
	AuditLog
	UserName  string `json:"user_name"`
	UserEmail string `json:"user_email"`
}

type EditLock struct {
	EntityType string    `db:"entity_type" json:"entity_type"`
	EntityID   int64     `db:"entity_id" json:"entity_id"`
	UserID     int64     `db:"user_id" json:"user_id"`
	LockedAt   time.Time `db:"locked_at" json:"locked_at"`
	ExpiresAt  time.Time `db:"expires_at" json:"expires_at"`
}

type LoginAttempt struct {
	ID        int64     `db:"id" json:"id"`
	IPAddress string    `db:"ip_address" json:"ip_address"`
	Email     string    `db:"email" json:"email"`
	Success   bool      `db:"success" json:"success"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}

type HomeData struct {
	Headline       *Post
	LatestNews     []Post
	Bastidores     []Post
	HeroBanner     *Banner
	InFeedBanner   *Banner
	StickyBanner   *Banner
	FeaturedStores []Store
	FeaturedPromos []Promotion
	SponsoredPosts []Post
	Neighborhoods  []Neighborhood
	Influencers    []Influencer
}

type SEOData struct {
	Title        string
	Description  string
	Image        string
	URL          string
	Type         string
	NoIndex      bool
	PublishedAt  *time.Time
	ModifiedAt   *time.Time
	Author       string
	Tags         []string
	CanonicalURL string
}

type MetricTotal struct {
	MetricType string
	Total      int
}

type MetricEntityTotal struct {
	MetricType string
	EntityType string
	EntityID   int64
	Total      int
}

type SitemapEntry struct {
	Path    string
	LastMod time.Time
}
