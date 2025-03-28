package validation

import (
	"fmt"
	"strings"
)

// ListingValidator provides validation for marketplace listings
type ListingValidator struct {
	TitleMinLength       int
	TitleMaxLength       int
	DescriptionMinLength int
	DescriptionMaxLength int
	MinPrice             int64
	MaxPrice             int64
	AllowedCategories    []string
	AllowedConditions    []string
}

// NewListingValidator creates a validator with default settings
func NewListingValidator() *ListingValidator {
	return &ListingValidator{
		TitleMinLength:       5,
		TitleMaxLength:       100,
		DescriptionMinLength: 20,
		DescriptionMaxLength: 2000,
		MinPrice:             0,
		MaxPrice:             1000000,
		AllowedCategories: []string{
			"electronics", "clothing", "furniture", "books", "sports",
			"homegoods", "gardening", "toys", "jewelry", "art", "other",
		},
		AllowedConditions: []string{
			"new", "like_new", "good", "fair", "poor",
		},
	}
}

// ValidationResult contains the result of validation
type ValidationResult struct {
	Valid  bool
	Errors map[string]string
}

// NewValidationResult creates a new validation result
func NewValidationResult() *ValidationResult {
	return &ValidationResult{
		Valid:  true,
		Errors: make(map[string]string),
	}
}

// AddError adds an error to the validation result
func (vr *ValidationResult) AddError(field, message string) {
	vr.Valid = false
	vr.Errors[field] = message
}

// ValidateListing validates a listing
func (v *ListingValidator) ValidateListing(title, description, category, condition string, price int64) *ValidationResult {
	result := NewValidationResult()

	// Validate title
	if len(title) < v.TitleMinLength || len(title) > v.TitleMaxLength {
		result.AddError("title", fmt.Sprintf("Title must be between %d and %d characters", v.TitleMinLength, v.TitleMaxLength))
	}

	// Validate description
	if len(description) < v.DescriptionMinLength || len(description) > v.DescriptionMaxLength {
		result.AddError("description", fmt.Sprintf("Description must be between %d and %d characters", v.DescriptionMinLength, v.DescriptionMaxLength))
	}

	// Validate price
	if price < v.MinPrice || price > v.MaxPrice {
		result.AddError("price", fmt.Sprintf("Price must be between %d and %d", v.MinPrice, v.MaxPrice))
	}

	// Validate category
	categoryValid := false
	for _, validCategory := range v.AllowedCategories {
		if strings.EqualFold(category, validCategory) {
			categoryValid = true
			break
		}
	}
	if !categoryValid {
		result.AddError("category", "Invalid category")
	}

	// Validate condition
	conditionValid := false
	for _, validCondition := range v.AllowedConditions {
		if strings.EqualFold(condition, validCondition) {
			conditionValid = true
			break
		}
	}
	if !conditionValid {
		result.AddError("condition", "Invalid condition")
	}

	return result
}

// ValidateListing is a convenience function using the default validator
func ValidateListing(title, description, category, condition string, price int64) *ValidationResult {
	validator := NewListingValidator()
	return validator.ValidateListing(title, description, category, condition, price)
}
