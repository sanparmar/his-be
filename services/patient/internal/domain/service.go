package domain

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// PatientService handles patient domain business rules and use cases.
type PatientService struct {
	repo PatientRepository
}

// NewPatientService creates a new patient domain service.
func NewPatientService(repo PatientRepository) *PatientService {
	return &PatientService{repo: repo}
}

// RegisterPatientInput contains patient registration data
type RegisterPatientInput struct {
	TenantID         uuid.UUID
	Name             string
	DOB              time.Time
	Gender           Gender
	Phone            string
	Email            string
	BloodGroup       BloodGroup
	Address          *Address
	EmergencyContact *EmergencyContact
	Insurance        *Insurance
	Allergies        []string
	HomeMedications  []Medication
	CreatedBy        uuid.UUID
}

// RegisterPatient creates a new patient with all required demographics.
// Returns PatientRegistered event for publishing to Kafka.
func (s *PatientService) RegisterPatient(
	ctx context.Context,
	input RegisterPatientInput,
) (*Patient, *PatientRegistered, error) {
	// Create new patient aggregate
	patient := NewPatient(
		input.TenantID,
		input.Name,
		input.DOB,
		input.Gender,
		input.Phone,
		input.BloodGroup,
		input.EmergencyContact,
		input.CreatedBy,
	)

	// Set optional fields
	patient.Email = input.Email
	patient.Address = input.Address
	patient.Insurance = input.Insurance
	patient.Allergies = input.Allergies
	patient.HomeMedications = input.HomeMedications
	patient.BloodGroup = input.BloodGroup

	// Persist to repository
	if err := s.repo.Create(ctx, patient); err != nil {
		return nil, nil, fmt.Errorf("register patient: %w", err)
	}

	// Create domain event for publishing
	event := &PatientRegistered{
		PatientID:        patient.ID,
		TenantID:         patient.TenantID,
		MRN:              patient.MRN,
		Name:             patient.Name,
		DOB:              patient.DOB,
		Gender:           patient.Gender,
		BloodGroup:       patient.BloodGroup,
		Phone:            patient.Phone,
		Email:            patient.Email,
		Address:          patient.Address,
		EmergencyContact: patient.EmergencyContact,
		Status:           patient.Status,
		CorrelID:         uuid.New().String(), // TODO: Use context correlation ID
		RecordedAt:       time.Now().UTC(),
		CreatedBy:        patient.CreatedBy,
	}

	return patient, event, nil
}

// GetPatientByMRN retrieves patient by MRN (tenant-scoped)
func (s *PatientService) GetPatientByMRN(
	ctx context.Context,
	tenantID uuid.UUID,
	mrn string,
) (*Patient, error) {
	patient, err := s.repo.GetByMRN(ctx, tenantID, mrn)
	if err != nil {
		return nil, fmt.Errorf("get patient by mrn: %w", err)
	}
	return patient, nil
}

// ListPatients lists patients with pagination (tenant-scoped)
func (s *PatientService) ListPatients(
	ctx context.Context,
	tenantID uuid.UUID,
	offset, limit int,
) ([]*Patient, int64, error) {
	// Exclude archived patients by default
	status := PatientStatusActive
	patients, total, err := s.repo.List(ctx, tenantID, offset, limit, &status)
	if err != nil {
		return nil, 0, fmt.Errorf("list patients: %w", err)
	}
	return patients, total, nil
}

// SearchPatients performs full-text search on patient records
func (s *PatientService) SearchPatients(
	ctx context.Context,
	tenantID uuid.UUID,
	query string,
	offset, limit int,
) ([]*Patient, int64, error) {
	patients, total, err := s.repo.Search(ctx, tenantID, query, offset, limit)
	if err != nil {
		return nil, 0, fmt.Errorf("search patients: %w", err)
	}
	return patients, total, nil
}

// UpdateDemographics updates patient contact and address information
func (s *PatientService) UpdateDemographics(
	ctx context.Context,
	tenantID uuid.UUID,
	mrn string,
	name, phone, email string,
	address *Address,
	updatedBy uuid.UUID,
) (*Patient, *PatientUpdated, error) {
	patient, err := s.repo.GetByMRN(ctx, tenantID, mrn)
	if err != nil {
		return nil, nil, fmt.Errorf("update demographics: %w", err)
	}

	// Record old values for audit
	oldName := patient.Name
	oldPhone := patient.Phone
	oldEmail := patient.Email

	// Apply domain method
	if err := patient.UpdateDemographics(name, phone, email, address, updatedBy); err != nil {
		return nil, nil, err
	}

	// Persist
	if err := s.repo.Update(ctx, patient); err != nil {
		return nil, nil, fmt.Errorf("update demographics: %w", err)
	}

	// Create domain event
	changedFields := []string{}
	if oldName != name {
		changedFields = append(changedFields, "name")
	}
	if oldPhone != phone {
		changedFields = append(changedFields, "phone")
	}
	if oldEmail != email {
		changedFields = append(changedFields, "email")
	}
	if address != nil {
		changedFields = append(changedFields, "address")
	}

	event := &PatientUpdated{
		PatientID:     patient.ID,
		TenantID:      patient.TenantID,
		MRN:           patient.MRN,
		Name:          patient.Name,
		Phone:         patient.Phone,
		Email:         patient.Email,
		Address:       patient.Address,
		Status:        patient.Status,
		ChangedFields: changedFields,
		CorrelID:      uuid.New().String(),
		RecordedAt:    time.Now().UTC(),
		UpdatedBy:     updatedBy,
	}

	return patient, event, nil
}

// UpdateInsurance updates patient insurance information
func (s *PatientService) UpdateInsurance(
	ctx context.Context,
	tenantID uuid.UUID,
	mrn string,
	insurance *Insurance,
	updatedBy uuid.UUID,
) (*Patient, error) {
	patient, err := s.repo.GetByMRN(ctx, tenantID, mrn)
	if err != nil {
		return nil, fmt.Errorf("update insurance: %w", err)
	}

	patient.UpdateInsurance(insurance, updatedBy)

	if err := s.repo.Update(ctx, patient); err != nil {
		return nil, fmt.Errorf("update insurance: %w", err)
	}

	return patient, nil
}

// RecordAllergies records patient allergies
func (s *PatientService) RecordAllergies(
	ctx context.Context,
	tenantID uuid.UUID,
	mrn string,
	allergies []string,
	updatedBy uuid.UUID,
) (*Patient, error) {
	patient, err := s.repo.GetByMRN(ctx, tenantID, mrn)
	if err != nil {
		return nil, fmt.Errorf("record allergies: %w", err)
	}

	patient.RecordAllergies(allergies, updatedBy)

	if err := s.repo.Update(ctx, patient); err != nil {
		return nil, fmt.Errorf("record allergies: %w", err)
	}

	return patient, nil
}

// RecordMedications records home medications
func (s *PatientService) RecordMedications(
	ctx context.Context,
	tenantID uuid.UUID,
	mrn string,
	medications []Medication,
	updatedBy uuid.UUID,
) (*Patient, error) {
	patient, err := s.repo.GetByMRN(ctx, tenantID, mrn)
	if err != nil {
		return nil, fmt.Errorf("record medications: %w", err)
	}

	patient.RecordMedications(medications, updatedBy)

	if err := s.repo.Update(ctx, patient); err != nil {
		return nil, fmt.Errorf("record medications: %w", err)
	}

	return patient, nil
}

// RecordVitalSigns records vital measurements
func (s *PatientService) RecordVitalSigns(
	ctx context.Context,
	tenantID uuid.UUID,
	mrn string,
	vitals *VitalSigns,
	updatedBy uuid.UUID,
) (*Patient, error) {
	patient, err := s.repo.GetByMRN(ctx, tenantID, mrn)
	if err != nil {
		return nil, fmt.Errorf("record vitals: %w", err)
	}

	patient.RecordVitalSigns(vitals, updatedBy)

	if err := s.repo.Update(ctx, patient); err != nil {
		return nil, fmt.Errorf("record vitals: %w", err)
	}

	return patient, nil
}
