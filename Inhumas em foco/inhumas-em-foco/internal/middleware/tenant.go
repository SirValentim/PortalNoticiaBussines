package middleware

import (
	"net/http"

	"inhumas-em-foco/internal/repository"
	tenantctx "inhumas-em-foco/internal/tenant"
)

func ResolveTenant(repo *repository.Repository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if repo == nil {
				next.ServeHTTP(w, r)
				return
			}

			current, err := repo.TenantGetByDomain(r.Context(), tenantctx.DomainFromRequest(r))
			if err != nil {
				http.Error(w, "erro ao resolver tenant", http.StatusInternalServerError)
				return
			}
			if current == nil {
				current, err = repo.TenantGetBySlug(r.Context(), "default")
				if err != nil {
					http.Error(w, "erro ao carregar tenant default", http.StatusInternalServerError)
					return
				}
			}
			if current == nil {
				http.Error(w, "tenant default nao encontrado", http.StatusInternalServerError)
				return
			}
			if current.Status != "active" {
				http.Error(w, "tenant inativo", http.StatusServiceUnavailable)
				return
			}

			next.ServeHTTP(w, r.WithContext(tenantctx.WithTenant(r.Context(), current)))
		})
	}
}
