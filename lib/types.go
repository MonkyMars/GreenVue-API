package lib

import (
	"github.com/gofiber/fiber/v2"
)

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User         User   `json:"user"`
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

type CompleteUser struct {
	User      User              `json:"user"`
	Listings  []FetchedListing  `json:"listings"`
	Reviews   []FetchedReview   `json:"reviews"`
	Messages  []Message         `json:"messages"`
	Favorites []FetchedFavorite `json:"favorites"`
}
