package auth

import (
	"encoding/json"
	"fmt"
	"greenvue-eu/internal/db"
	"greenvue-eu/lib"
	"greenvue-eu/lib/errors"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"

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

func VerifyEmailRedirect(c *fiber.Ctx) error {
	redirectURI := c.Query("redirect_uri")
	metaData := c.Query("metadata")

	log.Println("Received metadata:", metaData)

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
	userID, ok := parsedMetadata["sub"].(string)
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
	query := fmt.Sprintf("id=eq.%s", userID)
	data, err := client.GET("users", query)
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
	httpClient := &http.Client{}
	req, err := http.NewRequest("GET", redirectURI, nil)
	if err != nil {
		return errors.InternalServerError("Failed to create request: " + err.Error())
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return errors.InternalServerError("Failed to send request: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.InternalServerError("Email verification link is expired or invalid")
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.InternalServerError("Failed to read response body: " + err.Error())
	}

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
