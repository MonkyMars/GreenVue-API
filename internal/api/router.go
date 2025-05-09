package api

import (
	"greenvue-eu/internal/auth"
	"greenvue-eu/internal/chat"
	"greenvue-eu/internal/config"
	"greenvue-eu/internal/favorites"
	"greenvue-eu/internal/health"
	"greenvue-eu/internal/listings"
	"greenvue-eu/internal/reviews"
	"greenvue-eu/internal/seller"
	"greenvue-eu/lib/errors"
	"log"
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
	DevMode := cfg.Environment != "production"

	// Configure with custom error handler, explicitly providing the logger
	app := fiber.New(fiber.Config{
		ServerHeader:      "GreenVue",
		ReadTimeout:       cfg.Server.ReadTimeout,
		WriteTimeout:      cfg.Server.WriteTimeout,
		IdleTimeout:       cfg.Server.IdleTimeout,
		ReduceMemoryUsage: true,
		ErrorHandler:      errors.ErrorHandler(errors.ErrorResponseConfig{DevMode: DevMode, Logger: log.Printf}), // Explicitly set Logger
	})

	// Setup middleware
	setupMiddleware(app, cfg)

	// Setup routes
	setupRoutes(app)

	return app
}

// setupMiddleware adds all middleware to the app
func setupMiddleware(app *fiber.App, cfg *config.Config) {
	// Add request ID middleware early in the chain
	app.Use(errors.RequestID())

	// Add structured logging middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] [${ip}] ${status} - ${method} ${path} - ${latency}\n",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: func() string {
			if cfg.Environment != "production" {
				return "*" // Allow all origins in development
			}
			// Specify allowed origins in production
			allowedOrigins := []string{
				"https://www.greenvue.eu",
			}
			return strings.Join(allowedOrigins, ",")
		}(),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: false,
	}))

	// Configure custom rate limiter with different limits for different endpoints
	rateLimiter := errors.NewRateLimiter()
	rateLimiter.Max = 120                // Allow 120 requests
	rateLimiter.Expiration = time.Minute // Per minute
	app.Use(rateLimiter.Middleware())

	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))

	app.Use(recover.New(recover.Config{
		EnableStackTrace: cfg.Environment != "production",
	}))

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
			if strings.Contains(path, "health") || strings.HasPrefix(path, "/chat") {
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
	setupHealthRoutes(api)
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
	router.Post("/auth/resend_email", auth.ResendConfirmationEmail)
	router.Put("/auth/user/:id", auth.UpdateUser)
}

// setupChatRoutes configures chat routes
func setupChatRoutes(router fiber.Router) {
	// Conversation routes
	router.Get("/chat/conversation/:userId", chat.GetConversations)
	router.Post("/chat/conversation", chat.CreateConversation)

	// Message routes
	router.Get("/chat/messages/:conversation_id", chat.GetMessagesByConversationID)
	router.Post("/chat/message", chat.PostMessage)
}

// setupProtectedReviewRoutes configures protected review routes
func setupProtectedReviewRoutes(router fiber.Router) {
	router.Post("/reviews", reviews.PostReview)
}

// setupPublicReviewRoutes configures public review routes
func setupPublicReviewRoutes(router fiber.Router) {
	router.Get("/reviews/:seller_id", reviews.GetReviews)
}

// setupFavoritesRoutes configures favorites routes
func setupFavoritesRoutes(router fiber.Router) {
	router.Get("/favorites/:user_id", favorites.GetFavorites)
	router.Get("/favorites/check/:listing_id/:user_id", favorites.IsFavorite)
	router.Post("/favorites", favorites.AddFavorite)
	router.Delete("/favorites/:listing_id/:user_id", favorites.DeleteFavorite)
}

// setupHealthRoutes configures health check routes
func setupHealthRoutes(router fiber.Router) {
	router.Get("/health", health.HealthCheck)
	router.Get("/health/detailed", health.DetailedHealth)

	// Prevents 404 spam for favicon.ico
	router.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return errors.ErrNotFound
	})
}
