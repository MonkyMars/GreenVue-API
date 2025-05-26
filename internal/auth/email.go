package auth

import (
	"fmt"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/email"
	"greenvue/lib/errors"
	"net/http"
	"net/url"
	"strings"
	"time"

	"encoding/json"

	"github.com/go-resty/resty/v2"
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
		return errors.BadRequest("Email is required")
	}

	// Check if the user exists before queuing the email
	query := fmt.Sprintf("email=eq.%s", requestBody.Email)
	data, err := client.GET("user_details", query)
	if err != nil {
		return errors.InternalServerError("Failed to verify user: " + err.Error())
	}

	var users []lib.User
	if err := json.Unmarshal(data, &users); err != nil {
		return errors.InternalServerError("Failed to parse user data: " + err.Error())
	}

	if len(users) == 0 {
		return errors.NotFound("User not found")
	}

	if users[0].EmailVerified {
		return errors.BadRequest("Email already confirmed")
	}
	// Create email object for sending
	emailToSend := email.Email{
		ID:         lib.GenerateUUID(),
		To:         requestBody.Email,
		Subject:    "Please confirm your email address",
		Type:       email.ConfirmationEmail,
		TemplateID: "signup", // This is the resendType
		Status:     "pending",
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}

	// Queue the email but don't process it immediately
	// Let the scheduled job handle it to prevent duplicate sends
	err = email.QueueEmail(emailToSend)
	if err != nil {
		return errors.InternalServerError("Failed to queue confirmation email: " + err.Error())
	}
	return errors.SuccessResponse(c, "Confirmation email queued successfully")
}

func VerifyEmailRedirect(c *fiber.Ctx) error {
	redirectURI := c.Query("redirect_uri")
	metaData := c.Query("metadata")

	if redirectURI == "" {
		return errors.BadRequest("Missing redirect_uri parameter")
	}

	// Check if redirect_uri is a valid URL
	parsedURI, err := url.ParseRequestURI(redirectURI)
	if err != nil || parsedURI.Host == "" {
		return errors.BadRequest("Invalid redirect_uri format")
	}

	// Parse base64-encoded JSON metadata
	if metaData == "" {
		return errors.BadRequest("Missing metadata")
	}

	var parsedMetadata map[string]any
	if metaData != "" {
		decodedMetadata, err := url.QueryUnescape(metaData)
		if err != nil {
			return errors.BadRequest("Invalid metadata format: " + err.Error())
		}

		// Parse the metadata string directly
		parsedMetadata = make(map[string]any)

		// Split the string by spaces and commas to extract key-value pairs
		// Format is typically "map[key:value key2:value2]"
		if len(decodedMetadata) > 4 && decodedMetadata[:4] == "map[" {
			// Remove the "map[" prefix and "]" suffix
			content := decodedMetadata[4 : len(decodedMetadata)-1]

			// Custom parsing for the map string format
			var key string
			inKey := true
			current := ""

			for i := 0; i < len(content); i++ {
				char := content[i]

				if inKey && char == ':' {
					key = current
					current = ""
					inKey = false
				} else if !inKey && (char == ' ' || i == len(content)-1) {
					// If it's the last character, include it
					if i == len(content)-1 {
						current += string(char)
					}

					// Convert value to appropriate type if needed
					if current == "true" {
						parsedMetadata[key] = true
					} else if current == "false" {
						parsedMetadata[key] = false
					} else {
						parsedMetadata[key] = current
					}

					current = ""
					inKey = true
					// Skip the space
					if i < len(content)-1 && content[i+1] == ' ' {
						i++
					}
				} else {
					current += string(char)
				}
			}
		} else {
			return errors.BadRequest("Unrecognized metadata format")
		}
	}

	// Extract fields
	userId, ok := parsedMetadata["sub"].(string)
	if !ok {
		return errors.BadRequest("Missing or invalid user ID in metadata")
	}

	email, ok := parsedMetadata["email"].(string)
	if !ok {
		return errors.BadRequest("Missing or invalid email in metadata")
	}

	// Get DB client
	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to create database client")
	}

	// Check if user exists and matches
	query := fmt.Sprintf("id=eq.%s", userId)
	data, err := client.GET("user_details", query)
	if err != nil {
		return errors.InternalServerError("Failed to fetch user data: " + err.Error())
	}

	var user []lib.User
	if err := json.Unmarshal(data, &user); err != nil {
		return errors.InternalServerError("Failed to parse user data: " + err.Error())
	}

	if len(user) == 0 {
		return errors.BadRequest("User not found")
	}
	if user[0].Email != email {
		return errors.BadRequest("Email does not match user ID")
	}
	// Check if email link is expired
	restyClient := resty.New().SetTimeout(10 * time.Second)
	resp, err := restyClient.R().Get(redirectURI)

	if err != nil {
		return errors.InternalServerError("Failed to send request: " + err.Error())
	}
	if resp.StatusCode() != http.StatusOK {
		return errors.InternalServerError("Email verification link is expired or invalid")
	}

	bodyBytes := resp.Body()
	if strings.Contains(string(bodyBytes), "expired") {
		return errors.InternalServerError("Email verification link is expired or invalid")
	}

	// Mark email as verified
	_, err = client.PATCH("users", user[0].ID, map[string]any{
		"email_verified": true,
	})
	if err != nil {
		return errors.InternalServerError("Failed to update user data: " + err.Error())
	}

	// Redirect user
	return c.Redirect(redirectURI, fiber.StatusFound)
}
