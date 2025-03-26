package auth

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func GetUserById(c *fiber.Ctx) error {
	userId := c.Params("id")

	if userId == "" {
		return errors.BadRequest("User ID is required")
	}

	client := db.NewSupabaseClient()

	// Get user by ID
	user, err := client.GetUserById(userId)
	if err != nil {
		return errors.DatabaseError("Failed to fetch user: " + err.Error())
	}

	return errors.SuccessResponse(c, fiber.Map{
		"user": user,
	})
}

func GetUserByAccessToken(c *fiber.Ctx) error {
	accessToken := c.Get("Authorization")

	if accessToken == "" {
		errors.Unauthorized("Access token is required")
		return nil
	}

	client := db.NewSupabaseClient()

	// Get user by access token
	user, err := client.GetUserByAccessToken(accessToken)
	if err != nil {
		errors.Unauthorized("Invalid access token")
		return nil
	}

	errors.SuccessResponse(c, fiber.Map{
		"user": user,
	})

	return nil
}
