# Image Processing Background Jobs in GreenVue API

This document describes the image processing system in the GreenVue API, which exclusively uses background jobs for all image handling.

## Overview

The GreenVue API implements a fully asynchronous image processing system with the following advantages:

1. API endpoints respond quickly without waiting for image processing
2. Failed image processing can be automatically retried
3. Image processing can be batched to avoid overwhelming the server
4. Image processing can be monitored and managed through the jobs API
5. Immediate URL availability while processing happens in background

## Architecture

The image processing system consists of several components:

1. **Image Queue**: An in-memory queue for storing images to be processed (`lib/image/image.go`)
2. **Image Processing Job**: A background job that processes the image queue (`internal/jobs/tasks.go`)
3. **API Integration**: Functions for queuing images within API handlers (`internal/listings/queuedUpload.go`)

## Processing Steps

The image processing pipeline includes:

1. **WebP Conversion**: Converting uploaded images to the WebP format for smaller file sizes
2. **Image Resizing**: Resizing images to a maximum height (640px) while maintaining aspect ratio
3. **Storage Upload**: Uploading the processed images to Supabase storage

## Image Job Creation

When an image is uploaded via the `/api/upload/listing_image` endpoint, it is:

1. Read into memory
2. Converted to WebP format
3. Added to the image processing queue with status "pending"
4. Public URL is generated and returned immediately to the client

Even though the image is still queued for processing, the endpoint returns the expected final URLs. This allows your application to immediately reference these URLs in your database or UI, knowing that they will be valid once the background processing completes.

### Example Response

```json
{
  "status": "success",
  "data": {
    "message": "2 images queued for processing",
    "image_count": 2,
    "images": [
      {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "url": "https://example.supabase.co/storage/v1/object/public/listing-images/my-listing-550e8400.webp",
        "file_name": "my-listing-550e8400.webp",
        "status": "pending"
      },
      {
        "id": "6ba7b810-9dad-11d1-80b4-00c04fd430c8",
        "url": "https://example.supabase.co/storage/v1/object/public/listing-images/my-listing-6ba7b810.webp",
        "file_name": "my-listing-6ba7b810.webp",
        "status": "pending"
      }
    ],
    "status": "processing"
  }
}
```

## Image Processing Job

Images are processed by a background job that runs every 10 seconds. The job:

1. Takes a batch of images from the queue (default: 10 at a time)
2. Uploads each image to Supabase storage
3. Updates the status of each image (processed, failed, retry)
4. Retries failed images up to MaxRetries times (default: 3)

## Managing Image Jobs

Image processing can be controlled through the jobs API:

- View image job status: `GET /api/jobs/process-image-queue`
- Delete image job: `DELETE /api/jobs/process-image-queue` (stops image processing)
- Create/restart image job: `POST /api/jobs` with appropriate parameters:

```json
{
  "id": "process-image-queue",
  "name": "Process Image Queue",
  "description": "Process pending images in queue",
  "type": "process_images",
  "interval": "10s",
  "payload": {
    "batch_size": 10
  }
}
```

## Debug Endpoints

For testing and debugging the image processing system, the following endpoints are available in non-production environments:

### Check Image Queue Status

```
GET /debug/image-queue-status
```

Returns information about the current state of the image queue:

```json
{
  "pending_count": 3,
  "queue_status": {
    "initialized": true,
    "status": "active"
  }
}
```

### Test Image Upload

```
POST /debug/upload-test-image
```

Payload:

```json
{
  "listing_title": "Test Listing",
  "image_path": "/path/to/test/image.jpg",
  "process_now": true
}
```

This endpoint allows you to test the image processing queue with a file from the server's filesystem. If `process_now` is set to `true`, it will immediately process the queued image.

## Image Error Handling

If image processing fails, the system:

1. Records the error message in the image's `Error` field
2. Increments the `Retries` counter
3. Sets the status to `retry` if retries remain, or `failed` if maximum retries reached
4. Logs the failure

## Implementation Details

- Images are stored in memory for efficient processing
- Basic queue persistence is implemented (saves pending jobs on shutdown and restores on startup)
- For production use, consider implementing more robust persistent storage
- The system is designed to be easily extended with additional image processing steps
