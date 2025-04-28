package errors

import (
	"github.com/gofiber/fiber/v2"
)

// HandleError simplifies error handling in route handlers
// It checks if the error is nil, and if not, passes it to the Fiber context
func HandleError(c *fiber.Ctx, err error) error {
	if err == nil {
		return nil
	}
	return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
		"success": false,
		"message": "An error occurred",
		"error":   err.Error(),
	})
}

// SuccessResponse sends a standardized success response
func SuccessResponse(c *fiber.Ctx, data any) error {
	return c.JSON(fiber.Map{
		"success": true,
		"data":    data,
	})
}

// ErrorResponse sends a standardized error response
func ErrorResponse(c *fiber.Ctx, statusCode int, message string) error {
	return c.Status(statusCode).JSON(fiber.Map{
		"success": false,
		"message": message,
	})
}

// ValidateFields is a helper to validate multiple fields and return appropriate errors
func ValidateFields(fields map[string]string) *AppError {
	for field, value := range fields {
		if value == "" {
			return ValidationError(field+" is required", field)
		}
	}
	return nil
}

// ValidateRequest validates that the request body matches the expected structure
func ValidateRequest(c *fiber.Ctx, data any) error {
	if err := c.BodyParser(data); err != nil {
		return BadRequest("Invalid request body format")
	}
	return nil
}
