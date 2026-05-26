package handler

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
	"inhumas-em-foco/internal/repository"
)

func (h *Handler) AdminTenants(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermTenantsManage); !ok {
		return
	}
	tenants, _ := h.repo.TenantList(r.Context())
	h.Render(w, r, "admin_tenants.html", map[string]any{
		"Title":   "Portais",
		"Active":  "tenants",
		"Tenants": tenants,
	})
}

func (h *Handler) AdminTenantCreate(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermTenantsManage)
	if !ok {
		return
	}
	tenant := &model.Tenant{
		Name:          r.FormValue("name"),
		Slug:          r.FormValue("slug"),
		Status:        r.FormValue("status"),
		PrimaryDomain: r.FormValue("primary_domain"),
	}
	if tenant.Name == "" || tenant.Slug == "" {
		h.renderTenantsWithError(w, r, "Nome e slug sao obrigatorios")
		return
	}
	if err := h.repo.TenantCreate(r.Context(), tenant); err != nil {
		h.renderTenantsWithError(w, r, "Nao foi possivel criar o portal")
		return
	}
	if err := h.provisionTenant(r.Context(), tenant); err != nil {
		h.renderTenantsWithError(w, r, "Portal criado, mas nao foi possivel concluir o provisionamento inicial")
		return
	}
	h.auditAdminAction(r, currentUser, "create", "tenant", auditEntityID(tenant.ID), map[string]any{
		"name": tenant.Name,
		"slug": tenant.Slug,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/tenants", http.StatusSeeOther)
}

func (h *Handler) provisionTenant(ctx context.Context, tenant *model.Tenant) error {
	if strings.TrimSpace(tenant.PrimaryDomain) != "" {
		if err := h.repo.TenantDomainCreate(ctx, &model.TenantDomain{
			TenantID:  tenant.ID,
			Domain:    tenant.PrimaryDomain,
			IsPrimary: true,
		}); err != nil {
			return err
		}
	}
	settings := repository.DefaultPortalSettings()
	settings.TenantID = tenant.ID
	settings.SiteName = tenant.Name
	settings.SEOTitle = tenant.Name
	settings.SEODescription = "Portal " + tenant.Name
	settings.ContactEmail = provisionContactEmail(tenant)
	if err := h.repo.PortalSettingsUpdate(ctx, &settings); err != nil {
		return err
	}
	defaultFeatures := []model.TenantFeature{
		{TenantID: tenant.ID, Feature: "automation", Enabled: false},
		{TenantID: tenant.ID, Feature: "media", Enabled: true},
		{TenantID: tenant.ID, Feature: "commercial", Enabled: true},
	}
	for i := range defaultFeatures {
		if err := h.repo.TenantFeatureUpsert(ctx, &defaultFeatures[i]); err != nil {
			return err
		}
	}
	return nil
}

func provisionContactEmail(tenant *model.Tenant) string {
	domain := strings.TrimSpace(tenant.PrimaryDomain)
	if domain == "" {
		return "contato@" + strings.TrimSpace(tenant.Slug) + ".local"
	}
	domain = strings.TrimPrefix(domain, "https://")
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimSuffix(domain, "/")
	return "contato@" + domain
}

func (h *Handler) AdminTenantEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermTenantsManage); !ok {
		return
	}
	tenant := h.tenantFromPath(w, r)
	if tenant == nil {
		return
	}
	h.renderTenantForm(w, r, tenant, "")
}

func (h *Handler) AdminTenantUpdate(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermTenantsManage)
	if !ok {
		return
	}
	tenant := h.tenantFromPath(w, r)
	if tenant == nil {
		return
	}
	tenant.Name = r.FormValue("name")
	tenant.Slug = r.FormValue("slug")
	tenant.Status = r.FormValue("status")
	tenant.PrimaryDomain = r.FormValue("primary_domain")
	if tenant.Name == "" || tenant.Slug == "" {
		h.renderTenantForm(w, r, tenant, "Nome e slug sao obrigatorios")
		return
	}
	if err := h.repo.TenantUpdate(r.Context(), tenant); err != nil {
		h.renderTenantForm(w, r, tenant, "Nao foi possivel atualizar o portal")
		return
	}
	h.auditAdminAction(r, currentUser, "update", "tenant", auditEntityID(tenant.ID), map[string]any{
		"name": tenant.Name,
		"slug": tenant.Slug,
	})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/tenants/"+strconv.FormatInt(tenant.ID, 10)+"/edit", http.StatusSeeOther)
}

func (h *Handler) AdminTenantDomainCreate(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermTenantsManage)
	if !ok {
		return
	}
	tenant := h.tenantFromPath(w, r)
	if tenant == nil {
		return
	}
	domain := &model.TenantDomain{
		TenantID:  tenant.ID,
		Domain:    r.FormValue("domain"),
		IsPrimary: r.FormValue("is_primary") == "on",
	}
	if domain.Domain == "" {
		h.renderTenantForm(w, r, tenant, "Dominio e obrigatorio")
		return
	}
	if err := h.repo.TenantDomainCreate(r.Context(), domain); err != nil {
		h.renderTenantForm(w, r, tenant, "Nao foi possivel adicionar o dominio")
		return
	}
	h.auditAdminAction(r, currentUser, "create", "tenant_domain", auditEntityID(domain.ID), map[string]any{"tenant_id": tenant.ID, "domain": domain.Domain})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/tenants/"+strconv.FormatInt(tenant.ID, 10)+"/edit", http.StatusSeeOther)
}

func (h *Handler) AdminTenantDomainDelete(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermTenantsManage)
	if !ok {
		return
	}
	tenant := h.tenantFromPath(w, r)
	if tenant == nil {
		return
	}
	domainID, _ := strconv.ParseInt(r.PathValue("domainID"), 10, 64)
	_ = h.repo.TenantDomainDelete(r.Context(), tenant.ID, domainID)
	h.auditAdminAction(r, currentUser, "delete", "tenant_domain", auditEntityID(domainID), map[string]any{"tenant_id": tenant.ID})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/tenants/"+strconv.FormatInt(tenant.ID, 10)+"/edit", http.StatusSeeOther)
}

func (h *Handler) AdminTenantFeatureUpsert(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermTenantsManage)
	if !ok {
		return
	}
	tenant := h.tenantFromPath(w, r)
	if tenant == nil {
		return
	}
	var limit *int64
	if raw := r.FormValue("limit_value"); raw != "" {
		if value, err := strconv.ParseInt(raw, 10, 64); err == nil {
			limit = &value
		}
	}
	feature := &model.TenantFeature{
		TenantID: tenant.ID,
		Feature:  r.FormValue("feature"),
		Enabled:  r.FormValue("enabled") == "on",
		Limit:    limit,
	}
	if feature.Feature == "" {
		h.renderTenantForm(w, r, tenant, "Feature e obrigatoria")
		return
	}
	if err := h.repo.TenantFeatureUpsert(r.Context(), feature); err != nil {
		h.renderTenantForm(w, r, tenant, "Nao foi possivel salvar a feature")
		return
	}
	h.auditAdminAction(r, currentUser, "upsert", "tenant_feature", auditEntityID(feature.ID), map[string]any{"tenant_id": tenant.ID, "feature": feature.Feature})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/tenants/"+strconv.FormatInt(tenant.ID, 10)+"/edit", http.StatusSeeOther)
}

func (h *Handler) AdminTenantUserUpsert(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermTenantsManage)
	if !ok {
		return
	}
	tenant := h.tenantFromPath(w, r)
	if tenant == nil {
		return
	}
	userID, _ := strconv.ParseInt(r.FormValue("user_id"), 10, 64)
	userEmail := r.FormValue("user_email")
	if userID <= 0 && userEmail != "" {
		user, err := h.repo.UserGetAnyByEmail(r.Context(), userEmail)
		if err != nil || user == nil {
			h.renderTenantForm(w, r, tenant, "Usuario nao encontrado pelo email informado")
			return
		}
		userID = user.ID
	}
	tenantUser := &model.TenantUser{
		TenantID: tenant.ID,
		UserID:   userID,
		Role:     model.UserRole(r.FormValue("role")),
		Active:   r.FormValue("active") == "on",
	}
	if tenantUser.UserID <= 0 {
		h.renderTenantForm(w, r, tenant, "Informe o email ou ID do usuario")
		return
	}
	if err := h.repo.TenantUserUpsert(r.Context(), tenantUser); err != nil {
		h.renderTenantForm(w, r, tenant, "Nao foi possivel salvar o vinculo do usuario")
		return
	}
	h.auditAdminAction(r, currentUser, "upsert", "tenant_user", auditEntityID(tenantUser.ID), map[string]any{"tenant_id": tenant.ID, "user_id": userID, "role": tenantUser.Role})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/tenants/"+strconv.FormatInt(tenant.ID, 10)+"/edit", http.StatusSeeOther)
}

func (h *Handler) AdminTenantUserDelete(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermTenantsManage)
	if !ok {
		return
	}
	tenant := h.tenantFromPath(w, r)
	if tenant == nil {
		return
	}
	userID, _ := strconv.ParseInt(r.PathValue("userID"), 10, 64)
	_ = h.repo.TenantUserDeleteForTenant(r.Context(), tenant.ID, userID)
	h.auditAdminAction(r, currentUser, "delete", "tenant_user", auditEntityID(userID), map[string]any{"tenant_id": tenant.ID})
	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/tenants/"+strconv.FormatInt(tenant.ID, 10)+"/edit", http.StatusSeeOther)
}

func (h *Handler) renderTenantsWithError(w http.ResponseWriter, r *http.Request, message string) {
	tenants, _ := h.repo.TenantList(r.Context())
	h.Render(w, r, "admin_tenants.html", map[string]any{
		"Title":   "Portais",
		"Active":  "tenants",
		"Tenants": tenants,
		"Error":   message,
	})
}

func (h *Handler) renderTenantForm(w http.ResponseWriter, r *http.Request, tenant *model.Tenant, errorMessage string) {
	domains, _ := h.repo.TenantDomainList(r.Context(), tenant.ID)
	features, _ := h.repo.TenantFeatureListForTenant(r.Context(), tenant.ID)
	users, _ := h.repo.TenantUserListForTenant(r.Context(), tenant.ID)
	allUsers, _ := h.repo.UserListAny(r.Context(), 200)
	h.Render(w, r, "admin_tenant_form.html", map[string]any{
		"Title":       "Editar Portal",
		"Active":      "tenants",
		"Tenant":      tenant,
		"Domains":     domains,
		"Features":    features,
		"TenantUsers": users,
		"AllUsers":    allUsers,
		"Error":       errorMessage,
	})
}

func (h *Handler) tenantFromPath(w http.ResponseWriter, r *http.Request) *model.Tenant {
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	tenant, err := h.repo.TenantGetByID(r.Context(), id)
	if err != nil || tenant == nil {
		http.NotFound(w, r)
		return nil
	}
	return tenant
}
