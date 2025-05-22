package listings

import (
	"fmt"
	"greenvue/lib/errors"
	_ "image/jpeg" // register JPEG format
	_ "image/png"  // register PNG format
	"io"

	"bytes"
	"log"
	"os"

	"greenvue/lib"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	storage "github.com/supabase-community/storage-go"
)

func UploadHandler(c *fiber.Ctx) error {
	// Extract listing title from form data
	listingTitle := c.FormValue("listing_title")

	// Get files from formdata
	form, err := c.MultipartForm()

	if err != nil {
		log.Println("Error parsing form data:", err)
		return errors.BadRequest("Failed to parse form data: " + err.Error())
	}

	uploadedURLs := []string{}

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
					continue
				}
				defer src.Close()

				// Generate a unique filename
				fileName := fmt.Sprintf("%s-%s.webp", lib.SanitizeFilename(listingTitle), uuid.New().String())

				// Convert image to WebP
				webpData, err := convertToWebP(src)
				if err != nil {
					continue
				}

				// Upload WebP file to Supabase
				publicURL, err := uploadToSupabase(fileName, webpData)
				if err != nil {
					continue
				}

				// Append the public URL to the list
				uploadedURLs = append(uploadedURLs, publicURL)
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
				continue
			}
			defer src.Close()

			// Generate a unique filename
			fileName := fmt.Sprintf("%s-%s.webp", lib.SanitizeFilename(listingTitle), uuid.New().String())
			// Convert image to WebP
			webpData, err := convertToWebP(src)
			if err != nil {
				continue
			}

			// Upload WebP file to Supabase
			publicURL, err := uploadToSupabase(fileName, webpData)
			if err != nil {
				continue
			}

			// Append the public URL to the list
			uploadedURLs = append(uploadedURLs, publicURL)
		}
	}

	// Return error if no files were successfully processed
	if len(uploadedURLs) == 0 {
		return errors.BadRequest("No valid files were processed. Make sure to provide valid image files.")
	}

	return errors.SuccessResponse(c, fiber.Map{
		"urls": uploadedURLs,
	})
}

func uploadToSupabase(filename string, fileData *bytes.Buffer) (string, error) {
	// Get Supabase URL and key from environment variables
	supabaseUrl := os.Getenv("SUPABASE_URL")
	supabaseKey := os.Getenv("SUPABASE_SERVICE_KEY")

	// Declare the bucket name (fixed)
	bucket := "listing-images"

	// Create a new Supabase client
	client := storage.NewClient(supabaseUrl+"/storage/v1", supabaseKey, nil)

	// Set file options
	upsert := true
	cacheControl := "3600"
	fileType := "image/webp"

	fileOptions := storage.FileOptions{
		CacheControl: &cacheControl,
		Upsert:       &upsert,
		ContentType:  &fileType,
	}

	// Upload the file to Supabase
	_, err := client.UploadFile(bucket, filename, io.NopCloser(fileData), fileOptions)
	if err != nil {
		log.Println("Error uploading to Supabase:", err)
		return "", err
	}

	// Return the public URL of the uploaded file
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, bucket, filename)
	return publicURL, nil
}
