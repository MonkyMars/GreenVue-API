package listings

import (
	"greentrade-eu/internal/db"
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
		return c.Status(500).SendString("Failed to delete listing")
	}

	// Return 204 No Content
	return c.Status(204).Send(nil)
}
