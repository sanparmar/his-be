package application

import (
	"context"

	"github.com/deloitte-us-consulting/his-be/services/auth-service/internal/domain"
)

type ListRolesUseCase struct {
	roleRepo domain.RoleRepository
}

func NewListRolesUseCase(roleRepo domain.RoleRepository) *ListRolesUseCase {
	return &ListRolesUseCase{roleRepo: roleRepo}
}

type ListRolesResponse struct {
	Roles []RoleResponse
}

func (uc *ListRolesUseCase) Execute(ctx context.Context, req ListRolesRequest) (*ListRolesResponse, error) {
	roles, err := uc.roleRepo.List(ctx)
	if err != nil {
		return nil, err
	}

	var resp []RoleResponse
	for _, r := range roles {
		resp = append(resp, RoleResponse{
			ID:          r.ID,
			Name:        r.Name,
			DisplayName: r.DisplayName,
			Category:    r.Category,
			IsSystem:    r.IsSystem,
			Description: r.Description,
		})
	}

	return &ListRolesResponse{Roles: resp}, nil
}