package db_test

import (
	"context"
	"errors"
	"fmt"
	"greenvue-eu/lib"
	"sync"
	"time"
)

// Define common errors for testing
var (
	ErrNotFound      = errors.New("entity not found")
	ErrAlreadyExists = errors.New("entity already exists")
)

// MockListingRepository provides a simple in-memory repository for testing
type MockListingRepository struct {
	listings    map[string]lib.Listing
	lastID      int64
	uploadedImg map[string][]byte
	mu          sync.RWMutex
}

// NewMockListingRepository creates a new mock repository
func NewMockListingRepository() *MockListingRepository {
	return &MockListingRepository{
		listings:    make(map[string]lib.Listing),
		uploadedImg: make(map[string][]byte),
	}
}

// GetListings fetches listings with pagination
func (r *MockListingRepository) GetListings(ctx context.Context, limit, offset int) ([]lib.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []lib.Listing
	count := 0
	// Note: Iteration order over maps is not guaranteed. For stable pagination, consider sorting keys.
	for _, listing := range r.listings {
		if count >= offset && len(result) < limit {
			result = append(result, listing)
		}
		count++
	}

	return result, nil
}

// GetListingByID fetches a listing by ID
func (r *MockListingRepository) GetListingByID(ctx context.Context, id string) (*lib.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if listing, ok := r.listings[id]; ok {
		return &listing, nil
	}

	return nil, ErrNotFound
}

// GetListingsByCategory fetches listings by category
func (r *MockListingRepository) GetListingsByCategory(ctx context.Context, category string) ([]lib.Listing, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []lib.Listing
	for _, listing := range r.listings {
		if listing.Category == category {
			result = append(result, listing)
		}
	}

	return result, nil
}

// CreateListing creates a new listing
func (r *MockListingRepository) CreateListing(ctx context.Context, listing lib.Listing) (*lib.Listing, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Simulate auto-increment ID if not provided
	if listing.ID == "" {
		r.lastID++
		listing.ID = fmt.Sprintf("%d", r.lastID) // Use fmt.Sprintf for string ID
	} else {
		// Check if ID already exists
		if _, exists := r.listings[listing.ID]; exists {
			return nil, ErrAlreadyExists // Or handle as appropriate
		}
	}

	// Simulate created_at timestamp
	if listing.CreatedAt == "" {
		listing.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	r.listings[listing.ID] = listing
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

	// Return a plausible mock URL
	return fmt.Sprintf("https://mock-storage.example.com/%s", key), nil
}

// MockSellerRepository provides a simple in-memory repository for testing
type MockSellerRepository struct {
	sellers map[string]lib.User
	mu      sync.RWMutex
}

// NewMockSellerRepository creates a new mock repository
func NewMockSellerRepository() *MockSellerRepository {
	return &MockSellerRepository{
		sellers: make(map[string]lib.User),
	}
}

// GetSellers fetches all sellers
func (r *MockSellerRepository) GetSellers(ctx context.Context) ([]lib.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []lib.User
	for _, seller := range r.sellers {
		// Add filtering logic if sellers need specific attributes (e.g., is_seller flag)
		result = append(result, seller)
	}

	return result, nil
}

// GetSellerByID fetches a seller by ID
func (r *MockSellerRepository) GetSellerByID(ctx context.Context, id string) (*lib.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if seller, ok := r.sellers[id]; ok {
		return &seller, nil
	}

	return nil, ErrNotFound
}

// CreateSeller creates a new seller
func (r *MockSellerRepository) CreateSeller(ctx context.Context, seller lib.User) (*lib.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if seller.ID == "" {
		// Generate an ID if needed, e.g., using UUID or incrementing counter
		return nil, errors.New("seller ID cannot be empty")
	}

	if _, exists := r.sellers[seller.ID]; exists {
		return nil, ErrAlreadyExists
	}

	// Simulate created_at timestamp
	if seller.CreatedAt == "" {
		seller.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	r.sellers[seller.ID] = seller
	return &seller, nil
}

// UpdateSeller updates a seller
func (r *MockSellerRepository) UpdateSeller(ctx context.Context, id string, updates map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	seller, exists := r.sellers[id]
	if !exists {
		return ErrNotFound
	}

	// Apply updates (simplified for mock, add more fields as needed)
	if name, ok := updates["name"].(string); ok {
		seller.Name = name
	}
	if location, ok := updates["location"].(string); ok {
		seller.Location = location
	}
	if bio, ok := updates["bio"].(string); ok {
		seller.Bio = bio
	}
	// Add other updatable fields based on lib.User needs

	r.sellers[id] = seller
	return nil
}

// --- MockUserRepository ---

// MockUserRepository provides a simple in-memory repository for testing User operations
type MockUserRepository struct {
	users map[string]lib.User
	mu    sync.RWMutex
}

// NewMockUserRepository creates a new mock repository for users
func NewMockUserRepository() *MockUserRepository {
	return &MockUserRepository{
		users: make(map[string]lib.User),
	}
}

// GetUserByID fetches a user by ID
func (r *MockUserRepository) GetUserByID(ctx context.Context, id string) (*lib.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if user, ok := r.users[id]; ok {
		return &user, nil
	}

	return nil, ErrNotFound
}

// GetUserByEmail fetches a user by email
func (r *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*lib.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, user := range r.users {
		if user.Email == email {
			return &user, nil
		}
	}

	return nil, ErrNotFound
}

// CreateUser creates a new user
func (r *MockUserRepository) CreateUser(ctx context.Context, user lib.User) (*lib.User, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if user.ID == "" {
		// Generate an ID if needed, e.g., using UUID
		return nil, errors.New("user ID cannot be empty")
	}
	if user.Email == "" {
		return nil, errors.New("user email cannot be empty")
	}

	// Check for existing ID or Email
	if _, exists := r.users[user.ID]; exists {
		return nil, ErrAlreadyExists
	}
	for _, existingUser := range r.users {
		if existingUser.Email == user.Email {
			return nil, fmt.Errorf("user with email %s already exists", user.Email)
		}
	}

	// Simulate created_at timestamp
	if user.CreatedAt == "" {
		user.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	r.users[user.ID] = user
	return &user, nil
}

// UpdateUser updates a user by ID
func (r *MockUserRepository) UpdateUser(ctx context.Context, id string, updates map[string]any) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	user, exists := r.users[id]
	if !exists {
		return ErrNotFound
	}

	// Apply updates (simplified for mock)
	if name, ok := updates["name"].(string); ok {
		user.Name = name
	}
	if location, ok := updates["location"].(string); ok {
		user.Location = location
	}
	if bio, ok := updates["bio"].(string); ok {
		user.Bio = bio
	}

	// Note: Email updates might need special handling (checking for uniqueness)
	if email, ok := updates["email"].(string); ok {
		// Check if the new email is already taken by another user
		for otherID, otherUser := range r.users {
			if otherID != id && otherUser.Email == email {
				return fmt.Errorf("email %s is already in use", email)
			}
		}
		user.Email = email
	}

	r.users[id] = user
	return nil
}
