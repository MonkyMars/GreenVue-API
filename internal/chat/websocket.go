package chat

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

// --- Rate Limiting Constants ---
const (
	messageRateLimitWindow = 5 * time.Second // Example: 5-second window
	messageRateLimitMax    = 10              // Example: Max 10 messages per window (adjust as needed)
	pingTimeout            = 5 * time.Second // Close connection if no ping received within this time
)

// Client represents a connected user via WebSocket
type Client struct {
	Conn           *websocket.Conn
	ConversationID string
	UserID         string // Keep track of the user ID if needed for authorization etc.

	// --- Rate Limiting State ---
	messageCount    int       // Messages sent in the current window
	windowStartTime time.Time // Start time of the current window
	// Use a mutex per client if handling messages concurrently within the handler (unlikely here)
	// rateLimitMux sync.Mutex

	// Ping/Pong tracking
	lastPingTime time.Time
	pingTimer    *time.Timer
}

var (
	// clients stores active WebSocket connections mapped by their connection pointer
	clients    = make(map[*websocket.Conn]*Client)
	clientsMux sync.RWMutex // Use RWMutex for better concurrent read performance
)

// RegisterWebsocketRoutes sets up the WebSocket endpoint for the chat
func RegisterWebsocketRoutes(app *fiber.App) {
	// Configure CORS specifically for WebSocket routes
	wsGroup := app.Group("/ws")
	wsGroup.Use(cors.New(cors.Config{
		AllowOrigins:     "http://192.168.178.10,http://localhost:3000,http://localhost:8081,https://greenvue.eu,https://www.greenvue.eu,http://10.0.2.2:3000",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowMethods:     "GET, POST",
		AllowCredentials: true,
		ExposeHeaders:    "Upgrade, Connection",
	}))

	// Middleware to check if it's a WebSocket upgrade request
	wsGroup.Use(func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	// WebSocket endpoint: /ws/chat/:conversation_id/:user_id
	// UserID might be used later for authentication/authorization checks
	wsGroup.Get("/chat/:conversation_id/:user_id", websocket.New(handleChatWebSocket, websocket.Config{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		// Add basic configuration to improve stability
		EnableCompression: true,
	}))
}

// handleChatWebSocket manages individual WebSocket connections
func handleChatWebSocket(c *websocket.Conn) {
	conversationID := c.Params("conversation_id")
	userID := c.Params("user_id")

	// Basic validation
	if conversationID == "" || userID == "" {
		log.Println("WebSocket connection rejected: Missing conversation_id or user_id")
		if err := c.Close(); err != nil {
			log.Println("Error closing WebSocket:", err)
			return
		}
		return
	}

	// Create and register the client
	client := &Client{
		Conn:           c,
		ConversationID: conversationID,
		UserID:         userID,
		// --- Initialize Rate Limiting State ---
		windowStartTime: time.Now(),
		messageCount:    0,
		lastPingTime:    time.Now(),
		pingTimer:       time.NewTimer(pingTimeout),
	}

	clientsMux.Lock()
	clients[c] = client
	clientsMux.Unlock()

	log.Printf("WebSocket client connected: User %s, Conversation %s", userID, conversationID)

	// Send a connection confirmation message
	welcomeMsg := map[string]string{
		"type":        "connection_established",
		"message":     "Connected to chat",
		"pingTimeout": pingTimeout.String(),
	}
	welcomeJSON, _ := json.Marshal(welcomeMsg)
	if err := c.WriteMessage(websocket.TextMessage, welcomeJSON); err != nil {
		log.Printf("Error sending welcome message: %v", err)
	}

	// Defer cleanup: remove client and close connection when the function returns
	defer func() {
		// Stop the ping timer to prevent leaks
		client.pingTimer.Stop()
		clientsMux.Lock()
		delete(clients, c)
		clientsMux.Unlock()
		if err := c.Close(); err != nil {
			log.Println("Error closing WebSocket:", err)
			return
		}
		log.Printf("WebSocket client disconnected: User %s, Conversation %s", userID, conversationID)
	}()

	// Start goroutine to monitor ping timeouts
	go func() {
		done := make(chan struct{})
		defer close(done)

		for range client.pingTimer.C {
			if time.Since(client.lastPingTime) > pingTimeout {
				log.Printf("WebSocket ping timeout for User %s, Conversation %s", userID, conversationID)
				c.Close()
				return
			}
			client.pingTimer.Reset(pingTimeout)
		}
	}()

	for {
		messageType, msg, err := c.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket read error: %v", err)
			} else {
				log.Printf("WebSocket connection closed normally for User %s, Conversation %s", userID, conversationID)
			}
			break // Exit loop on error or close
		}

		now := time.Now()
		// Reset window if it has expired
		if now.Sub(client.windowStartTime) > messageRateLimitWindow {
			client.windowStartTime = now
			client.messageCount = 0
		}

		// Increment message count for this window
		client.messageCount++

		// Check if limit is exceeded
		if client.messageCount > messageRateLimitMax {
			log.Printf("WebSocket message rate limit exceeded for User %s (Conv %s)", client.UserID, client.ConversationID)
			continue // Skip processing this message, read the next one
		}
		// --- End Rate Limiting Check ---

		if messageType == websocket.TextMessage {
			// Handle incoming messages from client if necessary (e.g., typing indicators, or if clients send messages via WS)
			log.Printf("Received message from User %s: %s", userID, string(msg))

			// Parse the message to check if it's a ping
			var msgData map[string]string
			if err := json.Unmarshal(msg, &msgData); err == nil {
				if msgData["type"] == "ping" {
					// Update last ping time
					client.lastPingTime = time.Now()
					// Reset the ping timer
					client.pingTimer.Reset(pingTimeout)

					// Send pong response immediately
					pongMsg := map[string]string{"type": "pong"}
					pongJSON, _ := json.Marshal(pongMsg)
					if err := c.WriteMessage(websocket.TextMessage, pongJSON); err != nil {
						log.Printf("Error sending pong response: %v", err)
					}
					continue // Skip the general acknowledgment for ping messages
				}
			}

			// Send back an acknowledgment to confirm message receipt
			ackMsg := map[string]string{"type": "ack", "message": "Message received"}
			ackJSON, _ := json.Marshal(ackMsg)
			if err := c.WriteMessage(websocket.TextMessage, ackJSON); err != nil {
				log.Printf("Error sending message acknowledgment: %v", err)
			}
		} else {
			log.Printf("Received non-text message type: %d", messageType)
		}
	}
}

// BroadcastMessage sends a message to all clients connected to a specific conversation
func BroadcastMessage(conversationID string, message Message) {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		log.Printf("Error marshalling message for broadcast: %v", err)
		return
	}

	clientsMux.RLock() // Use read lock for iterating
	defer clientsMux.RUnlock()

	activeConnections := 0
	for conn, client := range clients {
		if client.ConversationID == conversationID {
			activeConnections++
			// Launch sending in a goroutine to avoid blocking the broadcast loop
			go func(c *websocket.Conn) {
				// Need to lock writing to a single connection if multiple goroutines might write concurrently
				// However, websocket library might handle this internally. Check docs if issues arise.
				err := c.WriteMessage(websocket.TextMessage, messageJSON)
				if err != nil {
					log.Printf("Error sending message to client %s: %v", client.UserID, err)
					// Optional: Consider closing the connection or marking the client for removal
					// c.Close() // Be careful with concurrent map modification if you do this
				}
			}(conn) // Pass conn to the goroutine
		}
	}
	log.Printf("Broadcasted message to %d clients in conversation %s", activeConnections, conversationID)
}
