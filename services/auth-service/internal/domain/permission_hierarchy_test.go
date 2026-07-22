package domain

import (
	"testing"
)

func TestPermissionHierarchy_Implies(t *testing.T) {
	h := NewPermissionHierarchy()

	tests := []struct {
		name     string
		has      string
		needs    string
		expected bool
	}{
		{
			name:     "same permission",
			has:      "patient:read:own",
			needs:    "patient:read:own",
			expected: true,
		},
		{
			name:     "write implies read",
			has:      "patient:write:own",
			needs:    "patient:read:own",
			expected: true,
		},
		{
			name:     "approve implies write and read",
			has:      "order:approve:own",
			needs:    "order:read:own",
			expected: true,
		},
		{
			name:     "approve implies write",
			has:      "order:approve:own",
			needs:    "order:write:own",
			expected: true,
		},
		{
			name:     "delete implies write and read",
			has:      "patient:delete:own",
			needs:    "patient:read:own",
			expected: true,
		},
		{
			name:     "manage implies all",
			has:      "patient:manage:own",
			needs:    "patient:delete:own",
			expected: true,
		},
		{
			name:     "different resource",
			has:      "patient:read:own",
			needs:    "order:read:own",
			expected: false,
		},
		{
			name:     "lower scope does not imply higher",
			has:      "patient:read:own",
			needs:    "patient:read:department",
			expected: false,
		},
		{
			name:     "higher scope implies lower",
			has:      "patient:read:department",
			needs:    "patient:read:own",
			expected: true,
		},
		{
			name:     "sign implies write and read",
			has:      "clinical:sign:own",
			needs:    "clinical:read:own",
			expected: true,
		},
		{
			name:     "verify implies write and read",
			has:      "lab:result:verify:department",
			needs:    "lab:result:read:own",
			expected: true,
		},
		{
			name:     "dispense implies write and read",
			has:      "pharmacy:dispense:department",
			needs:    "pharmacy:read:own",
			expected: true,
		},
		{
			name:     "administer implies write and read",
			has:      "medication:administer:own",
			needs:    "medication:read:own",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.Implies(tt.has, tt.needs)
			if result != tt.expected {
				t.Errorf("Implies(%q, %q) = %v, want %v", tt.has, tt.needs, result, tt.expected)
			}
		})
	}
}

func TestPermissionHierarchy_ImpliesAll(t *testing.T) {
	h := NewPermissionHierarchy()

	has := []string{
		"patient:write:own",
		"order:approve:own",
		"clinical:read:own",
	}

	tests := []struct {
		name     string
		needs    string
		expected bool
	}{
		{
			name:     "has write implies read",
			needs:    "patient:read:own",
			expected: true,
		},
		{
			name:     "has approve implies read",
			needs:    "order:read:own",
			expected: true,
		},
		{
			name:     "missing permission",
			needs:    "patient:delete:own",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := h.ImpliesAll(has, tt.needs)
			if result != tt.expected {
				t.Errorf("ImpliesAll() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestPermissionHierarchy_ExpandPermissions(t *testing.T) {
	h := NewPermissionHierarchy()

	perms := []string{
		"patient:write:own",
		"order:approve:own",
	}

	expanded := h.ExpandPermissions(perms)

	if len(expanded) <= len(perms) {
		t.Errorf("ExpandPermissions() should return more permissions than input, got %d vs %d", len(expanded), len(perms))
	}

	foundRead := false
	for _, p := range expanded {
		if p == "patient:read:own" {
			foundRead = true
			break
		}
	}
	if !foundRead {
		t.Error("expanded permissions should include patient:read:own")
	}
}

func TestActionLevel(t *testing.T) {
	tests := []struct {
		action  string
		level   int
	}{
		{ActionRead, 0},
		{ActionWrite, 1},
		{ActionCreate, 1},
		{ActionDelete, 2},
		{ActionApprove, 3},
		{ActionSign, 2},
		{ActionManage, 3},
	}

	for _, tt := range tests {
		t.Run(tt.action, func(t *testing.T) {
			level := ActionLevel(tt.action)
			if level != tt.level {
				t.Errorf("ActionLevel(%s) = %d, want %d", tt.action, level, tt.level)
			}
		})
	}
}