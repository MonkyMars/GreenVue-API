package api

import (
	"greenvue/lib"
	"greenvue/lib/email"
	"greenvue/lib/errors"
	"time"

	"github.com/gofiber/fiber/v2"
)

// TestEmailHandler is a debug handler to test the email system
func TestEmailHandler(c *fiber.Ctx) error {
	// Parse request body
	var req struct {
		Email    string                 `json:"email"`
		Subject  string                 `json:"subject"`
		Type     string                 `json:"type"`     // confirmation, password_reset, notification, etc.
		Template string                 `json:"template"` // Only used for certain email types
		Vars     map[string]interface{} `json:"vars"`     // Template variables
	}

	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	// Validate
	if req.Email == "" {
		return errors.BadRequest("Email is required")
	}

	if req.Subject == "" {
		req.Subject = "Test Email from GreenVue API"
	}

	// Determine email type
	var emailType email.EmailType
	switch req.Type {
	case "confirmation":
		emailType = email.ConfirmationEmail
	case "password_reset":
		emailType = email.PasswordResetEmail
	case "notification":
		emailType = email.NotificationEmail
	case "welcome":
		emailType = email.WelcomeEmail
	case "marketing":
		emailType = email.MarketingEmail
	default:
		emailType = email.NotificationEmail
	}

	// Create test email
	testEmail := email.Email{
		ID:         lib.GenerateUUID(),
		To:         req.Email,
		Subject:    req.Subject,
		Type:       emailType,
		TemplateID: req.Template,
		Variables:  req.Vars,
		CreatedAt:  time.Now(),
		Status:     "pending",
		MaxRetries: 3,
	}
	// Queue the email
	err := email.QueueEmail(testEmail)
	if err != nil {
		return errors.InternalServerError("Failed to queue test email: " + err.Error())
	}

	// Process the queue immediately for test emails
	if req.Type == "immediate" {
		go func() {
			if email.GlobalEmailQueue != nil {
				email.GlobalEmailQueue.ProcessQueue(10)
			}
		}()
	}

	return errors.SuccessResponse(c, "Test email queued successfully with ID: "+testEmail.ID)
}

// GetEmailQueueStatusHandler returns the status of the email queue
func GetEmailQueueStatusHandler(c *fiber.Ctx) error {
	if email.GlobalEmailQueue == nil {
		return errors.InternalServerError("Email queue not initialized")
	}

	// Get queue status
	pendingCount := 0
	if email.GlobalEmailQueue.HasPendingEmails() {
		pendingCount = email.GlobalEmailQueue.PendingCount()
	}

	return errors.SuccessResponse(c, fiber.Map{
		"queueInitialized":   email.GlobalEmailQueue != nil,
		"serviceInitialized": email.DefaultEmailService != nil,
		"pendingEmails":      pendingCount,
	})
}
