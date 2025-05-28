package auth

import (
	"encoding/json"
	"fmt"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"
	"greenvue/lib/validation"
	"log"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

func RegisterUser(c *fiber.Ctx) error {
	client := db.NewSupabaseClient(true)
	if client == nil {
		return errors.InternalServerError("Failed to create database client")
	}

	// Define payload struct
	var payload struct {
		Name     string `json:"name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	// Parse JSON request body
	if err := errors.ValidateRequest(c, &payload); err != nil {
		return err
	}

	if payload.Email == "" || payload.Password == "" || payload.Name == "" {
		return errors.BadRequest("Missing required user registration fields")
	}

	// Validate required fields
	validUsername, msg := validation.ValidateUsername(payload.Name)
	if !validUsername {
		return errors.BadRequest("Invalid username: " + msg)
	}

	validEmail, msg := validation.ValidateEmail(payload.Email)
	if !validEmail {
		return errors.BadRequest("Invalid email: " + msg)
	}

	sanitizedName := lib.SanitizeInput(payload.Name)
	sanitizedEmail := lib.SanitizeInput(payload.Email)

	if sanitizedName == "" || sanitizedEmail == "" {
		return errors.BadRequest("Invalid input: name and email cannot be empty")
	}

	// Check if the email is already registered
	existingUser, err := client.GET("users", "email=eq."+sanitizedEmail)
	if err != nil {
		return errors.DatabaseError("Failed to check existing user: " + err.Error())
	}

	var users []lib.User
	if err := json.Unmarshal(existingUser, &users); err != nil {
		return errors.InternalServerError("Failed to parse existing user data: " + err.Error())
	}

	if len(users) > 0 {
		return errors.AlreadyExists("Email is already registered")
	}

	// Sign up the user with the authentication provider
	user, err := client.SignUp(lib.SanitizeInput(payload.Email), payload.Password)

	if err != nil {
		return errors.DatabaseError("Failed to register user: " + err.Error())
	}

	// Check if user or user.ID is nil
	if user == nil {
		return errors.DatabaseError("User registration failed: received nil user from auth provider")
	}

	if user.ID == uuid.Nil {
		return errors.DatabaseError("User registration failed: received empty user ID from auth provider")
	}

	// Create a user record using the standardized type lib.User
	newUser := lib.User{
		ID:       user.ID,
		Email:    sanitizedEmail,
		Name:     sanitizedName,
		Provider: "email",
	}

	// Insert user into the database using standardized Create operation
	_, err = client.POST("users", newUser)

	if err != nil {
		return errors.DatabaseError("Failed to store user in database: " + err.Error())
	}

	// Generate JWT tokens for the new user
	tokens, err := GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return errors.InternalServerError("Failed to generate authentication tokens")
	}

	// Set the tokens as secure cookies for web clients
	SetAuthCookies(c, tokens)

	// Return success response with tokens for React Native clients
	return errors.SuccessResponse(c, fiber.Map{
		"userId":       user.ID,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
	})
}

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
