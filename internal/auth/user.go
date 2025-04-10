package auth

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
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

func UpdateUser(c *fiber.Ctx) error {
	userId := c.Params("id")

	if userId == "" {
		return errors.BadRequest("User ID is required")
	}

	// Parse request body into a user object
	var userUpdate lib.UpdateUser
	if err := c.BodyParser(&userUpdate); err != nil {
		return errors.BadRequest("Invalid user data format")
	}
	fmt.Println(userUpdate)
	// Validate the user data
	if err := errors.ValidateRequest(c, &userUpdate); err != nil {
		return errors.BadRequest(err.Error())
	}

	client := db.NewSupabaseClient()

	if client == nil {
		return errors.InternalServerError("Failed to create Supabase client")
	}

	// Make sure the ID is set correctly
	userUpdate.ID = userId

	// Update user (safely handle errors)
	updatedUser, err := client.UpdateUser(&userUpdate)
	if err != nil {
		// Check for specific "not found" error message
		if err.Error() == fmt.Sprintf("user not found with ID: %s", userId) {
			return errors.NotFound("User not found")
		}
		return errors.InternalServerError("Failed to update user")
	}

	// Check if updatedUser is nil first, then check ID
	if updatedUser == nil || updatedUser.ID == "" {
		return errors.NotFound("User not found")
	}

	return errors.SuccessResponse(c, updatedUser)
}
