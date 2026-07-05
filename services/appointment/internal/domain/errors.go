package domain

import "errors"

var (
	ErrAppointmentNotFound = errors.New("appointment not found")
	ErrAppointmentExists   = errors.New("appointment already exists")
	ErrInvalidAppointment  = errors.New("invalid appointment")
	ErrInvalidStatusChange = errors.New("invalid appointment status change")
	ErrTenantScopeMismatch = errors.New("tenant scope mismatch")
)
