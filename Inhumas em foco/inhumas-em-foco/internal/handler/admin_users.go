package handler

import (
	"net/http"
	"strconv"

	"inhumas-em-foco/internal/auth"
	"inhumas-em-foco/internal/model"
	usersvc "inhumas-em-foco/internal/users"
)

func (h *Handler) AdminUsers(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermUsersManage); !ok {
		return
	}
	users, _ := h.repo.UserList(r.Context())
	h.Render(w, r, "admin_users.html", map[string]any{
		"Title":  "Usuarios",
		"Active": "users",
		"Users":  users,
	})
}

func (h *Handler) AdminUserCreate(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermUsersManage)
	if !ok {
		return
	}
	u, msg := h.userSvc.Create(r.Context(), usersvc.UserCreateInput{
		Name:     r.FormValue("name"),
		Email:    r.FormValue("email"),
		Password: r.FormValue("password"),
		Role:     model.UserRole(r.FormValue("role")),
		Active:   r.FormValue("active") == "on",
	})
	if msg != "" {
		users, _ := h.repo.UserList(r.Context())
		h.Render(w, r, "admin_users.html", map[string]any{
			"Users": users,
			"Error": msg,
		})
		return
	}
	h.auditAdminAction(r, currentUser, "create", "user", auditEntityID(u.ID), map[string]any{
		"name":   u.Name,
		"email":  u.Email,
		"role":   u.Role,
		"active": u.Active,
	})

	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/users", http.StatusSeeOther)
}

func (h *Handler) AdminUserEdit(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermUsersManage); !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	user, err := h.repo.UserGetByID(r.Context(), id)
	if err != nil || user == nil {
		http.NotFound(w, r)
		return
	}

	h.Render(w, r, "admin_user_form.html", map[string]any{
		"Title":    "Editar Usuario",
		"Active":   "users",
		"EditUser": user,
	})
}

func (h *Handler) AdminUserDetail(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.requirePermission(w, r, auth.PermUsersManage); !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	user, err := h.repo.UserGetByID(r.Context(), id)
	if err != nil || user == nil {
		http.NotFound(w, r)
		return
	}
	var tenantLinks []model.TenantUser
	if h.authSvc.HasPermission(auth.UserFromContext(r.Context()), auth.PermTenantsManage) {
		tenantLinks, _ = h.repo.TenantUserListByUser(r.Context(), user.ID)
	} else if link, _ := h.repo.TenantUserGet(r.Context(), user.ID); link != nil {
		tenantLinks = append(tenantLinks, *link)
	}
	h.Render(w, r, "admin_user_detail.html", map[string]any{
		"Title":       "Usuario",
		"Active":      "users",
		"EditUser":    user,
		"TenantLinks": tenantLinks,
	})
}

func (h *Handler) AdminUserUpdate(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermUsersManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	user, err := h.repo.UserGetByID(r.Context(), id)
	if err != nil || user == nil {
		http.NotFound(w, r)
		return
	}

	if msg := h.userSvc.Update(r.Context(), currentUser, user, usersvc.UserUpdateInput{
		Name:   r.FormValue("name"),
		Email:  r.FormValue("email"),
		Role:   model.UserRole(r.FormValue("role")),
		Active: r.FormValue("active") == "on",
	}); msg != "" {
		h.Render(w, r, "admin_user_form.html", map[string]any{
			"Title":    "Editar Usuario",
			"Active":   "users",
			"EditUser": user,
			"Error":    msg,
		})
		return
	}
	h.auditAdminAction(r, currentUser, "update", "user", auditEntityID(user.ID), map[string]any{
		"name":   user.Name,
		"email":  user.Email,
		"role":   user.Role,
		"active": user.Active,
	})

	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/users", http.StatusSeeOther)
}

func (h *Handler) AdminUserUpdatePassword(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermUsersManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if msg := h.userSvc.UpdatePassword(r.Context(), id, r.FormValue("password"), r.FormValue("password_confirm")); msg != "" {
		users, _ := h.repo.UserList(r.Context())
		h.Render(w, r, "admin_users.html", map[string]any{
			"Users": users,
			"Error": msg,
		})
		return
	}
	h.auditAdminAction(r, currentUser, "password_update", "user", auditEntityID(id), nil)

	http.Redirect(w, r, h.cfg.AdminPathPrefix+"/users", http.StatusSeeOther)
}

func (h *Handler) AdminUserActivate(w http.ResponseWriter, r *http.Request) {
	h.adminUserSetActive(w, r, true)
}

func (h *Handler) AdminUserDeactivate(w http.ResponseWriter, r *http.Request) {
	h.adminUserSetActive(w, r, false)
}

func (h *Handler) adminUserSetActive(w http.ResponseWriter, r *http.Request, active bool) {
	currentUser, ok := h.requirePermission(w, r, auth.PermUsersManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	user, err := h.repo.UserGetByID(r.Context(), id)
	if err != nil || user == nil {
		http.NotFound(w, r)
		return
	}
	if msg := h.userSvc.Update(r.Context(), currentUser, user, usersvc.UserUpdateInput{
		Name:   user.Name,
		Email:  user.Email,
		Role:   user.Role,
		Active: active,
	}); msg != "" {
		h.renderUsersWithMessage(w, r, msg, "")
		return
	}
	action := "activate"
	success := "Conta ativada com sucesso."
	if !active {
		action = "deactivate"
		success = "Conta desativada com sucesso."
	}
	h.auditAdminAction(r, currentUser, action, "user", auditEntityID(user.ID), map[string]any{
		"email":  user.Email,
		"active": active,
	})
	h.renderUsersWithMessage(w, r, "", success)
}

func (h *Handler) AdminUserDelete(w http.ResponseWriter, r *http.Request) {
	currentUser, ok := h.requirePermission(w, r, auth.PermUsersManage)
	if !ok {
		return
	}
	id, _ := strconv.ParseInt(r.PathValue("id"), 10, 64)
	user, err := h.repo.UserGetByID(r.Context(), id)
	if err != nil || user == nil {
		http.NotFound(w, r)
		return
	}
	if msg := h.userSvc.Delete(r.Context(), currentUser, user); msg != "" {
		h.renderUsersWithMessage(w, r, msg, "")
		return
	}
	h.auditAdminAction(r, currentUser, "delete", "user", auditEntityID(user.ID), map[string]any{
		"email": user.Email,
		"mode":  "soft_delete",
	})
	h.renderUsersWithMessage(w, r, "", "Conta excluida com sucesso.")
}

func (h *Handler) renderUsersWithMessage(w http.ResponseWriter, r *http.Request, errorMsg, successMsg string) {
	users, _ := h.repo.UserList(r.Context())
	data := map[string]any{
		"Title":  "Usuarios",
		"Active": "users",
		"Users":  users,
	}
	if errorMsg != "" {
		data["Error"] = errorMsg
	}
	if successMsg != "" {
		data["Success"] = successMsg
	}
	h.Render(w, r, "admin_users.html", data)
}

func (h *Handler) AdminProfile(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	h.Render(w, r, "admin_profile.html", map[string]any{
		"Title":       "Perfil",
		"Active":      "profile",
		"ProfileUser": user,
	})
}

func (h *Handler) AdminProfilePasswordUpdate(w http.ResponseWriter, r *http.Request) {
	user, ok := h.requireCurrentUser(w, r)
	if !ok {
		return
	}
	if msg := h.userSvc.UpdatePassword(r.Context(), user.ID, r.FormValue("password"), r.FormValue("password_confirm")); msg != "" {
		h.Render(w, r, "admin_profile.html", map[string]any{
			"Title":       "Perfil",
			"Active":      "profile",
			"ProfileUser": user,
			"Error":       msg,
		})
		return
	}
	h.auditAdminAction(r, user, "password_update", "user", auditEntityID(user.ID), map[string]any{"scope": "profile"})
	h.Render(w, r, "admin_profile.html", map[string]any{
		"Title":       "Perfil",
		"Active":      "profile",
		"ProfileUser": user,
		"Success":     "Senha atualizada com sucesso.",
	})
}
