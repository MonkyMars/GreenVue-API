package db

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
)

type Seller struct { // this struct is used in the supabase database: Seller.
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Rating   int    `json:"rating"`
	Verified bool   `json:"verified"`
}

type Listing struct { // this struct is used in the supabase database: Listing.
	ID            int      `json:"id,omitempty"`
	CreatedAt     string   `json:"created_at,omitempty"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Condition     string   `json:"condition"`
	Price         int      `json:"price"`
	Location      string   `json:"location"`
	EcoAttributes []string `json:"ecoAttributes"`
	Negotiable    bool     `json:"negotiable"`
	Title         string   `json:"title"`
	SellerID      int      `json:"seller_id"`
}

type ExpectedListing struct { // this struct is expected from the frontend.
	ID            int      `json:"id,omitempty"`
	CreatedAt     string   `json:"created_at,omitempty"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Condition     string   `json:"condition"`
	Price         int      `json:"price"`
	Location      string   `json:"location"`
	EcoAttributes []string `json:"ecoAttributes"`
	Negotiable    bool     `json:"negotiable"`
	Title         string   `json:"title"`
	Seller        Seller   `json:"seller"`
}

type SupabaseClient struct {
	URL       string
	APIKey    string
	AuthToken string
}

func NewSupabaseClient() *SupabaseClient {
	url := os.Getenv("SUPABASE_URL")
	apiKey := os.Getenv("SUPABASE_ANON")

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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
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

func (s *SupabaseClient) IsSellerInDB(seller Seller) (bool, int, error) {
	url := fmt.Sprintf("%s/rest/v1/sellers?id=eq.%d", s.URL, seller.ID)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, 0, err
	}

	req.Header.Add("apikey", s.APIKey)
	req.Header.Add("Authorization", "Bearer "+s.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, 0, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, 0, err
	}

	var sellers []Seller
	if err := json.Unmarshal(body, &sellers); err != nil {
		return false, 0, err
	}

	// Check if the array has elements before accessing index 0
	if len(sellers) > 0 {
		return true, sellers[0].ID, nil
	}

	// If sellers array is empty, return false with ID 0
	return false, 0, nil
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
