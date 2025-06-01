package validation

import (
	"fmt"
	"greenvue/lib"
	"strings"
	"unicode"
)

// PasswordValidator provides validation for passwords
type PasswordValidator struct {
	MinLength        int
	MaxLength        int
	RequireSpecial   bool
	RequireUppercase bool
	RequireLowercase bool
	RequireDigit     bool
}

// NewPasswordValidator creates a validator with default settings
func NewPasswordValidator() *PasswordValidator {
	return &PasswordValidator{
		MinLength:        8,
		MaxLength:        64,
		RequireSpecial:   true,
		RequireUppercase: true,
		RequireLowercase: true,
		RequireDigit:     true,
	}
}

// Validate checks if a password is valid and returns result with reason
func (v *PasswordValidator) Validate(password string) (bool, string) {
	if len(password) < v.MinLength {
		return false, fmt.Sprintf("Password must be at least %d characters long", v.MinLength)
	}
	if len(password) > v.MaxLength {
		return false, fmt.Sprintf("Password must be at most %d characters long", v.MaxLength)
	}
	if v.RequireSpecial && !containsSpecialCharacter(password) {
		return false, "Password must contain at least one special character"
	}
	if v.RequireUppercase && !containsUppercaseCharacter(password) {
		return false, "Password must contain at least one uppercase letter"
	}
	if v.RequireLowercase && !containsLowercaseCharacter(password) {
		return false, "Password must contain at least one lowercase letter"
	}
	if v.RequireDigit && !containsDigit(password) {
		return false, "Password must contain at least one digit"
	}
	return true, ""
}

// containsSpecialCharacter checks if the password contains special characters
func containsSpecialCharacter(s string) bool {
	specialChars := `!@#$%^&*()-_=+[]{}|;:'",.<>?/`
	for _, char := range s {
		if strings.ContainsRune(specialChars, char) {
			return true
		}
	}
	return false
}

// containsUppercaseCharacter checks if the password contains uppercase letters
func containsUppercaseCharacter(s string) bool {
	for _, char := range s {
		if unicode.IsUpper(char) {
			return true
		}
	}
	return false
}

// containsLowercaseCharacter checks if the password contains lowercase letters
func containsLowercaseCharacter(s string) bool {
	for _, char := range s {
		if unicode.IsLower(char) {
			return true
		}
	}
	return false
}

// containsDigit checks if the password contains digits
func containsDigit(s string) bool {
	for _, char := range s {
		if unicode.IsDigit(char) {
			return true
		}
	}
	return false
}

// SanitizePassword removes any unsafe characters from the password
func SanitizePassword(password string) string {
	// Remove any unsafe characters
	sanitized := lib.SanitizeInput(password)
	// Ensure the sanitized password is still within length limits
	if len(sanitized) < 8 {
		return ""
	}
	if len(sanitized) > 64 {
		return sanitized[:64]
	}
	return sanitized
}

// ValidatePassword checks if a password meets the validation criteria
func ValidatePassword(password string) (bool, string) {
	validator := NewPasswordValidator()
	return validator.Validate(password)
}
