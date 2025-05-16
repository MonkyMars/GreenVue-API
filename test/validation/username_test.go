package validation_test

import (
	"greenvue/lib/validation"
	"testing"
)

func TestUsernameValidation(t *testing.T) {
	// Test cases structure: input username, expected validity, reason for failing (if not valid)
	testCases := []struct {
		name           string
		username       string
		expectedValid  bool
		expectedReason string
	}{
		{
			name:           "Valid username",
			username:       "johndoe123",
			expectedValid:  true,
			expectedReason: "",
		},
		{
			name:           "Username too short",
			username:       "ab",
			expectedValid:  false,
			expectedReason: "Username must be between 3 and 30 characters",
		},
		{
			name:           "Username too long",
			username:       "thisusernameistoolongandexceedsthe30characterslimit12345",
			expectedValid:  false,
			expectedReason: "Username must be between 3 and 30 characters",
		},
		{
			name:           "Username with reserved word",
			username:       "admin_user",
			expectedValid:  false,
			expectedReason: "Username contains reserved words",
		},
		{
			name:           "Username with repeating characters",
			username:       "helloooooworld",
			expectedValid:  false,
			expectedReason: "Username contains too many repeating characters",
		},
		{
			name:           "Username with invalid characters",
			username:       "user@name",
			expectedValid:  false,
			expectedReason: "Username can only contain letters, numbers, and underscores",
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test using the convenience function
			valid, reason := validation.ValidateUsername(tc.username)
			if valid != tc.expectedValid {
				t.Errorf("Expected validity to be %v, got %v", tc.expectedValid, valid)
			}
			if !valid && reason != tc.expectedReason {
				t.Errorf("Expected reason '%s', got '%s'", tc.expectedReason, reason)
			}

			// Test using the validator object directly
			validator := validation.NewUsernameValidator()
			valid, reason = validator.Validate(tc.username)
			if valid != tc.expectedValid {
				t.Errorf("Validator object: Expected validity to be %v, got %v", tc.expectedValid, valid)
			}
			if !valid && reason != tc.expectedReason {
				t.Errorf("Validator object: Expected reason '%s', got '%s'", tc.expectedReason, reason)
			}
		})
	}
}

func TestCustomUsernameValidator(t *testing.T) {
	// Test a custom validator with different settings
	validator := validation.NewUsernameValidator()
	validator.MinLength = 5
	validator.MaxLength = 20
	validator.BlockedWords = []string{"test", "dummy"}
	validator.RepeatLimit = 3

	// Test custom length validation
	valid, reason := validator.Validate("user")
	if valid {
		t.Errorf("Expected 'user' to be invalid with MinLength=5")
	}
	if reason != "Username must be between 5 and 20 characters" {
		t.Errorf("Wrong reason: %s", reason)
	}

	// Test custom blocked words
	valid, reason = validator.Validate("testuser123")
	if valid {
		t.Errorf("Expected 'testuser123' to be invalid with BlockedWords containing 'test'")
	}
	if reason != "Username contains reserved words" {
		t.Errorf("Wrong reason: %s", reason)
	}

	// Test custom repeating characters limit
	valid, reason = validator.Validate("helloooworld")
	if valid {
		t.Errorf("Expected 'helloooworld' to be invalid with RepeatLimit=3")
	}
	if reason != "Username contains too many repeating characters" {
		t.Errorf("Wrong reason: %s", reason)
	}

	// Test valid username with custom settings
	valid, reason = validator.Validate("valid_user123")
	if !valid {
		t.Errorf("Expected 'valid_user123' to be valid, but got reason: %s", reason)
	}
}
