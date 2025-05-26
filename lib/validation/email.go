package validation

import (
	"fmt"
	"regexp"
)

// EmailValidator provides validation for email addresses
type EmailValidator struct {
	MinLength int
	MaxLength int
}

// NewEmailValidator creates a validator with default settings
func NewEmailValidator() *EmailValidator {
	return &EmailValidator{
		MinLength: 5,
		MaxLength: 254, // Maximum length for email addresses
	}
}

// Validate checks if an email is valid and returns result with reason
func (v *EmailValidator) Validate(email string) (bool, string) {
	// Length checks
	if len(email) < v.MinLength || len(email) > v.MaxLength {
		return false, fmt.Sprintf("Email must be between %d and %d characters", v.MinLength, v.MaxLength)
	}

	// Basic email format validation using regex
	validEmail := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !validEmail.MatchString(email) {
		return false, "Invalid email format"
	}

	return true, ""
}

func ValidateEmail(email string) (bool, string) {
	validator := NewEmailValidator()
	return validator.Validate(email)
}
