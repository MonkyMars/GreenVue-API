package main

import (
	"greentrade-eu/internal/api"
	"greentrade-eu/internal/config"
	"log"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env.local"); err != nil {
		log.Printf("Warning: Error loading .env.local file: %v", err)
	}

	// Load configuration using the config package
	cfg := config.LoadConfig()

	// Setup the Fiber app using the api package's function
	// This likely includes middleware and basic route setup
	app := api.SetupApp(cfg)

	// Start server using port from config
	port := cfg.Server.Port
	if port == "" {
		port = "8080" // Fallback default if config doesn't provide one
	}

	log.Fatal(app.Listen(":" + port))
}
