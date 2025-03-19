package listings

import (
	"fmt"
	"greentrade-eu/internal/db"

	// "greentrade-eu/lib"
	"log"
	"mime/multipart"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	storage "github.com/supabase-community/storage-go"
)

func PostListing(c *fiber.Ctx) error {
	client := db.NewSupabaseClient()

	// Parse JSON payload
	var payload struct {
		Title         string                 `json:"title"`
		Description   string                 `json:"description"`
		Category      string                 `json:"category"`
		Condition     string                 `json:"condition"`
		Location      string                 `json:"location"`
		Price         int64                  `json:"price"`
		Negotiable    bool                   `json:"negotiable"`
		EcoScore      float32                `json:"ecoScore"`
		EcoAttributes []string               `json:"ecoAttributes"`
		ImageUrl      map[string]interface{} `json:"imageUrl"`
		Seller        db.Seller              `json:"seller"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Failed to parse JSON body: " + err.Error(),
		})
	}

	// Extract fields from parsed JSON
	title := payload.Title
	description := payload.Description
	category := payload.Category
	condition := payload.Condition
	location := payload.Location
	price := payload.Price
	negotiable := payload.Negotiable
	ecoAttributes := payload.EcoAttributes
	ecoScore := payload.EcoScore

	// Handle imageUrl safely with proper type assertion
	var imageUrl []string
	if payload.ImageUrl != nil {
		if urls, exists := payload.ImageUrl["urls"]; exists && urls != nil {
			// Try to handle the case where urls is a []interface{}
			if urlsArray, ok := urls.([]interface{}); ok {
				for _, item := range urlsArray {
					if str, ok := item.(string); ok {
						imageUrl = append(imageUrl, str)
					}
				}
			} else {
				// Log what type it actually is for debugging
				log.Printf("urls is not []interface{} but %T: %v", urls, urls)
			}
		}
	}

	seller := payload.Seller
	log.Printf("Processed imageUrl: %v", imageUrl)
	log.Printf("EcoScore: %v", ecoScore)
	// Check if seller exists
	// exists, sellerID, err := client.IsSellerInDB(seller)
	// if err != nil {
	// 	return c.Status(500).JSON(fiber.Map{
	// 		"error": "Failed to check if seller exists: " + err.Error(),
	// 	})
	// }

	// // Create seller if doesn't exist
	// if !exists {
	// 	sellerData, err := json.Marshal(seller)
	// 	if err != nil {
	// 		return c.Status(500).JSON(fiber.Map{
	// 			"error": "Failed to marshal seller data: " + err.Error(),
	// 		})
	// 	}

	// 	resp, err := client.PostRaw("sellers", sellerData)
	// 	if err != nil {
	// 		return c.Status(500).JSON(fiber.Map{
	// 			"error": "Failed to create seller: " + err.Error(),
	// 		})
	// 	}

	// 	var newSeller []db.Seller
	// 	if err := json.Unmarshal(resp, &newSeller); err != nil || len(newSeller) == 0 {
	// 		return c.Status(500).JSON(fiber.Map{"error": "Failed to parse seller response"})
	// 	}
	// 	sellerID = newSeller[0].ID
	// }

	// Create the listing
	listing := db.Listing{
		Description:   description,
		Category:      category,
		Condition:     condition,
		Price:         price,
		Location:      location,
		EcoScore:      ecoScore,
		EcoAttributes: ecoAttributes,
		Negotiable:    negotiable,
		Title:         title,
		ImageUrl:      imageUrl,
		Seller:        seller,
	}

	// Post the listing to Supabase
	listingData, err := client.POST("listings", listing)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create listing: " + err.Error(),
		})
	}

	response := map[string]any{
		"listing": listingData,
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
	// Parse listing title
	listingTitle := c.FormValue("listing_title")

	fileName := fmt.Sprintf("%s-%s", listingTitle, uuid.New().String())

	// Parse files from request
	files, err := c.MultipartForm()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Failed to parse form-data"})
	}

	uploadedURLs := []string{}

	for _, file := range files.File["file"] { // Supports multiple files under "file"
		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
		}
		defer src.Close()

		// Upload to Supabase
		publicURL, err := uploadToSupabase(sanitizeFilename(fileName), src)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Upload failed: " + err.Error()})
		}

		uploadedURLs = append(uploadedURLs, publicURL)
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
