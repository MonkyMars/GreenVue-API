package auth

import (
	"fmt"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"encoding/json"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type GoogleTokenResponse struct {
	IDToken string `json:"id_token"`
}

type User struct {
	Id uuid.UUID `json:"id"`
}

type SupabaseResp struct {
	UserId       User   `json:"user"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
}

func HandleGoogleCallback(c *fiber.Ctx) error {
	// Check for error parameter from OAuth provider
	if errorMsg := c.Query("error"); errorMsg != "" {
		errorDescription := c.Query("error_description")
		return errors.BadRequest(fmt.Sprintf("OAuth error: %s", errorDescription))
	}

	code := c.Query("code")
	if code == "" {
		return errors.BadRequest("Missing code parameter")
	}

	// Check if this is a registration or login flow
	stateFromQuery := c.Query("state")

	if stateFromQuery == "" {
		return errors.BadRequest("Missing state parameter")
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

	// Set the tokens in a cookie
	tokens := TokenPair{
		AccessToken:  supabaseResp.AccessToken,
		RefreshToken: supabaseResp.RefreshToken,
	}

	SetAuthCookies(c, &tokens)

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
		err = handleUserRegistration(supabaseResp)
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
	formData := map[string]string{
		"code":          code,
		"client_id":     os.Getenv("GOOGLE_CLIENT_ID"),
		"client_secret": os.Getenv("GOOGLE_CLIENT_SECRET"),
		"redirect_uri":  os.Getenv("REDIRECT_URI"),
		"grant_type":    "authorization_code",
	}

	client := resty.New().SetTimeout(10 * time.Second)
	resp, err := client.R().
		SetFormData(formData).
		SetResult(&GoogleTokenResponse{}).
		Post("https://oauth2.googleapis.com/token")

	if err != nil {
		return nil, err
	}

	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("failed to exchange code: status %d", resp.StatusCode())
	}

	tokenResp := resp.Result().(*GoogleTokenResponse)
	return tokenResp, nil
}

func signInWithSupabase(idToken string) (SupabaseResp, error) {
	body := map[string]string{
		"provider": "google",
		"id_token": idToken,
	}

	client := resty.New().SetTimeout(10 * time.Second)
	resp, err := client.R().
		SetHeader("Accept", "application/json").
		SetHeader("Content-Type", "application/json").
		SetHeader("apikey", os.Getenv("SUPABASE_ANON")).
		SetBody(body).
		Post(os.Getenv("SUPABASE_URL") + "/auth/v1/token?grant_type=id_token")

	if err != nil {
		return SupabaseResp{}, err
	}

	var supabaseResp map[string]any
	if err := json.Unmarshal(resp.Body(), &supabaseResp); err != nil {
		return SupabaseResp{}, err
	}

	UserId := supabaseResp["user"].(map[string]any)["id"].(string)

	UserUuid, err := uuid.Parse(UserId)

	if err != nil || UserUuid == uuid.Nil {
		return SupabaseResp{}, fmt.Errorf("user ID not found in Supabase response")
	}

	Email := supabaseResp["user"].(map[string]any)["email"].(string)
	if Email == "" {
		return SupabaseResp{}, fmt.Errorf("email not found in Supabase response")
	}

	authTokens, err := GenerateTokenPair(UserUuid, Email)
	if err != nil {
		return SupabaseResp{}, fmt.Errorf("failed to generate token pair: %v", err)
	}

	tokens := SupabaseResp{
		UserId: User{
			Id: UserUuid,
		},
		AccessToken:  authTokens.AccessToken,
		RefreshToken: authTokens.RefreshToken,
		ExpiresIn:    authTokens.ExpiresIn,
	}

	log.Println(authTokens.AccessToken)

	return tokens, nil
}

// handleUserRegistration handles creating a user record in our custom users table
// when a user signs up with Google OAuth.
func handleUserRegistration(supabaseResp SupabaseResp) error {

	// Create Supabase client to interact with the database
	client := db.NewSupabaseClient(true)
	if client == nil {
		return fmt.Errorf("failed to create database client")
	}

	// Check if user already exists in our users table
	query := fmt.Sprintf("id=eq.%s", supabaseResp.UserId.Id)
	data, err := client.GET("user_details", query)

	userExists := false
	if err == nil && len(data) > 0 && string(data) != "[]" {
		// User already exists in our database
		userExists = true
		return nil // Nothing to do if user exists
	}

	// Get user details (email, name) from Supabase user profile
	userEmail := ""
	userName := ""
	picture := ""

	// Get user profile from Supabase to extract email and name
	userUrl := fmt.Sprintf("%s/auth/v1/user", os.Getenv("SUPABASE_URL"))

	restyClient := resty.New().SetTimeout(10 * time.Second)
	resp, err := restyClient.R().
		SetHeader("Authorization", "Bearer "+supabaseResp.AccessToken).
		SetHeader("apikey", os.Getenv("SUPABASE_ANON")).
		Get(userUrl)

	if err != nil {
		return fmt.Errorf("failed to get user profile: %v", err)
	}

	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("failed to get user profile, status: %d", resp.StatusCode())
	}

	var userProfile map[string]any
	if err := json.Unmarshal(resp.Body(), &userProfile); err != nil {
		return fmt.Errorf("failed to decode user profile: %v", err)
	}

	// Extract email
	if email, ok := userProfile["email"].(string); ok {
		userEmail = email
	} else {
		return fmt.Errorf("no email found in user profile")
	}

	// Extract user's name
	// First try to get from user_metadata
	if userMetadata, ok := userProfile["user_metadata"].(map[string]any); ok {
		if fullName, ok := userMetadata["full_name"].(string); ok && fullName != "" {
			userName = fullName
		} else if name, ok := userMetadata["name"].(string); ok && name != "" {
			userName = name
		}
		if userPicture, ok := userMetadata["picture"].(string); ok && userPicture != "" {
			picture = userPicture
		}
	}

	// If name is still empty, try to get it from other fields
	if userName == "" {
		if name, ok := userProfile["name"].(string); ok && name != "" {
			userName = name
		} else {
			// Use email prefix as fallback
			if userEmail != "" {
				userName = strings.Split(userEmail, "@")[0]
			} else {
				userName = "User"
			}
		}
	}

	// If user doesn't exist in our custom users table, create them
	if !userExists {
		// Create a user record using the standardized type
		newUser := lib.User{
			ID:            supabaseResp.UserId.Id,
			Email:         lib.SanitizeInput(userEmail),
			Name:          lib.SanitizeInput(userName),
			EmailVerified: true,
			Picture:       picture,
			Provider:      "google",
		}

		// Insert user into the database
		_, err = client.POST("users", newUser)
		if err != nil {
			return fmt.Errorf("failed to store user in database: %v", err)
		}

	}
	return nil
}
