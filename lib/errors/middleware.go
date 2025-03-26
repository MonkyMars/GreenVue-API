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
		// Get the error that might be wrapped in a *fiber.Error
		var originalErr = err
		var fiberErr *fiber.Error
		if errors.As(err, &fiberErr) {
			originalErr = fiberErr
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

		// Handle AppError type
		var appErr *AppError
		if errors.As(originalErr, &appErr) {
			statusCode = appErr.StatusCode
			errorMsg = appErr.Message

			// Only include detailed error info in dev mode or for non-internal errors
			if cfg.DevMode || !appErr.Internal {
				errorDetails["error"] = appErr.Err.Error()
				if appErr.Field != "" {
					errorDetails["field"] = appErr.Field
				}
			}

			// Log internal server errors
			if appErr.Internal || appErr.StatusCode >= 500 {
				cfg.Logger("[%s] Error: %v", requestID, appErr.Error())
			}
		} else {
			// Handle fiber.Error
			if fiberErr != nil {
				statusCode = fiberErr.Code
				errorMsg = fiberErr.Message
			} else {
				// For other error types, only show details in dev mode
				if cfg.DevMode {
					errorMsg = err.Error()
				}
				// Always log unexpected errors
				cfg.Logger("[%s] Unexpected error: %v", requestID, err.Error())
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
