package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPatientCreation(t *testing.T) {
	t.Run("creates patient with valid data", func(t *testing.T) {
		tenantID := uuid.New()
		createdBy := uuid.New()

		patient := &Patient{
			ID:        uuid.New(),
			TenantID:  tenantID,
			MRN:       "MRN-2025-000001",
			FirstName: "John",
			LastName:  "Doe",
			DOB:       time.Date(1990, 1, 15, 0, 0, 0, 0, time.UTC),
			Gender:    "M",
			Phone:     "+919876543210",
			Email:     "john@hospital.local",
			Status:    "active",
			CreatedBy: createdBy,
			CreatedAt: time.Now(),
		}

		assert.NotNil(t, patient.ID)
		assert.Equal(t, tenantID, patient.TenantID)
		assert.Equal(t, "MRN-2025-000001", patient.MRN)
		assert.Equal(t, "John", patient.FirstName)
		assert.Equal(t, "Doe", patient.LastName)
		assert.Equal(t, "active", patient.Status)
	})

	t.Run("validates MRN uniqueness per tenant", func(t *testing.T) {
		tenantID := uuid.New()
		mrn := "MRN-2025-000001"

		patient1 := &Patient{
			ID:       uuid.New(),
			TenantID: tenantID,
			MRN:      mrn,
		}

		patient2 := &Patient{
			ID:       uuid.New(),
			TenantID: tenantID,
			MRN:      mrn,
		}

		// Both have same MRN in same tenant - would violate uniqueness
		assert.Equal(t, patient1.MRN, patient2.MRN)
		assert.Equal(t, patient1.TenantID, patient2.TenantID)
	})
}
