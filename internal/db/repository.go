package db

import (
	"context"
)

// ListingRepository defines operations for the Listing entity
type ListingRepository interface {
	GetListings(ctx context.Context, limit, offset int) ([]Listing, error)
	GetListingByID(ctx context.Context, id string) (*Listing, error)
	GetListingsByCategory(ctx context.Context, category string) ([]Listing, error)
	CreateListing(ctx context.Context, listing Listing) (*Listing, error)
	DeleteListing(ctx context.Context, id string) error
	UploadImage(ctx context.Context, filename, bucket string, image []byte) (string, error)
}

// SellerRepository defines operations for the Seller entity
type SellerRepository interface {
	GetSellers(ctx context.Context) ([]Seller, error)
	GetSellerByID(ctx context.Context, id string) (*Seller, error)
	CreateSeller(ctx context.Context, seller Seller) (*Seller, error)
	UpdateSeller(ctx context.Context, id string, updates map[string]interface{}) error
}

// UserRepository defines operations for the User entity
type UserRepository interface {
	GetUserByID(ctx context.Context, id string) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	CreateUser(ctx context.Context, user User) (*User, error)
	UpdateUser(ctx context.Context, id string, updates map[string]interface{}) error
}
