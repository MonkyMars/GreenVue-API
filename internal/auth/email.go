package auth

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib/errors"

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
