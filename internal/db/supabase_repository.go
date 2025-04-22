package db

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// SupabaseListingRepository implements ListingRepository
type SupabaseListingRepository struct {
	client *SupabaseClient
}

// NewSupabaseListingRepository creates a new Supabase repository for listings
func NewSupabaseListingRepository(client *SupabaseClient) *SupabaseListingRepository {
	return &SupabaseListingRepository{
		client: client,
	}
}

// GetListings fetches listings with pagination
func (r *SupabaseListingRepository) GetListings(ctx context.Context, limit, offset int) ([]Listing, error) {
	query := fmt.Sprintf("select=*&limit=%d&offset=%d&order=created_at.desc", limit, offset)
	resp, err := r.client.Query("listings", query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch listings: %w", err)
	}

	var listings []Listing
	if err := json.Unmarshal(resp, &listings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal listings: %w", err)
	}

	return listings, nil
}

// GetListingByID fetches a listing by ID
func (r *SupabaseListingRepository) GetListingByID(ctx context.Context, id string) (*Listing, error) {
	query := fmt.Sprintf("id=eq.%s&select=*", id)
	resp, err := r.client.Query("listings", query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch listing by ID: %w", err)
	}

	var listings []Listing
	if err := json.Unmarshal(resp, &listings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal listing: %w", err)
	}

	if len(listings) == 0 {
		return nil, fmt.Errorf("listing not found")
	}

	return &listings[0], nil
}

// GetListingsByCategory fetches listings by category
func (r *SupabaseListingRepository) GetListingsByCategory(ctx context.Context, category string) ([]Listing, error) {
	query := fmt.Sprintf("category=eq.%s&select=*&order=created_at.desc", category)
	resp, err := r.client.Query("listings", query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch listings by category: %w", err)
	}

	var listings []Listing
	if err := json.Unmarshal(resp, &listings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal listings: %w", err)
	}

	return listings, nil
}

// CreateListing creates a new listing
func (r *SupabaseListingRepository) CreateListing(ctx context.Context, listing Listing) (*Listing, error) {
	resp, err := r.client.POST("listings", listing)
	if err != nil {
		return nil, fmt.Errorf("failed to create listing: %w", err)
	}

	var createdListing []Listing
	if err := json.Unmarshal(resp, &createdListing); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created listing: %w", err)
	}

	if len(createdListing) == 0 {
		return nil, fmt.Errorf("failed to create listing: no record returned")
	}

	return &createdListing[0], nil
}

// DeleteListing deletes a listing by ID
func (r *SupabaseListingRepository) DeleteListing(ctx context.Context, id string) error {
	_, err := r.client.DELETE("listings", id)
	if err != nil {
		return fmt.Errorf("failed to delete listing: %w", err)
	}

	return nil
}

// UploadImage uploads an image to Supabase storage
func (r *SupabaseListingRepository) UploadImage(ctx context.Context, filename, bucket string, image []byte) (string, error) {
	resp, err := r.client.UploadImage(filename, bucket, image)
	if err != nil {
		return "", fmt.Errorf("failed to upload image: %w", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(resp, &result); err != nil {
		return "", fmt.Errorf("failed to parse upload response: %w", err)
	}

	// Construct image URL
	imageURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", r.client.URL, bucket, filename)
	return imageURL, nil
}

// SupabaseSellerRepository implements SellerRepository
type SupabaseSellerRepository struct {
	client *SupabaseClient
}

// NewSupabaseSellerRepository creates a new Supabase repository for sellers
func NewSupabaseSellerRepository(client *SupabaseClient) *SupabaseSellerRepository {
	return &SupabaseSellerRepository{
		client: client,
	}
}

// GetSellers fetches all sellers
func (r *SupabaseSellerRepository) GetSellers(ctx context.Context) ([]User, error) {
	resp, err := r.client.Query("sellers", "select=*")
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sellers: %w", err)
	}

	var sellers []User
	if err := json.Unmarshal(resp, &sellers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sellers: %w", err)
	}

	return sellers, nil
}

// GetSellerByID fetches a seller by ID
func (r *SupabaseSellerRepository) GetSellerByID(ctx context.Context, id string) (*User, error) {
	query := fmt.Sprintf("id=eq.%s&select=*", id)
	resp, err := r.client.Query("sellers", query)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch seller by ID: %w", err)
	}

	var sellers []User
	if err := json.Unmarshal(resp, &sellers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal seller: %w", err)
	}

	if len(sellers) == 0 {
		return nil, fmt.Errorf("seller not found")
	}

	return &sellers[0], nil
}

// CreateSeller creates a new seller
func (r *SupabaseSellerRepository) CreateSeller(ctx context.Context, seller User) (*User, error) {
	jsonData, err := json.Marshal(seller)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal seller: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/%s", r.client.URL, "sellers")
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("apikey", r.client.APIKey)
	req.Header.Add("Authorization", "Bearer "+r.client.APIKey)

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

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("supabase error (%d): %s", resp.StatusCode, string(body))
	}

	var createdSeller User
	if err := json.Unmarshal(body, &createdSeller); err != nil {
		return nil, fmt.Errorf("failed to unmarshal created seller: %w", err)
	}

	return &createdSeller, nil
}

// UpdateSeller updates a seller by ID
func (r *SupabaseSellerRepository) UpdateSeller(ctx context.Context, id string, updates map[string]interface{}) error {
	jsonData, err := json.Marshal(updates)
	if err != nil {
		return fmt.Errorf("failed to marshal updates: %w", err)
	}

	url := fmt.Sprintf("%s/rest/v1/%s?id=eq.%s", r.client.URL, "sellers", id)
	req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Add("apikey", r.client.APIKey)
	req.Header.Add("Authorization", "Bearer "+r.client.APIKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("supabase error (%d): %s", resp.StatusCode, string(body))
	}

	return nil
}
