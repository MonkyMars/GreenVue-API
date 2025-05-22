# Email Background Jobs in GreenVue API

This document describes how to use the background email job system in the GreenVue API.

## Overview

The email background job system allows you to queue emails to be sent asynchronously by a background job. This has several advantages:

1. API endpoints respond quickly without waiting for emails to be sent
2. Failed emails can be automatically retried
3. Email sending can be batched to avoid overwhelming email providers
4. Email sending can be monitored and managed through the jobs API

## Architecture

The email system consists of several components:

1. **Email Service**: A service for sending different types of emails (lib/email/email.go)
2. **Email Queue**: An in-memory queue for storing emails to be sent (lib/email/email.go)
3. **Email Processing Job**: A background job that processes the email queue (internal/jobs/tasks.go)
4. **API Integration**: Functions for queuing emails within API handlers (internal/api/email.go)

## Email Types

The system supports several email types:

- `ConfirmationEmail`: For account email verification
- `PasswordResetEmail`: For password reset links
- `NotificationEmail`: For general notifications
- `WelcomeEmail`: For new user onboarding
- `MarketingEmail`: For marketing communications

## Queuing Emails

To queue an email for asynchronous sending, use the `QueueEmail` function:

```go
import (
    "greenvue/lib"
    "greenvue/lib/email"
    "time"
)

// Create an email
e := email.Email{
    ID:         lib.GenerateUUID(),
    To:         "recipient@example.com",
    Subject:    "Welcome to GreenVue",
    Type:       email.WelcomeEmail,
    TemplateID: "welcome_template",
    Variables: map[string]interface{}{
        "username": "JohnDoe",
    },
    CreatedAt:  time.Now(),
    Status:     "pending",
    MaxRetries: 3,
}

// Add to queue
err := email.QueueEmail(e)
```

## Convenience Functions

For common email types, use the convenience functions in `internal/api/email.go`:

```go
import "greenvue/internal/api"

// Queue a confirmation email
err := api.QueueConfirmationEmail("user@example.com", "signup")

// Queue a password reset email
err := api.QueuePasswordResetEmail("user@example.com")

// Queue a notification email
variables := map[string]interface{}{
    "listingName": "Vintage Chair",
    "sellerName": "Jane Doe",
}
err := api.QueueNotificationEmail(
    "buyer@example.com",
    "New Message About Your Listing",
    "new_message_template",
    variables,
)
```

## Email Processing Job

Emails are processed by a background job that runs every 30 seconds (configurable). The job:

1. Takes a batch of emails from the queue
2. Attempts to send each email
3. Updates the status of each email (sent, failed, retry)
4. Retries failed emails up to MaxRetries times

## Managing Email Jobs

Email processing can be controlled through the jobs API:

- View email job status: `GET /api/jobs/process-email-queue`
- Delete email job: `DELETE /api/jobs/process-email-queue` (stops email processing)
- Create/restart email job: `POST /api/jobs` with appropriate parameters:

```json
{
  "id": "process-email-queue",
  "name": "Process Email Queue",
  "description": "Process pending emails in queue",
  "type": "process_emails",
  "interval": "30s",
  "payload": {
    "batch_size": 25
  }
}
```

## Email Error Handling

If an email fails to send, the system:

1. Records the error message in the email's `Error` field
2. Increments the `Retries` counter
3. Sets the status to `retry` if retries remain, or `failed` if maximum retries reached
4. Logs the failure

## Implementation Details

- Emails are stored in memory, so they will be lost if the server restarts
- For production use, consider implementing persistent storage for the email queue
- The system is designed to be easily extended with additional email providers
