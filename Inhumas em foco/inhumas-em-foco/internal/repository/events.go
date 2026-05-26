package repository

import (
	"context"
	"database/sql"
	"strings"

	"inhumas-em-foco/internal/model"
)

func (r *Repository) EventCreate(ctx context.Context, event *model.Event) error {
	if event.TenantID <= 0 {
		event.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO events (tenant_id, slug, title, description, location, organizer, ticket_url, price_display, image_key, status, is_featured, is_sponsored, meta_title, meta_description, start_at, end_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16)`,
		event.TenantID, event.Slug, event.Title, event.Description, event.Location, event.Organizer, event.TicketURL, event.PriceDisplay, event.ImageKey, normalizeEventStatus(event.Status), event.IsFeatured, event.IsSponsored, event.MetaTitle, event.MetaDescription, event.StartAt, event.EndAt)
	if err != nil {
		return err
	}
	event.ID = id
	return nil
}

func (r *Repository) EventUpdate(ctx context.Context, event *model.Event) error {
	if event.TenantID <= 0 {
		event.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE events
		SET slug=$1, title=$2, description=$3, location=$4, organizer=$5, ticket_url=$6, price_display=$7, image_key=$8,
		    status=$9, is_featured=$10, is_sponsored=$11, meta_title=$12, meta_description=$13, start_at=$14, end_at=$15, updated_at=CURRENT_TIMESTAMP
		WHERE tenant_id=$16 AND id=$17`,
		event.Slug, event.Title, event.Description, event.Location, event.Organizer, event.TicketURL, event.PriceDisplay, event.ImageKey,
		normalizeEventStatus(event.Status), event.IsFeatured, event.IsSponsored, event.MetaTitle, event.MetaDescription, event.StartAt, event.EndAt, event.TenantID, event.ID)
	return err
}

func (r *Repository) EventGetBySlug(ctx context.Context, slug string) (*model.Event, error) {
	row := r.db.QueryRowContext(ctx, eventSelectSQL()+` WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug)
	event, err := scanEvent(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *Repository) EventGetByID(ctx context.Context, id int64) (*model.Event, error) {
	row := r.db.QueryRowContext(ctx, eventSelectSQL()+` WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	event, err := scanEvent(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &event, nil
}

func (r *Repository) EventList(ctx context.Context, activeOnly bool, limit int) ([]model.Event, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	query := eventSelectSQL() + ` WHERE tenant_id = $1`
	args := []any{tenantIDFromContext(ctx)}
	if activeOnly {
		query += ` AND status = 'active'`
	}
	query += ` ORDER BY is_featured DESC, is_sponsored DESC, start_at ASC, id DESC LIMIT $2`
	args = append(args, limit)
	return r.eventList(ctx, query, args...)
}

func (r *Repository) EventDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM events WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) EventSlugExists(ctx context.Context, slug string) bool {
	var count int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM events WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug).Scan(&count)
	return count > 0
}

func (r *Repository) eventList(ctx context.Context, query string, args ...any) ([]model.Event, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var events []model.Event
	for rows.Next() {
		event, err := scanEvent(rows)
		if err != nil {
			return nil, err
		}
		events = append(events, event)
	}
	return events, rows.Err()
}

func eventSelectSQL() string {
	return `SELECT tenant_id, id, slug, title, COALESCE(description, ''), COALESCE(location, ''), COALESCE(organizer, ''), COALESCE(ticket_url, ''), COALESCE(price_display, ''), COALESCE(image_key, ''), COALESCE(status, 'active'), COALESCE(is_featured, false), COALESCE(is_sponsored, false), COALESCE(meta_title, ''), COALESCE(meta_description, ''), start_at, end_at, created_at, updated_at FROM events`
}

type eventScanner interface {
	Scan(dest ...any) error
}

func scanEvent(scanner eventScanner) (model.Event, error) {
	var event model.Event
	var endAt sql.NullTime
	err := scanner.Scan(
		&event.TenantID,
		&event.ID,
		&event.Slug,
		&event.Title,
		&event.Description,
		&event.Location,
		&event.Organizer,
		&event.TicketURL,
		&event.PriceDisplay,
		&event.ImageKey,
		&event.Status,
		&event.IsFeatured,
		&event.IsSponsored,
		&event.MetaTitle,
		&event.MetaDescription,
		&event.StartAt,
		&endAt,
		&event.CreatedAt,
		&event.UpdatedAt,
	)
	if endAt.Valid {
		event.EndAt = &endAt.Time
	}
	return event, err
}

func normalizeEventStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "draft", "archived":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "active"
	}
}
