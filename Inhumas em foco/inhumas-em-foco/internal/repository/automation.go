package repository

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"inhumas-em-foco/internal/model"
)

func (r *Repository) AutomationSourceCreate(ctx context.Context, source *model.AutomationSource) error {
	if source.TenantID <= 0 {
		source.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO automation_sources (tenant_id, name, source_type, url, default_category_id, active)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		source.TenantID, source.Name, normalizeAutomationSourceType(source.SourceType), source.URL, source.DefaultCategoryID, source.Active)
	if err != nil {
		return err
	}
	source.ID = id
	return nil
}

func (r *Repository) AutomationSourceUpdate(ctx context.Context, source *model.AutomationSource) error {
	source.TenantID = tenantIDFromContext(ctx)
	_, err := r.db.ExecContext(ctx, `
		UPDATE automation_sources
		SET name=$1, source_type=$2, url=$3, default_category_id=$4, active=$5, updated_at=CURRENT_TIMESTAMP
		WHERE tenant_id=$6 AND id=$7`,
		source.Name, normalizeAutomationSourceType(source.SourceType), source.URL, source.DefaultCategoryID, source.Active, source.TenantID, source.ID)
	return err
}

func (r *Repository) AutomationSourceDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM automation_sources WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) AutomationSourceGetByID(ctx context.Context, id int64) (*model.AutomationSource, error) {
	row := r.db.QueryRowContext(ctx, automationSourceSelect()+` WHERE s.tenant_id = $1 AND s.id = $2`, tenantIDFromContext(ctx), id)
	var source model.AutomationSource
	if err := scanAutomationSource(row, &source); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &source, nil
}

func (r *Repository) AutomationSourceList(ctx context.Context, activeOnly bool, limit int) ([]model.AutomationSource, error) {
	if limit <= 0 {
		limit = 200
	}
	query := automationSourceSelect()
	args := []any{tenantIDFromContext(ctx)}
	query += ` WHERE s.tenant_id = $1`
	if activeOnly {
		query += ` AND s.active = true`
	}
	args = append(args, limit)
	query += ` ORDER BY s.active DESC, s.name ASC LIMIT $2`
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var sources []model.AutomationSource
	for rows.Next() {
		var source model.AutomationSource
		if err := scanAutomationSource(rows, &source); err != nil {
			return nil, err
		}
		sources = append(sources, source)
	}
	return sources, rows.Err()
}

func (r *Repository) AutomationSourceMarkRun(ctx context.Context, id int64, at time.Time) error {
	_, err := r.db.ExecContext(ctx, `
		UPDATE automation_sources
		SET last_run_at=$1, updated_at=CURRENT_TIMESTAMP
		WHERE tenant_id=$2 AND id=$3`, at, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) AutomationRunCreate(ctx context.Context, run *model.AutomationRun) error {
	if run.TenantID <= 0 {
		run.TenantID = tenantIDFromContext(ctx)
	}
	if run.StartedAt.IsZero() {
		run.StartedAt = time.Now()
	}
	id, err := r.insertID(ctx, `
		INSERT INTO automation_runs (tenant_id, source_id, status, items_found, drafts_created, duplicates, error, log, started_at, finished_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		run.TenantID, run.SourceID, run.Status, run.ItemsFound, run.DraftsCreated, run.Duplicates, run.Error, run.Log, run.StartedAt, run.FinishedAt)
	if err != nil {
		return err
	}
	run.ID = id
	return nil
}

func (r *Repository) AutomationRunUpdate(ctx context.Context, run *model.AutomationRun) error {
	run.TenantID = tenantIDFromContext(ctx)
	_, err := r.db.ExecContext(ctx, `
		UPDATE automation_runs
		SET status=$1, items_found=$2, drafts_created=$3, duplicates=$4, error=$5, log=$6, finished_at=$7
		WHERE tenant_id=$8 AND id=$9`,
		run.Status, run.ItemsFound, run.DraftsCreated, run.Duplicates, run.Error, run.Log, run.FinishedAt, run.TenantID, run.ID)
	return err
}

func (r *Repository) AutomationRunList(ctx context.Context, limit int) ([]model.AutomationRun, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT r.tenant_id, r.id, r.source_id, COALESCE(s.name, ''), r.status, r.items_found, r.drafts_created, r.duplicates, COALESCE(r.error, ''), COALESCE(r.log, ''), r.started_at, r.finished_at
		FROM automation_runs r
		LEFT JOIN automation_sources s ON s.id = r.source_id AND s.tenant_id = r.tenant_id
		WHERE r.tenant_id = $1
		ORDER BY r.started_at DESC
		LIMIT $2`, tenantIDFromContext(ctx), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var runs []model.AutomationRun
	for rows.Next() {
		var run model.AutomationRun
		if err := rows.Scan(&run.TenantID, &run.ID, &run.SourceID, &run.SourceName, &run.Status, &run.ItemsFound, &run.DraftsCreated, &run.Duplicates, &run.Error, &run.Log, &run.StartedAt, &run.FinishedAt); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, rows.Err()
}

func (r *Repository) AutomationDraftQueue(ctx context.Context, limit int) ([]model.Post, error) {
	if limit <= 0 {
		limit = 50
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT p.tenant_id, p.id, p.title, p.slug, p.excerpt, p.cover_image_key, p.category_id, p.author_id, p.status, p.is_sponsored, COALESCE(p.is_featured, false), p.published_at, p.created_at, p.updated_at, COALESCE(c.name, ''), COALESCE(u.name, '')
		FROM posts p
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		WHERE p.tenant_id = $1 AND p.status = 'draft' AND COALESCE(p.source_url, '') <> ''
		ORDER BY p.created_at DESC
		LIMIT $2`, tenantIDFromContext(ctx), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPostList(rows)
}

func (r *Repository) AutomationPostDuplicateExact(ctx context.Context, title, sourceURL string) (bool, string, error) {
	if strings.TrimSpace(sourceURL) != "" {
		var count int
		if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE tenant_id = $1 AND source_url = $2`, tenantIDFromContext(ctx), strings.TrimSpace(sourceURL)).Scan(&count); err != nil {
			return false, "", err
		}
		if count > 0 {
			return true, "URL ja cadastrada", nil
		}
	}
	if strings.TrimSpace(title) != "" {
		var count int
		if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts WHERE tenant_id = $1 AND lower(title) = lower($2)`, tenantIDFromContext(ctx), strings.TrimSpace(title)).Scan(&count); err != nil {
			return false, "", err
		}
		if count > 0 {
			return true, "titulo identico", nil
		}
	}
	return false, "", nil
}

func (r *Repository) AutomationRecentPostTitles(ctx context.Context, limit int) ([]model.Post, error) {
	if limit <= 0 {
		limit = 300
	}
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, title, COALESCE(source_url, '')
		FROM posts
		WHERE tenant_id = $1
		ORDER BY created_at DESC
		LIMIT $2`, tenantIDFromContext(ctx), limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []model.Post
	for rows.Next() {
		var post model.Post
		if err := rows.Scan(&post.ID, &post.Title, &post.SourceURL); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func automationSourceSelect() string {
	return `
		SELECT s.tenant_id, s.id, s.name, s.source_type, s.url, s.default_category_id, s.active, s.last_run_at, s.created_at, s.updated_at, COALESCE(c.name, '')
		FROM automation_sources s
		LEFT JOIN categories c ON c.id = s.default_category_id AND c.tenant_id = s.tenant_id`
}

type automationSourceScanner interface {
	Scan(dest ...any) error
}

func scanAutomationSource(scanner automationSourceScanner, source *model.AutomationSource) error {
	return scanner.Scan(&source.TenantID, &source.ID, &source.Name, &source.SourceType, &source.URL, &source.DefaultCategoryID, &source.Active, &source.LastRunAt, &source.CreatedAt, &source.UpdatedAt, &source.CategoryName)
}

func normalizeAutomationSourceType(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(model.AutomationSourceOfficial):
		return string(model.AutomationSourceOfficial)
	default:
		return string(model.AutomationSourceRSS)
	}
}
