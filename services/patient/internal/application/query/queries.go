package query

import (
	"context"

	"github.com/google/uuid"

	"github.com/deloitte-us-consulting/his-be/services/patient/internal/domain"
)

// GetPatientByMRNQuery represents the query to fetch a patient by MRN.
type GetPatientByMRNQuery struct {
	TenantID uuid.UUID
	MRN      string
}

// GetPatientByMRNHandler handles the GetPatientByMRN query.
type GetPatientByMRNHandler struct {
	patientService *domain.PatientService
}

// NewGetPatientByMRNHandler creates a new handler.
func NewGetPatientByMRNHandler(svc *domain.PatientService) *GetPatientByMRNHandler {
	return &GetPatientByMRNHandler{patientService: svc}
}

// Handle executes the query.
func (h *GetPatientByMRNHandler) Handle(ctx context.Context, query *GetPatientByMRNQuery) (*domain.Patient, error) {
	return h.patientService.GetPatientByMRN(ctx, query.TenantID, query.MRN)
}

// SearchPatientsQuery represents the query to search patients.
type SearchPatientsQuery struct {
	TenantID uuid.UUID
	Query    string // search term (name, MRN, phone)
	Offset   int    // pagination offset
	Limit    int    // pagination limit
}

// SearchPatientsHandler handles the SearchPatients query.
type SearchPatientsHandler struct {
	patientService *domain.PatientService
}

// NewSearchPatientsHandler creates a new handler.
func NewSearchPatientsHandler(svc *domain.PatientService) *SearchPatientsHandler {
	return &SearchPatientsHandler{patientService: svc}
}

// Handle executes the query.
func (h *SearchPatientsHandler) Handle(ctx context.Context, query *SearchPatientsQuery) ([]*domain.Patient, int64, error) {
	return h.patientService.SearchPatients(ctx, query.TenantID, query.Query, query.Offset, query.Limit)
}

// ListPatientsQuery represents the query to list patients with pagination.
type ListPatientsQuery struct {
	TenantID uuid.UUID
	Offset   int // pagination offset
	Limit    int // pagination limit
}

// ListPatientsHandler handles the ListPatients query.
type ListPatientsHandler struct {
	patientService *domain.PatientService
}

// NewListPatientsHandler creates a new handler.
func NewListPatientsHandler(svc *domain.PatientService) *ListPatientsHandler {
	return &ListPatientsHandler{patientService: svc}
}

// Handle executes the query.
func (h *ListPatientsHandler) Handle(ctx context.Context, query *ListPatientsQuery) ([]*domain.Patient, int64, error) {
	return h.patientService.ListPatients(ctx, query.TenantID, query.Offset, query.Limit)
}
