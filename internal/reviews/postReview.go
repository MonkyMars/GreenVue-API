package reviews

import (
	"fmt"
	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
	"greentrade-eu/lib/errors"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

func PostReview(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	if client == nil {
		return errors.InternalServerError("Failed to create client")
	}

	var review lib.Review
	if err := c.BodyParser(&review); err != nil {
		return errors.BadRequest("Invalid request body: " + err.Error())
	}

	if review.UserID == "" || review.SellerID == "" {
		return errors.BadRequest("UserID, SellerID, and ListingID are required")
	}

	if review.Rating < 1 || review.Rating > 5 {
		return errors.BadRequest("Rating must be between 1 and 5")
	}

	reviewJSON, err := json.Marshal(review)
	if err != nil {
		return errors.InternalServerError("Failed to marshal review: " + err.Error())
	}

	fmt.Println("Review JSON:", string(reviewJSON))

	data, err := client.PostRaw("reviews", reviewJSON)

	if err != nil {
		return errors.DatabaseError("Failed to post review: " + err.Error())
	}

	if len(data) == 0 || string(data) == "[]" {
		return errors.InternalServerError("Failed to create review")
	}

	return errors.SuccessResponse(c, data)
}
