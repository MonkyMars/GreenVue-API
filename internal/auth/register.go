package auth

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func RegisterUser(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()

	// Define payload struct
	var payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
		Location string `json:"location"`
	}

	// Parse JSON request body
	if err := errors.ValidateRequest(c, &payload); err != nil {
		return err
	}

	// no need to validate username and email since it happens on the frontend using zod.

	// Sign up the user
	user, err := client.SignUp(payload.Email, payload.Password)
	if err != nil {
		return errors.DatabaseError("Failed to register user: " + err.Error())
	}

	// Check if user or user.ID is nil
	if user == nil {
		return errors.DatabaseError("User registration failed: received nil user from auth provider")
	}

	if user.ID == "" {
		return errors.DatabaseError("User registration failed: received empty user ID from auth provider")
	}

	parsedPayload := db.User{
		ID:       user.ID,
		Name:     payload.Name,
		Email:    payload.Email,
		Location: payload.Location,
	}

	// Insert user into the database
	err = client.InsertUser(parsedPayload)
	if err != nil {
		return errors.DatabaseError("Failed to store user in database: " + err.Error())
	}

	// Generate JWT tokens for the new user
	tokens, err := GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return errors.InternalServerError("Failed to generate authentication tokens")
	}

	// Return success response with tokens
	return errors.SuccessResponse(c, fiber.Map{
		"userId":       user.ID,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
	})
}
