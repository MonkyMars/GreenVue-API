package listings

import (
	"encoding/json"
	"fmt"
	"greentrade-eu/internal/db"

	"greentrade-eu/lib/errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const viewName string = "listings_with_seller"

func GetListings(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}

	limit := c.Query("limit", "50")

	data, err := client.Query(viewName, "select=*&limit="+limit)

	if err != nil {
		return errors.DatabaseError("Failed to fetch listings: " + err.Error())
	}
	if len(data) == 0 || string(data) == "[]" {
		return errors.NotFound("No listings found")
	}

	var listings []db.FetchedListing
	if err := json.Unmarshal(data, &listings); err != nil {
		return errors.InternalServerError("Failed to parse listings data")
	}

	if listings == nil {
		listings = []db.FetchedListing{}
	}

	return errors.SuccessResponse(c, listings)
}

func GetListingById(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	listingID := c.Params("id")

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}
	if listingID == "" {
		return errors.BadRequest("Listing ID is required")
	}

	query := fmt.Sprintf("select=*&id=eq.%s", listingID)

	data, err := client.Query(viewName, query)

	if err != nil {
		if strings.Contains(err.Error(), "invalid input syntax") {
			return errors.BadRequest("Invalid listing ID format")
		}
		return errors.DatabaseError("Failed to fetch listing: " + err.Error())
	}

	if len(data) == 0 || string(data) == "[]" {
		return errors.NotFound("Listing not found")
	}

	var listings []db.FetchedListing
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

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}
	if category == "" {
		return errors.BadRequest("Category is required")
	}

	data, err := client.Query(viewName, "select=*&category=eq."+category)
	if err != nil {
		return errors.DatabaseError("Failed to fetch listings by category: " + err.Error())
	}
	if len(data) == 0 {
		return errors.NotFound("No listings found in this category")
	}

	var listings []db.FetchedListing
	if err := json.Unmarshal(data, &listings); err != nil {
		return errors.InternalServerError("Failed to parse listings data")
	}

	return errors.SuccessResponse(c, listings)
}

func GetListingBySeller(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	sellerID := c.Params("sellerId")

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}
	if sellerID == "" {
		return errors.BadRequest("Seller ID is required")
	}

	query := fmt.Sprintf("select=*&seller_id=eq.%s", sellerID)
	data, err := client.Query(viewName, query)
	if err != nil {
		fmt.Printf("Error from Supabase: %v\n", err)
		return errors.DatabaseError("Failed to fetch listings by seller: " + err.Error())
	}
	if len(data) == 0 || string(data) == "[]" {
		return errors.NotFound("No listings found for this seller")
	}

	var listings []db.FetchedListing
	if err := json.Unmarshal(data, &listings); err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		return errors.InternalServerError("Failed to parse listings data")
	}
	if listings == nil {
		listings = []db.FetchedListing{}
	}

	return errors.SuccessResponse(c, listings)
}
