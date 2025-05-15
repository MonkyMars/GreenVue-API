package auth

import (
	"encoding/json"
	"fmt"
	"greenvue-eu/internal/db"
	"greenvue-eu/lib"
	"greenvue-eu/lib/errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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
		Location string `json:"location"`
	}

	// Parse JSON request body
	if err := errors.ValidateRequest(c, &payload); err != nil {
		return err
	}

	// No need to validate username and email since it happens on the frontend using zod.

	// Sign up the user (this is a specialized operation that doesn't fit standard CRUD)
	// We'll continue to use the SignUp method which is kept in the repository for auth operations
	user, err := client.SignUp(lib.SanitizeInput(payload.Email), payload.Password)

	if err != nil {
		return errors.DatabaseError("Failed to register user: " + err.Error())
	}

	// Check if user or user.ID is nil
	if user == nil {
		return errors.DatabaseError("User registration failed: received nil user from auth provider")
	}

	if user.ID == "" {
		return errors.DatabaseError("User registration failed: received empty user ID from auth provider")
	}

	// Create a user record using the standardized type
	newUser := lib.User{
		ID:       user.ID,
		Name:     payload.Name,
		Email:    payload.Email,
		Location: payload.Location,
	}

	// Insert user into the database using standardized Create operation
	_, err = client.POST(c, "users", newUser, true)

	if err != nil {
		return errors.DatabaseError("Failed to store user in database: " + err.Error())
	}

	// Generate JWT tokens for the new user
	tokens, err := GenerateTokenPair(user.ID, user.Email)
	if err != nil {
		return errors.InternalServerError("Failed to generate authentication tokens")
	} // Set the tokens as secure cookies for web clients
	SetTokenCookie(c, tokens.AccessToken)
	SetRefreshTokenCookie(c, tokens.RefreshToken)

	// Return success response with tokens for React Native clients
	return errors.SuccessResponse(c, fiber.Map{
		"userId":       user.ID,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
	})
}

func HandleGoogleRegister(c *fiber.Ctx) error {
	// Parse request body
	var payload struct {
		IDToken  string `json:"id_token"` // Google ID token
		Location string `json:"location"` // Optional location from user
	}

	// Parse JSON request body
	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Invalid request format")
	}

	// Validate that we have the required ID token
	if payload.IDToken == "" {
		return errors.BadRequest("Missing ID token")
	}

	// Location is optional, no need to validate

	// Use the Google ID token to sign in with Supabase
	supabaseResp, err := signInWithSupabase(payload.IDToken)
	if err != nil {
		log.Printf("Google auth error: %v", err)
		return errors.InternalServerError("Failed to authenticate with Google: " + err.Error())
	}

	// Log successful authentication
	log.Printf("Successfully authenticated Google user with ID: %s", supabaseResp.UserId.Id)

	// Create Supabase client to interact with the database
	client := db.NewSupabaseClient(true)
	if client == nil {
		return errors.InternalServerError("Failed to create database client")
	}

	// Check if user already exists in our users table
	query := fmt.Sprintf("id=eq.%s", supabaseResp.UserId.Id)
	data, err := client.GET("users", query)

	userExists := false
	if err == nil && len(data) > 0 && string(data) != "[]" {
		// User already exists in our database
		userExists = true
		log.Printf("Google user already exists in database: %s", supabaseResp.UserId.Id)
	} // Get user details (email, name) from Supabase user profile
	userEmail := ""
	userName := ""

	// Get user profile from Supabase to extract email and name
	userUrl := fmt.Sprintf("%s/auth/v1/user", os.Getenv("SUPABASE_URL"))
	req, _ := http.NewRequest("GET", userUrl, nil)
	req.Header.Set("Authorization", "Bearer "+supabaseResp.AccessToken)
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON"))
	httpClient := &http.Client{Timeout: 10 * time.Second}
	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("Failed to get user profile: %v", err)
	} else if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to get user profile, status: %d", resp.StatusCode)
	} else {
		defer resp.Body.Close()
		var userProfile map[string]any
		if err := json.NewDecoder(resp.Body).Decode(&userProfile); err != nil {
			log.Printf("Failed to decode user profile: %v", err)
		} else {
			// Extract email
			if email, ok := userProfile["email"].(string); ok {
				userEmail = email
				log.Printf("Successfully retrieved email %s for Google user", userEmail)
			} else {
				log.Printf("No email found in user profile")
			}

			// Extract user's name
			// First try to get from user_metadata
			if userMetadata, ok := userProfile["user_metadata"].(map[string]any); ok {
				if fullName, ok := userMetadata["full_name"].(string); ok && fullName != "" {
					userName = fullName
					log.Printf("Using full_name from user_metadata: %s", userName)
				} else if name, ok := userMetadata["name"].(string); ok && name != "" {
					userName = name
					log.Printf("Using name from user_metadata: %s", userName)
				}
			}

			// If name is still empty, try to get it from other fields
			if userName == "" {
				if name, ok := userProfile["name"].(string); ok && name != "" {
					userName = name
					log.Printf("Using name from root profile: %s", userName)
				} else {
					// Use email prefix as fallback
					if userEmail != "" {
						userName = strings.Split(userEmail, "@")[0]
						log.Printf("Using email prefix as name: %s", userName)
					} else {
						userName = "User"
						log.Printf("No name found, using default: %s", userName)
					}
				}
			}
		}
	}
	// If user doesn't exist in our custom users table, create them
	if !userExists {
		// Create a user record using the standardized type
		newUser := lib.User{
			ID:            supabaseResp.UserId.Id,
			Email:         userEmail,
			Name:          userName,
			Location:      payload.Location,
			EmailVerified: true,
		}

		// Insert user into the database
		data, err = client.POST(c, "users", newUser, true)
		if err != nil {
			log.Printf("Failed to store Google user in database: %v", err)
			return errors.DatabaseError("Failed to store Google user in database: " + err.Error())
		}

		log.Println(string(data))

		log.Printf("Successfully created new Google user in database with ID: %s", supabaseResp.UserId.Id)
	} else {
		log.Println("Google user already exists in database, no need to create")
	}

	// Generate JWT tokens for the user
	tokens, err := GenerateTokenPair(supabaseResp.UserId.Id, userEmail)
	if err != nil {
		return errors.InternalServerError("Failed to generate authentication tokens")
	}

	// Set the tokens as secure cookies for web clients
	SetTokenCookie(c, tokens.AccessToken)
	SetRefreshTokenCookie(c, tokens.RefreshToken)

	// Return success response with tokens for React Native clients
	return errors.SuccessResponse(c, fiber.Map{
		"userId":       supabaseResp.UserId.Id,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
	})
}
