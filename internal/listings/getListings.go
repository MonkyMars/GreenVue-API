package listings

import (
	"encoding/json"
	"fmt"
	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
	"greentrade-eu/lib/errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

func GetListings(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	data, err := client.Query("listings", "select=*")

	if err != nil {
		return errors.DatabaseError("Failed to fetch listings: " + err.Error())
	}

	if data == nil {
		return errors.NotFound("No listings found")
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

	if !lib.IsNumeric(listingID) {
		return errors.BadRequest("Invalid listing ID format - must be a number")
	}

	query := fmt.Sprintf("select=*&id=eq.%s", listingID)
	data, err := client.Query("listings", query)

	if err != nil {
		if strings.Contains(err.Error(), "invalid input syntax") {
			return errors.BadRequest("Invalid listing ID format")
		}
		return errors.DatabaseError("Failed to fetch listing: " + err.Error())
	}

	if len(data) == 0 || string(data) == "[]" {
		return errors.NotFound("Listing not found")
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
