package api

import (
	"greenvue/lib/errors"
	"greenvue/lib/image"

	"github.com/gofiber/fiber/v2"
)

// GetImageQueueStatusHandler provides information about the current status of the image processing queue
func GetImageQueueStatusHandler(c *fiber.Ctx) error {
	if image.GlobalImageQueue == nil {
		return errors.InternalServerError("Image queue not initialized")
	}

	pendingCount := image.GlobalImageQueue.PendingCount()

	return errors.SuccessResponse(c, fiber.Map{
		"pending_count": pendingCount,
		"queue_status": map[string]any{
			"initialized": image.GlobalImageQueue != nil,
			"status":      "active",
		},
	})
}
