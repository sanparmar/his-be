package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ListFilter struct {
	Department string
	DoctorID   string
	Status     *AppointmentStatus
	Date       *time.Time
	Offset     int
	Limit      int
}

type SearchFilter struct {
	Text   string
	Date   *time.Time
	Status *AppointmentStatus
	Offset int
	Limit  int
}

type AppointmentRepository interface {
	Create(ctx context.Context, appointment *Appointment) error
	GetByID(ctx context.Context, tenantID uuid.UUID, appointmentID string) (*Appointment, error)
	Update(ctx context.Context, appointment *Appointment) error
	List(ctx context.Context, tenantID uuid.UUID, filter ListFilter) ([]*Appointment, int64, error)
	Search(ctx context.Context, tenantID uuid.UUID, filter SearchFilter) ([]*Appointment, int64, error)
}
