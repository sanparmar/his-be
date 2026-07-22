package domain

import (
	"testing"

	"github.com/google/uuid"
)

func TestPermission_ParsePermission(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    *Permission
	}{
		{
			name:     "valid permission - own scope",
			input:    "patient:read:own",
			expected: &Permission{Resource: "patient", Action: "read", Scope: "own"},
		},
		{
			name:     "valid permission - department scope",
			input:    "patient:write:department",
			expected: &Permission{Resource: "patient", Action: "write", Scope: "department"},
		},
		{
			name:     "valid permission - all scope",
			input:    "patient:read:all",
			expected: &Permission{Resource: "patient", Action: "read", Scope: "all"},
		},
		{
			name:        "invalid format - too few parts",
			input:       "patient:read",
			expectError: true,
		},
		{
			name:        "invalid format - too many parts",
			input:       "patient:read:own:extra",
			expectError: true,
		},
		{
			name:        "invalid resource",
			input:       "invalid:read:own",
			expectError: true,
		},
		{
			name:        "invalid action",
			input:       "patient:invalid:own",
			expectError: true,
		},
		{
			name:        "invalid scope",
			input:       "patient:read:invalid",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParsePermission(tt.input)
			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if result.Resource != tt.expected.Resource {
				t.Errorf("expected resource %s, got %s", tt.expected.Resource, result.Resource)
			}
			if result.Action != tt.expected.Action {
				t.Errorf("expected action %s, got %s", tt.expected.Action, result.Action)
			}
			if result.Scope != tt.expected.Scope {
				t.Errorf("expected scope %s, got %s", tt.expected.Scope, result.Scope)
			}
		})
	}
}

func TestPermission_ScopeLevel(t *testing.T) {
	tests := []struct {
		scope     string
		expected  int
	}{
		{ScopeOwn, 0},
		{ScopeDepartment, 1},
		{ScopeHospital, 2},
		{ScopeOrganization, 3},
		{ScopeTenant, 4},
		{ScopeAll, 5},
	}

	for _, tt := range tests {
		p := &Permission{Scope: tt.scope}
		if level := p.ScopeLevel(); level != tt.expected {
			t.Errorf("scope %s: expected level %d, got %d", tt.scope, tt.expected, level)
		}
	}
}

func TestPermission_Implies(t *testing.T) {
	tests := []struct {
		name     string
		has      *Permission
		needs    *Permission
		expected bool
	}{
		{
			name:     "same permission",
			has:      &Permission{Resource: "patient", Action: "read", Scope: "own"},
			needs:    &Permission{Resource: "patient", Action: "read", Scope: "own"},
			expected: true,
		},
		{
			name:     "higher scope implies lower",
			has:      &Permission{Resource: "patient", Action: "read", Scope: "department"},
			needs:    &Permission{Resource: "patient", Action: "read", Scope: "own"},
			expected: true,
		},
		{
			name:     "write implies read",
			has:      &Permission{Resource: "patient", Action: "write", Scope: "own"},
			needs:    &Permission{Resource: "patient", Action: "read", Scope: "own"},
			expected: true,
		},
		{
			name:     "approve implies write and read",
			has:      &Permission{Resource: "order", Action: "approve", Scope: "own"},
			needs:    &Permission{Resource: "order", Action: "read", Scope: "own"},
			expected: true,
		},
		{
			name:     "different resource",
			has:      &Permission{Resource: "patient", Action: "read", Scope: "own"},
			needs:    &Permission{Resource: "order", Action: "read", Scope: "own"},
			expected: false,
		},
		{
			name:     "lower scope does not imply higher",
			has:      &Permission{Resource: "patient", Action: "read", Scope: "own"},
			needs:    &Permission{Resource: "patient", Action: "read", Scope: "department"},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.has.Implies(tt.needs)
			if result != tt.expected {
				t.Errorf("Implies() = %v, want %v", result, tt.expected)
			}
		})
	}
}