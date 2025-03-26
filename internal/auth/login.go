package auth

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
	"greentrade-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
)

func LoginUser(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()

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

	// Authenticate user
	authResp, err := client.Login(lib.SanitizeInput(payload.Email), lib.SanitizeInput(payload.Password))

	if err != nil {
		// Determine if this is a credentials error or a server error
		return errors.Unauthorized("Invalid credentials")
	}

	// Return login success response
	return errors.SuccessResponse(c, fiber.Map{
		"userId":       authResp.User.ID,
		"accessToken":  authResp.AccessToken,
		"refreshToken": authResp.RefreshToken,
		"expiresIn":    authResp.ExpiresIn,
	})
}
