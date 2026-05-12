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

	_ "modernc.org/sqlite"
)

type Repository struct {
	db *sql.DB
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
	db, err := sql.Open("sqlite", dbPath+"?_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_pragma=cache_size(-64000)&_pragma=temp_store(memory)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(30 * time.Minute)

	repo := &Repository{db: db}
	if err := repo.Migrate(); err != nil {
		return nil, fmt.Errorf("migrate: %w", err)
	}
	return repo, nil
}

func (r *Repository) DB() *sql.DB {
	return r.db
}

func (r *Repository) Migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    role TEXT DEFAULT 'editor',
    active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    meta_title TEXT,
    meta_description TEXT,
    image_key TEXT,
    sort_order INTEGER DEFAULT 0,
    active BOOLEAN DEFAULT true,
    requires_editorial_notes BOOLEAN DEFAULT false
);

CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    meta_title TEXT,
    meta_description TEXT,
    active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS posts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
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
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
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
    key TEXT UNIQUE NOT NULL,
    original_name TEXT NOT NULL,
    title TEXT,
    alt_text TEXT,
    content_type TEXT NOT NULL,
    size_bytes INTEGER DEFAULT 0,
    uploaded_by INTEGER REFERENCES users(id),
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_media_assets_created ON media_assets(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_assets_search ON media_assets(original_name, title, alt_text);

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
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    category TEXT,
    address TEXT,
    phone TEXT,
    whatsapp TEXT,
    logo_key TEXT,
    cover_image_key TEXT,
    is_sponsored BOOLEAN DEFAULT false,
    is_featured BOOLEAN DEFAULT false,
    neighborhood_id INTEGER,
    active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS promotions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    store_id INTEGER REFERENCES stores(id),
    title TEXT NOT NULL,
    slug TEXT UNIQUE NOT NULL,
    description TEXT,
    price_display TEXT,
    image_key TEXT,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    status TEXT DEFAULT 'active',
    is_sponsored BOOLEAN DEFAULT false,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS banners (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    description TEXT,
    meta_title TEXT,
    meta_description TEXT,
    cover_image_key TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS influencers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    slug TEXT UNIQUE NOT NULL,
    name TEXT NOT NULL,
    bio TEXT,
    city_area TEXT,
    instagram TEXT,
    tiktok TEXT,
    youtube TEXT,
    whatsapp TEXT,
    avatar_key TEXT,
    cover_image_key TEXT,
    is_featured BOOLEAN DEFAULT false,
    is_sponsored BOOLEAN DEFAULT false,
    active BOOLEAN DEFAULT true,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS jobs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
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

INSERT OR IGNORE INTO categories (slug, name, requires_editorial_notes) VALUES
('noticias', 'Noticias', false),
('politica-bastidores', 'Politica & Bastidores', true),
('influencers', 'Influencers da Cidade', false),
('eventos', 'Eventos', false);
`
	if _, err := r.db.Exec(schema); err != nil {
		return err
	}
	return r.ensureSQLiteColumns()
}

func (r *Repository) ensureSQLiteColumns() error {
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
	return nil
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

func (r *Repository) UserGetByEmail(ctx context.Context, email string) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, email, password_hash, role, active, created_at FROM users WHERE lower(email) = lower($1)`, email)
	var u model.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Active, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("usuario nao encontrado")
	}
	return &u, err
}

func (r *Repository) UserGetByID(ctx context.Context, id int64) (*model.User, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, email, password_hash, role, active, created_at FROM users WHERE id = $1`, id)
	var u model.User
	err := row.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Active, &u.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("usuario nao encontrado")
	}
	return &u, err
}

func (r *Repository) UserCreate(ctx context.Context, u *model.User) error {
	res, err := r.db.ExecContext(ctx, `INSERT INTO users (name, email, password_hash, role, active) VALUES ($1, $2, $3, $4, $5)`,
		u.Name, u.Email, u.PasswordHash, u.Role, u.Active)
	if err != nil {
		return err
	}
	u.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) UserUpdatePassword(ctx context.Context, id int64, hash string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE users SET password_hash = $1 WHERE id = $2`, hash, id)
	return err
}

func (r *Repository) UserUpdate(ctx context.Context, u *model.User) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET name = $1, email = $2, role = $3, active = $4
		WHERE id = $5`,
		u.Name, u.Email, u.Role, u.Active, u.ID)
	return err
}

func (r *Repository) UserList(ctx context.Context) ([]model.User, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, email, password_hash, role, active, created_at FROM users ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []model.User
	for rows.Next() {
		var u model.User
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.Role, &u.Active, &u.CreatedAt); err != nil {
			return nil, err
		}
		users = append(users, u)
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
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO password_reset_tokens (user_id, token_hash, requested_ip, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5)`,
		token.UserID, token.TokenHash, token.RequestedIP, token.UserAgent, token.ExpiresAt)
	if err != nil {
		return err
	}
	token.ID, _ = res.LastInsertId()
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
	row := r.db.QueryRowContext(ctx, categorySelectSQL()+` WHERE slug = $1`, slug)
	var c model.Category
	err := scanCategory(row, &c)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *Repository) CategoryGetByID(ctx context.Context, id int64) (*model.Category, error) {
	row := r.db.QueryRowContext(ctx, categorySelectSQL()+` WHERE id = $1`, id)
	var c model.Category
	err := scanCategory(row, &c)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &c, err
}

func (r *Repository) CategoryList(ctx context.Context) ([]model.Category, error) {
	rows, err := r.db.QueryContext(ctx, categorySelectSQL()+` ORDER BY sort_order ASC, name ASC`)
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
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO categories (slug, name, description, meta_title, meta_description, image_key, sort_order, active, requires_editorial_notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		c.Slug, c.Name, c.Description, c.MetaTitle, c.MetaDescription, c.ImageKey, c.SortOrder, c.Active, c.RequiresEditorialNotes)
	if err != nil {
		return err
	}
	c.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) CategoryUpdate(ctx context.Context, c *model.Category) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE categories
		SET slug=$1, name=$2, description=$3, meta_title=$4, meta_description=$5, image_key=$6, sort_order=$7, active=$8, requires_editorial_notes=$9
		WHERE id=$10`,
		c.Slug, c.Name, c.Description, c.MetaTitle, c.MetaDescription, c.ImageKey, c.SortOrder, c.Active, c.RequiresEditorialNotes, c.ID)
	return err
}

func (r *Repository) CategoryDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	return err
}

func (r *Repository) CategoryPostCount(ctx context.Context, id int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE category_id = $1`, id).Scan(&count)
	return count, err
}

func (r *Repository) CategorySlugExists(ctx context.Context, slug string) bool {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM categories WHERE slug = $1`, slug).Scan(&id)
	return err == nil
}

func categorySelectSQL() string {
	return `SELECT id, slug, name, COALESCE(description, ''), COALESCE(meta_title, ''), COALESCE(meta_description, ''), COALESCE(image_key, ''), COALESCE(sort_order, 0), COALESCE(active, true), requires_editorial_notes FROM categories`
}

type categoryScanner interface {
	Scan(dest ...any) error
}

func scanCategory(scanner categoryScanner, c *model.Category) error {
	return scanner.Scan(&c.ID, &c.Slug, &c.Name, &c.Description, &c.MetaTitle, &c.MetaDescription, &c.ImageKey, &c.SortOrder, &c.Active, &c.RequiresEditorialNotes)
}

// Tags

func (r *Repository) TagGetBySlug(ctx context.Context, slug string) (*model.Tag, error) {
	row := r.db.QueryRowContext(ctx, tagSelectSQL()+` WHERE slug = $1`, slug)
	var tag model.Tag
	err := scanTag(row, &tag)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tag, err
}

func (r *Repository) TagGetByID(ctx context.Context, id int64) (*model.Tag, error) {
	row := r.db.QueryRowContext(ctx, tagSelectSQL()+` WHERE id = $1`, id)
	var tag model.Tag
	err := scanTag(row, &tag)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &tag, err
}

func (r *Repository) TagList(ctx context.Context, activeOnly bool) ([]model.Tag, error) {
	query := tagSelectSQL()
	if activeOnly {
		query += ` WHERE active = true`
	}
	query += ` ORDER BY name ASC`
	rows, err := r.db.QueryContext(ctx, query)
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
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO tags (slug, name, description, meta_title, meta_description, active)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		tag.Slug, tag.Name, tag.Description, tag.MetaTitle, tag.MetaDescription, tag.Active)
	if err != nil {
		return err
	}
	tag.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) TagUpdate(ctx context.Context, tag *model.Tag) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE tags
		SET slug=$1, name=$2, description=$3, meta_title=$4, meta_description=$5, active=$6
		WHERE id=$7`,
		tag.Slug, tag.Name, tag.Description, tag.MetaTitle, tag.MetaDescription, tag.Active, tag.ID)
	return err
}

func (r *Repository) TagDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tags WHERE id = $1`, id)
	return err
}

func (r *Repository) TagPostCount(ctx context.Context, id int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM post_tags WHERE tag_id = $1`, id).Scan(&count)
	return count, err
}

func (r *Repository) TagSlugExists(ctx context.Context, slug string) bool {
	var id int64
	err := r.db.QueryRowContext(ctx, `SELECT id FROM tags WHERE slug = $1`, slug).Scan(&id)
	return err == nil
}

func (r *Repository) PostTags(ctx context.Context, postID int64) ([]model.Tag, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.slug, t.name, COALESCE(t.description, ''), COALESCE(t.meta_title, ''), COALESCE(t.meta_description, ''), COALESCE(t.active, true), t.created_at
		FROM tags t
		JOIN post_tags pt ON pt.tag_id = t.id
		WHERE pt.post_id = $1
		ORDER BY t.name ASC`, postID)
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
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM post_tags WHERE post_id = $1`, postID); err != nil {
		return err
	}
	seen := map[int64]bool{}
	for _, tagID := range tagIDs {
		if tagID <= 0 || seen[tagID] {
			continue
		}
		seen[tagID] = true
		if _, err := tx.ExecContext(ctx, `INSERT OR IGNORE INTO post_tags (post_id, tag_id) VALUES ($1, $2)`, postID, tagID); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func tagSelectSQL() string {
	return `SELECT id, slug, name, COALESCE(description, ''), COALESCE(meta_title, ''), COALESCE(meta_description, ''), COALESCE(active, true), created_at FROM tags`
}

type tagScanner interface {
	Scan(dest ...any) error
}

func scanTag(scanner tagScanner, tag *model.Tag) error {
	return scanner.Scan(&tag.ID, &tag.Slug, &tag.Name, &tag.Description, &tag.MetaTitle, &tag.MetaDescription, &tag.Active, &tag.CreatedAt)
}

// Posts

func (r *Repository) PostCreate(ctx context.Context, p *model.Post) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO posts (title, slug, excerpt, content, cover_image_key, gallery_image_keys, meta_title, meta_description, seo_keyword, canonical_url, source_name, source_url, reading_time_minutes, category_id, author_id, status, is_sponsored, is_featured, is_pinned, editorial_notes, editor_responsible, published_at, publish_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23)`,
		p.Title, p.Slug, p.Excerpt, p.Content, p.CoverImageKey, encodeStringList(p.GalleryImageKeys), p.MetaTitle, p.MetaDescription, p.SEOKeyword, p.CanonicalURL, p.SourceName, p.SourceURL, p.ReadingTimeMinutes, p.CategoryID, p.AuthorID, p.Status, p.IsSponsored, p.IsFeatured, p.IsPinned, p.EditorialNotes, p.EditorResponsible, p.PublishedAt, p.PublishAt)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) PostUpdate(ctx context.Context, p *model.Post) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE posts SET title=$1, slug=$2, excerpt=$3, content=$4, cover_image_key=$5, gallery_image_keys=$6, meta_title=$7, meta_description=$8, seo_keyword=$9, canonical_url=$10, source_name=$11, source_url=$12, reading_time_minutes=$13, category_id=$14, status=$15, is_sponsored=$16, is_featured=$17, is_pinned=$18, editorial_notes=$19, editor_responsible=$20, published_at=$21, publish_at=$22, updated_at=CURRENT_TIMESTAMP
		WHERE id=$23`,
		p.Title, p.Slug, p.Excerpt, p.Content, p.CoverImageKey, encodeStringList(p.GalleryImageKeys), p.MetaTitle, p.MetaDescription, p.SEOKeyword, p.CanonicalURL, p.SourceName, p.SourceURL, p.ReadingTimeMinutes, p.CategoryID, p.Status, p.IsSponsored, p.IsFeatured, p.IsPinned, p.EditorialNotes, p.EditorResponsible, p.PublishedAt, p.PublishAt, p.ID)
	return err
}

func (r *Repository) PostGetBySlug(ctx context.Context, slug string) (*model.Post, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.title, p.slug, p.excerpt, p.content, p.cover_image_key, COALESCE(p.gallery_image_keys, '[]'), COALESCE(p.meta_title, ''), COALESCE(p.meta_description, ''), COALESCE(p.seo_keyword, ''), COALESCE(p.canonical_url, ''), COALESCE(p.source_name, ''), COALESCE(p.source_url, ''), COALESCE(p.reading_time_minutes, 1), p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), COALESCE(p.is_pinned, false), COALESCE(p.editorial_notes, ''), COALESCE(p.editor_responsible, ''), p.published_at, p.publish_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.slug = $1`, slug)
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
		SELECT p.id, p.title, p.slug, p.excerpt, p.content, p.cover_image_key, COALESCE(p.gallery_image_keys, '[]'), COALESCE(p.meta_title, ''), COALESCE(p.meta_description, ''), COALESCE(p.seo_keyword, ''), COALESCE(p.canonical_url, ''), COALESCE(p.source_name, ''), COALESCE(p.source_url, ''), COALESCE(p.reading_time_minutes, 1), p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), COALESCE(p.is_pinned, false), COALESCE(p.editorial_notes, ''), COALESCE(p.editor_responsible, ''), p.published_at, p.publish_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.id = $1`, id)
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
	if err := scanner.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.Content, &p.CoverImageKey, &galleryJSON, &p.MetaTitle, &p.MetaDescription, &p.SEOKeyword, &p.CanonicalURL, &p.SourceName, &p.SourceURL, &p.ReadingTimeMinutes, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.IsPinned, &p.EditorialNotes, &p.EditorResponsible, &p.PublishedAt, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
		return err
	}
	p.GalleryImageKeys = decodeStringList(galleryJSON)
	return nil
}

func (r *Repository) PostListPublished(ctx context.Context, limit int, offset int) ([]model.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.status = 'published'
		ORDER BY p.published_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *Repository) PostGetFeaturedPublished(ctx context.Context) (*model.Post, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), COALESCE(p.is_pinned, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.status = 'published' AND (COALESCE(p.is_featured, false) = true OR COALESCE(p.is_pinned, false) = true)
		ORDER BY COALESCE(p.is_pinned, false) DESC, p.updated_at DESC, p.published_at DESC
		LIMIT 1`)
	var p model.Post
	err := row.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.IsPinned, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *Repository) PostListByCategory(ctx context.Context, catID int64, limit int) ([]model.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.category_id = $1 AND p.status = 'published'
		ORDER BY p.published_at DESC
		LIMIT $2`, catID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *Repository) PostListAll(ctx context.Context, limit int, offset int) ([]model.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.publish_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		ORDER BY p.created_at DESC
		LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []model.Post
	for rows.Next() {
		var p model.Post
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.PublishedAt, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *Repository) PostListAdmin(ctx context.Context, status string, categoryID int64, tagID int64, query string, limit int, offset int) ([]model.Post, error) {
	where, args := adminPostWhere(status, categoryID, tagID, query)
	args = append(args, limit, offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.publish_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
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
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.PublishedAt, &p.PublishAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
			return nil, err
		}
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func (r *Repository) PostDelete(ctx context.Context, id int64) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM post_tags WHERE post_id = $1`, id); err != nil {
		return err
	}
	_, err := r.db.ExecContext(ctx, `DELETE FROM posts WHERE id = $1`, id)
	return err
}

func (r *Repository) PostCount(ctx context.Context) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts`).Scan(&count)
	return count, err
}

func (r *Repository) PostCountAdmin(ctx context.Context, status string, categoryID int64, tagID int64, query string) (int, error) {
	where, args := adminPostWhere(status, categoryID, tagID, query)
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts p `+where, args...).Scan(&count)
	return count, err
}

func adminPostWhere(status string, categoryID int64, tagID int64, query string) (string, []any) {
	var clauses []string
	var args []any
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
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE slug = $1`, slug).Scan(&count)
	return count > 0
}

func (r *Repository) PostSearch(ctx context.Context, query string, limit int) ([]model.Post, error) {
	ftsQuery := buildFTSQuery(query)
	if ftsQuery != "" {
		rows, err := r.db.QueryContext(ctx, `
			SELECT p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
			FROM posts_fts f
			JOIN posts p ON p.id = f.rowid
			LEFT JOIN categories c ON c.id = p.category_id
			LEFT JOIN users u ON u.id = p.author_id
			WHERE posts_fts MATCH $1 AND p.status = 'published'
			ORDER BY bm25(posts_fts), p.published_at DESC
			LIMIT $2`, ftsQuery, limit)
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
		SELECT p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE (
			p.title LIKE $1 OR p.excerpt LIKE $1 OR p.content LIKE $1 OR
			EXISTS (
				SELECT 1 FROM post_tags pt
				JOIN tags t ON t.id = pt.tag_id
				WHERE pt.post_id = p.id AND t.active = true AND t.name LIKE $1
			)
		) AND p.status = 'published'
		ORDER BY p.published_at DESC
		LIMIT $2`, q, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostList(rows)
}

func (r *Repository) PostListByTag(ctx context.Context, tagID int64, limit int) ([]model.Post, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		JOIN post_tags pt ON pt.post_id = p.id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE pt.tag_id = $1 AND p.status = 'published'
		ORDER BY p.published_at DESC
		LIMIT $2`, tagID, limit)
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
		if err := rows.Scan(&p.ID, &p.Title, &p.Slug, &p.Excerpt, &p.CoverImageKey, &p.CategoryID, &p.AuthorID, &p.Status, &p.IsSponsored, &p.IsFeatured, &p.PublishedAt, &p.CreatedAt, &p.UpdatedAt, &p.CategoryName, &p.AuthorName); err != nil {
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
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET status=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`, status, id)
	return err
}

func (r *Repository) PostUpdateStatusAndEditorialNotes(ctx context.Context, id int64, status model.PostStatus, notes string, responsible string) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE posts
		SET status=$1, editorial_notes=$2, editor_responsible=$3, updated_at=CURRENT_TIMESTAMP
		WHERE id=$4`, status, notes, responsible, id)
	return err
}

func (r *Repository) PostRevisionCreate(ctx context.Context, revision *model.PostRevision) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO post_revisions (post_id, user_id, action, title, status, snapshot)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		revision.PostID, revision.UserID, revision.Action, revision.Title, revision.Status, revision.Snapshot)
	if err != nil {
		return err
	}
	revision.ID, _ = res.LastInsertId()
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
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO media_assets (key, original_name, title, alt_text, content_type, size_bytes, uploaded_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		asset.Key, asset.OriginalName, asset.Title, asset.AltText, asset.ContentType, asset.SizeBytes, asset.UploadedBy)
	if err != nil {
		return err
	}
	asset.ID, _ = res.LastInsertId()
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
	where, args := mediaAssetWhere(filter)
	args = append(args, filter.Limit, filter.Offset)
	rows, err := r.db.QueryContext(ctx, `
		SELECT m.id, m.key, m.original_name, COALESCE(m.title, ''), COALESCE(m.alt_text, ''), m.content_type, COALESCE(m.size_bytes, 0), m.uploaded_by, COALESCE(u.name, ''), m.created_at, m.updated_at
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
	where, args := mediaAssetWhere(filter)
	row := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM media_assets m `+where, args...)
	var count int
	err := row.Scan(&count)
	return count, err
}

func (r *Repository) MediaAssetArchiveMonths(ctx context.Context) ([]model.MediaArchiveMonth, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT substr(CAST(created_at AS TEXT), 1, 7) AS month, COUNT(*)
		FROM media_assets
		GROUP BY month
		ORDER BY month DESC`)
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
		SELECT m.id, m.key, m.original_name, COALESCE(m.title, ''), COALESCE(m.alt_text, ''), m.content_type, COALESCE(m.size_bytes, 0), m.uploaded_by, COALESCE(u.name, ''), m.created_at, m.updated_at
		FROM media_assets m
		LEFT JOIN users u ON u.id = m.uploaded_by
		WHERE m.id = $1`, id)
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
		SELECT m.id, m.key, m.original_name, COALESCE(m.title, ''), COALESCE(m.alt_text, ''), m.content_type, COALESCE(m.size_bytes, 0), m.uploaded_by, COALESCE(u.name, ''), m.created_at, m.updated_at
		FROM media_assets m
		LEFT JOIN users u ON u.id = m.uploaded_by
		WHERE m.key = $1`, key)
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
	_, err := r.db.ExecContext(ctx, `
		UPDATE media_assets
		SET title=$1, alt_text=$2, updated_at=CURRENT_TIMESTAMP
		WHERE id=$3`, asset.Title, asset.AltText, asset.ID)
	return err
}

func (r *Repository) MediaAssetDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM media_assets WHERE id = $1`, id)
	return err
}

func (r *Repository) MediaAssetUsageCount(ctx context.Context, key string) (int, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total), 0) FROM (
			SELECT COUNT(*) AS total FROM posts WHERE cover_image_key = $1 OR COALESCE(gallery_image_keys, '') LIKE '%' || $1 || '%'
			UNION ALL SELECT COUNT(*) FROM categories WHERE image_key = $1
			UNION ALL SELECT COUNT(*) FROM stores WHERE logo_key = $1 OR cover_image_key = $1
			UNION ALL SELECT COUNT(*) FROM promotions WHERE image_key = $1
			UNION ALL SELECT COUNT(*) FROM banners WHERE image_key = $1
			UNION ALL SELECT COUNT(*) FROM neighborhoods WHERE cover_image_key = $1
			UNION ALL SELECT COUNT(*) FROM influencers WHERE avatar_key = $1 OR cover_image_key = $1
		)`, key)
	var count int
	err := row.Scan(&count)
	return count, err
}

func mediaAssetWhere(filter MediaAssetFilter) (string, []any) {
	var conditions []string
	var args []any
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
	if len(conditions) == 0 {
		return "", nil
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
	if err := scanner.Scan(&asset.ID, &asset.Key, &asset.OriginalName, &asset.Title, &asset.AltText, &asset.ContentType, &asset.SizeBytes, &uploadedBy, &asset.UploadedByName, &asset.CreatedAt, &asset.UpdatedAt); err != nil {
		return asset, err
	}
	if uploadedBy.Valid {
		id := uploadedBy.Int64
		asset.UploadedBy = &id
	}
	return asset, nil
}

func (r *Repository) PostSetPublished(ctx context.Context, id int64, t time.Time) error {
	_, err := r.db.ExecContext(ctx, `UPDATE posts SET status='published', published_at=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`, t, id)
	return err
}

func (r *Repository) PostSetFeatured(ctx context.Context, id int64, featured bool) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if featured {
		if _, err := tx.ExecContext(ctx, `UPDATE posts SET is_featured=false WHERE is_featured=true`); err != nil {
			return err
		}
	}
	if _, err := tx.ExecContext(ctx, `UPDATE posts SET is_featured=$1, updated_at=CURRENT_TIMESTAMP WHERE id=$2`, featured, id); err != nil {
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
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO stores (slug, name, description, category, address, phone, whatsapp, logo_key, cover_image_key, is_sponsored, is_featured, neighborhood_id, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		s.Slug, s.Name, s.Description, s.Category, s.Address, s.Phone, s.Whatsapp, s.LogoKey, s.CoverImageKey, s.IsSponsored, s.IsFeatured, s.NeighborhoodID, s.Active)
	if err != nil {
		return err
	}
	s.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) StoreUpdate(ctx context.Context, s *model.Store) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE stores SET slug=$1, name=$2, description=$3, category=$4, address=$5, phone=$6, whatsapp=$7, logo_key=$8, cover_image_key=$9, is_sponsored=$10, is_featured=$11, neighborhood_id=$12, active=$13
		WHERE id=$14`,
		s.Slug, s.Name, s.Description, s.Category, s.Address, s.Phone, s.Whatsapp, s.LogoKey, s.CoverImageKey, s.IsSponsored, s.IsFeatured, s.NeighborhoodID, s.Active, s.ID)
	return err
}

func (r *Repository) StoreGetBySlug(ctx context.Context, slug string) (*model.Store, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, name, description, category, address, phone, whatsapp, logo_key, cover_image_key, is_sponsored, is_featured, neighborhood_id, active, created_at FROM stores WHERE slug = $1`, slug)
	var s model.Store
	err := row.Scan(&s.ID, &s.Slug, &s.Name, &s.Description, &s.Category, &s.Address, &s.Phone, &s.Whatsapp, &s.LogoKey, &s.CoverImageKey, &s.IsSponsored, &s.IsFeatured, &s.NeighborhoodID, &s.Active, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *Repository) StoreGetByID(ctx context.Context, id int64) (*model.Store, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, name, description, category, address, phone, whatsapp, logo_key, cover_image_key, is_sponsored, is_featured, neighborhood_id, active, created_at FROM stores WHERE id = $1`, id)
	var s model.Store
	err := row.Scan(&s.ID, &s.Slug, &s.Name, &s.Description, &s.Category, &s.Address, &s.Phone, &s.Whatsapp, &s.LogoKey, &s.CoverImageKey, &s.IsSponsored, &s.IsFeatured, &s.NeighborhoodID, &s.Active, &s.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &s, err
}

func (r *Repository) StoreList(ctx context.Context, activeOnly bool, limit int) ([]model.Store, error) {
	q := `SELECT id, slug, name, description, category, address, phone, whatsapp, logo_key, cover_image_key, is_sponsored, is_featured, neighborhood_id, active, created_at FROM stores`
	var args []any
	if activeOnly {
		q += ` WHERE active = true`
	}
	q += ` ORDER BY created_at DESC LIMIT $1`
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stores []model.Store
	for rows.Next() {
		var s model.Store
		if err := rows.Scan(&s.ID, &s.Slug, &s.Name, &s.Description, &s.Category, &s.Address, &s.Phone, &s.Whatsapp, &s.LogoKey, &s.CoverImageKey, &s.IsSponsored, &s.IsFeatured, &s.NeighborhoodID, &s.Active, &s.CreatedAt); err != nil {
			return nil, err
		}
		stores = append(stores, s)
	}
	return stores, rows.Err()
}

func (r *Repository) StoreListFeatured(ctx context.Context, limit int) ([]model.Store, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, slug, name, description, category, address, phone, whatsapp, logo_key, cover_image_key, is_sponsored, is_featured, neighborhood_id, active, created_at 
		FROM stores WHERE active = true AND is_featured = true
		ORDER BY created_at DESC LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var stores []model.Store
	for rows.Next() {
		var s model.Store
		if err := rows.Scan(&s.ID, &s.Slug, &s.Name, &s.Description, &s.Category, &s.Address, &s.Phone, &s.Whatsapp, &s.LogoKey, &s.CoverImageKey, &s.IsSponsored, &s.IsFeatured, &s.NeighborhoodID, &s.Active, &s.CreatedAt); err != nil {
			return nil, err
		}
		stores = append(stores, s)
	}
	return stores, rows.Err()
}

func (r *Repository) StoreDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM stores WHERE id = $1`, id)
	return err
}

func (r *Repository) StoreSlugExists(ctx context.Context, slug string) bool {
	var count int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM stores WHERE slug = $1`, slug).Scan(&count)
	return count > 0
}

// Promotions

func (r *Repository) PromotionCreate(ctx context.Context, p *model.Promotion) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO promotions (store_id, title, slug, description, price_display, image_key, start_date, end_date, status, is_sponsored)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		p.StoreID, p.Title, p.Slug, p.Description, p.PriceDisplay, p.ImageKey, p.StartDate, p.EndDate, p.Status, p.IsSponsored)
	if err != nil {
		return err
	}
	p.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) PromotionUpdate(ctx context.Context, p *model.Promotion) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE promotions SET store_id=$1, title=$2, slug=$3, description=$4, price_display=$5, image_key=$6, start_date=$7, end_date=$8, status=$9, is_sponsored=$10
		WHERE id=$11`,
		p.StoreID, p.Title, p.Slug, p.Description, p.PriceDisplay, p.ImageKey, p.StartDate, p.EndDate, p.Status, p.IsSponsored, p.ID)
	return err
}

func (r *Repository) PromotionGetBySlug(ctx context.Context, slug string) (*model.Promotion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.store_id, p.title, p.slug, p.description, p.price_display, p.image_key, p.start_date, p.end_date, p.status, p.is_sponsored, p.created_at, s.name, s.slug
		FROM promotions p
		JOIN stores s ON s.id = p.store_id
		WHERE p.slug = $1`, slug)
	var p model.Promotion
	err := row.Scan(&p.ID, &p.StoreID, &p.Title, &p.Slug, &p.Description, &p.PriceDisplay, &p.ImageKey, &p.StartDate, &p.EndDate, &p.Status, &p.IsSponsored, &p.CreatedAt, &p.StoreName, &p.StoreSlug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *Repository) PromotionGetByID(ctx context.Context, id int64) (*model.Promotion, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT p.id, p.store_id, p.title, p.slug, p.description, p.price_display, p.image_key, p.start_date, p.end_date, p.status, p.is_sponsored, p.created_at, s.name, s.slug
		FROM promotions p
		JOIN stores s ON s.id = p.store_id
		WHERE p.id = $1`, id)
	var p model.Promotion
	err := row.Scan(&p.ID, &p.StoreID, &p.Title, &p.Slug, &p.Description, &p.PriceDisplay, &p.ImageKey, &p.StartDate, &p.EndDate, &p.Status, &p.IsSponsored, &p.CreatedAt, &p.StoreName, &p.StoreSlug)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &p, err
}

func (r *Repository) PromotionListActive(ctx context.Context, limit int) ([]model.Promotion, error) {
	today := time.Now().Format("2006-01-02")
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.store_id, p.title, p.slug, p.description, p.price_display, p.image_key, p.start_date, p.end_date, p.status, p.is_sponsored, p.created_at, s.name, s.slug
		FROM promotions p
		JOIN stores s ON s.id = p.store_id
		WHERE p.status = 'active' AND p.start_date <= $1 AND p.end_date >= $1
		ORDER BY p.created_at DESC
		LIMIT $2`, today, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var promos []model.Promotion
	for rows.Next() {
		var p model.Promotion
		if err := rows.Scan(&p.ID, &p.StoreID, &p.Title, &p.Slug, &p.Description, &p.PriceDisplay, &p.ImageKey, &p.StartDate, &p.EndDate, &p.Status, &p.IsSponsored, &p.CreatedAt, &p.StoreName, &p.StoreSlug); err != nil {
			return nil, err
		}
		promos = append(promos, p)
	}
	return promos, rows.Err()
}

func (r *Repository) PromotionUpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE promotions SET status=$1 WHERE id=$2`, status, id)
	return err
}

func (r *Repository) PromotionDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM promotions WHERE id = $1`, id)
	return err
}

func (r *Repository) PromotionSlugExists(ctx context.Context, slug string) bool {
	var count int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM promotions WHERE slug = $1`, slug).Scan(&count)
	return count > 0
}

// Banners

func (r *Repository) BannerCreate(ctx context.Context, b *model.Banner) error {
	if b.Status == "" {
		b.Status = bannerStatusFromActive(b.Active)
	}
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO banners (name, advertiser_name, contact_name, contact_phone, contact_whatsapp, price_display, notes, position, image_key, link_url, start_date, end_date, status, active, priority)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		b.Name, b.AdvertiserName, b.ContactName, b.ContactPhone, b.ContactWhatsapp, b.PriceDisplay, b.Notes, b.Position, b.ImageKey, b.LinkURL, b.StartDate, b.EndDate, b.Status, b.Active, b.Priority)
	if err != nil {
		return err
	}
	b.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) BannerUpdate(ctx context.Context, b *model.Banner) error {
	if b.Status == "" {
		b.Status = bannerStatusFromActive(b.Active)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE banners SET name=$1, advertiser_name=$2, contact_name=$3, contact_phone=$4, contact_whatsapp=$5, price_display=$6, notes=$7, position=$8, image_key=$9, link_url=$10, start_date=$11, end_date=$12, status=$13, active=$14, priority=$15
		WHERE id=$16`,
		b.Name, b.AdvertiserName, b.ContactName, b.ContactPhone, b.ContactWhatsapp, b.PriceDisplay, b.Notes, b.Position, b.ImageKey, b.LinkURL, b.StartDate, b.EndDate, b.Status, b.Active, b.Priority, b.ID)
	return err
}

func (r *Repository) BannerGetByID(ctx context.Context, id int64) (*model.Banner, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, name, COALESCE(advertiser_name, ''), COALESCE(contact_name, ''), COALESCE(contact_phone, ''), COALESCE(contact_whatsapp, ''), COALESCE(price_display, ''), COALESCE(notes, ''), position, image_key, link_url, start_date, end_date, COALESCE(status, CASE WHEN active THEN 'active' ELSE 'paused' END), active, priority, created_at FROM banners WHERE id = $1`, id)
	var b model.Banner
	err := row.Scan(&b.ID, &b.Name, &b.AdvertiserName, &b.ContactName, &b.ContactPhone, &b.ContactWhatsapp, &b.PriceDisplay, &b.Notes, &b.Position, &b.ImageKey, &b.LinkURL, &b.StartDate, &b.EndDate, &b.Status, &b.Active, &b.Priority, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &b, err
}

func (r *Repository) BannerGetActiveByPosition(ctx context.Context, position string) (*model.Banner, error) {
	today := time.Now().Format("2006-01-02")
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, COALESCE(advertiser_name, ''), COALESCE(contact_name, ''), COALESCE(contact_phone, ''), COALESCE(contact_whatsapp, ''), COALESCE(price_display, ''), COALESCE(notes, ''), position, image_key, link_url, start_date, end_date, COALESCE(status, CASE WHEN active THEN 'active' ELSE 'paused' END), active, priority, created_at
		FROM banners
		WHERE position = $1 AND active = true AND COALESCE(status, 'active') = 'active' AND start_date <= $2 AND end_date >= $2 AND image_key <> ''
		ORDER BY priority DESC, created_at DESC
		LIMIT 1`, position, today)
	var b model.Banner
	err := row.Scan(&b.ID, &b.Name, &b.AdvertiserName, &b.ContactName, &b.ContactPhone, &b.ContactWhatsapp, &b.PriceDisplay, &b.Notes, &b.Position, &b.ImageKey, &b.LinkURL, &b.StartDate, &b.EndDate, &b.Status, &b.Active, &b.Priority, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &b, err
}

func (r *Repository) BannerList(ctx context.Context) ([]model.Banner, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, name, COALESCE(advertiser_name, ''), COALESCE(contact_name, ''), COALESCE(contact_phone, ''), COALESCE(contact_whatsapp, ''), COALESCE(price_display, ''), COALESCE(notes, ''), position, image_key, link_url, start_date, end_date, COALESCE(status, CASE WHEN active THEN 'active' ELSE 'paused' END), active, priority, created_at FROM banners ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var banners []model.Banner
	for rows.Next() {
		var b model.Banner
		if err := rows.Scan(&b.ID, &b.Name, &b.AdvertiserName, &b.ContactName, &b.ContactPhone, &b.ContactWhatsapp, &b.PriceDisplay, &b.Notes, &b.Position, &b.ImageKey, &b.LinkURL, &b.StartDate, &b.EndDate, &b.Status, &b.Active, &b.Priority, &b.CreatedAt); err != nil {
			return nil, err
		}
		banners = append(banners, b)
	}
	return banners, rows.Err()
}

func (r *Repository) BannerDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM banners WHERE id = $1`, id)
	return err
}

func (r *Repository) BannerCountActiveInPeriod(ctx context.Context, position string, start, end time.Time) (int, error) {
	return r.BannerCountActiveInPeriodExcluding(ctx, position, start, end, 0)
}

func (r *Repository) BannerCountActiveInPeriodExcluding(ctx context.Context, position string, start, end time.Time, excludeID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM banners 
		WHERE position = $1 AND active = true AND COALESCE(status, 'active') = 'active'
		AND substr(start_date, 1, 10) <= $2
		AND substr(end_date, 1, 10) >= $3
		AND ($4 = 0 OR id <> $4)`, position, end.Format("2006-01-02"), start.Format("2006-01-02"), excludeID).Scan(&count)
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
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO neighborhoods (slug, name, description, meta_title, meta_description, cover_image_key)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		n.Slug, n.Name, n.Description, n.MetaTitle, n.MetaDescription, n.CoverImageKey)
	if err != nil {
		return err
	}
	n.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) NeighborhoodUpdate(ctx context.Context, n *model.Neighborhood) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE neighborhoods SET slug=$1, name=$2, description=$3, meta_title=$4, meta_description=$5, cover_image_key=$6
		WHERE id=$7`,
		n.Slug, n.Name, n.Description, n.MetaTitle, n.MetaDescription, n.CoverImageKey, n.ID)
	return err
}

func (r *Repository) NeighborhoodGetBySlug(ctx context.Context, slug string) (*model.Neighborhood, error) {
	row := r.db.QueryRowContext(ctx, `SELECT id, slug, name, description, meta_title, meta_description, cover_image_key, created_at FROM neighborhoods WHERE slug = $1`, slug)
	var n model.Neighborhood
	err := row.Scan(&n.ID, &n.Slug, &n.Name, &n.Description, &n.MetaTitle, &n.MetaDescription, &n.CoverImageKey, &n.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &n, err
}

func (r *Repository) NeighborhoodList(ctx context.Context) ([]model.Neighborhood, error) {
	rows, err := r.db.QueryContext(ctx, `SELECT id, slug, name, description, meta_title, meta_description, cover_image_key, created_at FROM neighborhoods ORDER BY name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var neighborhoods []model.Neighborhood
	for rows.Next() {
		var n model.Neighborhood
		if err := rows.Scan(&n.ID, &n.Slug, &n.Name, &n.Description, &n.MetaTitle, &n.MetaDescription, &n.CoverImageKey, &n.CreatedAt); err != nil {
			return nil, err
		}
		neighborhoods = append(neighborhoods, n)
	}
	return neighborhoods, rows.Err()
}

func (r *Repository) NeighborhoodDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM neighborhoods WHERE id = $1`, id)
	return err
}

func (r *Repository) NeighborhoodSlugExists(ctx context.Context, slug string) bool {
	var count int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM neighborhoods WHERE slug = $1`, slug).Scan(&count)
	return count > 0
}

// Influencers

func (r *Repository) InfluencerCreate(ctx context.Context, i *model.Influencer) error {
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO influencers (slug, name, bio, city_area, instagram, tiktok, youtube, whatsapp, avatar_key, cover_image_key, is_featured, is_sponsored, active)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)`,
		i.Slug, i.Name, i.Bio, i.CityArea, i.Instagram, i.TikTok, i.YouTube, i.Whatsapp, i.AvatarKey, i.CoverImageKey, i.IsFeatured, i.IsSponsored, i.Active)
	if err != nil {
		return err
	}
	i.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) InfluencerUpdate(ctx context.Context, i *model.Influencer) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE influencers SET slug=$1, name=$2, bio=$3, city_area=$4, instagram=$5, tiktok=$6, youtube=$7, whatsapp=$8, avatar_key=$9, cover_image_key=$10, is_featured=$11, is_sponsored=$12, active=$13
		WHERE id=$14`,
		i.Slug, i.Name, i.Bio, i.CityArea, i.Instagram, i.TikTok, i.YouTube, i.Whatsapp, i.AvatarKey, i.CoverImageKey, i.IsFeatured, i.IsSponsored, i.Active, i.ID)
	return err
}

func (r *Repository) InfluencerGetBySlug(ctx context.Context, slug string) (*model.Influencer, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, slug, name, bio, city_area, instagram, tiktok, youtube, whatsapp, avatar_key, cover_image_key, is_featured, is_sponsored, active, created_at
		FROM influencers WHERE slug = $1`, slug)
	var i model.Influencer
	err := row.Scan(&i.ID, &i.Slug, &i.Name, &i.Bio, &i.CityArea, &i.Instagram, &i.TikTok, &i.YouTube, &i.Whatsapp, &i.AvatarKey, &i.CoverImageKey, &i.IsFeatured, &i.IsSponsored, &i.Active, &i.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &i, err
}

func (r *Repository) InfluencerGetByID(ctx context.Context, id int64) (*model.Influencer, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, slug, name, bio, city_area, instagram, tiktok, youtube, whatsapp, avatar_key, cover_image_key, is_featured, is_sponsored, active, created_at
		FROM influencers WHERE id = $1`, id)
	var i model.Influencer
	err := row.Scan(&i.ID, &i.Slug, &i.Name, &i.Bio, &i.CityArea, &i.Instagram, &i.TikTok, &i.YouTube, &i.Whatsapp, &i.AvatarKey, &i.CoverImageKey, &i.IsFeatured, &i.IsSponsored, &i.Active, &i.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	return &i, err
}

func (r *Repository) InfluencerList(ctx context.Context, activeOnly bool, limit int) ([]model.Influencer, error) {
	q := `SELECT id, slug, name, bio, city_area, instagram, tiktok, youtube, whatsapp, avatar_key, cover_image_key, is_featured, is_sponsored, active, created_at FROM influencers`
	var args []any
	if activeOnly {
		q += ` WHERE active = true`
	}
	q += ` ORDER BY is_featured DESC, is_sponsored DESC, created_at DESC`
	if limit > 0 {
		q += ` LIMIT $1`
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
		if err := rows.Scan(&i.ID, &i.Slug, &i.Name, &i.Bio, &i.CityArea, &i.Instagram, &i.TikTok, &i.YouTube, &i.Whatsapp, &i.AvatarKey, &i.CoverImageKey, &i.IsFeatured, &i.IsSponsored, &i.Active, &i.CreatedAt); err != nil {
			return nil, err
		}
		influencers = append(influencers, i)
	}
	return influencers, rows.Err()
}

func (r *Repository) InfluencerDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM influencers WHERE id = $1`, id)
	return err
}

func (r *Repository) InfluencerSlugExists(ctx context.Context, slug string) bool {
	var count int
	r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM influencers WHERE slug = $1`, slug).Scan(&count)
	return count > 0
}

// Jobs

func (r *Repository) JobCreate(ctx context.Context, j *model.Job) error {
	if j.MaxAttempts <= 0 {
		j.MaxAttempts = 3
	}
	res, err := r.db.ExecContext(ctx, `
		INSERT INTO jobs (type, payload, status, run_at, max_attempts)
		VALUES ($1, $2, $3, $4, $5)`,
		j.Type, j.Payload, j.Status, j.RunAt, j.MaxAttempts)
	if err != nil {
		return err
	}
	j.ID, _ = res.LastInsertId()
	return nil
}

func (r *Repository) JobHasActiveType(ctx context.Context, jobType model.JobType) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM jobs
		WHERE type = $1 AND status IN ('pending', 'running')`, jobType).Scan(&count)
	return count > 0, err
}

func (r *Repository) JobGetPending(ctx context.Context, limit int) ([]model.Job, error) {
	now := time.Now()
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, type, payload, status, run_at, created_at, attempts, max_attempts, error, processed_at
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
		if err := rows.Scan(&j.ID, &j.Type, &j.Payload, &j.Status, &j.RunAt, &j.CreatedAt, &j.Attempts, &j.MaxAttempts, &j.Error, &j.ProcessedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, j)
	}
	return jobs, rows.Err()
}

func (r *Repository) JobClaimPending(ctx context.Context, limit int) ([]model.Job, error) {
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
			SELECT id, type, payload, status, run_at, created_at, attempts, max_attempts, error, processed_at
			FROM jobs
			WHERE id = $1`, id).
			Scan(&job.ID, &job.Type, &job.Payload, &job.Status, &job.RunAt, &job.CreatedAt, &job.Attempts, &job.MaxAttempts, &job.Error, &job.ProcessedAt)
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
		INSERT INTO dead_jobs (type, payload, status, run_at, created_at, attempts, max_attempts, error)
		SELECT type, payload, 'failed', run_at, created_at, attempts, max_attempts, error FROM jobs WHERE id = $1`, id)
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
		SELECT id, type, payload, status, run_at, created_at, attempts, max_attempts, error, processed_at
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
		if err := rows.Scan(&job.ID, &job.Type, &job.Payload, &job.Status, &job.RunAt, &job.CreatedAt, &job.Attempts, &job.MaxAttempts, &job.Error, &job.ProcessedAt); err != nil {
			return nil, err
		}
		jobs = append(jobs, job)
	}
	return jobs, rows.Err()
}

// Metrics

func (r *Repository) MetricTrack(ctx context.Context, m *model.Metric) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO metrics (metric_type, entity_type, entity_id, user_id, ip_address, user_agent, referrer)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		m.MetricType, m.EntityType, m.EntityID, m.UserID, m.IPAddress, m.UserAgent, m.Referrer)
	return err
}

func (r *Repository) MetricCountByEntity(ctx context.Context, metricType, entityType string, entityID int64) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM metrics 
		WHERE metric_type = $1 AND entity_type = $2 AND entity_id = $3`, metricType, entityType, entityID).Scan(&count)
	return count, err
}

func (r *Repository) MetricCountByType(ctx context.Context, metricType string) (int, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM metrics WHERE metric_type = $1`, metricType).Scan(&count)
	return count, err
}

func (r *Repository) MetricTotals(ctx context.Context, limit int) ([]model.MetricTotal, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT metric_type, COUNT(*) AS total
		FROM metrics
		GROUP BY metric_type
		ORDER BY total DESC, metric_type ASC
		LIMIT $1`, limit)
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
		WHERE metric_type = $1
		GROUP BY metric_type, entity_type, entity_id
		ORDER BY total DESC, entity_id ASC
		LIMIT $2`, metricType, limit)
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
	queries := []struct {
		prefix string
		sql    string
	}{
		{
			prefix: "/noticia/",
			sql:    `SELECT slug, updated_at FROM posts WHERE status = 'published' ORDER BY updated_at DESC`,
		},
		{
			prefix: "/loja/",
			sql:    `SELECT slug, created_at FROM stores WHERE active = true ORDER BY created_at DESC`,
		},
		{
			prefix: "/promocao/",
			sql: `SELECT slug, created_at FROM promotions
				WHERE status = 'active' AND start_date <= date('now') AND end_date >= date('now')
				ORDER BY created_at DESC`,
		},
		{
			prefix: "/bairro/",
			sql:    `SELECT slug, created_at FROM neighborhoods ORDER BY name`,
		},
		{
			prefix: "/influencer/",
			sql:    `SELECT slug, created_at FROM influencers WHERE active = true ORDER BY created_at DESC`,
		},
		{
			prefix: "/tag/",
			sql:    `SELECT slug, created_at FROM tags WHERE active = true ORDER BY name`,
		},
	}

	for _, query := range queries {
		rows, err := r.db.QueryContext(ctx, query.sql)
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
