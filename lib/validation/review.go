package validation

import (
	"fmt"
	"greenvue/lib"

	"github.com/google/uuid"
)

// ReviewValidator provides validation for reviews
type ReviewValidator struct {
	MinRating            int
	MaxRating            int
	TitleMinLength       int
	TitleMaxLength       int
	DescriptionMinLength int
	DescriptionMaxLength int
}

// NewReviewValidator creates a validator with default settings
func NewReviewValidator() *ReviewValidator {
	return &ReviewValidator{
		MinRating:            1,
		MaxRating:            5,
		TitleMinLength:       5,
		TitleMaxLength:       100,
		DescriptionMinLength: 20,
		DescriptionMaxLength: 1000,
	}
}

// ValidateReview validates a review
func (v *ReviewValidator) ValidateReview(review lib.Review) *ValidationResult {
	result := NewValidationResult()

	// Validate rating (most critical field)
	if review.Rating < v.MinRating || review.Rating > v.MaxRating {
		result.AddError("rating", fmt.Sprintf("Rating must be between %d and %d", v.MinRating, v.MaxRating))
	}

	// Validate title
	if len(review.Title) < v.TitleMinLength || len(review.Title) > v.TitleMaxLength {
		result.AddError("title", fmt.Sprintf("Title must be between %d and %d characters", v.TitleMinLength, v.TitleMaxLength))
	}

	// Validate content
	if len(review.Content) < v.DescriptionMinLength || len(review.Content) > v.DescriptionMaxLength {
		result.AddError("content", fmt.Sprintf("Content must be between %d and %d characters", v.DescriptionMinLength, v.DescriptionMaxLength))
	}

	// Validate required UUIDs
	if review.UserID == uuid.Nil {
		result.AddError("user_id", "UserID is required")
	}
	if review.SellerID == uuid.Nil {
		result.AddError("seller_id", "SellerID is required")
	}

	// Validate that user is not reviewing themselves
	if review.UserID == review.SellerID {
		result.AddError("seller_id", "Users cannot review themselves")
	}

	return result
}

// ValidateReview is a convenience function using the default validator
func ValidateReview(review lib.Review) *ValidationResult {
	validator := NewReviewValidator()
	return validator.ValidateReview(review)
}
