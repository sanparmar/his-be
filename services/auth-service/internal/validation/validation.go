package validation

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/google/uuid"
)

var (
	emailRegex    = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9._-]{3,50}$`)
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	var msgs []string
	for _, e := range v {
		msgs = append(msgs, e.Error())
	}
	return strings.Join(msgs, "; ")
}

func ValidateUsername(username string) error {
	if username == "" {
		return ValidationError{Field: "username", Message: "username is required"}
	}
	if len(username) < 3 || len(username) > 50 {
		return ValidationError{Field: "username", Message: "username must be between 3 and 50 characters"}
	}
	if !usernameRegex.MatchString(username) {
		return ValidationError{Field: "username", Message: "username can only contain letters, numbers, dots, underscore, and hyphen"}
	}
	return nil
}

func ValidateEmail(email string) error {
	if email == "" {
		return ValidationError{Field: "email", Message: "email is required"}
	}
	if !emailRegex.MatchString(email) {
		return ValidationError{Field: "email", Message: "invalid email format"}
	}
	return nil
}

func ValidatePassword(password string) error {
	if password == "" {
		return ValidationError{Field: "password", Message: "password is required"}
	}
	if len(password) < 8 {
		return ValidationError{Field: "password", Message: "password must be at least 8 characters"}
	}
	if len(password) > 128 {
		return ValidationError{Field: "password", Message: "password must not exceed 128 characters"}
	}

	hasUpper := regexp.MustCompile(`[A-Z]`).MatchString(password)
	hasLower := regexp.MustCompile(`[a-z]`).MatchString(password)
	hasNumber := regexp.MustCompile(`[0-9]`).MatchString(password)
	hasSpecial := regexp.MustCompile(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`).MatchString(password)

	if !hasUpper {
		return ValidationError{Field: "password", Message: "password must contain at least one uppercase letter"}
	}
	if !hasLower {
		return ValidationError{Field: "password", Message: "password must contain at least one lowercase letter"}
	}
	if !hasNumber {
		return ValidationError{Field: "password", Message: "password must contain at least one number"}
	}
	if !hasSpecial {
		return ValidationError{Field: "password", Message: "password must contain at least one special character"}
	}

	return nil
}

func ValidateUUID(id string, fieldName string) error {
	if id == "" {
		return ValidationError{Field: fieldName, Message: fieldName + " is required"}
	}
	if _, err := uuid.Parse(id); err != nil {
		return ValidationError{Field: fieldName, Message: fieldName + " must be a valid UUID"}
	}
	return nil
}

func ValidateUUIDs(ids []string, fieldName string) error {
	if len(ids) == 0 {
		return nil // Empty is valid for optional lists
	}
	for i, id := range ids {
		if _, err := uuid.Parse(id); err != nil {
			return ValidationError{Field: fmt.Sprintf("%s[%d]", fieldName, i), Message: "must be a valid UUID"}
		}
	}
	return nil
}

func ValidateMFAType(mfaType string) error {
	if mfaType == "" {
		return nil // Optional
	}
	validTypes := map[string]bool{"totp": true, "webauthn": true}
	if !validTypes[mfaType] {
		return ValidationError{Field: "mfa_type", Message: "mfa_type must be 'totp' or 'webauthn'"}
	}
	return nil
}

func ValidateTenantID(tenantID string) error {
	return ValidateUUID(tenantID, "tenant_id")
}

func ValidateOrganizationID(orgID string) error {
	if orgID == "" {
		return nil // Optional
	}
	return ValidateUUID(orgID, "organization_id")
}

func ValidateHospitalID(hospitalID string) error {
	if hospitalID == "" {
		return nil // Optional
	}
	return ValidateUUID(hospitalID, "hospital_id")
}

func ValidateUserID(userID string) error {
	return ValidateUUID(userID, "user_id")
}

func ValidateRoleIDs(roleIDs []string) error {
	return ValidateUUIDs(roleIDs, "role_ids")
}

func ValidateRemoveRoleIDs(roleIDs []string) error {
	return ValidateUUIDs(roleIDs, "remove_role_ids")
}

func ValidateOptionalUUID(id string, fieldName string) error {
	if id == "" {
		return nil
	}
	if _, err := uuid.Parse(id); err != nil {
		return ValidationError{Field: fieldName, Message: fieldName + " must be a valid UUID"}
	}
	return nil
}