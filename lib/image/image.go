package image

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	storage "github.com/supabase-community/storage-go"
)

// ImageJob represents an image processing job
type ImageJob struct {
	ID           string     `json:"id"`
	FileName     string     `json:"file_name"`
	ListingTitle string     `json:"listing_title"`
	ImageData    []byte     `json:"image_data,omitempty"` // Will be cleared after processing to prevent memory leaks
	CreatedAt    time.Time  `json:"created_at"`
	ProcessedAt  *time.Time `json:"processed_at,omitempty"`
	Status       string     `json:"status"`
	PublicURL    string     `json:"public_url,omitempty"`
	Retries      int        `json:"retries"`
	MaxRetries   int        `json:"max_retries"`
	Error        string     `json:"error,omitempty"`
}

// Queue manages a queue of images to be processed
type Queue struct {
	pendingImages   []ImageJob
	completedImages []ImageJob // Store completed jobs separately for cleanup
	mu              sync.Mutex
	persistPath     string    // Path to persist the queue
	maxCompleted    int       // Maximum number of completed jobs to keep
	lastCleanup     time.Time // Last time cleanup was performed
}

// Global image queue
var GlobalImageQueue *Queue

// Initialize the global image queue
func InitializeImageQueue() {
	if GlobalImageQueue == nil {
		GlobalImageQueue = NewImageQueue()
		// Try to restore the queue from disk if it exists
		GlobalImageQueue.RestoreFromDisk()
	}
}

// NewImageQueue creates a new image queue
func NewImageQueue() *Queue {
	return &Queue{
		pendingImages:   make([]ImageJob, 0),
		completedImages: make([]ImageJob, 0),
		persistPath:     "image_queue_backup.json", // Default path
		maxCompleted:    100,                       // Keep max 100 completed jobs for debugging
		lastCleanup:     time.Now(),
	}
}

// SetPersistPath sets the path for persisting the queue
func (q *Queue) SetPersistPath(path string) {
	q.persistPath = path
}

// QueueImage adds an image to the processing queue
func QueueImage(imageJob ImageJob) error {
	if GlobalImageQueue == nil {
		return fmt.Errorf("image queue not initialized")
	}

	GlobalImageQueue.AddToQueue(imageJob)
	return nil
}

// GetImageJob retrieves an image job from the queue by ID
func GetImageJob(id string) (*ImageJob, error) {
	if GlobalImageQueue == nil {
		return nil, fmt.Errorf("image queue not initialized")
	}

	return GlobalImageQueue.GetJobByID(id)
}

// AddToQueue adds an image to the queue
func (q *Queue) AddToQueue(imageJob ImageJob) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Set default values if not provided
	if imageJob.CreatedAt.IsZero() {
		imageJob.CreatedAt = time.Now()
	}
	if imageJob.Status == "" {
		imageJob.Status = "pending"
	}
	if imageJob.MaxRetries == 0 {
		imageJob.MaxRetries = 3
	}
	if imageJob.ID == "" {
		imageJob.ID = uuid.New().String()
	}

	// Generate the expected URL for this image
	if imageJob.PublicURL == "" {
		imageJob.PublicURL = GenerateImageURL(imageJob.FileName)
	}

	q.pendingImages = append(q.pendingImages, imageJob)
}

// HasPendingImages checks if there are any pending images in the queue
func (q *Queue) HasPendingImages() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pendingImages) > 0
}

// PendingCount returns the number of pending images in the queue
func (q *Queue) PendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pendingImages)
}

// GetJobByID returns a job by its ID
func (q *Queue) GetJobByID(id string) (*ImageJob, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Check pending jobs first
	for i, job := range q.pendingImages {
		if job.ID == id {
			// Return a copy to prevent external modification
			jobCopy := q.pendingImages[i]
			return &jobCopy, nil
		}
	}

	// Check completed jobs
	for i, job := range q.completedImages {
		if job.ID == id {
			// Return a copy to prevent external modification
			jobCopy := q.completedImages[i]
			return &jobCopy, nil
		}
	}

	return nil, fmt.Errorf("job not found with ID: %s", id)
}

// ProcessQueue processes the image queue with memory leak prevention
func (q *Queue) ProcessQueue(batchSize int) ([]string, error) {
	q.mu.Lock()

	if len(q.pendingImages) == 0 {
		q.mu.Unlock()
		// Perform cleanup if it's been a while
		q.cleanupCompletedJobs()
		return nil, nil
	}

	// Process images in batches
	endIdx := min(batchSize, len(q.pendingImages))
	batch := make([]ImageJob, endIdx)
	copy(batch, q.pendingImages[:endIdx])
	q.pendingImages = q.pendingImages[endIdx:]

	// Release the lock while processing
	q.mu.Unlock()

	processedURLs := make([]string, 0, len(batch))

	for i := range batch {
		imageJob := &batch[i]

		if imageJob.Status != "pending" && imageJob.Status != "retry" {
			continue
		}

		// Convert image data to a buffer that we can pass to the processor
		buffer := bytes.NewBuffer(imageJob.ImageData)

		// Upload to Supabase
		publicURL, err := UploadToSupabase(imageJob.FileName, buffer)
		now := time.Now()

		if err != nil {
			imageJob.Retries++
			imageJob.Error = err.Error()

			if imageJob.Retries >= imageJob.MaxRetries {
				imageJob.Status = "failed"
				// Clear image data to free memory
				imageJob.ImageData = nil
				log.Printf("Failed to process image %s after %d retries: %v",
					imageJob.ID, imageJob.Retries, err)
			} else {
				imageJob.Status = "retry"
				log.Printf("Image processing for %s failed, will retry (attempt %d/%d)",
					imageJob.ID, imageJob.Retries, imageJob.MaxRetries)
				// Re-queue for retry but don't clear image data yet
				q.mu.Lock()
				q.pendingImages = append(q.pendingImages, *imageJob)
				q.mu.Unlock()
				continue
			}
		} else {
			imageJob.Status = "processed"
			imageJob.ProcessedAt = &now
			imageJob.PublicURL = publicURL
			// Clear image data to free memory after successful upload
			imageJob.ImageData = nil
			processedURLs = append(processedURLs, publicURL)
		}

		// Move completed/failed jobs to completed queue
		q.mu.Lock()
		q.completedImages = append(q.completedImages, *imageJob)
		q.mu.Unlock()
	}

	// Cleanup old completed jobs periodically
	q.cleanupCompletedJobs()

	return processedURLs, nil
}

// UploadToSupabase uploads an image buffer to Supabase storage
func UploadToSupabase(filename string, fileData *bytes.Buffer) (string, error) {
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

	// Create a reader that will automatically close when upload is done
	reader := io.NopCloser(bytes.NewReader(fileData.Bytes()))

	// Upload the file to Supabase
	_, err := client.UploadFile(bucket, filename, reader, fileOptions)
	if err != nil {
		log.Println("Error uploading to Supabase:", err)
		return "", err
	}

	// Return the public URL of the uploaded file
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, bucket, filename)
	return publicURL, nil
}

// GenerateImageURL generates the public URL for an image without uploading it
// This can be used to return URLs immediately before background processing
func GenerateImageURL(filename string) string {
	supabaseUrl := os.Getenv("SUPABASE_URL")
	bucket := "listing-images"
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", supabaseUrl, bucket, filename)
}

// PersistToDisk saves the current queue state to a JSON file
func (q *Queue) PersistToDisk() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Skip if no images or no persist path
	if len(q.pendingImages) == 0 || q.persistPath == "" {
		return nil
	}

	// Marshal the queue to JSON
	data, err := json.Marshal(q.pendingImages)
	if err != nil {
		return fmt.Errorf("error marshaling image queue: %w", err)
	}

	// Write to file
	err = os.WriteFile(q.persistPath, data, 0644)
	if err != nil {
		return fmt.Errorf("error writing image queue to disk: %w", err)
	}

	log.Printf("Persisted %d pending images to %s", len(q.pendingImages), q.persistPath)
	return nil
}

// RestoreFromDisk loads the queue state from a JSON file if it exists
func (q *Queue) RestoreFromDisk() error {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Skip if no persist path or file doesn't exist
	if q.persistPath == "" {
		return nil
	}

	data, err := os.ReadFile(q.persistPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, not an error
			return nil
		}
		return fmt.Errorf("error reading image queue from disk: %w", err)
	}

	// Unmarshal the queue
	var images []ImageJob
	err = json.Unmarshal(data, &images)
	if err != nil {
		return fmt.Errorf("error unmarshaling image queue: %w", err)
	}

	// Add the images to the queue
	q.pendingImages = images
	log.Printf("Restored %d pending images from %s", len(images), q.persistPath)

	// Delete the file after successful restore
	os.Remove(q.persistPath)

	return nil
}

// cleanupCompletedJobs removes old completed jobs to prevent memory leaks
func (q *Queue) cleanupCompletedJobs() {
	// Only cleanup every 5 minutes to avoid excessive work
	if time.Since(q.lastCleanup) < 5*time.Minute {
		return
	}

	q.mu.Lock()
	defer q.mu.Unlock()

	// Remove excess completed jobs, keeping only the most recent ones
	if len(q.completedImages) > q.maxCompleted {
		// Keep only the last maxCompleted jobs
		toRemove := len(q.completedImages) - q.maxCompleted

		// Clear image data from jobs being removed to free memory
		for i := 0; i < toRemove; i++ {
			q.completedImages[i].ImageData = nil
		}

		// Keep only recent jobs
		q.completedImages = q.completedImages[toRemove:]
		log.Printf("Cleaned up %d old completed image jobs", toRemove)
	}

	// Also cleanup very old jobs (older than 24 hours) regardless of count
	cutoffTime := time.Now().Add(-24 * time.Hour)
	newCompleted := make([]ImageJob, 0, len(q.completedImages))
	cleanedCount := 0

	for _, job := range q.completedImages {
		if job.ProcessedAt == nil || job.ProcessedAt.After(cutoffTime) {
			newCompleted = append(newCompleted, job)
		} else {
			// Clear image data before removing
			job.ImageData = nil
			cleanedCount++
		}
	}

	if cleanedCount > 0 {
		q.completedImages = newCompleted
		log.Printf("Cleaned up %d jobs older than 24 hours", cleanedCount)
	}

	q.lastCleanup = time.Now()
}

// GetCompletedJobsCount returns the number of completed jobs
func (q *Queue) GetCompletedJobsCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.completedImages)
}

// ForceCleanup manually triggers cleanup of completed jobs
func (q *Queue) ForceCleanup() {
	q.mu.Lock()
	q.lastCleanup = time.Time{} // Reset to force cleanup
	q.mu.Unlock()
	q.cleanupCompletedJobs()
}
