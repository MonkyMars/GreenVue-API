package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"greenvue-eu/lib"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

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
func (s *SupabaseClient) GET(c *fiber.Ctx, table, query string) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?%s", s.URL, table, query)

	if c == nil {
		return nil, fmt.Errorf("fiber context is nil")
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", c.Get("Authorization"))
	req.Header.Add("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("supabase error: status %d - %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// POST creates a new record
func (s *SupabaseClient) POST(c *fiber.Ctx, table string, data any, useServiceKey ...bool) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?select=*", s.URL, table)

	jsonData, err := json.Marshal(data)
	if err != nil {
		fmt.Println("Error marshaling data:", err)
		return nil, fmt.Errorf("error marshaling request data: %w", err)
	}

	payload := strings.NewReader(string(jsonData))

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}

	var Authorization string
	if len(useServiceKey) > 0 && useServiceKey[0] {
		Authorization = "Bearer " + s.APIKey
	} else {
		Authorization = c.Get("Authorization")
		fmt.Println("Using Authorization from context", Authorization)
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Set("Authorization", Authorization)
	req.Header.Add("Prefer", "return=representation")
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	fmt.Println("Auth Header:", req.Header.Get("Authorization"))
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	// Empty response is valid in some cases
	if len(body) == 0 {
		return []byte("{}"), nil
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("supabase error: %s", string(body))
	}

	return body, nil
}

// PATCH updates an existing record by ID
func (s *SupabaseClient) PATCH(c *fiber.Ctx, table string, id string, data any) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?id=eq.%s", s.URL, table, id)

	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", c.Get("Authorization"))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Prefer", "return=representation") // Ensures Supabase returns the updated record

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("supabase PATCH error (%d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// DELETE removes a record based on condition
func (s *SupabaseClient) DELETE(c *fiber.Ctx, table, conditions string) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?%s", s.URL, table, conditions)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create DELETE request: %w", err)
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", c.Get("Authorization"))
	req.Header.Add("Prefer", "return=minimal") // More efficient for DELETE operations

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute DELETE request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read DELETE response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("DELETE operation failed (status %d): %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// UploadImage uploads an image to Supabase storage
func (s *SupabaseClient) UploadImage(filename, bucket string, image []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.URL, bucket, filename)
	fmt.Printf("Uploading to URL: %s\n", url)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(image))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return nil, err
	}

	contentType := "image/jpeg"
	if strings.HasSuffix(filename, ".png") {
		contentType = "image/png"
	} else if strings.HasSuffix(filename, ".gif") {
		contentType = "image/gif"
	} else if strings.HasSuffix(filename, ".webp") {
		contentType = "image/webp"
	}

	fmt.Printf("Using content type: %s\n", contentType)

	req.Header.Set("Content-Type", contentType)
	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Printf("Error reading response: %v\n", err)
		return nil, err
	}

	fmt.Printf("Status code: %d\n", resp.StatusCode)
	if resp.StatusCode >= 400 {
		fmt.Printf("Error response: %s\n", string(body))
		return nil, fmt.Errorf("supabase storage error (%d): %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// SignUp registers a new user
func (s *SupabaseClient) SignUp(email, password string) (*lib.User, error) {
	url := fmt.Sprintf("%s/auth/v1/signup", s.URL)

	// Create request payload
	payload, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.APIKey)
	req.Header.Set("Authorization", "Bearer "+s.APIKey)
	req.Header.Set("Prefer", "return=representation")

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
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
	payload, err := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.APIKey)
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
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
	payload, err := json.Marshal(map[string]string{
		"type":  resend_type,
		"email": email,
	})

	if err != nil {
		return fmt.Errorf("failed to marshal request body: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.APIKey)
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	// Send request
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("resend confirmation email failed: %s", string(body))
	}

	return nil
}
