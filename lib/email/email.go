package email

import (
	"encoding/json"
	"fmt"
	"greenvue/lib"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

// EmailType represents different types of emails that can be sent
type EmailType string

const (
	ConfirmationEmail  EmailType = "confirmation"
	PasswordResetEmail EmailType = "password_reset"
	NotificationEmail  EmailType = "notification"
	WelcomeEmail       EmailType = "welcome"
	MarketingEmail     EmailType = "marketing"
)

// Email represents an email message to be sent
type Email struct {
	ID          string         `json:"id"`
	To          string         `json:"to"`
	Subject     string         `json:"subject"`
	Type        EmailType      `json:"type"`
	TemplateID  string         `json:"template_id,omitempty"`
	Variables   map[string]any `json:"variables,omitempty"`
	HTMLContent string         `json:"html_content,omitempty"`
	TextContent string         `json:"text_content,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	SentAt      *time.Time     `json:"sent_at,omitempty"`
	Status      string         `json:"status"`
	Retries     int            `json:"retries"`
	MaxRetries  int            `json:"max_retries"`
	Error       string         `json:"error,omitempty"`
}

// Service is an interface for sending emails
type Service interface {
	SendEmail(email *Email) error
	SendConfirmationEmail(email string, resendType string) error
}

// SupabaseEmailService sends emails using Supabase Auth API
type SupabaseEmailService struct {
	BaseURL string
	APIKey  string
	Client  *resty.Client
}

// NewSupabaseEmailService creates a new email service
func NewSupabaseEmailService(baseURL, apiKey string) *SupabaseEmailService {
	client := resty.New().
		SetTimeout(10*time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("apikey", apiKey).
		SetHeader("Authorization", "Bearer "+apiKey)

	return &SupabaseEmailService{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Client:  client,
	}
}

// SendEmail sends an email using the appropriate method based on type
func (s *SupabaseEmailService) SendEmail(email *Email) error {
	switch email.Type {
	case ConfirmationEmail:
		return s.SendConfirmationEmail(email.To, email.TemplateID)
	case PasswordResetEmail:
		return s.SendPasswordResetEmail(email.To)
	default:
		return s.sendGenericEmail(email)
	}
}

// SendConfirmationEmail specifically sends a confirmation email
func (s *SupabaseEmailService) SendConfirmationEmail(email string, resendType string) error {
	url := fmt.Sprintf("%s/auth/v1/resend", s.BaseURL)

	// Create request payload
	payload := map[string]string{
		"type":  resendType,
		"email": email,
	}

	client := resty.New().
		SetTimeout(10*time.Second).
		SetHeader("Content-Type", "application/json").
		SetHeader("apikey", s.APIKey).
		SetHeader("Authorization", "Bearer "+s.APIKey)

	resp, err := client.R().
		SetBody(payload).
		Post(url)

	if err != nil {
		log.Printf("Error sending confirmation email to %s: %v", email, err)
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode() != http.StatusOK {
		log.Printf("API error sending confirmation email to %s: %s", email, string(resp.Body()))
		return fmt.Errorf("resend confirmation email failed: %s", string(resp.Body()))
	}

	return nil
}

// SendPasswordResetEmail sends a password reset email
func (s *SupabaseEmailService) SendPasswordResetEmail(email string) error {
	redirectUrl := os.Getenv("URL") + "/reset_password"
	url := fmt.Sprintf("%s/auth/v1/recover?redirect_to=%s", s.BaseURL, url.QueryEscape(redirectUrl))

	log.Println(url)
	fmt.Println(url)

	// Create request payload
	payload := map[string]string{
		"email": email,
	}

	resp, err := s.Client.R().
		SetBody(payload).
		Post(url)

	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode() != http.StatusOK {
		return fmt.Errorf("send reset password email failed: %s", string(resp.Body()))
	}

	return nil
}

// sendGenericEmail is a placeholder for sending custom emails
// In a real implementation, this would use an email service like SendGrid, MailChimp, etc.
func (s *SupabaseEmailService) sendGenericEmail(email *Email) error {
	// Implementation would depend on your email provider
	return nil
}

// Queue represents a queue of emails to send
type Queue struct {
	pendingEmails []Email
	mu            sync.Mutex // Using a proper mutex for thread safety
	emailService  Service
}

// NewEmailQueue creates a new email queue
func NewEmailQueue(service Service) *Queue {
	return &Queue{
		pendingEmails: make([]Email, 0),
		emailService:  service,
	}
}

// AddToQueue adds an email to the queue
func (q *Queue) AddToQueue(email Email) {
	q.mu.Lock()
	defer q.mu.Unlock()

	// Set default values if not provided
	if email.CreatedAt.IsZero() {
		email.CreatedAt = time.Now()
	}
	if email.Status == "" {
		email.Status = "pending"
	}
	if email.MaxRetries == 0 {
		email.MaxRetries = 3
	}

	q.pendingEmails = append(q.pendingEmails, email)
}

// HasPendingEmails checks if there are any pending emails in the queue
func (q *Queue) HasPendingEmails() bool {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pendingEmails) > 0
}

// PendingCount returns the number of pending emails in the queue
func (q *Queue) PendingCount() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.pendingEmails)
}

// ProcessQueue processes the email queue
func (q *Queue) ProcessQueue(batchSize int) error {
	q.mu.Lock()

	if len(q.pendingEmails) == 0 {
		q.mu.Unlock()
		return nil
	}

	// Process emails in batches
	endIdx := lib.Min(batchSize, len(q.pendingEmails))
	batch := q.pendingEmails[:endIdx]
	q.pendingEmails = q.pendingEmails[endIdx:]
	// Release the lock while processing
	q.mu.Unlock()

	for i := range batch {
		email := &batch[i]

		if email.Status != "pending" && email.Status != "retry" {
			continue
		}

		err := q.emailService.SendEmail(email)
		now := time.Now()

		if err != nil {
			email.Retries++
			email.Error = err.Error()

			if email.Retries >= email.MaxRetries {
				email.Status = "failed"
				log.Printf("Failed to send email to %s after %d retries: %v",
					email.To, email.Retries, err)
			} else {
				email.Status = "retry"
				// Only log retry attempts for clarity
				log.Printf("Email to %s failed, will retry (attempt %d/%d)",
					email.To, email.Retries, email.MaxRetries)
			}
		} else {
			email.Status = "sent"
			email.SentAt = &now
		}
	}

	return nil
}

// Global instances for the application to use
var (
	DefaultEmailService Service
	GlobalEmailQueue    *Queue
)

// InitializeEmailService sets up the default email service and queue
func InitializeEmailService(baseURL, apiKey string) {
	DefaultEmailService = NewSupabaseEmailService(baseURL, apiKey)
	GlobalEmailQueue = NewEmailQueue(DefaultEmailService)
}

// QueueEmail is a convenience method to add an email to the global queue
func QueueEmail(email Email) error {
	if GlobalEmailQueue == nil {
		return fmt.Errorf("email queue not initialized")
	}
	GlobalEmailQueue.AddToQueue(email)
	return nil
}

// MarshalEmail serializes an email to JSON
func MarshalEmail(email Email) (string, error) {
	data, err := json.Marshal(email)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// UnmarshalEmail deserializes an email from JSON
func UnmarshalEmail(data string) (Email, error) {
	var email Email
	err := json.Unmarshal([]byte(data), &email)
	if err != nil {
		return Email{}, err
	}
	return email, nil
}
