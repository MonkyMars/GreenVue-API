package api

import (
	"greenvue/internal/jobs"
	"greenvue/lib/image"
	"log"
	"time"
)

// initImageProcessingQueue initializes the image processing queue
func initImageProcessingQueue() {
	// Initialize the global image queue
	image.InitializeImageQueue()
	log.Printf("Image processing queue initialized")
}

// setupDefaultImageProcessingJob sets up a background job to process images
func setupDefaultImageProcessingJob() {
	// Set up a job to process images every 10 seconds in non-production environments
	imageProcessingOptions := &jobs.ImageProcessingOptions{
		BatchSize: 10, // Process up to 10 images at once
	}

	err := jobs.GlobalScheduler.AddJob(
		"process-image-queue",                                 // Job ID
		"Process Image Queue",                                 // Job Name
		"Process pending images in queue",                     // Description
		jobs.CreateImageProcessingJob(imageProcessingOptions), // Job function
		10*time.Second,                                        // Run every 10 seconds
	)

	if err != nil {
		log.Printf("Warning: Could not add image processing job: %v", err)
	} else {
		log.Printf("Image processing background job initialized")
	}
}
