package lib

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type User struct {
	ID           string  `json:"id"`
	Email        string  `json:"email"`
	Name         string  `json:"name,omitempty"`
	Location     string  `json:"location,omitempty"`
	Bio          string  `json:"bio,omitempty"`
	CreatedAt    string  `json:"created_at,omitempty"`
	Rating       float32 `json:"rating,omitempty"`
	Verified     bool    `json:"verified,omitempty"`
	LastSignInAt string  `json:"last_sign_in_at,omitempty"`
}

type PublicUser struct {
	ID        string  `json:"id"`
	Name      string  `json:"name"`
	Location  string  `json:"location"`
	Bio       string  `json:"bio"`
	CreatedAt string  `json:"created_at"`
	Rating    float32 `json:"rating"`
	Verified  bool    `json:"verified"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
}

type UpdateUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Bio      string `json:"bio"`
	Location string `json:"location"`
}

type Message struct {
	ConversationID string `json:"conversation_id"`
	SenderID       string `json:"sender_id"`
	Content        string `json:"content"`
}

type Review struct {
	Rating           int    `json:"rating"`
	UserID           string `json:"user_id"`
	SellerID         string `json:"seller_id"`
	Title            string `json:"title"`
	Content          string `json:"content"`
	VerifiedPurchase bool   `json:"verified_purchase"`
}

type FetchedReview struct {
	ID               uuid.UUID  `json:"id,omitempty"`
	CreatedAt        *time.Time `json:"created_at,omitempty"`
	Rating           int        `json:"rating"`
	UserID           string     `json:"user_id"`
	UserName         *string    `json:"user_name,omitempty"`
	SellerID         string     `json:"seller_id"`
	Title            string     `json:"title"`
	Content          string     `json:"content"`
	HelpfulCount     int        `json:"helpful_count"`
	VerifiedPurchase bool       `json:"verified_purchase"`
}

type Favorite struct {
	ID        string `json:"id,omitempty"`
	UserID    string `json:"user_id"`
	ListingID string `json:"listing_id"`
	CreatedAt string `json:"created_at,omitempty"`
}

type FetchedFavorite struct {
	ID          string `json:"id,omitempty"`
	UserID      string `json:"user_id"`
	ListingID   string `json:"listing_id"`
	FavoritedAt string `json:"favorited_at,omitempty"`

	CreatedAt     string   `json:"created_at,omitempty"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Condition     string   `json:"condition"`
	Price         float64  `json:"price"`
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

// Moved from db package
type Listing struct {
	ID            string   `json:"id,omitempty"`
	CreatedAt     string   `json:"created_at,omitempty"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Condition     string   `json:"condition"`
	Price         float64  `json:"price"`
	EcoScore      float32  `json:"ecoScore"`
	EcoAttributes []string `json:"ecoAttributes"`
	Negotiable    bool     `json:"negotiable"`
	Title         string   `json:"title"`
	ImageUrl      []string `json:"imageUrl"`
	SellerID      string   `json:"seller_id"`
}

// Moved from db package
type FetchedListing struct {
	ID            string   `json:"id,omitempty"`
	CreatedAt     string   `json:"created_at,omitempty"`
	Description   string   `json:"description"`
	Category      string   `json:"category"`
	Condition     string   `json:"condition"`
	Price         float64  `json:"price"`
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

// Generic query parameters for data access
type QueryParams struct {
	Table     string         `json:"-"`
	ID        string         `json:"-"`
	Filter    string         `json:"-"`
	Limit     int            `json:"-"`
	Offset    int            `json:"-"`
	OrderBy   string         `json:"-"`
	Direction string         `json:"-"`
	Data      map[string]any `json:"-"`
	FiberMap  fiber.Map      `json:"-"`
}
