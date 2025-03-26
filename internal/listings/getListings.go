package listings

import (
	"encoding/json"
	"fmt"
	"greentrade-eu/internal/db"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func GetListings(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	data, err := client.Query("listings", "select=*")

	if err != nil {
		return errors.DatabaseError("Failed to fetch listings: " + err.Error())
	}

	var listings []db.Listing
	if err := json.Unmarshal(data, &listings); err != nil {
		return errors.InternalServerError("Failed to parse listings data")
	}

	return errors.SuccessResponse(c, listings)
}

func GetListingById(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	listingID := c.Params("id")

	if listingID == "" {
		return errors.BadRequest("Listing ID is required")
	}

	query := fmt.Sprintf("select=*&id=eq.%s", listingID)
	data, err := client.Query("listings", query)

	if err != nil {
		return errors.DatabaseError("Failed to fetch listing: " + err.Error())
	}

	var listings []db.Listing
	if err := json.Unmarshal(data, &listings); err != nil {
		return errors.InternalServerError("Failed to parse listing data")
	}

	if len(listings) == 0 {
		return errors.NotFound("Listing not found")
	}

	return errors.SuccessResponse(c, listings[0])
}

func GetListingByCategory(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	category := c.Params("category")

	if category == "" {
		return errors.BadRequest("Category is required")
	}

	data, err := client.Query("listings", "select=*&category=eq."+category)

	if err != nil {
		return errors.DatabaseError("Failed to fetch listings by category: " + err.Error())
	}

	if len(data) == 0 {
		return errors.NotFound("No listings found in this category")
	}

	var listings []db.Listing
	if err := json.Unmarshal(data, &listings); err != nil {
		return errors.InternalServerError("Failed to parse listings data")
	}

	return errors.SuccessResponse(c, listings)
}
