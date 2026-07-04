package grpc

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"

	"github.com/deloitte-us-consulting/his-be/services/patient/internal/domain"
)

// Helper to create test patient
func newTestPatient() *domain.Patient {
	now := time.Now().UTC()
	tenantID := uuid.New()
	patientID := uuid.New()
	createdBy := uuid.New()
	updatedBy := uuid.New()

	hr := 72
	temp := float32(98.6)
	spo2 := 98
	rr := 16
	weight := float32(75.0)
	height := 180
	bmi := float32(23.1)

	return &domain.Patient{
		ID:         patientID,
		TenantID:   tenantID,
		MRN:        "MRN-2025-000001",
		Name:       "John Doe",
		DOB:        time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
		Gender:     domain.GenderMale,
		Phone:      "+919876543210",
		Email:      "john@example.com",
		BloodGroup: domain.BloodGroupOPlus,
		Address: &domain.Address{
			Street:     "123 Main St",
			City:       "Mumbai",
			State:      "Maharashtra",
			PostalCode: "400001",
			Country:    "India",
		},
		Status:    domain.PatientStatusActive,
		CreatedAt: now,
		UpdatedAt: now,
		CreatedBy: createdBy,
		UpdatedBy: updatedBy,
		VitalSigns: &domain.VitalSigns{
			BP:     "120/80",
			HR:     &hr,
			Temp:   &temp,
			SpO2:   &spo2,
			RR:     &rr,
			Weight: &weight,
			Height: &height,
			BMI:    &bmi,
		},
	}
}

// TestDomainPatientToProto tests the proto conversion logic
func TestDomainPatientToProto(t *testing.T) {
	t.Run("converts domain patient to proto with all fields", func(t *testing.T) {
		patient := newTestPatient()

		proto := domainPatientToProto(patient)

		assert.NotNil(t, proto)
		assert.Equal(t, patient.ID.String(), proto.Id)
		assert.Equal(t, patient.TenantID.String(), proto.TenantId)
		assert.Equal(t, patient.MRN, proto.Mrn)
		assert.Equal(t, patient.Name, proto.Name)
		assert.Equal(t, "1990-01-15", proto.Dob)
		assert.Equal(t, string(patient.Gender), proto.Gender)
		assert.Equal(t, string(patient.BloodGroup), proto.BloodGroup)
		assert.Equal(t, patient.Phone, proto.Phone)
		assert.Equal(t, patient.Email, proto.Email)
		assert.Equal(t, string(patient.Status), proto.Status)
		assert.Equal(t, patient.CreatedBy.String(), proto.CreatedBy)
		assert.Equal(t, patient.UpdatedBy.String(), proto.UpdatedBy)
	})

	t.Run("converts address correctly", func(t *testing.T) {
		patient := newTestPatient()

		proto := domainPatientToProto(patient)

		assert.NotNil(t, proto.Address)
		assert.Equal(t, patient.Address.Street, proto.Address.Street)
		assert.Equal(t, patient.Address.City, proto.Address.City)
		assert.Equal(t, patient.Address.State, proto.Address.State)
		assert.Equal(t, patient.Address.PostalCode, proto.Address.PostalCode)
		assert.Equal(t, patient.Address.Country, proto.Address.Country)
	})

	t.Run("converts vital signs with pointer handling", func(t *testing.T) {
		patient := newTestPatient()

		proto := domainPatientToProto(patient)

		assert.NotNil(t, proto.VitalSigns)
		assert.Equal(t, "120/80", proto.VitalSigns.Bp)
		assert.Equal(t, int32(72), proto.VitalSigns.Hr)
		assert.Equal(t, float32(98.6), proto.VitalSigns.Temp)
		assert.Equal(t, int32(98), proto.VitalSigns.Spo2)
		assert.Equal(t, int32(16), proto.VitalSigns.Rr)
		assert.Equal(t, float32(75.0), proto.VitalSigns.Weight)
		assert.Equal(t, int32(180), proto.VitalSigns.Height)
		assert.Equal(t, float32(23.1), proto.VitalSigns.Bmi)
	})

	t.Run("handles patient with minimal fields", func(t *testing.T) {
		patient := &domain.Patient{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			MRN:       "TEST-001",
			Name:      "Test Patient",
			DOB:       time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
			Gender:    domain.GenderFemale,
			Status:    domain.PatientStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: uuid.New(),
			UpdatedBy: uuid.New(),
		}

		proto := domainPatientToProto(patient)

		assert.NotNil(t, proto)
		assert.Equal(t, "TEST-001", proto.Mrn)
		assert.Nil(t, proto.Address)
		assert.Nil(t, proto.VitalSigns)
		assert.Nil(t, proto.EmergencyContact)
		assert.Nil(t, proto.Insurance)
		assert.Len(t, proto.HomeMedications, 0)
		assert.Len(t, proto.Allergies, 0)
	})

	t.Run("converts emergency contact", func(t *testing.T) {
		patient := newTestPatient()
		patient.EmergencyContact = &domain.EmergencyContact{
			Name:     "Jane Doe",
			Relation: "Spouse",
			Phone:    "+919876543211",
		}

		proto := domainPatientToProto(patient)

		assert.NotNil(t, proto.EmergencyContact)
		assert.Equal(t, "Jane Doe", proto.EmergencyContact.Name)
		assert.Equal(t, "Spouse", proto.EmergencyContact.Relation)
		assert.Equal(t, "+919876543211", proto.EmergencyContact.Phone)
	})

	t.Run("converts insurance", func(t *testing.T) {
		patient := newTestPatient()
		policyTill := time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)
		patient.Insurance = &domain.Insurance{
			Provider:        "Aditya Birla",
			PolicyNo:        "POL-123456",
			TPA:             "TATA AIG",
			MemberID:        "MEMBER-789",
			SumInsured:      500000,
			PolicyValidTill: &policyTill,
			CorporateName:   "ABC Corp",
		}

		proto := domainPatientToProto(patient)

		assert.NotNil(t, proto.Insurance)
		assert.Equal(t, "Aditya Birla", proto.Insurance.Provider)
		assert.Equal(t, "POL-123456", proto.Insurance.PolicyNo)
		assert.Equal(t, "TATA AIG", proto.Insurance.Tpa)
		assert.Equal(t, "MEMBER-789", proto.Insurance.MemberId)
		assert.Equal(t, int64(500000), proto.Insurance.SumInsured)
		assert.Equal(t, "2026-12-31", proto.Insurance.PolicyValidTill)
		assert.Equal(t, "ABC Corp", proto.Insurance.CorporateName)
	})

	t.Run("converts home medications", func(t *testing.T) {
		patient := newTestPatient()
		patient.HomeMedications = []domain.Medication{
			{
				Drug:      "Aspirin",
				Dose:      "500mg",
				Frequency: "OD",
				Notes:     "After meals",
			},
			{
				Drug:      "Metformin",
				Dose:      "1000mg",
				Frequency: "BD",
				Notes:     "For diabetes",
			},
		}

		proto := domainPatientToProto(patient)

		assert.Len(t, proto.HomeMedications, 2)
		assert.Equal(t, "Aspirin", proto.HomeMedications[0].Drug)
		assert.Equal(t, "500mg", proto.HomeMedications[0].Dose)
		assert.Equal(t, "OD", proto.HomeMedications[0].Frequency)
		assert.Equal(t, "After meals", proto.HomeMedications[0].Notes)
	})

	t.Run("converts allergies", func(t *testing.T) {
		patient := newTestPatient()
		patient.Allergies = []string{"Penicillin", "Shellfish"}

		proto := domainPatientToProto(patient)

		assert.Len(t, proto.Allergies, 2)
		assert.Contains(t, proto.Allergies, "Penicillin")
		assert.Contains(t, proto.Allergies, "Shellfish")
	})

	t.Run("handles nil vital signs gracefully", func(t *testing.T) {
		patient := newTestPatient()
		patient.VitalSigns = nil

		proto := domainPatientToProto(patient)

		assert.NotNil(t, proto)
		assert.Nil(t, proto.VitalSigns)
	})

	t.Run("formats DOB correctly", func(t *testing.T) {
		patient := &domain.Patient{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			MRN:       "TEST-001",
			Name:      "Test",
			DOB:       time.Date(1985, 12, 25, 0, 0, 0, 0, time.UTC),
			Status:    domain.PatientStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: uuid.New(),
			UpdatedBy: uuid.New(),
		}

		proto := domainPatientToProto(patient)

		assert.Equal(t, "1985-12-25", proto.Dob)
	})

	t.Run("preserves null timestamps in proto", func(t *testing.T) {
		patient := &domain.Patient{
			ID:        uuid.New(),
			TenantID:  uuid.New(),
			MRN:       "TEST-001",
			Name:      "Test",
			Status:    domain.PatientStatusActive,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			CreatedBy: uuid.New(),
			UpdatedBy: uuid.New(),
			// DeletedAt is nil
		}

		proto := domainPatientToProto(patient)

		assert.Nil(t, proto.DeletedAt)
	})
}

// TestDomainPatientToProtoEdgeCases tests edge cases and boundary conditions
func TestDomainPatientToProtoEdgeCases(t *testing.T) {
	t.Run("handles patient with deleted_at set", func(t *testing.T) {
		patient := newTestPatient()
		deletedAt := time.Now().UTC()
		patient.DeletedAt = &deletedAt
		patient.Status = domain.PatientStatusArchived

		proto := domainPatientToProto(patient)

		assert.NotNil(t, proto.DeletedAt)
		assert.Equal(t, deletedAt, *proto.DeletedAt)
	})

	t.Run("handles all blood group types", func(t *testing.T) {
		bloodGroups := []domain.BloodGroup{
			domain.BloodGroupAPlusPlus,
			domain.BloodGroupAMinus,
			domain.BloodGroupBPlus,
			domain.BloodGroupBMinus,
			domain.BloodGroupABPlus,
			domain.BloodGroupABMinus,
			domain.BloodGroupOPlus,
			domain.BloodGroupOMinus,
			domain.BloodGroupUnknown,
		}

		for _, bg := range bloodGroups {
			patient := newTestPatient()
			patient.BloodGroup = bg

			proto := domainPatientToProto(patient)

			assert.Equal(t, string(bg), proto.BloodGroup)
		}
	})

	t.Run("handles all gender types", func(t *testing.T) {
		genders := []domain.Gender{
			domain.GenderMale,
			domain.GenderFemale,
			domain.GenderOther,
			domain.GenderUnknown,
		}

		for _, g := range genders {
			patient := newTestPatient()
			patient.Gender = g

			proto := domainPatientToProto(patient)

			assert.Equal(t, string(g), proto.Gender)
		}
	})

	t.Run("handles all patient statuses", func(t *testing.T) {
		statuses := []domain.PatientStatus{
			domain.PatientStatusActive,
			domain.PatientStatusInactive,
			domain.PatientStatusArchived,
		}

		for _, status := range statuses {
			patient := newTestPatient()
			patient.Status = status

			proto := domainPatientToProto(patient)

			assert.Equal(t, string(status), proto.Status)
		}
	})

	t.Run("handles large medication list", func(t *testing.T) {
		patient := newTestPatient()
		for i := 0; i < 10; i++ {
			patient.HomeMedications = append(patient.HomeMedications, domain.Medication{
				Drug:      "Drug" + string(rune(i)),
				Dose:      "Dose" + string(rune(i)),
				Frequency: "OD",
			})
		}

		proto := domainPatientToProto(patient)

		assert.Len(t, proto.HomeMedications, 10)
	})

	t.Run("preserves zero vital sign values", func(t *testing.T) {
		patient := newTestPatient()
		zero := int(0)
		zeroFloat := float32(0)
		patient.VitalSigns.HR = &zero
		patient.VitalSigns.Temp = &zeroFloat

		proto := domainPatientToProto(patient)

		assert.NotNil(t, proto.VitalSigns.Hr)
		assert.Equal(t, int32(0), proto.VitalSigns.Hr)
		assert.NotNil(t, proto.VitalSigns.Temp)
		assert.Equal(t, float32(0), proto.VitalSigns.Temp)
	})
}
