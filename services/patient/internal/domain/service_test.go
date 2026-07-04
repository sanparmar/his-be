package domain

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// MockPatientRepository is a mock implementation for testing.
type MockPatientRepository struct {
	createFunc  func(ctx context.Context, patient *Patient) error
	getByIDFunc func(ctx context.Context, id, tenantID uuid.UUID) (*Patient, error)
	searchFunc  func(ctx context.Context, tenantID uuid.UUID, query *PatientQuery) ([]*Patient, int64, error)
	listFunc    func(ctx context.Context, tenantID uuid.UUID, page, pageSize int, sortBy, sortDir string) ([]*Patient, int64, error)
	updateFunc  func(ctx context.Context, patient *Patient) error
}

func (m *MockPatientRepository) Create(ctx context.Context, patient *Patient) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, patient)
	}
	return nil
}

func (m *MockPatientRepository) GetByID(ctx context.Context, id, tenantID uuid.UUID) (*Patient, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id, tenantID)
	}
	return nil, ErrNotFound
}

func (m *MockPatientRepository) GetByMRN(ctx context.Context, mrn string, tenantID uuid.UUID) (*Patient, error) {
	return nil, ErrNotFound
}

func (m *MockPatientRepository) Search(ctx context.Context, tenantID uuid.UUID, query *PatientQuery) ([]*Patient, int64, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, tenantID, query)
	}
	return []*Patient{}, 0, nil
}

func (m *MockPatientRepository) List(ctx context.Context, tenantID uuid.UUID, page, pageSize int, sortBy, sortDir string) ([]*Patient, int64, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, tenantID, page, pageSize, sortBy, sortDir)
	}
	return []*Patient{}, 0, nil
}

func (m *MockPatientRepository) Update(ctx context.Context, patient *Patient) error {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, patient)
	}
	return nil
}

// Test cases

func TestNewPatient(t *testing.T) {
	tenantID := uuid.New()
	createdBy := uuid.New()

	tests := []struct {
		name      string
		firstName string
		lastName  string
		dob       string
		gender    string
		phone     string
		email     string
	}{
		{
			name:      "valid patient",
			firstName: "John",
			lastName:  "Doe",
			dob:       "1990-01-15",
			gender:    "MALE",
			phone:     "1234567890",
			email:     "john@example.com",
		},
		{
			name:      "female patient",
			firstName: "Jane",
			lastName:  "Smith",
			dob:       "1985-05-20",
			gender:    "FEMALE",
			phone:     "0987654321",
			email:     "jane@example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patient := NewPatient(tenantID, tt.firstName, tt.lastName, tt.dob, tt.gender, tt.phone, tt.email, createdBy)

			assert.NotEqual(t, uuid.Nil, patient.ID)
			assert.Equal(t, tenantID, patient.TenantID)
			assert.Equal(t, tt.firstName, patient.FirstName)
			assert.Equal(t, tt.lastName, patient.LastName)
			assert.Equal(t, tt.dob, patient.DOB)
			assert.Equal(t, tt.gender, patient.Gender)
			assert.Equal(t, tt.phone, patient.Phone)
			assert.Equal(t, tt.email, patient.Email)
			assert.Equal(t, string(PatientStatusRegistered), patient.Status)
			assert.Equal(t, createdBy, patient.CreatedBy)
			assert.NotNil(t, patient.CreatedAt)
		})
	}
}

func TestPatientUpdateContact(t *testing.T) {
	patient := &Patient{
		ID:       uuid.New(),
		TenantID: uuid.New(),
		Phone:    "1234567890",
		Email:    "old@example.com",
	}

	updatedBy := uuid.New()
	newPhone := "9876543210"
	newEmail := "new@example.com"
	oldUpdatedAt := patient.UpdatedAt

	time.Sleep(10 * time.Millisecond)
	patient.UpdateContact(newPhone, newEmail, updatedBy)

	assert.Equal(t, newPhone, patient.Phone)
	assert.Equal(t, newEmail, patient.Email)
	assert.Equal(t, updatedBy, patient.UpdatedBy)
	assert.True(t, patient.UpdatedAt.After(oldUpdatedAt))
}

func TestPatientActivate(t *testing.T) {
	patient := &Patient{
		ID:     uuid.New(),
		Status: string(PatientStatusRegistered),
	}

	updatedBy := uuid.New()
	patient.Activate(updatedBy)

	assert.Equal(t, string(PatientStatusActive), patient.Status)
	assert.Equal(t, updatedBy, patient.UpdatedBy)
}

func TestPatientIsActive(t *testing.T) {
	tests := []struct {
		name     string
		status   string
		expected bool
	}{
		{name: "active status", status: string(PatientStatusActive), expected: true},
		{name: "registered status", status: string(PatientStatusRegistered), expected: false},
		{name: "inactive status", status: string(PatientStatusInactive), expected: false},
		{name: "discharged status", status: string(PatientStatusDischarged), expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patient := &Patient{Status: tt.status}
			assert.Equal(t, tt.expected, patient.IsActive())
		})
	}
}

func TestPatientServiceRegisterPatient(t *testing.T) {
	tenantID := uuid.New()
	createdBy := uuid.New()

	tests := []struct {
		name      string
		firstName string
		lastName  string
		dob       string
		gender    string
		phone     string
		email     string
		mockError error
		expectErr bool
	}{
		{
			name:      "successful registration",
			firstName: "John",
			lastName:  "Doe",
			dob:       "1990-01-15",
			gender:    "MALE",
			phone:     "1234567890",
			email:     "john@example.com",
			expectErr: false,
		},
		{
			name:      "repository error",
			firstName: "Jane",
			lastName:  "Smith",
			dob:       "1985-05-20",
			gender:    "FEMALE",
			phone:     "0987654321",
			email:     "jane@example.com",
			mockError: errors.New("db error"),
			expectErr: true,
		},
		{
			name:      "missing first name",
			firstName: "",
			lastName:  "Smith",
			dob:       "1985-05-20",
			gender:    "FEMALE",
			phone:     "0987654321",
			email:     "jane@example.com",
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockPatientRepository{
				createFunc: func(ctx context.Context, patient *Patient) error {
					return tt.mockError
				},
			}

			svc := NewPatientService(mockRepo)
			patient, err := svc.RegisterPatient(ctx, tenantID, tt.firstName, tt.lastName, tt.dob, tt.gender, tt.phone, tt.email, createdBy)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, patient)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, patient)
				assert.Equal(t, tt.firstName, patient.FirstName)
				assert.Equal(t, tt.lastName, patient.LastName)
			}
		})
	}
}

func TestPatientServiceGetPatient(t *testing.T) {
	patientID := uuid.New()
	tenantID := uuid.New()
	expectedPatient := &Patient{ID: patientID, TenantID: tenantID}

	tests := []struct {
		name      string
		mockError error
		expectErr bool
	}{
		{
			name:      "patient found",
			expectErr: false,
		},
		{
			name:      "patient not found",
			mockError: ErrNotFound,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := &MockPatientRepository{
				getByIDFunc: func(ctx context.Context, id, tid uuid.UUID) (*Patient, error) {
					if tt.mockError != nil {
						return nil, tt.mockError
					}
					return expectedPatient, nil
				},
			}

			svc := NewPatientService(mockRepo)
			patient, err := svc.GetPatient(ctx, patientID, tenantID)

			if tt.expectErr {
				assert.Error(t, err)
				assert.Nil(t, patient)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, expectedPatient, patient)
			}
		})
	}
}

var ctx = context.Background()
