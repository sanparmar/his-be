package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewAppointment(t *testing.T) {
	appointment, err := NewAppointment(NewAppointmentParams{
		TenantID:        uuid.New(),
		Patient:         PatientSummary{MRN: "MRN-001", Name: "John Doe"},
		Doctor:          DoctorSummary{ID: "DOC-1", Name: "Dr. Rao"},
		Department:      "Cardiology",
		AppointmentDate: time.Date(2026, time.July, 5, 10, 30, 0, 0, time.FixedZone("IST", 19800)),
		SlotTime:        "10:30",
		VisitType:       VisitTypeConsultation,
		Notes:           "First visit",
		CreatedBy:       uuid.New(),
	})

	require.NoError(t, err)
	assert.NotEmpty(t, appointment.ID)
	assert.Equal(t, AppointmentStatusScheduled, appointment.Status)
	assert.Equal(t, "Cardiology", appointment.Department)
	assert.Equal(t, "MRN-001", appointment.Patient.MRN)
	assert.Equal(t, "10:30", appointment.SlotTime)
	assert.Equal(t, 0, appointment.AppointmentDate.Hour())
}

func TestAppointmentUpdateStatus(t *testing.T) {
	appointment := &Appointment{Status: AppointmentStatusScheduled}

	err := appointment.UpdateStatus(AppointmentStatusCheckedIn, uuid.New())

	require.NoError(t, err)
	assert.Equal(t, AppointmentStatusCheckedIn, appointment.Status)
}

func TestAppointmentUpdateStatusRejectsInvalidTransition(t *testing.T) {
	appointment := &Appointment{Status: AppointmentStatusCompleted}

	err := appointment.UpdateStatus(AppointmentStatusScheduled, uuid.New())

	require.Error(t, err)
	assert.ErrorIs(t, err, ErrInvalidStatusChange)
}
