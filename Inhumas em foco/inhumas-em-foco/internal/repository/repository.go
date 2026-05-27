package repository

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"inhumas-em-foco/internal/model"

	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"
)

type Repository struct {
	db     *sql.DB
	driver string
}

type AuditLogFilter struct {
	UserID     int64
	Action     string
	EntityType string
	EntityID   int64
	DateFrom   *time.Time
	DateTo     *time.Time
	Limit      int
	Offset     int
}

type MediaAssetFilter struct {
	Query    string
	DateFrom *time.Time
	DateTo   *time.Time
	Limit    int
	Offset   int
}

func New(dbPath string) (*Repository, error) {
	return Open("sqlite", dbPath, "")
}

func Open(driver, databaseURL, migrationsDir string) (*Repository, error) {
	driver = normalizeDriver(driver, databaseURL)
	openURL := databaseURL
	if driver == "sqlite" {
		openURL = databaseURL + "?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=cache_size(-64000)&_pragma=temp_store(memory)"
	}
	db, err := sql.Open(driver, openURL)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)
	if driver == "sqlite" && isInMemorySQLite(databaseURL) {
		db.SetMaxOpenConns(1)
		db.SetMaxIdleConns(1)
	}

	repo := &Repository{db: db, driver: driver}
	if err := repo.MigrateWithDir(migrationsDir); err != nil {
		db.Close()
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return repo, nil
}

func isInMemorySQLite(databaseURL string) bool {
	normalized := strings.ToLower(strings.TrimSpace(databaseURL))
	return normalized == ":memory:" || strings.Contains(normalized, "mode=memory")
}

func normalizeDriver(driver, databaseURL string) string {
	driver = strings.ToLower(strings.TrimSpace(driver))
	switch driver {
	case "postgres", "postgresql", "pg":
		return "postgres"
	case "sqlite", "sqlite3":
		return "sqlite"
	}
	if strings.HasPrefix(strings.ToLower(databaseURL), "postgres://") || strings.HasPrefix(strings.ToLower(databaseURL), "postgresql://") {
		return "postgres"
	}
	return "sqlite"
}

func (r *Repository) DB() *sql.DB {
	return r.db
}

func (r *Repository) Driver() string {
	return r.driver
}

func (r *Repository) Migrate() error {
	return r.MigrateWithDir("")
}

func (r *Repository) MigrateWithDir(migrationsDir string) error {
	if r.driver == "postgres" {
		return r.runPostgresMigrations(migrationsDir)
	}
	schema := `
CREATE TABLE IF NOT EXISTS tenants (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    primary_domain TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tenant_domains (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    domain TEXT NOT NULL UNIQUE,
    is_primary BOOLEAN NOT NULL DEFAULT false,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_tenant_domains_tenant ON tenant_domains(tenant_id);
CREATE INDEX IF NOT EXISTS idx_tenants_status ON tenants(status);

CREATE TABLE IF NOT EXISTS tenant_features (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    feature TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    limit_value INTEGER,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, feature)
);
CREATE INDEX IF NOT EXISTS idx_tenant_features_tenant ON tenant_features(tenant_id, enabled);

CREATE TABLE IF NOT EXISTS tenant_users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'editor',
    active BOOLEAN NOT NULL DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_tenant_users_user ON tenant_users(user_id, active);
CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant_role ON tenant_users(tenant_id, role, active);

INSERT OR IGNORE INTO tenants (name, slug, status, primary_domain) VALUES ('Default Portal', 'default', 'active', NULL);

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    email TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'editor',
    active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, email)
);

CREATE TABLE IF NOT EXISTS password_reset_tokens (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER NOT NULL REFERENCES users(id),
    token_hash TEXT UNIQUE NOT NULL,
    requested_ip TEXT,
    user_agent TEXT,
    expires_at DATETIME NOT NULL,
    used_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_password_reset_hash ON password_reset_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_password_reset_user ON password_reset_tokens(user_id, created_at DESC);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    meta_title TEXT,
    meta_description TEXT,
    image_key TEXT,
    sort_order INTEGER DEFAULT 0,
    active BOOLEAN DEFAULT true,
    requires_editorial_notes BOOLEAN DEFAULT false,
    UNIQUE(tenant_id, slug)
);

CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    meta_title TEXT,
    meta_description TEXT,
    active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);

CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    title TEXT NOT NULL,
    slug TEXT NOT NULL,
    excerpt TEXT,
    content TEXT NOT NULL,
    cover_image_key TEXT,
    gallery_image_keys TEXT,
    meta_title TEXT,
    meta_description TEXT,
    seo_keyword TEXT,
    canonical_url TEXT,
    source_name TEXT,
    source_url TEXT,
    reading_time_minutes INTEGER DEFAULT 1,
    category_id INTEGER REFERENCES categories(id),
    author_id INTEGER REFERENCES users(id),
    status TEXT DEFAULT 'draft',
    is_sponsored BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    is_pinned BOOLEAN DEFAULT false,
    editorial_notes TEXT,
    editor_responsible TEXT,
    published_at DATETIME,
    publish_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);

CREATE TABLE IF NOT EXISTS post_revisions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id INTEGER REFERENCES users(id),
    action TEXT NOT NULL,
    title TEXT,
    status TEXT,
    snapshot TEXT NOT NULL DEFAULT '{}',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_post_revisions_post ON post_revisions(post_id, created_at DESC);

CREATE TABLE IF NOT EXISTS media_assets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    key TEXT NOT NULL,
    original_name TEXT NOT NULL,
    title TEXT,
    alt_text TEXT,
    content_type TEXT NOT NULL,
    size_bytes INTEGER DEFAULT 0,
    uploaded_by INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, key)
);
CREATE INDEX IF NOT EXISTS idx_media_assets_created ON media_assets(tenant_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_assets_search ON media_assets(tenant_id, original_name, title, alt_text);

CREATE TABLE IF NOT EXISTS portal_settings (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    site_name TEXT NOT NULL,
    tagline TEXT,
    logo_key TEXT,
    favicon_key TEXT,
    contact_email TEXT,
    contact_whatsapp TEXT,
    contact_phone TEXT,
    city TEXT,
    state TEXT,
    seo_title TEXT,
    seo_description TEXT,
    facebook_url TEXT,
    instagram_url TEXT,
    youtube_url TEXT,
    tiktok_url TEXT,
    upload_max_mb INTEGER DEFAULT 2,
    automation_enabled BOOLEAN DEFAULT false,
    automation_interval_minutes INTEGER DEFAULT 60,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id)
);

CREATE TABLE IF NOT EXISTS automation_sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    source_type TEXT NOT NULL DEFAULT 'rss',
    url TEXT NOT NULL,
    default_category_id INTEGER REFERENCES categories(id),
    active BOOLEAN DEFAULT true,
    last_run_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_automation_sources_active ON automation_sources(tenant_id, active, source_type);

CREATE TABLE IF NOT EXISTS automation_runs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    source_id INTEGER REFERENCES automation_sources(id),
    status TEXT NOT NULL DEFAULT 'success',
    items_found INTEGER DEFAULT 0,
    drafts_created INTEGER DEFAULT 0,
    duplicates INTEGER DEFAULT 0,
    error TEXT,
    log TEXT,
    started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    finished_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_automation_runs_started ON automation_runs(tenant_id, started_at DESC);

CREATE TABLE IF NOT EXISTS ai_usage_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    post_id INTEGER REFERENCES posts(id) ON DELETE SET NULL,
    user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
    action TEXT NOT NULL,
    provider TEXT NOT NULL,
    input_title TEXT,
    output TEXT NOT NULL,
    source_name TEXT,
    source_url TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_ai_usage_post ON ai_usage_logs(post_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ai_usage_created ON ai_usage_logs(created_at DESC);

CREATE TABLE IF NOT EXISTS post_tags (
    post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);
CREATE INDEX IF NOT EXISTS idx_post_tags_tag ON post_tags(tag_id, post_id);

CREATE VIRTUAL TABLE IF NOT EXISTS posts_fts USING fts5(title, excerpt, content, content_rowid=rowid, content=posts);

CREATE TRIGGER IF NOT EXISTS posts_fts_insert AFTER INSERT ON posts BEGIN
    INSERT INTO posts_fts(rowid, title, excerpt, content) VALUES (new.id, new.title, new.excerpt, new.content);
END;

CREATE TRIGGER IF NOT EXISTS posts_fts_delete AFTER DELETE ON posts BEGIN
    INSERT INTO posts_fts(posts_fts, rowid, title, excerpt, content) VALUES ('delete', old.id, old.title, old.excerpt, old.content);
END;

CREATE TRIGGER IF NOT EXISTS posts_fts_update AFTER UPDATE ON posts BEGIN
    INSERT INTO posts_fts(posts_fts, rowid, title, excerpt, content) VALUES ('delete', old.id, old.title, old.excerpt, old.content);
    INSERT INTO posts_fts(rowid, title, excerpt, content) VALUES (new.id, new.title, new.excerpt, new.content);
END;

CREATE TABLE IF NOT EXISTS slug_redirects (
    old_slug TEXT PRIMARY KEY,
    new_slug TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS stores (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    category TEXT,
    address TEXT,
    phone TEXT,
    whatsapp TEXT,
    website_url TEXT,
    logo_key TEXT,
    cover_image_key TEXT,
    commercial_status TEXT DEFAULT 'active',
    meta_title TEXT,
    meta_description TEXT,
    is_sponsored BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    neighborhood_id INTEGER,
    active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);

CREATE TABLE IF NOT EXISTS promotions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    store_id INTEGER REFERENCES stores(id),
    title TEXT NOT NULL,
    slug TEXT NOT NULL,
    description TEXT,
    price_display TEXT,
    coupon_code TEXT,
    image_key TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status TEXT DEFAULT 'active',
    is_sponsored BOOLEAN DEFAULT false,
    meta_title TEXT,
    meta_description TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);

CREATE TABLE IF NOT EXISTS events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    location TEXT,
    organizer TEXT,
    ticket_url TEXT,
    price_display TEXT,
    image_key TEXT,
    status TEXT DEFAULT 'active',
    is_featured BOOLEAN DEFAULT false,
    is_sponsored BOOLEAN DEFAULT false,
    meta_title TEXT,
    meta_description TEXT,
    start_at DATETIME NOT NULL,
    end_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);
CREATE INDEX IF NOT EXISTS idx_events_status_start ON events(status, start_at);
CREATE INDEX IF NOT EXISTS idx_events_featured ON events(is_featured, status, start_at);

CREATE TABLE IF NOT EXISTS classifieds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT,
    category TEXT,
    price_display TEXT,
    contact_name TEXT,
    contact_phone TEXT,
    contact_whatsapp TEXT,
    location TEXT,
    image_key TEXT,
    status TEXT DEFAULT 'active',
    is_featured BOOLEAN DEFAULT false,
    is_sponsored BOOLEAN DEFAULT false,
    meta_title TEXT,
    meta_description TEXT,
    expires_at DATETIME,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);
CREATE INDEX IF NOT EXISTS idx_classifieds_status_category ON classifieds(status, category, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_classifieds_featured ON classifieds(is_featured, status, created_at DESC);

CREATE TABLE IF NOT EXISTS banners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    advertiser_name TEXT,
    contact_name TEXT,
    contact_phone TEXT,
    contact_whatsapp TEXT,
    price_display TEXT,
    notes TEXT,
    position TEXT NOT NULL,
    image_key TEXT NOT NULL,
    link_url TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status TEXT DEFAULT 'active',
    active BOOLEAN DEFAULT true,
    priority INTEGER DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS neighborhoods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    meta_title TEXT,
    meta_description TEXT,
    cover_image_key TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);

CREATE TABLE IF NOT EXISTS influencers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    name TEXT NOT NULL,
    bio TEXT,
    city_area TEXT,
    niche TEXT,
    instagram TEXT,
    tiktok TEXT,
    youtube TEXT,
    whatsapp TEXT,
    avatar_key TEXT,
    cover_image_key TEXT,
    meta_title TEXT,
    meta_description TEXT,
    is_featured BOOLEAN DEFAULT false,
    is_sponsored BOOLEAN DEFAULT false,
    active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(tenant_id, slug)
);

CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    payload TEXT NOT NULL DEFAULT '{}',
    status TEXT DEFAULT 'pending',
    run_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    error TEXT,
    processed_at DATETIME
);

CREATE TABLE IF NOT EXISTS dead_jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    type TEXT NOT NULL,
    payload TEXT NOT NULL DEFAULT '{}',
    status TEXT DEFAULT 'failed',
    run_at DATETIME NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    attempts INTEGER DEFAULT 0,
    max_attempts INTEGER DEFAULT 3,
    error TEXT,
    processed_at DATETIME
);

CREATE TABLE IF NOT EXISTS metrics (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
    metric_type TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    user_id INTEGER,
    ip_address TEXT,
    user_agent TEXT,
    referrer TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS audit_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    user_id INTEGER,
    action TEXT NOT NULL,
    entity_type TEXT,
    entity_id INTEGER,
    changes TEXT,
    ip_address TEXT,
    user_agent TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS edit_locks (
    entity_type TEXT NOT NULL,
    entity_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    locked_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    expires_at DATETIME NOT NULL,
    PRIMARY KEY (entity_type, entity_id)
);

CREATE TABLE IF NOT EXISTS login_attempts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ip_address TEXT NOT NULL,
    email TEXT,
    success BOOLEAN DEFAULT false,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT OR IGNORE INTO categories (tenant_id, slug, name, requires_editorial_notes) VALUES
(1, 'noticias', 'Noticias', false),
(1, 'politica-bastidores', 'Politica & Bastidores', true),
(1, 'influencers', 'Influencers da Cidade', false),
(1, 'eventos', 'Eventos', false);
`
	if _, err := r.db.Exec(schema); err != nil {
		return err
	}
	if err := r.ensureSQLiteColumns(); err != nil {
		return err
	}
	return r.markSQLiteSchemaVersion()
}

func (r *Repository) ensureSQLiteColumns() error {
	if err := r.ensureSQLitePortalSettingsTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteUsersTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteTenantUsersSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteCategoriesTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteAutomationTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteTagsTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLitePostsTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteMediaAssetsTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteNeighborhoodsTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteStoresTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLitePromotionsTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteEventsTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteClassifiedsTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteInfluencersTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteMetricsTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureSQLiteJobsTenantSchema(); err != nil {
		return err
	}
	if err := r.ensureColumn("posts", "is_featured", "BOOLEAN DEFAULT false"); err != nil {
		return err
	}
	postColumns := []struct {
		name       string
		definition string
	}{
		{"meta_title", "TEXT DEFAULT ''"},
		{"meta_description", "TEXT DEFAULT ''"},
		{"seo_keyword", "TEXT DEFAULT ''"},
		{"canonical_url", "TEXT DEFAULT ''"},
		{"source_name", "TEXT DEFAULT ''"},
		{"source_url", "TEXT DEFAULT ''"},
		{"reading_time_minutes", "INTEGER DEFAULT 1"},
		{"is_pinned", "BOOLEAN DEFAULT false"},
		{"gallery_image_keys", "TEXT DEFAULT '[]'"},
	}
	for _, column := range postColumns {
		if err := r.ensureColumn("posts", column.name, column.definition); err != nil {
			return err
		}
	}
	categoryColumns := []struct {
		name       string
		definition string
	}{
		{"image_key", "TEXT DEFAULT ''"},
		{"sort_order", "INTEGER DEFAULT 0"},
		{"active", "BOOLEAN DEFAULT true"},
		{"meta_title", "TEXT DEFAULT ''"},
		{"meta_description", "TEXT DEFAULT ''"},
	}
	for _, column := range categoryColumns {
		if err := r.ensureColumn("categories", column.name, column.definition); err != nil {
			return err
		}
	}
	tagColumns := []struct {
		name       string
		definition string
	}{
		{"meta_title", "TEXT DEFAULT ''"},
		{"meta_description", "TEXT DEFAULT ''"},
	}
	for _, column := range tagColumns {
		if err := r.ensureColumn("tags", column.name, column.definition); err != nil {
			return err
		}
	}
	bannerColumns := []struct {
		name       string
		definition string
	}{
		{"tenant_id", "INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE"},
		{"advertiser_name", "TEXT DEFAULT ''"},
		{"contact_name", "TEXT DEFAULT ''"},
		{"contact_phone", "TEXT DEFAULT ''"},
		{"contact_whatsapp", "TEXT DEFAULT ''"},
		{"price_display", "TEXT DEFAULT ''"},
		{"notes", "TEXT DEFAULT ''"},
		{"status", "TEXT DEFAULT 'active'"},
	}
	for _, column := range bannerColumns {
		if err := r.ensureColumn("banners", column.name, column.definition); err != nil {
			return err
		}
	}
	storeColumns := []struct {
		name       string
		definition string
	}{
		{"website_url", "TEXT DEFAULT ''"},
		{"commercial_status", "TEXT DEFAULT 'active'"},
		{"meta_title", "TEXT DEFAULT ''"},
		{"meta_description", "TEXT DEFAULT ''"},
	}
	for _, column := range storeColumns {
		if err := r.ensureColumn("stores", column.name, column.definition); err != nil {
			return err
		}
	}
	promotionColumns := []struct {
		name       string
		definition string
	}{
		{"coupon_code", "TEXT DEFAULT ''"},
		{"meta_title", "TEXT DEFAULT ''"},
		{"meta_description", "TEXT DEFAULT ''"},
	}
	for _, column := range promotionColumns {
		if err := r.ensureColumn("promotions", column.name, column.definition); err != nil {
			return err
		}
	}
	influencerColumns := []struct {
		name       string
		definition string
	}{
		{"niche", "TEXT DEFAULT ''"},
		{"meta_title", "TEXT DEFAULT ''"},
		{"meta_description", "TEXT DEFAULT ''"},
	}
	for _, column := range influencerColumns {
		if err := r.ensureColumn("influencers", column.name, column.definition); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ensureSQLiteCategoriesTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='categories'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "slug text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE categories RENAME TO categories_legacy;
			CREATE TABLE categories (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				slug TEXT NOT NULL,
				name TEXT NOT NULL,
				description TEXT,
				meta_title TEXT,
				meta_description TEXT,
				image_key TEXT,
				sort_order INTEGER DEFAULT 0,
				active BOOLEAN DEFAULT true,
				requires_editorial_notes BOOLEAN DEFAULT false,
				UNIQUE(tenant_id, slug)
			);
			INSERT INTO categories (
				id, tenant_id, slug, name, description, meta_title, meta_description, image_key, sort_order, active, requires_editorial_notes
			)
			SELECT
				id, 1, slug, name, description, meta_title, meta_description, image_key, sort_order, active, requires_editorial_notes
			FROM categories_legacy;
			DROP TABLE categories_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_categories_tenant_slug ON categories(tenant_id, slug);
		CREATE INDEX IF NOT EXISTS idx_categories_tenant_active ON categories(tenant_id, active, sort_order);
	`)
	return err
}

func (r *Repository) ensureSQLiteTagsTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='tags'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "slug text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE tags RENAME TO tags_legacy;
			CREATE TABLE tags (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				slug TEXT NOT NULL,
				name TEXT NOT NULL,
				description TEXT,
				meta_title TEXT,
				meta_description TEXT,
				active BOOLEAN DEFAULT true,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, slug)
			);
			INSERT INTO tags (
				id, tenant_id, slug, name, description, meta_title, meta_description, active, created_at
			)
			SELECT
				id, 1, slug, name, description, meta_title, meta_description, active, created_at
			FROM tags_legacy;
			DROP TABLE tags_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_tags_tenant_slug ON tags(tenant_id, slug);
		CREATE INDEX IF NOT EXISTS idx_tags_tenant_active ON tags(tenant_id, active, name);
	`)
	return err
}

func (r *Repository) ensureSQLitePostsTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='posts'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "slug text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			DROP TRIGGER IF EXISTS posts_fts_insert;
			DROP TRIGGER IF EXISTS posts_fts_delete;
			DROP TRIGGER IF EXISTS posts_fts_update;
			DROP TABLE IF EXISTS posts_fts;
			PRAGMA foreign_keys=off;
			ALTER TABLE posts RENAME TO posts_legacy;
			CREATE TABLE posts (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				title TEXT NOT NULL,
				slug TEXT NOT NULL,
				excerpt TEXT,
				content TEXT NOT NULL,
				cover_image_key TEXT,
				gallery_image_keys TEXT,
				meta_title TEXT,
				meta_description TEXT,
				seo_keyword TEXT,
				canonical_url TEXT,
				source_name TEXT,
				source_url TEXT,
				reading_time_minutes INTEGER DEFAULT 1,
				category_id INTEGER REFERENCES categories(id),
				author_id INTEGER REFERENCES users(id),
				status TEXT DEFAULT 'draft',
				is_sponsored BOOLEAN DEFAULT false,
				is_featured BOOLEAN DEFAULT false,
				is_pinned BOOLEAN DEFAULT false,
				editorial_notes TEXT,
				editor_responsible TEXT,
				published_at DATETIME,
				publish_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, slug)
			);
			INSERT INTO posts (
				id, tenant_id, title, slug, excerpt, content, cover_image_key, gallery_image_keys, meta_title, meta_description, seo_keyword, canonical_url, source_name, source_url, reading_time_minutes, category_id, author_id, status, is_sponsored, is_featured, is_pinned, editorial_notes, editor_responsible, published_at, publish_at, created_at, updated_at
			)
			SELECT
				id, 1, title, slug, excerpt, content, cover_image_key, gallery_image_keys, meta_title, meta_description, seo_keyword, canonical_url, source_name, source_url, reading_time_minutes, category_id, author_id, status, is_sponsored, is_featured, is_pinned, editorial_notes, editor_responsible, published_at, publish_at, created_at, updated_at
			FROM posts_legacy;
			DROP TABLE posts_legacy;
			PRAGMA foreign_keys=on;
			CREATE VIRTUAL TABLE IF NOT EXISTS posts_fts USING fts5(title, excerpt, content, content_rowid=rowid, content=posts);
			INSERT INTO posts_fts(rowid, title, excerpt, content)
			SELECT id, title, excerpt, content FROM posts;
			CREATE TRIGGER IF NOT EXISTS posts_fts_insert AFTER INSERT ON posts BEGIN
				INSERT INTO posts_fts(rowid, title, excerpt, content) VALUES (new.id, new.title, new.excerpt, new.content);
			END;
			CREATE TRIGGER IF NOT EXISTS posts_fts_delete AFTER DELETE ON posts BEGIN
				INSERT INTO posts_fts(posts_fts, rowid, title, excerpt, content) VALUES ('delete', old.id, old.title, old.excerpt, old.content);
			END;
			CREATE TRIGGER IF NOT EXISTS posts_fts_update AFTER UPDATE ON posts BEGIN
				INSERT INTO posts_fts(posts_fts, rowid, title, excerpt, content) VALUES ('delete', old.id, old.title, old.excerpt, old.content);
				INSERT INTO posts_fts(rowid, title, excerpt, content) VALUES (new.id, new.title, new.excerpt, new.content);
			END;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_posts_tenant_slug ON posts(tenant_id, slug);
		CREATE INDEX IF NOT EXISTS idx_posts_tenant_status ON posts(tenant_id, status, published_at DESC);
		CREATE INDEX IF NOT EXISTS idx_posts_tenant_category ON posts(tenant_id, category_id, status);
		CREATE INDEX IF NOT EXISTS idx_posts_tenant_featured ON posts(tenant_id, is_featured, status, updated_at DESC);
	`)
	return err
}

func (r *Repository) ensureSQLiteMediaAssetsTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='media_assets'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "key text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE media_assets RENAME TO media_assets_legacy;
			CREATE TABLE media_assets (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				key TEXT NOT NULL,
				original_name TEXT NOT NULL,
				title TEXT,
				alt_text TEXT,
				content_type TEXT NOT NULL,
				size_bytes INTEGER DEFAULT 0,
				uploaded_by INTEGER REFERENCES users(id),
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, key)
			);
			INSERT INTO media_assets (
				id, tenant_id, key, original_name, title, alt_text, content_type, size_bytes, uploaded_by, created_at, updated_at
			)
			SELECT
				id, 1, key, original_name, title, alt_text, content_type, size_bytes, uploaded_by, created_at, updated_at
			FROM media_assets_legacy;
			DROP TABLE media_assets_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_media_assets_tenant_key ON media_assets(tenant_id, key);
		CREATE INDEX IF NOT EXISTS idx_media_assets_tenant_created ON media_assets(tenant_id, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_media_assets_tenant_search ON media_assets(tenant_id, original_name, title, alt_text);
	`)
	return err
}

func (r *Repository) ensureSQLiteNeighborhoodsTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='neighborhoods'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "slug text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE neighborhoods RENAME TO neighborhoods_legacy;
			CREATE TABLE neighborhoods (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				slug TEXT NOT NULL,
				name TEXT NOT NULL,
				description TEXT,
				meta_title TEXT,
				meta_description TEXT,
				cover_image_key TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, slug)
			);
			INSERT INTO neighborhoods (
				id, tenant_id, slug, name, description, meta_title, meta_description, cover_image_key, created_at
			)
			SELECT
				id, 1, slug, name, description, meta_title, meta_description, cover_image_key, created_at
			FROM neighborhoods_legacy;
			DROP TABLE neighborhoods_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_neighborhoods_tenant_slug ON neighborhoods(tenant_id, slug);
		CREATE INDEX IF NOT EXISTS idx_neighborhoods_tenant_name ON neighborhoods(tenant_id, name);
	`)
	return err
}

func (r *Repository) ensureSQLiteStoresTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='stores'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "slug text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE stores RENAME TO stores_legacy;
			CREATE TABLE stores (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				slug TEXT NOT NULL,
				name TEXT NOT NULL,
				description TEXT,
				category TEXT,
				address TEXT,
				phone TEXT,
				whatsapp TEXT,
				website_url TEXT,
				logo_key TEXT,
				cover_image_key TEXT,
				commercial_status TEXT DEFAULT 'active',
				meta_title TEXT,
				meta_description TEXT,
				is_sponsored BOOLEAN DEFAULT false,
				is_featured BOOLEAN DEFAULT false,
				neighborhood_id INTEGER,
				active BOOLEAN DEFAULT true,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, slug)
			);
			INSERT INTO stores (
				id, tenant_id, slug, name, description, category, address, phone, whatsapp, website_url, logo_key, cover_image_key, commercial_status, meta_title, meta_description, is_sponsored, is_featured, neighborhood_id, active, created_at
			)
			SELECT
				id, 1, slug, name, description, category, address, phone, whatsapp, website_url, logo_key, cover_image_key, commercial_status, meta_title, meta_description, is_sponsored, is_featured, neighborhood_id, active, created_at
			FROM stores_legacy;
			DROP TABLE stores_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_stores_tenant_slug ON stores(tenant_id, slug);
		CREATE INDEX IF NOT EXISTS idx_stores_tenant_active ON stores(tenant_id, active, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_stores_tenant_featured ON stores(tenant_id, active, is_featured, created_at DESC);
	`)
	return err
}

func (r *Repository) ensureSQLitePromotionsTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='promotions'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "slug text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE promotions RENAME TO promotions_legacy;
			CREATE TABLE promotions (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				store_id INTEGER REFERENCES stores(id),
				title TEXT NOT NULL,
				slug TEXT NOT NULL,
				description TEXT,
				price_display TEXT,
				coupon_code TEXT,
				image_key TEXT,
				start_date DATE NOT NULL,
				end_date DATE NOT NULL,
				status TEXT DEFAULT 'active',
				is_sponsored BOOLEAN DEFAULT false,
				meta_title TEXT,
				meta_description TEXT,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, slug)
			);
			INSERT INTO promotions (
				id, tenant_id, store_id, title, slug, description, price_display, coupon_code, image_key, start_date, end_date, status, is_sponsored, meta_title, meta_description, created_at
			)
			SELECT
				id, 1, store_id, title, slug, description, price_display, coupon_code, image_key, start_date, end_date, status, is_sponsored, meta_title, meta_description, created_at
			FROM promotions_legacy;
			DROP TABLE promotions_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_promotions_tenant_slug ON promotions(tenant_id, slug);
		CREATE INDEX IF NOT EXISTS idx_promotions_tenant_active ON promotions(tenant_id, status, end_date);
	`)
	return err
}

func (r *Repository) ensureSQLiteEventsTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='events'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "slug text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE events RENAME TO events_legacy;
			CREATE TABLE events (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				slug TEXT NOT NULL,
				title TEXT NOT NULL,
				description TEXT,
				location TEXT,
				organizer TEXT,
				ticket_url TEXT,
				price_display TEXT,
				image_key TEXT,
				status TEXT DEFAULT 'active',
				is_featured BOOLEAN DEFAULT false,
				is_sponsored BOOLEAN DEFAULT false,
				meta_title TEXT,
				meta_description TEXT,
				start_at DATETIME NOT NULL,
				end_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, slug)
			);
			INSERT INTO events (
				id, tenant_id, slug, title, description, location, organizer, ticket_url, price_display, image_key, status, is_featured, is_sponsored, meta_title, meta_description, start_at, end_at, created_at, updated_at
			)
			SELECT
				id, 1, slug, title, description, location, organizer, ticket_url, price_display, image_key, status, is_featured, is_sponsored, meta_title, meta_description, start_at, end_at, created_at, updated_at
			FROM events_legacy;
			DROP TABLE events_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_events_tenant_slug ON events(tenant_id, slug);
		CREATE INDEX IF NOT EXISTS idx_events_tenant_status_start ON events(tenant_id, status, start_at);
		CREATE INDEX IF NOT EXISTS idx_events_tenant_featured ON events(tenant_id, is_featured, status, start_at);
	`)
	return err
}

func (r *Repository) ensureSQLiteClassifiedsTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='classifieds'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "slug text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE classifieds RENAME TO classifieds_legacy;
			CREATE TABLE classifieds (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				slug TEXT NOT NULL,
				title TEXT NOT NULL,
				description TEXT,
				category TEXT,
				price_display TEXT,
				contact_name TEXT,
				contact_phone TEXT,
				contact_whatsapp TEXT,
				location TEXT,
				image_key TEXT,
				status TEXT DEFAULT 'active',
				is_featured BOOLEAN DEFAULT false,
				is_sponsored BOOLEAN DEFAULT false,
				meta_title TEXT,
				meta_description TEXT,
				expires_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, slug)
			);
			INSERT INTO classifieds (
				id, tenant_id, slug, title, description, category, price_display, contact_name, contact_phone, contact_whatsapp, location, image_key, status, is_featured, is_sponsored, meta_title, meta_description, expires_at, created_at, updated_at
			)
			SELECT
				id, 1, slug, title, description, category, price_display, contact_name, contact_phone, contact_whatsapp, location, image_key, status, is_featured, is_sponsored, meta_title, meta_description, expires_at, created_at, updated_at
			FROM classifieds_legacy;
			DROP TABLE classifieds_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_classifieds_tenant_slug ON classifieds(tenant_id, slug);
		CREATE INDEX IF NOT EXISTS idx_classifieds_tenant_status_category ON classifieds(tenant_id, status, category, created_at DESC);
		CREATE INDEX IF NOT EXISTS idx_classifieds_tenant_featured ON classifieds(tenant_id, is_featured, status, created_at DESC);
	`)
	return err
}

func (r *Repository) ensureSQLiteInfluencersTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='influencers'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "slug text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE influencers RENAME TO influencers_legacy;
			CREATE TABLE influencers (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				slug TEXT NOT NULL,
				name TEXT NOT NULL,
				bio TEXT,
				city_area TEXT,
				niche TEXT,
				instagram TEXT,
				tiktok TEXT,
				youtube TEXT,
				whatsapp TEXT,
				avatar_key TEXT,
				cover_image_key TEXT,
				meta_title TEXT,
				meta_description TEXT,
				is_featured BOOLEAN DEFAULT false,
				is_sponsored BOOLEAN DEFAULT false,
				active BOOLEAN DEFAULT true,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, slug)
			);
			INSERT INTO influencers (
				id, tenant_id, slug, name, bio, city_area, niche, instagram, tiktok, youtube, whatsapp, avatar_key, cover_image_key, meta_title, meta_description, is_featured, is_sponsored, active, created_at
			)
			SELECT
				id, 1, slug, name, bio, city_area, niche, instagram, tiktok, youtube, whatsapp, avatar_key, cover_image_key, meta_title, meta_description, is_featured, is_sponsored, active, created_at
			FROM influencers_legacy;
			DROP TABLE influencers_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_influencers_tenant_slug ON influencers(tenant_id, slug);
		CREATE INDEX IF NOT EXISTS idx_influencers_tenant_active ON influencers(tenant_id, active, is_featured, created_at DESC);
	`)
	return err
}

func (r *Repository) ensureSQLiteMetricsTenantSchema() error {
	if err := r.ensureColumn("metrics", "tenant_id", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		return err
	}
	_, err := r.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_metrics_tenant_entity ON metrics(tenant_id, entity_type, entity_id, metric_type);
		CREATE INDEX IF NOT EXISTS idx_metrics_tenant_date ON metrics(tenant_id, created_at);
	`)
	return err
}

func (r *Repository) ensureSQLitePortalSettingsTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='portal_settings'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "check (id = 1)")
	if needsRebuild {
		_, err := r.db.Exec(`
			ALTER TABLE portal_settings RENAME TO portal_settings_legacy;
			CREATE TABLE portal_settings (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				site_name TEXT NOT NULL,
				tagline TEXT,
				logo_key TEXT,
				favicon_key TEXT,
				contact_email TEXT,
				contact_whatsapp TEXT,
				contact_phone TEXT,
				city TEXT,
				state TEXT,
				seo_title TEXT,
				seo_description TEXT,
				facebook_url TEXT,
				instagram_url TEXT,
				youtube_url TEXT,
				tiktok_url TEXT,
				upload_max_mb INTEGER DEFAULT 2,
				automation_enabled BOOLEAN DEFAULT false,
				automation_interval_minutes INTEGER DEFAULT 60,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id)
			);
			INSERT INTO portal_settings (
				id, tenant_id, site_name, tagline, logo_key, favicon_key, contact_email, contact_whatsapp, contact_phone,
				city, state, seo_title, seo_description, facebook_url, instagram_url, youtube_url, tiktok_url,
				upload_max_mb, automation_enabled, automation_interval_minutes, updated_at
			)
			SELECT
				id, 1, site_name, tagline, logo_key, favicon_key, contact_email, contact_whatsapp, contact_phone,
				city, state, seo_title, seo_description, facebook_url, instagram_url, youtube_url, tiktok_url,
				upload_max_mb, automation_enabled, automation_interval_minutes, updated_at
			FROM portal_settings_legacy;
			DROP TABLE portal_settings_legacy;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS idx_portal_settings_tenant_id ON portal_settings(tenant_id)`)
	return err
}

func (r *Repository) ensureSQLiteUsersTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='users'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	normalized := strings.ToLower(createSQL)
	needsRebuild := !strings.Contains(normalized, "tenant_id") || strings.Contains(normalized, "email text unique")
	if needsRebuild {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE users RENAME TO users_legacy;
			CREATE TABLE users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				name TEXT NOT NULL,
				email TEXT NOT NULL,
				password_hash TEXT NOT NULL,
				role TEXT DEFAULT 'editor',
				active BOOLEAN DEFAULT true,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				UNIQUE(tenant_id, email)
			);
			INSERT INTO users (id, tenant_id, name, email, password_hash, role, active, created_at)
			SELECT id, 1, name, email, password_hash, role, active, created_at
			FROM users_legacy;
			DROP TABLE users_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`
		CREATE UNIQUE INDEX IF NOT EXISTS idx_users_tenant_email ON users(tenant_id, email);
		CREATE INDEX IF NOT EXISTS idx_users_tenant_role ON users(tenant_id, role, active);
	`)
	return err
}

func (r *Repository) ensureSQLiteTenantUsersSchema() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS tenant_users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			tenant_id INTEGER NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			role TEXT NOT NULL DEFAULT 'editor',
			active BOOLEAN NOT NULL DEFAULT true,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			UNIQUE(tenant_id, user_id)
		);
		INSERT OR IGNORE INTO tenant_users (tenant_id, user_id, role, active, created_at, updated_at)
		SELECT tenant_id, id, role, active, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
		FROM users;
		CREATE INDEX IF NOT EXISTS idx_tenant_users_user ON tenant_users(user_id, active);
		CREATE INDEX IF NOT EXISTS idx_tenant_users_tenant_role ON tenant_users(tenant_id, role, active);
	`)
	return err
}

func (r *Repository) ensureSQLiteAutomationTenantSchema() error {
	if err := r.ensureSQLiteAutomationSourcesTenantSchema(); err != nil {
		return err
	}
	return r.ensureSQLiteAutomationRunsTenantSchema()
}

func (r *Repository) ensureSQLiteAutomationSourcesTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='automation_sources'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if !strings.Contains(strings.ToLower(createSQL), "tenant_id") {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE automation_sources RENAME TO automation_sources_legacy;
			CREATE TABLE automation_sources (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				name TEXT NOT NULL,
				source_type TEXT NOT NULL DEFAULT 'rss',
				url TEXT NOT NULL,
				default_category_id INTEGER REFERENCES categories(id),
				active BOOLEAN DEFAULT true,
				last_run_at DATETIME,
				created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
			);
			INSERT INTO automation_sources (
				id, tenant_id, name, source_type, url, default_category_id, active, last_run_at, created_at, updated_at
			)
			SELECT id, 1, name, source_type, url, default_category_id, active, last_run_at, created_at, updated_at
			FROM automation_sources_legacy;
			DROP TABLE automation_sources_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`CREATE INDEX IF NOT EXISTS idx_automation_sources_active ON automation_sources(tenant_id, active, source_type)`)
	return err
}

func (r *Repository) ensureSQLiteAutomationRunsTenantSchema() error {
	var createSQL string
	err := r.db.QueryRow(`SELECT sql FROM sqlite_master WHERE type='table' AND name='automation_runs'`).Scan(&createSQL)
	if err == sql.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if !strings.Contains(strings.ToLower(createSQL), "tenant_id") {
		_, err := r.db.Exec(`
			PRAGMA foreign_keys=off;
			ALTER TABLE automation_runs RENAME TO automation_runs_legacy;
			CREATE TABLE automation_runs (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				tenant_id INTEGER NOT NULL DEFAULT 1 REFERENCES tenants(id) ON DELETE CASCADE,
				source_id INTEGER REFERENCES automation_sources(id),
				status TEXT NOT NULL DEFAULT 'success',
				items_found INTEGER DEFAULT 0,
				drafts_created INTEGER DEFAULT 0,
				duplicates INTEGER DEFAULT 0,
				error TEXT,
				log TEXT,
				started_at DATETIME DEFAULT CURRENT_TIMESTAMP,
				finished_at DATETIME
			);
			INSERT INTO automation_runs (
				id, tenant_id, source_id, status, items_found, drafts_created, duplicates, error, log, started_at, finished_at
			)
			SELECT id, 1, source_id, status, items_found, drafts_created, duplicates, error, log, started_at, finished_at
			FROM automation_runs_legacy;
			DROP TABLE automation_runs_legacy;
			PRAGMA foreign_keys=on;
		`)
		if err != nil {
			return err
		}
	}
	_, err = r.db.Exec(`CREATE INDEX IF NOT EXISTS idx_automation_runs_started ON automation_runs(tenant_id, started_at DESC)`)
	return err
}

func (r *Repository) ensureSQLiteJobsTenantSchema() error {
	if err := r.ensureColumn("jobs", "tenant_id", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		return err
	}
	if err := r.ensureColumn("dead_jobs", "tenant_id", "INTEGER NOT NULL DEFAULT 1"); err != nil {
		return err
	}
	_, err := r.db.Exec(`
		CREATE INDEX IF NOT EXISTS idx_jobs_tenant_pending ON jobs(tenant_id, run_at, status);
		CREATE INDEX IF NOT EXISTS idx_dead_jobs_tenant_created ON dead_jobs(tenant_id, created_at DESC);
	`)
	return err
}

func (r *Repository) ensureColumn(table, column, definition string) error {
	rows, err := r.db.Query(`PRAGMA table_info(` + table + `)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var name, columnType string
		var notNull int
		var defaultValue any
		var pk int
		if err := rows.Scan(&cid, &name, &columnType, &notNull, &defaultValue, &pk); err != nil {
			return err
		}
		if name == column {
			return nil
		}
	}
	if err := rows.Err(); err != nil {
		return err
	}
	_, err = r.db.Exec(`ALTER TABLE ` + table + ` ADD COLUMN ` + column + ` ` + definition)
	return err
}

func (r *Repository) Close() error {
	return r.db.Close()
}

// Users

func userSelectSQL() string {
	return `
		SELECT tu.tenant_id, u.id, u.name, u.email, u.password_hash, tu.role, (u.active AND tu.active), u.created_at
		FROM users u
		INNER JOIN tenant_users tu ON tu.user_id = u.id`
}

func scanUser(scanner interface {
	Scan(dest ...any) error
}) (*model.User, error) {
	var u model.User
	err := scanner.Scan(&u.TenantID, &u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Active, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("usuario nao encontrado")
	}
	return &u, err
}

func (r *Repository) UserGetByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, userSelectSQL()+`
		WHERE tu.tenant_id = $1 AND lower(u.email) = lower($2)
		ORDER BY CASE WHEN u.tenant_id = tu.tenant_id THEN 0 ELSE 1 END
		LIMIT 1`, tenantIDFromContext(ctx), email)
	return scanUser(row)
}

func (r *Repository) UserGetByID(ctx context.Context, id int64) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, userSelectSQL()+` WHERE tu.tenant_id = $1 AND u.id = $2`, tenantIDFromContext(ctx), id)
	return scanUser(row)
}

func (r *Repository) UserGetAnyByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT tenant_id, id, name, email, password_hash, role, active, created_at
		FROM users
		WHERE lower(email) = lower($1)
		ORDER BY id ASC
		LIMIT 1`, email)
	return scanUser(row)
}

func (r *Repository) UserListAny(ctx context.Context, limit int) ([]model.User, error) {
	if limit <= 0 {
		limit = 200
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT tenant_id, id, name, email, password_hash, role, active, created_at
		FROM users
		ORDER BY name ASC, email ASC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *Repository) UserCreate(ctx context.Context, u *model.User) error {
	if u.TenantID <= 0 {
		u.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `INSERT INTO users (tenant_id, name, email, password_hash, role, active) VALUES ($1, $2, $3, $4, $5, $6)`,
		u.TenantID, u.Name, u.Email, u.PasswordHash, u.Role, u.Active)
	if err != nil {
		return err
	}
	u.ID = id
	if err := r.TenantUserUpsert(ctx, &model.TenantUser{
		TenantID: u.TenantID,
		UserID:   u.ID,
		Role:     u.Role,
		Active:   u.Active,
	}); err != nil {
		return err
	}
	return nil
}

func (r *Repository) UserUpdatePassword(ctx context.Context, id int64, hash string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET password_hash = $1
		WHERE id = $2
			AND EXISTS (
				SELECT 1 FROM tenant_users tu
				WHERE tu.tenant_id = $3 AND tu.user_id = users.id AND tu.active = true
			)`, hash, id, tenantIDFromContext(ctx))
	return err
}

func (r *Repository) UserUpdate(ctx context.Context, u *model.User) error {
	if u.TenantID <= 0 {
		u.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET name = $1,
			email = $2,
			role = CASE WHEN tenant_id = $5 THEN $3 ELSE role END,
			active = CASE WHEN tenant_id = $5 THEN $4 ELSE active END
		WHERE id = $6
			AND EXISTS (
				SELECT 1 FROM tenant_users tu
				WHERE tu.tenant_id = $5 AND tu.user_id = users.id
			)`,
		u.Name, u.Email, u.Role, u.Active, u.TenantID, u.ID)
	if err != nil {
		return err
	}
	return r.TenantUserUpsert(ctx, &model.TenantUser{
		TenantID: u.TenantID,
		UserID:   u.ID,
		Role:     u.Role,
		Active:   u.Active,
	})
}

func (r *Repository) UserSoftDelete(ctx context.Context, id int64) error {
	tenantID := tenantIDFromContext(ctx)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE tenant_users
		SET active = false, updated_at = CURRENT_TIMESTAMP
		WHERE tenant_id = $1 AND user_id = $2`, tenantID, id); err != nil {
		tx.Rollback()
		return err
	}
	if _, err := tx.ExecContext(ctx, `
		UPDATE users
		SET active = false
		WHERE id = $1
			AND NOT EXISTS (
				SELECT 1 FROM tenant_users
				WHERE user_id = $1 AND active = true
			)`, id); err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *Repository) UserActiveSuperAdminCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		INNER JOIN tenant_users tu ON tu.user_id = u.id
		WHERE u.active = true AND tu.active = true AND tu.role = 'super_admin'`).Scan(&count)
	return count, err
}

func (r *Repository) UserList(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, userSelectSQL()+` WHERE tu.tenant_id = $1 ORDER BY u.created_at DESC`, tenantIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		u, err := scanUser(rows)
		if err != nil {
			return nil, err
		}
		users = append(users, *u)
	}
	return users, rows.Err()
}

func (r *Repository) PasswordResetInvalidateUser(ctx context.Context, userID int64) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE password_reset_tokens
		SET used_at = CURRENT_TIMESTAMP
		WHERE user_id = $1 AND used_at IS NULL`, userID)
	return err
}

func (r *Repository) PasswordResetCreate(ctx context.Context, token *model.PasswordResetToken) error {
	id, err := r.insertID(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, requested_ip, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		token.UserID, token.TokenHash, token.RequestedIP, token.UserAgent, token.ExpiresAt)
	if err != nil {
		return err
	}
	token.ID = id
	return nil
}

func (r *Repository) PasswordResetGetActive(ctx context.Context, tokenHash string, now time.Time) (*model.PasswordResetToken, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, token_hash, COALESCE(requested_ip, ''), COALESCE(user_agent, ''), expires_at, used_at, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1 AND used_at IS NULL AND expires_at > $2`, tokenHash, now)
	var token model.PasswordResetToken
	var usedAt sql.NullTime
	err := row.Scan(&token.ID, &token.UserID, &token.TokenHash, &token.RequestedIP, &token.UserAgent, &token.ExpiresAt, &usedAt, &token.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if usedAt.Valid {
		token.UsedAt = &usedAt.Time
	}
	return &token, err
}

func (r *Repository) PasswordResetMarkUsed(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE password_reset_tokens SET used_at = CURRENT_TIMESTAMP WHERE id = $1 AND used_at IS NULL`, id)
	return err
}

// Categories

func (r *Repository) CategoryGetBySlug(ctx context.Context, slug string) (*model.Category, error) {
	row := r.db.QueryRowContext(ctx, categorySelectSQL()+` WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug)
	var c model.Category
	err := scanCategory(row, &c)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *Repository) CategoryGetByID(ctx context.Context, id int64) (*model.Category, error) {
	row := r.db.QueryRowContext(ctx, categorySelectSQL()+` WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	var c model.Category
	err := scanCategory(row, &c)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *Repository) CategoryList(ctx context.Context) ([]model.Category, error) {
	rows, err := r.db.QueryContext(ctx, categorySelectSQL()+` WHERE tenant_id = $1 ORDER BY sort_order ASC, name ASC`, tenantIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []model.Category
	for rows.Next() {
		var c model.Category
		if err := scanCategory(rows, &c); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *Repository) CategoryCreate(ctx context.Context, c *model.Category) error {
	if c.TenantID <= 0 {
		c.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO categories (tenant_id, slug, name, description, meta_title, meta_description, image_key, sort_order, active, requires_editorial_notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		c.TenantID, c.Slug, c.Name, c.Description, c.MetaTitle, c.MetaDescription, c.ImageKey, c.SortOrder, c.Active, c.RequiresEditorialNotes)
	if err != nil {
		return err
	}
	c.ID = id
	return nil
}

func (r *Repository) CategoryUpdate(ctx context.Context, c *model.Category) error {
	if c.TenantID <= 0 {
		c.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE categories
		SET slug=$1, name=$2, description=$3, meta_title=$4, meta_description=$5, image_key=$6, sort_order=$7, active=$8, requires_editorial_notes=$9
		WHERE id=$10 AND tenant_id=$11`,
		c.Slug, c.Name, c.Description, c.MetaTitle, c.MetaDescription, c.ImageKey, c.SortOrder, c.Active, c.RequiresEditorialNotes, c.ID, c.TenantID)
	return err
}

func (r *Repository) CategoryDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) CategoryPostCount(ctx context.Context, id int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE tenant_id = $1 AND category_id = $2`, tenantIDFromContext(ctx), id).Scan(&count)
	return count, err
}

func (r *Repository) CategorySlugExists(ctx context.Context, slug string) bool {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM categories WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug).Scan(&id)
	return err == nil
}

func categorySelectSQL() string {
	return `SELECT tenant_id, id, slug, name, COALESCE(description, ''), COALESCE(meta_title, ''), COALESCE(meta_description, ''), COALESCE(image_key, ''), COALESCE(sort_order, 0), COALESCE(active, true), requires_editorial_notes FROM categories`
}

type categoryScanner interface {
	Scan(dest ...any) error
}

func scanCategory(scanner categoryScanner, c *model.Category) error {
	return scanner.Scan(&c.TenantID, &c.ID, &c.Slug, &c.Name, &c.Description, &c.MetaTitle, &c.MetaDescription, &c.ImageKey, &c.SortOrder, &c.Active, &c.RequiresEditorialNotes)
}

// Tags

func (r *Repository) TagGetBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	row := r.db.QueryRowContext(ctx, tagSelectSQL()+` WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug)
	var tag model.Tag
	err := scanTag(row, &tag)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tag, err
}

func (r *Repository) TagGetByID(ctx context.Context, id int64) (*model.Tag, error) {
	row := r.db.QueryRowContext(ctx, tagSelectSQL()+` WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	var tag model.Tag
	err := scanTag(row, &tag)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tag, err
}

func (r *Repository) TagList(ctx context.Context, activeOnly bool) ([]model.Tag, error) {
	query := tagSelectSQL() + ` WHERE tenant_id = $1`
	if activeOnly {
		query += ` AND active = true`
	}
	query += ` ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query, tenantIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if err := scanTag(rows, &tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *Repository) TagCreate(ctx context.Context, tag *model.Tag) error {
	if tag.TenantID <= 0 {
		tag.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO tags (tenant_id, slug, name, description, meta_title, meta_description, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		tag.TenantID, tag.Slug, tag.Name, tag.Description, tag.MetaTitle, tag.MetaDescription, tag.Active)
	if err != nil {
		return err
	}
	tag.ID = id
	return nil
}

func (r *Repository) TagUpdate(ctx context.Context, tag *model.Tag) error {
	if tag.TenantID <= 0 {
		tag.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE tags
		SET slug=$1, name=$2, description=$3, meta_title=$4, meta_description=$5, active=$6
		WHERE id=$7 AND tenant_id=$8`,
		tag.Slug, tag.Name, tag.Description, tag.MetaTitle, tag.MetaDescription, tag.Active, tag.ID, tag.TenantID)
	return err
}

func (r *Repository) TagDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tags WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) TagPostCount(ctx context.Context, id int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM post_tags pt
		JOIN posts p ON p.id = pt.post_id
		WHERE p.tenant_id = $1 AND pt.tag_id = $2`, tenantIDFromContext(ctx), id).Scan(&count)
	return count, err
}

func (r *Repository) TagSlugExists(ctx context.Context, slug string) bool {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM tags WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug).Scan(&id)
	return err == nil
}

func (r *Repository) PostTags(ctx context.Context, postID int64) ([]model.Tag, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.tenant_id, t.id, t.slug, t.name, COALESCE(t.description, ''), COALESCE(t.meta_title, ''), COALESCE(t.meta_description, ''), COALESCE(t.active, true), t.created_at
		FROM tags t
		JOIN post_tags pt ON pt.tag_id = t.id
		WHERE pt.post_id = $1 AND t.tenant_id = $2
		ORDER BY t.name ASC`, postID, tenantIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tags []model.Tag
	for rows.Next() {
		var tag model.Tag
		if err := scanTag(rows, &tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, rows.Err()
}

func (r *Repository) PostSetTags(ctx context.Context, postID int64, tagIDs []int64) error {
	tenantID := tenantIDFromContext(ctx)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM post_tags WHERE post_id IN (SELECT id FROM posts WHERE tenant_id = $1 AND id = $2)`, tenantID, postID); err != nil {
		return err
	}
	seen := map[int64]bool{}
	for _, tagID := range tagIDs {
		if tagID <= 0 || seen[tagID] {
			continue
		}
		seen[tagID] = true
		var exists int
		if err := tx.QueryRowContext(ctx, `
			SELECT COUNT(*)
			FROM posts p
			JOIN tags t ON t.tenant_id = p.tenant_id
			WHERE p.tenant_id = $1 AND p.id = $2 AND t.id = $3`, tenantID, postID, tagID).Scan(&exists); err != nil {
			return err
		}
		if exists == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES ($1, $2)`, postID, tagID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func tagSelectSQL() string {
	return `SELECT tenant_id, id, slug, name, COALESCE(description, ''), COALESCE(meta_title, ''), COALESCE(meta_description, ''), COALESCE(active, true), created_at FROM tags`
}

type tagScanner interface {
	Scan(dest ...any) error
}

func scanTag(scanner tagScanner, tag *model.Tag) error {
	return scanner.Scan(&tag.TenantID, &tag.ID, &tag.Slug, &tag.Name, &tag.Description, &tag.MetaTitle, &tag.MetaDescription, &tag.Active, &tag.CreatedAt)
}

// Posts

func (r *Repository) PostCreate(ctx context.Context, p *model.Post) error {
	if p.TenantID <= 0 {
		p.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO posts (tenant_id, title, slug, excerpt, content, cover_image_key, gallery_image_keys, meta_title, meta_description, seo_keyword, canonical_url, source_name, source_url, reading_time_minutes, category_id, author_id, status, is_sponsored, is_featured, is_pinned, editorial_notes, editor_responsible, published_at, publish_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24)`,
		p.TenantID, p.Title, p.Slug, p.Excerpt, p.Content, p.CoverImageKey, encodeStringList(p.GalleryImageKeys), p.MetaTitle, p.MetaDescription, p.SEOKeyword, p.CanonicalURL, p.SourceName, p.SourceURL, p.ReadingTimeMinutes, p.CategoryID, p.AuthorID, p.Status, p.IsSponsored, p.IsFeatured, p.IsPinned, p.EditorialNotes, p.EditorResponsible, p.PublishedAt, p.PublishAt)
	if err != nil {
		return err
	}
	p.ID = id
	return nil
}

func (r *Repository) PostUpdate(ctx context.Context, p *model.Post) error {
	if p.TenantID <= 0 {
		p.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE posts SET title=$1, slug=$2, excerpt=$3, content=$4, cover_image_key=$5, gallery_image_keys=$6, meta_title=$7, meta_description=$8, seo_keyword=$9, canonical_url=$10, source_name=$11, source_url=$12, reading_time_minutes=$13, category_id=$14, status=$15, is_sponsored=$16, is_featured=$17, is_pinned=$18, editorial_notes=$19, editor_responsible=$20, published_at=$21, publish_at=$22, updated_at=CURRENT_TIMESTAMP
		WHERE id=$23 AND tenant_id=$24`,
		p.Title, p.Slug, p.Excerpt, p.Content, p.CoverImageKey, encodeStringList(p.GalleryImageKeys), p.MetaTitle, p.MetaDescription, p.SEOKeyword, p.CanonicalURL, p.SourceName, p.SourceURL, p.ReadingTimeMinutes, p.CategoryID, p.Status, p.IsSponsored, p.IsFeatured, p.IsPinned, p.EditorialNotes, p.EditorResponsible, p.PublishedAt, p.PublishAt, p.ID, p.TenantID)
	return err
}

func (r *Repository) PostGetBySlug(ctx context.Context, slug string) (*model.Post, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.content, p.cover_image_key, COALESCE(p.gallery_image_keys, '[]'), COALESCE(p.meta_title, ''), COALESCE(p.meta_description, ''), COALESCE(p.seo_keyword, ''), COALESCE(p.canonical_url, ''), COALESCE(p.source_name, ''), COALESCE(p.source_url, ''), COALESCE(p.reading_time_minutes, 1), p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), COALESCE(p.is_pinned, false), COALESCE(p.editorial_notes, ''), COALESCE(p.editor_responsible, ''), p.published_at, p.publish_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.tenant_id = $1 AND p.slug = $2`, tenantIDFromContext(ctx), slug)
	var p model.Post
	err := scanPostDetail(row, &p)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err == nil {
		p.Tags, _ = r.PostTags(ctx, p.ID)
	}
	return &p, err
}

func (r *Repository) PostGetByID(ctx context.Context, id int64) (*model.Post, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.content, p.cover_image_key, COALESCE(p.gallery_image_keys, '[]'), COALESCE(p.meta_title, ''), COALESCE(p.meta_description, ''), COALESCE(p.seo_keyword, ''), COALESCE(p.canonical_url, ''), COALESCE(p.source_name, ''), COALESCE(p.source_url, ''), COALESCE(p.reading_time_minutes, 1), p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), COALESCE(p.is_pinned, false), COALESCE(p.editorial_notes, ''), COALESCE(p.editor_responsible, ''), p.published_at, p.publish_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.tenant_id = $1 AND p.id = $2`, tenantIDFromContext(ctx), id)
	var p model.Post
	err := scanPostDetail(row, &p)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err == nil {
		p.Tags, _ = r.PostTags(ctx, p.ID)
	}
	return &p, err
}

type postDetailScanner interface {
	Scan(dest ...any) error
}

func scanPostDetail(scanner postDetailScanner, p *model.Post) error {
	var galleryJSON string
	if err := scanner.Scan(&p.TenantID, &p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.Content, &p.CoverImageKey, &galleryJSON, &p.MetaTitle, &p.MetaDescription, &p.SEOKeyword, &p.CanonicalURL, &p.SourceName, &p.SourceURL, &p.ReadingTimeMinutes, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.IsPinned, &p.EditorialNotes, &p.EditorResponsible, &p.PublishedAt, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
		return err
	}
	p.GalleryImageKeys = decodeStringList(galleryJSON)
	return nil
}

func (r *Repository) PostListPublished(ctx context.Context, limit int, offset int) ([]model.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.tenant_id = $1 AND p.status = 'published'
		ORDER BY p.published_at DESC
		LIMIT $2 OFFSET $3`, tenantIDFromContext(ctx), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.TenantID, &p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *Repository) PostGetFeaturedPublished(ctx context.Context) (*model.Post, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), COALESCE(p.is_pinned, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.tenant_id = $1 AND p.status = 'published' AND (COALESCE(p.is_featured, false) = true OR COALESCE(p.is_pinned, false) = true)
		ORDER BY COALESCE(p.is_pinned, false) DESC, p.updated_at DESC, p.published_at DESC
		LIMIT 1`, tenantIDFromContext(ctx))
	var p model.Post
	err := row.Scan(&p.TenantID, &p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.IsPinned, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *Repository) PostListByCategory(ctx context.Context, catID int64, limit int) ([]model.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.tenant_id = $1 AND p.category_id = $2 AND p.status = 'published'
		ORDER BY p.published_at DESC
		LIMIT $3`, tenantIDFromContext(ctx), catID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.TenantID, &p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *Repository) PostListAll(ctx context.Context, limit int, offset int) ([]model.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.publish_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.tenant_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2 OFFSET $3`, tenantIDFromContext(ctx), limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.TenantID, &p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.PublishedAt, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *Repository) PostListAdmin(ctx context.Context, status string, categoryID int64, tagID int64, query string, limit int, offset int) ([]model.Post, error) {
	where, args := adminPostWhere(tenantIDFromContext(ctx), status, categoryID, tagID, query)
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.publish_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		`+where+`
		ORDER BY p.updated_at DESC, p.created_at DESC
		LIMIT $`+fmt.Sprint(len(args)-1)+` OFFSET $`+fmt.Sprint(len(args)), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.TenantID, &p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.PublishedAt, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *Repository) PostDelete(ctx context.Context, id int64) error {
	tenantID := tenantIDFromContext(ctx)
	if _, err := r.db.ExecContext(ctx, `DELETE FROM post_tags WHERE post_id IN (SELECT id FROM posts WHERE tenant_id = $1 AND id = $2)`, tenantID, id); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	return err
}

func (r *Repository) PostCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE tenant_id = $1`, tenantIDFromContext(ctx)).Scan(&count)
	return count, err
}

func (r *Repository) PostCountAdmin(ctx context.Context, status string, categoryID int64, tagID int64, query string) (int, error) {
	where, args := adminPostWhere(tenantIDFromContext(ctx), status, categoryID, tagID, query)
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts p `+where, args...).Scan(&count)
	return count, err
}

func adminPostWhere(tenantID int64, status string, categoryID int64, tagID int64, query string) (string, []any) {
	clauses := []string{"p.tenant_id = $1"}
	args := []any{tenantID}
	if status != "" {
		args = append(args, status)
		clauses = append(clauses, fmt.Sprintf("p.status = $%d", len(args)))
	}
	if categoryID > 0 {
		args = append(args, categoryID)
		clauses = append(clauses, fmt.Sprintf("p.category_id = $%d", len(args)))
	}
	if tagID > 0 {
		args = append(args, tagID)
		clauses = append(clauses, fmt.Sprintf("p.id IN (SELECT post_id FROM post_tags WHERE tag_id = $%d)", len(args)))
	}
	if strings.TrimSpace(query) != "" {
		args = append(args, "%"+strings.TrimSpace(query)+"%")
		clauses = append(clauses, fmt.Sprintf("(p.title LIKE $%d OR p.excerpt LIKE $%d OR p.seo_keyword LIKE $%d)", len(args), len(args), len(args)))
	}
	if len(clauses) == 0 {
		return "", args
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func (r *Repository) PostSlugExists(ctx context.Context, slug string) bool {
	var count int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug).Scan(&count)
	return count > 0
}

func (r *Repository) PostSearch(ctx context.Context, query string, limit int) ([]model.Post, error) {
	tenantID := tenantIDFromContext(ctx)
	if r.driver == "postgres" {
		rows, err := r.db.QueryContext(ctx, `
			SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
			FROM posts p
			LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
			LEFT JOIN users u ON u.id = p.author_id
			WHERE p.tenant_id = $1 AND p.search_vector @@ plainto_tsquery('portuguese', $2) AND p.status = 'published'
			ORDER BY ts_rank_cd(p.search_vector, plainto_tsquery('portuguese', $2)) DESC, p.published_at DESC
			LIMIT $3`, tenantID, query, limit)
		if err == nil {
			defer rows.Close()
			posts, err := scanPostList(rows)
			if err != nil {
				return nil, err
			}
			if len(posts) > 0 {
				return posts, nil
			}
		}
	}
	ftsQuery := buildFTSQuery(query)
	if ftsQuery != "" {
		rows, err := r.db.QueryContext(ctx, `
			SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
			FROM posts_fts f
			JOIN posts p ON p.id = f.rowid
			LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
			LEFT JOIN users u ON u.id = p.author_id
			WHERE posts_fts MATCH $1 AND p.tenant_id = $2 AND p.status = 'published'
			ORDER BY bm25(posts_fts), p.published_at DESC
			LIMIT $3`, ftsQuery, tenantID, limit)
		if err == nil {
			defer rows.Close()
			posts, err := scanPostList(rows)
			if err != nil {
				return nil, err
			}
			if len(posts) > 0 {
				return posts, nil
			}
		}
	}

	q := "%" + query + "%"
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.tenant_id = $2 AND (
			p.title LIKE $1 OR p.excerpt LIKE $1 OR p.content LIKE $1 OR
			EXISTS (
				SELECT 1 FROM post_tags pt
				JOIN tags t ON t.id = pt.tag_id
				WHERE pt.post_id = p.id AND t.tenant_id = p.tenant_id AND t.active = true AND t.name LIKE $1
			)
		) AND p.status = 'published'
		ORDER BY p.published_at DESC
		LIMIT $3`, q, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostList(rows)
}

func (r *Repository) PostListByTag(ctx context.Context, tagID int64, limit int) ([]model.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.tenant_id = $1 AND pt.tag_id = $2 AND p.status = 'published'
		ORDER BY p.published_at DESC
		LIMIT $3`, tenantIDFromContext(ctx), tagID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostList(rows)
}

func buildFTSQuery(query string) string {
	terms := strings.Fields(query)
	if len(terms) == 0 {
		return ""
	}
	for i, term := range terms {
		terms[i] = `"` + strings.ReplaceAll(term, `"`, `""`) + `"`
	}
	return strings.Join(terms, " ")
}

func scanPostList(rows *sql.Rows) ([]model.Post, error) {
	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.TenantID, &p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func encodeStringList(values []string) string {
	if values == nil {
		values = []string{}
	}
	data, err := json.Marshal(values)
	if err != nil {
		return "[]"
	}
	return string(data)
}

func decodeStringList(raw string) []string {
	var values []string
	if err := json.Unmarshal([]byte(raw), &values); err != nil {
		return []string{}
	}
	return values
}

func (r *Repository) PostUpdateStatus(ctx context.Context, id int64, status model.PostStatus) error {
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET status=$1, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$2 AND id=$3`, status, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) PostUpdateStatusAndEditorialNotes(ctx context.Context, id int64, status model.PostStatus, notes string, responsible string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE posts
		SET status=$1, editorial_notes=$2, editor_responsible=$3, updated_at=CURRENT_TIMESTAMP
		WHERE tenant_id=$4 AND id=$5`, status, notes, responsible, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) PostRevisionCreate(ctx context.Context, revision *model.PostRevision) error {
	id, err := r.insertID(ctx, `
		INSERT INTO post_revisions (post_id, user_id, action, title, status, snapshot)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		revision.PostID, revision.UserID, revision.Action, revision.Title, revision.Status, revision.Snapshot)
	if err != nil {
		return err
	}
	revision.ID = id
	return nil
}

func (r *Repository) PostRevisionList(ctx context.Context, postID int64, limit int) ([]model.PostRevision, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT pr.id, pr.post_id, pr.user_id, pr.action, COALESCE(pr.title, ''), pr.status, pr.snapshot, pr.created_at, COALESCE(u.name, '')
		FROM post_revisions pr
		LEFT JOIN users u ON u.id = pr.user_id
		WHERE pr.post_id = $1
		ORDER BY pr.created_at DESC
		LIMIT $2`, postID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var revisions []model.PostRevision
	for rows.Next() {
		var rev model.PostRevision
		if err := rows.Scan(&rev.ID, &rev.PostID, &rev.UserID, &rev.Action, &rev.Title, &rev.Status, &rev.Snapshot, &rev.CreatedAt, &rev.UserName); err != nil {
			return nil, err
		}
		revisions = append(revisions, rev)
	}
	return revisions, rows.Err()
}

func (r *Repository) MediaAssetCreate(ctx context.Context, asset *model.MediaAsset) error {
	if asset.TenantID <= 0 {
		asset.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO media_assets (tenant_id, key, original_name, title, alt_text, content_type, size_bytes, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		asset.TenantID, asset.Key, asset.OriginalName, asset.Title, asset.AltText, asset.ContentType, asset.SizeBytes, asset.UploadedBy)
	if err != nil {
		return err
	}
	asset.ID = id
	return nil
}

func (r *Repository) MediaAssetList(ctx context.Context, query string, limit, offset int) ([]model.MediaAsset, error) {
	return r.MediaAssetListFiltered(ctx, MediaAssetFilter{
		Query:  query,
		Limit:  limit,
		Offset: offset,
	})
}

func (r *Repository) MediaAssetListFiltered(ctx context.Context, filter MediaAssetFilter) ([]model.MediaAsset, error) {
	if filter.Limit <= 0 || filter.Limit > 100 {
		filter.Limit = 24
	}
	where, args := mediaAssetWhere(tenantIDFromContext(ctx), filter)
	args = append(args, filter.Limit, filter.Offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.tenant_id, m.id, m.key, m.original_name, COALESCE(m.title, ''), COALESCE(m.alt_text, ''), m.content_type, COALESCE(m.size_bytes, 0), m.uploaded_by, COALESCE(u.name, ''), m.created_at, m.updated_at
		FROM media_assets m
		LEFT JOIN users u ON u.id = m.uploaded_by
		`+where+`
		ORDER BY m.created_at DESC, m.id DESC
		LIMIT $`+strconv.Itoa(len(args)-1)+` OFFSET $`+strconv.Itoa(len(args)), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var assets []model.MediaAsset
	for rows.Next() {
		asset, err := scanMediaAsset(rows)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range assets {
		assets[i].UsageCount, _ = r.MediaAssetUsageCount(ctx, assets[i].Key)
	}
	return assets, nil
}

func (r *Repository) MediaAssetCount(ctx context.Context, query string) (int, error) {
	return r.MediaAssetCountFiltered(ctx, MediaAssetFilter{Query: query})
}

func (r *Repository) MediaAssetCountFiltered(ctx context.Context, filter MediaAssetFilter) (int, error) {
	where, args := mediaAssetWhere(tenantIDFromContext(ctx), filter)
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM media_assets m `+where, args...)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (r *Repository) MediaAssetArchiveMonths(ctx context.Context) ([]model.MediaArchiveMonth, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT substr(CAST(created_at AS TEXT), 1, 7) AS month, COUNT(*)
		FROM media_assets
		WHERE tenant_id = $1
		GROUP BY month
		ORDER BY month DESC`, tenantIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var months []model.MediaArchiveMonth
	for rows.Next() {
		var month model.MediaArchiveMonth
		if err := rows.Scan(&month.Month, &month.Count); err != nil {
			return nil, err
		}
		month.Label = mediaMonthLabel(month.Month)
		months = append(months, month)
	}
	return months, rows.Err()
}

func (r *Repository) MediaAssetGetByID(ctx context.Context, id int64) (*model.MediaAsset, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT m.tenant_id, m.id, m.key, m.original_name, COALESCE(m.title, ''), COALESCE(m.alt_text, ''), m.content_type, COALESCE(m.size_bytes, 0), m.uploaded_by, COALESCE(u.name, ''), m.created_at, m.updated_at
		FROM media_assets m
		LEFT JOIN users u ON u.id = m.uploaded_by
		WHERE m.tenant_id = $1 AND m.id = $2`, tenantIDFromContext(ctx), id)
	asset, err := scanMediaAsset(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	asset.UsageCount, _ = r.MediaAssetUsageCount(ctx, asset.Key)
	return &asset, nil
}

func (r *Repository) MediaAssetGetByKey(ctx context.Context, key string) (*model.MediaAsset, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT m.tenant_id, m.id, m.key, m.original_name, COALESCE(m.title, ''), COALESCE(m.alt_text, ''), m.content_type, COALESCE(m.size_bytes, 0), m.uploaded_by, COALESCE(u.name, ''), m.created_at, m.updated_at
		FROM media_assets m
		LEFT JOIN users u ON u.id = m.uploaded_by
		WHERE m.tenant_id = $1 AND m.key = $2`, tenantIDFromContext(ctx), key)
	asset, err := scanMediaAsset(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	asset.UsageCount, _ = r.MediaAssetUsageCount(ctx, asset.Key)
	return &asset, nil
}

func (r *Repository) MediaAssetUpdate(ctx context.Context, asset *model.MediaAsset) error {
	if asset.TenantID <= 0 {
		asset.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE media_assets
		SET title=$1, alt_text=$2, updated_at=CURRENT_TIMESTAMP
		WHERE tenant_id=$3 AND id=$4`, asset.Title, asset.AltText, asset.TenantID, asset.ID)
	return err
}

func (r *Repository) MediaAssetDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM media_assets WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) MediaAssetUsageCount(ctx context.Context, key string) (int, error) {
	tenantID := tenantIDFromContext(ctx)
	row := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total), 0) FROM (
			SELECT COUNT(*) AS total FROM posts WHERE tenant_id = $2 AND (cover_image_key = $1 OR COALESCE(gallery_image_keys, '') LIKE '%' || $1 || '%')
			UNION ALL SELECT COUNT(*) FROM categories WHERE tenant_id = $2 AND image_key = $1
			UNION ALL SELECT COUNT(*) FROM stores WHERE tenant_id = $2 AND (logo_key = $1 OR cover_image_key = $1)
			UNION ALL SELECT COUNT(*) FROM promotions WHERE tenant_id = $2 AND image_key = $1
			UNION ALL SELECT COUNT(*) FROM events WHERE tenant_id = $2 AND image_key = $1
			UNION ALL SELECT COUNT(*) FROM classifieds WHERE tenant_id = $2 AND image_key = $1
			UNION ALL SELECT COUNT(*) FROM banners WHERE tenant_id = $2 AND image_key = $1
			UNION ALL SELECT COUNT(*) FROM neighborhoods WHERE tenant_id = $2 AND cover_image_key = $1
			UNION ALL SELECT COUNT(*) FROM influencers WHERE tenant_id = $2 AND (avatar_key = $1 OR cover_image_key = $1)
			UNION ALL SELECT COUNT(*) FROM portal_settings WHERE tenant_id = $2 AND (logo_key = $1 OR favicon_key = $1)
		)`, key, tenantID)
	var count int
	err := row.Scan(&count)
	return count, err
}

func mediaAssetWhere(tenantID int64, filter MediaAssetFilter) (string, []any) {
	conditions := []string{"m.tenant_id = $1"}
	args := []any{tenantID}
	query := strings.TrimSpace(filter.Query)
	if query != "" {
		args = append(args, "%"+strings.ToLower(query)+"%")
		placeholder := "$" + strconv.Itoa(len(args))
		conditions = append(conditions, "(LOWER(m.original_name) LIKE "+placeholder+" OR LOWER(COALESCE(m.title, '')) LIKE "+placeholder+" OR LOWER(COALESCE(m.alt_text, '')) LIKE "+placeholder+" OR LOWER(m.key) LIKE "+placeholder+")")
	}
	if filter.DateFrom != nil {
		args = append(args, *filter.DateFrom)
		conditions = append(conditions, "m.created_at >= $"+strconv.Itoa(len(args)))
	}
	if filter.DateTo != nil {
		args = append(args, *filter.DateTo)
		conditions = append(conditions, "m.created_at < $"+strconv.Itoa(len(args)))
	}
	return "WHERE " + strings.Join(conditions, " AND "), args
}

func mediaMonthLabel(month string) string {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return month
	}
	names := []string{"Janeiro", "Fevereiro", "Marco", "Abril", "Maio", "Junho", "Julho", "Agosto", "Setembro", "Outubro", "Novembro", "Dezembro"}
	return names[int(t.Month())-1] + " " + strconv.Itoa(t.Year())
}

type mediaAssetScanner interface {
	Scan(dest ...any) error
}

func scanMediaAsset(scanner mediaAssetScanner) (model.MediaAsset, error) {
	var asset model.MediaAsset
	var uploadedBy sql.NullInt64
	if err := scanner.Scan(&asset.TenantID, &asset.ID, &asset.Key, &asset.OriginalName, &asset.Title, &asset.AltText, &asset.ContentType, &asset.SizeBytes, &uploadedBy, &asset.UploadedByName, &asset.CreatedAt, &asset.UpdatedAt); err != nil {
		return asset, err
	}
	if uploadedBy.Valid {
		id := uploadedBy.Int64
		asset.UploadedBy = &id
	}
	return asset, nil
}

func (r *Repository) PostSetPublished(ctx context.Context, id int64, t time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET status='published', published_at=$1, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$2 AND id=$3`, t, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) PostSetFeatured(ctx context.Context, id int64, featured bool) error {
	tenantID := tenantIDFromContext(ctx)
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if featured {
		if _, err := tx.ExecContext(ctx, `UPDATE posts SET is_featured=false WHERE tenant_id=$1 AND is_featured=true`, tenantID); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE posts SET is_featured=$1, updated_at=CURRENT_TIMESTAMP WHERE tenant_id=$2 AND id=$3`, featured, tenantID, id); err != nil {
		return err
	}
	return tx.Commit()
}

// Slug Redirects

func (r *Repository) SlugRedirectCreate(ctx context.Context, oldSlug, newSlug, entityType string) error {
	_, err := r.db.ExecContext(ctx, `INSERT OR REPLACE INTO slug_redirects (old_slug, new_slug, entity_type) VALUES ($1, $2, $3)`, oldSlug, newSlug, entityType)
	return err
}

func (r *Repository) SlugRedirectGet(ctx context.Context, oldSlug string) (*model.SlugRedirect, error) {
	row := r.db.QueryRowContext(ctx, `SELECT old_slug, new_slug, entity_type, created_at FROM slug_redirects WHERE old_slug = $1`, oldSlug)
	var s model.SlugRedirect
	err := row.Scan(&s.OldSlug, &s.NewSlug, &s.EntityType, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

// Stores

func (r *Repository) StoreCreate(ctx context.Context, s *model.Store) error {
	if s.TenantID <= 0 {
		s.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO stores (tenant_id, slug, name, description, category, address, phone, whatsapp, website_url, logo_key, cover_image_key, commercial_status, meta_title, meta_description, is_sponsored, is_featured, neighborhood_id, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)`,
		s.TenantID, s.Slug, s.Name, s.Description, s.Category, s.Address, s.Phone, s.Whatsapp, s.WebsiteURL, s.LogoKey, s.CoverImageKey, normalizeStoreCommercialStatus(s.CommercialStatus), s.MetaTitle, s.MetaDescription, s.IsSponsored, s.IsFeatured, s.NeighborhoodID, s.Active)
	if err != nil {
		return err
	}
	s.ID = id
	return nil
}

func (r *Repository) StoreUpdate(ctx context.Context, s *model.Store) error {
	if s.TenantID <= 0 {
		s.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE stores SET slug=$1, name=$2, description=$3, category=$4, address=$5, phone=$6, whatsapp=$7, website_url=$8, logo_key=$9, cover_image_key=$10, commercial_status=$11, meta_title=$12, meta_description=$13, is_sponsored=$14, is_featured=$15, neighborhood_id=$16, active=$17
		WHERE tenant_id=$18 AND id=$19`,
		s.Slug, s.Name, s.Description, s.Category, s.Address, s.Phone, s.Whatsapp, s.WebsiteURL, s.LogoKey, s.CoverImageKey, normalizeStoreCommercialStatus(s.CommercialStatus), s.MetaTitle, s.MetaDescription, s.IsSponsored, s.IsFeatured, s.NeighborhoodID, s.Active, s.TenantID, s.ID)
	return err
}

func (r *Repository) StoreGetBySlug(ctx context.Context, slug string) (*model.Store, error) {
	row := r.db.QueryRowContext(ctx, storeSelectSQL()+` WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug)
	s, err := scanStore(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *Repository) StoreGetByID(ctx context.Context, id int64) (*model.Store, error) {
	row := r.db.QueryRowContext(ctx, storeSelectSQL()+` WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	s, err := scanStore(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *Repository) StoreList(ctx context.Context, activeOnly bool, limit int) ([]model.Store, error) {
	q := storeSelectSQL() + ` WHERE tenant_id = $1`
	args := []any{tenantIDFromContext(ctx)}
	if activeOnly {
		q += ` AND active = true`
	}
	q += ` ORDER BY created_at DESC LIMIT $2`
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stores []model.Store
	for rows.Next() {
		s, err := scanStore(rows)
		if err != nil {
			return nil, err
		}
		stores = append(stores, s)
	}
	return stores, rows.Err()
}

func (r *Repository) StoreListFeatured(ctx context.Context, limit int) ([]model.Store, error) {
	rows, err := r.db.QueryContext(ctx, storeSelectSQL()+`
		WHERE tenant_id = $1 AND active = true AND is_featured = true
		ORDER BY created_at DESC LIMIT $2`, tenantIDFromContext(ctx), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stores []model.Store
	for rows.Next() {
		s, err := scanStore(rows)
		if err != nil {
			return nil, err
		}
		stores = append(stores, s)
	}
	return stores, rows.Err()
}

func (r *Repository) StoreDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM stores WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) StoreSlugExists(ctx context.Context, slug string) bool {
	var count int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM stores WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug).Scan(&count)
	return count > 0
}

func storeSelectSQL() string {
	return `SELECT tenant_id, id, slug, name, description, category, address, phone, whatsapp, COALESCE(website_url, ''), logo_key, cover_image_key, COALESCE(commercial_status, 'active'), COALESCE(meta_title, ''), COALESCE(meta_description, ''), is_sponsored, is_featured, neighborhood_id, active, created_at FROM stores`
}

func scanStore(scanner interface{ Scan(dest ...any) error }) (model.Store, error) {
	var s model.Store
	err := scanner.Scan(
		&s.TenantID,
		&s.ID,
		&s.Slug,
		&s.Name,
		&s.Description,
		&s.Category,
		&s.Address,
		&s.Phone,
		&s.Whatsapp,
		&s.WebsiteURL,
		&s.LogoKey,
		&s.CoverImageKey,
		&s.CommercialStatus,
		&s.MetaTitle,
		&s.MetaDescription,
		&s.IsSponsored,
		&s.IsFeatured,
		&s.NeighborhoodID,
		&s.Active,
		&s.CreatedAt,
	)
	return s, err
}

func normalizeStoreCommercialStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "lead", "paused", "inactive":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "active"
	}
}

// Promotions

func (r *Repository) PromotionCreate(ctx context.Context, p *model.Promotion) error {
	if p.TenantID <= 0 {
		p.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO promotions (tenant_id, store_id, title, slug, description, price_display, coupon_code, image_key, start_date, end_date, status, is_sponsored, meta_title, meta_description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		p.TenantID, p.StoreID, p.Title, p.Slug, p.Description, p.PriceDisplay, p.CouponCode, p.ImageKey, p.StartDate, p.EndDate, normalizePromotionStatus(p.Status), p.IsSponsored, p.MetaTitle, p.MetaDescription)
	if err != nil {
		return err
	}
	p.ID = id
	return nil
}

func (r *Repository) PromotionUpdate(ctx context.Context, p *model.Promotion) error {
	if p.TenantID <= 0 {
		p.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE promotions SET store_id=$1, title=$2, slug=$3, description=$4, price_display=$5, coupon_code=$6, image_key=$7, start_date=$8, end_date=$9, status=$10, is_sponsored=$11, meta_title=$12, meta_description=$13
		WHERE tenant_id=$14 AND id=$15`,
		p.StoreID, p.Title, p.Slug, p.Description, p.PriceDisplay, p.CouponCode, p.ImageKey, p.StartDate, p.EndDate, normalizePromotionStatus(p.Status), p.IsSponsored, p.MetaTitle, p.MetaDescription, p.TenantID, p.ID)
	return err
}

func (r *Repository) PromotionGetBySlug(ctx context.Context, slug string) (*model.Promotion, error) {
	row := r.db.QueryRowContext(ctx, promotionSelectSQL()+` WHERE p.tenant_id = $1 AND p.slug = $2`, tenantIDFromContext(ctx), slug)
	p, err := scanPromotion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *Repository) PromotionGetByID(ctx context.Context, id int64) (*model.Promotion, error) {
	row := r.db.QueryRowContext(ctx, promotionSelectSQL()+` WHERE p.tenant_id = $1 AND p.id = $2`, tenantIDFromContext(ctx), id)
	p, err := scanPromotion(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *Repository) PromotionListActive(ctx context.Context, limit int) ([]model.Promotion, error) {
	today := time.Now().Format("2006-01-02")
	rows, err := r.db.QueryContext(ctx, promotionSelectSQL()+`
		WHERE p.tenant_id = $1 AND p.status = 'active' AND p.start_date <= $2 AND p.end_date >= $2
		ORDER BY p.created_at DESC
		LIMIT $3`, tenantIDFromContext(ctx), today, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var promos []model.Promotion
	for rows.Next() {
		p, err := scanPromotion(rows)
		if err != nil {
			return nil, err
		}
		promos = append(promos, p)
	}
	return promos, rows.Err()
}

func (r *Repository) PromotionListAdmin(ctx context.Context, limit int) ([]model.Promotion, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, promotionSelectSQL()+`
		WHERE p.tenant_id = $1
		ORDER BY p.created_at DESC
		LIMIT $2`, tenantIDFromContext(ctx), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var promos []model.Promotion
	for rows.Next() {
		p, err := scanPromotion(rows)
		if err != nil {
			return nil, err
		}
		promos = append(promos, p)
	}
	return promos, rows.Err()
}

func (r *Repository) PromotionUpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE promotions SET status=$1 WHERE tenant_id=$2 AND id=$3`, status, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) PromotionDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM promotions WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) PromotionSlugExists(ctx context.Context, slug string) bool {
	var count int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM promotions WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug).Scan(&count)
	return count > 0
}

func promotionSelectSQL() string {
	return `
		SELECT p.tenant_id, p.id, p.store_id, p.title, p.slug, p.description, p.price_display, COALESCE(p.coupon_code, ''), p.image_key, p.start_date, p.end_date, p.status, p.is_sponsored, COALESCE(p.meta_title, ''), COALESCE(p.meta_description, ''), p.created_at, s.name, s.slug,
		       COALESCE((SELECT COUNT(*) FROM metrics m WHERE m.tenant_id = p.tenant_id AND m.metric_type = 'promo_click' AND m.entity_type = 'promotion' AND m.entity_id = p.id), 0)
		FROM promotions p
		JOIN stores s ON s.id = p.store_id AND s.tenant_id = p.tenant_id`
}

func scanPromotion(scanner interface{ Scan(dest ...any) error }) (model.Promotion, error) {
	var p model.Promotion
	err := scanner.Scan(
		&p.TenantID,
		&p.ID,
		&p.StoreID,
		&p.Title,
		&p.Slug,
		&p.Description,
		&p.PriceDisplay,
		&p.CouponCode,
		&p.ImageKey,
		&p.StartDate,
		&p.EndDate,
		&p.Status,
		&p.IsSponsored,
		&p.MetaTitle,
		&p.MetaDescription,
		&p.CreatedAt,
		&p.StoreName,
		&p.StoreSlug,
		&p.ClickCount,
	)
	return p, err
}

func normalizePromotionStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "draft", "expired":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "active"
	}
}

// Banners

func (r *Repository) BannerCreate(ctx context.Context, b *model.Banner) error {
	if b.TenantID <= 0 {
		b.TenantID = tenantIDFromContext(ctx)
	}
	if b.Status == "" {
		b.Status = bannerStatusFromActive(b.Active)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO banners (tenant_id, name, advertiser_name, contact_name, contact_phone, contact_whatsapp, price_display, notes, position, image_key, link_url, start_date, end_date, status, active, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		b.TenantID, b.Name, b.AdvertiserName, b.ContactName, b.ContactPhone, b.ContactWhatsapp, b.PriceDisplay, b.Notes, b.Position, b.ImageKey, b.LinkURL, b.StartDate, b.EndDate, b.Status, b.Active, b.Priority)
	if err != nil {
		return err
	}
	b.ID = id
	return nil
}

func (r *Repository) BannerUpdate(ctx context.Context, b *model.Banner) error {
	if b.TenantID <= 0 {
		b.TenantID = tenantIDFromContext(ctx)
	}
	if b.Status == "" {
		b.Status = bannerStatusFromActive(b.Active)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE banners SET name=$1, advertiser_name=$2, contact_name=$3, contact_phone=$4, contact_whatsapp=$5, price_display=$6, notes=$7, position=$8, image_key=$9, link_url=$10, start_date=$11, end_date=$12, status=$13, active=$14, priority=$15
		WHERE tenant_id=$16 AND id=$17`,
		b.Name, b.AdvertiserName, b.ContactName, b.ContactPhone, b.ContactWhatsapp, b.PriceDisplay, b.Notes, b.Position, b.ImageKey, b.LinkURL, b.StartDate, b.EndDate, b.Status, b.Active, b.Priority, b.TenantID, b.ID)
	return err
}

func (r *Repository) BannerGetByID(ctx context.Context, id int64) (*model.Banner, error) {
	row := r.db.QueryRowContext(ctx, bannerSelectSQL()+` WHERE b.tenant_id = $1 AND b.id = $2`, tenantIDFromContext(ctx), id)
	b, err := scanBanner(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return b, err
}

func (r *Repository) BannerGetActiveByPosition(ctx context.Context, position string) (*model.Banner, error) {
	today := time.Now().Format("2006-01-02")
	row := r.db.QueryRowContext(ctx, bannerSelectSQL()+`
		WHERE b.tenant_id = $1 AND b.position = $2 AND b.active = true AND COALESCE(b.status, 'active') = 'active' AND b.start_date <= $3 AND b.end_date >= $3 AND b.image_key <> ''
		ORDER BY b.priority DESC, b.created_at DESC
		LIMIT 1`, tenantIDFromContext(ctx), position, today)
	b, err := scanBanner(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return b, err
}

func (r *Repository) BannerList(ctx context.Context) ([]model.Banner, error) {
	rows, err := r.db.QueryContext(ctx, bannerSelectSQL()+` WHERE b.tenant_id = $1 ORDER BY b.created_at DESC`, tenantIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var banners []model.Banner
	for rows.Next() {
		b, err := scanBanner(rows)
		if err != nil {
			return nil, err
		}
		banners = append(banners, *b)
	}
	return banners, rows.Err()
}

func bannerSelectSQL() string {
	return `
		SELECT b.tenant_id, b.id, b.name, COALESCE(b.advertiser_name, ''), COALESCE(b.contact_name, ''), COALESCE(b.contact_phone, ''), COALESCE(b.contact_whatsapp, ''),
		       COALESCE(b.price_display, ''), COALESCE(b.notes, ''), b.position, b.image_key, b.link_url, b.start_date, b.end_date,
		       COALESCE(b.status, CASE WHEN b.active THEN 'active' ELSE 'paused' END), b.active, b.priority, b.created_at,
		       COALESCE((SELECT COUNT(*) FROM metrics m WHERE m.tenant_id = b.tenant_id AND m.metric_type = 'banner_impression' AND m.entity_type = 'banner' AND m.entity_id = b.id), 0),
		       COALESCE((SELECT COUNT(*) FROM metrics m WHERE m.tenant_id = b.tenant_id AND m.metric_type = 'banner_click' AND m.entity_type = 'banner' AND m.entity_id = b.id), 0)
		FROM banners b`
}

func scanBanner(scanner interface{ Scan(dest ...any) error }) (*model.Banner, error) {
	var b model.Banner
	err := scanner.Scan(&b.TenantID, &b.ID, &b.Name, &b.AdvertiserName, &b.ContactName, &b.ContactPhone, &b.ContactWhatsapp, &b.PriceDisplay, &b.Notes, &b.Position, &b.ImageKey, &b.LinkURL, &b.StartDate, &b.EndDate, &b.Status, &b.Active, &b.Priority, &b.CreatedAt, &b.ImpressionCount, &b.ClickCount)
	return &b, err
}

func (r *Repository) BannerDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM banners WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) BannerCountActiveInPeriod(ctx context.Context, position string, start, end time.Time) (int, error) {
	return r.BannerCountActiveInPeriodExcluding(ctx, position, start, end, 0)
}

func (r *Repository) BannerCountActiveInPeriodExcluding(ctx context.Context, position string, start, end time.Time, excludeID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM banners 
		WHERE tenant_id = $1 AND position = $2 AND active = true AND COALESCE(status, 'active') = 'active'
		AND substr(start_date, 1, 10) <= $3
		AND substr(end_date, 1, 10) >= $4
		AND ($5 = 0 OR id <> $5)`, tenantIDFromContext(ctx), position, end.Format("2006-01-02"), start.Format("2006-01-02"), excludeID).Scan(&count)
	return count, err
}

func bannerStatusFromActive(active bool) string {
	if active {
		return "active"
	}
	return "paused"
}

// Neighborhoods

func (r *Repository) NeighborhoodCreate(ctx context.Context, n *model.Neighborhood) error {
	if n.TenantID <= 0 {
		n.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO neighborhoods (tenant_id, slug, name, description, meta_title, meta_description, cover_image_key)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		n.TenantID, n.Slug, n.Name, n.Description, n.MetaTitle, n.MetaDescription, n.CoverImageKey)
	if err != nil {
		return err
	}
	n.ID = id
	return nil
}

func (r *Repository) NeighborhoodUpdate(ctx context.Context, n *model.Neighborhood) error {
	if n.TenantID <= 0 {
		n.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE neighborhoods SET slug=$1, name=$2, description=$3, meta_title=$4, meta_description=$5, cover_image_key=$6
		WHERE tenant_id=$7 AND id=$8`,
		n.Slug, n.Name, n.Description, n.MetaTitle, n.MetaDescription, n.CoverImageKey, n.TenantID, n.ID)
	return err
}

func (r *Repository) NeighborhoodGetBySlug(ctx context.Context, slug string) (*model.Neighborhood, error) {
	row := r.db.QueryRowContext(ctx, `SELECT tenant_id, id, slug, name, description, meta_title, meta_description, cover_image_key, created_at FROM neighborhoods WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug)
	var n model.Neighborhood
	err := row.Scan(&n.TenantID, &n.ID, &n.Slug, &n.Name, &n.Description, &n.MetaTitle, &n.MetaDescription, &n.CoverImageKey, &n.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &n, err
}

func (r *Repository) NeighborhoodList(ctx context.Context) ([]model.Neighborhood, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT tenant_id, id, slug, name, description, meta_title, meta_description, cover_image_key, created_at FROM neighborhoods WHERE tenant_id = $1 ORDER BY name`, tenantIDFromContext(ctx))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var neighborhoods []model.Neighborhood
	for rows.Next() {
		var n model.Neighborhood
		if err := rows.Scan(&n.TenantID, &n.ID, &n.Slug, &n.Name, &n.Description, &n.MetaTitle, &n.MetaDescription, &n.CoverImageKey, &n.CreatedAt); err != nil {
			return nil, err
		}
		neighborhoods = append(neighborhoods, n)
	}
	return neighborhoods, rows.Err()
}

func (r *Repository) NeighborhoodDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM neighborhoods WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) NeighborhoodSlugExists(ctx context.Context, slug string) bool {
	var count int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM neighborhoods WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug).Scan(&count)
	return count > 0
}

// Influencers

func (r *Repository) InfluencerCreate(ctx context.Context, i *model.Influencer) error {
	if i.TenantID <= 0 {
		i.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO influencers (tenant_id, slug, name, bio, city_area, niche, instagram, tiktok, youtube, whatsapp, avatar_key, cover_image_key, meta_title, meta_description, is_featured, is_sponsored, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		i.TenantID, i.Slug, i.Name, i.Bio, i.CityArea, i.Niche, i.Instagram, i.TikTok, i.YouTube, i.Whatsapp, i.AvatarKey, i.CoverImageKey, i.MetaTitle, i.MetaDescription, i.IsFeatured, i.IsSponsored, i.Active)
	if err != nil {
		return err
	}
	i.ID = id
	return nil
}

func (r *Repository) InfluencerUpdate(ctx context.Context, i *model.Influencer) error {
	if i.TenantID <= 0 {
		i.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE influencers SET slug=$1, name=$2, bio=$3, city_area=$4, niche=$5, instagram=$6, tiktok=$7, youtube=$8, whatsapp=$9, avatar_key=$10, cover_image_key=$11, meta_title=$12, meta_description=$13, is_featured=$14, is_sponsored=$15, active=$16
		WHERE tenant_id=$17 AND id=$18`,
		i.Slug, i.Name, i.Bio, i.CityArea, i.Niche, i.Instagram, i.TikTok, i.YouTube, i.Whatsapp, i.AvatarKey, i.CoverImageKey, i.MetaTitle, i.MetaDescription, i.IsFeatured, i.IsSponsored, i.Active, i.TenantID, i.ID)
	return err
}

func (r *Repository) InfluencerGetBySlug(ctx context.Context, slug string) (*model.Influencer, error) {
	row := r.db.QueryRowContext(ctx, influencerSelectSQL()+` WHERE i.tenant_id = $1 AND i.slug = $2`, tenantIDFromContext(ctx), slug)
	var i model.Influencer
	err := scanInfluencer(row, &i)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &i, err
}

func (r *Repository) InfluencerGetByID(ctx context.Context, id int64) (*model.Influencer, error) {
	row := r.db.QueryRowContext(ctx, influencerSelectSQL()+` WHERE i.tenant_id = $1 AND i.id = $2`, tenantIDFromContext(ctx), id)
	var i model.Influencer
	err := scanInfluencer(row, &i)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &i, err
}

func (r *Repository) InfluencerList(ctx context.Context, activeOnly bool, limit int) ([]model.Influencer, error) {
	q := influencerSelectSQL() + ` WHERE i.tenant_id = $1`
	args := []any{tenantIDFromContext(ctx)}
	if activeOnly {
		q += ` AND i.active = true`
	}
	q += ` ORDER BY i.is_featured DESC, i.is_sponsored DESC, i.created_at DESC`
	if limit > 0 {
		q += ` LIMIT $2`
		args = append(args, limit)
	}
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var influencers []model.Influencer
	for rows.Next() {
		var i model.Influencer
		if err := scanInfluencer(rows, &i); err != nil {
			return nil, err
		}
		influencers = append(influencers, i)
	}
	return influencers, rows.Err()
}

func influencerSelectSQL() string {
	return `
		SELECT i.tenant_id, i.id, i.slug, i.name, i.bio, i.city_area, i.niche, i.instagram, i.tiktok, i.youtube, i.whatsapp, i.avatar_key, i.cover_image_key,
		       i.meta_title, i.meta_description, i.is_featured, i.is_sponsored, i.active, i.created_at,
		       COALESCE((SELECT COUNT(*) FROM metrics m WHERE m.tenant_id = i.tenant_id AND m.metric_type = 'influencer_view' AND m.entity_type = 'influencer' AND m.entity_id = i.id), 0)
		FROM influencers i`
}

type influencerScanner interface {
	Scan(dest ...any) error
}

func scanInfluencer(scanner influencerScanner, i *model.Influencer) error {
	return scanner.Scan(&i.TenantID, &i.ID, &i.Slug, &i.Name, &i.Bio, &i.CityArea, &i.Niche, &i.Instagram, &i.TikTok, &i.YouTube, &i.Whatsapp, &i.AvatarKey, &i.CoverImageKey, &i.MetaTitle, &i.MetaDescription, &i.IsFeatured, &i.IsSponsored, &i.Active, &i.CreatedAt, &i.ViewCount)
}

func (r *Repository) InfluencerDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM influencers WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) InfluencerSlugExists(ctx context.Context, slug string) bool {
	var count int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM influencers WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug).Scan(&count)
	return count > 0
}

// Jobs

func (r *Repository) JobCreate(ctx context.Context, j *model.Job) error {
	if j.TenantID <= 0 {
		j.TenantID = tenantIDFromContext(ctx)
	}
	if j.MaxAttempts <= 0 {
		j.MaxAttempts = 3
	}
	id, err := r.insertID(ctx, `
		INSERT INTO jobs (tenant_id, type, payload, status, run_at, max_attempts)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		j.TenantID, j.Type, j.Payload, j.Status, j.RunAt, j.MaxAttempts)
	if err != nil {
		return err
	}
	j.ID = id
	return nil
}

func (r *Repository) JobHasActiveType(ctx context.Context, jobType model.JobType) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM jobs
		WHERE tenant_id = $1 AND type = $2 AND status IN ('pending', 'running')`, tenantIDFromContext(ctx), jobType).Scan(&count)
	return count > 0, err
}

func (r *Repository) JobGetPending(ctx context.Context, limit int) ([]model.Job, error) {
	now := time.Now()
	rows, err := r.db.QueryContext(ctx, `
		SELECT tenant_id, id, type, payload, status, run_at, created_at, attempts, max_attempts, error, processed_at
		FROM jobs
		WHERE status = 'pending' AND run_at <= $1
		ORDER BY run_at ASC
		LIMIT $2`, now, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var jobs []model.Job
	for rows.Next() {
		var j model.Job
		if err := rows.Scan(&j.TenantID, &j.ID, &j.Type, &j.Payload, &j.Status, &j.RunAt, &j.CreatedAt, &j.Attempts, &j.MaxAttempts, &j.Error, &j.ProcessedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (r *Repository) JobClaimPending(ctx context.Context, limit int) ([]model.Job, error) {
	if r.driver == "postgres" {
		now := time.Now()
		rows, err := r.db.QueryContext(ctx, `
			WITH due AS (
				SELECT id
				FROM jobs
				WHERE status = 'pending' AND run_at <= $1
				ORDER BY run_at ASC
				LIMIT $2
				FOR UPDATE SKIP LOCKED
			)
			UPDATE jobs
			SET status=$3, error='', processed_at=NULL
			FROM due
			WHERE jobs.id = due.id
			RETURNING jobs.tenant_id, jobs.id, jobs.type, jobs.payload, jobs.status, jobs.run_at, jobs.created_at, jobs.attempts, jobs.max_attempts, jobs.error, jobs.processed_at`,
			now, limit, model.JobRunning)
		if err != nil {
			return nil, err
		}
		defer rows.Close()
		var jobs []model.Job
		for rows.Next() {
			var job model.Job
			if err := rows.Scan(&job.TenantID, &job.ID, &job.Type, &job.Payload, &job.Status, &job.RunAt, &job.CreatedAt, &job.Attempts, &job.MaxAttempts, &job.Error, &job.ProcessedAt); err != nil {
				return nil, err
			}
			jobs = append(jobs, job)
		}
		return jobs, rows.Err()
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	now := time.Now()
	rows, err := tx.QueryContext(ctx, `
		SELECT id
		FROM jobs
		WHERE status = 'pending' AND run_at <= $1
		ORDER BY run_at ASC
		LIMIT $2`, now, limit)
	if err != nil {
		return nil, err
	}
	var ids []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			rows.Close()
			return nil, err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	var jobs []model.Job
	for _, id := range ids {
		res, err := tx.ExecContext(ctx, `
			UPDATE jobs
			SET status=$1, error='', processed_at=NULL
			WHERE id=$2 AND status='pending' AND run_at <= $3`, model.JobRunning, id, now)
		if err != nil {
			return nil, err
		}
		affected, err := res.RowsAffected()
		if err != nil {
			return nil, err
		}
		if affected != 1 {
			continue
		}

		var job model.Job
		err = tx.QueryRowContext(ctx, `
			SELECT tenant_id, id, type, payload, status, run_at, created_at, attempts, max_attempts, error, processed_at
			FROM jobs
			WHERE id = $1`, id).
			Scan(&job.TenantID, &job.ID, &job.Type, &job.Payload, &job.Status, &job.RunAt, &job.CreatedAt, &job.Attempts, &job.MaxAttempts, &job.Error, &job.ProcessedAt)
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return jobs, nil
}

func (r *Repository) JobPendingCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM jobs WHERE status = 'pending'`).Scan(&count)
	return count, err
}

func (r *Repository) JobUpdateStatus(ctx context.Context, id int64, status model.JobStatus, errMsg string) error {
	var processed *time.Time
	if status == model.JobCompleted || status == model.JobFailed {
		now := time.Now()
		processed = &now
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE jobs SET status=$1, error=$2, processed_at=$3, attempts=attempts+1
		WHERE id=$4`, status, errMsg, processed, id)
	return err
}

func (r *Repository) JobRecordFailure(ctx context.Context, job model.Job, errMsg string) (bool, error) {
	nextAttempt := job.Attempts + 1
	if nextAttempt >= job.MaxAttempts {
		if err := r.JobUpdateStatus(ctx, job.ID, model.JobFailed, errMsg); err != nil {
			return false, err
		}
		return true, r.JobMoveToDead(ctx, job.ID)
	}

	runAt := time.Now().Add(jobBackoff(nextAttempt))
	_, err := r.db.ExecContext(ctx, `
		UPDATE jobs
		SET status=$1, error=$2, processed_at=NULL, attempts=attempts+1, run_at=$3
		WHERE id=$4`, model.JobPending, errMsg, runAt, job.ID)
	return false, err
}

func jobBackoff(attempt int) time.Duration {
	switch attempt {
	case 1:
		return time.Minute
	case 2:
		return 5 * time.Minute
	default:
		return 15 * time.Minute
	}
}

func (r *Repository) JobMoveToDead(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO dead_jobs (tenant_id, type, payload, status, run_at, created_at, attempts, max_attempts, error)
		SELECT tenant_id, type, payload, 'failed', run_at, created_at, attempts, max_attempts, error FROM jobs WHERE id = $1`, id)
	if err != nil {
		return err
	}
	_, err = r.db.ExecContext(ctx, `DELETE FROM jobs WHERE id = $1`, id)
	return err
}

func (r *Repository) JobCleanupCompleted(ctx context.Context, days int) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM jobs WHERE status = 'completed' AND processed_at < datetime('now', $1)`,
		fmt.Sprintf("-%d days", days))
	return err
}

func (r *Repository) DeadJobCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM dead_jobs`).Scan(&count)
	return count, err
}

func (r *Repository) DeadJobList(ctx context.Context, limit int) ([]model.Job, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT tenant_id, id, type, payload, status, run_at, created_at, attempts, max_attempts, error, processed_at
		FROM dead_jobs
		ORDER BY created_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var jobs []model.Job
	for rows.Next() {
		var job model.Job
		if err := rows.Scan(&job.TenantID, &job.ID, &job.Type, &job.Payload, &job.Status, &job.RunAt, &job.CreatedAt, &job.Attempts, &job.MaxAttempts, &job.Error, &job.ProcessedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

// Metrics

func (r *Repository) MetricTrack(ctx context.Context, m *model.Metric) error {
	if m.TenantID <= 0 {
		m.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO metrics (tenant_id, metric_type, entity_type, entity_id, user_id, ip_address, user_agent, referrer)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		m.TenantID, m.MetricType, m.EntityType, m.EntityID, m.UserID, m.IPAddress, m.UserAgent, m.Referrer)
	return err
}

func (r *Repository) MetricCountByEntity(ctx context.Context, metricType, entityType string, entityID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM metrics 
		WHERE tenant_id = $1 AND metric_type = $2 AND entity_type = $3 AND entity_id = $4`, tenantIDFromContext(ctx), metricType, entityType, entityID).Scan(&count)
	return count, err
}

func (r *Repository) MetricCountByType(ctx context.Context, metricType string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM metrics WHERE tenant_id = $1 AND metric_type = $2`, tenantIDFromContext(ctx), metricType).Scan(&count)
	return count, err
}

func (r *Repository) MetricTotals(ctx context.Context, limit int) ([]model.MetricTotal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT metric_type, COUNT(*) AS total
		FROM metrics
		WHERE tenant_id = $1
		GROUP BY metric_type
		ORDER BY total DESC, metric_type ASC
		LIMIT $2`, tenantIDFromContext(ctx), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totals []model.MetricTotal
	for rows.Next() {
		var total model.MetricTotal
		if err := rows.Scan(&total.MetricType, &total.Total); err != nil {
			return nil, err
		}
		totals = append(totals, total)
	}
	return totals, rows.Err()
}

func (r *Repository) MetricTopEntities(ctx context.Context, metricType string, limit int) ([]model.MetricEntityTotal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT metric_type, entity_type, entity_id, COUNT(*) AS total
		FROM metrics
		WHERE tenant_id = $1 AND metric_type = $2
		GROUP BY metric_type, entity_type, entity_id
		ORDER BY total DESC, entity_id ASC
		LIMIT $3`, tenantIDFromContext(ctx), metricType, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var totals []model.MetricEntityTotal
	for rows.Next() {
		var total model.MetricEntityTotal
		if err := rows.Scan(&total.MetricType, &total.EntityType, &total.EntityID, &total.Total); err != nil {
			return nil, err
		}
		totals = append(totals, total)
	}
	return totals, rows.Err()
}

// Audit Logs

func (r *Repository) AuditLogCreate(ctx context.Context, a *model.AuditLog) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, changes, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		a.UserID, a.Action, a.EntityType, a.EntityID, a.Changes, a.IPAddress, a.UserAgent)
	return err
}

func (r *Repository) AuditLogList(ctx context.Context, entityType string, entityID int64, limit int) ([]model.AuditLog, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, action, entity_type, entity_id, changes, ip_address, user_agent, created_at
		FROM audit_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY created_at DESC, id DESC
		LIMIT $3`, entityType, entityID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var a model.AuditLog
		if err := rows.Scan(&a.ID, &a.UserID, &a.Action, &a.EntityType, &a.EntityID, &a.Changes, &a.IPAddress, &a.UserAgent, &a.CreatedAt); err != nil {
			return nil, err
		}
		logs = append(logs, a)
	}
	return logs, rows.Err()
}

func (r *Repository) AuditLogSearch(ctx context.Context, filter AuditLogFilter) ([]model.AuditLogEntry, int, error) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	where, args := auditLogWhere(filter)
	countQuery := "SELECT COUNT(*) FROM audit_logs a " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT a.id, a.user_id, a.action, a.entity_type, a.entity_id, a.changes, a.ip_address, a.user_agent, a.created_at,
		       COALESCE(u.name, ''), COALESCE(u.email, '')
		FROM audit_logs a
		LEFT JOIN users u ON u.id = a.user_id
	` + where + `
		ORDER BY a.created_at DESC, a.id DESC
		LIMIT $` + strconv.Itoa(len(args)+1) + ` OFFSET $` + strconv.Itoa(len(args)+2)
	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var logs []model.AuditLogEntry
	for rows.Next() {
		var entry model.AuditLogEntry
		if err := rows.Scan(
			&entry.ID,
			&entry.UserID,
			&entry.Action,
			&entry.EntityType,
			&entry.EntityID,
			&entry.Changes,
			&entry.IPAddress,
			&entry.UserAgent,
			&entry.CreatedAt,
			&entry.UserName,
			&entry.UserEmail,
		); err != nil {
			return nil, 0, err
		}
		logs = append(logs, entry)
	}
	return logs, total, rows.Err()
}

func auditLogWhere(filter AuditLogFilter) (string, []any) {
	clauses := []string{"1=1"}
	args := []any{}
	add := func(clause string, value any) {
		args = append(args, value)
		clauses = append(clauses, fmt.Sprintf(clause, len(args)))
	}
	if filter.UserID > 0 {
		add("a.user_id = $%d", filter.UserID)
	}
	if filter.Action != "" {
		add("a.action = $%d", filter.Action)
	}
	if filter.EntityType != "" {
		add("a.entity_type = $%d", filter.EntityType)
	}
	if filter.EntityID > 0 {
		add("a.entity_id = $%d", filter.EntityID)
	}
	if filter.DateFrom != nil {
		add("a.created_at >= $%d", *filter.DateFrom)
	}
	if filter.DateTo != nil {
		add("a.created_at < $%d", *filter.DateTo)
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

// Login Attempts

func (r *Repository) LoginAttemptCreate(ctx context.Context, l *model.LoginAttempt) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO login_attempts (ip_address, email, success)
		VALUES ($1, $2, $3)`, l.IPAddress, l.Email, l.Success)
	return err
}

func (r *Repository) LoginAttemptCountRecent(ctx context.Context, ip string, minutes int) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM login_attempts 
		WHERE ip_address = $1 AND success = false AND created_at > datetime('now', $2)`,
		ip, fmt.Sprintf("-%d minutes", minutes)).Scan(&count)
	return count, err
}

// Edit Locks

func (r *Repository) EditLockCreate(ctx context.Context, l *model.EditLock) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO edit_locks (entity_type, entity_id, user_id, locked_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT(entity_type, entity_id) DO UPDATE SET user_id=$2, locked_at=$4, expires_at=$5`,
		l.EntityType, l.EntityID, l.UserID, l.LockedAt, l.ExpiresAt)
	return err
}

func (r *Repository) EditLockGet(ctx context.Context, entityType string, entityID int64) (*model.EditLock, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT entity_type, entity_id, user_id, locked_at, expires_at
		FROM edit_locks WHERE entity_type = $1 AND entity_id = $2`, entityType, entityID)
	var l model.EditLock
	err := row.Scan(&l.EntityType, &l.EntityID, &l.UserID, &l.LockedAt, &l.ExpiresAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &l, err
}

func (r *Repository) EditLockDelete(ctx context.Context, entityType string, entityID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM edit_locks WHERE entity_type = $1 AND entity_id = $2`, entityType, entityID)
	return err
}

func (r *Repository) SitemapEntries(ctx context.Context) ([]model.SitemapEntry, error) {
	var entries []model.SitemapEntry
	tenantID := tenantIDFromContext(ctx)
	queries := []struct {
		prefix string
		sql    string
		args   []any
	}{
		{
			prefix: "/noticia/",
			sql:    `SELECT slug, updated_at FROM posts WHERE tenant_id = $1 AND status = 'published' ORDER BY updated_at DESC`,
			args:   []any{tenantID},
		},
		{
			prefix: "/loja/",
			sql:    `SELECT slug, created_at FROM stores WHERE tenant_id = $1 AND active = true ORDER BY created_at DESC`,
			args:   []any{tenantID},
		},
		{
			prefix: "/promocao/",
			sql: `SELECT slug, created_at FROM promotions
				WHERE tenant_id = $1 AND status = 'active' AND start_date <= date('now') AND end_date >= date('now')
				ORDER BY created_at DESC`,
			args: []any{tenantID},
		},
		{
			prefix: "/evento/",
			sql:    `SELECT slug, updated_at FROM events WHERE tenant_id = $1 AND status = 'active' ORDER BY start_at ASC`,
			args:   []any{tenantID},
		},
		{
			prefix: "/classificado/",
			sql:    `SELECT slug, updated_at FROM classifieds WHERE tenant_id = $1 AND status = 'active' AND (expires_at IS NULL OR expires_at >= date('now')) ORDER BY created_at DESC`,
			args:   []any{tenantID},
		},
		{
			prefix: "/bairro/",
			sql:    `SELECT slug, created_at FROM neighborhoods WHERE tenant_id = $1 ORDER BY name`,
			args:   []any{tenantID},
		},
		{
			prefix: "/influencer/",
			sql:    `SELECT slug, created_at FROM influencers WHERE tenant_id = $1 AND active = true ORDER BY created_at DESC`,
			args:   []any{tenantID},
		},
		{
			prefix: "/tag/",
			sql:    `SELECT slug, created_at FROM tags WHERE tenant_id = $1 AND active = true ORDER BY name`,
			args:   []any{tenantID},
		},
	}

	for _, query := range queries {
		rows, err := r.db.QueryContext(ctx, query.sql, query.args...)
		if err != nil {
			return nil, err
		}
		for rows.Next() {
			var slug string
			var lastMod time.Time
			if err := rows.Scan(&slug, &lastMod); err != nil {
				rows.Close()
				return nil, err
			}
			entries = append(entries, model.SitemapEntry{
				Path:    query.prefix + slug,
				LastMod: lastMod,
			})
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return nil, err
		}
		rows.Close()
	}

	return entries, nil
}
