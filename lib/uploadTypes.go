package lib

import (
	"time"

	"github.com/google/uuid"
)

type UpdateUser struct {
	ID       uuid.UUID `json:"id"`
	Name     string    `json:"name"`
	Bio      string    `json:"bio"`
	Location string    `json:"location"`
}

type Favorite struct {
	ID        uuid.UUID `json:"id,omitempty"`
	UserID    uuid.UUID `json:"user_id"`
	ListingID uuid.UUID `json:"listing_id"`
	CreatedAt time.Time `json:"created_at"`
}

type Review struct {
	Rating           int       `json:"rating"`
	UserID           uuid.UUID `json:"user_id"`
	SellerID         uuid.UUID `json:"seller_id"`
	Title            string    `json:"title"`
	Content          string    `json:"content"`
	VerifiedPurchase bool      `json:"verified_purchase"`
}

type Listing struct {
	ID            uuid.UUID `json:"id,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Condition     string    `json:"condition"`
	Price         float64   `json:"price"`
	EcoScore      float32   `json:"ecoScore"`
	EcoAttributes []string  `json:"ecoAttributes"`
	Negotiable    bool      `json:"negotiable"`
	Title         string    `json:"title"`
	ImageUrl      []string  `json:"imageUrl"`
	SellerID      uuid.UUID `json:"seller_id"`
}

type Message struct {
	ConversationID uuid.UUID `json:"conversation_id"`
	SenderID       uuid.UUID `json:"sender_id"`
	Content        string    `json:"content"`
}

type Bid struct {
	ListingID uuid.UUID `json:"listing_id"`
	UserID    uuid.UUID `json:"user_id"`
	Price     float64   `json:"price"`
}
