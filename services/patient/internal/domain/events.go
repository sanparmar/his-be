package domain

import (
	"time"

	"github.com/google/uuid"
)

// ─── Domain Events ────────────────────────────────────────────────

// DomainEvent is the base interface for all domain events.
// Events are published to Kafka for other bounded contexts to consume.
type DomainEvent interface {
	EventType() string
	Timestamp() time.Time
	AggregateID() uuid.UUID
	AggregateTenantID() uuid.UUID
	CorrelationID() string
}

// PatientRegistered is published when a new patient is registered.
// Kafka topic: his.patient.patient.registered
type PatientRegistered struct {
	PatientID        uuid.UUID         `json:"patient_id"`
	TenantID         uuid.UUID         `json:"tenant_id"`
	MRN              string            `json:"mrn"`
	Name             string            `json:"name"`
	DOB              time.Time         `json:"dob"`
	Gender           Gender            `json:"gender"`
	BloodGroup       BloodGroup        `json:"blood_group"`
	Phone            string            `json:"phone"`
	Email            string            `json:"email"`
	Address          *Address          `json:"address,omitempty"`
	EmergencyContact *EmergencyContact `json:"emergency_contact,omitempty"`
	Status           PatientStatus     `json:"status"`
	CorrelID         string            `json:"correlation_id"`
	RecordedAt       time.Time         `json:"recorded_at"`
	CreatedBy        uuid.UUID         `json:"created_by"`
}

func (e *PatientRegistered) EventType() string            { return "his.patient.patient.registered" }
func (e *PatientRegistered) Timestamp() time.Time         { return e.RecordedAt }
func (e *PatientRegistered) AggregateID() uuid.UUID       { return e.PatientID }
func (e *PatientRegistered) AggregateTenantID() uuid.UUID { return e.TenantID }
func (e *PatientRegistered) CorrelationID() string        { return e.CorrelID }

// PatientUpdated is published when a patient record is updated.
// Kafka topic: his.patient.patient.updated
type PatientUpdated struct {
	PatientID       uuid.UUID     `json:"patient_id"`
	TenantID        uuid.UUID     `json:"tenant_id"`
	MRN             string        `json:"mrn"`
	Name            string        `json:"name"`
	Phone           string        `json:"phone"`
	Email           string        `json:"email"`
	Address         *Address      `json:"address,omitempty"`
	Insurance       *Insurance    `json:"insurance,omitempty"`
	HomeMedications []Medication  `json:"home_medications,omitempty"`
	Allergies       []string      `json:"allergies,omitempty"`
	VitalSigns      *VitalSigns   `json:"vital_signs,omitempty"`
	Status          PatientStatus `json:"status"`
	ChangedFields   []string      `json:"changed_fields"`
	CorrelID        string        `json:"correlation_id"`
	RecordedAt      time.Time     `json:"recorded_at"`
	UpdatedBy       uuid.UUID     `json:"updated_by"`
}

func (e *PatientUpdated) EventType() string            { return "his.patient.patient.updated" }
func (e *PatientUpdated) Timestamp() time.Time         { return e.RecordedAt }
func (e *PatientUpdated) AggregateID() uuid.UUID       { return e.PatientID }
func (e *PatientUpdated) AggregateTenantID() uuid.UUID { return e.TenantID }
func (e *PatientUpdated) CorrelationID() string        { return e.CorrelID }

// PatientArchived is published when a patient is archived (soft deleted).
// Kafka topic: his.patient.patient.archived
type PatientArchived struct {
	PatientID  uuid.UUID `json:"patient_id"`
	TenantID   uuid.UUID `json:"tenant_id"`
	MRN        string    `json:"mrn"`
	CorrelID   string    `json:"correlation_id"`
	RecordedAt time.Time `json:"recorded_at"`
	ArchivedBy uuid.UUID `json:"archived_by"`
}

func (e *PatientArchived) EventType() string            { return "his.patient.patient.archived" }
func (e *PatientArchived) Timestamp() time.Time         { return e.RecordedAt }
func (e *PatientArchived) AggregateID() uuid.UUID       { return e.PatientID }
func (e *PatientArchived) AggregateTenantID() uuid.UUID { return e.TenantID }
func (e *PatientArchived) CorrelationID() string        { return e.CorrelID }

// PatientActivated is published when a patient is activated.
// Kafka topic: his.patient.patient.activated
type PatientActivated struct {
	PatientID       uuid.UUID `json:"patient_id"`
	TenantID        uuid.UUID `json:"tenant_id"`
	MRN             string    `json:"mrn"`
	Status          string    `json:"status"`
	CorrelID        string    `json:"correlation_id"`
	ActivatedBy     uuid.UUID `json:"activated_by"`
	ActivatedAtTime time.Time `json:"activated_at"`
}

func (e *PatientActivated) EventType() string            { return "his.patient.patient.activated" }
func (e *PatientActivated) Timestamp() time.Time         { return e.ActivatedAtTime }
func (e *PatientActivated) AggregateID() uuid.UUID       { return e.PatientID }
func (e *PatientActivated) AggregateTenantID() uuid.UUID { return e.TenantID }
func (e *PatientActivated) CorrelationID() string        { return e.CorrelID }
