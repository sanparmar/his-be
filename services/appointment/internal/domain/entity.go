package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type VisitType string

const (
	VisitTypeConsultation VisitType = "CONSULTATION"
	VisitTypeFollowUp     VisitType = "FOLLOW_UP"
	VisitTypeProcedure    VisitType = "PROCEDURE"
	VisitTypeTeleconsult  VisitType = "TELECONSULTATION"
)

type AppointmentStatus string

const (
	AppointmentStatusScheduled AppointmentStatus = "SCHEDULED"
	AppointmentStatusCheckedIn AppointmentStatus = "CHECKED_IN"
	AppointmentStatusCompleted AppointmentStatus = "COMPLETED"
	AppointmentStatusCancelled AppointmentStatus = "CANCELLED"
	AppointmentStatusNoShow    AppointmentStatus = "NO_SHOW"
)

type PatientSummary struct {
	MRN  string `json:"mrn" db:"patient_mrn" validate:"required"`
	Name string `json:"name" db:"patient_name" validate:"required"`
}

type DoctorSummary struct {
	ID   string `json:"id" db:"doctor_id" validate:"required"`
	Name string `json:"name" db:"doctor_name" validate:"required"`
}

type Appointment struct {
	ID              string            `json:"id" db:"id" validate:"required"`
	TenantID        uuid.UUID         `json:"tenant_id" db:"tenant_id" validate:"required"`
	Patient         PatientSummary    `json:"patient" validate:"required"`
	Doctor          DoctorSummary     `json:"doctor" validate:"required"`
	Department      string            `json:"department" db:"department" validate:"required"`
	AppointmentDate time.Time         `json:"appointment_date" db:"appointment_date" validate:"required"`
	SlotTime        string            `json:"slot_time" db:"slot_time" validate:"required"`
	VisitType       VisitType         `json:"visit_type" db:"visit_type" validate:"required"`
	Status          AppointmentStatus `json:"status" db:"status" validate:"required"`
	Notes           string            `json:"notes" db:"notes" validate:"omitempty"`
	CreatedAt       time.Time         `json:"created_at" db:"created_at" validate:"required"`
	UpdatedAt       time.Time         `json:"updated_at" db:"updated_at" validate:"required"`
	CreatedBy       uuid.UUID         `json:"created_by" db:"created_by" validate:"required"`
	UpdatedBy       uuid.UUID         `json:"updated_by" db:"updated_by" validate:"required"`
}

type NewAppointmentParams struct {
	TenantID        uuid.UUID
	Patient         PatientSummary
	Doctor          DoctorSummary
	Department      string
	AppointmentDate time.Time
	SlotTime        string
	VisitType       VisitType
	Notes           string
	CreatedBy       uuid.UUID
}

func NewAppointment(params NewAppointmentParams) (*Appointment, error) {
	if params.TenantID == uuid.Nil || params.CreatedBy == uuid.Nil {
		return nil, ErrInvalidAppointment
	}
	if strings.TrimSpace(params.Patient.MRN) == "" || strings.TrimSpace(params.Patient.Name) == "" {
		return nil, fmt.Errorf("patient summary: %w", ErrInvalidAppointment)
	}
	if strings.TrimSpace(params.Doctor.ID) == "" || strings.TrimSpace(params.Doctor.Name) == "" {
		return nil, fmt.Errorf("doctor summary: %w", ErrInvalidAppointment)
	}
	if strings.TrimSpace(params.Department) == "" {
		return nil, fmt.Errorf("department: %w", ErrInvalidAppointment)
	}
	if !isValidVisitType(params.VisitType) {
		return nil, fmt.Errorf("visit type: %w", ErrInvalidAppointment)
	}
	if err := validateSlotTime(params.SlotTime); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	date := normalizeDate(params.AppointmentDate)
	id := fmt.Sprintf("APT-%s-%s", strings.ToUpper(params.TenantID.String()[:8]), uuid.NewString()[:8])

	return &Appointment{
		ID:              id,
		TenantID:        params.TenantID,
		Patient:         params.Patient,
		Doctor:          params.Doctor,
		Department:      strings.TrimSpace(params.Department),
		AppointmentDate: date,
		SlotTime:        params.SlotTime,
		VisitType:       params.VisitType,
		Status:          AppointmentStatusScheduled,
		Notes:           strings.TrimSpace(params.Notes),
		CreatedAt:       now,
		UpdatedAt:       now,
		CreatedBy:       params.CreatedBy,
		UpdatedBy:       params.CreatedBy,
	}, nil
}

func (a *Appointment) UpdateStatus(newStatus AppointmentStatus, updatedBy uuid.UUID) error {
	if updatedBy == uuid.Nil {
		return ErrInvalidAppointment
	}
	if a.Status == newStatus {
		a.UpdatedAt = time.Now().UTC()
		a.UpdatedBy = updatedBy
		return nil
	}
	if !canTransition(a.Status, newStatus) {
		return fmt.Errorf("%w: %s -> %s", ErrInvalidStatusChange, a.Status, newStatus)
	}

	a.Status = newStatus
	a.UpdatedAt = time.Now().UTC()
	a.UpdatedBy = updatedBy
	return nil
}

func normalizeDate(date time.Time) time.Time {
	utc := date.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 0, 0, 0, 0, time.UTC)
}

func validateSlotTime(value string) error {
	if _, err := time.Parse("15:04", strings.TrimSpace(value)); err != nil {
		return fmt.Errorf("slot time: %w", ErrInvalidAppointment)
	}
	return nil
}

func isValidVisitType(value VisitType) bool {
	switch value {
	case VisitTypeConsultation, VisitTypeFollowUp, VisitTypeProcedure, VisitTypeTeleconsult:
		return true
	default:
		return false
	}
}

func canTransition(from, to AppointmentStatus) bool {
	switch from {
	case AppointmentStatusScheduled:
		return to == AppointmentStatusCheckedIn || to == AppointmentStatusCancelled || to == AppointmentStatusNoShow
	case AppointmentStatusCheckedIn:
		return to == AppointmentStatusCompleted || to == AppointmentStatusCancelled
	default:
		return false
	}
}
