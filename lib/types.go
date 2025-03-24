package lib

type User struct {
	ID    string `json:"id"`
	Email string `json:"email"`
}

type AuthResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
	User        User   `json:"user"`
}