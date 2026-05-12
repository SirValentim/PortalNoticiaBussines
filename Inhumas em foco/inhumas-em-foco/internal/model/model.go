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
)

type User struct {
	ID           int64     `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Email        string    `db:"email" json:"email"`
	PasswordHash string    `db:"password_hash" json:"-"`
	Role         UserRole  `db:"role" json:"role"`
	Active       bool      `db:"active" json:"active"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
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

type SlugRedirect struct {
	OldSlug    string    `db:"old_slug" json:"old_slug"`
	NewSlug    string    `db:"new_slug" json:"new_slug"`
	EntityType string    `db:"entity_type" json:"entity_type"`
	CreatedAt  time.Time `db:"created_at" json:"created_at"`
}

type Store struct {
	ID             int64     `db:"id" json:"id"`
	Slug           string    `db:"slug" json:"slug"`
	Name           string    `db:"name" json:"name"`
	Description    string    `db:"description" json:"description"`
	Category       string    `db:"category" json:"category"`
	Address        string    `db:"address" json:"address"`
	Phone          string    `db:"phone" json:"phone"`
	Whatsapp       string    `db:"whatsapp" json:"whatsapp"`
	LogoKey        string    `db:"logo_key" json:"logo_key"`
	CoverImageKey  string    `db:"cover_image_key" json:"cover_image_key"`
	IsSponsored    bool      `db:"is_sponsored" json:"is_sponsored"`
	IsFeatured     bool      `db:"is_featured" json:"is_featured"`
	NeighborhoodID *int64    `db:"neighborhood_id" json:"neighborhood_id"`
	Active         bool      `db:"active" json:"active"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
}

type Promotion struct {
	ID           int64     `db:"id" json:"id"`
	StoreID      int64     `db:"store_id" json:"store_id"`
	Title        string    `db:"title" json:"title"`
	Slug         string    `db:"slug" json:"slug"`
	Description  string    `db:"description" json:"description"`
	PriceDisplay string    `db:"price_display" json:"price_display"`
	ImageKey     string    `db:"image_key" json:"image_key"`
	StartDate    time.Time `db:"start_date" json:"start_date"`
	EndDate      time.Time `db:"end_date" json:"end_date"`
	Status       string    `db:"status" json:"status"`
	IsSponsored  bool      `db:"is_sponsored" json:"is_sponsored"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	StoreName    string    `db:"store_name" json:"store_name,omitempty"`
	StoreSlug    string    `db:"store_slug" json:"store_slug,omitempty"`
}

type Banner struct {
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
	CreatedAt       time.Time `db:"created_at" json:"created_at"`
}

type Neighborhood struct {
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
	ID            int64     `db:"id" json:"id"`
	Slug          string    `db:"slug" json:"slug"`
	Name          string    `db:"name" json:"name"`
	Bio           string    `db:"bio" json:"bio"`
	CityArea      string    `db:"city_area" json:"city_area"`
	Instagram     string    `db:"instagram" json:"instagram"`
	TikTok        string    `db:"tiktok" json:"tiktok"`
	YouTube       string    `db:"youtube" json:"youtube"`
	Whatsapp      string    `db:"whatsapp" json:"whatsapp"`
	AvatarKey     string    `db:"avatar_key" json:"avatar_key"`
	CoverImageKey string    `db:"cover_image_key" json:"cover_image_key"`
	IsFeatured    bool      `db:"is_featured" json:"is_featured"`
	IsSponsored   bool      `db:"is_sponsored" json:"is_sponsored"`
	Active        bool      `db:"active" json:"active"`
	CreatedAt     time.Time `db:"created_at" json:"created_at"`
}

type Job struct {
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
