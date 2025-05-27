package validation

import (
	"fmt"
	"greenvue/lib"
	"strings"

	"github.com/google/uuid"
)

// ListingValidator provides validation for marketplace listings
type ListingValidator struct {
	TitleMinLength       int
	TitleMaxLength       int
	DescriptionMinLength int
	DescriptionMaxLength int
	MinPrice             float64
	MaxPrice             float64
	AllowedCategories    []string
	AllowedConditions    []string
	AllowedEcoAttributes []string
}

// NewListingValidator creates a validator with default settings
func NewListingValidator() *ListingValidator {
	return &ListingValidator{
		TitleMinLength:       5,
		TitleMaxLength:       40,
		DescriptionMinLength: 20,
		DescriptionMaxLength: 1000,
		MinPrice:             0,
		MaxPrice:             1000000,
		AllowedCategories:    lib.Categories,
		AllowedConditions:    lib.Conditions,
		AllowedEcoAttributes: lib.EcoAttributes,
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
func (v *ListingValidator) ValidateListing(listing lib.Listing) *ValidationResult {
	result := NewValidationResult()

	// Validate title
	if len(listing.Title) < v.TitleMinLength || len(listing.Title) > v.TitleMaxLength {
		result.AddError("title", fmt.Sprintf("Title must be between %d and %d characters", v.TitleMinLength, v.TitleMaxLength))
	}

	// Validate description
	if len(listing.Description) < v.DescriptionMinLength || len(listing.Description) > v.DescriptionMaxLength {
		result.AddError("description", fmt.Sprintf("Description must be between %d and %d characters", v.DescriptionMinLength, v.DescriptionMaxLength))
	}

	// Validate price
	if listing.Price < v.MinPrice || listing.Price > v.MaxPrice {
		result.AddError("price", fmt.Sprintf("Price must be between %f and %f", v.MinPrice, v.MaxPrice))
	}

	// Validate eco score
	if listing.EcoScore < 0 || listing.EcoScore >= 5 {
		result.AddError("eco_score", "Eco score must be between 0 and 5")
	}

	// Validate seller ID
	if listing.SellerID == uuid.Nil {
		result.AddError("seller_id", "Seller ID must be provided")
	}

	// Validate category
	categoryValid := false
	for _, validCategory := range v.AllowedCategories {
		if strings.EqualFold(listing.Category, validCategory) {
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
		if strings.EqualFold(listing.Condition, validCondition) {
			conditionValid = true
			break
		}
	}
	if !conditionValid {
		result.AddError("condition", "Invalid condition")
	}

	// Validate eco attributes
	if len(listing.EcoAttributes) > 0 {
		for _, attr := range listing.EcoAttributes {
			attrValid := false
			for _, validAttr := range v.AllowedEcoAttributes {
				if strings.EqualFold(attr, validAttr) {
					attrValid = true
					break
				}
			}
			if !attrValid {
				result.AddError("eco_attributes", fmt.Sprintf("Invalid eco attribute: %s", attr))
			}
		}
	} else {
		result.AddError("eco_attributes", "At least one eco attribute must be specified")
	}

	return result
}

// ValidateListing is a convenience function using the default validator
func ValidateListing(listing lib.Listing) *ValidationResult {
	validator := NewListingValidator()
	return validator.ValidateListing(listing)
}
