package domain

import (
	"errors"
	"strings"
)

var (
	ErrInvalidPermissionFormat = errors.New("invalid permission format, expected resource:action:scope")
	ErrInvalidResource         = errors.New("invalid resource")
	ErrInvalidAction           = errors.New("invalid action")
	ErrInvalidScope            = errors.New("invalid scope")
)

const (
	ScopeOwn         = "own"
	ScopeDepartment  = "department"
	ScopeHospital    = "hospital"
	ScopeOrganization = "organization"
	ScopeTenant      = "tenant"
	ScopeAll         = "all"
)

const (
	ActionRead       = "read"
	ActionWrite      = "write"
	ActionCreate     = "create"
	ActionUpdate     = "update"
	ActionDelete     = "delete"
	ActionClose      = "close"
	ActionApprove    = "approve"
	ActionSign       = "sign"
	ActionExport     = "export"
	ActionDiscontinue = "discontinue"
	ActionAdminister = "administer"
	ActionDispense   = "dispense"
	ActionCode       = "code"
	ActionOrder      = "order"
	ActionResultRead = "result:read"
	ActionResultVerify = "result:verify"
	ActionResultApprove = "result:approve"
	ActionCollect    = "collect"
	ActionInterpret  = "interpret"
	ActionAssign     = "assign"
	ActionTransfer   = "transfer"
	ActionRelease    = "release"
	ActionSchedule   = "schedule"
	ActionPerform    = "perform"
	ActionAssist     = "assist"
	ActionAdmit      = "admit"
	ActionManage     = "manage"
	ActionReschedule = "reschedule"
	ActionCancel     = "cancel"
	ActionConduct    = "conduct"
	ActionProcess    = "process"
	ActionAssess     = "assess"
	ActionDisposition = "disposition"
	ActionPost       = "post"
	ActionGenerate   = "generate"
	ActionAdjust     = "adjust"
	ActionVoid       = "void"
	ActionSubmit     = "submit"
	ActionAdjudicate = "adjudicate"
	ActionRefund     = "refund"
	ActionVerify     = "verify"
	ActionObtain     = "obtain"
	ActionRun        = "run"
	ActionDeactivate = "deactivate"
	ActionAdmin    = "admin"
	ActionManageFormulary = "manage"
)

var validScopes = map[string]int{
	ScopeOwn:         0,
	ScopeDepartment:  1,
	ScopeHospital:    2,
	ScopeOrganization: 3,
	ScopeTenant:      4,
	ScopeAll:         5,
}

var validResources = map[string]bool{
	"patient":      true,
	"encounter":    true,
	"admission":    true,
	"transfer":     true,
	"discharge":    true,
	"clinical":     true,
	"order":        true,
	"medication":   true,
	"diagnosis":    true,
	"procedure":    true,
	"allergy":      true,
	"vitals":       true,
	"lab":          true,
	"specimen":     true,
	"rad":          true,
	"blood":        true,
	"ipd":          true,
	"bed":          true,
	"surgery":      true,
	"anesthesia":   true,
	"icu":          true,
	"ventilator":   true,
	"opd":          true,
	"appointment":  true,
	"telemedicine": true,
	"checkin":      true,
	"ed":           true,
	"triage":       true,
	"fasttrack":    true,
	"billing":      true,
	"invoice":      true,
	"claim":        true,
	"payment":      true,
	"coverage":     true,
	"authorization": true,
	"pharmacy":     true,
	"formulary":    true,
	"interaction":  true,
	"stock":        true,
	"controlled":   true,
	"user":         true,
	"role":         true,
	"audit":        true,
	"config":       true,
	"facility":     true,
	"report":       true,
}

var validActions = map[string]bool{
	"read":          true,
	"write":         true,
	"create":        true,
	"update":        true,
	"delete":        true,
	"close":         true,
	"approve":       true,
	"sign":          true,
	"export":        true,
	"discontinue":   true,
	"administer":    true,
	"dispense":      true,
	"code":          true,
	"order":         true,
	"result:read":   true,
	"result:verify": true,
	"result:approve": true,
	"collect":       true,
	"interpret":     true,
	"assign":        true,
	"transfer":      true,
	"release":       true,
	"schedule":      true,
	"perform":       true,
	"assist":        true,
	"admit":         true,
	"manage":        true,
	"reschedule":    true,
	"cancel":        true,
	"conduct":       true,
	"process":       true,
	"assess":        true,
	"disposition":   true,
	"post":          true,
	"generate":      true,
	"adjust":        true,
	"void":          true,
	"submit":        true,
	"adjudicate":    true,
	"refund":        true,
	"verify":        true,
	"obtain":        true,
	"run":           true,
	"deactivate":    true,
}

type Permission struct {
	ID          string `json:"id" db:"id"`
	Name        string `json:"name" db:"name"`
	Resource    string `json:"resource" db:"resource"`
	Action      string `json:"action" db:"action"`
	Scope       string `json:"scope" db:"scope"`
	Category    string `json:"category" db:"category"`
	Description string `json:"description" db:"description"`
}

// ParsePermission accepts both the stored `resource:action` name format
// (his-be's seeded permissions.name column — scope lives in its own DB
// column, never appended to name) and an explicit `resource:action:scope`
// form. A bare `resource:action` string defaults to the most restrictive
// scope (own) since no scope is available to parse. Some actions are
// themselves compound (e.g. "result:approve" for lab/rad permissions), so
// the whole remainder is tried as the action first before assuming a
// trailing scope segment.
func ParsePermission(s string) (*Permission, error) {
	firstColon := strings.Index(s, ":")
	if firstColon < 0 || firstColon == len(s)-1 {
		return nil, ErrInvalidPermissionFormat
	}
	resource := s[:firstColon]
	rest := s[firstColon+1:]

	if !validResources[resource] {
		return nil, ErrInvalidResource
	}

	action, scope := rest, ScopeOwn
	if !validActions[action] {
		if lastColon := strings.LastIndex(rest, ":"); lastColon >= 0 {
			action, scope = rest[:lastColon], rest[lastColon+1:]
		}
	}

	if !validActions[action] {
		return nil, ErrInvalidAction
	}
	if _, ok := validScopes[scope]; !ok {
		return nil, ErrInvalidScope
	}

	return &Permission{
		Name:     s,
		Resource: resource,
		Action:   action,
		Scope:    scope,
	}, nil
}

func (p *Permission) ScopeLevel() int {
	return validScopes[p.Scope]
}

func (p *Permission) String() string {
	return p.Name
}

func (p *Permission) Matches(resource, action, scope string) bool {
	if p.Resource != resource {
		return false
	}
	if p.Action != action {
		return false
	}
	return p.ScopeLevel() >= validScopes[scope]
}

func ValidScopes() []string {
	scopes := make([]string, 0, len(validScopes))
	for s := range validScopes {
		scopes = append(scopes, s)
	}
	return scopes
}

func ValidResources() []string {
	resources := make([]string, 0, len(validResources))
	for r := range validResources {
		resources = append(resources, r)
	}
	return resources
}

func ValidActions() []string {
	actions := make([]string, 0, len(validActions))
	for a := range validActions {
		actions = append(actions, a)
	}
	return actions
}