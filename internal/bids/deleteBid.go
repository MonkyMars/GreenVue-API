package bids

import (
	"encoding/json"
	"fmt"
	"greenvue/internal/auth"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// DeleteBid handles the deletion of a bid with proper authorization
func DeleteBid(c *fiber.Ctx) error {
	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to get database client")
	}

	// Get bid ID from parameters
	bidID := c.Params("bid_id")
	if bidID == "" {
		return errors.BadRequest("Bid ID is required")
	}

	// Validate bid ID format
	if _, err := uuid.Parse(bidID); err != nil {
		return errors.BadRequest("Invalid bid ID format")
	}

	// Get authenticated user from JWT middleware
	claims, ok := c.Locals("user").(*auth.Claims)
	if !ok {
		return errors.Unauthorized("User authentication required")
	}

	// First, check if the bid exists and get bid details
	query := fmt.Sprintf("id=eq.%s", bidID)
	bidData, err := client.GET("bids", query)
	if err != nil {
		return errors.InternalServerError("Failed to retrieve bid: " + err.Error())
	}

	var bids []lib.Bid
	if err := json.Unmarshal(bidData, &bids); err != nil {
		return errors.InternalServerError("Failed to unmarshal bid data: " + err.Error())
	}

	if len(bids) == 0 {
		return errors.NotFound("Bid not found")
	}

	bid := bids[0]

	// Check if the authenticated user owns this bid
	if bid.UserID != claims.UserId {
		return errors.Forbidden("You can only delete your own bids")
	}

	// Check if this is the highest bid - optionally prevent deletion of highest bids
	// Get all bids for this listing to check if this is the highest
	listingQuery := fmt.Sprintf("listing_id=eq.%s&order=price.desc,created_at.desc", bid.ListingID.String())
	allBidsData, err := client.GET("bids", listingQuery)
	if err != nil {
		return errors.InternalServerError("Failed to retrieve listing bids: " + err.Error())
	}

	var allBids []lib.Bid
	if err := json.Unmarshal(allBidsData, &allBids); err != nil {
		return errors.InternalServerError("Failed to unmarshal listing bids: " + err.Error())
	}

	// Delete the bid
	deleteQuery := fmt.Sprintf("id=eq.%s", bidID)
	_, err = client.DELETE("bids", deleteQuery)
	if err != nil {
		return errors.InternalServerError("Failed to delete bid: " + err.Error())
	}

	// Return success response with deleted bid information
	response := map[string]any{
		"message":    "Bid deleted successfully",
		"bid_id":     bidID,
		"listing_id": bid.ListingID,
		"amount":     bid.Price,
	}

	return errors.SuccessResponse(c, response)
}
