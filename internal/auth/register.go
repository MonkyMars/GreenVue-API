package auth

import (
	"greentrade-eu/internal/db"
	
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
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid JSON payload: " + err.Error(),
		})
	}

	// Validate required fields
	if payload.Name == "" || payload.Email == "" || payload.Password == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Name, email, and password are required",
		})
	}

	// Register user in Supabase Auth
	user, err := client.SignUp(payload.Email, payload.Password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to register user: " + err.Error(),
		})
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
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to store user in database: " + err.Error(),
		})
	}

	// Return success response
	return c.Status(201).JSON(fiber.Map{
		"message": "User registered successfully",
		"userId":  user.ID,
	})
}