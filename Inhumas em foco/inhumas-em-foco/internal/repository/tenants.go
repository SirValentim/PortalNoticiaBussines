package repository

import (
	"context"
	"database/sql"
	"strings"

	"inhumas-em-foco/internal/model"
)

func (r *Repository) TenantCreate(ctx context.Context, tenant *model.Tenant) error {
	normalizeTenant(tenant)
	id, err := r.insertID(ctx, `
		INSERT INTO tenants (name, slug, status, primary_domain, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`,
		tenant.Name,
		tenant.Slug,
		tenant.Status,
		tenant.PrimaryDomain,
	)
	if err != nil {
		return err
	}
	tenant.ID = id
	return nil
}

func (r *Repository) TenantUpdate(ctx context.Context, tenant *model.Tenant) error {
	normalizeTenant(tenant)
	_, err := r.db.ExecContext(ctx, `
		UPDATE tenants
		SET name=$1, slug=$2, status=$3, primary_domain=$4, updated_at=CURRENT_TIMESTAMP
		WHERE id=$5`,
		tenant.Name,
		tenant.Slug,
		tenant.Status,
		tenant.PrimaryDomain,
		tenant.ID,
	)
	return err
}

func (r *Repository) TenantDomainCreate(ctx context.Context, domain *model.TenantDomain) error {
	domain.Domain = normalizeDomain(domain.Domain)
	id, err := r.insertID(ctx, `
		INSERT INTO tenant_domains (tenant_id, domain, is_primary, created_at)
		VALUES ($1, $2, $3, CURRENT_TIMESTAMP)`,
		domain.TenantID,
		domain.Domain,
		domain.IsPrimary,
	)
	if err != nil {
		return err
	}
	domain.ID = id
	return nil
}

func (r *Repository) TenantDomainList(ctx context.Context, tenantID int64) ([]model.TenantDomain, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, tenant_id, domain, is_primary, created_at
		FROM tenant_domains
		WHERE tenant_id = $1
		ORDER BY is_primary DESC, domain ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var domains []model.TenantDomain
	for rows.Next() {
		var domain model.TenantDomain
		if err := rows.Scan(&domain.ID, &domain.TenantID, &domain.Domain, &domain.IsPrimary, &domain.CreatedAt); err != nil {
			return nil, err
		}
		domains = append(domains, domain)
	}
	return domains, rows.Err()
}

func (r *Repository) TenantDomainDelete(ctx context.Context, tenantID, id int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tenant_domains WHERE tenant_id = $1 AND id = $2`, tenantID, id)
	return err
}

func (r *Repository) TenantGetBySlug(ctx context.Context, slug string) (*model.Tenant, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, slug, status, COALESCE(primary_domain, ''), created_at, updated_at
		FROM tenants
		WHERE slug = $1`, strings.TrimSpace(strings.ToLower(slug)))
	return scanTenant(row)
}

func (r *Repository) TenantGetByID(ctx context.Context, id int64) (*model.Tenant, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, name, slug, status, COALESCE(primary_domain, ''), created_at, updated_at
		FROM tenants
		WHERE id = $1`, id)
	return scanTenant(row)
}

func (r *Repository) TenantGetByDomain(ctx context.Context, domain string) (*model.Tenant, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT tenants.id, tenants.name, tenants.slug, tenants.status, COALESCE(tenants.primary_domain, ''), tenants.created_at, tenants.updated_at
		FROM tenants
		INNER JOIN tenant_domains ON tenant_domains.tenant_id = tenants.id
		WHERE tenant_domains.domain = $1`, normalizeDomain(domain))
	return scanTenant(row)
}

func (r *Repository) TenantList(ctx context.Context) ([]model.Tenant, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, name, slug, status, COALESCE(primary_domain, ''), created_at, updated_at
		FROM tenants
		ORDER BY name ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tenants []model.Tenant
	for rows.Next() {
		var tenant model.Tenant
		if err := rows.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Status, &tenant.PrimaryDomain, &tenant.CreatedAt, &tenant.UpdatedAt); err != nil {
			return nil, err
		}
		tenants = append(tenants, tenant)
	}
	return tenants, rows.Err()
}

func (r *Repository) TenantFeatureUpsert(ctx context.Context, feature *model.TenantFeature) error {
	if feature.TenantID <= 0 {
		feature.TenantID = tenantIDFromContext(ctx)
	}
	feature.Feature = normalizeTenantFeatureName(feature.Feature)
	id, err := r.insertID(ctx, `
		INSERT INTO tenant_features (tenant_id, feature, enabled, limit_value, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (tenant_id, feature) DO UPDATE SET
			enabled=excluded.enabled,
			limit_value=excluded.limit_value,
			updated_at=CURRENT_TIMESTAMP
		RETURNING id`,
		feature.TenantID,
		feature.Feature,
		feature.Enabled,
		feature.Limit,
	)
	if err != nil {
		return err
	}
	feature.ID = id
	return nil
}

func (r *Repository) TenantFeatureGet(ctx context.Context, feature string) (*model.TenantFeature, error) {
	row := r.db.QueryRowContext(ctx, tenantFeatureSelect()+` WHERE tenant_id = $1 AND feature = $2`, tenantIDFromContext(ctx), normalizeTenantFeatureName(feature))
	return scanTenantFeature(row)
}

func (r *Repository) TenantFeatureEnabled(ctx context.Context, feature string) (bool, error) {
	tenantFeature, err := r.TenantFeatureGet(ctx, feature)
	if err != nil || tenantFeature == nil {
		return false, err
	}
	return tenantFeature.Enabled, nil
}

func (r *Repository) TenantFeatureEnabledOrDefault(ctx context.Context, feature string, fallback bool) (bool, error) {
	tenantFeature, err := r.TenantFeatureGet(ctx, feature)
	if err != nil {
		return false, err
	}
	if tenantFeature == nil {
		return fallback, nil
	}
	return tenantFeature.Enabled, nil
}

func (r *Repository) TenantFeatureList(ctx context.Context) ([]model.TenantFeature, error) {
	return r.TenantFeatureListForTenant(ctx, tenantIDFromContext(ctx))
}

func (r *Repository) TenantFeatureListForTenant(ctx context.Context, tenantID int64) ([]model.TenantFeature, error) {
	rows, err := r.db.QueryContext(ctx, tenantFeatureSelect()+` WHERE tenant_id = $1 ORDER BY feature ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var features []model.TenantFeature
	for rows.Next() {
		var feature model.TenantFeature
		if err := scanTenantFeatureRows(rows, &feature); err != nil {
			return nil, err
		}
		features = append(features, feature)
	}
	return features, rows.Err()
}

func (r *Repository) TenantUserUpsert(ctx context.Context, tenantUser *model.TenantUser) error {
	if tenantUser.TenantID <= 0 {
		tenantUser.TenantID = tenantIDFromContext(ctx)
	}
	if !tenantUser.Role.IsValid() {
		tenantUser.Role = model.RoleEditor
	}
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO tenant_users (tenant_id, user_id, role, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)
		ON CONFLICT (tenant_id, user_id) DO UPDATE SET
			role=excluded.role,
			active=excluded.active,
			updated_at=CURRENT_TIMESTAMP`,
		tenantUser.TenantID,
		tenantUser.UserID,
		tenantUser.Role,
		tenantUser.Active,
	)
	if err != nil {
		return err
	}
	row := r.db.QueryRowContext(ctx, `SELECT id FROM tenant_users WHERE tenant_id = $1 AND user_id = $2`, tenantUser.TenantID, tenantUser.UserID)
	return row.Scan(&tenantUser.ID)
}

func (r *Repository) TenantUserGet(ctx context.Context, userID int64) (*model.TenantUser, error) {
	row := r.db.QueryRowContext(ctx, tenantUserSelect()+`
		WHERE tu.tenant_id = $1 AND tu.user_id = $2`, tenantIDFromContext(ctx), userID)
	return scanTenantUser(row)
}

func (r *Repository) TenantUserList(ctx context.Context) ([]model.TenantUser, error) {
	return r.TenantUserListForTenant(ctx, tenantIDFromContext(ctx))
}

func (r *Repository) TenantUserListForTenant(ctx context.Context, tenantID int64) ([]model.TenantUser, error) {
	rows, err := r.db.QueryContext(ctx, tenantUserSelect()+`
		WHERE tu.tenant_id = $1
		ORDER BY u.name ASC, u.email ASC`, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []model.TenantUser
	for rows.Next() {
		var tenantUser model.TenantUser
		if err := scanTenantUserRows(rows, &tenantUser); err != nil {
			return nil, err
		}
		users = append(users, tenantUser)
	}
	return users, rows.Err()
}

func (r *Repository) TenantUserListByUser(ctx context.Context, userID int64) ([]model.TenantUser, error) {
	rows, err := r.db.QueryContext(ctx, tenantUserSelect()+`
		WHERE tu.user_id = $1
		ORDER BY tu.created_at DESC`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var users []model.TenantUser
	for rows.Next() {
		var tenantUser model.TenantUser
		if err := scanTenantUserRows(rows, &tenantUser); err != nil {
			return nil, err
		}
		users = append(users, tenantUser)
	}
	return users, rows.Err()
}

func (r *Repository) TenantUserDelete(ctx context.Context, userID int64) error {
	return r.TenantUserDeleteForTenant(ctx, tenantIDFromContext(ctx), userID)
}

func (r *Repository) TenantUserDeleteForTenant(ctx context.Context, tenantID, userID int64) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM tenant_users WHERE tenant_id = $1 AND user_id = $2`, tenantID, userID)
	return err
}

func scanTenant(row *sql.Row) (*model.Tenant, error) {
	var tenant model.Tenant
	err := row.Scan(&tenant.ID, &tenant.Name, &tenant.Slug, &tenant.Status, &tenant.PrimaryDomain, &tenant.CreatedAt, &tenant.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

func tenantFeatureSelect() string {
	return `SELECT id, tenant_id, feature, enabled, limit_value, created_at, updated_at FROM tenant_features`
}

func scanTenantFeature(row *sql.Row) (*model.TenantFeature, error) {
	var feature model.TenantFeature
	err := scanTenantFeatureRows(row, &feature)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &feature, nil
}

type tenantFeatureScanner interface {
	Scan(dest ...any) error
}

func scanTenantFeatureRows(scanner tenantFeatureScanner, feature *model.TenantFeature) error {
	return scanner.Scan(&feature.ID, &feature.TenantID, &feature.Feature, &feature.Enabled, &feature.Limit, &feature.CreatedAt, &feature.UpdatedAt)
}

func tenantUserSelect() string {
	return `
		SELECT tu.id, tu.tenant_id, tu.user_id, tu.role, tu.active, tu.created_at, tu.updated_at, COALESCE(u.name, ''), COALESCE(u.email, '')
		FROM tenant_users tu
		LEFT JOIN users u ON u.id = tu.user_id`
}

func scanTenantUser(row *sql.Row) (*model.TenantUser, error) {
	var tenantUser model.TenantUser
	err := scanTenantUserRows(row, &tenantUser)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &tenantUser, nil
}

type tenantUserScanner interface {
	Scan(dest ...any) error
}

func scanTenantUserRows(scanner tenantUserScanner, tenantUser *model.TenantUser) error {
	return scanner.Scan(&tenantUser.ID, &tenantUser.TenantID, &tenantUser.UserID, &tenantUser.Role, &tenantUser.Active, &tenantUser.CreatedAt, &tenantUser.UpdatedAt, &tenantUser.UserName, &tenantUser.UserEmail)
}

func normalizeTenant(tenant *model.Tenant) {
	tenant.Name = strings.TrimSpace(tenant.Name)
	tenant.Slug = strings.TrimSpace(strings.ToLower(tenant.Slug))
	tenant.Status = strings.TrimSpace(strings.ToLower(tenant.Status))
	if tenant.Status == "" {
		tenant.Status = "active"
	}
	tenant.PrimaryDomain = normalizeDomain(tenant.PrimaryDomain)
}

func normalizeTenantFeatureName(feature string) string {
	return strings.TrimSpace(strings.ToLower(feature))
}

func normalizeDomain(domain string) string {
	domain = strings.TrimSpace(strings.ToLower(domain))
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimSuffix(domain, "/")
	return domain
}
