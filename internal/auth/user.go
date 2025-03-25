package auth

import (
	"greentrade-eu/internal/db"

	"github.com/gofiber/fiber/v2"
)

func GetUserById(c *fiber.Ctx) error {
	userId := c.Params("id")

	client := db.NewSupabaseClient()

	// Get user by ID
	user, err := client.GetUserById(userId)

	if err != nil {
		return err
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "User found",
		"user":    user,
	})
}
