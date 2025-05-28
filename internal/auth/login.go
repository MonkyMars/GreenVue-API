package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"greenvue/internal/config"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"
	"log"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

func LoginUser(c *fiber.Ctx) error {
	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to create database client")
	}

	// Define payload struct
	var payload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse JSON request body
	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Invalid request format")
	}

	// Validate required fields
	if err := errors.ValidateFields(map[string]string{
		"email":    payload.Email,
		"password": payload.Password,
	}); err != nil {
		return err
	}

	// Authenticate user (this is a specialized operation that doesn't fit standard CRUD)
	// We'll continue to use the Login method which is kept in the client for auth operations
	authResp, err := client.Login(lib.SanitizeInput(payload.Email), lib.SanitizeInput(payload.Password))
	if err != nil {
		switch err.Error() {
		case "invalid_credentials":
			return errors.Unauthorized("Invalid email or password")
		case "user_not_found":
			return errors.NotFound("User not found")
		case "email_not_confirmed":
			return errors.Forbidden("Email not confirmed")
		}
	}

	// Generate JWT tokens
	tokens, err := GenerateTokenPair(authResp.User.ID, authResp.User.Email)
	if err != nil {
		return errors.InternalServerError("Failed to generate tokens")
	} // Set the tokens as secure cookies for web clients
	SetAuthCookies(c, tokens)

	// Return login success response with JWT tokens for React Native clients
	return errors.SuccessResponse(c, fiber.Map{
		"userId":       authResp.User.ID,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
	})
}

// LogoutUser handles user logout by clearing cookies
func LogoutUser(c *fiber.Ctx) error {
	// Clear all authentication cookies
	ClearAuthCookies(c)

	// Clear the user session
	c.Locals("user", nil)

	// Send success response
	return errors.SuccessResponse(c, fiber.Map{
		"message": "Successfully logged out",
	})
}

func generateStateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

func HandleGoogleLogin(c *fiber.Ctx) error {
	state := generateStateToken()
	// Load config to determine environment
	cfg := config.LoadConfig()

	// Get the hostname and extract domain for the cookie
	host := c.Hostname()
	var domain string
	// For local development, don't set the domain at all
	if cfg.Environment != "production" {
		domain = ""
	} else {
		// For production, use the parent domain to share cookies across subdomains
		// This ensures the cookie works for all subdomains (www, api, etc)
		if strings.Contains(host, "greenvue.eu") {
			domain = "greenvue.eu" // Parent domain shared by all subdomains
		} else {
			domain = host
		}
	}

	// Log the state for debugging
	log.Printf("Generated OAuth state: %s", state)

	// Save state in cookie with proper settings
	c.Cookie(&fiber.Cookie{
		Name:     "oauthstate",
		Value:    state,
		Path:     "/",
		Domain:   domain,
		MaxAge:   3600,                      // 1 hour
		Expires:  time.Now().Add(time.Hour), // 1 hour
		HTTPOnly: true,
		Secure:   cfg.Environment == "production", // Only secure in production
		SameSite: "Lax",                           // Use Lax for better compatibility
	})

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	redirectURI := os.Getenv("REDIRECT_URI")
	scope := "openid email profile"

	authURL := fmt.Sprintf(
		"https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s",
		clientID,
		url.QueryEscape(redirectURI),
		url.QueryEscape(scope),
		state,
	)

	return c.Redirect(authURL)
}
