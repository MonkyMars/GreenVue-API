package errors

import (
	"errors"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// RequestID middleware adds a unique ID to each request
func RequestID() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Generate request ID if not already set
		requestID := c.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.New().String()
			c.Set("X-Request-ID", requestID)
		}

		// Store in locals for access in other middleware/handlers
		c.Locals("requestID", requestID)

		return c.Next()
	}
}

// ErrorResponseConfig contains config for ErrorResponse middleware
type ErrorResponseConfig struct {
	// ShowDetails determines if detailed errors are shown in non-production
	DevMode bool
	// Logger is a custom logger function, defaults to log.Printf
	Logger func(format string, args ...interface{})
}

// DefaultErrorResponseConfig is the default config for ErrorResponse middleware
var DefaultErrorResponseConfig = ErrorResponseConfig{
	DevMode: false, // Set to true for development environments
	Logger:  log.Printf,
}

// ErrorHandler creates a middleware that handles errors consistently
func ErrorHandler(config ...ErrorResponseConfig) fiber.ErrorHandler {
	// Use default config if none is provided
	cfg := DefaultErrorResponseConfig
	if len(config) > 0 {
		cfg = config[0]
	}

	return func(ctx *fiber.Ctx, err error) error {
		// Early return if err is nil
		if err == nil {
			return nil
		}

		// Default error response
		statusCode := fiber.StatusInternalServerError
		errorMsg := "Internal Server Error"
		errorDetails := map[string]interface{}{}

		// Get request ID for logging
		requestID, _ := ctx.Locals("requestID").(string)
		if requestID == "" {
			requestID = "unknown"
		}

		// Handle AppError type - with safe type assertion
		var appErr *AppError
		if errors.As(err, &appErr) && appErr != nil {
			statusCode = appErr.StatusCode
			errorMsg = appErr.Message

			// Only include detailed error info in dev mode or for non-internal errors
			if cfg.DevMode || !appErr.Internal {
				if appErr.Err != nil {
					// Safely get error message
					errorDetails["error"] = appErr.Err.Error()
				} else {
					errorDetails["message"] = appErr.Message
				}
				if appErr.Field != "" {
					errorDetails["field"] = appErr.Field
				}
			}

			// Log internal server errors - with safe logging
			if appErr.Internal || appErr.StatusCode >= 500 {
				if appErr.Err != nil {
					// Safe logging that doesn't rely on appErr.Error()
					cfg.Logger("[%s] Error: %s - %s", requestID, appErr.Message, appErr.Err.Error())
				} else {
					cfg.Logger("[%s] Error: %s", requestID, appErr.Message)
				}
			}
		} else {
			// Handle fiber.Error with safe type assertion
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) && fiberErr != nil {
				statusCode = fiberErr.Code
				errorMsg = fiberErr.Message
			} else {
				// For other error types, only show details in dev mode
				if cfg.DevMode && err != nil {
					errorMsg = err.Error()
				}
				// Always log unexpected errors - safely
				if err != nil {
					cfg.Logger("[%s] Unexpected error: %s", requestID, err.Error())
				} else {
					cfg.Logger("[%s] Unknown error (nil error with non-nil wrapper)", requestID)
				}
			}
		}

		// Send JSON response
		response := fiber.Map{
			"success": false,
			"message": errorMsg,
		}

		// Add detailed error info if available
		if len(errorDetails) > 0 {
			response["error"] = errorDetails
		}

		// Add request ID to response
		response["requestId"] = requestID

		return ctx.Status(statusCode).JSON(response)
	}
}
