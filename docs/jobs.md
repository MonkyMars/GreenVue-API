# Background Jobs in GreenVue API

This document explains how to use the background job system in the GreenVue API.

## Overview

The background job system allows you to run scheduled tasks in the background while using the same HTTP port as your main API. Jobs are managed through REST endpoints and can be scheduled to run at specific intervals.

## API Endpoints

All job endpoints are protected by authentication and available at the `/api/jobs` path:

- `GET /api/jobs` - List all background jobs
- `GET /api/jobs/:job_id` - Get details for a specific job
- `POST /api/jobs` - Create a new background job
- `DELETE /api/jobs/:job_id` - Delete a job

## Creating a Job

To create a background job, send a POST request to `/api/jobs` with a payload like:

```json
{
  "id": "cleanup-expired-listings",
  "name": "Cleanup Expired Listings",
  "description": "Remove listings that are older than 30 days",
  "type": "cleanup_expired_listings",
  "interval": "24h",
  "payload": {
    "days_old": 30
  }
}
```

### Available Job Types

The system includes several predefined job types:

1. `cleanup_expired_listings` - Removes old listings

   - Parameters: `days_old` (int) - Age in days of listings to clean up

2. `send_notifications` - Sends scheduled notifications

   - Parameters:
     - `template` (string) - Email template name
     - `batchSize` (int) - Batch size for processing

3. `update_search_index` - Updates search indexes
   - Parameters:
     - `fullReindex` (boolean) - Whether to perform a full reindex

## Interval Format

The interval is specified using Go's duration format:

- `5m`: 5 minutes
- `1h`: 1 hour
- `24h`: 24 hours (1 day)
- `168h`: 1 week

## Creating Custom Job Types

To add a new job type:

1. Define a new job function in `internal/jobs/tasks.go`
2. Add a new case in the `CreateJob` function in `internal/jobs/handlers.go`

Example:

```go
// In tasks.go
type CustomJobOptions struct {
    Parameter1 string `json:"parameter1"`
    Parameter2 int    `json:"parameter2"`
}

func CreateCustomJob(opts *CustomJobOptions) JobFunc {
    if opts == nil {
        opts = &CustomJobOptions{
            Parameter1: "default",
            Parameter2: 10,
        }
    }

    return func(ctx context.Context) error {
        // Implement your job logic here
        return nil
    }
}

// Then in handlers.go, add to the switch statement:
case "custom_job":
    jobFunc = createCustomJob(req.Payload)
```

## Sample Usage

Here's an example of creating a job to clean up expired listings:

```bash
curl -X POST http://localhost:8080/api/jobs \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "id": "cleanup-expired-listings",
    "name": "Cleanup Expired Listings",
    "description": "Remove listings older than 30 days",
    "type": "cleanup_expired_listings",
    "interval": "24h",
    "payload": {
      "days_old": 30
    }
  }'
```

And to list all jobs:

```bash
curl -X GET http://localhost:8080/api/jobs \
  -H "Authorization: Bearer YOUR_TOKEN"
```
