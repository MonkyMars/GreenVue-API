package auth

import (
	"greentrade-eu/internal/db"

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
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON payload: " + err.Error(),
		})
	}

	// Validate required fields
	if payload.Email == "" || payload.Password == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Email and password are required",
		})
	}

	// Authenticate user
	authResp, err := client.Login(payload.Email, payload.Password)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid credentials: " + err.Error(),
		})
	}

	// Return login success response
	return c.Status(200).JSON(fiber.Map{
		"message":      "Login successful",
		"userId":       authResp.User.ID,
		"accessToken":  authResp.AccessToken,
		"refreshToken": authResp.RefreshToken,
		"expiresIn":    authResp.ExpiresIn,
	})
}
