package router

import (
	"net/http"

	"github.com/gorilla/mux"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/middleware"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/transport/http/handlers"
)

func NewRouter(
	h *handlers.AuthHandler,
	authMiddleware *middleware.AuthorizationMiddleware,
) *mux.Router {
	r := mux.NewRouter()

	api := r.PathPrefix("/api/v1").Subrouter()

	api.HandleFunc("/auth/login", h.Login).Methods(http.MethodPost)
	api.HandleFunc("/auth/refresh", h.Refresh).Methods(http.MethodPost)

	protected := api.PathPrefix("").Subrouter()
	protected.Use(authMiddleware.RequireAuth)

	protected.HandleFunc("/auth/logout", h.Logout).Methods(http.MethodPost)
	protected.HandleFunc("/auth/me", h.Me).Methods(http.MethodGet)

	rbac := protected.PathPrefix("/rbac").Subrouter()

	rbac.HandleFunc("/roles", h.ListRoles).Methods(http.MethodGet)
	rbac.HandleFunc("/permissions", h.ListPermissions).Methods(http.MethodGet)
	rbac.HandleFunc("/user-permissions", h.ListUserPermissions).Methods(http.MethodGet)
	rbac.HandleFunc("/effective-permissions", h.GetEffectivePermissions).Methods(http.MethodGet)

	admin := rbac.PathPrefix("").Subrouter()
	admin.Use(authMiddleware.RequirePermission("role:assign"))
	admin.HandleFunc("/assign-role", h.AssignRole).Methods(http.MethodPost)
	admin.Use(authMiddleware.RequirePermission("role:manage"))
	admin.HandleFunc("/revoke-role", h.RevokeRole).Methods(http.MethodPost)

	return r
}