package listings

import (
	"bytes"
	"fmt"
	"greenvue/lib"
	"greenvue/lib/errors"
	"greenvue/lib/image"
	"io"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// QueuedUploadHandler processes image uploads asynchronously using the background job system
func QueuedUploadHandler(c *fiber.Ctx) error {
	// Extract listing title from form data
	listingTitle := c.FormValue("listing_title")

	if listingTitle == "" {
		return errors.BadRequest("Missing listing title in form data")
	}

	// Get files from formdata
	form, err := c.MultipartForm()

	if err != nil {
		return errors.BadRequest("Failed to parse form data: " + err.Error())
	}

	queuedImageIds := []string{}

	fileHeaders, ok := form.File["file"]
	if !ok || len(fileHeaders) == 0 {
		// If no files with key "file", check if there are any files at all
		totalFiles := 0
		for _, v := range form.File {
			totalFiles += len(v)
			// Process these files instead
			for _, fileHeader := range v {
				if fileHeader.Size == 0 {
					continue
				}

				src, err := fileHeader.Open()
				if err != nil {
					log.Printf("Error opening file: %v", err)
					continue
				}

				// Generate a unique filename
				fileName := fmt.Sprintf("%s-%s.webp", lib.SanitizeFilename(listingTitle), uuid.New().String()) // Need to read the file data completely as we'll need to close the file
				fileData, err := io.ReadAll(src)
				src.Close() // Close immediately after reading

				if err != nil {
					log.Printf("Error reading file data: %v", err)
					continue
				}

				// First validate the image format
				format, err := validateImage(fileData)
				if err != nil {
					log.Printf("Invalid image file: %v", err)
					continue
				}

				log.Printf("Processing valid %s image", format)

				// Convert image to WebP here so we don't queue massive raw images
				// This also strips metadata and resizes the image
				webpData, err := convertToWebP(bytes.NewReader(fileData))
				if err != nil {
					log.Printf("Error converting to WebP: %v", err)
					continue
				}

				// Create an image job
				imageJob := image.ImageJob{
					ID:           uuid.New().String(),
					FileName:     fileName,
					ListingTitle: listingTitle,
					ImageData:    webpData.Bytes(),
					CreatedAt:    time.Now(),
					Status:       "pending",
					MaxRetries:   3,
				}

				// Add to the queue instead of processing immediately
				err = image.QueueImage(imageJob)
				if err != nil {
					log.Printf("Error queueing image: %v", err)
					continue
				}

				// Append the image ID to the list
				queuedImageIds = append(queuedImageIds, imageJob.ID)
			}
		}

		if totalFiles == 0 {
			log.Println("No files found in the form data")
			return errors.BadRequest("No files were submitted. Make sure to include files in your FormData.")
		}
	} else {
		// Process files from the expected "file" key
		for _, fileHeader := range fileHeaders {
			if fileHeader.Size == 0 {
				log.Println("Skipping empty file")
				continue
			}

			src, err := fileHeader.Open()
			if err != nil {
				log.Printf("Error opening file: %v", err)
				continue
			}

			// Generate a unique filename
			fileName := fmt.Sprintf("%s-%s.webp", lib.SanitizeFilename(listingTitle), uuid.New().String()) // Read file data
			fileData, err := io.ReadAll(src)
			src.Close() // Close immediately after reading

			if err != nil {
				log.Printf("Error reading file data: %v", err)
				continue
			}

			// First validate the image format
			format, err := validateImage(fileData)
			if err != nil {
				log.Printf("Invalid image file: %v", err)
				continue
			}

			log.Printf("Processing valid %s image", format)

			// Convert to WebP (also strips metadata and resizes)
			webpData, err := convertToWebP(bytes.NewReader(fileData))
			if err != nil {
				log.Printf("Error converting to WebP: %v", err)
				continue
			}

			// Create an image job
			imageJob := image.ImageJob{
				ID:           uuid.New().String(),
				FileName:     fileName,
				ListingTitle: listingTitle,
				ImageData:    webpData.Bytes(),
				CreatedAt:    time.Now(),
				Status:       "pending",
				MaxRetries:   3,
			}

			// Add to the queue instead of processing immediately
			err = image.QueueImage(imageJob)
			if err != nil {
				log.Printf("Error queueing image: %v", err)
				continue
			}

			// Append the image ID to the list
			queuedImageIds = append(queuedImageIds, imageJob.ID)
		}
	}
	// Return error if no files were successfully queued
	if len(queuedImageIds) == 0 {
		return errors.BadRequest("No valid files were processed. Make sure to provide valid image files.")
	}
	// Collect the image information with URLs
	type ImageInfo struct {
		ID       string `json:"id"`
		URL      string `json:"url"`
		FileName string `json:"file_name"`
		Status   string `json:"status"`
	}

	imageInfos := make([]ImageInfo, 0, len(queuedImageIds))

	for _, imageID := range queuedImageIds {
		// Lookup the image job from the queue
		imageJob, err := image.GetImageJob(imageID)
		if err == nil && imageJob != nil {
			// The URL is already generated when the job was queued
			imageInfos = append(imageInfos, ImageInfo{
				ID:       imageJob.ID,
				URL:      imageJob.PublicURL,
				FileName: imageJob.FileName,
				Status:   imageJob.Status,
			})
		}
	}

	// Return a response with detailed image information
	return errors.SuccessResponse(c, fiber.Map{
		"message":     fmt.Sprintf("%d images queued for processing", len(queuedImageIds)),
		"image_count": len(imageInfos),
		"images":      imageInfos,
		"status":      "processing",
	})
}
