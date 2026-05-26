package repository

import (
	"context"
	"database/sql"
	"strconv"
	"strings"

	"inhumas-em-foco/internal/model"
)

type ClassifiedFilter struct {
	Query      string
	Category   string
	ActiveOnly bool
	Limit      int
	Offset     int
}

func (r *Repository) ClassifiedCreate(ctx context.Context, classified *model.Classified) error {
	if classified.TenantID <= 0 {
		classified.TenantID = tenantIDFromContext(ctx)
	}
	id, err := r.insertID(ctx, `
		INSERT INTO classifieds (tenant_id, slug, title, description, category, price_display, contact_name, contact_phone, contact_whatsapp, location, image_key, status, is_featured, is_sponsored, meta_title, meta_description, expires_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17)`,
		classified.TenantID, classified.Slug, classified.Title, classified.Description, classified.Category, classified.PriceDisplay, classified.ContactName, classified.ContactPhone, classified.ContactWhatsapp, classified.Location, classified.ImageKey, normalizeClassifiedStatus(classified.Status), classified.IsFeatured, classified.IsSponsored, classified.MetaTitle, classified.MetaDescription, classified.ExpiresAt)
	if err != nil {
		return err
	}
	classified.ID = id
	return nil
}

func (r *Repository) ClassifiedUpdate(ctx context.Context, classified *model.Classified) error {
	if classified.TenantID <= 0 {
		classified.TenantID = tenantIDFromContext(ctx)
	}
	_, err := r.db.ExecContext(ctx, `
		UPDATE classifieds
		SET slug=$1, title=$2, description=$3, category=$4, price_display=$5, contact_name=$6, contact_phone=$7, contact_whatsapp=$8, location=$9, image_key=$10,
		    status=$11, is_featured=$12, is_sponsored=$13, meta_title=$14, meta_description=$15, expires_at=$16, updated_at=CURRENT_TIMESTAMP
		WHERE tenant_id=$17 AND id=$18`,
		classified.Slug, classified.Title, classified.Description, classified.Category, classified.PriceDisplay, classified.ContactName, classified.ContactPhone, classified.ContactWhatsapp, classified.Location, classified.ImageKey,
		normalizeClassifiedStatus(classified.Status), classified.IsFeatured, classified.IsSponsored, classified.MetaTitle, classified.MetaDescription, classified.ExpiresAt, classified.TenantID, classified.ID)
	return err
}

func (r *Repository) ClassifiedGetBySlug(ctx context.Context, slug string) (*model.Classified, error) {
	row := r.db.QueryRowContext(ctx, classifiedSelectSQL()+` WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug)
	classified, err := scanClassified(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &classified, nil
}

func (r *Repository) ClassifiedGetByID(ctx context.Context, id int64) (*model.Classified, error) {
	row := r.db.QueryRowContext(ctx, classifiedSelectSQL()+` WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	classified, err := scanClassified(row)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &classified, nil
}

func (r *Repository) ClassifiedList(ctx context.Context, filter ClassifiedFilter) ([]model.Classified, error) {
	if filter.Limit <= 0 || filter.Limit > 500 {
		filter.Limit = 100
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	where, args := classifiedWhere(tenantIDFromContext(ctx), filter)
	args = append(args, filter.Limit, filter.Offset)
	query := classifiedSelectSQL() + ` ` + where + `
		ORDER BY is_featured DESC, is_sponsored DESC, created_at DESC, id DESC
		LIMIT $` + strconv.Itoa(len(args)-1) + ` OFFSET $` + strconv.Itoa(len(args))
	return r.classifiedList(ctx, query, args...)
}

func (r *Repository) ClassifiedDelete(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM classifieds WHERE tenant_id = $1 AND id = $2`, tenantIDFromContext(ctx), id)
	return err
}

func (r *Repository) ClassifiedSlugExists(ctx context.Context, slug string) bool {
	var count int
	_ = r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM classifieds WHERE tenant_id = $1 AND slug = $2`, tenantIDFromContext(ctx), slug).Scan(&count)
	return count > 0
}

func (r *Repository) classifiedList(ctx context.Context, query string, args ...any) ([]model.Classified, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var classifieds []model.Classified
	for rows.Next() {
		classified, err := scanClassified(rows)
		if err != nil {
			return nil, err
		}
		classifieds = append(classifieds, classified)
	}
	return classifieds, rows.Err()
}

func classifiedWhere(tenantID int64, filter ClassifiedFilter) (string, []any) {
	clauses := []string{"tenant_id = $1"}
	args := []any{tenantID}
	if filter.ActiveOnly {
		clauses = append(clauses, "status = 'active'")
		clauses = append(clauses, "(expires_at IS NULL OR expires_at >= date('now'))")
	}
	if strings.TrimSpace(filter.Category) != "" {
		args = append(args, strings.TrimSpace(filter.Category))
		clauses = append(clauses, "category = $"+strconv.Itoa(len(args)))
	}
	if strings.TrimSpace(filter.Query) != "" {
		args = append(args, "%"+strings.ToLower(strings.TrimSpace(filter.Query))+"%")
		placeholder := "$" + strconv.Itoa(len(args))
		clauses = append(clauses, "(LOWER(title) LIKE "+placeholder+" OR LOWER(COALESCE(description, '')) LIKE "+placeholder+" OR LOWER(COALESCE(category, '')) LIKE "+placeholder+" OR LOWER(COALESCE(location, '')) LIKE "+placeholder+")")
	}
	return "WHERE " + strings.Join(clauses, " AND "), args
}

func classifiedSelectSQL() string {
	return `SELECT tenant_id, id, slug, title, COALESCE(description, ''), COALESCE(category, ''), COALESCE(price_display, ''), COALESCE(contact_name, ''), COALESCE(contact_phone, ''), COALESCE(contact_whatsapp, ''), COALESCE(location, ''), COALESCE(image_key, ''), COALESCE(status, 'active'), COALESCE(is_featured, false), COALESCE(is_sponsored, false), COALESCE(meta_title, ''), COALESCE(meta_description, ''), expires_at, created_at, updated_at FROM classifieds`
}

type classifiedScanner interface {
	Scan(dest ...any) error
}

func scanClassified(scanner classifiedScanner) (model.Classified, error) {
	var classified model.Classified
	var expiresAt sql.NullTime
	err := scanner.Scan(
		&classified.TenantID,
		&classified.ID,
		&classified.Slug,
		&classified.Title,
		&classified.Description,
		&classified.Category,
		&classified.PriceDisplay,
		&classified.ContactName,
		&classified.ContactPhone,
		&classified.ContactWhatsapp,
		&classified.Location,
		&classified.ImageKey,
		&classified.Status,
		&classified.IsFeatured,
		&classified.IsSponsored,
		&classified.MetaTitle,
		&classified.MetaDescription,
		&expiresAt,
		&classified.CreatedAt,
		&classified.UpdatedAt,
	)
	if expiresAt.Valid {
		classified.ExpiresAt = &expiresAt.Time
	}
	return classified, err
}

func normalizeClassifiedStatus(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "draft", "archived", "sold":
		return strings.ToLower(strings.TrimSpace(status))
	default:
		return "active"
	}
}
