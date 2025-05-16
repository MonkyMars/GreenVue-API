package listings

import (
	"encoding/json"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func PostListing(c *fiber.Ctx) error {
	client := db.GetGlobalClient()

	if client == nil {
		return errors.InternalServerError("Failed to create client")
	}

	var payload struct {
		Title         string         `json:"title"`
		Description   string         `json:"description"`
		Category      string         `json:"category"`
		Condition     string         `json:"condition"`
		Price         float64        `json:"price"`
		Negotiable    bool           `json:"negotiable"`
		EcoScore      float32        `json:"ecoScore"`
		EcoAttributes []string       `json:"ecoAttributes"`
		ImageUrl      map[string]any `json:"imageUrl"`
		SellerID      string         `json:"seller_id"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Failed to parse JSON payload: " + err.Error())
	}

	// Parse image URLs (as before)
	var imageUrl []string
	if payload.ImageUrl != nil {
		if urls, exists := payload.ImageUrl["urls"]; exists {
			switch urlsTyped := urls.(type) {
			case []any:
				for _, item := range urlsTyped {
					if str, ok := item.(string); ok {
						imageUrl = append(imageUrl, str)
					}
				}
			case string:
				imageUrl = append(imageUrl, urlsTyped)
			}
		}
	}

	// Build the listing object using the lib.Listing type now
	listing := lib.Listing{
		Title:         lib.SanitizeInput(payload.Title),
		Description:   lib.SanitizeInput(payload.Description),
		Category:      payload.Category,
		Condition:     payload.Condition,
		Price:         payload.Price,
		Negotiable:    payload.Negotiable,
		EcoScore:      payload.EcoScore,
		EcoAttributes: payload.EcoAttributes,
		ImageUrl:      imageUrl,
		SellerID:      payload.SellerID,
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
