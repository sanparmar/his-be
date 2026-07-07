package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// ─── Enumerations ─────────────────────────────────────────────────

// Gender represents patient biological sex
type Gender string

const (
	GenderMale    Gender = "MALE"
	GenderFemale  Gender = "FEMALE"
	GenderOther   Gender = "OTHER"
	GenderUnknown Gender = "UNKNOWN"
)

// BloodGroup represents ABO/Rh blood type
type BloodGroup string

const (
	BloodGroupAPlusPlus BloodGroup = "A+"
	BloodGroupAMinus    BloodGroup = "A-"
	BloodGroupBPlus     BloodGroup = "B+"
	BloodGroupBMinus    BloodGroup = "B-"
	BloodGroupABPlus    BloodGroup = "AB+"
	BloodGroupABMinus   BloodGroup = "AB-"
	BloodGroupOPlus     BloodGroup = "O+"
	BloodGroupOMinus    BloodGroup = "O-"
	BloodGroupUnknown   BloodGroup = "UNKNOWN"
)

func ParseGender(s string) (Gender, error) {
	switch s {
	case "MALE":
		return GenderMale, nil
	case "FEMALE":
		return GenderFemale, nil
	case "OTHER":
		return GenderOther, nil
	default:
		return GenderUnknown, fmt.Errorf("unknown gender: %s", s)
	}
}

func ParseBloodGroup(s string) (BloodGroup, error) {
	switch s {
	case "A+":
		return BloodGroupAPlusPlus, nil
	case "A-":
		return BloodGroupAMinus, nil
	case "B+":
		return BloodGroupBPlus, nil
	case "B-":
		return BloodGroupBMinus, nil
	case "AB+":
		return BloodGroupABPlus, nil
	case "AB-":
		return BloodGroupABMinus, nil
	case "O+":
		return BloodGroupOPlus, nil
	case "O-":
		return BloodGroupOMinus, nil
	default:
		return BloodGroupUnknown, fmt.Errorf("unknown blood group: %s", s)
	}
}

// PatientStatus represents patient lifecycle state
type PatientStatus string

const (
	PatientStatusActive   PatientStatus = "ACTIVE"
	PatientStatusInactive PatientStatus = "INACTIVE"
	PatientStatusArchived PatientStatus = "ARCHIVED"
)

// ─── Value Objects ────────────────────────────────────────────────

// Address represents physical location
type Address struct {
	Street     string `json:"street"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postal_code"`
	Country    string `json:"country"`
}

// EmergencyContact represents contact for emergencies
type EmergencyContact struct {
	Name     string `json:"name"`
	Relation string `json:"relation"`
	Phone    string `json:"phone"`
}

// Insurance represents health insurance policy
type Insurance struct {
	Provider        string     `json:"provider"`
	PolicyNo        string     `json:"policy_no"`
	TPA             string     `json:"tpa"`
	MemberID        string     `json:"member_id"`
	SumInsured      int64      `json:"sum_insured"`
	PolicyValidTill *time.Time `json:"policy_valid_till"`
	CorporateName   string     `json:"corporate_name"`
}

// Medication represents home medication
type Medication struct {
	Drug      string `json:"drug"`
	Dose      string `json:"dose"`
	Frequency string `json:"frequency"`
	Notes     string `json:"notes"`
}

// VitalSigns represents current vital measurements
type VitalSigns struct {
	BP         string     `json:"bp"`
	HR         *int       `json:"hr"`
	Temp       *float32   `json:"temp"`
	SpO2       *int       `json:"spo2"`
	RR         *int       `json:"rr"`
	Weight     *float32   `json:"weight"`
	Height     *int       `json:"height"`
	BMI        *float32   `json:"bmi"`
	RecordedAt *time.Time `json:"recorded_at"`
}

// ─── Aggregate Root ───────────────────────────────────────────────

// Patient is the patient aggregate root for the patient bounded context.
// FHIR R4 Patient resource aligned.
//
// Invariants:
// 1. TenantID is required (multi-tenancy)
// 2. MRN is unique per tenant
// 3. Name, DOB, Gender, Phone, BloodGroup, EmergencyContact are required on registration
// 4. PHI fields (name, DOB, phone, address) encrypted at application layer
// 5. Status transitions: ACTIVE ↔ INACTIVE → ARCHIVED (final)
// 6. All mutations are audit-logged
type Patient struct {
	// Identifiers
	ID       uuid.UUID `db:"id" json:"id"`
	TenantID uuid.UUID `db:"tenant_id" json:"tenant_id"`
	MRN      string    `db:"mrn" json:"mrn"` // Unique per tenant

	// Demographics
	Name       string     `db:"name" json:"name"`               // PHI - encrypted at rest
	DOB        time.Time  `db:"dob" json:"dob"`                 // PHI - encrypted at rest
	Gender     Gender     `db:"gender" json:"gender"`           // MALE, FEMALE, OTHER, UNKNOWN
	BloodGroup BloodGroup `db:"blood_group" json:"blood_group"` // A+, A-, B+, etc.

	// Contact Information
	Phone   string   `db:"phone" json:"phone"`  // PHI - encrypted at rest
	Email   string   `db:"email" json:"email"`  // PHI - encrypted at rest
	Address *Address `json:"address,omitempty"` // PHI - encrypted at rest

	// Health Identifiers
	ABHA         string `db:"abha" json:"abha"`                           // PHI - encrypted
	AadhaarLast4 string `db:"aadhaar_last_four" json:"aadhaar_last_four"` // Masked

	// Emergency Contact
	EmergencyContact *EmergencyContact `json:"emergency_contact,omitempty"`

	// Insurance (optional)
	Insurance *Insurance `json:"insurance,omitempty"`

	// Clinical Information
	HomeMedications []Medication `json:"home_medications,omitempty"`
	Allergies       []string     `json:"allergies,omitempty"`
	VitalSigns      *VitalSigns  `json:"vital_signs,omitempty"`

	// Status & Lifecycle
	Status PatientStatus `db:"status" json:"status"`

	// Audit Fields
	CreatedAt time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt time.Time  `db:"updated_at" json:"updated_at"`
	CreatedBy uuid.UUID  `db:"created_by" json:"created_by"`
	UpdatedBy uuid.UUID  `db:"updated_by" json:"updated_by"`
	DeletedAt *time.Time `db:"deleted_at" json:"deleted_at,omitempty"`
}

// NewPatient creates a new patient aggregate root with ACTIVE status.
func NewPatient(
	tenantID uuid.UUID,
	name string,
	dob time.Time,
	gender Gender,
	phone string,
	bloodGroup BloodGroup,
	emergencyContact *EmergencyContact,
	createdBy uuid.UUID,
) *Patient {
	return &Patient{
		ID:               uuid.New(),
		TenantID:         tenantID,
		MRN:              GenerateMRN(tenantID),
		Name:             name,
		DOB:              dob,
		Gender:           gender,
		BloodGroup:       bloodGroup,
		Phone:            phone,
		EmergencyContact: emergencyContact,
		Status:           PatientStatusActive,
		CreatedAt:        time.Now().UTC(),
		UpdatedAt:        time.Now().UTC(),
		CreatedBy:        createdBy,
		UpdatedBy:        createdBy,
	}
}

// ─── Domain Methods ───────────────────────────────────────────────

// CanTransitionTo checks if status transition is valid
func (p *Patient) CanTransitionTo(newStatus PatientStatus) bool {
	switch p.Status {
	case PatientStatusActive:
		return newStatus == PatientStatusInactive || newStatus == PatientStatusArchived
	case PatientStatusInactive:
		return newStatus == PatientStatusActive || newStatus == PatientStatusArchived
	case PatientStatusArchived:
		return false // Terminal state
	}
	return false
}

// TransitionStatus changes patient status with validation
func (p *Patient) TransitionStatus(newStatus PatientStatus, updatedBy uuid.UUID) error {
	if !p.CanTransitionTo(newStatus) {
		return fmt.Errorf("invalid status transition from %s to %s", p.Status, newStatus)
	}
	p.Status = newStatus
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = updatedBy
	return nil
}

// UpdateDemographics updates patient demographics with audit
func (p *Patient) UpdateDemographics(
	name string,
	phone string,
	email string,
	address *Address,
	updatedBy uuid.UUID,
) error {
	if name == "" {
		return fmt.Errorf("patient name is required")
	}
	p.Name = name
	p.Phone = phone
	p.Email = email
	p.Address = address
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = updatedBy
	return nil
}

// UpdateInsurance updates insurance information
func (p *Patient) UpdateInsurance(insurance *Insurance, updatedBy uuid.UUID) {
	p.Insurance = insurance
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = updatedBy
}

// RecordAllergies records patient allergies
func (p *Patient) RecordAllergies(allergies []string, updatedBy uuid.UUID) {
	p.Allergies = allergies
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = updatedBy
}

// RecordMedications records home medications
func (p *Patient) RecordMedications(medications []Medication, updatedBy uuid.UUID) {
	p.HomeMedications = medications
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = updatedBy
}

// RecordVitalSigns records vital signs
func (p *Patient) RecordVitalSigns(vitals *VitalSigns, updatedBy uuid.UUID) {
	p.VitalSigns = vitals
	p.UpdatedAt = time.Now().UTC()
	p.UpdatedBy = updatedBy
}

// Archive marks patient as archived (soft delete)
func (p *Patient) Archive(updatedBy uuid.UUID) error {
	if !p.CanTransitionTo(PatientStatusArchived) {
		return fmt.Errorf("cannot archive patient from status %s", p.Status)
	}
	now := time.Now().UTC()
	p.DeletedAt = &now
	p.Status = PatientStatusArchived
	p.UpdatedAt = now
	p.UpdatedBy = updatedBy
	return nil
}

// ─── Helper Functions ────────────────────────────────────────────

// GenerateMRN generates unique Medical Record Number per tenant
func GenerateMRN(tenantID uuid.UUID) string {
	// TODO: Integrate with hospital MRN numbering scheme (sequential per hospital)
	// Example: MRN-{hospital_code}-{sequential_number}
	return "MRN-" + uuid.New().String()[:8]
}
