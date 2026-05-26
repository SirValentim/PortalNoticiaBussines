package handler

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/automation"
	"inhumas-em-foco/internal/model"
)

func (h *Handler) AdminAutomation(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermAutomationManage); !ok {
		return
	}
	h.renderAutomationDashboard(w, r, automationSuccessMessage(r.URL.Query().Get("success")), "")
}

func (h *Handler) AdminAutomationSourceNew(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermAutomationManage); !ok {
		return
	}
	if !h.requireAutomationFeature(w, r) {
		return
	}
	h.renderAutomationSourceForm(w, r, nil, "")
}

func (h *Handler) AdminAutomationSourceCreate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermAutomationManage)
	if !ok {
		return
	}
	if !h.requireAutomationFeature(w, r) {
		return
	}
	if err := r.ParseForm(); err != nil {
		h.renderAutomationSourceForm(w, r, nil, "Nao foi possivel ler o formulario")
		return
	}
	source := automationSourceFromRequest(r, nil)
	if msg := validateAutomationSource(source); msg != "" {
		h.renderAutomationSourceForm(w, r, source, msg)
		return
	}
	if err := h.repo.AutomationSourceCreate(r.Context(), source); err != nil {
		h.renderAutomationSourceForm(w, r, source, "Nao foi possivel criar a fonte")
		return
	}
	h.auditAdminAction(r, user, "create", "automation_source", auditEntityID(source.ID), map[string]any{
		"name":        source.Name,
		"source_type": source.SourceType,
		"url":         source.URL,
		"active":      source.Active,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/automation?success=source_create", http.StatusSeeOther)
}

func (h *Handler) AdminAutomationSourceEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermAutomationManage); !ok {
		return
	}
	if !h.requireAutomationFeature(w, r) {
		return
	}
	source := h.automationSourceFromPath(w, r)
	if source == nil {
		return
	}
	h.renderAutomationSourceForm(w, r, source, "")
}

func (h *Handler) AdminAutomationSourceUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermAutomationManage)
	if !ok {
		return
	}
	if !h.requireAutomationFeature(w, r) {
		return
	}
	source := h.automationSourceFromPath(w, r)
	if source == nil {
		return
	}
	if err := r.ParseForm(); err != nil {
		h.renderAutomationSourceForm(w, r, source, "Nao foi possivel ler o formulario")
		return
	}
	source = automationSourceFromRequest(r, source)
	if msg := validateAutomationSource(source); msg != "" {
		h.renderAutomationSourceForm(w, r, source, msg)
		return
	}
	if err := h.repo.AutomationSourceUpdate(r.Context(), source); err != nil {
		h.renderAutomationSourceForm(w, r, source, "Nao foi possivel atualizar a fonte")
		return
	}
	h.auditAdminAction(r, user, "update", "automation_source", auditEntityID(source.ID), map[string]any{
		"name":        source.Name,
		"source_type": source.SourceType,
		"url":         source.URL,
		"active":      source.Active,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/automation?success=source_update", http.StatusSeeOther)
}

func (h *Handler) AdminAutomationSourceDelete(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermAutomationManage)
	if !ok {
		return
	}
	if !h.requireAutomationFeature(w, r) {
		return
	}
	source := h.automationSourceFromPath(w, r)
	if source == nil {
		return
	}
	if err := h.repo.AutomationSourceDelete(r.Context(), source.ID); err != nil {
		h.renderAutomationDashboard(w, r, "", "Nao foi possivel excluir a fonte")
		return
	}
	h.auditAdminAction(r, user, "delete", "automation_source", auditEntityID(source.ID), map[string]any{
		"name": source.Name,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/automation?success=source_delete", http.StatusSeeOther)
}

func (h *Handler) AdminAutomationSourceRun(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermAutomationManage)
	if !ok {
		return
	}
	if !h.requireAutomationFeature(w, r) {
		return
	}
	source := h.automationSourceFromPath(w, r)
	if source == nil {
		return
	}
	run, err := automation.NewService(h.repo).RunSource(r.Context(), source.ID)
	if err != nil {
		h.auditAdminAction(r, user, "run_error", "automation_source", auditEntityID(source.ID), map[string]any{
			"name":  source.Name,
			"error": err.Error(),
		})
		h.renderAutomationDashboard(w, r, "", "Execucao concluida com erro: "+err.Error())
		return
	}
	h.auditAdminAction(r, user, "run", "automation_source", auditEntityID(source.ID), map[string]any{
		"name":           source.Name,
		"run_id":         run.ID,
		"drafts_created": run.DraftsCreated,
		"duplicates":     run.Duplicates,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/automation?success=run_source", http.StatusSeeOther)
}

func (h *Handler) AdminAutomationRunAll(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requirePermission(w, r, auth.PermAutomationManage)
	if !ok {
		return
	}
	if !h.requireAutomationFeature(w, r) {
		return
	}
	runs, err := automation.NewService(h.repo).RunAllActive(r.Context())
	h.auditAdminAction(r, user, "run_all", "automation", nil, map[string]any{
		"runs": len(runs),
	})
	if err != nil {
		h.renderAutomationDashboard(w, r, "", "Uma ou mais fontes falharam: "+err.Error())
		return
	}
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/automation?success=run_all", http.StatusSeeOther)
}

func (h *Handler) renderAutomationDashboard(w http.ResponseWriter, r *http.Request, success, errorMessage string) {
	sources, _ := h.repo.AutomationSourceList(r.Context(), false, 200)
	runs, _ := h.repo.AutomationRunList(r.Context(), 40)
	queue, _ := h.repo.AutomationDraftQueue(r.Context(), 30)
	featureEnabled := h.automationFeatureEnabled(r)
	activeCount := 0
	for _, source := range sources {
		if source.Active {
			activeCount++
		}
	}
	settings := h.portalSettings(r.Context())
	h.Render(w, r, "admin_automation.html", map[string]any{
		"Title":                    "Automacao de Noticias",
		"Active":                   "automation",
		"Sources":                  sources,
		"Runs":                     runs,
		"DraftQueue":               queue,
		"SourceCount":              len(sources),
		"ActiveCount":              activeCount,
		"Settings":                 settings,
		"Success":                  success,
		"Error":                    errorMessage,
		"AutomationFeatureEnabled": featureEnabled,
	})
}

func (h *Handler) automationFeatureEnabled(r *http.Request) bool {
	return h.tenantFeatureEnabled(r, "automation", true)
}

func (h *Handler) requireAutomationFeature(w http.ResponseWriter, r *http.Request) bool {
	if h.automationFeatureEnabled(r) {
		return true
	}
	h.renderAutomationDashboard(w, r, "", "Automacao nao habilitada para este portal.")
	return false
}

func (h *Handler) renderAutomationSourceForm(w http.ResponseWriter, r *http.Request, source *model.AutomationSource, errorMessage string) {
	if source == nil {
		source = &model.AutomationSource{Active: true, SourceType: string(model.AutomationSourceRSS)}
	}
	title := "Nova Fonte"
	if source != nil && source.ID > 0 {
		title = "Editar Fonte"
	}
	categories, _ := h.repo.CategoryList(r.Context())
	h.Render(w, r, "admin_automation_source_form.html", map[string]any{
		"Title":      title,
		"Active":     "automation",
		"Source":     source,
		"Categories": categories,
		"Error":      errorMessage,
	})
}

func (h *Handler) automationSourceFromPath(w http.ResponseWriter, r *http.Request) *model.AutomationSource {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	source, err := h.repo.AutomationSourceGetByID(r.Context(), id)
	if err != nil || source == nil {
		http.NotFound(w, r)
		return nil
	}
	return source
}

func automationSourceFromRequest(r *http.Request, source *model.AutomationSource) *model.AutomationSource {
	if source == nil {
		source = &model.AutomationSource{Active: true, SourceType: string(model.AutomationSourceRSS)}
	}
	source.Name = strings.TrimSpace(r.FormValue("name"))
	source.SourceType = normalizeAutomationSourceTypeForHandler(r.FormValue("source_type"))
	source.URL = strings.TrimSpace(r.FormValue("url"))
	source.DefaultCategoryID = automationCategoryIDFromForm(r)
	source.Active = r.FormValue("active") == "on"
	return source
}

func validateAutomationSource(source *model.AutomationSource) string {
	if strings.TrimSpace(source.Name) == "" {
		return "Nome da fonte e obrigatorio"
	}
	parsed, err := url.ParseRequestURI(source.URL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return "URL da fonte invalida"
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "A fonte precisa usar HTTP ou HTTPS"
	}
	return ""
}

func normalizeAutomationSourceTypeForHandler(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case string(model.AutomationSourceOfficial):
		return string(model.AutomationSourceOfficial)
	default:
		return string(model.AutomationSourceRSS)
	}
}

func automationCategoryIDFromForm(r *http.Request) *int64 {
	id, err := strconv.ParseInt(strings.TrimSpace(r.FormValue("default_category_id")), 10, 64)
	if err != nil || id <= 0 {
		return nil
	}
	return &id
}

func automationSuccessMessage(code string) string {
	switch code {
	case "source_create":
		return "Fonte criada com sucesso."
	case "source_update":
		return "Fonte atualizada com sucesso."
	case "source_delete":
		return "Fonte excluida com sucesso."
	case "run_source":
		return "Fonte executada. Rascunhos novos entraram na fila de revisao."
	case "run_all":
		return "Coleta manual executada para todas as fontes ativas."
	default:
		return ""
	}
}
