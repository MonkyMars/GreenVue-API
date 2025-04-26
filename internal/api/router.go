package api

import (
	"greentrade-eu/internal/auth"
	"greentrade-eu/internal/chat"
	"greentrade-eu/internal/config"
	"greentrade-eu/internal/favorites"
	"greentrade-eu/internal/health"
	"greentrade-eu/internal/listings"
	"greentrade-eu/internal/reviews"
	"greentrade-eu/internal/seller"
	"greentrade-eu/lib/errors"
	"log" // Import log package
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// SetupApp configures the Fiber app with all middleware and routes
func SetupApp(cfg *config.Config) *fiber.App {
	// Check environment
	devMode := cfg.Environment != "production"

	// Configure with custom error handler, explicitly providing the logger
	app := fiber.New(fiber.Config{
		ServerHeader:      "GreenTrade.eu",
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReduceMemoryUsage: true,
		ErrorHandler:      errors.ErrorHandler(errors.ErrorResponseConfig{DevMode: devMode, Logger: log.Printf}), // Explicitly set Logger
	})

	// Setup middleware
	setupMiddleware(app)

	// Setup routes
	setupRoutes(app)

	return app
}

// setupMiddleware adds all middleware to the app
func setupMiddleware(app *fiber.App) {
	// Add request ID middleware early in the chain
	app.Use(errors.RequestID())

	// Add structured logging middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] [${ip}] ${status} - ${method} ${path} - ${latency}\n",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins:     "http://192.168.178.10,http://localhost:3000,http://localhost:8081,https://greentrade.eu,https://www.greentrade.eu,http://10.0.2.2:3000",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
	}))

	// Configure custom rate limiter with different limits for different endpoints
	rateLimiter := errors.NewRateLimiter()
	rateLimiter.Max = 120                // Allow 120 requests
	rateLimiter.Expiration = time.Minute // Per minute
	// Skip rate limiting for certain paths
	rateLimiter.SkipFunc = func(c *fiber.Ctx) bool {
		// Don't rate limit static assets or health check
		path := c.Path()
		return path == "/health" || path == "/favicon.ico"
	}
	app.Use(rateLimiter.Middleware())

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Use(recover.New())

	// Cache middleware with exclusions for auth and non-GET requests
	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			// Don't cache non-GET requests
			if c.Method() != fiber.MethodGet {
				return true
			}

			path := c.Path()

			// Don't cache auth-related routes
			if strings.HasPrefix(path, "/auth") {
				return true
			}

			// Don't cache health checks and chat routes.
			if path == "/health" || path == "/health/detailed" || strings.HasPrefix(path, "/chat") {
				return true
			}

			// Don't cache favicon
			if path == "/favicon.ico" {
				return true
			}

			return false
		},
		Expiration:   time.Minute,
		CacheControl: true,
	}))

	app.Use(etag.New(etag.Config{
		Weak: true,
	}))
}

// setupRoutes configures all the routes for the application
func setupRoutes(app *fiber.App) {
	// Health routes
	setupHealthRoutes(app)

	// Auth routes (public)
	setupAuthRoutes(app)

	// Public listing routes
	setupPublicListingRoutes(app)

	// Seller routes (public), doesn't need auth since info is not sensitive.
	setupSellerRoutes(app)

	// Public review routes
	setupPublicReviewRoutes(app)

	chat.RegisterWebsocketRoutes(app)

	// Protected routes
	api := app.Group("/api", auth.AuthMiddleware())
	setupProtectedListingRoutes(api)
	setupUserRoutes(api)
	setupChatRoutes(api)
	setupProtectedReviewRoutes(api)
	setupFavoritesRoutes(api)
}

// setupHealthRoutes configures health check routes
func setupHealthRoutes(app *fiber.App) {
	app.Get("/health", health.HealthCheck)
	app.Get("/health/detailed", health.DetailedHealth)

	// Prevents 404 spam for favicon.ico
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return errors.ErrNotFound
	})
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(app *fiber.App) {
	app.Post("/auth/login", auth.LoginUser)
	app.Post("/auth/register", auth.RegisterUser)
	app.Post("/auth/refresh", auth.RefreshTokenHandler)

}

// setupPublicListingRoutes configures public listing routes
func setupPublicListingRoutes(app *fiber.App) {
	app.Get("/listings", listings.GetListings)
	app.Get("/listings/category/:category", listings.GetListingByCategory)
	app.Get("/listings/seller/:sellerId", listings.GetListingBySeller)
	// This route should come last as it's a catch-all for any path parameter
	app.Get("/listings/:id", listings.GetListingById)
}

// setupProtectedListingRoutes configures protected listing routes
func setupProtectedListingRoutes(router fiber.Router) {
	router.Post("/listings", listings.PostListing)
	router.Post("/upload/listing_image", listings.UploadHandler)
	router.Delete("/listings/:id", listings.DeleteListingById)
}

// setupSellerRoutes configures seller routes
func setupSellerRoutes(router fiber.Router) {
	router.Get("/seller/:user_id", seller.GetSeller)
}

// setupUserRoutes configures user routes
func setupUserRoutes(router fiber.Router) {
	router.Get("/auth/me", auth.GetUserByAccessToken)
	router.Get("/auth/user/:id", auth.GetUserById)
	router.Put("/auth/user/:id", auth.UpdateUser)
}

// setupChatRoutes configures chat routes
func setupChatRoutes(router fiber.Router) {
	// Conversation routes
	router.Get("/chat/conversation/:userId", chat.GetConversations) // Get conversations for userId
	router.Post("/chat/conversation", chat.CreateConversation)      // Create a new conversation

	// Message routes
	router.Get("/chat/messages/:conversation_id", chat.GetMessagesByConversationID) // Get messages for a conversation
	router.Post("/chat/message", chat.PostMessage)                                  // Post a new message
}

func setupProtectedReviewRoutes(router fiber.Router) {
	router.Post("/reviews", reviews.PostReview)
}

func setupPublicReviewRoutes(router fiber.Router) {
	router.Get("/reviews/:seller_id", reviews.GetReviews)
}

func setupFavoritesRoutes(router fiber.Router) {
	router.Get("/favorites/:id", favorites.GetFavorites)
	router.Post("/favorites/:listing_id/:user_id", favorites.AddFavorite)
	router.Delete("/favorites/:listing_id/:user_id", favorites.DeleteFavorite)
}
