package auth

import (
	"fmt"
	"log"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
)

// HandleGoogleRegistrationStart initiates the OAuth flow for Google registration
func HandleGoogleRegistrationStart(c *fiber.Ctx) error {
	// Build the Google OAuth URL
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	redirectURI := os.Getenv("REDIRECT_URI")
	state := "register" // This state indicates registration vs login

	// If testing locally, make sure the redirect URI is properly set
	if redirectURI == "" {
		log.Println("Warning: REDIRECT_URI is not set. Using default callback URL.")
		frontendURL := c.Query("redirect_uri", "https://api.greenvue.eu/auth/callback/google")
		redirectURI = fmt.Sprintf("%s/auth/callback/google", frontendURL)
	}

	// Build the authorization URL with specific scopes
	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		clientID,
		url.QueryEscape(redirectURI),
		url.QueryEscape("profile email"),
		state,
	)

	// Redirect the user to Google's OAuth page
	return c.Redirect(authURL, fiber.StatusTemporaryRedirect)
}
