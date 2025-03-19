package listings

import (
	"encoding/json"
	"greentrade-eu/internal/db"

	"github.com/gofiber/fiber/v2"
)

func GetListings(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	data, err := client.Query("listings", "select=*")

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch listings: " + err.Error(),
		})
	}

	var listings []db.Listing
	if err := json.Unmarshal(data, &listings); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse listings: " + err.Error(),
		})
	}

	return c.JSON(listings)
}
