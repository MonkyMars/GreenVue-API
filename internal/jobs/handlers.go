package jobs

import (
	"context"
	"encoding/json"
	"greenvue/lib/errors"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Global scheduler instance
var GlobalScheduler *Scheduler

// Initialize creates a new global scheduler
func Initialize() {
	if GlobalScheduler == nil {
		GlobalScheduler = NewScheduler()
	}
}

// JobRequest represents a request to create a job
type JobRequest struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Type        string `json:"type"`
	Interval    string `json:"interval"`
	Payload     any    `json:"payload,omitempty"`
}

// JobResponse represents a job response
type JobResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Interval    string    `json:"interval"`
	LastRun     time.Time `json:"last_run,omitempty"`
	NextRun     time.Time `json:"next_run"`
	IsRunning   bool      `json:"is_running"`
}

// GetJobs returns all jobs
func GetJobs(c *fiber.Ctx) error {
	jobs := GlobalScheduler.GetJobs()
	response := make([]JobResponse, 0, len(jobs))

	for _, job := range jobs {
		response = append(response, JobResponse{
			ID:          job.ID,
			Name:        job.Name,
			Description: job.Description,
			Type:        "", // This would be populated based on job type
			Interval:    job.Interval.String(),
			LastRun:     job.LastRun,
			NextRun:     job.NextRun,
			IsRunning:   job.IsRunning,
		})
	}

	return errors.SuccessResponse(c, response)
}

// GetJobByID returns a specific job by ID
func GetJobByID(c *fiber.Ctx) error {
	id := c.Params("job_id")
	job, err := GlobalScheduler.GetJob(id)
	if err != nil {
		return err
	}

	response := JobResponse{
		ID:          job.ID,
		Name:        job.Name,
		Description: job.Description,
		Type:        "", // This would be populated based on job type
		Interval:    job.Interval.String(),
		LastRun:     job.LastRun,
		NextRun:     job.NextRun,
		IsRunning:   job.IsRunning,
	}

	return errors.SuccessResponse(c, response)
}

// CreateJob creates a new job
func CreateJob(c *fiber.Ctx) error {
	var req JobRequest
	if err := c.BodyParser(&req); err != nil {
		return errors.BadRequest("Invalid request body")
	}

	// Validate
	if req.ID == "" || req.Name == "" || req.Type == "" || req.Interval == "" {
		return errors.BadRequest("Missing required fields: id, name, type, interval")
	}

	// Parse interval
	interval, err := time.ParseDuration(req.Interval)
	if err != nil {
		return errors.BadRequest("Invalid interval format")
	} // Select job handler based on type
	var jobFunc JobFunc
	switch req.Type {
	case "cleanup_expired_listings":
		jobFunc = createCleanupExpiredListingsJob(req.Payload)
	case "update_search_index":
		jobFunc = createUpdateSearchIndexJob(req.Payload)
	case "send_notifications":
		jobFunc = createSendNotificationsJob(req.Payload)
	case "process_emails":
		jobFunc = createEmailProcessingJob(req.Payload)
	case "process_images":
		jobFunc = createImageProcessingJob(req.Payload)
	default:
		return errors.BadRequest("Unknown job type")
	}

	// Add job to scheduler
	err = GlobalScheduler.AddJob(req.ID, req.Name, req.Description, jobFunc, interval)
	if err != nil {
		return err
	}

	return errors.SuccessResponse(c, fiber.Map{
		"status":  "success",
		"message": "Job created successfully",
	})
}

// DeleteJob removes a job
func DeleteJob(c *fiber.Ctx) error {
	id := c.Params("job_id")
	err := GlobalScheduler.RemoveJob(id)
	if err != nil {
		return err
	}

	return errors.SuccessResponse(c, fiber.Map{
		"status":  "success",
		"message": "Job removed successfully",
	})
}

// Job creation helper functions
func createCleanupExpiredListingsJob(payload any) JobFunc {
	return func(ctx context.Context) error {
		// TODO: Implement cleanup logic for expired listings
		return nil
	}
}

func createUpdateSearchIndexJob(payload any) JobFunc {
	return func(ctx context.Context) error {
		// TODO: Implement search index update logic
		return nil
	}
}

func createSendNotificationsJob(payload any) JobFunc {
	return func(ctx context.Context) error {
		// TODO: Implement notification sending logic
		return nil
	}
}

func createImageProcessingJob(payload any) JobFunc {
	var options ImageProcessingOptions
	if payload != nil {
		data, _ := json.Marshal(payload)
		json.Unmarshal(data, &options)
	}
	return CreateImageProcessingJob(&options)
}

func createEmailProcessingJob(payload any) JobFunc {
	var options EmailProcessingOptions
	if payload != nil {
		data, _ := json.Marshal(payload)
		json.Unmarshal(data, &options)
	}
	return CreateEmailProcessingJob(&options)
}
