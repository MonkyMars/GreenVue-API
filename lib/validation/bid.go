package validation

import (
	"fmt"
	"greenvue/lib"

	"github.com/google/uuid"
)

// BidValidator provides validation for bids
type BidValidator struct {
	MinBidAmount    float64
	MaxBidAmount    float64
	MinBidIncrement float64
}

// NewBidValidator creates a validator with default settings
func NewBidValidator() *BidValidator {
	return &BidValidator{
		MinBidAmount:    0.01,    // Minimum bid of 1 cent
		MaxBidAmount:    1000000, // Maximum bid of 1 million
		MinBidIncrement: 0.01,    // Minimum increment of 1 cent
	}
}

// BidValidationContext contains context needed for bid validation
type BidValidationContext struct {
	Listing      *lib.FetchedListing
	HighestBid   *lib.FetchedBid
	ExistingBids []lib.FetchedBid
	BidderID     uuid.UUID
}

// ValidateBid validates a bid with context
func (v *BidValidator) ValidateBid(bid lib.Bid, context *BidValidationContext) *ValidationResult {
	result := NewValidationResult()

	// Validate required UUIDs
	if bid.UserID == uuid.Nil {
		result.AddError("user_id", "User ID is required")
	}
	if bid.ListingID == uuid.Nil {
		result.AddError("listing_id", "Listing ID is required")
	}

	// Validate bid amount range
	if bid.Price < v.MinBidAmount {
		result.AddError("price", fmt.Sprintf("Bid amount must be at least %.2f", v.MinBidAmount))
	}
	if bid.Price > v.MaxBidAmount {
		result.AddError("price", fmt.Sprintf("Bid amount cannot exceed %.2f", v.MaxBidAmount))
	}

	// Context-based validations (if context is provided)
	if context != nil {
		// Validate against listing
		if context.Listing != nil {
			// Check if listing allows bidding (negotiable)
			if !context.Listing.Negotiable {
				result.AddError("listing", "This listing does not accept bids")
			}

			// Prevent self-bidding
			if context.Listing.SellerID == bid.UserID {
				result.AddError("user_id", "Cannot bid on your own listing")
			}

			// Validate bid against listing price
			if bid.Price > context.Listing.Price {
				result.AddError("price", fmt.Sprintf("Bid amount cannot exceed the listing price of %.2f", context.Listing.Price))
			}

			// Minimum bid should be reasonable percentage of listing price
			minReasonableBid := context.Listing.Price * 0.1 // 10% of listing price
			if bid.Price < minReasonableBid {
				result.AddError("price", fmt.Sprintf("Bid amount should be at least %.2f (10%% of listing price)", minReasonableBid))
			}
		}

		// Validate against existing bids
		if context.HighestBid != nil {
			minNextBid := context.HighestBid.Price + v.MinBidIncrement
			if bid.Price <= context.HighestBid.Price {
				result.AddError("price", fmt.Sprintf("Bid must be higher than current highest bid of %.2f", context.HighestBid.Price))
			}
			if bid.Price < minNextBid {
				result.AddError("price", fmt.Sprintf("Bid must be at least %.2f (minimum increment of %.2f)", minNextBid, v.MinBidIncrement))
			}
		}

		// Check for duplicate bids from same user
		for _, existingBid := range context.ExistingBids {
			if existingBid.UserID == bid.UserID && existingBid.Price == bid.Price {
				result.AddError("price", "You have already placed this exact bid amount")
			}
		}
	}

	return result
}

// ValidateBidIncrement checks if a bid meets increment requirements
func (v *BidValidator) ValidateBidIncrement(newBid, currentHighest float64) bool {
	return newBid >= currentHighest+v.MinBidIncrement
}

// CalculateMinimumBid calculates the minimum valid bid for a listing
func (v *BidValidator) CalculateMinimumBid(listing *lib.FetchedListing, highestBid *lib.FetchedBid) float64 {
	if highestBid != nil {
		return highestBid.Price + v.MinBidIncrement
	}

	// If no bids exist, minimum is 10% of listing price or minimum bid amount, whichever is higher
	if listing != nil {
		minFromPrice := listing.Price * 0.1
		if minFromPrice > v.MinBidAmount {
			return minFromPrice
		}
	}

	return v.MinBidAmount
}

// ValidateBid is a convenience function using the default validator
func ValidateBid(bid lib.Bid, context *BidValidationContext) *ValidationResult {
	validator := NewBidValidator()
	return validator.ValidateBid(bid, context)
}
