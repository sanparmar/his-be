package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPatient(t *testing.T) {
	tenantID := uuid.New()
	createdBy := uuid.New()
	dob := time.Date(1990, time.January, 15, 0, 0, 0, 0, time.UTC)
	emergencyContact := &EmergencyContact{Name: "Jane Doe", Relation: "Spouse", Phone: "+919900000000"}

	patient := NewPatient(tenantID, "John Doe", dob, GenderMale, "+919876543210", BloodGroupOPlus, emergencyContact, createdBy)

	require.NotNil(t, patient)
	assert.NotEqual(t, uuid.Nil, patient.ID)
	assert.Equal(t, tenantID, patient.TenantID)
	assert.NotEmpty(t, patient.MRN)
	assert.Equal(t, "John Doe", patient.Name)
	assert.True(t, patient.DOB.Equal(dob))
	assert.Equal(t, GenderMale, patient.Gender)
	assert.Equal(t, BloodGroupOPlus, patient.BloodGroup)
	assert.Equal(t, PatientStatusActive, patient.Status)
	assert.Equal(t, emergencyContact, patient.EmergencyContact)
	assert.Equal(t, createdBy, patient.CreatedBy)
	assert.Equal(t, createdBy, patient.UpdatedBy)
	assert.False(t, patient.CreatedAt.IsZero())
	assert.False(t, patient.UpdatedAt.IsZero())
}

func TestPatientTransitionStatus(t *testing.T) {
	patient := &Patient{Status: PatientStatusActive}
	updatedBy := uuid.New()

	err := patient.TransitionStatus(PatientStatusInactive, updatedBy)

	require.NoError(t, err)
	assert.Equal(t, PatientStatusInactive, patient.Status)
	assert.Equal(t, updatedBy, patient.UpdatedBy)
	assert.False(t, patient.UpdatedAt.IsZero())
}

func TestPatientTransitionStatusRejectsInvalidTransition(t *testing.T) {
	patient := &Patient{Status: PatientStatusArchived}

	err := patient.TransitionStatus(PatientStatusActive, uuid.New())

	require.Error(t, err)
	assert.Equal(t, PatientStatusArchived, patient.Status)
}

func TestPatientUpdateDemographics(t *testing.T) {
	patient := &Patient{
		Name:      "Old Name",
		Phone:     "+911111111111",
		Email:     "old@example.com",
		UpdatedAt: time.Now().Add(-time.Hour),
	}
	updatedBy := uuid.New()
	address := &Address{Street: "123 Street", City: "Mumbai", State: "MH", PostalCode: "400001", Country: "IN"}

	err := patient.UpdateDemographics("New Name", "+922222222222", "new@example.com", address, updatedBy)

	require.NoError(t, err)
	assert.Equal(t, "New Name", patient.Name)
	assert.Equal(t, "+922222222222", patient.Phone)
	assert.Equal(t, "new@example.com", patient.Email)
	assert.Equal(t, address, patient.Address)
	assert.Equal(t, updatedBy, patient.UpdatedBy)
}

func TestPatientUpdateDemographicsRequiresName(t *testing.T) {
	patient := &Patient{Name: "Old Name"}

	err := patient.UpdateDemographics("", "+922222222222", "new@example.com", nil, uuid.New())

	require.Error(t, err)
	assert.Equal(t, "Old Name", patient.Name)
}

func TestPatientArchive(t *testing.T) {
	patient := &Patient{Status: PatientStatusInactive}
	updatedBy := uuid.New()

	err := patient.Archive(updatedBy)

	require.NoError(t, err)
	assert.Equal(t, PatientStatusArchived, patient.Status)
	assert.NotNil(t, patient.DeletedAt)
	assert.Equal(t, updatedBy, patient.UpdatedBy)
}
