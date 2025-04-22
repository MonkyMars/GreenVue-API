package listings

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib/errors"
	"log"

	"github.com/gofiber/fiber/v2"
)

func DeleteListingById(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()

	// Extract listing ID from request path
	listingId := c.Params("id")

	// Delete listing from database
	_, err := client.DELETE("listings", listingId)
	if err != nil {
		log.Println(err)
		return errors.InternalServerError("Failed to delete listing: " + err.Error())
	}

	// Return 204 No Content
	return errors.SuccessResponse(c, nil)
}
