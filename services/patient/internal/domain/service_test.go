package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type MockPatientRepository struct {
	createFunc   func(ctx context.Context, patient *Patient) error
	getByIDFunc  func(ctx context.Context, tenantID, patientID uuid.UUID) (*Patient, error)
	getByMRNFunc func(ctx context.Context, tenantID uuid.UUID, mrn string) (*Patient, error)
	updateFunc   func(ctx context.Context, patient *Patient) error
	deleteFunc   func(ctx context.Context, tenantID, patientID uuid.UUID, deletedBy uuid.UUID) error
	listFunc     func(ctx context.Context, tenantID uuid.UUID, offset, limit int, statusFilter *PatientStatus) ([]*Patient, int64, error)
	searchFunc   func(ctx context.Context, tenantID uuid.UUID, query string, offset, limit int) ([]*Patient, int64, error)
}

func (m *MockPatientRepository) Create(ctx context.Context, patient *Patient) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, patient)
	}
	return nil
}

func (m *MockPatientRepository) GetByID(ctx context.Context, tenantID, patientID uuid.UUID) (*Patient, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, tenantID, patientID)
	}
	return nil, ErrNotFound
}

func (m *MockPatientRepository) GetByMRN(ctx context.Context, tenantID uuid.UUID, mrn string) (*Patient, error) {
	if m.getByMRNFunc != nil {
		return m.getByMRNFunc(ctx, tenantID, mrn)
	}
	return nil, ErrNotFound
}

func (m *MockPatientRepository) Update(ctx context.Context, patient *Patient) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, patient)
	}
	return nil
}

func (m *MockPatientRepository) Delete(ctx context.Context, tenantID, patientID uuid.UUID, deletedBy uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, tenantID, patientID, deletedBy)
	}
	return nil
}

func (m *MockPatientRepository) List(ctx context.Context, tenantID uuid.UUID, offset, limit int, statusFilter *PatientStatus) ([]*Patient, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, tenantID, offset, limit, statusFilter)
	}
	return nil, 0, nil
}

func (m *MockPatientRepository) Search(ctx context.Context, tenantID uuid.UUID, query string, offset, limit int) ([]*Patient, int64, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, tenantID, query, offset, limit)
	}
	return nil, 0, nil
}

func (m *MockPatientRepository) GetAuditLog(ctx context.Context, tenantID, patientID uuid.UUID, offset, limit int) ([]*AuditLog, error) {
	return nil, nil
}

func TestPatientServiceRegisterPatient(t *testing.T) {
	tenantID := uuid.New()
	createdBy := uuid.New()
	dob := time.Date(1990, time.January, 15, 0, 0, 0, 0, time.UTC)

	t.Run("successful registration", func(t *testing.T) {
		mockRepo := &MockPatientRepository{}
		svc := NewPatientService(mockRepo)

		patient, event, err := svc.RegisterPatient(context.Background(), RegisterPatientInput{
			TenantID:         tenantID,
			Name:             "John Doe",
			DOB:              dob,
			Gender:           GenderMale,
			Phone:            "+919876543210",
			Email:            "john@example.com",
			BloodGroup:       BloodGroupOPlus,
			EmergencyContact: &EmergencyContact{Name: "Jane Doe", Relation: "Spouse", Phone: "+919900000000"},
			CreatedBy:        createdBy,
		})

		require.NoError(t, err)
		require.NotNil(t, patient)
		require.NotNil(t, event)
		assert.Equal(t, tenantID, patient.TenantID)
		assert.Equal(t, patient.ID, event.PatientID)
		assert.Equal(t, patient.MRN, event.MRN)
	})

	t.Run("repository error", func(t *testing.T) {
		mockRepo := &MockPatientRepository{
			createFunc: func(ctx context.Context, patient *Patient) error {
				return errors.New("db error")
			},
		}
		svc := NewPatientService(mockRepo)

		patient, event, err := svc.RegisterPatient(context.Background(), RegisterPatientInput{
			TenantID:         tenantID,
			Name:             "Jane Doe",
			DOB:              dob,
			Gender:           GenderFemale,
			Phone:            "+919800000000",
			BloodGroup:       BloodGroupAPlusPlus,
			EmergencyContact: &EmergencyContact{Name: "John Doe", Relation: "Brother", Phone: "+919811111111"},
			CreatedBy:        createdBy,
		})

		require.Error(t, err)
		assert.Nil(t, patient)
		assert.Nil(t, event)
	})
}

func TestPatientServiceGetPatientByMRN(t *testing.T) {
	expected := &Patient{ID: uuid.New(), TenantID: uuid.New(), MRN: "MRN-123", Name: "John Doe"}
	mockRepo := &MockPatientRepository{
		getByMRNFunc: func(ctx context.Context, tenantID uuid.UUID, mrn string) (*Patient, error) {
			assert.Equal(t, expected.TenantID, tenantID)
			assert.Equal(t, expected.MRN, mrn)
			return expected, nil
		},
	}

	svc := NewPatientService(mockRepo)
	patient, err := svc.GetPatientByMRN(context.Background(), expected.TenantID, expected.MRN)

	require.NoError(t, err)
	assert.Equal(t, expected, patient)
}

func TestPatientServiceListPatients(t *testing.T) {
	tenantID := uuid.New()
	expected := []*Patient{{ID: uuid.New(), TenantID: tenantID, Status: PatientStatusActive}}
	mockRepo := &MockPatientRepository{
		listFunc: func(ctx context.Context, gotTenantID uuid.UUID, offset, limit int, statusFilter *PatientStatus) ([]*Patient, int64, error) {
			require.NotNil(t, statusFilter)
			assert.Equal(t, PatientStatusActive, *statusFilter)
			assert.Equal(t, tenantID, gotTenantID)
			assert.Equal(t, 0, offset)
			assert.Equal(t, 25, limit)
			return expected, int64(len(expected)), nil
		},
	}

	svc := NewPatientService(mockRepo)
	patients, total, err := svc.ListPatients(context.Background(), tenantID, 0, 25)

	require.NoError(t, err)
	assert.Equal(t, expected, patients)
	assert.Equal(t, int64(1), total)
}

func TestPatientServiceSearchPatients(t *testing.T) {
	tenantID := uuid.New()
	query := "john"
	expected := []*Patient{{ID: uuid.New(), TenantID: tenantID, Name: "John Doe"}}
	mockRepo := &MockPatientRepository{
		searchFunc: func(ctx context.Context, gotTenantID uuid.UUID, gotQuery string, offset, limit int) ([]*Patient, int64, error) {
			assert.Equal(t, tenantID, gotTenantID)
			assert.Equal(t, query, gotQuery)
			assert.Equal(t, 10, offset)
			assert.Equal(t, 5, limit)
			return expected, int64(len(expected)), nil
		},
	}

	svc := NewPatientService(mockRepo)
	patients, total, err := svc.SearchPatients(context.Background(), tenantID, query, 10, 5)

	require.NoError(t, err)
	assert.Equal(t, expected, patients)
	assert.Equal(t, int64(1), total)
}

func TestPatientServiceUpdateDemographics(t *testing.T) {
	tenantID := uuid.New()
	updatedBy := uuid.New()
	patient := &Patient{ID: uuid.New(), TenantID: tenantID, MRN: "MRN-123", Name: "Old Name", Phone: "+911111111111", Email: "old@example.com", Status: PatientStatusActive}
	address := &Address{Street: "123 Street", City: "Mumbai", State: "MH", PostalCode: "400001", Country: "IN"}
	mockRepo := &MockPatientRepository{
		getByMRNFunc: func(ctx context.Context, gotTenantID uuid.UUID, mrn string) (*Patient, error) {
			assert.Equal(t, tenantID, gotTenantID)
			assert.Equal(t, patient.MRN, mrn)
			return patient, nil
		},
		updateFunc: func(ctx context.Context, updated *Patient) error {
			assert.Equal(t, "New Name", updated.Name)
			assert.Equal(t, "+922222222222", updated.Phone)
			return nil
		},
	}

	svc := NewPatientService(mockRepo)
	updatedPatient, event, err := svc.UpdateDemographics(context.Background(), tenantID, patient.MRN, "New Name", "+922222222222", "new@example.com", address, updatedBy)

	require.NoError(t, err)
	require.NotNil(t, event)
	assert.Equal(t, "New Name", updatedPatient.Name)
	assert.ElementsMatch(t, []string{"name", "phone", "email", "address"}, event.ChangedFields)
	assert.Equal(t, updatedBy, event.UpdatedBy)
}
