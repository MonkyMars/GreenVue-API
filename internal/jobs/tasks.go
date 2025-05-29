package jobs

import (
	"context"
	"greenvue/lib/email"
	"greenvue/lib/image"
	"log"
)

// Task definitions for common background jobs

// CleanupExpiredListingsOptions defines options for the cleanup job
type CleanupExpiredListingsOptions struct {
	DaysOld int `json:"days_old"` // Listings older than this will be cleaned
}

// CreateCleanupExpiredListingsJob creates a job to clean up old listings
func CreateCleanupExpiredListingsJob(opts *CleanupExpiredListingsOptions) JobFunc {
	if opts == nil {
		opts = &CleanupExpiredListingsOptions{DaysOld: 30} // Default to 30 days
	}

	return func(ctx context.Context) error {
		// Calculate the cutoff date
		// cutoffDate := time.Now().AddDate(0, 0, -opts.DaysOld)

		// This is a simplified example - in a real app you would:
		// 1. Query the database for expired listings
		// 2. Process them in batches to avoid memory issues
		// 3. Either delete them or mark them as archived

		// Example implementation:
		// client := db.GetSupabaseClient()
		// query := client.From("listings").Select("*").Lt("created_at", cutoffDate.Format(time.RFC3339))
		// ... process the results

		return nil
	}
}

// ImageProcessingOptions defines options for the image processing job
type ImageProcessingOptions struct {
	BatchSize int `json:"batch_size"` // Number of images to process in each batch
}

// CreateImageProcessingJob creates a job to process the image queue
func CreateImageProcessingJob(opts *ImageProcessingOptions) JobFunc {
	if opts == nil {
		opts = &ImageProcessingOptions{BatchSize: 5} // Reduced from 10 to prevent memory spikes
	}

	return func(ctx context.Context) error {
		if image.GlobalImageQueue == nil {
			return nil
		}

		// Only process if there are images in the queue
		if image.GlobalImageQueue.HasPendingImages() {
			pendingCount := image.GlobalImageQueue.PendingCount()
			completedCount := image.GlobalImageQueue.GetCompletedJobsCount()

			log.Printf("Processing image queue: %d pending, %d completed", pendingCount, completedCount)

			// Process the image queue
			processedURLs, err := image.GlobalImageQueue.ProcessQueue(opts.BatchSize)
			if err != nil {
				log.Printf("Error processing image queue: %v", err)
				return err
			}

			if len(processedURLs) > 0 {
				log.Printf("Successfully processed %d images", len(processedURLs))
			}
		}

		return nil
	}
}

// EmailNotificationOptions defines options for sending email notifications
type EmailNotificationOptions struct {
	Template  string `json:"template"`  // Email template name
	BatchSize int    `json:"batchSize"` // Number of emails to send in each batch
}

// CreateSendEmailNotificationsJob creates a job to send scheduled emails
func CreateSendEmailNotificationsJob(opts *EmailNotificationOptions) JobFunc {
	// if opts == nil {
	// 	opts = &EmailNotificationOptions{
	// 		Template:  "default",
	// 		BatchSize: 50,
	// 	}
	// }

	return func(ctx context.Context) error {
		// Implementation would:
		// 1. Query users who should receive notifications
		// 2. Process in batches of the specified size
		// 3. Send emails and record sending status

		return nil
	}
}

// SearchIndexOptions defines options for updating the search index
type SearchIndexOptions struct {
	FullReindex bool `json:"fullReindex"` // Whether to reindex everything
}

// CreateUpdateSearchIndexJob creates a job to update search indexes
func CreateUpdateSearchIndexJob(opts *SearchIndexOptions) JobFunc {
	if opts == nil {
		opts = &SearchIndexOptions{FullReindex: false}
	}

	return func(ctx context.Context) error {
		// Implementation would:
		// 1. Either get changed records since last index or all records
		// 2. Update search index entries

		return nil
	}
}

// EmailProcessingOptions defines options for the email processing job
type EmailProcessingOptions struct {
	BatchSize int `json:"batch_size"` // Number of emails to process in each batch
}

// CreateEmailProcessingJob creates a job to process the email queue
func CreateEmailProcessingJob(opts *EmailProcessingOptions) JobFunc {
	if opts == nil {
		opts = &EmailProcessingOptions{BatchSize: 25} // Default to processing 25 emails at a time
	}

	return func(ctx context.Context) error {
		if email.GlobalEmailQueue == nil {
			return nil
		}

		// Only process if there are emails in the queue
		if email.GlobalEmailQueue.HasPendingEmails() {
			// Process the email queue
			err := email.GlobalEmailQueue.ProcessQueue(opts.BatchSize)
			if err != nil {
				log.Printf("Error processing email queue: %v", err)
				return err
			}
		}

		return nil
	}
}
