package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/his-platform/auth-service/internal/application"
)

type AuthHandler struct {
	loginUseCase    *application.LoginUseCase
	refreshUseCase  *application.RefreshUseCase
	logoutUseCase   *application.LogoutUseCase
	meUseCase       *application.MeUseCase
}

func NewAuthHandler(
	loginUseCase *application.LoginUseCase,
	refreshUseCase *application.RefreshUseCase,
	logoutUseCase *application.LogoutUseCase,
	meUseCase *application.MeUseCase,
) *AuthHandler {
	return &AuthHandler{
		loginUseCase:    loginUseCase,
		refreshUseCase:  refreshUseCase,
		logoutUseCase:   logoutUseCase,
		meUseCase:       meUseCase,
	}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req application.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.loginUseCase.Execute(r.Context(), req.Username, req.Password)
	if err != nil {
		// For simplicity, treat all errors as 401. In production, use specific error types.
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req application.RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	resp, err := h.refreshUseCase.Execute(r.Context(), req.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	// Token is usually extracted from header or cookie in a real app
	token := r.Header.Get("Authorization")
	if token == "" {
		http.Error(w, "missing token", http.StatusBadRequest)
		return
	}

	err := h.logoutUseCase.Execute(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	// Token is usually extracted from context (populated by middleware)
	// For this prototype, we assume the token is in the header and we extract it.
	token := r.Header.Get("Authorization")
	
	// In a real app, we would use a middleware to validate the token and 
	// put the user object into the request context.
	
	resp, err := h.meUseCase.Execute(r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
