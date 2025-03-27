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

	if user.ID == "" {
		return errors.NotFound("User not found")
	}

	return errors.SuccessResponse(c, fiber.Map{
		"user": user,
	})
}

func GetUserByAccessToken(c *fiber.Ctx) error {
	accessToken := c.Get("Authorization")

	if accessToken == "" {
		return errors.Unauthorized("Access token is required") // RETURN the error
	}

	client := db.NewSupabaseClient()

	// Get user by access token
	user, err := client.GetUserByAccessToken(accessToken)
	if err != nil {
		return errors.Unauthorized("Invalid access token") // RETURN the error
	}

	// Check if the user exists using zero value comparison or a method provided by the User type
	if user.ID == "" {
		return errors.NotFound("User not found") // RETURN the error
	}

	// Return the success response
	return errors.SuccessResponse(c, fiber.Map{
		"user": user,
	})
}

func RefreshAccessToken(c *fiber.Ctx) error {
	refreshToken := c.Get("Authorization")

	if refreshToken == "" {
		return errors.Unauthorized("Refresh token is required") // RETURN the error
	}

	client := db.NewSupabaseClient()

	// Refresh access token
	user, err := client.RefreshAccessToken(refreshToken)
	if err != nil {
		return errors.Unauthorized("Invalid refresh token") // RETURN the error
	}

	if user.ID == "" {
		return errors.NotFound("User not found") // RETURN the error
	}

	// Return user and success message
	return errors.SuccessResponse(c, fiber.Map{ // RETURN the response
		"user":    user,
		"success": true,
		"message": "Access token refreshed successfully",
	})
}
