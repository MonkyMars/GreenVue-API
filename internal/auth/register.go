package auth

import (
	"greentrade-eu/internal/db"

	"github.com/gofiber/fiber/v2"
)

func RegisterUser(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()

	var payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to parse JSON body: " + err.Error(),
		})
	}

	name := payload.Name
	email := payload.Email
	password := payload.Password

	// Register user
	_, err := client.SignUp(name, email, password)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "User registered successfully",
	})
}
