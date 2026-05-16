CREATE TYPE job_status AS ENUM ('pending','running','completed','failed');
CREATE TYPE job_type AS ENUM ('publish_post','expire_promotion','expire_banner','backup_database','vacuum_db','generate_sitemap','cleanup_old_jobs','compress_old_uploads','collect_news');
CREATE TYPE user_role AS ENUM ('super_admin','admin','editor','redator','revisor','comercial');
CREATE TYPE post_status AS ENUM ('draft','review','approved','scheduled','published','archived');

CREATE TABLE users (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    email VARCHAR(200) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    role user_role DEFAULT 'editor',
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE password_reset_tokens (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id),
    token_hash VARCHAR(128) UNIQUE NOT NULL,
    requested_ip INET,
    user_agent TEXT,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_password_reset_hash ON password_reset_tokens(token_hash);
CREATE INDEX idx_password_reset_user ON password_reset_tokens(user_id, created_at DESC);

CREATE TABLE categories (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    meta_title VARCHAR(200),
    meta_description VARCHAR(300),
    image_key VARCHAR(500),
    sort_order INTEGER DEFAULT 0,
    active BOOLEAN DEFAULT true,
    requires_editorial_notes BOOLEAN DEFAULT false
);

INSERT INTO categories (slug, name, requires_editorial_notes) VALUES
('noticias','Noticias',false),
('politica-bastidores','Politica & Bastidores',true),
('influencers','Influencers da Cidade',false),
('eventos','Eventos',false)
ON CONFLICT (slug) DO NOTHING;

CREATE TABLE tags (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(120) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    meta_title VARCHAR(200),
    meta_description VARCHAR(300),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    title VARCHAR(300) NOT NULL,
    slug VARCHAR(300) UNIQUE NOT NULL,
    excerpt TEXT,
    content TEXT NOT NULL,
    cover_image_key VARCHAR(500),
    gallery_image_keys JSONB DEFAULT '[]'::jsonb,
    meta_title VARCHAR(200),
    meta_description VARCHAR(300),
    seo_keyword VARCHAR(200),
    canonical_url VARCHAR(500),
    source_name VARCHAR(200),
    source_url VARCHAR(500),
    reading_time_minutes INTEGER DEFAULT 1,
    category_id INT REFERENCES categories(id),
    author_id BIGINT REFERENCES users(id),
    status post_status DEFAULT 'draft',
    is_sponsored BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    is_pinned BOOLEAN DEFAULT false,
    editorial_notes TEXT,
    editor_responsible VARCHAR(200),
    published_at TIMESTAMP,
    publish_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    search_vector tsvector
);

CREATE INDEX idx_posts_status ON posts(status, published_at DESC);
CREATE INDEX idx_posts_category ON posts(category_id, status);
CREATE INDEX idx_posts_featured ON posts(is_featured, status, updated_at DESC);
CREATE INDEX idx_posts_search ON posts USING GIN(search_vector);

CREATE TABLE post_revisions (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id),
    action VARCHAR(80) NOT NULL,
    title VARCHAR(300),
    status post_status,
    snapshot JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_post_revisions_post ON post_revisions(post_id, created_at DESC);

CREATE TABLE media_assets (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(500) UNIQUE NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    title VARCHAR(255),
    alt_text VARCHAR(255),
    content_type VARCHAR(120) NOT NULL,
    size_bytes BIGINT DEFAULT 0,
    uploaded_by BIGINT REFERENCES users(id),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_media_assets_created ON media_assets(created_at DESC);
CREATE INDEX idx_media_assets_search ON media_assets(original_name, title, alt_text);

CREATE TABLE portal_settings (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    site_name VARCHAR(200) NOT NULL,
    tagline TEXT,
    logo_key VARCHAR(500),
    favicon_key VARCHAR(500),
    contact_email VARCHAR(200),
    contact_whatsapp VARCHAR(60),
    contact_phone VARCHAR(60),
    city VARCHAR(120),
    state VARCHAR(2),
    seo_title VARCHAR(200),
    seo_description VARCHAR(300),
    facebook_url TEXT,
    instagram_url TEXT,
    youtube_url TEXT,
    tiktok_url TEXT,
    upload_max_mb INTEGER DEFAULT 2,
    automation_enabled BOOLEAN DEFAULT false,
    automation_interval_minutes INTEGER DEFAULT 60,
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE automation_sources (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(200) NOT NULL,
    source_type VARCHAR(40) NOT NULL DEFAULT 'rss',
    url TEXT NOT NULL,
    default_category_id INT REFERENCES categories(id),
    active BOOLEAN DEFAULT true,
    last_run_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_automation_sources_active ON automation_sources(active, source_type);

CREATE TABLE automation_runs (
    id BIGSERIAL PRIMARY KEY,
    source_id BIGINT REFERENCES automation_sources(id),
    status VARCHAR(40) NOT NULL DEFAULT 'success',
    items_found INTEGER DEFAULT 0,
    drafts_created INTEGER DEFAULT 0,
    duplicates INTEGER DEFAULT 0,
    error TEXT,
    log TEXT,
    started_at TIMESTAMP DEFAULT NOW(),
    finished_at TIMESTAMP
);
CREATE INDEX idx_automation_runs_started ON automation_runs(started_at DESC);

CREATE TABLE ai_usage_logs (
    id BIGSERIAL PRIMARY KEY,
    post_id BIGINT REFERENCES posts(id) ON DELETE SET NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(80) NOT NULL,
    provider VARCHAR(80) NOT NULL,
    input_title VARCHAR(300),
    output JSONB NOT NULL DEFAULT '{}'::jsonb,
    source_name VARCHAR(200),
    source_url TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_ai_usage_post ON ai_usage_logs(post_id, created_at DESC);
CREATE INDEX idx_ai_usage_created ON ai_usage_logs(created_at DESC);

CREATE TABLE post_tags (
    post_id BIGINT NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
    tag_id BIGINT NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (post_id, tag_id)
);
CREATE INDEX idx_post_tags_tag ON post_tags(tag_id, post_id);

CREATE OR REPLACE FUNCTION posts_search_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('portuguese', COALESCE(NEW.title,'')), 'A') ||
        setweight(to_tsvector('portuguese', COALESCE(NEW.excerpt,'')), 'B') ||
        setweight(to_tsvector('portuguese', COALESCE(NEW.content,'')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER posts_search_trigger BEFORE INSERT OR UPDATE ON posts
    FOR EACH ROW EXECUTE FUNCTION posts_search_update();

CREATE TABLE slug_redirects (
    old_slug VARCHAR(300) PRIMARY KEY,
    new_slug VARCHAR(300) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE neighborhoods (
    id SERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    meta_title VARCHAR(60),
    meta_description VARCHAR(160),
    cover_image_key VARCHAR(500),
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE influencers (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(200) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    bio TEXT,
    city_area VARCHAR(200),
    niche VARCHAR(120),
    instagram TEXT,
    tiktok TEXT,
    youtube TEXT,
    whatsapp VARCHAR(40),
    avatar_key TEXT,
    cover_image_key TEXT,
    meta_title VARCHAR(200),
    meta_description VARCHAR(300),
    is_featured BOOLEAN DEFAULT false,
    is_sponsored BOOLEAN DEFAULT false,
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE stores (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(200) UNIQUE NOT NULL,
    name VARCHAR(200) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    address TEXT,
    phone VARCHAR(50),
    whatsapp VARCHAR(50),
    website_url TEXT,
    logo_key VARCHAR(500),
    cover_image_key VARCHAR(500),
    commercial_status VARCHAR(50) DEFAULT 'active',
    meta_title VARCHAR(200),
    meta_description VARCHAR(300),
    is_sponsored BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    neighborhood_id INT REFERENCES neighborhoods(id),
    active BOOLEAN DEFAULT true,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE promotions (
    id BIGSERIAL PRIMARY KEY,
    store_id BIGINT REFERENCES stores(id),
    title VARCHAR(300) NOT NULL,
    slug VARCHAR(300) UNIQUE NOT NULL,
    description TEXT,
    price_display VARCHAR(100),
    coupon_code VARCHAR(80),
    image_key VARCHAR(500),
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(50) DEFAULT 'active',
    is_sponsored BOOLEAN DEFAULT false,
    meta_title VARCHAR(200),
    meta_description VARCHAR(300),
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_promotions_active ON promotions(store_id, status, end_date) WHERE status = 'active';

CREATE TABLE events (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(300) UNIQUE NOT NULL,
    title VARCHAR(300) NOT NULL,
    description TEXT,
    location VARCHAR(300),
    organizer VARCHAR(200),
    ticket_url TEXT,
    price_display VARCHAR(120),
    image_key VARCHAR(500),
    status VARCHAR(50) DEFAULT 'active',
    is_featured BOOLEAN DEFAULT false,
    is_sponsored BOOLEAN DEFAULT false,
    meta_title VARCHAR(200),
    meta_description VARCHAR(300),
    start_at TIMESTAMP NOT NULL,
    end_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_events_status_start ON events(status, start_at);
CREATE INDEX idx_events_featured ON events(is_featured, status, start_at);

CREATE TABLE classifieds (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(300) UNIQUE NOT NULL,
    title VARCHAR(300) NOT NULL,
    description TEXT,
    category VARCHAR(120),
    price_display VARCHAR(120),
    contact_name VARCHAR(200),
    contact_phone VARCHAR(60),
    contact_whatsapp VARCHAR(60),
    location VARCHAR(200),
    image_key VARCHAR(500),
    status VARCHAR(50) DEFAULT 'active',
    is_featured BOOLEAN DEFAULT false,
    is_sponsored BOOLEAN DEFAULT false,
    meta_title VARCHAR(200),
    meta_description VARCHAR(300),
    expires_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_classifieds_status_category ON classifieds(status, category, created_at DESC);
CREATE INDEX idx_classifieds_featured ON classifieds(is_featured, status, created_at DESC);

CREATE TABLE banners (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    advertiser_name VARCHAR(200),
    contact_name VARCHAR(200),
    contact_phone VARCHAR(60),
    contact_whatsapp VARCHAR(60),
    price_display VARCHAR(120),
    notes TEXT,
    position VARCHAR(50) NOT NULL,
    image_key VARCHAR(500) NOT NULL,
    link_url TEXT NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status VARCHAR(40) DEFAULT 'active',
    active BOOLEAN DEFAULT true,
    priority INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_banners_position ON banners(position, active, start_date, end_date);

CREATE TABLE jobs (
    id BIGSERIAL PRIMARY KEY,
    type job_type NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status job_status DEFAULT 'pending',
    run_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    error TEXT,
    processed_at TIMESTAMP WITH TIME ZONE
);
CREATE INDEX idx_jobs_pending ON jobs(run_at) WHERE status = 'pending';

CREATE TABLE dead_jobs (
    id BIGSERIAL PRIMARY KEY,
    type job_type NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status job_status DEFAULT 'failed',
    run_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    attempts INT DEFAULT 0,
    max_attempts INT DEFAULT 3,
    error TEXT,
    processed_at TIMESTAMP WITH TIME ZONE
);

CREATE TABLE metrics (
    id BIGSERIAL PRIMARY KEY,
    metric_type VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    user_id BIGINT REFERENCES users(id),
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_metrics_entity ON metrics(entity_type, entity_id, metric_type);
CREATE INDEX idx_metrics_date ON metrics(created_at);

CREATE TABLE audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    action VARCHAR(50) NOT NULL,
    entity_type VARCHAR(50),
    entity_id BIGINT,
    changes JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE edit_locks (
    entity_type VARCHAR(50) NOT NULL,
    entity_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL REFERENCES users(id),
    locked_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    PRIMARY KEY (entity_type, entity_id)
);

CREATE TABLE login_attempts (
    id BIGSERIAL PRIMARY KEY,
    ip_address INET NOT NULL,
    email VARCHAR(200),
    success BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_login_ip ON login_attempts(ip_address, created_at DESC);
