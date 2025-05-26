package listings

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	_ "image/jpeg" // register JPEG format
	"image/png"
	_ "image/png" // register PNG format
	"io"
	"log"
	"strings"

	"github.com/chai2010/webp"
	"github.com/nfnt/resize"
)

// validateImage checks if the data is a valid image file
func validateImage(data []byte) (string, error) {
	// Create a bytes reader from the data
	reader := bytes.NewReader(data)

	// Attempt to decode as image
	_, format, err := image.DecodeConfig(reader)
	if err != nil {
		return "", errors.New("invalid image file: failed to decode image")
	}
	// Only allow specific formats
	format = strings.ToLower(format)
	allowedFormats := map[string]bool{
		"jpeg": true,
		"jpg":  true,
		"png":  true,
		"webp": true,
	}
	if !allowedFormats[format] {
		if format == "gif" {
			return "", errors.New("GIF format is not supported, please convert to JPG or PNG")
		}
		return "", errors.New("unsupported image format: " + format)
	}

	return format, nil
}

// stripExifMetadata removes all EXIF metadata from an image
func stripExifMetadata(imgData []byte) []byte {
	// Create a new reader from the image data
	reader := bytes.NewReader(imgData)

	// Decode the image to get its basic properties (without metadata)
	img, format, err := image.Decode(reader)
	if err != nil {
		// If we can't decode it, return the original data
		log.Println("Error decoding image for metadata stripping:", err)
		return imgData
	}

	// Create a new buffer to hold the clean image
	cleanBuffer := new(bytes.Buffer)
	// Re-encode the image based on its original format, which drops the metadata
	switch strings.ToLower(format) {
	case "jpeg", "jpg":
		// Re-encode as JPEG (this drops metadata)
		err = jpeg.Encode(cleanBuffer, img, &jpeg.Options{Quality: 85})
	case "png":
		// Re-encode as PNG (this drops metadata)
		err = png.Encode(cleanBuffer, img)
	default:
		// For unknown formats, return the original
		return imgData
	}

	if err != nil {
		log.Println("Error re-encoding image for metadata stripping:", err)
		return imgData // Return original if re-encoding fails
	}

	return cleanBuffer.Bytes()
}

// convertToWebP converts any image to WebP format with validation and metadata stripping
func convertToWebP(reader io.Reader) (*bytes.Buffer, error) {
	// Read all data from the reader
	imgData, err := io.ReadAll(reader)
	if err != nil {
		log.Println("Error reading image data:", err)
		return nil, err
	}

	// Validate that this is a supported image
	format, err := validateImage(imgData)
	if err != nil {
		log.Println("Image validation failed:", err)
		return nil, err
	}

	// Strip EXIF metadata
	cleanImgData := stripExifMetadata(imgData)

	// Decode the cleaned image
	img, _, err := image.Decode(bytes.NewReader(cleanImgData))
	if err != nil {
		log.Println("Error decoding cleaned image:", err)
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

	log.Printf("Successfully converted %s image to WebP with metadata stripped", format)

	// Return the WebP buffer
	return webpBuffer, nil
}
