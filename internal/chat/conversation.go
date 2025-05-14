package chat

import (
	"encoding/json"
	"fmt"
	"greenvue-eu/internal/db"
	"greenvue-eu/lib/errors"

	"github.com/gofiber/fiber/v2"
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

	jsonPayload := fmt.Sprintf(`{"buyer_id": "%s", "seller_id": "%s", "listing_id": "%s"}`, buyerId, sellerId, listingId)
	body, err := client.POST(c, "conversations", []byte(jsonPayload))

	if err != nil {
		return errors.InternalServerError("Failed to create conversation: " + err.Error())
	}

	if body == nil {
		return errors.InternalServerError("Failed to create conversation: No response from server")
	}

	// Parse the response directly into a slice of Conversation
	var conversations []Conversation
	err = json.Unmarshal(body, &conversations)
	if err != nil {
		return errors.InternalServerError("Failed to decode conversations: " + err.Error())
	}

	// Return the first conversation
	if len(conversations) > 0 {
		createdConversation := conversations[0]
		query := fmt.Sprintf("id=eq.%s", createdConversation.Id)
		data, err := client.GET(viewName, query)
		if err != nil {
			return errors.InternalServerError("Failed to fetch created conversation: " + err.Error())
		}
		if data == nil {
			return errors.InternalServerError("Failed to fetch created conversation: No data returned")
		}
		var createdConversations []Conversation
		err = json.Unmarshal(data, &createdConversations)
		if err != nil {
			return errors.InternalServerError("Failed to decode created conversation: " + err.Error())
		}
		if len(createdConversations) > 0 {
			return errors.SuccessResponse(c, createdConversations[0])
		} else {
			return errors.InternalServerError("Failed to fetch created conversation: No data returned")
		}
	} else {
		return errors.InternalServerError("No conversation was created")
	}
}

func GetConversations(c *fiber.Ctx) error {
	userId := c.Params("userId")
	if userId == "" {
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
