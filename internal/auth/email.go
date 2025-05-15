package auth

import (
	"encoding/json"
	"fmt"
	"greenvue-eu/internal/db"
	"greenvue-eu/lib"
	"greenvue-eu/lib/errors"
	"net/url"

	"github.com/gofiber/fiber/v2"
)

func ResendConfirmationEmail(c *fiber.Ctx) error {
	client := db.GetGlobalClient()
	// Get user ID from the request body
	var requestBody struct {
		Email string `json:"email"`
	}
	if err := c.BodyParser(&requestBody); err != nil {
		return errors.BadRequest("Invalid request body: " + err.Error())
	}

	if requestBody.Email == "" {
		return errors.BadRequest("User ID is required")
	}

	err := client.ResendConfirmationEmail(requestBody.Email, "signup")

	if err != nil {
		if err.Error() == "user not found" {
			return errors.NotFound("User not found")
		}
		if err.Error() == "email already confirmed" {
			return errors.BadRequest("Email already confirmed")
		}
		if err.Error() == "email not verified" {
			return errors.BadRequest("Email not verified")
		}
		return errors.InternalServerError("Failed to send confirmation email: " + err.Error())
	}

	return errors.SuccessResponse(c, "Confirmation email resent successfully")
}

func VerifyEmailRedirect(c *fiber.Ctx) error {
	redirect_uri := c.Query("redirect_uri")
	metaData := c.Query("metadata")

	if redirect_uri == "" {
		return errors.BadRequest("Missing redirect_uri parameter")
	}

	// Check if the redirect_uri is a valid URL
	_, err := url.ParseRequestURI(redirect_uri)
	if err != nil {
		return errors.BadRequest("Invalid redirect_uri parameter")
	}

	var parsedMetadata map[string]any
	if metaData != "" {
		decodedMetadata, err := url.QueryUnescape(metaData)
		if err != nil {
			return errors.BadRequest("Invalid metadata format: " + err.Error())
		}

		if err := json.Unmarshal([]byte(decodedMetadata), &parsedMetadata); err != nil {
			return errors.BadRequest("Invalid metadata JSON: " + err.Error())
		}
	}

	// Extract the sub value
	userID, ok := parsedMetadata["sub"].(string)
	Email := parsedMetadata["email"].(string)
	if !ok {
		return errors.BadRequest("Missing or invalid user ID in metadata")
	}

	// Fetch user data to see if user exists and if email and user id match
	client := db.GetGlobalClient()

	if client == nil {
		return errors.InternalServerError("Failed to create database client")
	}

	query := fmt.Sprintf("id=eq.%s", userID)
	data, err := client.GET("users", query)
	if err != nil {
		return errors.InternalServerError("Failed to fetch user data: " + err.Error())
	}

	if len(data) == 0 {
		return errors.NotFound("User not found")
	}

	var user *lib.User

	if err := json.Unmarshal(data, &user); err != nil {
		return errors.InternalServerError("Failed to parse user data: " + err.Error())
	}

	if user.Email != Email {
		return errors.BadRequest("Email does not match user ID")
	}

	// Set verified to true
	_, err = client.PATCH("users", query, map[string]any{
		"email_verified": true,
	})

	if err != nil {
		return errors.InternalServerError("Failed to update user data: " + err.Error())
	}

	// Redirect to the specified URL
	return c.Redirect(redirect_uri, fiber.StatusFound)
}
