package main

import (
	"greentrade-eu/internal/listings"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cache"
	"github.com/gofiber/fiber/v2/middleware/compress"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/etag"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env.local"); err != nil {
		log.Printf("Warning: Error loading .env.local file: %v", err)
	}

	app := fiber.New(fiber.Config{
		ServerHeader:      "GreenTrade.eu",
		ReadTimeout:       5 * time.Second,
		WriteTimeout:      5 * time.Second,
		IdleTimeout:       120 * time.Second,
		ReduceMemoryUsage: true,
	})

	app.Use(cors.New())
	app.Use(limiter.New(limiter.Config{
		Max:        50,
		Expiration: 30 * time.Second,
		KeyGenerator: func(c *fiber.Ctx) string {
			return c.IP()
		},
	}))
	app.Use(compress.New(compress.Config{
		Level: compress.LevelBestSpeed,
	}))
	app.Use(recover.New())
	app.Use(cache.New(cache.Config{
		Next: func(c *fiber.Ctx) bool {
			return c.Method() != fiber.MethodGet
		},
		Expiration:   10 * time.Minute,
		CacheControl: true,
	}))
	app.Use(etag.New(etag.Config{
		Weak: true,
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "OK"})
	})

	app.Get("/listings", listings.GetListings)
	app.Post("/listings", listings.PostListing)
	app.Post("/upload/listing_image", listings.UploadHandler)
	// app.Delete("/listings/:id", listings.DeleteListing) - not implemented yet
	// TODO: Implement DeleteListing

	app.Listen(":8080")
}
