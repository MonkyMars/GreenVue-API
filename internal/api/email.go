package api

import (
	"greenvue/internal/config"
	"greenvue/internal/jobs"
	"greenvue/lib"
	"greenvue/lib/email"
	"log"
	"time"
)

// initEmailService initializes the email service with configuration
func initEmailService(cfg *config.Config) {
	// Initialize email service using Supabase URL and key
	email.InitializeEmailService(cfg.Database.SupabaseURL, cfg.Database.SupabaseKey)
}

// setupDefaultEmailJob sets up a background job to process emails
func setupDefaultEmailJob() {
	// Set up a job to process emails every 30 seconds in non-production environments
	emailProcessingOptions := &jobs.EmailProcessingOptions{
		BatchSize: 25, // Process up to 25 emails at once
	}

	err := jobs.GlobalScheduler.AddJob(
		"process-email-queue",                                 // Job ID
		"Process Email Queue",                                 // Job Name
		"Process pending emails in queue",                     // Description
		jobs.CreateEmailProcessingJob(emailProcessingOptions), // Job function
		30*time.Second,                                        // Run every 30 seconds
	)

	if err != nil {
		log.Printf("Warning: Could not add email processing job: %v", err)
	}
}

// QueueConfirmationEmail queues a confirmation email to be sent asynchronously
func QueueConfirmationEmail(recipient, resendType string) error {
	// Create a new email object
	e := email.Email{
		ID:         lib.GenerateUUID(),
		To:         recipient,
		Type:       email.ConfirmationEmail,
		TemplateID: resendType, // Using resendType as template ID
		CreatedAt:  time.Now(),
		Status:     "pending",
		MaxRetries: 3,
	}

	// Add it to the queue
	return email.QueueEmail(e)
}

// QueuePasswordResetEmail queues a password reset email
func QueuePasswordResetEmail(recipient string) error {
	// Create a new email object
	e := email.Email{
		ID:         lib.GenerateUUID(),
		To:         recipient,
		Type:       email.PasswordResetEmail,
		CreatedAt:  time.Now(),
		Status:     "pending",
		MaxRetries: 3,
	}

	// Add it to the queue
	return email.QueueEmail(e)
}

// QueueNotificationEmail queues a generic notification email
func QueueNotificationEmail(recipient, subject, template string, variables map[string]any) error {
	// Create a new email object
	e := email.Email{
		ID:         lib.GenerateUUID(),
		To:         recipient,
		Subject:    subject,
		Type:       email.NotificationEmail,
		TemplateID: template,
		Variables:  variables,
		CreatedAt:  time.Now(),
		Status:     "pending",
		MaxRetries: 3,
	}

	// Add it to the queue
	return email.QueueEmail(e)
}
