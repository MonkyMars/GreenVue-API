package listings

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func PostListing(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()

	if client == nil {
		return errors.InternalServerError("Failed to create client")
	}

	var payload struct {
		Title         string         `json:"title"`
		Description   string         `json:"description"`
		Category      string         `json:"category"`
		Condition     string         `json:"condition"`
		Location      string         `json:"location"`
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

	// Build the listing object
	listing := db.Listing{
		Title:         lib.SanitizeInput(payload.Title),
		Description:   lib.SanitizeInput(payload.Description),
		Category:      payload.Category,
		Condition:     payload.Condition,
		Location:      payload.Location,
		Price:         payload.Price,
		Negotiable:    payload.Negotiable,
		EcoScore:      payload.EcoScore,
		EcoAttributes: payload.EcoAttributes,
		ImageUrl:      imageUrl,
		SellerID:      payload.SellerID,
	}

	// Insert into Supabase
	listingData, err := client.POST("listings", listing)
	if err != nil {
		return errors.DatabaseError("Failed to create listing: " + err.Error())
	}

	response := map[string]any{
		"listing": listingData,
	}

	return errors.SuccessResponse(c, response)
}
