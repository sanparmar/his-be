package domain

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInvalidToken       = errors.New("invalid token")
	ErrSessionExpired     = errors.New("session expired")
	ErrInvalidSession     = errors.New("invalid session")
	ErrInternal           = errors.New("internal server error")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrMFAAlreadyEnabled  = errors.New("MFA already enabled")
	ErrMFAInvalidCode     = errors.New("invalid MFA code")
	ErrMFANotEnabled      = errors.New("MFA not enabled")
	ErrRoleNotFound       = errors.New("role not found")
	ErrRoleTenantMismatch = errors.New("role tenant mismatch")
)
