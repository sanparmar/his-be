package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/application"
	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/middleware"
)

type AuthHandler struct {
	loginUseCase           *application.LoginUseCase
	refreshUseCase         *application.RefreshUseCase
	logoutUseCase          *application.LogoutUseCase
	meUseCase              *application.MeUseCase
	assignRoleUseCase      *application.AssignRoleUseCase
	revokeRoleUseCase      *application.RevokeRoleUseCase
	listUserPermsUseCase   *application.ListUserPermissionsUseCase
	getEffectivePermsUseCase *application.GetEffectivePermissionsUseCase
	listRolesUseCase       *application.ListRolesUseCase
	listPermsUseCase       *application.ListPermissionsUseCase
}

func NewAuthHandler(
	loginUseCase *application.LoginUseCase,
	refreshUseCase *application.RefreshUseCase,
	logoutUseCase *application.LogoutUseCase,
	meUseCase *application.MeUseCase,
	assignRoleUseCase *application.AssignRoleUseCase,
	revokeRoleUseCase *application.RevokeRoleUseCase,
	listUserPermsUseCase *application.ListUserPermissionsUseCase,
	getEffectivePermsUseCase *application.GetEffectivePermissionsUseCase,
	listRolesUseCase *application.ListRolesUseCase,
	listPermsUseCase *application.ListPermissionsUseCase,
) *AuthHandler {
	return &AuthHandler{
		loginUseCase:           loginUseCase,
		refreshUseCase:         refreshUseCase,
		logoutUseCase:          logoutUseCase,
		meUseCase:              meUseCase,
		assignRoleUseCase:      assignRoleUseCase,
		revokeRoleUseCase:      revokeRoleUseCase,
		listUserPermsUseCase:   listUserPermsUseCase,
		getEffectivePermsUseCase: getEffectivePermsUseCase,
		listRolesUseCase:       listRolesUseCase,
		listPermsUseCase:       listPermsUseCase,
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
	rawToken := r.Header.Get("Authorization")
	token := stripBearerPrefix(rawToken)
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

func stripBearerPrefix(token string) string {
	if len(token) > 7 && token[:7] == "Bearer " {
		return token[7:]
	}
	return token
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	roles := make([]string, len(user.Roles))
	for i, r := range user.Roles {
		roles[i] = r.Name
	}

	var orgID, hospID uuid.UUID
	if user.OrganizationID != nil {
		orgID = *user.OrganizationID
	}
	if user.HospitalID != nil {
		hospID = *user.HospitalID
	}

	resp := application.UserResponse{
		ID:             user.ID,
		Username:       user.Username,
		Email:          user.Email,
		TenantID:       user.TenantID,
		OrganizationID: orgID,
		HospitalID:     hospID,
		Roles:          roles,
		Permissions:    user.Permissions,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) AssignRole(w http.ResponseWriter, r *http.Request) {
	var req application.AssignRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.assignRoleUseCase.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *AuthHandler) RevokeRole(w http.ResponseWriter, r *http.Request) {
	var req application.RevokeRoleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err := h.revokeRoleUseCase.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]bool{"success": true})
}

func (h *AuthHandler) ListUserPermissions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	resp, err := h.listUserPermsUseCase.Execute(r.Context(), user.ID, user.TenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) GetEffectivePermissions(w http.ResponseWriter, r *http.Request) {
	user := middleware.GetUserFromContext(r.Context())
	if user == nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	perms, err := h.getEffectivePermsUseCase.Execute(r.Context(), user.ID, user.TenantID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":      user.ID,
		"tenant_id":    user.TenantID,
		"permissions":  perms,
	})
}

func (h *AuthHandler) ListRoles(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	req := application.ListRolesRequest{Category: category}

	resp, err := h.listRolesUseCase.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *AuthHandler) ListPermissions(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	resource := r.URL.Query().Get("resource")
	req := application.ListPermissionsRequest{Category: category, Resource: resource}

	resp, err := h.listPermsUseCase.Execute(r.Context(), req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}