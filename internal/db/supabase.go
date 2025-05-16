package db

import (
	"encoding/json"
	"fmt"
	"greenvue/lib"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/gofiber/fiber/v2"
)

// Global client instance and mutex for thread safety
var (
	globalClient     *SupabaseClient
	globalClientOnce sync.Once
	globalClientMu   sync.RWMutex
)

type SupabaseClient struct {
	URL    string
	APIKey string
}

// InitGlobalClient initializes the global Supabase client if it doesn't exist yet
func InitGlobalClient(useServiceKey ...bool) (*SupabaseClient, error) {
	globalClientOnce.Do(func() {
		globalClient = NewSupabaseClient(useServiceKey...)
	})

	if globalClient == nil {
		return nil, fmt.Errorf("failed to initialize global Supabase client")
	}

	return globalClient, nil
}

// GetGlobalClient returns the global Supabase client instance
// If it hasn't been initialized yet, it creates a new instance
func GetGlobalClient() *SupabaseClient {
	globalClientMu.RLock()
	if globalClient != nil {
		defer globalClientMu.RUnlock()
		return globalClient
	}
	globalClientMu.RUnlock()

	// If client doesn't exist, initialize it
	globalClientMu.Lock()
	defer globalClientMu.Unlock()

	if globalClient == nil {
		globalClient = NewSupabaseClient()
	}

	return globalClient
}

// NewSupabaseClient creates a new Supabase client using environment variables
func NewSupabaseClient(useServiceKey ...bool) *SupabaseClient {
	url := os.Getenv("SUPABASE_URL")

	var apiKey string
	// Check if we should use the service key or the anon key
	if len(useServiceKey) > 0 && useServiceKey[0] {
		apiKey = os.Getenv("SUPABASE_SERVICE_KEY")
	} else {
		// Default to using the anon key
		apiKey = os.Getenv("SUPABASE_ANON")
	}

	// Validate that the required environment variables are set
	if url == "" || apiKey == "" {
		fmt.Println("ERROR: Supabase environment variables not set. SUPABASE_URL and SUPABASE_ANON or SUPABASE_SERVICE_KEY are required.")
		return nil
	}

	return &SupabaseClient{
		URL:    url,
		APIKey: apiKey,
	}
}

// GET performs a GET request to fetch data with optional query parameters
func (s *SupabaseClient) GET(table, query string) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?%s", s.URL, table, query)

	client := resty.New().
		SetTimeout(10*time.Second).
		SetHeader("apikey", s.APIKey).
		SetHeader("Authorization", s.APIKey).
		SetHeader("Accept", "application/json")

	resp, err := client.R().Get(url)
	if err != nil {
		return nil, err
	}

	body := resp.Body()

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("supabase error: status %d - %s", resp.StatusCode(), string(body))
	}

	return body, nil
}

// POST creates a new record
func (s *SupabaseClient) POST(table string, data any) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?select=*", s.URL, table)

	client := resty.New().
		SetTimeout(10*time.Second).
		SetHeader("apikey", s.APIKey).
		SetHeader("Authorization", "Bearer "+s.APIKey).
		SetHeader("Prefer", "return=representation").
		SetHeader("Content-Type", "application/json")

	resp, err := client.R().
		SetBody(data).
		Post(url)

	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}

	body := resp.Body()

	// Empty response is valid in some cases
	if len(body) == 0 {
		return []byte("{}"), nil
	}

	if resp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("supabase error: %s", string(body))
	}

	return body, nil
}

// PATCH updates an existing record by ID
func (s *SupabaseClient) PATCH(table string, id string, data any) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?id=eq.%s", s.URL, table, id)

	client := resty.New().
		SetTimeout(10*time.Second).
		SetHeader("apikey", s.APIKey).
		SetHeader("Authorization", "Bearer "+s.APIKey).
		SetHeader("Content-Type", "application/json").
		SetHeader("Prefer", "return=representation") // Ensures Supabase returns the updated record

	resp, err := client.R().
		SetBody(data).
		Patch(url)

	if err != nil {
		return nil, err
	}

	respBody := resp.Body()

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("supabase PATCH error (%d): %s", resp.StatusCode(), string(respBody))
	}

	return respBody, nil
}

// DELETE removes a record based on condition
func (s *SupabaseClient) DELETE(c *fiber.Ctx, table, conditions string) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?%s", s.URL, table, conditions)

	client := resty.New().
		SetTimeout(10*time.Second).
		SetHeader("apikey", s.APIKey).
		SetHeader("Authorization", c.Get("Authorization")).
		SetHeader("Prefer", "return=minimal") // More efficient for DELETE operations

	resp, err := client.R().Delete(url)
	if err != nil {
		return nil, fmt.Errorf("failed to execute DELETE request: %w", err)
	}

	respBody := resp.Body()

	if resp.StatusCode() < 200 || resp.StatusCode() >= 300 {
		return nil, fmt.Errorf("DELETE operation failed (status %d): %s", resp.StatusCode(), string(respBody))
	}

	return respBody, nil
}

// UploadImage uploads an image to Supabase storage
func (s *SupabaseClient) UploadImage(filename, bucket string, image []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.URL, bucket, filename)
	fmt.Printf("Uploading to URL: %s\n", url)

	contentType := "image/jpeg"
	if strings.HasSuffix(filename, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(filename, ".gif") {
		contentType = "image/gif"
	} else if strings.HasSuffix(filename, ".webp") {
		contentType = "image/webp"
	}

	fmt.Printf("Using content type: %s\n", contentType)

	client := resty.New().
		SetTimeout(30*time.Second).
		SetHeader("Content-Type", contentType).
		SetHeader("apikey", s.APIKey).
		SetHeader("Authorization", "Bearer "+s.APIKey)

	resp, err := client.R().
		SetBody(image).
		Post(url)

	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, err
	}

	body := resp.Body()

	fmt.Printf("Status code: %d\n", resp.StatusCode())
	if resp.StatusCode() >= 400 {
		fmt.Printf("Error response: %s\n", string(body))
		return nil, fmt.Errorf("supabase storage error (%d): %s", resp.StatusCode(), string(body))
	}

	return body, nil
}

// SignUp registers a new user
func (s *SupabaseClient) SignUp(email, password string) (*lib.User, error) {
	url := fmt.Sprintf("%s/auth/v1/signup", s.URL)

	// Create request payload
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	client := resty.New().
		SetTimeout(10*time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("apikey", s.APIKey).
		SetHeader("Authorization", "Bearer "+s.APIKey).
		SetHeader("Prefer", "return=representation")

	resp, err := client.R().
		SetBody(payload).
		Post(url)

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	body := resp.Body()

	// Check for HTTP errors
	if resp.StatusCode() != http.StatusOK && resp.StatusCode() != http.StatusCreated {
		return nil, fmt.Errorf("failed to sign up user: %s", string(body))
	}

	// Parse JSON response based on actual Supabase structure
	var userResp struct {
		ID    string `json:"id"`
		Email string `json:"email"`
	}

	if err := json.Unmarshal(body, &userResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	// Check if user ID exists
	if userResp.ID == "" {
		return nil, fmt.Errorf("user ID missing in response")
	}

	// Create the user object
	user := &lib.User{
		ID:    userResp.ID,
		Email: userResp.Email,
	}

	return user, nil
}

// Login authenticates a user
func (s *SupabaseClient) Login(email, password string) (*lib.AuthResponse, error) {
	url := fmt.Sprintf("%s/auth/v1/token?grant_type=password", s.URL)

	// Create request payload
	payload := map[string]string{
		"email":    email,
		"password": password,
	}

	client := resty.New().
		SetTimeout(10*time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("apikey", s.APIKey).
		SetHeader("Authorization", "Bearer "+s.APIKey)

	resp, err := client.R().
		SetBody(payload).
		Post(url)

	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	body := resp.Body()

	// Check for HTTP errors
	if resp.StatusCode() != http.StatusOK {
		return nil, fmt.Errorf("login failed: %s", string(body))
	}

	// Parse JSON response
	var authResp lib.AuthResponse
	if err := json.Unmarshal(body, &authResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &authResp, nil
}

func (s *SupabaseClient) ResendConfirmationEmail(email, resend_type string) error {
	url := fmt.Sprintf("%s/auth/v1/resend", s.URL)

	// Create request payload
	payload := map[string]string{
		"type":  resend_type,
		"email": email,
	}

	client := resty.New().
		SetTimeout(10*time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("apikey", s.APIKey).
		SetHeader("Authorization", "Bearer "+s.APIKey)

	resp, err := client.R().
		SetBody(payload).
		Post(url)

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("resend confirmation email failed: %s", string(resp.Body()))
	}

	return nil
}
