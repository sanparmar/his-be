package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
)

type ListPermissionsUseCase struct {
	permRepo domain.PermissionRepository
}

func NewListPermissionsUseCase(permRepo domain.PermissionRepository) *ListPermissionsUseCase {
	return &ListPermissionsUseCase{permRepo: permRepo}
}

type ListPermissionsRequest struct {
	Category string
	Resource string
}

type ListPermissionsResponse struct {
	Permissions []PermissionResponse
}

func (uc *ListPermissionsUseCase) Execute(ctx context.Context, req ListPermissionsRequest) (*ListPermissionsResponse, error) {
	var perms []domain.Permission
	var err error

	if req.Category != "" {
		perms, err = uc.permRepo.ListByCategory(ctx, req.Category)
	} else {
		perms, err = uc.permRepo.ListAll(ctx)
	}

	if err != nil {
		return nil, err
	}

	if req.Resource != "" {
		filtered := make([]domain.Permission, 0)
		for _, p := range perms {
			if p.Resource == req.Resource {
				filtered = append(filtered, p)
			}
		}
		perms = filtered
	}

	var resp []PermissionResponse
	for _, p := range perms {
		resp = append(resp, PermissionResponse{
			ID:          p.ID,
			Name:        p.Name,
			Resource:    p.Resource,
			Action:      p.Action,
			Scope:       p.Scope,
			Category:    p.Category,
			Description: p.Description,
		})
	}

	return &ListPermissionsResponse{Permissions: resp}, nil
}