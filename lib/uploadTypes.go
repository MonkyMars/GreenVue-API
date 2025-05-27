package lib

import (
	"github.com/google/uuid"
)

type UpdateUser struct {
	ID   uuid.UUID `json:"id"`
	Name string    `json:"name"`
	Bio  string    `json:"bio"`
}

type Favorite struct {
	UserID    uuid.UUID `json:"user_id"`
	ListingID uuid.UUID `json:"listing_id"`
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
	Title         string    `json:"title"`
	Description   string    `json:"description"`
	Category      string    `json:"category"`
	Condition     string    `json:"condition"`
	Price         float64   `json:"price"`
	Negotiable    bool      `json:"negotiable"`
	EcoScore      float32   `json:"eco_score"`
	EcoAttributes []string  `json:"eco_attributes"`
	ImageUrls     []string  `json:"image_urls"`
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
