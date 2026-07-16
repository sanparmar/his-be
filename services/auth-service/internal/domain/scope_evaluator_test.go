package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestScopeEvaluator_Evaluate(t *testing.T) {
	eval := NewScopeEvaluator()

	userID := uuid.New()
	tenantID := uuid.New()
	orgID := uuid.New()
	hospitalID := uuid.New()
	deptID := uuid.New()

	userCtx := &UserContext{
		UserID:         userID,
		TenantID:       tenantID,
		OrganizationID: &orgID,
		HospitalID:     &hospitalID,
		DepartmentID:   &deptID,
	}

	tests := []struct {
		name         string
		resourceCtx  *ResourceContext
		requiredScope string
		expected     bool
	}{
		{
			name: "own scope - same user",
			resourceCtx: &ResourceContext{
				ResourceID:  userID,
				TenantID:    tenantID,
				DepartmentID: &deptID,
			},
			requiredScope: ScopeOwn,
			expected:      true,
		},
		{
			name: "own scope - different user",
			resourceCtx: &ResourceContext{
				ResourceID:  uuid.New(),
				TenantID:    tenantID,
				DepartmentID: &deptID,
			},
			requiredScope: ScopeOwn,
			expected:      false,
		},
		{
			name: "department scope - same department",
			resourceCtx: &ResourceContext{
				ResourceID:  uuid.New(),
				TenantID:    tenantID,
				DepartmentID: &deptID,
			},
			requiredScope: ScopeDepartment,
			expected:      true,
		},
		{
			name: "department scope - different department",
			resourceCtx: &ResourceContext{
				ResourceID:  uuid.New(),
				TenantID:    tenantID,
				DepartmentID: uuidPtr(uuid.New()),
			},
			requiredScope: ScopeDepartment,
			expected:      false,
		},
		{
			name: "hospital scope - same hospital",
			resourceCtx: &ResourceContext{
				ResourceID:  uuid.New(),
				TenantID:    tenantID,
				HospitalID:  &hospitalID,
			},
			requiredScope: ScopeHospital,
			expected:      true,
		},
		{
			name: "hospital scope - different hospital",
			resourceCtx: &ResourceContext{
				ResourceID:  uuid.New(),
				TenantID:    tenantID,
				HospitalID:  uuidPtr(uuid.New()),
			},
			requiredScope: ScopeHospital,
			expected:      false,
		},
		{
			name: "organization scope - same organization",
			resourceCtx: &ResourceContext{
				ResourceID:      uuid.New(),
				TenantID:        tenantID,
				OrganizationID:  &orgID,
			},
			requiredScope: ScopeOrganization,
			expected:      true,
		},
		{
			name: "tenant scope - same tenant",
			resourceCtx: &ResourceContext{
				ResourceID: uuid.New(),
				TenantID:   tenantID,
			},
			requiredScope: ScopeTenant,
			expected:      true,
		},
		{
			name: "tenant scope - different tenant",
			resourceCtx: &ResourceContext{
				ResourceID: uuid.New(),
				TenantID:   uuid.New(),
			},
			requiredScope: ScopeTenant,
			expected:      false,
		},
		{
			name: "all scope - always allowed",
			resourceCtx: &ResourceContext{
				ResourceID: uuid.New(),
				TenantID:   uuid.New(),
			},
			requiredScope: ScopeAll,
			expected:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.Evaluate(userCtx, tt.resourceCtx, tt.requiredScope)
			if result != tt.expected {
				t.Errorf("Evaluate() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestScopeEvaluator_CanAccessResource(t *testing.T) {
	eval := NewScopeEvaluator()

	userID := uuid.New()
	tenantID := uuid.New()
	deptID := uuid.New()
	hospitalID := uuid.New()
	orgID := uuid.New()

	userCtx := &UserContext{
		UserID:         userID,
		TenantID:       tenantID,
		OrganizationID: &orgID,
		HospitalID:     &hospitalID,
		DepartmentID:   &deptID,
	}

	tests := []struct {
		name         string
		resourceCtx  *ResourceContext
		permission   *Permission
		expected     bool
	}{
		{
			name: "can access own resource",
			resourceCtx: &ResourceContext{
				ResourceID:  userID,
				TenantID:    tenantID,
				DepartmentID: &deptID,
			},
			permission: &Permission{Resource: "patient", Action: "read", Scope: ScopeOwn},
			expected:   true,
		},
		{
			name: "cannot access other user's resource with own scope",
			resourceCtx: &ResourceContext{
				ResourceID:  uuid.New(),
				TenantID:    tenantID,
				DepartmentID: &deptID,
			},
			permission: &Permission{Resource: "patient", Action: "read", Scope: ScopeOwn},
			expected:   false,
		},
		{
			name: "can access department resource",
			resourceCtx: &ResourceContext{
				ResourceID:  uuid.New(),
				TenantID:    tenantID,
				DepartmentID: &deptID,
			},
			permission: &Permission{Resource: "patient", Action: "read", Scope: ScopeDepartment},
			expected:   true,
		},
		{
			name: "can access hospital resource",
			resourceCtx: &ResourceContext{
				ResourceID:  uuid.New(),
				TenantID:    tenantID,
				HospitalID:  &hospitalID,
			},
			permission: &Permission{Resource: "patient", Action: "read", Scope: ScopeHospital},
			expected:   true,
		},
		{
			name: "can access tenant resource",
			resourceCtx: &ResourceContext{
				ResourceID: uuid.New(),
				TenantID:   tenantID,
			},
			permission: &Permission{Resource: "user", Action: "read", Scope: ScopeTenant},
			expected:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.CanAccessResource(userCtx, tt.resourceCtx, tt.permission)
			if result != tt.expected {
				t.Errorf("CanAccessResource() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestScopeEvaluator_GetEffectiveScope(t *testing.T) {
	eval := NewScopeEvaluator()

	userID := uuid.New()
	tenantID := uuid.New()
	deptID := uuid.New()
	hospitalID := uuid.New()
	orgID := uuid.New()

	userCtx := &UserContext{
		UserID:         userID,
		TenantID:       tenantID,
		OrganizationID: &orgID,
		HospitalID:     &hospitalID,
		DepartmentID:   &deptID,
	}

	tests := []struct {
		name        string
		resourceCtx *ResourceContext
		expected    string
	}{
		{
			name: "own scope for own resource",
			resourceCtx: &ResourceContext{
				ResourceID:  userID,
				TenantID:    tenantID,
				DepartmentID: &deptID,
			},
			expected: ScopeOwn,
		},
		{
			name: "department scope for same department",
			resourceCtx: &ResourceContext{
				ResourceID:  uuid.New(),
				TenantID:    tenantID,
				DepartmentID: &deptID,
			},
			expected: ScopeDepartment,
		},
		{
			name: "hospital scope for same hospital",
			resourceCtx: &ResourceContext{
				ResourceID:  uuid.New(),
				TenantID:    tenantID,
				HospitalID:  &hospitalID,
			},
			expected: ScopeHospital,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := eval.GetEffectiveScope(userCtx, tt.resourceCtx)
			if result != tt.expected {
				t.Errorf("GetEffectiveScope() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func uuidPtr(u uuid.UUID) *uuid.UUID {
	return &u
}