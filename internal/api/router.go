package api

import (
	"greenvue/internal/auth"
	"greenvue/internal/chat"
	"greenvue/internal/config"
	"greenvue/internal/favorites"
	"greenvue/internal/health"
	"greenvue/internal/jobs"
	"greenvue/internal/listings"
	"greenvue/internal/reviews"
	"greenvue/internal/seller"
	"greenvue/lib/errors"
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
		ErrorHandler:      errors.ErrorHandler(errors.ErrorResponseConfig{DevMode: DevMode, Logger: log.Printf}),
		BodyLimit:         20 * 1024 * 1024, // 20 MB
	})

	// Setup middleware
	setupMiddleware(app, cfg)
	// Setup routes
	setupRoutes(app, cfg)

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
				return "http://localhost:3000,http://localhost:8080,http://localhost:8081,http://127.0.0.1:3000,http://192.168.178.10:3000,http://192.168.178.10"
			}
			// Specify allowed origins in production
			allowedOrigins := []string{
				"https://www.greenvue.eu",
				"https://greenvue.eu",
			}
			return strings.Join(allowedOrigins, ",")
		}(),
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS,PATCH",
		AllowHeaders:     "Origin,Content-Type,Accept,Authorization",
		AllowCredentials: true, // Enable credentials for cookies
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

			// Don't cache in development
			if cfg.Environment != "production" {
				return true
			}

			// Don't cache non-GET requests
			if c.Method() != fiber.MethodGet {
				return true
			}

			path := c.Path()

			// Don't cache health checks, chat and auth routes.
			if strings.Contains(path, "/health") || strings.Contains(path, "/chat") || strings.Contains(path, "/auth") {
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
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.OriginalURL()
		},
	}))

	app.Use(etag.New(etag.Config{
		Weak: true,
	}))
}

// setupRoutes configures all the routes for the application
func setupRoutes(app *fiber.App, cfg *config.Config) {
	// Initialize the job scheduler
	jobs.Initialize()

	// Initialize email service
	initEmailService(cfg)

	// Initialize image processing queue
	initImageProcessingQueue()

	// Setup default background jobs if not in production
	if cfg.Environment != "production" {
		setupDefaultEmailJob()
		setupDefaultImageProcessingJob()
		setupDebugRoutes(app) // Only set up debug routes in non-production
	}

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
	setupJobRoutes(api)
}

// setupAuthRoutes configures authentication routes
func setupAuthRoutes(app *fiber.App) {
	app.Post("/auth/login", auth.LoginUser)
	app.Get("/auth/login/google", auth.HandleGoogleLogin)
	app.Get("/auth/register/google", auth.HandleGoogleRegistrationStart)
	app.Get("/auth/callback/google", auth.HandleGoogleCallback)
	app.Post("/auth/register", auth.RegisterUser)
	app.Post("/auth/refresh", auth.RefreshTokenHandler)
	app.Post("/auth/logout", auth.LogoutUser)
	app.Get("/auth/confirm_email", auth.VerifyEmailRedirect)
}

// setupPublicListingRoutes configures public listing routes
func setupPublicListingRoutes(app *fiber.App) {
	app.Get("/listings", listings.GetListings)
	app.Get("/listings/category/:category", listings.GetListingByCategory)
	app.Get("/listings/seller/:seller_id", listings.GetListingBySeller)
	app.Get("/listings/:listing_id", listings.GetListingById)
}

// setupProtectedListingRoutes configures protected listing routes
func setupProtectedListingRoutes(router fiber.Router) {
	router.Post("/listings", listings.PostListing)
	router.Post("/upload/listing_image", listings.QueuedUploadHandler) // Queued image processing
	router.Delete("/listings/:listing_id", listings.DeleteListingById)
}

// setupSellerRoutes configures seller routes
func setupSellerRoutes(router fiber.Router) {
	router.Get("/seller/:seller_id", seller.GetSeller)
}

// setupUserRoutes configures user routes
func setupUserRoutes(router fiber.Router) {
	router.Get("/auth/me", auth.GetUserByAccessToken)
	router.Post("/auth/resend_email", auth.ResendConfirmationEmail)
	router.Patch("/auth/user", auth.UpdateUser)
	router.Get("/auth/download_user_data", auth.DownloadUserData)
}

// setupChatRoutes configures chat routes
func setupChatRoutes(router fiber.Router) {
	// Conversation routes
	router.Get("/chat/conversation", chat.GetConversations)
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
func setupPublicReviewRoutes(app *fiber.App) {
	app.Get("/reviews/:seller_id", reviews.GetReviews)
}

// setupFavoritesRoutes configures favorites routes
func setupFavoritesRoutes(router fiber.Router) {
	router.Get("/favorites", favorites.GetFavorites)
	router.Get("/favorites/check/:listing_id", favorites.IsFavorite)
	router.Post("/favorites", favorites.AddFavorite)
	router.Delete("/favorites/:listing_id", favorites.DeleteFavorite)
}

// setupHealthRoutes configures health check routes
func setupHealthRoutes(router fiber.Router) {
	router.Get("/health", health.HealthCheck)
	router.Get("/health/detailed", health.DetailedHealth)
}

// setupJobRoutes configures background job routes
func setupJobRoutes(router fiber.Router) {
	router.Get("/jobs", jobs.GetJobs)
	router.Get("/jobs/:job_id", jobs.GetJobByID)
	router.Post("/jobs", jobs.CreateJob)
	router.Delete("/jobs/:job_id", jobs.DeleteJob)
}

// setupDebugRoutes configures debug routes for development/testing
func setupDebugRoutes(app *fiber.App) {
	debug := app.Group("/debug")
	debug.Post("/send-test-email", TestEmailHandler)
	debug.Get("/email-queue-status", GetEmailQueueStatusHandler)
	debug.Get("/image-queue-status", GetImageQueueStatusHandler)
	debug.Post("/upload-test-image", TestImageQueueHandler)
}
