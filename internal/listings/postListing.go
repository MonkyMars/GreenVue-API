package listings

import (
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

func PostListing(c *fiber.Ctx) error {
	client := db.GetGlobalClient()

	if client == nil {
		return errors.InternalServerError("Failed to create client")
	}

	var listing lib.Listing
	if err := c.BodyParser(&listing); err != nil {
		return errors.BadRequest("Failed to parse JSON payload: " + err.Error())
	}

	// Build the listing object using the lib.Listing type now

	listing = lib.Listing{
		Title:         lib.SanitizeInput(listing.Title),
		Description:   lib.SanitizeInput(listing.Description),
		Category:      listing.Category,
		Condition:     listing.Condition,
		Price:         listing.Price,
		Negotiable:    listing.Negotiable,
		EcoScore:      listing.EcoScore,
		EcoAttributes: listing.EcoAttributes,
		ImageUrl:      listing.ImageUrl,
		SellerID:      listing.SellerID,
	}

	// Insert into Supabase using standardized repository method
	listingData, err := client.POST("listings", listing)
	if err != nil {
		return errors.DatabaseError("Failed to create listing: " + err.Error())
	}

	// Parse the response - first try as an array since Supabase returns arrays
	var listingArray []lib.Listing
	if err := json.Unmarshal(listingData, &listingArray); err != nil {
		// If that fails, try as a single object
		var createdListing lib.Listing
		if err := json.Unmarshal(listingData, &createdListing); err != nil {
			return errors.InternalServerError("Failed to parse created listing: " + err.Error())
		}
		return errors.SuccessResponse(c, fiber.Map{
			"listing": createdListing,
		})
	}

	// If we got here, we parsed an array successfully
	if len(listingArray) == 0 {
		return errors.InternalServerError("No listing was created")
	}

	// Return the first (and likely only) listing in the array
	return errors.SuccessResponse(c, fiber.Map{
		"listing": listingArray[0],
	})
}
