package listings

import (
	"encoding/json"
	"fmt"
	"greentrade-eu/internal/db"
	"log"
	"mime/multipart"
	"os"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	storage "github.com/supabase-community/storage-go"
)

func PostListing(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()

	// Parse multipart form
	form, err := c.MultipartForm()
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to parse multipart form: " + err.Error(),
		})
	}

	// Extract listing data from form fields
	title := form.Value["title"][0]
	description := form.Value["description"][0]
	category := form.Value["category"][0]
	condition := form.Value["condition"][0]
	location := form.Value["location"][0]

	priceStr := form.Value["price"][0]
	price, err := strconv.Atoi(priceStr)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid price format: " + err.Error(),
		})
	}

	negotiableStr := form.Value["negotiable"][0]
	negotiable := negotiableStr == "true"

	var ecoAttributes []string
	if err := json.Unmarshal([]byte(form.Value["ecoAttributes"][0]), &ecoAttributes); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid ecoAttributes format: " + err.Error(),
		})
	}

	// Parse seller data
	var seller db.Seller
	if err := json.Unmarshal([]byte(form.Value["seller"][0]), &seller); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid seller data: " + err.Error(),
		})
	}

	// Check if seller exists
	exists, sellerID, err := client.IsSellerInDB(seller)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to check if seller exists: " + err.Error(),
		})
	}

	// Create seller if doesn't exist
	if !exists {
		sellerData, err := json.Marshal(seller)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to marshal seller data: " + err.Error(),
			})
		}

		resp, err := client.PostRaw("sellers", sellerData)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Failed to create seller: " + err.Error(),
			})
		}

		var newSeller []db.Seller
		if err := json.Unmarshal(resp, &newSeller); err != nil || len(newSeller) == 0 {
			return c.Status(500).JSON(fiber.Map{"error": "Failed to parse seller response"})
		}
		sellerID = newSeller[0].ID
	}

	// Create the listing
	listing := db.Listing{
		Description:   description,
		Category:      category,
		Condition:     condition,
		Price:         price,
		Location:      location,
		EcoAttributes: ecoAttributes,
		Negotiable:    negotiable,
		Title:         title,
		SellerID:      sellerID,
	}

	// Post the listing to Supabase
	listingData, err := client.POST("listings", listing)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create listing: " + err.Error(),
		})
	}

	// Parse the created listing to get its ID
	var createdListing db.Listing
	if err := json.Unmarshal(listingData, &createdListing); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to parse created listing: " + err.Error(),
		})
	}

	response := map[string]any{
		"listing": createdListing,
	}

	return c.JSON(response)
}

func sanitizeFilename(filename string) string {
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(filename)
}

func UploadHandler(c *fiber.Ctx) error {
	// Parse listing ID
	listingID := c.FormValue("listing_id")
	if listingID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Missing listing_id"})
	}

	// Parse files from request
	files, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to parse form-data"})
	}

	uploadedURLs := []string{}
	client := db.NewSupabaseClient()

	for _, file := range files.File["file"] { // Supports multiple files under "file"
		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
		}
		defer src.Close()

		// Upload to Supabase
		publicURL, err := uploadToSupabase(file.Filename, src)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Upload failed: " + err.Error()})
		}

		uploadedURLs = append(uploadedURLs, publicURL)
	}

	// Update listing with image URLs
	_, err = client.PATCH("listings", listingID, fiber.Map{"image_urls": uploadedURLs})
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update listing"})
	}

	// Return all uploaded URLs
	return c.JSON(fiber.Map{"urls": uploadedURLs})
}

func uploadToSupabase(filename string, file multipart.File) (string, error) {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_KEY")
	sanitizedFilename := sanitizeFilename(filename)
	bucket := "listing-images"

	client := storage.NewClient("https://etyqijlhxoncggbtzsbi.supabase.co/storage/v1", supabaseKey, nil)

	upsert := true
	cacheControl := "3600"

	fileOptions := storage.FileOptions{
		CacheControl: &cacheControl,
		Upsert:       &upsert,
	}

	_, err := client.UploadFile(bucket, sanitizedFilename, file, fileOptions)

	if err != nil {
		log.Println("Error uploading to Supabase:", err)
		return "", err
	}

	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, bucket, sanitizedFilename)
	return publicURL, nil
}
