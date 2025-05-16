package main

import (
	"greenvue/internal/api"
	"greenvue/internal/config"
	"greenvue/internal/db"
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
	app := api.SetupApp(cfg)

	// Perform a sanity check on the database connection
	if ok, err := db.SanityCheck(); !ok {
		log.Fatalf("Sanity check failed: %v", err)
	}

	// Start server using port from config
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	log.Fatal(app.Listen(":" + port))
}
