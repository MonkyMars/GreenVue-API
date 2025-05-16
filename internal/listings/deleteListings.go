package listings

import (
	"greenvue/internal/db"
	"greenvue/lib/errors"
	"log"

	"github.com/gofiber/fiber/v2"
)

func DeleteListingById(c *fiber.Ctx) error {
	client := db.GetGlobalClient()

	// Extract listing ID from request path
	listingId := c.Params("listing_id")

	if client == nil {
		return errors.InternalServerError("Database connection failed")
	}

	// Delete listing using the standardized DELETE operation
	_, err := client.DELETE(c, "listings", "id=eq."+listingId)
	if err != nil {
		log.Println("Error deleting listing:", err)
		return errors.InternalServerError("Failed to delete listing: " + err.Error())
	}

	// Return successful response with no content
	return errors.SuccessResponse(c, nil)
}
