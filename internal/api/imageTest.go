package api

import (
	"fmt"
	"greenvue/lib"
	"greenvue/lib/errors"
	"greenvue/lib/image"
	"io"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// TestImageQueueHandler is a debug handler to test the image processing system
func TestImageQueueHandler(c *fiber.Ctx) error {
	// Parse request body
	var req struct {
		ListingTitle string `json:"listing_title"` // Title for the listing
		ImagePath    string `json:"image_path"`    // Path to a test image file
		ProcessNow   bool   `json:"process_now"`   // Whether to process immediately
	}

	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	// Validate
	if req.ListingTitle == "" {
		req.ListingTitle = "Test Listing"
	}

	if req.ImagePath == "" {
		return errors.BadRequest("Image path is required")
	}

	// Read the test image file
	file, err := os.Open(req.ImagePath)
	if err != nil {
		return errors.BadRequest(fmt.Sprintf("Could not open image file: %v", err))
	}
	defer file.Close()

	// Read file data
	fileData, err := io.ReadAll(file)
	if err != nil {
		return errors.BadRequest(fmt.Sprintf("Error reading image file: %v", err))
	}

	// Generate a unique filename
	fileName := fmt.Sprintf("%s-%s.webp", lib.SanitizeFilename(req.ListingTitle), uuid.New().String())

	// Create an image job
	imageJob := image.ImageJob{
		ID:           uuid.New().String(),
		FileName:     fileName,
		ListingTitle: req.ListingTitle,
		ImageData:    fileData,
		CreatedAt:    time.Now(),
		Status:       "pending",
		MaxRetries:   3,
	}

	// Add to the queue
	err = image.QueueImage(imageJob)
	if err != nil {
		return errors.InternalServerError(fmt.Sprintf("Error queueing image: %v", err))
	}

	// Process the queue immediately if requested
	if req.ProcessNow {
		go func() {
			if image.GlobalImageQueue != nil {
				image.GlobalImageQueue.ProcessQueue(1)
			}
		}()
	}

	return errors.SuccessResponse(c, fiber.Map{
		"message":  "Test image queued successfully",
		"image_id": imageJob.ID,
		"status":   "pending",
	})
}
