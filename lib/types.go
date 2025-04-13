package lib

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
