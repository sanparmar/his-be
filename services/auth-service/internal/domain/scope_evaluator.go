package domain

import (
	"github.com/google/uuid"
)

type UserContext struct {
	UserID          uuid.UUID
	TenantID        uuid.UUID
	OrganizationID  *uuid.UUID
	HospitalID      *uuid.UUID
	DepartmentID    *uuid.UUID
	Roles           []Role
}

type ResourceContext struct {
	ResourceID      uuid.UUID
	ResourceType    string
	TenantID        uuid.UUID
	OrganizationID  *uuid.UUID
	HospitalID      *uuid.UUID
	DepartmentID    *uuid.UUID
}

type ScopeEvaluator struct{}

func NewScopeEvaluator() *ScopeEvaluator {
	return &ScopeEvaluator{}
}

func (e *ScopeEvaluator) CanAccessResource(userCtx *UserContext, resourceCtx *ResourceContext, perm *Permission) bool {
	if userCtx.TenantID != resourceCtx.TenantID {
		return false
	}

	requiredLevel := perm.ScopeLevel()

	switch requiredLevel {
	case 0: // ScopeOwn
		return e.checkOwnScope(userCtx, resourceCtx)
	case 1: // ScopeDepartment
		return e.checkDepartmentScope(userCtx, resourceCtx)
	case 2: // ScopeHospital
		return e.checkHospitalScope(userCtx, resourceCtx)
	case 3: // ScopeOrganization
		return e.checkOrganizationScope(userCtx, resourceCtx)
	case 4: // ScopeTenant
		return e.checkTenantScope(userCtx, resourceCtx)
	case 5: // ScopeAll
		return true
	default:
		return false
	}
}

func (e *ScopeEvaluator) checkOwnScope(userCtx *UserContext, resourceCtx *ResourceContext) bool {
	return userCtx.UserID == resourceCtx.ResourceID ||
		(resourceCtx.ResourceType == "patient" && e.isAssignedPatient(userCtx, resourceCtx.ResourceID))
}

func (e *ScopeEvaluator) checkDepartmentScope(userCtx *UserContext, resourceCtx *ResourceContext) bool {
	if userCtx.DepartmentID == nil || resourceCtx.DepartmentID == nil {
		return false
	}
	return *userCtx.DepartmentID == *resourceCtx.DepartmentID
}

func (e *ScopeEvaluator) checkHospitalScope(userCtx *UserContext, resourceCtx *ResourceContext) bool {
	if userCtx.HospitalID == nil || resourceCtx.HospitalID == nil {
		return false
	}
	return *userCtx.HospitalID == *resourceCtx.HospitalID
}

func (e *ScopeEvaluator) checkOrganizationScope(userCtx *UserContext, resourceCtx *ResourceContext) bool {
	if userCtx.OrganizationID == nil || resourceCtx.OrganizationID == nil {
		return false
	}
	return *userCtx.OrganizationID == *resourceCtx.OrganizationID
}

func (e *ScopeEvaluator) checkTenantScope(userCtx *UserContext, resourceCtx *ResourceContext) bool {
	return userCtx.TenantID == resourceCtx.TenantID
}

func (e *ScopeEvaluator) isAssignedPatient(userCtx *UserContext, patientID uuid.UUID) bool {
	for _, role := range userCtx.Roles {
		if role.Category == "clinical" || role.Category == "nursing" {
			return true
		}
	}
	return false
}

func (e *ScopeEvaluator) GetEffectiveScope(userCtx *UserContext, resourceCtx *ResourceContext) string {
	if e.checkOwnScope(userCtx, resourceCtx) {
		return "own"
	}
	if e.checkDepartmentScope(userCtx, resourceCtx) {
		return "department"
	}
	if e.checkHospitalScope(userCtx, resourceCtx) {
		return "hospital"
	}
	if e.checkOrganizationScope(userCtx, resourceCtx) {
		return "organization"
	}
	if e.checkTenantScope(userCtx, resourceCtx) {
		return "tenant"
	}
	return "all"
}