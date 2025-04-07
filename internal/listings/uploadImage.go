package listings

import (
	"fmt"
	"greentrade-eu/lib/errors"
	"image"
	_ "image/gif"  // register GIF format
	_ "image/jpeg" // register JPEG format
	_ "image/png"  // register PNG format
	"io"

	"bytes"
	"log"
	"mime/multipart"
	"os"

	"greentrade-eu/lib"

	"github.com/chai2010/webp"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
	storage "github.com/supabase-community/storage-go"
)

func UploadHandler(c *fiber.Ctx) error {
	// Extract listing title from form data
	listingTitle := c.FormValue("listing_title")

	// Get files from formdata
	files, err := c.MultipartForm()

	if err != nil {
		return errors.BadRequest("Failed to parse form data: " + err.Error())
	}

	uploadedURLs := []string{}

	// iterate over files
	for _, file := range files.File["file"] {
		src, err := file.Open()
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to open file"})
		}
		defer src.Close()

		// Generate a unique filename
		fileName := fmt.Sprintf("%s-%s.webp", lib.SanitizeFilename(listingTitle), uuid.New().String())

		// Convert image to WebP
		webpData, err := convertToWebP(src)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to convert image: " + err.Error()})
		}

		// Upload WebP file to Supabase
		publicURL, err := uploadToSupabase(fileName, webpData)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Upload failed: " + err.Error()})
		}
		// Append the public URL to the list
		uploadedURLs = append(uploadedURLs, publicURL)
	}

	// Return the list of uploaded URLs to the frontend.
	return c.JSON(fiber.Map{"urls": uploadedURLs})
}

func convertToWebP(file multipart.File) (*bytes.Buffer, error) {
	// Ensures we read from the beginning of the file
	file.Seek(0, 0)

	// Decode image
	img, format, err := image.Decode(file)
	if err != nil {
		log.Println("Error decoding image:", err)
		log.Println("Image format:", format)
		return nil, err
	}

	// Resize while maintaining aspect ratio (max 640px)
	img = resize.Resize(0, 640, img, resize.Lanczos3)

	// Encode to WebP
	webpBuffer := new(bytes.Buffer)

	webpOptions := &webp.Options{Quality: 80}
	err = webp.Encode(webpBuffer, img, webpOptions)

	if err != nil {
		log.Println("Error encoding WebP:", err)
		return nil, err
	}

	// Return the WebP buffer
	return webpBuffer, nil
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
