package validation

import (
	"errors"
	"fmt"
)

var (
	ErrInvalidUsername       = errors.New("username is required")
	ErrInvalidUsernameFormat = errors.New("username must be 3-64 characters, alphanumeric with ._-")
	ErrInvalidEmail          = errors.New("email is required")
	ErrInvalidEmailFormat    = errors.New("invalid email format")
	ErrInvalidPassword       = errors.New("password is required")
	ErrInvalidPasswordLength = errors.New("password must be 8-128 characters")
	ErrInvalidMFAType        = errors.New("mfa_type must be 'totp' or 'webauthn'")
)

func ErrMissingUUID(field string) error {
	return fmt.Errorf("%s is required", field)
}

func ErrInvalidUUID(field string) error {
	return fmt.Errorf("%s must be a valid UUID", field)
}