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

	// Load configuration
	cfg := config.LoadConfig()

	// Set up the application with routes and middleware
	app := api.SetupApp(cfg)

	// Listen on port from config
	port := cfg.Server.Port
	if port == "" {
		port = "8081"
	}
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
