package http

import (
	"encoding/json"
	"errors"
	"fmt"
	stdhttp "net/http"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/deloitte-us-consulting/his-be/services/appointment/internal/application/command"
	"github.com/deloitte-us-consulting/his-be/services/appointment/internal/application/query"
	"github.com/deloitte-us-consulting/his-be/services/appointment/internal/domain"
)

type Handler struct {
	createHandler       *command.CreateAppointmentHandler
	updateStatusHandler *command.UpdateAppointmentStatusHandler
	listHandler         *query.ListAppointmentsHandler
	searchHandler       *query.SearchAppointmentsHandler
}

func NewHandler(
	createHandler *command.CreateAppointmentHandler,
	updateStatusHandler *command.UpdateAppointmentStatusHandler,
	listHandler *query.ListAppointmentsHandler,
	searchHandler *query.SearchAppointmentsHandler,
) *Handler {
	return &Handler{
		createHandler:       createHandler,
		updateStatusHandler: updateStatusHandler,
		listHandler:         listHandler,
		searchHandler:       searchHandler,
	}
}

func (h *Handler) Register(mux *stdhttp.ServeMux) {
	mux.HandleFunc("/health", h.handleHealth)
	mux.HandleFunc("/appointments/search", h.handleSearchAppointments)
	mux.HandleFunc("/appointments/", h.handleAppointmentByID)
	mux.HandleFunc("/appointments", h.handleAppointments)
}

type createAppointmentRequest struct {
	TenantID        string `json:"tenant_id"`
	PatientMRN      string `json:"patient_mrn"`
	PatientName     string `json:"patient_name"`
	DoctorID        string `json:"doctor_id"`
	DoctorName      string `json:"doctor_name"`
	Department      string `json:"department"`
	AppointmentDate string `json:"appointment_date"`
	SlotTime        string `json:"slot_time"`
	VisitType       string `json:"visit_type"`
	Notes           string `json:"notes"`
	CreatedBy       string `json:"created_by"`
}

type updateStatusRequest struct {
	Status    string `json:"status"`
	UpdatedBy string `json:"updated_by"`
}

func (h *Handler) handleHealth(w stdhttp.ResponseWriter, _ *stdhttp.Request) {
	writeJSON(w, stdhttp.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) handleAppointments(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	switch r.Method {
	case stdhttp.MethodPost:
		h.createAppointment(w, r)
	case stdhttp.MethodGet:
		h.listAppointments(w, r)
	default:
		w.WriteHeader(stdhttp.StatusMethodNotAllowed)
	}
}

func (h *Handler) createAppointment(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	var req createAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, fmt.Errorf("decode request: %w", err))
		return
	}

	tenantID, err := uuid.Parse(req.TenantID)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, fmt.Errorf("tenant_id: %w", err))
		return
	}
	createdBy, err := uuid.Parse(req.CreatedBy)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, fmt.Errorf("created_by: %w", err))
		return
	}

	appointment, err := h.createHandler.Handle(r.Context(), command.CreateAppointmentCommand{
		TenantID:        tenantID,
		Patient:         domain.PatientSummary{MRN: req.PatientMRN, Name: req.PatientName},
		Doctor:          domain.DoctorSummary{ID: req.DoctorID, Name: req.DoctorName},
		Department:      req.Department,
		AppointmentDate: req.AppointmentDate,
		SlotTime:        req.SlotTime,
		VisitType:       domain.VisitType(strings.ToUpper(req.VisitType)),
		Notes:           req.Notes,
		CreatedBy:       createdBy,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusCreated, appointment)
}

func (h *Handler) listAppointments(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}

	status, err := statusFromQuery(r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, err)
		return
	}
	date, err := datePtrFromQuery(r.URL.Query().Get("date"))
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, err)
		return
	}

	appointments, total, err := h.listHandler.Handle(r.Context(), query.ListAppointmentsQuery{
		TenantID:   tenantID,
		Department: r.URL.Query().Get("department"),
		DoctorID:   r.URL.Query().Get("doctor_id"),
		Status:     status,
		Date:       date,
		Offset:     intFromQuery(r.URL.Query().Get("offset"), 0),
		Limit:      intFromQuery(r.URL.Query().Get("limit"), 20),
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusOK, map[string]any{"items": appointments, "total": total})
}

func (h *Handler) handleSearchAppointments(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodGet {
		w.WriteHeader(stdhttp.StatusMethodNotAllowed)
		return
	}

	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}

	status, err := statusFromQuery(r.URL.Query().Get("status"))
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, err)
		return
	}
	date, err := datePtrFromQuery(r.URL.Query().Get("date"))
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, err)
		return
	}

	appointments, total, err := h.searchHandler.Handle(r.Context(), query.SearchAppointmentsQuery{
		TenantID: tenantID,
		Text:     r.URL.Query().Get("q"),
		Date:     date,
		Status:   status,
		Offset:   intFromQuery(r.URL.Query().Get("offset"), 0),
		Limit:    intFromQuery(r.URL.Query().Get("limit"), 20),
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusOK, map[string]any{"items": appointments, "total": total})
}

func (h *Handler) handleAppointmentByID(w stdhttp.ResponseWriter, r *stdhttp.Request) {
	if r.Method != stdhttp.MethodPatch {
		w.WriteHeader(stdhttp.StatusMethodNotAllowed)
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/appointments/")
	parts := strings.Split(path, "/")
	if len(parts) != 2 || parts[0] == "" || parts[1] != "status" {
		w.WriteHeader(stdhttp.StatusNotFound)
		return
	}

	tenantID, ok := tenantIDFromRequest(w, r)
	if !ok {
		return
	}

	var req updateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, stdhttp.StatusBadRequest, fmt.Errorf("decode request: %w", err))
		return
	}
	updatedBy, err := uuid.Parse(req.UpdatedBy)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, fmt.Errorf("updated_by: %w", err))
		return
	}

	appointment, err := h.updateStatusHandler.Handle(r.Context(), command.UpdateAppointmentStatusCommand{
		TenantID:      tenantID,
		AppointmentID: parts[0],
		Status:        domain.AppointmentStatus(strings.ToUpper(req.Status)),
		UpdatedBy:     updatedBy,
	})
	if err != nil {
		writeDomainError(w, err)
		return
	}

	writeJSON(w, stdhttp.StatusOK, appointment)
}

func tenantIDFromRequest(w stdhttp.ResponseWriter, r *stdhttp.Request) (uuid.UUID, bool) {
	value := r.Header.Get("X-Tenant-ID")
	if value == "" {
		value = r.URL.Query().Get("tenant_id")
	}
	tenantID, err := uuid.Parse(value)
	if err != nil {
		writeError(w, stdhttp.StatusBadRequest, fmt.Errorf("tenant_id: %w", err))
		return uuid.Nil, false
	}
	return tenantID, true
}

func statusFromQuery(raw string) (*domain.AppointmentStatus, error) {
	if raw == "" {
		return nil, nil
	}
	status := domain.AppointmentStatus(strings.ToUpper(raw))
	switch status {
	case domain.AppointmentStatusScheduled, domain.AppointmentStatusCheckedIn, domain.AppointmentStatusCompleted, domain.AppointmentStatusCancelled, domain.AppointmentStatusNoShow:
		return &status, nil
	default:
		return nil, fmt.Errorf("status: %w", domain.ErrInvalidAppointment)
	}
}

func datePtrFromQuery(raw string) (*time.Time, error) {
	if raw == "" {
		return nil, nil
	}
	parsed, err := time.Parse("2006-01-02", raw)
	if err != nil {
		return nil, fmt.Errorf("date: %w", err)
	}
	return &parsed, nil
}

func intFromQuery(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	var value int
	_, err := fmt.Sscanf(raw, "%d", &value)
	if err != nil || value < 0 {
		return fallback
	}
	return value
}

func writeDomainError(w stdhttp.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrAppointmentNotFound):
		writeError(w, stdhttp.StatusNotFound, err)
	case errors.Is(err, domain.ErrInvalidAppointment), errors.Is(err, domain.ErrInvalidStatusChange):
		writeError(w, stdhttp.StatusBadRequest, err)
	default:
		writeError(w, stdhttp.StatusInternalServerError, err)
	}
}

func writeError(w stdhttp.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

func writeJSON(w stdhttp.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
