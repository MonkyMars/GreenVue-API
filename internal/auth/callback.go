package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"greenvue-eu/lib/errors"
	"net/http"
	"net/url"
	"os"

	"github.com/gofiber/fiber/v2"
)

type GoogleTokenResponse struct {
	IDToken string `json:"id_token"`
}

func HandleGoogleCallback(c *fiber.Ctx) error {
	code := c.Query("code")

	if code == "" {
		return errors.BadRequest("Missing code parameter")
	}

	stateFromQuery := c.Query("state")
	stateFromCookie := c.Cookies("oauthstate")

	fmt.Println("Query State:", stateFromQuery)
	fmt.Println("Cookie State:", stateFromCookie)

	if stateFromCookie == "" || stateFromQuery == "" {
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
