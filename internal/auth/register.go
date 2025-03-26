package auth

import (
	"greentrade-eu/internal/db"
	"greentrade-eu/lib"
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

	// Validate required fields
	if err := errors.ValidateFields(map[string]string{
		"name":     payload.Name,
		"email":    payload.Email,
		"password": payload.Password,
	}); err != nil {
		return err
	}

	valid, reason := lib.UsernameValidation(payload.Name)
	if !valid {
		return errors.ValidationError(reason, "name")
	}

	sanitizedEmail := lib.SanitizeInput(payload.Email)
	sanitizedUser := lib.SanitizeInput(payload.Name)

	// no need to validate username and email since it happens on the frontend.

	// Sign up the user
	user, err := client.SignUp(sanitizedEmail, payload.Password)
	if err != nil {
		return errors.DatabaseError("Failed to register user: " + err.Error())
	}

	parsedPayload := db.User{
		ID:       user.ID,
		Name:     sanitizedUser,
		Email:    sanitizedEmail,
		Location: lib.SanitizeInput(payload.Location),
	}

	// Insert user into the database
	err = client.InsertUser(parsedPayload)
	if err != nil {
		return errors.DatabaseError("Failed to store user in database: " + err.Error())
	}

	// Return success response
	return errors.SuccessResponse(c, fiber.Map{
		"message": "User registered successfully",
		"userId":  user.ID,
	})
}
