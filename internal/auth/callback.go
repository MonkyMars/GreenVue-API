package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"greenvue-eu/lib/errors"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
)

type GoogleTokenResponse struct {
	IDToken string `json:"id_token"`
}

func HandleGoogleCallback(c *fiber.Ctx) error {
	// Check for error parameter from OAuth provider
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

	stateFromQuery := c.Query("state")
	stateFromCookie := c.Cookies("oauthstate")

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
		return errors.BadRequest("Invalid state parameter")
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

	// Send token info to client (or set a cookie)
	return c.JSON(supabaseResp)
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

func signInWithSupabase(idToken string) (map[string]any, error) {
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
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", os.Getenv("SUPABASE_ANON_KEY"))

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var supabaseResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&supabaseResp); err != nil {
		return nil, err
	}

	return supabaseResp, nil
}
