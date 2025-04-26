package seller

import (
	"encoding/json"
	"fmt"
	"greentrade-eu/internal/db"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func GetSellers(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	data, err := client.Query("users", "select=*")

	if err != nil {
		return errors.InternalServerError("Failed to fetch sellers: " + err.Error())
	}

	var sellers []db.User
	if err := json.Unmarshal(data, &sellers); err != nil {
		return errors.BadRequest("Failed to parse sellers: " + err.Error())
	}

	if len(sellers) == 0 {
		return errors.SuccessResponse(c, []db.User{})
	}

	return errors.SuccessResponse(c, sellers)
}

func GetSellerById(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	sellerID := c.Params("id")
	query := fmt.Sprintf("select=id,created_at,name,location,bio,rating,verified&"+
		"id=eq.%s", sellerID)

	data, err := client.Query("users", query)

	if err != nil {
		return errors.InternalServerError("Failed to fetch seller: " + err.Error())
	}

	var sellers []db.User
	if err := json.Unmarshal(data, &sellers); err != nil {
		return errors.BadRequest("Failed to parse seller: " + err.Error())
	}

	if len(sellers) == 0 {
		return errors.SuccessResponse(c, db.User{})
	}

	return errors.SuccessResponse(c, sellers[0])
}
