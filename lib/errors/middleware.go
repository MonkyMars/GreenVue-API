package errors

import (
	"errors"
	"fmt"
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
	// Initialize final config with defaults
	finalCfg := DefaultErrorResponseConfig

	// Apply user config if provided
	if len(config) > 0 {
		userConfig := config[0]
		// Always apply DevMode from user config if provided
		finalCfg.DevMode = userConfig.DevMode
		// Only apply Logger from user config if it's not nil
		if userConfig.Logger != nil {
			finalCfg.Logger = userConfig.Logger
		}
	}

	// Final safety check: ensure logger is never nil *before* returning the handler
	if finalCfg.Logger == nil {
		log.Println("Warning: ErrorHandler logger was nil, falling back to default log.Printf")
		finalCfg.Logger = log.Printf
	}

	// Return the actual error handling closure, capturing the correctly configured finalCfg
	return func(ctx *fiber.Ctx, err error) error {
		cfg := finalCfg // Use the captured final config

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
		if errors.As(err, &appErr) {
			// Safety check to prevent nil pointer dereference
			if appErr == nil {
				cfg.Logger("[%s] Application error is nil after type assertion", requestID)
				goto SendResponse // Use default 500 error
			}

			statusCode = appErr.StatusCode
			errorMsg = appErr.Message

			// Only include detailed error info in dev mode or for non-internal errors
			if cfg.DevMode || !appErr.Internal {
				var underlyingErrMsg string
				// Safely get underlying error message
				if appErr.Err != nil {
					func() {
						defer func() {
							if r := recover(); r != nil {
								// Use the guaranteed non-nil logger
								cfg.Logger("[%s] Panic recovering from appErr.Err.Error(): %v", requestID, r)
								underlyingErrMsg = fmt.Sprintf("panic getting underlying error: %v", r)
							}
						}()
						underlyingErrMsg = appErr.Err.Error()
					}()
					errorDetails["error"] = underlyingErrMsg
				}

				if appErr.Field != "" {
					errorDetails["field"] = appErr.Field
				}
			}

			// Log internal server errors
			if appErr.Internal || appErr.StatusCode >= 500 {
				errDetailsLog := ""
				if appErr.Err != nil {
					// Use the safely obtained message for logging
					if errMsg, ok := errorDetails["error"].(string); ok {
						errDetailsLog = errMsg
					} else {
						// Fallback if casting failed (shouldn't happen often)
						errDetailsLog = fmt.Sprintf("underlying error type: %T", appErr.Err)
					}
				}

				if errDetailsLog != "" {
					// Use the guaranteed non-nil logger
					cfg.Logger("[%s] Error: %s - %s", requestID, appErr.Message, errDetailsLog)
				} else {
					// Use the guaranteed non-nil logger
					cfg.Logger("[%s] Error: %s", requestID, appErr.Message)
				}
			}
		} else {
			// Handle fiber.Error with safe type assertion
			var fiberErr *fiber.Error
			if errors.As(err, &fiberErr) {
				// Safety check for fiber error
				if fiberErr == nil {
					// Use the guaranteed non-nil logger
					cfg.Logger("[%s] Fiber error is nil after type assertion", requestID)
					goto SendResponse // Use default 500 error
				}

				statusCode = fiberErr.Code
				errorMsg = fiberErr.Message
			} else {
				// Handle other error types safely
				var errorString string = "An unexpected error occurred" // Default message for unexpected errors

				// Safely attempt to get the error string
				func() {
					defer func() {
						if r := recover(); r != nil {
							// Use the guaranteed non-nil logger
							cfg.Logger("[%s] Panic recovering from err.Error() for type %T: %v", requestID, err, r)
							errorString = fmt.Sprintf("error retrieving message from type %T", err)
						}
					}()
					// This is the line that might panic
					errorString = err.Error()
				}()

				// Log the unexpected error using the obtained/default string
				// Use the guaranteed non-nil logger
				cfg.Logger("[%s] Unexpected error: %s", requestID, errorString)

				// Set response message based on DevMode
				if cfg.DevMode {
					errorMsg = errorString // Show potentially detailed message in dev
				}
				// Otherwise, errorMsg remains "Internal Server Error" (set by default earlier)
			}
		}

	SendResponse:
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

		// Ensure status code is valid before sending
		if statusCode < 100 || statusCode > 599 {
			// Use the guaranteed non-nil logger
			cfg.Logger("[%s] Invalid status code (%d) detected in error handler. Defaulting to 500.", requestID, statusCode)
			statusCode = fiber.StatusInternalServerError
		}

		// Check if the response has already been sent
		if ctx.Response().StatusCode() != fiber.StatusOK { // Check if status is still default 200
			// If status is not 200, it might have been set previously.
			// Avoid sending response again if headers are already sent.
			// Note: Fiber might handle this internally, but adding a check can prevent potential issues.
			// This check might need refinement based on Fiber's exact behavior for double responses.
			// For now, we proceed, assuming Fiber handles it or the previous response wasn't fully sent.
		}

		return ctx.Status(statusCode).JSON(response)
	}
}
