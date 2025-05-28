package listings

import (
	"bytes"
	"fmt"
	"greenvue/lib"
	"greenvue/lib/errors"
	img "greenvue/lib/image"
	"image"
	"image/jpeg"
	_ "image/jpeg" // register JPEG format
	"image/png"
	_ "image/png" // register PNG format
	"io"
	"log"
	"mime/multipart"
	"strings"
	"time"

	"github.com/chai2010/webp"
	"github.com/google/uuid"
	"github.com/nfnt/resize"
)

// ImageProcessor handles bulk image processing
type ImageProcessor struct {
	processor *FileProcessor
}

// validateImage checks if the data is a valid image file
func validateImage(data []byte) (string, error) {
	// Create a bytes reader from the data
	reader := bytes.NewReader(data)

	// Attempt to decode as image
	_, format, err := image.DecodeConfig(reader)
	if err != nil {
		return "", fmt.Errorf("invalid image file: failed to decode image")
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
			return "", fmt.Errorf("GIF format is not supported, please convert to JPG or PNG")
		}
		return "", fmt.Errorf("Unsupported image format: %s", format)
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
		err = jpeg.Encode(cleanBuffer, img, &jpeg.Options{Quality: 100}) // 80% quality is applied when encoding to Webp, keep at 100% to prevent double compression artifacts
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
	_, err = validateImage(imgData)
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
	webpOptions := &webp.Options{Quality: 100} // Experimental
	err = webp.Encode(webpBuffer, img, webpOptions)

	if err != nil {
		log.Println("Error encoding WebP:", err)
		return nil, err
	}

	// Return the WebP buffer
	return webpBuffer, nil
}

// ProcessFile handles the processing of a single file
func (fp *FileProcessor) ProcessFile(fileHeader *multipart.FileHeader) (*img.ImageJob, error) {
	if fileHeader.Size == 0 {
		return nil, fmt.Errorf("empty file: %s", fileHeader.Filename)
	}

	if fileHeader.Size > maxFileSize {
		return nil, fmt.Errorf("file too large: %s (%d bytes)", fileHeader.Filename, fileHeader.Size)
	}

	// Open and read file
	src, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %w", fileHeader.Filename, err)
	}
	defer src.Close()

	fileData, err := io.ReadAll(src)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %w", fileHeader.Filename, err)
	}

	// Validate image format
	_, err = validateImage(fileData)
	if err != nil {
		return nil, fmt.Errorf("invalid image format for %s: %w", fileHeader.Filename, err)
	}

	// Convert to WebP
	webpData, err := convertToWebP(bytes.NewReader(fileData))
	if err != nil {
		return nil, fmt.Errorf("failed to convert %s to WebP: %w", fileHeader.Filename, err)
	}

	// Generate unique filename
	fileName := fmt.Sprintf("%s-%s.webp", lib.SanitizeFilename(fp.listingTitle), uuid.New().String())

	// Create image job
	imageJob := &img.ImageJob{
		ID:           uuid.New().String(),
		FileName:     fileName,
		ListingTitle: fp.listingTitle,
		ImageData:    webpData.Bytes(),
		CreatedAt:    time.Now(),
		Status:       "pending",
		MaxRetries:   3,
	}

	// Queue the image
	if err := img.QueueImage(*imageJob); err != nil {
		return nil, fmt.Errorf("failed to queue image %s: %w", fileName, err)
	}

	return imageJob, nil
}

// processFileHeaders processes a slice of file headers
func (ip *ImageProcessor) processFileHeaders(fileHeaders []*multipart.FileHeader) ([]string, error) {
	var imageIds []string
	var lastError error

	for _, fileHeader := range fileHeaders {
		imageJob, err := ip.processor.ProcessFile(fileHeader)
		if err != nil {
			log.Printf("Error processing file %s: %v", fileHeader.Filename, err)
			lastError = err
			continue
		}

		imageIds = append(imageIds, imageJob.ID)
	}

	// If no files were processed successfully, return the last error
	if len(imageIds) == 0 && lastError != nil {
		return nil, lastError
	}

	return imageIds, nil
}

// ProcessAllFiles processes all files from the form, checking multiple possible keys
func (ip *ImageProcessor) ProcessAllFiles(form *multipart.Form) ([]string, error) {
	var queuedImageIds []string
	var processedFiles int

	// First, try the expected "file" key
	if fileHeaders, exists := form.File[fileFormKey]; exists && len(fileHeaders) > 0 {
		ids, err := ip.processFileHeaders(fileHeaders)
		if err != nil {
			log.Printf("Error processing files from '%s' key: %v", fileFormKey, err)
		} else {
			queuedImageIds = append(queuedImageIds, ids...)
			processedFiles += len(ids)
		}
	}

	// If no files were processed from the expected key, try all other keys
	if processedFiles == 0 {
		for key, fileHeaders := range form.File {
			if key == fileFormKey {
				continue // Already processed above
			}

			ids, err := ip.processFileHeaders(fileHeaders)
			if err != nil {
				log.Printf("Error processing files from '%s' key: %v", key, err)
				continue
			}

			queuedImageIds = append(queuedImageIds, ids...)
			processedFiles += len(ids)

			// Limit total files processed
			if processedFiles >= maxTotalFiles {
				break
			}
		}
	}

	if len(queuedImageIds) == 0 {
		if processedFiles == 0 {
			return nil, errors.BadRequest("No files were submitted. Make sure to include files in your FormData.")
		}
		return nil, errors.BadRequest("No valid files were processed. Make sure to provide valid image files.")
	}

	return queuedImageIds, nil
}

// buildImageInfos creates ImageInfo slice from image IDs
func buildImageInfos(imageIds []string) []ImageInfo {
	imageInfos := make([]ImageInfo, 0, len(imageIds))

	for _, imageID := range imageIds {
		imageJob, err := img.GetImageJob(imageID)
		if err != nil {
			log.Printf("Error retrieving image job %s: %v", imageID, err)
			continue
		}

		if imageJob != nil {
			imageInfos = append(imageInfos, ImageInfo{
				ID:       imageJob.ID,
				URL:      imageJob.PublicURL,
				FileName: imageJob.FileName,
				Status:   imageJob.Status,
			})
		}
	}

	return imageInfos
}
