package db_test

import (
	"context"
	"errors"
	"greentrade-eu/internal/db"
	"sync"
)

// Define common errors for testing
var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
)

// MockListingRepository provides a simple in-memory repository for testing
type MockListingRepository struct {
	listings    map[string]db.Listing
	lastID      int64
	uploadedImg map[string][]byte
	mu          sync.RWMutex
}

// NewMockListingRepository creates a new mock repository
func NewMockListingRepository() *MockListingRepository {
	return &MockListingRepository{
		listings:    make(map[string]db.Listing),
		uploadedImg: make(map[string][]byte),
	}
}

// GetListings fetches listings with pagination
func (r *MockListingRepository) GetListings(ctx context.Context, limit, offset int) ([]db.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []db.Listing
	count := 0
	for _, listing := range r.listings {
		if count >= offset && len(result) < limit {
			result = append(result, listing)
		}
		count++
	}

	return result, nil
}

// GetListingByID fetches a listing by ID
func (r *MockListingRepository) GetListingByID(ctx context.Context, id string) (*db.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if listing, ok := r.listings[id]; ok {
		return &listing, nil
	}

	return nil, ErrNotFound
}

// GetListingsByCategory fetches listings by category
func (r *MockListingRepository) GetListingsByCategory(ctx context.Context, category string) ([]db.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []db.Listing
	for _, listing := range r.listings {
		if listing.Category == category {
			result = append(result, listing)
		}
	}

	return result, nil
}

// CreateListing creates a new listing
func (r *MockListingRepository) CreateListing(ctx context.Context, listing db.Listing) (*db.Listing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.lastID++
	listing.ID = r.lastID

	// Store by string ID for simplicity in mock
	r.listings[string(rune(listing.ID))] = listing

	return &listing, nil
}

// DeleteListing deletes a listing by ID
func (r *MockListingRepository) DeleteListing(ctx context.Context, id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.listings[id]; !ok {
		return ErrNotFound
	}

	delete(r.listings, id)
	return nil
}

// UploadImage uploads an image (mocked)
func (r *MockListingRepository) UploadImage(ctx context.Context, filename, bucket string, image []byte) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	key := bucket + "/" + filename
	r.uploadedImg[key] = image

	return "https://mock-storage.example.com/" + key, nil
}

// MockSellerRepository provides a simple in-memory repository for testing
type MockSellerRepository struct {
	sellers map[string]db.Seller
	mu      sync.RWMutex
}

// NewMockSellerRepository creates a new mock repository
func NewMockSellerRepository() *MockSellerRepository {
	return &MockSellerRepository{
		sellers: make(map[string]db.Seller),
	}
}

// GetSellers fetches all sellers
func (r *MockSellerRepository) GetSellers(ctx context.Context) ([]db.Seller, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []db.Seller
	for _, seller := range r.sellers {
		result = append(result, seller)
	}

	return result, nil
}

// GetSellerByID fetches a seller by ID
func (r *MockSellerRepository) GetSellerByID(ctx context.Context, id string) (*db.Seller, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if seller, ok := r.sellers[id]; ok {
		return &seller, nil
	}

	return nil, ErrNotFound
}

// CreateSeller creates a new seller
func (r *MockSellerRepository) CreateSeller(ctx context.Context, seller db.Seller) (*db.Seller, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.sellers[seller.ID]; exists {
		return nil, ErrAlreadyExists
	}

	r.sellers[seller.ID] = seller
	return &seller, nil
}

// UpdateSeller updates a seller
func (r *MockSellerRepository) UpdateSeller(ctx context.Context, id string, updates map[string]interface{}) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	seller, exists := r.sellers[id]
	if !exists {
		return ErrNotFound
	}

	// Apply updates (simplified for mock)
	if name, ok := updates["name"].(string); ok {
		seller.Name = name
	}
	if rating, ok := updates["rating"].(float32); ok {
		seller.Rating = rating
	}
	if verified, ok := updates["verified"].(bool); ok {
		seller.Verified = verified
	}

	r.sellers[id] = seller
	return nil
}
