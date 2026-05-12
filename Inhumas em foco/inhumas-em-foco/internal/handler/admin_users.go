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
	h.Render(w, r, "admin_users.html", map[string]any{"Users": users})
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
