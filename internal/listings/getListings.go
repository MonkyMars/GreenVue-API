package listings

import (
	"encoding/json"
	"fmt"
	"greenvue-eu/internal/db"
	"greenvue-eu/lib"
	"greenvue-eu/lib/errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

const viewName string = "listings_with_seller"

func GetListings(c *fiber.Ctx) error {
	client := db.GetGlobalClient()

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}

	limit := c.Query("limit", "50")

	query := fmt.Sprintf("select=*&limit=%s&order=created_at.desc", limit)

	data, err := client.GET(viewName, query)

	if err != nil {
		return errors.DatabaseError("Failed to fetch listings: " + err.Error())
	}
	if len(data) == 0 || string(data) == "[]" {
		return errors.SuccessResponse(c, []lib.FetchedListing{})
	}

	var listings []lib.FetchedListing
	if err := json.Unmarshal(data, &listings); err != nil {
		return errors.InternalServerError("Failed to parse listings data")
	}

	if listings == nil {
		listings = []lib.FetchedListing{}
	}

	return errors.SuccessResponse(c, listings)
}

func GetListingById(c *fiber.Ctx) error {
	client := db.GetGlobalClient()
	listingID := c.Params("listing_id")

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}
	if listingID == "" {
		return errors.BadRequest("Listing ID is required")
	}

	query := fmt.Sprintf("select=*&id=eq.%s", listingID)

	data, err := client.GET(viewName, query)

	if err != nil {
		if strings.Contains(err.Error(), "invalid input syntax") {
			return errors.BadRequest("Invalid listing ID format")
		}
		return errors.DatabaseError("Failed to fetch listing: " + err.Error())
	}

	if len(data) == 0 || string(data) == "[]" {
		return errors.SuccessResponse(c, lib.FetchedListing{}) // Return nil if no listing found
	}

	var listings []lib.FetchedListing
	if err := json.Unmarshal(data, &listings); err != nil {
		return errors.InternalServerError("Failed to parse listing data")
	}

	if len(listings) == 0 {
		return errors.SuccessResponse(c, lib.FetchedListing{})
	}

	return errors.SuccessResponse(c, listings[0])
}

func GetListingByCategory(c *fiber.Ctx) error {
	client := db.GetGlobalClient()
	category := c.Params("category")

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}
	if category == "" {
		return errors.BadRequest("Category is required")
	}

	data, err := client.GET(viewName, "select=*&category=eq."+category)
	if err != nil {
		return errors.DatabaseError("Failed to fetch listings by category: " + err.Error())
	}
	if len(data) == 0 {
		return errors.SuccessResponse(c, []lib.FetchedListing{})
	}

	var listings []lib.FetchedListing
	if err := json.Unmarshal(data, &listings); err != nil {
		return errors.InternalServerError("Failed to parse listings data")
	}

	return errors.SuccessResponse(c, listings)
}

func GetListingBySeller(c *fiber.Ctx) error {
	client := db.GetGlobalClient()
	sellerID := c.Params("seller_id")

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}
	if sellerID == "" {
		return errors.BadRequest("Seller ID is required")
	}

	query := fmt.Sprintf("select=*&seller_id=eq.%s", sellerID)
	data, err := client.GET(viewName, query)
	if err != nil {
		fmt.Printf("Error from Supabase: %v\n", err)
		return errors.DatabaseError("Failed to fetch listings by seller: " + err.Error())
	}
	if len(data) == 0 || string(data) == "[]" {
		return errors.SuccessResponse(c, []lib.FetchedListing{})
	}

	var listings []lib.FetchedListing
	if err := json.Unmarshal(data, &listings); err != nil {
		fmt.Printf("JSON unmarshal error: %v\n", err)
		return errors.InternalServerError("Failed to parse listings data")
	}

	if listings == nil {
		listings = []lib.FetchedListing{}
	}

	return errors.SuccessResponse(c, listings)
}
