package reviews

import (
	"encoding/json"
	"fmt"
	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

const viewName string = "review_with_username"

func getQuery(selectedSeller, limit string) string {
	if selectedSeller == "" {
		return fmt.Sprintf("select=*&limit=%s", limit)
	} else {
		return fmt.Sprintf("select=*&limit=%s&seller_id=eq.%s", limit, selectedSeller)
	}
}

func GetReviews(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	if client == nil {
		return errors.InternalServerError("Failed to create client")
	}

	selectedSeller := c.Params("sellerID")

	limit := c.Query("limit", "50")

	query := getQuery(selectedSeller, limit)

	data, err := client.Query(viewName, query)

	if err != nil {
		return errors.DatabaseError("Failed to fetch reviews: " + err.Error())
	}

	if len(data) == 0 || string(data) == "[]" {
		return errors.SuccessResponse(c, []lib.FetchedReview{})
	}

	var reviews []lib.FetchedReview

	if err := json.Unmarshal(data, &reviews); err != nil {
		return errors.InternalServerError("Failed to parse reviews data: " + err.Error())
	}

	if reviews == nil {
		reviews = []lib.FetchedReview{}
	}

	return errors.SuccessResponse(c, reviews)
}
