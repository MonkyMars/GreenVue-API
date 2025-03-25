package listings

import (
	"fmt"
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

func GetListingById(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	listingID := c.Params("id")
	query := fmt.Sprintf("select=*&id=eq.%s", listingID)
	data, err := client.Query("listings", query)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch listing: " + err.Error(),
		})
	}

	var listings []db.Listing
	if err := json.Unmarshal(data, &listings); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse listing: " + err.Error(),
		})
	}

	if len(listings) == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "Listing not found",
		})
	}

	return c.JSON(listings[0])
}

func GetListingByCategory(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	category := c.Params("category")
	data, err := client.Query("listings", "select=*&category=eq."+category)

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch listings: " + err.Error(),
		})
	}

	if len(data) == 0 {
		return c.Status(404).JSON(fiber.Map{
			"error": "No listings found",
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
