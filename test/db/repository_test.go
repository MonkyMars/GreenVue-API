package db_test

import (
	"context"
	"greentrade-eu/internal/db"
	"testing"
)

func TestMockListingRepository(t *testing.T) {
	repo := NewMockListingRepository()
	ctx := context.Background()

	// Test creating a listing
	listing := db.Listing{
		Title:       "Test Listing",
		Description: "This is a test listing for the mock repository",
		Category:    "electronics",
		Condition:   "good",
		Price:       10000,
		Location:    "Test City",
		Seller: db.Seller{
			ID:   "seller1",
			Name: "Test Seller",
		},
	}

	createdListing, err := repo.CreateListing(ctx, listing)
	if err != nil {
		t.Fatalf("Failed to create listing: %v", err)
	}
	if createdListing.ID <= 0 {
		t.Error("Created listing ID should be positive")
	}
	if createdListing.Title != listing.Title {
		t.Errorf("Created listing title mismatch: got %s, want %s",
			createdListing.Title, listing.Title)
	}

	// Test getting a listing by ID
	id := string(rune(createdListing.ID))
	fetchedListing, err := repo.GetListingByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get listing by ID: %v", err)
	}
	if fetchedListing.ID != createdListing.ID {
		t.Errorf("Fetched listing ID mismatch: got %d, want %d",
			fetchedListing.ID, createdListing.ID)
	}

	// Test getting a non-existent listing
	_, err = repo.GetListingByID(ctx, "non-existent-id")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for non-existent ID, got: %v", err)
	}

	// Test getting all listings
	listings, err := repo.GetListings(ctx, 10, 0)
	if err != nil {
		t.Fatalf("Failed to get listings: %v", err)
	}
	if len(listings) != 1 {
		t.Errorf("Expected 1 listing, got %d", len(listings))
	}

	// Create a second listing with a different category
	secondListing := db.Listing{
		Title:       "Test Listing 2",
		Description: "This is another test listing",
		Category:    "furniture",
		Condition:   "new",
		Price:       5000,
	}
	_, err = repo.CreateListing(ctx, secondListing)
	if err != nil {
		t.Fatalf("Failed to create second listing: %v", err)
	}

	// Test getting by category
	electronicsListings, err := repo.GetListingsByCategory(ctx, "electronics")
	if err != nil {
		t.Fatalf("Failed to get listings by category: %v", err)
	}
	if len(electronicsListings) != 1 {
		t.Errorf("Expected 1 electronics listing, got %d", len(electronicsListings))
	}

	furnitureListings, err := repo.GetListingsByCategory(ctx, "furniture")
	if err != nil {
		t.Fatalf("Failed to get listings by category: %v", err)
	}
	if len(furnitureListings) != 1 {
		t.Errorf("Expected 1 furniture listing, got %d", len(furnitureListings))
	}

	// Test upload image
	imgData := []byte("fake image data")
	imgURL, err := repo.UploadImage(ctx, "test.jpg", "listings", imgData)
	if err != nil {
		t.Fatalf("Failed to upload image: %v", err)
	}
	if imgURL == "" {
		t.Error("Image URL should not be empty")
	}

	// Test deleting a listing
	err = repo.DeleteListing(ctx, id)
	if err != nil {
		t.Fatalf("Failed to delete listing: %v", err)
	}

	// Verify listing is deleted
	_, err = repo.GetListingByID(ctx, id)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound after deletion, got: %v", err)
	}

	// Test deleting a non-existent listing
	err = repo.DeleteListing(ctx, "non-existent-id")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for deleting non-existent listing, got: %v", err)
	}
}

func TestMockSellerRepository(t *testing.T) {
	repo := NewMockSellerRepository()
	ctx := context.Background()

	// Test creating a seller
	seller := db.Seller{
		ID:       "seller1",
		Name:     "Test Seller",
		Rating:   4.5,
		Verified: true,
	}

	createdSeller, err := repo.CreateSeller(ctx, seller)
	if err != nil {
		t.Fatalf("Failed to create seller: %v", err)
	}
	if createdSeller.ID != seller.ID {
		t.Errorf("Created seller ID mismatch: got %s, want %s",
			createdSeller.ID, seller.ID)
	}
	if createdSeller.Name != seller.Name {
		t.Errorf("Created seller name mismatch: got %s, want %s",
			createdSeller.Name, seller.Name)
	}

	// Test creating a duplicate seller
	_, err = repo.CreateSeller(ctx, seller)
	if err != ErrAlreadyExists {
		t.Errorf("Expected ErrAlreadyExists for duplicate seller, got: %v", err)
	}

	// Test getting a seller by ID
	fetchedSeller, err := repo.GetSellerByID(ctx, seller.ID)
	if err != nil {
		t.Fatalf("Failed to get seller by ID: %v", err)
	}
	if fetchedSeller.ID != seller.ID {
		t.Errorf("Fetched seller ID mismatch: got %s, want %s",
			fetchedSeller.ID, seller.ID)
	}

	// Test getting a non-existent seller
	_, err = repo.GetSellerByID(ctx, "non-existent-id")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for non-existent seller ID, got: %v", err)
	}

	// Test updating a seller
	updates := map[string]interface{}{
		"name":     "Updated Seller Name",
		"rating":   float32(5.0),
		"verified": false,
	}
	err = repo.UpdateSeller(ctx, seller.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update seller: %v", err)
	}

	// Verify updates were applied
	updatedSeller, err := repo.GetSellerByID(ctx, seller.ID)
	if err != nil {
		t.Fatalf("Failed to get updated seller: %v", err)
	}
	if updatedSeller.Name != "Updated Seller Name" {
		t.Errorf("Updated name not applied: got %s, want %s",
			updatedSeller.Name, "Updated Seller Name")
	}
	if updatedSeller.Rating != 5.0 {
		t.Errorf("Updated rating not applied: got %f, want %f",
			updatedSeller.Rating, 5.0)
	}
	if updatedSeller.Verified != false {
		t.Errorf("Updated verified status not applied: got %v, want %v",
			updatedSeller.Verified, false)
	}

	// Test updating a non-existent seller
	err = repo.UpdateSeller(ctx, "non-existent-id", updates)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for updating non-existent seller, got: %v", err)
	}

	// Test getting all sellers
	sellers, err := repo.GetSellers(ctx)
	if err != nil {
		t.Fatalf("Failed to get all sellers: %v", err)
	}
	if len(sellers) != 1 {
		t.Errorf("Expected 1 seller, got %d", len(sellers))
	}
}
