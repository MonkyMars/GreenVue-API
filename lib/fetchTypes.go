package lib

import (
	"time"

	"github.com/google/uuid"
)

type Location struct {
	Country   string  `json:"country"`
	City      string  `json:"city"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

type User struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email"`
	Name          string    `json:"name"`
	Location      Location  `json:"location"`
	Bio           string    `json:"bio"`
	CreatedAt     time.Time `json:"created_at"`
	Rating        float32   `json:"rating"`
	Verified      bool      `json:"verified"`
	EmailVerified bool      `json:"email_verified"`
	Picture       string    `json:"picture"`
	Provider      string    `json:"provider"`
}

type PublicUser struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Location  Location  `json:"location"`
	Bio       string    `json:"bio"`
	CreatedAt time.Time `json:"created_at"`
	Rating    float32   `json:"rating"`
	Verified  bool      `json:"verified"`
	Picture   string    `json:"picture"`
}

type FetchedListing struct {
	ID            uuid.UUID `json:"id"`
	CreatedAt     time.Time `json:"created_at"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Condition     string    `json:"condition"`
	Price         float64   `json:"price"`
	Location      Location  `json:"location"`
	EcoScore      float32   `json:"eco_score"`
	EcoAttributes []string  `json:"eco_attributes"`
	Negotiable    bool      `json:"negotiable"`
	Title         string    `json:"title"`
	ImageUrl      []string  `json:"image_urls"`

	SellerID        uuid.UUID    `json:"seller_id"`
	SellerUsername  string       `json:"seller_username"`
	SellerBio       string       `json:"seller_bio"`
	SellerCreatedAt time.Time    `json:"seller_created_at"`
	SellerRating    float32      `json:"seller_rating"`
	SellerVerified  bool         `json:"seller_verified"`
	Bids            []FetchedBid `json:"bids"`
}

type FetchedFavorite struct {
	ID          uuid.UUID `json:"id"`
	UserID      uuid.UUID `json:"user_id"`
	ListingID   uuid.UUID `json:"listing_id"`
	FavoritedAt time.Time `json:"favorited_at"`

	CreatedAt     time.Time `json:"created_at"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Condition     string    `json:"condition"`
	Price         float64   `json:"price"`
	Location      Location  `json:"location"`
	EcoScore      float32   `json:"ecoScore"`
	EcoAttributes []string  `json:"eco_attributes"`
	Negotiable    bool      `json:"negotiable"`
	Title         string    `json:"title"`
	ImageUrl      []string  `json:"image_urls"`

	SellerID        uuid.UUID `json:"seller_id"`
	SellerUsername  string    `json:"seller_username"`
	SellerBio       string    `json:"seller_bio"`
	SellerCreatedAt time.Time `json:"seller_created_at"`
	SellerRating    float32   `json:"seller_rating"`
	SellerVerified  bool      `json:"seller_verified"`
}

type FetchedReview struct {
	ID               uuid.UUID `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	Rating           int       `json:"rating"`
	UserID           uuid.UUID `json:"user_id"`
	UserName         string    `json:"user_name"`
	SellerID         uuid.UUID `json:"seller_id"`
	Title            string    `json:"title"`
	Content          string    `json:"content"`
	HelpfulCount     int       `json:"helpful_count"`
	VerifiedPurchase bool      `json:"verified_purchase"`
}

type FetchedBid struct {
	// Bid data
	ID        uuid.UUID `json:"id"`
	ListingID uuid.UUID `json:"listing_id"`
	UserID    uuid.UUID `json:"user_id"`
	Price     float64   `json:"price"`
	CreatedAt time.Time `json:"created_at"`

	// User data
	UserName    string `json:"user_name"`
	UserPicture string `json:"user_picture"`
}
