package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Role struct {
	ID          uuid.UUID  `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	DisplayName string     `json:"display_name" db:"display_name"`
	Category    string     `json:"category" db:"category"`
	ParentID    *uuid.UUID `json:"parent_id,omitempty" db:"parent_role_id"`
	IsSystem    bool       `json:"is_system" db:"is_system"`
	Description string     `json:"description" db:"description"`
	CreatedAt   time.Time  `json:"created_at" db:"created_at"`
	Permissions []Permission `json:"permissions,omitempty"`
}

func (r *Role) GetInheritedRoles(ctx context.Context, repo RoleRepository) ([]Role, error) {
	var roles []Role
	current := r.ParentID

	for current != nil {
		parent, err := repo.GetByID(ctx, *current)
		if err != nil {
			return nil, err
		}
		if parent == nil {
			break
		}
		roles = append(roles, *parent)
		current = parent.ParentID
	}

	return roles, nil
}

func (r *Role) AllPermissions() []Permission {
	var perms []Permission
	for _, p := range r.Permissions {
		perms = append(perms, p)
	}
	return perms
}