package auth

import (
	"encoding/json"
	"fmt"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"

	"github.com/gofiber/fiber/v2"
)

// DeleteAccount handles the deletion of a user account
func DeleteAccount(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(*Claims)
	if !ok || claims == nil {
		return errors.Unauthorized("Unauthorized")
	}

	userID := claims.UserId

	// Get the client
	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to get database client")
	}

	// Check if the user exists
	query := fmt.Sprintf("id=eq.%s", userID)
	data, err := client.GET("users", query)
	if err != nil {
		return errors.InternalServerError("Failed to check user existence: " + err.Error())
	}

	if len(data) == 0 {
		return errors.NotFound("User not found")
	}

	// Unmarshal the user data
	var users []lib.User
	if err := json.Unmarshal(data, &users); err != nil {
		return errors.InternalServerError("Failed to unmarshal user data: " + err.Error())
	}

	if len(users) == 0 {
		return errors.NotFound("User not found")
	}

	// Ensure the user ID matches the authenticated user's ID
	if users[0].ID != userID {
		return errors.Forbidden("You can only delete your own account")
	}

	// Delete the user from the database
	data, err = client.DELETE("users", query)
	if err != nil {
		return errors.InternalServerError("Failed to delete user: " + err.Error())
	}

	return errors.SuccessResponse(c, fiber.Map{
		"message": "User account deleted successfully",
		"data":    data,
	})
}
