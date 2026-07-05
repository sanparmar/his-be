package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type AppointmentService struct {
	repo AppointmentRepository
}

func NewAppointmentService(repo AppointmentRepository) *AppointmentService {
	return &AppointmentService{repo: repo}
}

type CreateAppointmentInput struct {
	TenantID        uuid.UUID
	Patient         PatientSummary
	Doctor          DoctorSummary
	Department      string
	AppointmentDate string
	SlotTime        string
	VisitType       VisitType
	Notes           string
	CreatedBy       uuid.UUID
}

func (s *AppointmentService) CreateAppointment(ctx context.Context, input CreateAppointmentInput) (*Appointment, error) {
	appointmentDate, err := parseDate(input.AppointmentDate)
	if err != nil {
		return nil, err
	}

	appointment, err := NewAppointment(NewAppointmentParams{
		TenantID:        input.TenantID,
		Patient:         input.Patient,
		Doctor:          input.Doctor,
		Department:      input.Department,
		AppointmentDate: appointmentDate,
		SlotTime:        input.SlotTime,
		VisitType:       input.VisitType,
		Notes:           input.Notes,
		CreatedBy:       input.CreatedBy,
	})
	if err != nil {
		return nil, err
	}

	if err := s.repo.Create(ctx, appointment); err != nil {
		return nil, fmt.Errorf("create appointment: %w", err)
	}
	return appointment, nil
}

func (s *AppointmentService) ListAppointments(ctx context.Context, tenantID uuid.UUID, filter ListFilter) ([]*Appointment, int64, error) {
	items, total, err := s.repo.List(ctx, tenantID, sanitizeListFilter(filter))
	if err != nil {
		return nil, 0, fmt.Errorf("list appointments: %w", err)
	}
	return items, total, nil
}

func (s *AppointmentService) SearchAppointments(ctx context.Context, tenantID uuid.UUID, filter SearchFilter) ([]*Appointment, int64, error) {
	items, total, err := s.repo.Search(ctx, tenantID, sanitizeSearchFilter(filter))
	if err != nil {
		return nil, 0, fmt.Errorf("search appointments: %w", err)
	}
	return items, total, nil
}

func (s *AppointmentService) UpdateStatus(ctx context.Context, tenantID uuid.UUID, appointmentID string, status AppointmentStatus, updatedBy uuid.UUID) (*Appointment, error) {
	appointment, err := s.repo.GetByID(ctx, tenantID, appointmentID)
	if err != nil {
		return nil, fmt.Errorf("load appointment: %w", err)
	}
	if appointment.TenantID != tenantID {
		return nil, ErrTenantScopeMismatch
	}
	if err := appointment.UpdateStatus(status, updatedBy); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, appointment); err != nil {
		return nil, fmt.Errorf("persist appointment status: %w", err)
	}
	return appointment, nil
}

func parseDate(value string) (time.Time, error) {
	parsed, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("appointment date: %w", ErrInvalidAppointment)
	}
	return parsed, nil
}

func sanitizeListFilter(filter ListFilter) ListFilter {
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	return filter
}

func sanitizeSearchFilter(filter SearchFilter) SearchFilter {
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	if filter.Limit <= 0 {
		filter.Limit = 20
	}
	return filter
}
