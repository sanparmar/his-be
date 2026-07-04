package domain

import "errors"

var (
	// ErrNotFound is returned when a patient is not found.
	ErrNotFound = errors.New("patient not found")

	// ErrAlreadyExists is returned when a patient already exists.
	ErrAlreadyExists = errors.New("patient already exists")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when user lacks required permissions.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrTenantMismatch is returned when tenant isolation check fails.
	ErrTenantMismatch = errors.New("tenant mismatch")
)
