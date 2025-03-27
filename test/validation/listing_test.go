package validation_test

import (
	"greentrade-eu/lib/validation"
	"testing"
)

func TestListingValidation(t *testing.T) {
	// Test cases structure: input fields, expected valid flag, expected errors
	testCases := []struct {
		name              string
		title             string
		description       string
		category          string
		condition         string
		price             int64
		expectValid       bool
		expectedErrFields []string
	}{
		{
			name:              "Valid listing",
			title:             "iPhone 12 Pro Max",
			description:       "Lightly used iPhone 12 Pro Max in excellent condition",
			category:          "electronics",
			condition:         "good",
			price:             70000,
			expectValid:       true,
			expectedErrFields: nil,
		},
		{
			name:              "Title too short",
			title:             "iPad",
			description:       "Lightly used iPad in excellent condition",
			category:          "electronics",
			condition:         "good",
			price:             50000,
			expectValid:       false,
			expectedErrFields: []string{"title"},
		},
		{
			name:              "Description too short",
			title:             "iPhone 12 Pro Max",
			description:       "Short desc",
			category:          "electronics",
			condition:         "good",
			price:             70000,
			expectValid:       false,
			expectedErrFields: []string{"description"},
		},
		{
			name:              "Invalid category",
			title:             "iPhone 12 Pro Max",
			description:       "Lightly used iPhone 12 Pro Max in excellent condition",
			category:          "invalid-category",
			condition:         "good",
			price:             70000,
			expectValid:       false,
			expectedErrFields: []string{"category"},
		},
		{
			name:              "Invalid condition",
			title:             "iPhone 12 Pro Max",
			description:       "Lightly used iPhone 12 Pro Max in excellent condition",
			category:          "electronics",
			condition:         "invalid-condition",
			price:             70000,
			expectValid:       false,
			expectedErrFields: []string{"condition"},
		},
		{
			name:              "Negative price",
			title:             "iPhone 12 Pro Max",
			description:       "Lightly used iPhone 12 Pro Max in excellent condition",
			category:          "electronics",
			condition:         "good",
			price:             -100,
			expectValid:       false,
			expectedErrFields: []string{"price"},
		},
		{
			name:              "Multiple validation errors",
			title:             "iPad",
			description:       "Short",
			category:          "invalid",
			condition:         "unknown",
			price:             -100,
			expectValid:       false,
			expectedErrFields: []string{"title", "description", "category", "condition", "price"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Test using the convenience function
			result := validation.ValidateListing(tc.title, tc.description, tc.category, tc.condition, tc.price)

			if result.Valid != tc.expectValid {
				t.Errorf("Expected validity to be %v, got %v", tc.expectValid, result.Valid)
			}

			// Check if all expected error fields exist
			if !tc.expectValid {
				for _, field := range tc.expectedErrFields {
					if _, exists := result.Errors[field]; !exists {
						t.Errorf("Expected error for field '%s', but none found", field)
					}
				}

				// Check if there are no unexpected error fields
				if len(result.Errors) != len(tc.expectedErrFields) {
					t.Errorf("Expected %d errors, got %d errors", len(tc.expectedErrFields), len(result.Errors))
				}
			}
		})
	}
}

func TestCustomListingValidator(t *testing.T) {
	// Create a custom validator with more restrictive settings
	validator := validation.NewListingValidator()
	validator.TitleMinLength = 10
	validator.TitleMaxLength = 50
	validator.DescriptionMinLength = 50
	validator.MinPrice = 1000
	validator.MaxPrice = 50000
	validator.AllowedCategories = []string{"sustainable", "eco-friendly"}
	validator.AllowedConditions = []string{"new", "like_new"}

	// Test customized validation
	result := validator.ValidateListing(
		"Short",             // too short title
		"Short description", // too short description
		"electronics",       // not in allowed categories
		"good",              // not in allowed conditions
		500,                 // below minimum price
	)

	if result.Valid {
		t.Error("Expected invalid result with custom validator settings")
	}

	expectedErrors := []string{"title", "description", "category", "condition", "price"}
	for _, field := range expectedErrors {
		if _, exists := result.Errors[field]; !exists {
			t.Errorf("Expected error for field '%s', but none found", field)
		}
	}

	// Test valid listing with custom settings
	result = validator.ValidateListing(
		"This is a valid title for sustainable product",
		"This description is long enough to meet the custom minimum length requirement for our sustainable product descriptions that we have set in our custom validator.",
		"sustainable",
		"new",
		30000,
	)

	if !result.Valid {
		t.Errorf("Expected valid result, but got errors: %v", result.Errors)
	}
}
