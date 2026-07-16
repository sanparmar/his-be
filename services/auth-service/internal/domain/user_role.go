package domain

import (
	"github.com/google/uuid"
	"time"
)

type UserRole struct {
	UserID        uuid.UUID  `json:"user_id" db:"user_id"`
	RoleID        uuid.UUID  `json:"role_id" db:"role_id"`
	TenantID      uuid.UUID  `json:"tenant_id" db:"tenant_id"`
	OrganizationID *uuid.UUID `json:"organization_id,omitempty" db:"organization_id"`
	HospitalID    *uuid.UUID `json:"hospital_id,omitempty" db:"hospital_id"`
	DepartmentID  *uuid.UUID `json:"department_id,omitempty" db:"department_id"`
	AssignedBy    uuid.UUID  `json:"assigned_by" db:"assigned_by"`
	AssignedAt    time.Time  `json:"assigned_at" db:"assigned_at"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty" db:"expires_at"`
}

func (ur *UserRole) IsExpired() bool {
	if ur.ExpiresAt == nil {
		return false
	}
	return ur.ExpiresAt.Before(time.Now())
}

func (ur *UserRole) MatchesTenant(tid uuid.UUID) bool {
	return ur.TenantID == tid
}

func (ur *UserRole) MatchesOrganization(oid uuid.UUID) bool {
	if ur.OrganizationID == nil {
		return false
	}
	return *ur.OrganizationID == oid
}

func (ur *UserRole) MatchesHospital(hid uuid.UUID) bool {
	if ur.HospitalID == nil {
		return false
	}
	return *ur.HospitalID == hid
}

func (ur *UserRole) MatchesDepartment(did uuid.UUID) bool {
	if ur.DepartmentID == nil {
		return false
	}
	return *ur.DepartmentID == did
}