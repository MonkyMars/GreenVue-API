package bids

import (
	"greenvue/internal/auth"
	"greenvue/lib"
	"greenvue/lib/errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// UploadBid handles the creation of a new bid with comprehensive validation
func UploadBid(c *fiber.Ctx) error {

	claims, ok := c.Locals("user").(*auth.Claims)
	if !ok || claims == nil {
		return errors.Unauthorized("User not authenticated")
	}

	bidService := NewBidService()
	if bidService.client == nil {
		return errors.InternalServerError("Failed to get database client")
	}

	// Get listing ID from URL parameter
	listingID := c.Params("listing_id")
	if listingID == "" {
		return errors.BadRequest("Listing ID is required")
	}

	// Validate listing ID format
	listingUUID, err := uuid.Parse(listingID)
	if err != nil {
		return errors.BadRequest("Invalid listing ID format")
	}

	var payload struct {
		Price     float64 `json:"price"`
		ListingID string  `json:"listing_id"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Failed to parse bid data: " + err.Error())
	}

	if payload.ListingID != listingID {
		return errors.BadRequest("Listing ID in payload does not match URL parameter")
	}

	if payload.Price <= 0 {
		return errors.BadRequest("Bid price must be greater than zero")
	}

	// Create the bid object
	bid := lib.Bid{
		Price:     payload.Price,
		UserID:    claims.UserId,
		ListingID: listingUUID,
	}

	// Set the listing ID from URL parameter
	bid.ListingID = listingUUID

	// Place the bid using the service
	createdBid, err := bidService.PlaceBid(bid)
	if err != nil {
		log.Printf("Failed to place bid: %v", err)

		// Check if it's a validation error and return appropriate response
		if strings.Contains(err.Error(), "validation failed") {
			return errors.BadRequest("Bid validation failed: " + err.Error())
		}
		if strings.Contains(err.Error(), "not found") {
			return errors.NotFound("Listing not found")
		}
		if strings.Contains(err.Error(), "not accept bids") {
			return errors.BadRequest("This listing does not accept bids")
		}
		if strings.Contains(err.Error(), "your own listing") {
			return errors.BadRequest("Cannot bid on your own listing")
		}
		if strings.Contains(err.Error(), "higher than current") {
			return errors.BadRequest("Bid must be higher than current highest bid")
		}

		return errors.InternalServerError("Failed to place bid: " + err.Error())
	}

	return errors.SuccessResponse(c, createdBid)
}
