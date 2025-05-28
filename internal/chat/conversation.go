package chat

import (
	"fmt"
	"greenvue/internal/auth"
	"greenvue/internal/db"
	"greenvue/lib/errors"
	"strings"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

type Conversation struct {
	Id                 string `json:"id"`
	BuyerId            string `json:"buyer_id"`
	SellerId           string `json:"seller_id"`
	ListingId          string `json:"listing_id"`
	CreatedAt          string `json:"created_at"`
	SellerName         string `json:"seller_name"`
	BuyerName          string `json:"buyer_name"`
	ListingName        string `json:"listing_title"`
	LastMessageContent string `json:"last_message_content"`
	LastMessageTime    string `json:"last_message_time"`
}

const viewName string = "conversation_with_usernames"

// creates a conversation between a buyer and a seller for a specific listing and returns the conversation uuid.
func CreateConversation(c *fiber.Ctx) error {

	var payload struct {
		BuyerId   string `json:"buyer_id"`
		SellerId  string `json:"seller_id"`
		ListingId string `json:"listing_id"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Failed to parse JSON payload: " + err.Error())
	}

	buyerId := payload.BuyerId
	sellerId := payload.SellerId
	listingId := payload.ListingId

	client := db.GetGlobalClient()

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}

	if buyerId == "" || sellerId == "" || listingId == "" {
		return errors.BadRequest("Buyer ID and Seller ID are required")
	}

	query := fmt.Sprintf("buyer_id=eq.%s&seller_id=eq.%s&listing_id=eq.%s", buyerId, sellerId, listingId)
	existingData, err := client.GET(viewName, query)

	if err != nil {
		return errors.InternalServerError("Failed to check existing conversations: " + err.Error())
	}

	var existingConversations []Conversation
	if existingData != nil {
		if err := json.Unmarshal(existingData, &existingConversations); err != nil {
			return errors.InternalServerError("Failed to decode existing conversations: " + err.Error())
		}
	}

	if len(existingConversations) > 0 {
		return errors.SuccessResponse(c, existingConversations[0])
	}

	// No existing conversation, try to create one
	jsonPayload := fmt.Sprintf(`{"buyer_id": "%s", "seller_id": "%s", "listing_id": "%s"}`, buyerId, sellerId, listingId)
	body, err := client.POST("conversations", []byte(jsonPayload))

	// Optional: handle race condition fallback
	if err != nil {
		if strings.Contains(err.Error(), "duplicate key value") {
			// Another client just created it — fetch it again
			existingData, err := client.GET(viewName, query)
			if err == nil && existingData != nil {
				var fallbackConvos []Conversation
				if err := json.Unmarshal(existingData, &fallbackConvos); err == nil && len(fallbackConvos) > 0 {
					return errors.SuccessResponse(c, fallbackConvos[0])
				}
			}
		}
		return errors.InternalServerError("Failed to create conversation: " + err.Error())
	}

	var conversations []Conversation
	if err := json.Unmarshal(body, &conversations); err != nil || len(conversations) == 0 {
		return errors.InternalServerError("Failed to decode or create conversation.")
	}

	return errors.SuccessResponse(c, conversations[0])
}

func GetConversations(c *fiber.Ctx) error {
	claims, ok := c.Locals("user").(*auth.Claims)
	if !ok {
		return errors.Unauthorized("Invalid token claims")
	}

	userId := claims.UserId
	if userId == uuid.Nil {
		return errors.BadRequest("User ID is required")
	}

	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}

	// Construct the query string for Supabase: or=(seller_id.eq.userId,buyer_id.eq.userId)
	query := fmt.Sprintf("or=(seller_id.eq.%s,buyer_id.eq.%s)", userId, userId)

	// Execute the query
	data, err := client.GET(viewName, query)
	if err != nil {
		return errors.InternalServerError("Failed to fetch conversations: " + err.Error())
	}

	if data == nil {
		// Return an empty list if no data is found, which is not necessarily an error
		return errors.SuccessResponse(c, []Conversation{})
	}

	// Parse the JSON response into a slice of Conversation structs
	var conversations []Conversation
	err = json.Unmarshal(data, &conversations)
	if err != nil {
		// Log the raw data for debugging if unmarshalling fails
		fmt.Printf("Failed to unmarshal conversations data: %s\nRaw data: %s\n", err.Error(), string(data))
		return errors.InternalServerError("Failed to decode conversations: " + err.Error())
	}

	// Return the fetched conversations
	return errors.SuccessResponse(c, conversations)
}
