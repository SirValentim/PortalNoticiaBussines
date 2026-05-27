package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"inhumas-em-foco/internal/model"
)

type adminAnalyticsPeriod struct {
	Key        string
	Label      string
	From       *time.Time
	To         *time.Time
	DateFrom   string
	DateTo     string
	QueryPath  string
	IsFiltered bool
}

func adminPeriodFromRequest(r *http.Request) adminAnalyticsPeriod {
	key := strings.TrimSpace(r.URL.Query().Get("period"))
	if key == "" {
		key = "7d"
	}
	now := time.Now()
	today := startOfAdminDay(now)
	period := adminAnalyticsPeriod{Key: key, Label: "Ultimos 7 dias"}
	switch key {
	case "today":
		from := today
		period.From = &from
		period.Label = "Hoje"
	case "30d":
		from := today.AddDate(0, 0, -29)
		period.From = &from
		period.Label = "Ultimos 30 dias"
	case "month":
		from := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
		period.From = &from
		period.Label = "Este mes"
	case "custom":
		from := parseAdminDate(r.URL.Query().Get("date_from"))
		to := parseAdminDate(r.URL.Query().Get("date_to"))
		if to != nil {
			end := to.AddDate(0, 0, 1)
			to = &end
		}
		period.From = from
		period.To = to
		period.DateFrom = strings.TrimSpace(r.URL.Query().Get("date_from"))
		period.DateTo = strings.TrimSpace(r.URL.Query().Get("date_to"))
		period.Label = "Periodo personalizado"
	default:
		key = "7d"
		from := today.AddDate(0, 0, -6)
		period.Key = key
		period.From = &from
	}
	period.IsFiltered = period.From != nil || period.To != nil
	period.QueryPath = adminPeriodQueryPath(r)
	return period
}

func adminPeriodQueryPath(r *http.Request) string {
	values := r.URL.Query()
	encoded := values.Encode()
	if encoded == "" {
		return ""
	}
	return "?" + encoded
}

func selectedTenantIDFromRequest(r *http.Request) int64 {
	id, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("tenant_id")), 10, 64)
	if id < 0 {
		return 0
	}
	return id
}

func trendLabel(current, previous int) string {
	if previous <= 0 {
		if current > 0 {
			return "novo"
		}
		return "sem comparativo"
	}
	diff := float64(current-previous) / float64(previous) * 100
	if diff > 0 {
		return fmt.Sprintf("+%.0f%%", diff)
	}
	return fmt.Sprintf("%.0f%%", diff)
}

func previousPeriod(from, to *time.Time) (*time.Time, *time.Time) {
	if from == nil || to == nil || !to.After(*from) {
		return nil, nil
	}
	duration := to.Sub(*from)
	prevFrom := from.Add(-duration)
	prevTo := *from
	return &prevFrom, &prevTo
}

func dashboardAlerts(summary model.DashboardSummary, deadJobCount int, topPosts []model.PostMetricSummary) []model.OperationalAlert {
	alerts := []model.OperationalAlert{}
	if deadJobCount > 0 {
		alerts = append(alerts, model.OperationalAlert{
			Level:   "danger",
			Title:   "Jobs com falha",
			Message: "Ha rotinas operacionais aguardando revisao.",
			URL:     "/dead-jobs",
		})
	}
	if summary.ReviewPosts > 0 {
		alerts = append(alerts, model.OperationalAlert{
			Level:   "warning",
			Title:   "Materias em revisao",
			Message: "A fila editorial tem conteudos aguardando aprovacao.",
			URL:     "/posts?status=review",
		})
	}
	if summary.TotalViews == 0 || noPostViews(topPosts) {
		alerts = append(alerts, model.OperationalAlert{
			Level:   "info",
			Title:   "Sem dados suficientes",
			Message: "As metricas de audiencia aparecem quando o tracking registrar visualizacoes reais.",
			URL:     "/metrics",
		})
	}
	if summary.TotalTenants <= 1 {
		alerts = append(alerts, model.OperationalAlert{
			Level:   "info",
			Title:   "Primeiro portal ativo",
			Message: "Inhumas em Foco permanece como tenant inicial da plataforma.",
			URL:     "/tenants",
		})
	}
	return alerts
}

func noPostViews(posts []model.PostMetricSummary) bool {
	for _, post := range posts {
		if post.Views > 0 {
			return false
		}
	}
	return true
}

func startOfAdminDay(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
}
