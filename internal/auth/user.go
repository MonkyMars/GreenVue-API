package auth

import (
	"fmt"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

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
	query := fmt.Sprintf("id=eq.%s", claims.UserId)
	data, err := client.GET("users", query)
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
	claims, ok := c.Locals("user").(*Claims)
	if !ok {
		return errors.Unauthorized("Invalid token claims")
	}

	userId := claims.UserId

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
	data, err := client.PATCH("users", userId, userUpdate)
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

func DownloadUserData(c *fiber.Ctx) error {
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
	var User []lib.User

	query := fmt.Sprintf("id=eq.%s", claims.UserId)
	data, err := client.GET("users", query)
	if err != nil {
		return errors.DatabaseError("Failed to fetch user: " + err.Error())
	}

	// Handle empty response
	if len(data) == 0 || string(data) == "[]" {
		return errors.NotFound("User not found")
	}

	// Parse the user data
	if err := json.Unmarshal(data, &User); err != nil {
		return errors.InternalServerError("Failed to parse user data")
	}

	// Get User's listings
	var Listings []lib.FetchedListing

	query = fmt.Sprintf("seller_id=eq.%s", claims.UserId)
	data, err = client.GET("listings_with_seller", query)
	if err != nil {
		return errors.DatabaseError("Failed to fetch listings: " + err.Error())
	}

	// Empty response is okay due to the user having no listings
	if err = json.Unmarshal(data, &Listings); err != nil {
		return errors.InternalServerError("Failed to parse listings data")
	}

	// Get User's reviews
	var Reviews []lib.FetchedReview
	query = fmt.Sprintf("user_id=eq.%s", claims.UserId)
	data, err = client.GET("review_with_username", query)
	if err != nil {
		return errors.DatabaseError("Failed to fetch reviews: " + err.Error())
	}

	// Empty response is okay due to the user having no reviews
	if err = json.Unmarshal(data, &Reviews); err != nil {
		return errors.InternalServerError("Failed to parse reviews data")
	}

	// Get User's messages
	var Messages []lib.Message
	query = fmt.Sprintf("sender_id=eq.%s", claims.UserId)
	data, err = client.GET("messages", query)
	if err != nil {
		return errors.DatabaseError("Failed to fetch messages: " + err.Error())
	}
	// Empty response is okay due to the user having no messages
	if err = json.Unmarshal(data, &Messages); err != nil {
		return errors.InternalServerError("Failed to parse messages data")
	}

	// Get User's favorites
	var Favorites []lib.FetchedFavorite
	query = fmt.Sprintf("user_id=eq.%s", claims.UserId)
	data, err = client.GET("user_favorites_view", query)
	if err != nil {
		return errors.DatabaseError("Failed to fetch favorites: " + err.Error())
	}

	// Empty response is okay due to the user having no favorites
	if err = json.Unmarshal(data, &Favorites); err != nil {
		return errors.InternalServerError("Failed to parse favorites data")
	}

	// Create a complete user object
	completeUser := lib.CompleteUser{
		User:      User[0],
		Listings:  Listings,
		Reviews:   Reviews,
		Messages:  Messages,
		Favorites: Favorites,
	}

	return errors.SuccessResponse(c, completeUser)
}
