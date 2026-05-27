package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"inhumas-em-foco/internal/model"
)

func (r *Repository) DashboardSummary(ctx context.Context, allTenants bool, tenantID int64, from, to *time.Time) (model.DashboardSummary, error) {
	var summary model.DashboardSummary
	var firstErr error
	setErr := func(err error) {
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	if allTenants && tenantID <= 0 {
		setErr(r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM tenants`).Scan(&summary.TotalTenants))
	} else {
		summary.TotalTenants = 1
	}

	for _, item := range []struct {
		status string
		dest   *int
	}{
		{"published", &summary.PublishedPosts},
		{"draft", &summary.DraftPosts},
		{"review", &summary.ReviewPosts},
		{"scheduled", &summary.ScheduledPosts},
	} {
		where, args := tenantWhere(ctx, allTenants, tenantID, "p")
		args = append(args, item.status)
		where = appendWhere(where, fmt.Sprintf("p.status = $%d", len(args)))
		setErr(r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM posts p `+where, args...).Scan(item.dest))
	}

	userWhere, userArgs := tenantUserWhere(ctx, allTenants, tenantID)
	setErr(r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		INNER JOIN tenant_users tu ON tu.user_id = u.id
		`+appendWhere(userWhere, "u.active = true AND tu.active = true"), userArgs...).Scan(&summary.ActiveUsers))
	setErr(r.db.QueryRowContext(ctx, `
		SELECT COUNT(DISTINCT u.id)
		FROM users u
		INNER JOIN tenant_users tu ON tu.user_id = u.id
		`+userWhere, userArgs...).Scan(&summary.CreatedAccounts))

	summary.TotalViews, _ = r.MetricCountFiltered(ctx, "post_view", allTenants, tenantID, nil, nil)
	today := startOfDay(time.Now())
	sevenDays := today.AddDate(0, 0, -6)
	summary.ViewsToday, _ = r.MetricCountFiltered(ctx, "post_view", allTenants, tenantID, &today, nil)
	summary.ViewsLast7Days, _ = r.MetricCountFiltered(ctx, "post_view", allTenants, tenantID, &sevenDays, nil)
	summary.ViewsSelectedRange, _ = r.MetricCountFiltered(ctx, "post_view", allTenants, tenantID, from, to)

	return summary, firstErr
}

func (r *Repository) MetricCountFiltered(ctx context.Context, metricType string, allTenants bool, tenantID int64, from, to *time.Time) (int, error) {
	where, args := tenantWhere(ctx, allTenants, tenantID, "m")
	args = append(args, metricType)
	where = appendWhere(where, fmt.Sprintf("m.metric_type = $%d", len(args)))
	where, args = appendDateRange(where, args, "m.created_at", from, to)

	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM metrics m `+where, args...).Scan(&count)
	return count, err
}

func (r *Repository) MetricTotalsFiltered(ctx context.Context, allTenants bool, tenantID int64, from, to *time.Time, limit int) ([]model.MetricTotal, error) {
	if limit <= 0 {
		limit = 20
	}
	where, args := tenantWhere(ctx, allTenants, tenantID, "m")
	where, args = appendDateRange(where, args, "m.created_at", from, to)
	args = append(args, limit)

	rows, err := r.db.QueryContext(ctx, `
		SELECT m.metric_type, COUNT(*) AS total
		FROM metrics m
		`+where+`
		GROUP BY m.metric_type
		ORDER BY total DESC, m.metric_type ASC
		LIMIT $`+fmt.Sprint(len(args)), args...)
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

func (r *Repository) TopPostMetrics(ctx context.Context, allTenants bool, tenantID int64, from, to *time.Time, limit int) ([]model.PostMetricSummary, error) {
	return r.postMetricSummaries(ctx, allTenants, tenantID, from, to, limit, "views DESC, p.published_at DESC, p.updated_at DESC")
}

func (r *Repository) LatestPublishedPostMetrics(ctx context.Context, allTenants bool, tenantID int64, limit int) ([]model.PostMetricSummary, error) {
	return r.postMetricSummaries(ctx, allTenants, tenantID, nil, nil, limit, "p.published_at DESC, p.updated_at DESC")
}

func (r *Repository) postMetricSummaries(ctx context.Context, allTenants bool, tenantID int64, from, to *time.Time, limit int, orderBy string) ([]model.PostMetricSummary, error) {
	if limit <= 0 {
		limit = 10
	}
	where, args := tenantWhere(ctx, allTenants, tenantID, "p")
	where = appendWhere(where, "p.status = 'published'")
	periodStart := len(args) + 1
	periodViewClause, periodArgs := metricPeriodClause("m", periodStart, from, to)
	periodClickClause, _ := metricPeriodClause("mc", periodStart, from, to)
	args = append(args, periodArgs...)
	args = append(args, limit)
	if strings.TrimSpace(orderBy) == "" {
		orderBy = "p.published_at DESC"
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT p.id, p.tenant_id, COALESCE(t.name, ''), p.title, p.slug, p.status,
		       COALESCE(c.name, ''), COALESCE(u.name, ''), p.published_at, p.updated_at,
		       COALESCE((SELECT COUNT(*) FROM metrics m
		                 WHERE m.tenant_id = p.tenant_id AND m.metric_type = 'post_view' AND m.entity_type = 'post' AND m.entity_id = p.id`+periodViewClause+`), 0) AS views,
		       COALESCE((SELECT COUNT(*) FROM metrics mc
		                 WHERE mc.tenant_id = p.tenant_id AND mc.metric_type IN ('post_click', 'share_click') AND mc.entity_type = 'post' AND mc.entity_id = p.id`+periodClickClause+`), 0) AS clicks,
		       COALESCE(p.reading_time_minutes, 1)
		FROM posts p
		LEFT JOIN tenants t ON t.id = p.tenant_id
		LEFT JOIN categories c ON c.id = p.category_id AND c.tenant_id = p.tenant_id
		LEFT JOIN users u ON u.id = p.author_id
		`+where+`
		ORDER BY `+orderBy+`
		LIMIT $`+fmt.Sprint(len(args)), args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var posts []model.PostMetricSummary
	for rows.Next() {
		var post model.PostMetricSummary
		if err := rows.Scan(&post.PostID, &post.TenantID, &post.TenantName, &post.Title, &post.Slug, &post.Status, &post.CategoryName, &post.AuthorName, &post.PublishedAt, &post.UpdatedAt, &post.Views, &post.Clicks, &post.ReadingTimeMinutes); err != nil {
			return nil, err
		}
		posts = append(posts, post)
	}
	return posts, rows.Err()
}

func (r *Repository) TenantMetricSummaries(ctx context.Context, allTenants bool, tenantID int64, limit int) ([]model.TenantMetricSummary, error) {
	if limit <= 0 {
		limit = 100
	}
	where, args := tenantRecordWhere(ctx, allTenants, tenantID)
	args = append(args, limit)
	rows, err := r.db.QueryContext(ctx, `
		SELECT t.id, t.name, t.slug, t.status, COALESCE(t.primary_domain, ''), t.updated_at,
		       COALESCE((SELECT COUNT(*) FROM posts p WHERE p.tenant_id = t.id), 0) AS post_count,
		       COALESCE((SELECT COUNT(*) FROM posts p WHERE p.tenant_id = t.id AND p.status = 'published'), 0) AS published_count,
		       COALESCE((SELECT COUNT(*) FROM posts p WHERE p.tenant_id = t.id AND p.status = 'draft'), 0) AS draft_count,
		       COALESCE((SELECT COUNT(DISTINCT tu.user_id) FROM tenant_users tu INNER JOIN users u ON u.id = tu.user_id WHERE tu.tenant_id = t.id AND tu.active = true AND u.active = true), 0) AS active_users,
		       COALESCE((SELECT COUNT(*) FROM metrics m WHERE m.tenant_id = t.id AND m.metric_type = 'post_view'), 0) AS total_views
		FROM tenants t
		`+where+`
		ORDER BY post_count DESC, t.name ASC
		LIMIT $`+fmt.Sprint(len(args)), args...)
	if err != nil {
		return nil, err
	}

	var summaries []model.TenantMetricSummary
	for rows.Next() {
		var summary model.TenantMetricSummary
		if err := rows.Scan(&summary.TenantID, &summary.Name, &summary.Slug, &summary.Status, &summary.PrimaryDomain, &summary.UpdatedAt, &summary.PostCount, &summary.PublishedCount, &summary.DraftCount, &summary.ActiveUsers, &summary.TotalViews); err != nil {
			rows.Close()
			return nil, err
		}
		summaries = append(summaries, summary)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return nil, err
	}
	rows.Close()

	for i := range summaries {
		features, _ := r.TenantFeatureListForTenant(ctx, summaries[i].TenantID)
		for _, feature := range features {
			if feature.Enabled {
				summaries[i].FeatureNames = append(summaries[i].FeatureNames, feature.Feature)
			}
		}
	}
	return summaries, nil
}

func tenantWhere(ctx context.Context, allTenants bool, tenantID int64, alias string) (string, []any) {
	if tenantID <= 0 && !allTenants {
		tenantID = tenantIDFromContext(ctx)
	}
	if tenantID <= 0 {
		return "", nil
	}
	column := "tenant_id"
	if strings.TrimSpace(alias) != "" {
		column = alias + ".tenant_id"
	}
	return "WHERE " + column + " = $1", []any{tenantID}
}

func tenantUserWhere(ctx context.Context, allTenants bool, tenantID int64) (string, []any) {
	if tenantID <= 0 && !allTenants {
		tenantID = tenantIDFromContext(ctx)
	}
	if tenantID <= 0 {
		return "", nil
	}
	return "WHERE tu.tenant_id = $1", []any{tenantID}
}

func tenantRecordWhere(ctx context.Context, allTenants bool, tenantID int64) (string, []any) {
	if tenantID <= 0 && !allTenants {
		tenantID = tenantIDFromContext(ctx)
	}
	if tenantID <= 0 {
		return "", nil
	}
	return "WHERE t.id = $1", []any{tenantID}
}

func appendWhere(where, clause string) string {
	clause = strings.TrimSpace(clause)
	if clause == "" {
		return where
	}
	if strings.TrimSpace(where) == "" {
		return "WHERE " + clause
	}
	return where + " AND " + clause
}

func appendDateRange(where string, args []any, column string, from, to *time.Time) (string, []any) {
	if from != nil {
		args = append(args, *from)
		where = appendWhere(where, fmt.Sprintf("%s >= $%d", column, len(args)))
	}
	if to != nil {
		args = append(args, *to)
		where = appendWhere(where, fmt.Sprintf("%s < $%d", column, len(args)))
	}
	return where, args
}

func metricPeriodClause(alias string, start int, from, to *time.Time) (string, []any) {
	var clauses []string
	var args []any
	if from != nil {
		args = append(args, *from)
		clauses = append(clauses, fmt.Sprintf("%s.created_at >= $%d", alias, start+len(args)-1))
	}
	if to != nil {
		args = append(args, *to)
		clauses = append(clauses, fmt.Sprintf("%s.created_at < $%d", alias, start+len(args)-1))
	}
	if len(clauses) == 0 {
		return "", args
	}
	return " AND " + strings.Join(clauses, " AND "), args
}

func startOfDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
