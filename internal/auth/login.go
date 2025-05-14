package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"greenvue-eu/internal/db"
	"greenvue-eu/lib"
	"greenvue-eu/lib/errors"
	"log"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
)

func LoginUser(c *fiber.Ctx) error {
	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to create database client")
	}

	// Create a new repository instance to use standardized operations
	repo := db.NewSupabaseRepository(client)

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
	authResp, err := repo.Login(context.Background(), lib.SanitizeInput(payload.Email), lib.SanitizeInput(payload.Password))
	if err != nil {
		log.Printf("Login error for email %s: %v", payload.Email, err)
		return errors.Unauthorized("Invalid credentials")
	}

	// Generate JWT tokens
	tokens, err := GenerateTokenPair(authResp.User.ID, authResp.User.Email)
	if err != nil {
		return errors.InternalServerError("Failed to generate tokens")
	} // Set the tokens as secure cookies for web clients
	SetTokenCookie(c, tokens.AccessToken)
	SetRefreshTokenCookie(c, tokens.RefreshToken)

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
	log.Printf("Processing logout request from %s", c.IP())

	// Log auth state before logout
	cookies := c.GetReqHeaders()["Cookie"]
	log.Printf("Cookies before logout: %v", cookies)

	// Clear all authentication cookies
	ClearAuthCookies(c)

	// Log success
	log.Printf("User logged out successfully")

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

	// Save state in cookie or session
	c.Cookie(&fiber.Cookie{
		Name:     "oauthstate",
		Value:    state,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "None",
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
