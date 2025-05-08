package auth

import (
	"encoding/json"
	"fmt"
	"greenvue-eu/internal/db"
	"greenvue-eu/lib"
	"greenvue-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

const viewName = "users_with_auth_info"

func GetUserById(c *fiber.Ctx) error {
	userId := c.Params("id")

	if userId == "" {
		return errors.BadRequest("User ID is required")
	}

	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to create database client")
	}

	// Get user by ID using the standardized GET operation
	query := fmt.Sprintf("id=eq.%s", userId)
	data, err := client.GET(c, viewName, query)
	if err != nil {
		return errors.DatabaseError("Failed to fetch user: " + err.Error())
	}

	// Handle empty response
	if len(data) == 0 || string(data) == "[]" {
		return errors.NotFound("User not found")
	}

	// Parse the user data
	var users []lib.User
	if err := json.Unmarshal(data, &users); err != nil {
		return errors.InternalServerError("Failed to parse user data")
	}

	if len(users) == 0 {
		return errors.NotFound("User not found")
	}

	return errors.SuccessResponse(c, fiber.Map{
		"user": users[0],
	})
}

func GetUserByAccessToken(c *fiber.Ctx) error {
	// Get claims from context (set by AuthMiddleware)
	claims, ok := c.Locals("user").(*Claims)
	if !ok {
		return errors.Unauthorized("Invalid token claims")
	}

	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to create database client")
	}

	// Get user by ID using the standardized GET operation
	query := fmt.Sprintf("id=eq.%s", claims.UserID)
	data, err := client.GET(c, viewName, query)
	if err != nil {
		return errors.DatabaseError("Failed to fetch user: " + err.Error())
	}

	// Handle empty response
	if len(data) == 0 || string(data) == "[]" {
		return errors.NotFound("User not found")
	}

	// Parse the user data
	var users []lib.User
	if err := json.Unmarshal(data, &users); err != nil {
		return errors.InternalServerError("Failed to parse user data")
	}

	if len(users) == 0 {
		return errors.NotFound("User not found")
	}

	return errors.SuccessResponse(c, fiber.Map{
		"user": users[0],
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

	// Validate the user data
	if err := errors.ValidateRequest(c, &userUpdate); err != nil {
		return errors.BadRequest(err.Error())
	}

	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to create database client")
	}

	// Make sure the ID is set correctly
	userUpdate.ID = userId

	// Update user using the standardized PATCH operation
	data, err := client.PATCH(c, "users", userId, userUpdate)
	if err != nil {
		return errors.DatabaseError("Failed to update user: " + err.Error())
	}

	// Handle empty response
	if len(data) == 0 || string(data) == "[]" {
		return errors.NotFound("User not found")
	}

	// Parse the updated user data
	var updatedUsers []lib.UpdateUser
	if err := json.Unmarshal(data, &updatedUsers); err != nil {
		return errors.InternalServerError("Failed to parse updated user data")
	}

	if len(updatedUsers) == 0 {
		return errors.NotFound("User not found")
	}

	return errors.SuccessResponse(c, updatedUsers[0])
}
