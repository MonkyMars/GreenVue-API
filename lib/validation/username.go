package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// UsernameValidator provides validation for usernames
type UsernameValidator struct {
	MinLength    int
	MaxLength    int
	BlockedWords []string
	RepeatLimit  int
}

// NewUsernameValidator creates a validator with default settings
func NewUsernameValidator() *UsernameValidator {
	return &UsernameValidator{
		MinLength:    3,
		MaxLength:    30,
		BlockedWords: []string{"admin", "moderator", "support", "staff"},
		RepeatLimit:  5,
	}
}

// Validate checks if a username is valid and returns result with reason
func (v *UsernameValidator) Validate(username string) (bool, string) {
	// Length checks
	if len(username) < v.MinLength || len(username) > v.MaxLength {
		return false, fmt.Sprintf("Username must be between %d and %d characters", v.MinLength, v.MaxLength)
	}

	// Blocked words check
	lowerUsername := strings.ToLower(username)
	for _, word := range v.BlockedWords {
		if strings.Contains(lowerUsername, word) {
			return false, "Username contains reserved words"
		}
	}

	// Repeat character check
	if v.hasExcessiveRepeatingChars(username) {
		return false, "Username contains too many repeating characters"
	}

	// Character validation
	validChars := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validChars.MatchString(username) {
		return false, "Username can only contain letters, numbers, and underscores"
	}

	return true, ""
}

// hasExcessiveRepeatingChars checks for repeating characters
func (v *UsernameValidator) hasExcessiveRepeatingChars(s string) bool {
	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= v.RepeatLimit {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

// ValidateUsername is a convenience function using the default validator
func ValidateUsername(username string) (bool, string) {
	validator := NewUsernameValidator()
	return validator.Validate(username)
}
