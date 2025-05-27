package listings

import (
	"encoding/json"
	"greenvue/internal/db"
	"greenvue/lib"
	"greenvue/lib/errors"
	"greenvue/lib/image"
	"greenvue/lib/validation"
	"log"
	"mime/multipart"

	"github.com/gofiber/fiber/v2"
)

const (
	maxFileSize    = 10 << 20 // 10MB per file
	maxTotalFiles  = 10       // Maximum number of files
	listingFormKey = "listing"
	fileFormKey    = "file"
	titleFormKey   = "listing_title"
)

// ImageInfo represents the response structure for processed images
type ImageInfo struct {
	ID       string `json:"id"`
	URL      string `json:"url"`
	FileName string `json:"file_name"`
	Status   string `json:"status"`
}

// FileProcessor handles individual file processing
type FileProcessor struct {
	listingTitle string
}

// FormHandler handles form data extraction and validation
type FormHandler struct {
	ctx *fiber.Ctx
}

// ExtractListingTitle extracts and validates the listing title from form data
func (fh *FormHandler) ExtractListingTitle() (string, error) {
	title := fh.ctx.FormValue(titleFormKey)
	if title == "" {
		return "", errors.BadRequest("Missing listing title in form data")
	}
	return title, nil
}

// ExtractMultipartForm extracts the multipart form
func (fh *FormHandler) ExtractMultipartForm() (*multipart.Form, error) {
	form, err := fh.ctx.MultipartForm()
	if err != nil {
		return nil, errors.BadRequest("Failed to parse form data: " + err.Error())
	}
	return form, nil
}

// ExtractListingJSON extracts and parses the listing JSON from form data
func (fh *FormHandler) ExtractListingJSON(form *multipart.Form) (*lib.Listing, error) {
	jsonData, exists := form.Value[listingFormKey]
	if !exists || len(jsonData) == 0 {
		return nil, errors.BadRequest("Missing 'listing' JSON data in form")
	}

	var listing lib.Listing
	if err := json.Unmarshal([]byte(jsonData[0]), &listing); err != nil {
		return nil, errors.BadRequest("Failed to parse 'listing' JSON data: " + err.Error())
	}

	return &listing, nil
}

// buildFinalListing creates the final listing with sanitized data and image URLs
func buildFinalListing(listing *lib.Listing, imageInfos []ImageInfo) lib.Listing {
	imageUrls := make([]string, len(imageInfos))
	for i, img := range imageInfos {
		imageUrls[i] = img.URL
	}

	return lib.Listing{
		Title:         lib.SanitizeInput(listing.Title),
		Description:   lib.SanitizeInput(listing.Description),
		Category:      listing.Category,
		Condition:     listing.Condition,
		Price:         lib.SanitizePrice(listing.Price),
		Negotiable:    listing.Negotiable,
		EcoScore:      lib.CalculateEcoScore(listing.EcoAttributes),
		EcoAttributes: listing.EcoAttributes,
		ImageUrls:     imageUrls,
		SellerID:      listing.SellerID,
	}
}

// validateListing validates the listing structure
func validateListing(listing *lib.Listing) error {
	result := validation.ValidateListing(*listing)
	if result == nil {
		return errors.InternalServerError("Failed to validate listing")
	}

	if !result.Valid {
		log.Println("Listing validation failed:", result.Errors)
		return errors.BadRequest("Invalid listing")
	}

	return nil
}

// saveListing saves the listing to the database
func saveListing(client *db.SupabaseClient, listing lib.Listing) (*lib.Listing, error) {
	// Insert into database
	listingData, err := client.POST("listings", listing)
	if err != nil {
		return nil, errors.DatabaseError("Failed to create listing: " + err.Error())
	}

	// Parse response - try array first, then single object
	var listingArray []lib.Listing
	if err := json.Unmarshal(listingData, &listingArray); err == nil {
		if len(listingArray) == 0 {
			return nil, errors.InternalServerError("No listing was created")
		}
		return &listingArray[0], nil
	}

	// Try single object
	var createdListing lib.Listing
	if err := json.Unmarshal(listingData, &createdListing); err != nil {
		return nil, errors.InternalServerError("Failed to parse created listing: " + err.Error())
	}

	return &createdListing, nil
}

// PostListing handles the creation of a new listing with images
func PostListing(c *fiber.Ctx) error {
	// Get database client
	client := db.GetGlobalClient()
	if client == nil {
		return errors.InternalServerError("Failed to create client")
	}

	// Initialize handlers
	formHandler := &FormHandler{ctx: c}

	// Extract listing title
	listingTitle, err := formHandler.ExtractListingTitle()
	if err != nil {
		return err
	}

	// Extract multipart form
	form, err := formHandler.ExtractMultipartForm()
	if err != nil {
		return err
	}

	// Extract and parse listing JSON first
	listing, err := formHandler.ExtractListingJSON(form)
	if err != nil {
		return err
	}

	// Validate the listing before processing any images
	// This prevents unused images in the bucket
	if err := validateListing(listing); err != nil {
		return err
	}
	// Now that the listing is valid, process the images
	// Initialize processors
	fileProcessor := &FileProcessor{listingTitle: listingTitle}
	imageProcessor := &ImageProcessor{processor: fileProcessor}
	// Process all images
	queuedImageIds, err := imageProcessor.ProcessAllFiles(form)
	if err != nil {
		return err
	}

	// Store image information before processing to ensure URLs are available
	var imageInfos []ImageInfo
	for _, id := range queuedImageIds {
		imageJob, _ := image.GetImageJob(id)
		if imageJob != nil {
			// Use GenerateImageURL to create URLs in the same format as after upload
			if imageJob.PublicURL == "" {
				imageJob.PublicURL = image.GenerateImageURL(imageJob.FileName)
			}

			imageInfos = append(imageInfos, ImageInfo{
				ID:       imageJob.ID,
				URL:      imageJob.PublicURL,
				FileName: imageJob.FileName,
				Status:   "pending", // Will be processed soon
			})
		}
	}

	// Process the queued images in the background so we don't block the response
	if image.GlobalImageQueue != nil {
		go func() {
			// Process all queued images (using a large batch size to process all of them)
			image.GlobalImageQueue.ProcessQueue(len(queuedImageIds))
		}()
	}

	// Build final listing with sanitized data and image URLs
	finalListing := buildFinalListing(listing, imageInfos)

	// Save listing to database
	savedListing, err := saveListing(client, finalListing)
	if err != nil {
		return err
	}

	// Return success response
	return errors.SuccessResponse(c, fiber.Map{
		"listing": savedListing,
	})
}
