package listings

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func PostListing(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()

	// Parse JSON payload
	var payload struct {
		Title         string                 `json:"title"`
		Description   string                 `json:"description"`
		Category      string                 `json:"category"`
		Condition     string                 `json:"condition"`
		Location      string                 `json:"location"`
		Price         int64                  `json:"price"`
		Negotiable    bool                   `json:"negotiable"`
		EcoScore      float32                `json:"ecoScore"`
		EcoAttributes []string               `json:"ecoAttributes"`
		ImageUrl      map[string]interface{} `json:"imageUrl"`
		Seller        db.Seller              `json:"seller"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Failed to parse JSON payload: " + err.Error())
	}

	// Extract fields from parsed JSON
	title := payload.Title
	description := payload.Description
	category := payload.Category
	condition := payload.Condition
	location := payload.Location
	price := payload.Price
	negotiable := payload.Negotiable
	ecoAttributes := payload.EcoAttributes
	ecoScore := payload.EcoScore

	// Handle imageUrl safely with proper type assertion
	var imageUrl []string
	if payload.ImageUrl != nil {
		if urls, exists := payload.ImageUrl["urls"]; exists && urls != nil {
			// Try to handle the case where urls is a []interface{}
			if urlsArray, ok := urls.([]interface{}); ok {
				for _, item := range urlsArray {
					if str, ok := item.(string); ok {
						imageUrl = append(imageUrl, str)
					}
				}
			}
		} else {
			// Try to handle the case where urls is a string
			if url, ok := urls.(string); ok {
				imageUrl = append(imageUrl, url)
			}
		}
	}

	seller := payload.Seller

	// Create the listing
	listing := db.Listing{
		Description:   lib.SanitizeInput(description),
		Category:      category,
		Condition:     condition,
		Price:         price,
		Location:      location,
		EcoScore:      ecoScore,
		EcoAttributes: ecoAttributes,
		Negotiable:    negotiable,
		Title:         lib.SanitizeInput(title),
		ImageUrl:      imageUrl,
		Seller:        seller,
	}

	// Post the listing to Supabase
	listingData, err := client.POST("listings", listing)
	if err != nil {
		return errors.DatabaseError("Failed to create listing: " + err.Error())
	}

	response := map[string]any{
		"listing": listingData,
	}

	return errors.SuccessResponse(c, response)
}
