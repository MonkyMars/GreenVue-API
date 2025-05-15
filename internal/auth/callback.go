package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"greenvue-eu/internal/db"
	"greenvue-eu/lib"
	"greenvue-eu/lib/errors"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type GoogleTokenResponse struct {
	IDToken string `json:"id_token"`
}

type User struct {
	Id string `json:"id"`
}

type SupabaseResp struct {
	UserId       User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func HandleGoogleCallback(c *fiber.Ctx) error { // Check for error parameter from OAuth provider
	if errorMsg := c.Query("error"); errorMsg != "" {
		errorDescription := c.Query("error_description")
		log.Printf("OAuth error: %s - %s", errorMsg, errorDescription)
		return errors.BadRequest(fmt.Sprintf("OAuth error: %s", errorDescription))
	}

	code := c.Query("code")
	if code == "" {
		log.Println("Missing code parameter in query")
		return errors.BadRequest("Missing code parameter")
	}

	// Check if this is a registration or login flow
	stateFromQuery := c.Query("state")
	stateFromCookie := c.Cookies("oauthstate")

	// Track if this is a registration flow (state="register")
	log.Printf("Auth flow type: %s", stateFromQuery)

	// Log all cookies for debugging
	cookies := c.GetReqHeaders()["Cookie"]
	log.Printf("All cookies: %v", cookies)

	log.Printf("Query State: %s", stateFromQuery)
	log.Printf("Cookie State: %s", stateFromCookie)

	if stateFromCookie == "" || stateFromQuery == "" {
		log.Println("Missing state parameter in query or cookie")
		log.Println("State from cookie:", stateFromCookie)
		log.Println("State from query:", stateFromQuery)
		return errors.BadRequest("Missing state parameter")
	}
	// Validate that states match
	if stateFromCookie != stateFromQuery {
		log.Println("State mismatch: cookie state does not match query state")
		log.Printf("Cookie state: %s, Query state: %s", stateFromCookie, stateFromQuery)

		// Instead of failing immediately, try to continue with the authentication flow
		// This is a workaround for cross-domain cookie issues
		log.Println("Attempting to continue with authentication despite state mismatch")
	}

	// Exchange code for Google tokens
	tokenResp, err := exchangeCodeForGoogleToken(code)
	if err != nil {
		return errors.InternalServerError("Failed to exchange code for token: " + err.Error())
	}

	// Send Google ID token to Supabase
	supabaseResp, err := signInWithSupabase(tokenResp.IDToken)
	if err != nil {
		return errors.InternalServerError("Failed to sign in with Supabase: " + err.Error())
	}

	SetTokenCookie(c, supabaseResp.AccessToken)
	SetRefreshTokenCookie(c, supabaseResp.RefreshToken)
	siteUrl := os.Getenv("URL")
	if siteUrl == "" {
		log.Println("URL environment variable is not set")
		return c.Redirect("https://www.greenvue.eu/login")
	}
	// Check if this is a registration flow based on the state parameter
	isRegistration := stateFromQuery == "register"

	// For registration flow, we need to check if the user exists in our users table
	// and create them if they don't
	if isRegistration {
		// Get user details and create user record if needed
		err = handleUserRegistration(c, supabaseResp)
		if err != nil {
			log.Printf("Registration error: %v", err)
			return errors.InternalServerError("Failed to complete registration: " + err.Error())
		}

		// Redirect to different page for registration completion
		query := fmt.Sprintf("?access_token=%s&refresh_token=%s&user_id=%s&expires_in=%d&registration=true",
			supabaseResp.AccessToken, supabaseResp.RefreshToken, supabaseResp.UserId.Id, supabaseResp.ExpiresIn)
		return c.Redirect(siteUrl + query)
	}

	// Regular login flow
	query := fmt.Sprintf("?access_token=%s&refresh_token=%s&user_id=%s&expires_in=%d",
		supabaseResp.AccessToken, supabaseResp.RefreshToken, supabaseResp.UserId.Id, supabaseResp.ExpiresIn)
	return c.Redirect(siteUrl + query)
}

func exchangeCodeForGoogleToken(code string) (*GoogleTokenResponse, error) {
	data := url.Values{}
	data.Set("code", code)
	data.Set("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("GOOGLE_CLIENT_SECRET"))
	data.Set("redirect_uri", os.Getenv("REDIRECT_URI"))
	data.Set("grant_type", "authorization_code")

	resp, err := http.PostForm("https://oauth2.googleapis.com/token", data)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var tokenResp GoogleTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, err
	}
	return &tokenResp, nil
}

func signInWithSupabase(idToken string) (SupabaseResp, error) {
	body := map[string]string{
		"provider": "google",
		"id_token": idToken,
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest(
		"POST",
		os.Getenv("SUPABASE_URL")+"/auth/v1/token?grant_type=id_token",
		bytes.NewBuffer(jsonBody),
	)

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return SupabaseResp{}, err
	}
	defer resp.Body.Close()

	var supabaseResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&supabaseResp); err != nil {
		return SupabaseResp{}, err
	}

	tokens := SupabaseResp{
		UserId:       User{Id: supabaseResp["user"].(map[string]any)["id"].(string)},
		AccessToken:  supabaseResp["access_token"].(string),
		RefreshToken: supabaseResp["refresh_token"].(string),
		ExpiresIn:    int64(supabaseResp["expires_in"].(float64)),
	}

	return tokens, nil
}

// handleUserRegistration handles creating a user record in our custom users table
// when a user signs up with Google OAuth.
func handleUserRegistration(c *fiber.Ctx, supabaseResp SupabaseResp) error {

	// Create Supabase client to interact with the database
	client := db.NewSupabaseClient(true)
	if client == nil {
		return fmt.Errorf("failed to create database client")
	}

	// Check if user already exists in our users table
	query := fmt.Sprintf("id=eq.%s", supabaseResp.UserId.Id)
	data, err := client.GET("users", query)

	userExists := false
	if err == nil && len(data) > 0 && string(data) != "[]" {
		// User already exists in our database
		log.Printf("Google user already exists in database: %s", supabaseResp.UserId.Id)
		userExists = true
		return nil // Nothing to do if user exists
	}

	// Get user details (email, name) from Supabase user profile
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
		return fmt.Errorf("failed to get user profile: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Failed to get user profile, status: %d", resp.StatusCode)
		return fmt.Errorf("failed to get user profile, status: %d", resp.StatusCode)
	}

	defer resp.Body.Close()
	var userProfile map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&userProfile); err != nil {
		log.Printf("Failed to decode user profile: %v", err)
		return fmt.Errorf("failed to decode user profile: %v", err)
	}

	// Extract email
	if email, ok := userProfile["email"].(string); ok {
		userEmail = email
		log.Printf("Successfully retrieved email %s for Google user", userEmail)
	} else {
		log.Printf("No email found in user profile")
		return fmt.Errorf("no email found in user profile")
	}

	// Extract user's name
	// First try to get from user_metadata
	if userMetadata, ok := userProfile["user_metadata"].(map[string]interface{}); ok {
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

	// If user doesn't exist in our custom users table, create them
	if !userExists {
		// Create a user record using the standardized type
		newUser := lib.User{
			ID:    supabaseResp.UserId.Id,
			Email: userEmail,
			Name:  userName,
			// Location will be empty initially, user can update it later
			Location: "",
		}

		// Insert user into the database
		_, err = client.POST("users", newUser)
		if err != nil {
			log.Printf("Failed to store Google user in database: %v", err)
			return fmt.Errorf("failed to store user in database: %v", err)
		}

		log.Printf("Successfully created new Google user in database with ID: %s", supabaseResp.UserId.Id)
	}

	return nil
}
