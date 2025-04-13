package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"greentrade-eu/lib"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type Listing struct {
	ID            string   `json:"id,omitempty"`
	CreatedAt     string   `json:"created_at,omitempty"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Condition     string   `json:"condition"`
	Price         int64    `json:"price"`
	Location      string   `json:"location"`
	EcoScore      float32  `json:"ecoScore"`
	EcoAttributes []string `json:"ecoAttributes"`
	Negotiable    bool     `json:"negotiable"`
	Title         string   `json:"title"`
	ImageUrl      []string `json:"imageUrl"`
	SellerID      string   `json:"seller_id"`
}

type FetchedListing struct {
	ID            string   `json:"id,omitempty"`
	CreatedAt     string   `json:"created_at,omitempty"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Condition     string   `json:"condition"`
	Price         int64    `json:"price"`
	Location      string   `json:"location"`
	EcoScore      float32  `json:"ecoScore"`
	EcoAttributes []string `json:"ecoAttributes"`
	Negotiable    bool     `json:"negotiable"`
	Title         string   `json:"title"`
	ImageUrl      []string `json:"imageUrl"`

	SellerID        string  `json:"seller_id"`
	SellerUsername  string  `json:"seller_username"`
	SellerBio       *string `json:"seller_bio,omitempty"`
	SellerCreatedAt string  `json:"seller_created_at"`
	SellerRating    float32 `json:"seller_rating"`
	SellerVerified  bool    `json:"seller_verified"`
}

type SupabaseClient struct {
	URL       string
	APIKey    string
	AuthToken string
}

// NewSupabaseClient creates a new Supabase client using environment variables
func NewSupabaseClient() *SupabaseClient {
	url := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_ANON")

	// Validate that the required environment variables are set
	if url == "" || apiKey == "" {
		fmt.Println("ERROR: Supabase environment variables not set. SUPABASE_URL and SUPABASE_ANON are required.")
		return nil
	}

	return &SupabaseClient{
		URL:    url,
		APIKey: apiKey,
	}
}

func (s *SupabaseClient) Query(table string, query string) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?%s", s.URL, table, query)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)
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

func (s *SupabaseClient) POST(table string, data Listing) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s", s.URL, table)

	jsonData, _ := json.Marshal(data)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("supabase error: %s", string(body))
	}
	return body, nil
}

func (s *SupabaseClient) DELETE(table, id string) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s", s.URL, table)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return respBody, nil
}

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

	client := &http.Client{}
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

func (s *SupabaseClient) PATCH(table, ID string, query fiber.Map) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?id=eq.%s", s.URL, table, ID)

	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Prefer", "return=representation") // Ensures Supabase returns the updated record

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func (s *SupabaseClient) PostRaw(table string, jsonData []byte) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s", s.URL, table)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)
	req.Header.Add("Prefer", "return=representation")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("supabase error: %s", string(body))
	}

	return body, nil
}

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
	client := &http.Client{}
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

	// Debug logging
	fmt.Printf("Supabase signup response: %s\n", string(body))

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("failed to sign up user: %s", string(body))
	}

	// Parse JSON response based on actual Supabase structure
	// The user data is at the root level, not nested under a "user" property
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

type User struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	Location  string `json:"location"`
	Bio       string `json:"bio"`
	CreatedAt string `json:"created_at"`
}

func (s *SupabaseClient) InsertUser(user User) error {
	url := fmt.Sprintf("%s/rest/v1/users", s.URL)

	// Create request payload with all required fields
	payload, err := json.Marshal(map[string]interface{}{
		"id":       user.ID,
		"name":     user.Name,
		"email":    user.Email,
		"location": user.Location,
		"bio":      user.Bio,
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
	req.Header.Set("Prefer", "return=minimal") // No need for full response

	// Send request
	client := &http.Client{}
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
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to insert user: %s", string(body))
	}

	return nil
}

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
	client := &http.Client{}
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

func (s *SupabaseClient) InsertSeller(ID, description string) error {
	url := fmt.Sprintf("%s/rest/v1/sellers", s.URL)

	req, err := http.NewRequest("POST", url, strings.NewReader(fmt.Sprintf(`{"id": "%s", "description": "%s"}`, ID, description)))
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("apikey", s.APIKey)
	req.Header.Set("Authorization", "Bearer "+s.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("failed to insert seller: %s", resp.Status)
	}

	return nil
}

func (s *SupabaseClient) GetUserById(userId string) (User, error) {
	// Validate userId is not empty
	if userId == "" {
		return User{}, fmt.Errorf("user ID cannot be empty")
	}

	url := fmt.Sprintf("%s/rest/v1/users?id=eq.%s", s.URL, userId)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return User{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)
	req.Header.Add("Prefer", "return=representation")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error sending request: %v\n", err)
		return User{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return User{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP error responses
	if resp.StatusCode >= 400 {
		return User{}, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Try to unmarshal as array first
	var users []User
	if err := json.Unmarshal(body, &users); err != nil {
		// If array unmarshal fails, try as single object
		var user User
		if err := json.Unmarshal(body, &user); err != nil {
			return User{}, fmt.Errorf("failed to parse response: %w", err)
		}
		return user, nil
	}

	if len(users) == 0 {
		return User{}, fmt.Errorf("user not found with ID: %s", userId)
	}

	return users[0], nil
}

func (s *SupabaseClient) GetUserByAccessToken(accessToken string) (User, error) {
	url := fmt.Sprintf("%s/auth/v1/user", s.URL)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return User{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+accessToken)
	req.Header.Add("Prefer", "return=representation")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return User{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP error responses
	if resp.StatusCode >= 400 {
		return User{}, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Try to unmarshal as array first
	var users []User
	if err := json.Unmarshal(body, &users); err != nil {
		// If array unmarshal fails, try as single object
		var user User
		if err := json.Unmarshal(body, &user); err != nil {
			return User{}, fmt.Errorf("failed to parse response: %w", err)
		}
		return user, nil
	}

	if len(users) == 0 {
		return User{}, fmt.Errorf("user not found with access token: %s", accessToken)
	}

	return users[0], nil
}

func (s *SupabaseClient) RefreshAccessToken(refreshToken string) (User, error) {
	encodedRefreshToken := url.QueryEscape(refreshToken)
	url := fmt.Sprintf("%s/auth/v1/token?grant_type=refresh_token&refresh_token=%s", s.URL, encodedRefreshToken)

	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return User{}, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)
	req.Header.Add("Prefer", "return=representation")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return User{}, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return User{}, fmt.Errorf("failed to read response: %w", err)
	}

	// Handle HTTP error responses
	if resp.StatusCode >= 400 {
		return User{}, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Parse JSON response
	var user User
	if err := json.Unmarshal(body, &user); err != nil {
		return User{}, fmt.Errorf("failed to parse response: %w", err)
	}

	return user, nil
}

func (s *SupabaseClient) UpdateUser(user *lib.UpdateUser) (*lib.UpdateUser, error) {
	url := fmt.Sprintf("%s/rest/v1/users?id=eq.%s", s.URL, user.ID)

	jsonData, err := json.Marshal(user)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal user: %w", err)
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)
	req.Header.Add("Prefer", "return=representation")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(body))
	}

	// Handle empty response case
	if len(body) == 0 || string(body) == "[]" {
		return nil, fmt.Errorf("user not found with ID: %s", user.ID)
	}

	// Try to unmarshal as array first (which is what Supabase usually returns)
	var users []lib.UpdateUser
	if err := json.Unmarshal(body, &users); err != nil {
		// If array unmarshal fails, try as single object
		var singleUser lib.UpdateUser
		if err := json.Unmarshal(body, &singleUser); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}
		return &singleUser, nil
	}

	// Check if we got any results
	if len(users) == 0 {
		return nil, fmt.Errorf("user not found with ID: %s", user.ID)
	}

	return &users[0], nil
}

func (s *SupabaseClient) UPDATE(table, ID string, query fiber.Map) ([]byte, error) {
	url := fmt.Sprintf("%s/rest/v1/%s?id=eq.%s", s.URL, table, ID)

	jsonData, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Prefer", "return=representation") // Ensures Supabase returns the updated record

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return respBody, nil
}
