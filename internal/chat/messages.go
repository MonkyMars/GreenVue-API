package chat

import (
	"fmt"
	"greenvue/internal/db"
	"greenvue/lib/errors"
	"log" // Import log package
	"time"

	"encoding/json"

	"github.com/gofiber/fiber/v2"
)

type Message struct {
	ID             string    `json:"id"`
	ConversationID string    `json:"conversation_id"`
	SenderID       string    `json:"sender_id"`
	Content        string    `json:"content"`
	CreatedAt      time.Time `json:"created_at"`
}

func GetMessagesByConversationID(c *fiber.Ctx) error {
	client := db.GetGlobalClient()

	if client == nil {
		return errors.InternalServerError("Database connection failed. Please check SUPABASE_URL and SUPABASE_ANON.")
	}

	conversationID := c.Params("conversation_id")
	if conversationID == "" {
		return errors.BadRequest("Conversation ID is required")
	}

	query := fmt.Sprintf("conversation_id=eq.%s", conversationID)
	data, err := client.GET("messages", query)

	if err != nil {
		return errors.InternalServerError("Failed to retrieve messages: " + err.Error())
	}

	var messages []Message

	err = json.Unmarshal(data, &messages)
	if err != nil {
		return errors.InternalServerError("Failed to parse messages: " + err.Error())
	}

	return errors.SuccessResponse(c, messages)
}

func PostMessage(c *fiber.Ctx) error {
	var payload struct {
		ConversationID string `json:"conversation_id"`
		SenderID       string `json:"sender_id"`
		Content        string `json:"content"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return errors.BadRequest("Failed to parse JSON payload: " + err.Error())
	}

	// Validate payload
	if payload.ConversationID == "" || payload.SenderID == "" || payload.Content == "" {
		return errors.BadRequest("Missing required fields: conversation_id, sender_id, content")
	}

	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Database connection failed.")
	}

	// Prepare the message data for insertion
	newMessage := map[string]any{
		"conversation_id": payload.ConversationID,
		"sender_id":       payload.SenderID,
		"content":         payload.Content,
	}

	// Marshal the map into JSON bytes
	insertedData, err := client.POST("messages", newMessage)
	if err != nil {
		return errors.InternalServerError("Failed to post message: " + err.Error())
	}

	var createdMessages []Message
	if err := json.Unmarshal(insertedData, &createdMessages); err != nil || len(createdMessages) == 0 {
		log.Println("Warning: Failed to parse inserted message data after successful post:", err)
		return errors.SuccessResponse(c, fiber.Map{"status": "Message posted successfully, but response parsing failed"})
	}

	// Broadcast the newly created message to WebSocket clients
	if len(createdMessages) > 0 {
		go BroadcastMessage(payload.ConversationID, createdMessages[0]) // Run broadcast in a goroutine
	}

	// Return the newly created message
	return errors.SuccessResponse(c, createdMessages[0])
}
