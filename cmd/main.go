package main

import (
	"context"
	"greenvue/internal/api"
	"greenvue/internal/config"
	"greenvue/internal/db"
	"greenvue/lib/email"
	"greenvue/lib/image"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

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

	// Initialize the email service
	supabaseURL := cfg.Database.SupabaseURL
	supabaseKey := cfg.Database.SupabaseKey
	email.InitializeEmailService(supabaseURL, supabaseKey)
	log.Println("Email service initialized")

	// Start server using port from config
	port := cfg.Server.Port
	if port == "" {
		port = "8080"
	}

	// Import the jobs package to ensure it's properly initialized
	// This is needed because the scheduler is initialized in the api package
	_ = "greenvue/internal/jobs"

	// Start the server with graceful shutdown
	go func() {
		if err := app.Listen(":" + port); err != nil {
			log.Printf("Server shutting down: %v", err)
		}
	}()

	// Listen for shutdown signals
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	// Block until we receive a shutdown signal
	<-c
	log.Println("Gracefully shutting down...")
	// Shutdown the background job scheduler if it exists
	if api.GetJobScheduler() != nil {
		api.GetJobScheduler().Shutdown()
		log.Println("Background job scheduler shutdown")
	}

	// Persist the image queue if it exists
	if image.GlobalImageQueue != nil {
		if err := image.GlobalImageQueue.PersistToDisk(); err != nil {
			log.Printf("Error persisting image queue: %v", err)
		} else {
			log.Println("Image queue persisted to disk")
		}
	}

	// Shut down the server with a timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := app.ShutdownWithContext(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}
