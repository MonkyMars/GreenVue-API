package db_test

import (
	"context"
	"fmt"
	"greentrade-eu/lib"
	"testing"

	"github.com/google/uuid"
)

func TestMockListingRepository(t *testing.T) {
	repo := NewMockListingRepository()
	ctx := context.Background()

	// Test creating a listing
	sellerID := uuid.New().String() // Generate a UUID for the seller
	listing := lib.Listing{
		Title:       "Test Listing",
		Description: "This is a test listing for the mock repository",
		Category:    "electronics",
		Condition:   "good",
		Price:       10000,
		Location:    "Test City",
		SellerID:    sellerID, // Use the generated UUID string
	}

	createdListing, err := repo.CreateListing(ctx, listing)
	if err != nil {
		t.Fatalf("Failed to create listing: %v", err)
	}
	// Check if ID is assigned (non-empty string)
	if createdListing.ID == "" {
		t.Error("Created listing ID should not be empty")
	}
	if createdListing.Title != listing.Title {
		t.Errorf("Created listing title mismatch: got %s, want %s",
			createdListing.Title, listing.Title)
	}

	// Test getting a listing by ID
	id := createdListing.ID // ID is already a string
	fetchedListing, err := repo.GetListingByID(ctx, id)
	if err != nil {
		t.Fatalf("Failed to get listing by ID: %v", err)
	}
	if fetchedListing.ID != createdListing.ID {
		t.Errorf("Fetched listing ID mismatch: got %s, want %s", // Use %s for string ID
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
	secondListing := lib.Listing{
		Title:       "Test Listing 2",
		Description: "This is another test listing",
		Category:    "furniture",
		Condition:   "new",
		Price:       5000,
		SellerID:    "user2",
	}
	createdSecondListing, err := repo.CreateListing(ctx, secondListing)
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
	if electronicsListings[0].ID != createdListing.ID {
		t.Errorf("Incorrect listing returned for category 'electronics'")
	}

	furnitureListings, err := repo.GetListingsByCategory(ctx, "furniture")
	if err != nil {
		t.Fatalf("Failed to get listings by category: %v", err)
	}
	if len(furnitureListings) != 1 {
		t.Errorf("Expected 1 furniture listing, got %d", len(furnitureListings))
	}
	if furnitureListings[0].ID != createdSecondListing.ID {
		t.Errorf("Incorrect listing returned for category 'furniture'")
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
	expectedURLPrefix := "https://mock-storage.example.com/listings/test.jpg"
	if imgURL != expectedURLPrefix {
		t.Errorf("Image URL mismatch: got %s, want %s", imgURL, expectedURLPrefix)
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

	// Test creating a seller (using lib.User)
	sellerUser := lib.User{
		ID:       "seller1",
		Name:     "Test Seller User",
		Email:    "seller@example.com",
		Location: "Seller City",
		Bio:      "Sells things",
		// Add other relevant lib.User fields if needed
	}

	createdSeller, err := repo.CreateSeller(ctx, sellerUser)
	if err != nil {
		t.Fatalf("Failed to create seller: %v", err)
	}
	if createdSeller.ID != sellerUser.ID {
		t.Errorf("Created seller ID mismatch: got %s, want %s",
			createdSeller.ID, sellerUser.ID)
	}
	if createdSeller.Name != sellerUser.Name {
		t.Errorf("Created seller name mismatch: got %s, want %s",
			createdSeller.Name, sellerUser.Name)
	}
	if createdSeller.Email != sellerUser.Email {
		t.Errorf("Created seller email mismatch: got %s, want %s",
			createdSeller.Email, sellerUser.Email)
	}

	// Test creating a duplicate seller (by ID)
	_, err = repo.CreateSeller(ctx, sellerUser)
	if err != ErrAlreadyExists {
		t.Errorf("Expected ErrAlreadyExists for duplicate seller ID, got: %v", err)
	}

	// Test getting a seller by ID
	fetchedSeller, err := repo.GetSellerByID(ctx, sellerUser.ID)
	if err != nil {
		t.Fatalf("Failed to get seller by ID: %v", err)
	}
	if fetchedSeller.ID != sellerUser.ID {
		t.Errorf("Fetched seller ID mismatch: got %s, want %s",
			fetchedSeller.ID, sellerUser.ID)
	}
	if fetchedSeller.Name != sellerUser.Name {
		t.Errorf("Fetched seller name mismatch: got %s, want %s",
			fetchedSeller.Name, sellerUser.Name)
	}

	// Test getting a non-existent seller
	_, err = repo.GetSellerByID(ctx, "non-existent-id")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for non-existent seller ID, got: %v", err)
	}

	// Test updating a seller
	updates := map[string]interface{}{
		"name":     "Updated Seller Name",
		"location": "New Seller City",
		"bio":      "Updated bio",
		// Cannot update email via this method in the mock
	}
	err = repo.UpdateSeller(ctx, sellerUser.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update seller: %v", err)
	}

	// Verify updates were applied
	updatedSeller, err := repo.GetSellerByID(ctx, sellerUser.ID)
	if err != nil {
		t.Fatalf("Failed to get updated seller: %v", err)
	}
	if updatedSeller.Name != "Updated Seller Name" {
		t.Errorf("Updated name not applied: got %s, want %s",
			updatedSeller.Name, "Updated Seller Name")
	}
	if updatedSeller.Location != "New Seller City" {
		t.Errorf("Updated location not applied: got %s, want %s",
			updatedSeller.Location, "New Seller City")
	}
	if updatedSeller.Bio != "Updated bio" {
		t.Errorf("Updated bio not applied: got %s, want %s",
			updatedSeller.Bio, "Updated bio")
	}
	// Email should remain unchanged
	if updatedSeller.Email != sellerUser.Email {
		t.Errorf("Seller email should not have changed: got %s, want %s",
			updatedSeller.Email, sellerUser.Email)
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
	if sellers[0].ID != sellerUser.ID {
		t.Errorf("Incorrect seller returned in GetSellers")
	}
}

func TestMockUserRepository(t *testing.T) {
	repo := NewMockUserRepository()
	ctx := context.Background()

	// Test creating a user
	user := lib.User{
		ID:       "user1",
		Name:     "Test User",
		Email:    "test@example.com",
		Location: "Test Location",
		Bio:      "Test Bio",
	}

	createdUser, err := repo.CreateUser(ctx, user)
	if err != nil {
		t.Fatalf("Failed to create user: %v", err)
	}
	if createdUser.ID != user.ID {
		t.Errorf("Created user ID mismatch: got %s, want %s", createdUser.ID, user.ID)
	}
	if createdUser.Email != user.Email {
		t.Errorf("Created user email mismatch: got %s, want %s", createdUser.Email, user.Email)
	}
	if createdUser.CreatedAt == "" {
		t.Error("Created user should have a CreatedAt timestamp")
	}

	// Test creating user with empty ID
	_, err = repo.CreateUser(ctx, lib.User{Email: "no-id@example.com"})
	if err == nil {
		t.Error("Expected error when creating user with empty ID, got nil")
	}

	// Test creating user with empty Email
	_, err = repo.CreateUser(ctx, lib.User{ID: "no-email-user"})
	if err == nil {
		t.Error("Expected error when creating user with empty email, got nil")
	}

	// Test creating duplicate user (by ID)
	_, err = repo.CreateUser(ctx, user)
	if err != ErrAlreadyExists {
		t.Errorf("Expected ErrAlreadyExists for duplicate user ID, got: %v", err)
	}

	// Test creating duplicate user (by Email)
	duplicateEmailUser := lib.User{ID: "user2", Email: user.Email}
	_, err = repo.CreateUser(ctx, duplicateEmailUser)
	expectedErr := fmt.Errorf("user with email %s already exists", user.Email)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Expected error '%v' for duplicate email, got: %v", expectedErr, err)
	}

	// Test getting user by ID
	fetchedUserByID, err := repo.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get user by ID: %v", err)
	}
	if fetchedUserByID.ID != user.ID {
		t.Errorf("Fetched user ID mismatch: got %s, want %s", fetchedUserByID.ID, user.ID)
	}

	// Test getting non-existent user by ID
	_, err = repo.GetUserByID(ctx, "non-existent-id")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for non-existent user ID, got: %v", err)
	}

	// Test getting user by Email
	fetchedUserByEmail, err := repo.GetUserByEmail(ctx, user.Email)
	if err != nil {
		t.Fatalf("Failed to get user by email: %v", err)
	}
	if fetchedUserByEmail.ID != user.ID {
		t.Errorf("Fetched user by email returned wrong user: got ID %s, want %s", fetchedUserByEmail.ID, user.ID)
	}

	// Test getting non-existent user by Email
	_, err = repo.GetUserByEmail(ctx, "non-existent@example.com")
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for non-existent user email, got: %v", err)
	}

	// Test updating a user
	updates := map[string]interface{}{
		"name":     "Updated User Name",
		"location": "Updated Location",
		"email":    "updated@example.com", // Test email update
	}
	err = repo.UpdateUser(ctx, user.ID, updates)
	if err != nil {
		t.Fatalf("Failed to update user: %v", err)
	}

	// Verify updates
	updatedUser, err := repo.GetUserByID(ctx, user.ID)
	if err != nil {
		t.Fatalf("Failed to get updated user: %v", err)
	}
	if updatedUser.Name != "Updated User Name" {
		t.Errorf("Updated name not applied: got %s", updatedUser.Name)
	}
	if updatedUser.Location != "Updated Location" {
		t.Errorf("Updated location not applied: got %s", updatedUser.Location)
	}
	if updatedUser.Email != "updated@example.com" {
		t.Errorf("Updated email not applied: got %s", updatedUser.Email)
	}

	// Test updating user with email already in use
	// First, create another user
	otherUser := lib.User{ID: "otherUser", Email: "other@example.com"}
	_, err = repo.CreateUser(ctx, otherUser)
	if err != nil {
		t.Fatalf("Setup failed: could not create other user: %v", err)
	}
	// Try to update original user to other user's email
	emailConflictUpdates := map[string]interface{}{"email": otherUser.Email}
	err = repo.UpdateUser(ctx, user.ID, emailConflictUpdates)
	expectedErr = fmt.Errorf("email %s is already in use", otherUser.Email)
	if err == nil || err.Error() != expectedErr.Error() {
		t.Errorf("Expected error '%v' when updating to existing email, got: %v", expectedErr, err)
	}

	// Test updating non-existent user
	err = repo.UpdateUser(ctx, "non-existent-id", updates)
	if err != ErrNotFound {
		t.Errorf("Expected ErrNotFound for updating non-existent user, got: %v", err)
	}
}
