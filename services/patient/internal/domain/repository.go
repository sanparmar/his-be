package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// ─── Repository Interface (Port) ──────────────────────────────────

// PatientRepository defines the contract for patient persistence.
// Implements Repository Pattern from Domain-Driven Design.
// Adapter implementations: PostgreSQL, MongoDB, etc.
// Note: All methods are tenant-scoped for multi-tenancy isolation.
type PatientRepository interface {
	// Create persists a new patient
	Create(ctx context.Context, patient *Patient) error

	// GetByID retrieves patient by UUID (tenant-scoped)
	// Returns ErrNotFound if patient does not exist
	GetByID(ctx context.Context, tenantID, patientID uuid.UUID) (*Patient, error)

	// GetByMRN retrieves patient by Medical Record Number (tenant-scoped)
	// Returns ErrNotFound if patient does not exist
	GetByMRN(ctx context.Context, tenantID uuid.UUID, mrn string) (*Patient, error)

	// Update persists patient changes
	Update(ctx context.Context, patient *Patient) error

	// Delete performs soft-delete (sets deleted_at and status=ARCHIVED)
	Delete(ctx context.Context, tenantID, patientID uuid.UUID, deletedBy uuid.UUID) error

	// List retrieves patients with pagination (tenant-scoped)
	// offset: pagination offset (0-based)
	// limit: max records to return
	// statusFilter: optional status filter (if nil, returns ACTIVE and INACTIVE only, excludes ARCHIVED)
	List(ctx context.Context, tenantID uuid.UUID, offset, limit int, statusFilter *PatientStatus) ([]*Patient, int64, error)

	// Search performs full-text search on patient records (tenant-scoped)
	// query: search term (searches name, phone, mrn, email, aadhaar_last_four)
	// offset: pagination offset
	// limit: max records to return
	// Returns results and total count matching query
	Search(ctx context.Context, tenantID uuid.UUID, query string, offset, limit int) ([]*Patient, int64, error)

	// GetAuditLog retrieves change history for a patient (HIPAA compliance)
	GetAuditLog(ctx context.Context, tenantID, patientID uuid.UUID, offset, limit int) ([]*AuditLog, error)
}

// ─── Audit Log (for compliance & debugging) ───────────────────────

// AuditLog records all mutations to a patient record (HIPAA compliance).
// Every mutation (CREATE, UPDATE, DELETE) creates an audit log entry.
type AuditLog struct {
	ID            uuid.UUID              `json:"id"`
	PatientID     uuid.UUID              `json:"patient_id"`
	TenantID      uuid.UUID              `json:"tenant_id"`
	Action        string                 `json:"action"` // CREATE, UPDATE, DELETE, ARCHIVE
	ChangedFields []string               `json:"changed_fields"`
	OldValues     map[string]interface{} `json:"old_values"`
	NewValues     map[string]interface{} `json:"new_values"`
	ChangedBy     uuid.UUID              `json:"changed_by"`
	CorrelationID string                 `json:"correlation_id"` // For distributed tracing
	ChangedAt     time.Time              `json:"changed_at"`
}
