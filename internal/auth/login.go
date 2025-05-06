package auth

import (
	"context"
	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func LoginUser(c *fiber.Ctx) error {
	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to create database client")
	}

	// Create a new repository instance to use standardized operations
	repo := db.NewSupabaseRepository(client)

	// Define payload struct
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse JSON request body
	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Invalid request format")
	}

	// Validate required fields
	if err := errors.ValidateFields(map[string]string{
		"email":    payload.Email,
		"password": payload.Password,
	}); err != nil {
		return err
	}

	// Authenticate user (this is a specialized operation that doesn't fit standard CRUD)
	// We'll continue to use the Login method which is kept in the client for auth operations
	authResp, err := repo.Login(context.Background(), lib.SanitizeInput(payload.Email), lib.SanitizeInput(payload.Password))
	if err != nil {
		return errors.Unauthorized("Invalid credentials")
	}

	// Generate JWT tokens
	tokens, err := GenerateTokenPair(authResp.User.ID, authResp.User.Email)
	if err != nil {
		return errors.InternalServerError("Failed to generate tokens")
	}

	// Return login success response with JWT tokens
	return errors.SuccessResponse(c, fiber.Map{
		"userId":       authResp.User.ID,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
	})
}
