package lib

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
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
