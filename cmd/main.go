package main

import (
	"greentrade-eu/internal/auth"
	"greentrade-eu/internal/health"
	"greentrade-eu/internal/listings"
	"greentrade-eu/internal/seller"
	"greentrade-eu/lib/errors"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env.local"); err != nil {
		log.Printf("Warning: Error loading .env.local file: %v", err)
	}

	// Check environment
	devMode := os.Getenv("ENV") != "production"

	// Configure with custom error handler
	app := fiber.New(fiber.Config{
		ServerHeader:      "GreenTrade.eu",
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReduceMemoryUsage: true,
		ErrorHandler:      errors.ErrorHandler(errors.ErrorResponseConfig{DevMode: devMode}),
	})

	// Add request ID middleware early in the chain
	app.Use(errors.RequestID())

	// Add structured logging middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] [${ip}] ${status} - ${method} ${path} - ${latency}\n",
	}))

	app.Use(cors.New())

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

	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Method() != fiber.MethodGet
		},
		Expiration:   time.Minute,
		CacheControl: true,
	}))

	app.Use(etag.New(etag.Config{
		Weak: true,
	}))

	// Health checks
	app.Get("/health", health.HealthCheck)
	app.Get("/health/detailed", health.DetailedHealth)

	// Listings
	app.Get("/listings", listings.GetListings)
	app.Get("/listings/category/:category", listings.GetListingByCategory)
	app.Get("/listings/:id", listings.GetListingById)
	app.Post("/listings", listings.PostListing)
	app.Post("/upload/listing_image", listings.UploadHandler)
	app.Delete("/listings/:id", listings.DeleteListingById)

	// Auth
	app.Post("/auth/login", auth.LoginUser)
	app.Post("/auth/register", auth.RegisterUser)
	app.Get("/auth/user/:id", auth.GetUserById)
	app.Get("/auth/me", auth.GetUserByAccessToken)

	// Sellers
	app.Get("/sellers", seller.GetSellers)
	app.Get("/sellers/:id", seller.GetSellerById)
	app.Post("/sellers", seller.CreateSeller)

	// Prevents 404 spam for favicon.ico
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent) // 204 No Content
	})

	// Listen on port 8081
	if err := app.Listen(":8081"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
