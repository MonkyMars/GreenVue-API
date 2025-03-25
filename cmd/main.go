package main

import (
	"greentrade-eu/internal/auth"
	"greentrade-eu/internal/listings"
	"greentrade-eu/internal/seller"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	// "github.com/gofiber/fiber/v2/middleware/cache"
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
	// app.Use(cache.New(cache.Config{
	// 	Next: func(c *fiber.Ctx) bool {
	// 		return c.Method() != fiber.MethodGet
	// 	},
	// 	Expiration:   time.Minute,
	// 	CacheControl: true,
	// }))
	app.Use(etag.New(etag.Config{
		Weak: true,
	}))

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"message": "OK"})
	})

	// Listings
	app.Get("/listings", listings.GetListings)
	app.Get("/listings/:id", listings.GetListingById)
	app.Get("/listings/category/:category", listings.GetListingByCategory)
	app.Post("/listings", listings.PostListing)
	app.Post("/upload/listing_image", listings.UploadHandler)
	app.Delete("/listings/:id", listings.DeleteListingById)

	// Auth
	app.Post("/auth/login", auth.LoginUser)
	app.Post("/auth/register", auth.RegisterUser)
	app.Get("/auth/user/:id", auth.GetUserById)

	// Sellers
	app.Get("/sellers", seller.GetSellers)
	app.Get("/sellers/:id", seller.GetSellerById)
	app.Post("/sellers", seller.CreateSeller)

	// Prevents 404 spam for favicon.ico
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent) // 204 No Content
	})

	// Listen on port 8081
	app.Listen(":8081")
}
