package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/middleware"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
)

func auditEntityIDLabel(id *int64) string {
	if id == nil {
		return "-"
	}
	return friendlyID(id)
}

func auditUserLabel(entry model.AuditLogEntry) string {
	if entry.UserName != "" && entry.UserEmail != "" {
		return entry.UserName + " <" + entry.UserEmail + ">"
	}
	if entry.UserEmail != "" {
		return entry.UserEmail
	}
	if entry.UserName != "" {
		return entry.UserName
	}
	if entry.UserID != nil {
		return friendlyID(entry.UserID)
	}
	return "Sistema"
}

func (h *Handler) AdminMetrics(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermSettingsManage)
	if !ok {
		return
	}
	ctx := r.Context()
	period := adminPeriodFromRequest(r)
	allTenants := h.authSvc.HasPermission(user, auth.PermTenantsManage)
	selectedTenantID := int64(0)
	if allTenants {
		selectedTenantID = selectedTenantIDFromRequest(r)
	}
	totals, _ := h.repo.MetricTotalsFiltered(ctx, allTenants, selectedTenantID, period.From, period.To, 20)
	topPosts, _ := h.repo.TopPostMetrics(ctx, allTenants, selectedTenantID, period.From, period.To, 12)
	tenantSummaries, _ := h.repo.TenantMetricSummaries(ctx, allTenants, selectedTenantID, 12)
	summary, _ := h.repo.DashboardSummary(ctx, allTenants, selectedTenantID, period.From, period.To)
	tenants, _ := h.repo.TenantList(ctx)

	h.Render(w, r, "admin_metrics.html", map[string]any{
		"Title":            "Metricas",
		"Active":           "metrics",
		"Period":           period,
		"SelectedTenantID": selectedTenantID,
		"Tenants":          tenants,
		"Summary":          summary,
		"Totals":           totals,
		"TopPosts":         topPosts,
		"TenantSummaries":  tenantSummaries,
		"HasPostViewData":  !noPostViews(topPosts),
	})
}

func (h *Handler) AdminDeadJobs(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermSettingsManage); !ok {
		return
	}
	jobs, _ := h.repo.DeadJobList(r.Context(), 100)
	count, _ := h.repo.DeadJobCount(r.Context())

	h.Render(w, r, "admin_dead_jobs.html", map[string]any{
		"Title":  "Jobs com Falha",
		"Active": "dead_jobs",
		"Jobs":   jobs,
		"Count":  count,
	})
}

func (h *Handler) AdminAuditLogs(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermSettingsManage); !ok {
		return
	}

	page := 1
	if p := r.URL.Query().Get("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}
	limit := 50
	filter := repository.AuditLogFilter{
		Action:     strings.TrimSpace(r.URL.Query().Get("action")),
		EntityType: strings.TrimSpace(r.URL.Query().Get("entity_type")),
		Limit:      limit,
		Offset:     (page - 1) * limit,
	}
	filter.UserID, _ = strconv.ParseInt(r.URL.Query().Get("user_id"), 10, 64)
	filter.EntityID, _ = strconv.ParseInt(r.URL.Query().Get("entity_id"), 10, 64)
	if dateFrom := parseAdminDate(r.URL.Query().Get("date_from")); dateFrom != nil {
		filter.DateFrom = dateFrom
	}
	if dateTo := parseAdminDate(r.URL.Query().Get("date_to")); dateTo != nil {
		end := dateTo.AddDate(0, 0, 1)
		filter.DateTo = &end
	}

	logs, total, err := h.repo.AuditLogSearch(r.Context(), filter)
	if err != nil {
		h.Render(w, r, "admin_audit_logs.html", map[string]any{
			"Title":  "Auditoria",
			"Active": "audit",
			"Error":  "Nao foi possivel carregar a auditoria",
		})
		return
	}
	users, _ := h.repo.UserList(r.Context())
	h.Render(w, r, "admin_audit_logs.html", map[string]any{
		"Title":            "Auditoria",
		"Active":           "audit",
		"Logs":             logs,
		"Users":            users,
		"Actions":          auditLogActions(),
		"EntityTypes":      auditLogEntityTypes(),
		"FilterUserID":     filter.UserID,
		"FilterAction":     filter.Action,
		"FilterEntityType": filter.EntityType,
		"FilterEntityID":   filter.EntityID,
		"FilterDateFrom":   strings.TrimSpace(r.URL.Query().Get("date_from")),
		"FilterDateTo":     strings.TrimSpace(r.URL.Query().Get("date_to")),
		"Page":             page,
		"Total":            total,
		"TotalPages":       (total + limit - 1) / limit,
		"FilterQueryPath":  auditLogFilterPath(r.URL.Query()),
	})
}

func parseAdminDate(value string) *time.Time {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return nil
	}
	return &t
}

func auditLogActions() []string {
	return []string{"create", "update", "delete", "publish", "password_update", "password_reset_requested", "password_reset_completed", "password_reset_email_sent", "password_reset_email_error", "password_reset_email_not_configured"}
}

func auditLogEntityTypes() []string {
	return []string{"post", "user", "media", "store", "influencer", "banner", "promotion", "event", "classified", "neighborhood", "settings"}
}

func auditLogFilterPath(values url.Values) string {
	copyValues := url.Values{}
	for key, vals := range values {
		if key == "page" {
			continue
		}
		for _, value := range vals {
			if strings.TrimSpace(value) != "" && value != "0" {
				copyValues.Add(key, value)
			}
		}
	}
	encoded := copyValues.Encode()
	if encoded == "" {
		return ""
	}
	return "&" + encoded
}

func (h *Handler) APITrackMetric(w http.ResponseWriter, r *http.Request) {
	metricType := r.PathValue("type")
	entityType := r.URL.Query().Get("entity_type")
	entityID, _ := strconv.ParseInt(r.URL.Query().Get("entity_id"), 10, 64)

	go h.repo.MetricTrack(h.metricContext(r), &model.Metric{
		MetricType: metricType,
		EntityType: entityType,
		EntityID:   entityID,
		IPAddress:  middleware.ClientIP(r),
		UserAgent:  r.UserAgent(),
		Referrer:   r.Referer(),
	})

	w.Header().Set("Content-Type", "image/gif")
	w.Write([]byte("GIF89a\x01\x00\x01\x00\x80\x00\x00\xff\xff\xff\x00\x00\x00!\xf9\x04\x01\x00\x00\x00\x00,\x00\x00\x00\x00\x01\x00\x01\x00\x00\x02\x02D\x01\x00;"))
}

func metricLabel(metricType string) string {
	labels := map[string]string{
		"post_view":         "Visualizacoes de materias",
		"store_view":        "Visualizacoes de lojas",
		"promo_click":       "Cliques em promocoes",
		"banner_click":      "Cliques em banners",
		"banner_impression": "Impressoes de banners",
		"influencer_view":   "Visualizacoes de influencers",
	}
	if label, ok := labels[metricType]; ok {
		return label
	}
	return metricType
}
