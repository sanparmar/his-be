package domain

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAppointmentRepository struct {
	createFunc  func(context.Context, *Appointment) error
	getByIDFunc func(context.Context, uuid.UUID, string) (*Appointment, error)
	updateFunc  func(context.Context, *Appointment) error
	listFunc    func(context.Context, uuid.UUID, ListFilter) ([]*Appointment, int64, error)
	searchFunc  func(context.Context, uuid.UUID, SearchFilter) ([]*Appointment, int64, error)
}

func (m *mockAppointmentRepository) Create(ctx context.Context, appointment *Appointment) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, appointment)
	}
	return nil
}

func (m *mockAppointmentRepository) GetByID(ctx context.Context, tenantID uuid.UUID, appointmentID string) (*Appointment, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, tenantID, appointmentID)
	}
	return nil, ErrAppointmentNotFound
}

func (m *mockAppointmentRepository) Update(ctx context.Context, appointment *Appointment) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, appointment)
	}
	return nil
}

func (m *mockAppointmentRepository) List(ctx context.Context, tenantID uuid.UUID, filter ListFilter) ([]*Appointment, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, tenantID, filter)
	}
	return nil, 0, nil
}

func (m *mockAppointmentRepository) Search(ctx context.Context, tenantID uuid.UUID, filter SearchFilter) ([]*Appointment, int64, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, tenantID, filter)
	}
	return nil, 0, nil
}

func TestAppointmentServiceCreateAppointment(t *testing.T) {
	repo := &mockAppointmentRepository{}
	service := NewAppointmentService(repo)

	appointment, err := service.CreateAppointment(context.Background(), CreateAppointmentInput{
		TenantID:        uuid.New(),
		Patient:         PatientSummary{MRN: "MRN-001", Name: "John Doe"},
		Doctor:          DoctorSummary{ID: "DOC-1", Name: "Dr. Rao"},
		Department:      "Cardiology",
		AppointmentDate: "2026-07-05",
		SlotTime:        "10:30",
		VisitType:       VisitTypeConsultation,
		Notes:           "Walk-in converted to appointment",
		CreatedBy:       uuid.New(),
	})

	require.NoError(t, err)
	assert.Equal(t, AppointmentStatusScheduled, appointment.Status)
}

func TestAppointmentServiceListAndSearch(t *testing.T) {
	tenantID := uuid.New()
	today := time.Date(2026, time.July, 5, 0, 0, 0, 0, time.UTC)
	expected := []*Appointment{{ID: "APT-1", TenantID: tenantID, AppointmentDate: today, Status: AppointmentStatusScheduled}}
	repo := &mockAppointmentRepository{
		listFunc: func(ctx context.Context, gotTenantID uuid.UUID, filter ListFilter) ([]*Appointment, int64, error) {
			assert.Equal(t, tenantID, gotTenantID)
			assert.Equal(t, 20, filter.Limit)
			return expected, 1, nil
		},
		searchFunc: func(ctx context.Context, gotTenantID uuid.UUID, filter SearchFilter) ([]*Appointment, int64, error) {
			assert.Equal(t, tenantID, gotTenantID)
			assert.Equal(t, "john", filter.Text)
			assert.Equal(t, 5, filter.Limit)
			return expected, 1, nil
		},
	}
	service := NewAppointmentService(repo)

	list, total, err := service.ListAppointments(context.Background(), tenantID, ListFilter{})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, list, 1)

	search, total, err := service.SearchAppointments(context.Background(), tenantID, SearchFilter{Text: "john", Limit: 5})
	require.NoError(t, err)
	assert.Equal(t, int64(1), total)
	assert.Len(t, search, 1)
}

func TestAppointmentServiceUpdateStatus(t *testing.T) {
	tenantID := uuid.New()
	updatedBy := uuid.New()
	repo := &mockAppointmentRepository{
		getByIDFunc: func(ctx context.Context, gotTenantID uuid.UUID, appointmentID string) (*Appointment, error) {
			return &Appointment{ID: appointmentID, TenantID: gotTenantID, Status: AppointmentStatusScheduled}, nil
		},
		updateFunc: func(ctx context.Context, appointment *Appointment) error {
			assert.Equal(t, AppointmentStatusCheckedIn, appointment.Status)
			assert.Equal(t, updatedBy, appointment.UpdatedBy)
			return nil
		},
	}
	service := NewAppointmentService(repo)

	appointment, err := service.UpdateStatus(context.Background(), tenantID, "APT-1", AppointmentStatusCheckedIn, updatedBy)

	require.NoError(t, err)
	assert.Equal(t, AppointmentStatusCheckedIn, appointment.Status)
}
