package reviews

import (
	"fmt"
	"greenvue/internal/auth"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"
	"greenvue/lib/validation"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func PostReview(c *fiber.Ctx) error {
	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to create client")
	}

	claims, ok := c.Locals("user").(*auth.Claims)
	if !ok || claims == nil {
		return errors.Unauthorized("User not authenticated")
	}

	var payload struct {
		Rating           int       `json:"rating"`
		Title            string    `json:"title"`
		Content          string    `json:"content"`
		SellerID         uuid.UUID `json:"seller_id"`
		VerifiedPurchase bool      `json:"verified_purchase"`
	}
	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Invalid request body: " + err.Error())
	}

	review := lib.Review{
		Rating:           payload.Rating,
		Title:            lib.SanitizeInput(payload.Title),
		Content:          lib.SanitizeInput(payload.Content),
		SellerID:         payload.SellerID,
		VerifiedPurchase: payload.VerifiedPurchase,
		UserID:           claims.UserId,
	}
	// Validate the review using the validation package
	validationResult := validation.ValidateReview(review)
	if !validationResult.Valid {
		// Return the first validation error
		for _, message := range validationResult.Errors {
			return errors.BadRequest(message)
		}
	}

	// Check if the user has already reviewed this seller
	query := fmt.Sprintf("reviews?user_id=%s&seller_id=%s", claims.UserId.String(), payload.SellerID.String())
	existingReview, err := client.GET("reviews", query)
	if err != nil {
		return errors.DatabaseError("Failed to check existing reviews: " + err.Error())
	}

	var reviews []lib.Review
	if err := json.Unmarshal(existingReview, &reviews); err != nil {
		return errors.InternalServerError("Failed to parse existing review data: " + err.Error())
	}

	// If the user has already reviewed this seller, return an error
	if len(reviews) > 0 {
		return errors.AlreadyExists("You have already reviewed this seller")
	}

	// Use standardized POST operation
	data, err := client.POST("reviews", review)
	if err != nil {
		return errors.DatabaseError("Failed to post review: " + err.Error())
	}

	if len(data) == 0 || string(data) == "[]" {
		return errors.InternalServerError("Failed to create review")
	}

	// Parse the response
	var createdReview lib.Review
	if err := json.Unmarshal(data, &createdReview); err != nil {
		// If the response is an array, try parsing it as an array
		var reviewArray []lib.Review
		if err := json.Unmarshal(data, &reviewArray); err != nil {
			return errors.InternalServerError("Failed to parse review response: " + err.Error())
		}

		if len(reviewArray) == 0 {
			return errors.InternalServerError("Empty review response")
		}

		createdReview = reviewArray[0]
	}

	return errors.SuccessResponse(c, createdReview)
}
