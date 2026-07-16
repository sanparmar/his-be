package router

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/transport/http/handlers"
)

func NewRouter(h *handlers.AuthHandler) *mux.Router {
	r := mux.NewRouter()

	// API versioning prefix
	api := r.PathPrefix("/api/v1").Subrouter()

	// Auth routes
	api.HandleFunc("/auth/login", h.Login).Methods(http.MethodPost)
	api.HandleFunc("/auth/refresh", h.Refresh).Methods(http.MethodPost)
	api.HandleFunc("/auth/logout", h.Logout).Methods(http.MethodPost)
	api.HandleFunc("/auth/me", h.Me).Methods(http.MethodGet)

	return r
}
