package command

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/deloitte-us-consulting/his-be/services/patient/internal/domain"
)

// CreatePatientCommand represents the command to register a new patient.
type CreatePatientCommand struct {
	TenantID         uuid.UUID
	Name             string
	DOB              time.Time
	Gender           domain.Gender
	Phone            string
	Email            string
	BloodGroup       domain.BloodGroup
	Address          *domain.Address
	EmergencyContact *domain.EmergencyContact
	Insurance        *domain.Insurance
	Allergies        []string
	HomeMedications  []domain.Medication
	CreatedBy        uuid.UUID
}

// CreatePatientHandler handles the CreatePatient command.
type CreatePatientHandler struct {
	patientService *domain.PatientService
	eventPublisher domain.EventPublisher
}

// NewCreatePatientHandler creates a new handler.
func NewCreatePatientHandler(svc *domain.PatientService, pub domain.EventPublisher) *CreatePatientHandler {
	return &CreatePatientHandler{
		patientService: svc,
		eventPublisher: pub,
	}
}

// Handle executes the command.
func (h *CreatePatientHandler) Handle(ctx context.Context, cmd *CreatePatientCommand) (*domain.Patient, error) {
	input := domain.RegisterPatientInput{
		TenantID:         cmd.TenantID,
		Name:             cmd.Name,
		DOB:              cmd.DOB,
		Gender:           cmd.Gender,
		Phone:            cmd.Phone,
		Email:            cmd.Email,
		BloodGroup:       cmd.BloodGroup,
		Address:          cmd.Address,
		EmergencyContact: cmd.EmergencyContact,
		Insurance:        cmd.Insurance,
		Allergies:        cmd.Allergies,
		HomeMedications:  cmd.HomeMedications,
		CreatedBy:        cmd.CreatedBy,
	}

	patient, event, err := h.patientService.RegisterPatient(ctx, input)
	if err != nil {
		return nil, err
	}

	// Publish domain event asynchronously
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		// Log error but don't fail the command (eventual consistency)
		// In production, implement circuit breaker or retry logic
		return patient, nil
	}

	return patient, nil
}

// UpdateDemographicsCommand represents the command to update patient demographics.
type UpdateDemographicsCommand struct {
	TenantID  uuid.UUID
	MRN       string
	Name      string
	Phone     string
	Email     string
	Address   *domain.Address
	UpdatedBy uuid.UUID
}

// UpdateDemographicsHandler handles the UpdateDemographics command.
type UpdateDemographicsHandler struct {
	patientService *domain.PatientService
	eventPublisher domain.EventPublisher
}

// NewUpdateDemographicsHandler creates a new handler.
func NewUpdateDemographicsHandler(svc *domain.PatientService, pub domain.EventPublisher) *UpdateDemographicsHandler {
	return &UpdateDemographicsHandler{
		patientService: svc,
		eventPublisher: pub,
	}
}

// Handle executes the command.
func (h *UpdateDemographicsHandler) Handle(ctx context.Context, cmd *UpdateDemographicsCommand) (*domain.Patient, error) {
	patient, event, err := h.patientService.UpdateDemographics(
		ctx,
		cmd.TenantID,
		cmd.MRN,
		cmd.Name,
		cmd.Phone,
		cmd.Email,
		cmd.Address,
		cmd.UpdatedBy,
	)
	if err != nil {
		return nil, err
	}

	// Publish domain event
	if err := h.eventPublisher.Publish(ctx, event); err != nil {
		return patient, nil
	}

	return patient, nil
}

// UpdateInsuranceCommand represents the command to update patient insurance.
type UpdateInsuranceCommand struct {
	TenantID  uuid.UUID
	MRN       string
	Insurance *domain.Insurance
	UpdatedBy uuid.UUID
}

// UpdateInsuranceHandler handles the UpdateInsurance command.
type UpdateInsuranceHandler struct {
	patientService *domain.PatientService
	eventPublisher domain.EventPublisher
}

// NewUpdateInsuranceHandler creates a new handler.
func NewUpdateInsuranceHandler(svc *domain.PatientService, pub domain.EventPublisher) *UpdateInsuranceHandler {
	return &UpdateInsuranceHandler{
		patientService: svc,
		eventPublisher: pub,
	}
}

// Handle executes the command.
func (h *UpdateInsuranceHandler) Handle(ctx context.Context, cmd *UpdateInsuranceCommand) (*domain.Patient, error) {
	patient, err := h.patientService.UpdateInsurance(ctx, cmd.TenantID, cmd.MRN, cmd.Insurance, cmd.UpdatedBy)
	if err != nil {
		return nil, err
	}

	return patient, nil
}

// RecordAllergiesCommand represents the command to record patient allergies.
type RecordAllergiesCommand struct {
	TenantID  uuid.UUID
	MRN       string
	Allergies []string
	UpdatedBy uuid.UUID
}

// RecordAllergiesHandler handles the RecordAllergies command.
type RecordAllergiesHandler struct {
	patientService *domain.PatientService
	eventPublisher domain.EventPublisher
}

// NewRecordAllergiesHandler creates a new handler.
func NewRecordAllergiesHandler(svc *domain.PatientService, pub domain.EventPublisher) *RecordAllergiesHandler {
	return &RecordAllergiesHandler{
		patientService: svc,
		eventPublisher: pub,
	}
}

// Handle executes the command.
func (h *RecordAllergiesHandler) Handle(ctx context.Context, cmd *RecordAllergiesCommand) (*domain.Patient, error) {
	patient, err := h.patientService.RecordAllergies(ctx, cmd.TenantID, cmd.MRN, cmd.Allergies, cmd.UpdatedBy)
	if err != nil {
		return nil, err
	}

	return patient, nil
}

// RecordMedicationsCommand represents the command to record patient home medications.
type RecordMedicationsCommand struct {
	TenantID    uuid.UUID
	MRN         string
	Medications []domain.Medication
	UpdatedBy   uuid.UUID
}

// RecordMedicationsHandler handles the RecordMedications command.
type RecordMedicationsHandler struct {
	patientService *domain.PatientService
	eventPublisher domain.EventPublisher
}

// NewRecordMedicationsHandler creates a new handler.
func NewRecordMedicationsHandler(svc *domain.PatientService, pub domain.EventPublisher) *RecordMedicationsHandler {
	return &RecordMedicationsHandler{
		patientService: svc,
		eventPublisher: pub,
	}
}

// Handle executes the command.
func (h *RecordMedicationsHandler) Handle(ctx context.Context, cmd *RecordMedicationsCommand) (*domain.Patient, error) {
	patient, err := h.patientService.RecordMedications(ctx, cmd.TenantID, cmd.MRN, cmd.Medications, cmd.UpdatedBy)
	if err != nil {
		return nil, err
	}

	return patient, nil
}

// RecordVitalSignsCommand represents the command to record patient vital signs.
type RecordVitalSignsCommand struct {
	TenantID   uuid.UUID
	MRN        string
	VitalSigns *domain.VitalSigns
	UpdatedBy  uuid.UUID
}

// RecordVitalSignsHandler handles the RecordVitalSigns command.
type RecordVitalSignsHandler struct {
	patientService *domain.PatientService
	eventPublisher domain.EventPublisher
}

// NewRecordVitalSignsHandler creates a new handler.
func NewRecordVitalSignsHandler(svc *domain.PatientService, pub domain.EventPublisher) *RecordVitalSignsHandler {
	return &RecordVitalSignsHandler{
		patientService: svc,
		eventPublisher: pub,
	}
}

// Handle executes the command.
func (h *RecordVitalSignsHandler) Handle(ctx context.Context, cmd *RecordVitalSignsCommand) (*domain.Patient, error) {
	patient, err := h.patientService.RecordVitalSigns(ctx, cmd.TenantID, cmd.MRN, cmd.VitalSigns, cmd.UpdatedBy)
	if err != nil {
		return nil, err
	}

	return patient, nil
}
