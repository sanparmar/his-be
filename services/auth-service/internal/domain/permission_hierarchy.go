package domain

import (
	"context"
	"github.com/google/uuid"
)

type PermissionHierarchy struct{}

func NewPermissionHierarchy() *PermissionHierarchy {
	return &PermissionHierarchy{}
}

func (h *PermissionHierarchy) Implies(has, needs string) bool {
	if has == needs {
		return true
	}

	hasPerm, err := ParsePermission(has)
	if err != nil {
		return false
	}

	needsPerm, err := ParsePermission(needs)
	if err != nil {
		return false
	}

	if hasPerm.Resource != needsPerm.Resource {
		return false
	}

	return impliesAction(hasPerm.Action, needsPerm.Action) && hasPerm.ScopeLevel() >= needsPerm.ScopeLevel()
}

func (h *PermissionHierarchy) ImpliesAll(has []string, needs string) bool {
	for _, h := range has {
		if h.Implies(needs) {
			return true
		}
	}
	return false
}

func impliesAction(has, needs string) bool {
	hierarchy := map[string]int{
		ActionRead:    0,
		ActionWrite:   1,
		ActionCreate:  1,
		ActionUpdate:  1,
		ActionDelete:  2,
		ActionApprove: 3,
		ActionSign:    2,
		ActionExport:  2,
		ActionManage:  3,
		ActionAssign:  2,
		ActionVerify:  2,
		ActionDispense: 2,
		ActionAdminister: 2,
		ActionCollect:  1,
		ActionInterpret: 2,
		ActionSchedule: 1,
		ActionClose:    2,
		ActionCancel:   1,
		ActionReschedule: 1,
		ActionApproveAppointment: 2,
		ActionProcess:  1,
		ActionAssess:   1,
		ActionAssign:   2,
		ActionDisposition: 2,
		ActionManageFastTrack: 2,
		ActionPost:     1,
		ActionAdjust:   2,
		ActionVoid:     3,
		ActionCreate:   1,
		ActionSubmit:   2,
		ActionAdjudicate: 3,
		ActionPostPayment: 2,
		ActionRefund:   3,
		ActionVerify:   2,
		ActionObtain:   2,
		ActionRun:      1,
		ActionAdmin:    3,
		ActionDeactivate: 2,
		ActionManageFormulary: 3,
	}
	return hierarchy[has] >= hierarchy[needs]
}

func (h *PermissionHierarchy) ExpandPermissions(perms []string) []string {
	expanded := make(map[string]bool)
	for _, p := range perms {
		expanded[p] = true
		perm, err := ParsePermission(p)
		if err != nil {
			continue
		}
		for action, level := range map[string]int{
			ActionRead:    0,
			ActionWrite:   1,
			ActionCreate:  1,
			ActionUpdate:  1,
			ActionDelete:  2,
			ActionApprove: 3,
			ActionSign:    2,
			ActionExport:  2,
			ActionManage:  3,
			ActionAssign:  2,
			ActionVerify:  2,
		} {
			if level <= actionHierarchyLevel(perm.Action) {
				for _, scope := range []string{ScopeOwn, ScopeDepartment, ScopeHospital, ScopeOrganization, ScopeTenant, ScopeAll} {
					if validScopes[scope] <= perm.ScopeLevel() {
						expanded[perm.Resource+":"+action+":"+scope] = true
					}
				}
			}
		}
	}

	result := make([]string, 0, len(expanded))
	for p := range expanded {
		result = append(result, p)
	}
	return result
}

func actionHierarchyLevel(action string) int {
	hierarchy := map[string]int{
		ActionRead:    0,
		ActionWrite:   1,
		ActionCreate:  1,
		ActionUpdate:  1,
		ActionDelete:  2,
		ActionApprove: 3,
		ActionSign:    2,
		ActionExport:  2,
		ActionManage:  3,
		ActionAssign:  2,
		ActionVerify:  2,
	}
	return hierarchy[action]
}