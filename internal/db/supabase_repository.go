package db

import (
	"context"
	"encoding/json"
	"fmt"
	"greenvue/lib"
)

// SupabaseListingRepository implements ListingRepository using the new Repository interface
type SupabaseListingRepository struct {
	repo *SupabaseRepository
}

// NewSupabaseListingRepository creates a new repository for managing listings
func NewSupabaseListingRepository(client *SupabaseClient) *SupabaseListingRepository {
	// If no client is provided, use the global client
	if client == nil {
		client = GetGlobalClient()
	}
	return &SupabaseListingRepository{
		repo: NewSupabaseRepository(client),
	}
}

// GetListings fetches listings with pagination
func (r *SupabaseListingRepository) GetListings(ctx context.Context, limit, offset int) ([]lib.Listing, error) {
	params := lib.QueryParams{
		Table:     "listings",
		Limit:     limit,
		Offset:    offset,
		OrderBy:   "created_at",
		Direction: "desc",
	}

	resp, err := r.repo.Get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch listings: %w", err)
	}

	var listings []lib.Listing
	if err := json.Unmarshal(resp, &listings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal listings: %w", err)
	}

	return listings, nil
}

// GetListingByID fetches a listing by ID
func (r *SupabaseListingRepository) GetListingByID(ctx context.Context, id string) (*lib.Listing, error) {
	resp, err := r.repo.GetByID(ctx, "listings", id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch listing by ID: %w", err)
	}

	var listings []lib.Listing
	if err := json.Unmarshal(resp, &listings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal listing: %w", err)
	}

	if len(listings) == 0 {
		return nil, fmt.Errorf("listing not found")
	}

	return &listings[0], nil
}

// GetListingsByCategory fetches listings by category
func (r *SupabaseListingRepository) GetListingsByCategory(ctx context.Context, category string) ([]lib.Listing, error) {
	params := lib.QueryParams{
		Table:     "listings",
		Filter:    fmt.Sprintf("category=eq.%s", category),
		OrderBy:   "created_at",
		Direction: "desc",
	}

	resp, err := r.repo.Get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch listings by category: %w", err)
	}

	var listings []lib.Listing
	if err := json.Unmarshal(resp, &listings); err != nil {
		return nil, fmt.Errorf("failed to unmarshal listings: %w", err)
	}

	return listings, nil
}

// CreateListing creates a new listing
func (r *SupabaseListingRepository) CreateListing(ctx context.Context, listing lib.Listing) (*lib.Listing, error) {
	resp, err := r.repo.Create(ctx, "listings", listing)
	if err != nil {
		return nil, fmt.Errorf("failed to create listing: %w", err)
	}

	var createdListing []lib.Listing
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
	return r.repo.Delete(ctx, "listings", id)
}

// UploadImage uploads an image to Supabase storage
func (r *SupabaseListingRepository) UploadImage(ctx context.Context, filename, bucket string, image []byte) (string, error) {
	return r.repo.UploadImage(ctx, filename, bucket, image)
}

// SupabaseSellerRepository implements SellerRepository using the new Repository interface
type SupabaseSellerRepository struct {
	repo *SupabaseRepository
}

// NewSupabaseSellerRepository creates a new repository for managing sellers
func NewSupabaseSellerRepository(client *SupabaseClient) *SupabaseSellerRepository {
	// If no client is provided, use the global client
	if client == nil {
		client = GetGlobalClient()
	}
	return &SupabaseSellerRepository{
		repo: NewSupabaseRepository(client),
	}
}

// GetSellers fetches all sellers
func (r *SupabaseSellerRepository) GetSellers(ctx context.Context) ([]lib.User, error) {
	params := lib.QueryParams{
		Table: "sellers",
	}

	resp, err := r.repo.Get(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch sellers: %w", err)
	}

	var sellers []lib.User
	if err := json.Unmarshal(resp, &sellers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal sellers: %w", err)
	}

	return sellers, nil
}

// GetSellerByID fetches a seller by ID
func (r *SupabaseSellerRepository) GetSellerByID(ctx context.Context, id string) (*lib.User, error) {
	resp, err := r.repo.GetByID(ctx, "sellers", id)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch seller by ID: %w", err)
	}

	var sellers []lib.User
	if err := json.Unmarshal(resp, &sellers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal seller: %w", err)
	}

	if len(sellers) == 0 {
		return nil, fmt.Errorf("seller not found")
	}

	return &sellers[0], nil
}

// CreateSeller creates a new seller
func (r *SupabaseSellerRepository) CreateSeller(ctx context.Context, seller lib.User) (*lib.User, error) {
	resp, err := r.repo.Create(ctx, "sellers", seller)
	if err != nil {
		return nil, fmt.Errorf("failed to create seller: %w", err)
	}

	var createdSeller lib.User
	if err := json.Unmarshal(resp, &createdSeller); err != nil {
		// Try to unmarshal as array if single object fails
		var sellerArray []lib.User
		if err := json.Unmarshal(resp, &sellerArray); err != nil {
			return nil, fmt.Errorf("failed to unmarshal created seller: %w", err)
		}
		if len(sellerArray) == 0 {
			return nil, fmt.Errorf("no seller returned after creation")
		}
		return &sellerArray[0], nil
	}

	return &createdSeller, nil
}

// UpdateSeller updates a seller by ID
func (r *SupabaseSellerRepository) UpdateSeller(ctx context.Context, id string, updates map[string]any) error {
	_, err := r.repo.Update(ctx, "sellers", id, updates)
	return err
}
