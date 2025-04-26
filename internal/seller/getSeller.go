package seller

import (
	"encoding/json"
	"fmt"
	"net/url"

	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func GetSeller(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()
	sellerID := c.Params("user_id")
	query := fmt.Sprintf("select=id,created_at,name,location,bio,rating,verified&"+
		"id=eq.%s", url.QueryEscape(sellerID))

	data, err := client.GET("users", query)

	if err != nil {
		return errors.InternalServerError("Failed to fetch seller: " + err.Error())
	}

	var sellers []lib.PublicUser
	if err := json.Unmarshal(data, &sellers); err != nil {
		return errors.BadRequest("Failed to parse seller: " + err.Error())
	}

	if len(sellers) == 0 {
		return errors.SuccessResponse(c, lib.PublicUser{})
	}

	return errors.SuccessResponse(c, sellers[0])
}
