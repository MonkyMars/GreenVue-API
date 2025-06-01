package auth

import (
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/email"
	"greenvue/lib/errors"
	"greenvue/lib/validation"
	"time"

	"github.com/gofiber/fiber/v2"
)

func SendResetPasswordEmail(c *fiber.Ctx) error {
	var payload struct {
		Email string `json:"email"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Invalid request body: " + err.Error())
	}

	// Validate email format
	if payload.Email == "" {
		return errors.BadRequest("Email is required")
	}

	if valid, reason := validation.ValidateEmail(payload.Email); !valid {
		return errors.BadRequest("Invalid email format: " + reason)
	}
	// Queue a password reset email to be sent asynchronously
	resetEmail := email.Email{
		ID:         lib.GenerateUUID(),
		To:         payload.Email,
		Type:       email.PasswordResetEmail,
		CreatedAt:  time.Now(),
		Status:     "pending",
		MaxRetries: 3,
	}

	if err := email.QueueEmail(resetEmail); err != nil {
		return errors.InternalServerError("Failed to queue password reset email: " + err.Error())
	}
	return errors.SuccessResponse(c, fiber.Map{
		"message": "Password reset email has been queued successfully. Please check your inbox shortly.",
	})
}

func ChangePassword(c *fiber.Ctx) error {
	var payload struct {
		Password string `json:"password"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Invalid request body: " + err.Error())
	}

	if payload.Password == "" {
		return errors.BadRequest("New password is required")
	}

	if valid, reason := validation.ValidatePassword(payload.Password); !valid {
		return errors.BadRequest("Invalid password: " + reason)
	}

	claims, ok := c.Locals("user").(*Claims)
	if !ok || claims == nil {
		return errors.Unauthorized("User not authenticated")
	}

	client := db.GetGlobalClient()
	data, err := client.UpdateUser(claims.UserId, map[string]any{
		"password": payload.Password,
	})

	if err != nil {
		return errors.InternalServerError("Failed to update user: " + err.Error())
	}

	if data == nil {
		return errors.NotFound("User not found")
	}

	return errors.SuccessResponse(c, fiber.Map{
		"message": "Password changed successfully",
	})
}
