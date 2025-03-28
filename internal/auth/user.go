package auth

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"

	"fmt"
)

func GetUserById(c *fiber.Ctx) error {
	userId := c.Params("id")

	if userId == "" {
		return errors.BadRequest("User ID is required")
	}

	client := db.NewSupabaseClient()

	// Get user by ID (safely handle errors)
	user, err := client.GetUserById(userId)
	if err != nil {
		// Check for specific "not found" error message
		if err.Error() == fmt.Sprintf("user not found with ID: %s", userId) {
			return errors.NotFound("User not found")
		}
		return errors.InternalServerError("Failed to fetch user")
	}

	// if user.ID == "" {
	// 	return errors.NotFound("User not found")
	// }

	return errors.SuccessResponse(c, fiber.Map{
		"user": user,
	})
}

func GetUserByAccessToken(c *fiber.Ctx) error {
	// Get claims from context (set by AuthMiddleware)
	claims, ok := c.Locals("user").(*Claims)
	if !ok {
		return errors.Unauthorized("Invalid token claims")
	}

	client := db.NewSupabaseClient()

	// Get user by ID from claims
	user, err := client.GetUserById(claims.UserID)
	if err != nil {
		if err.Error() == fmt.Sprintf("user not found with ID: %s", claims.UserID) {
			return errors.NotFound("User not found")
		}
		return errors.InternalServerError("Failed to fetch user")
	}

	if user.ID == "" {
		return errors.NotFound("User not found")
	}

	return errors.SuccessResponse(c, fiber.Map{
		"user": user,
	})
}
