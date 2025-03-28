# Error Handling in GreenTrade.eu Backend

This package provides a consistent and structured approach to error handling across the GreenTrade.eu backend.

## Key Components

### 1. Error Types

Pre-defined error types for common scenarios:
- `ErrBadRequest`: Invalid input from client
- `ErrUnauthorized`: Authentication required
- `ErrForbidden`: Authentication successful but permission denied
- `ErrNotFound`: Requested resource not found
- `ErrInternalServerError`: Server-side error
- `ErrValidation`: Input validation error
- `ErrDatabaseError`: Database operation error
- `ErrAlreadyExists`: Duplicate resource error

### 2. AppError Structure

The `AppError` struct encapsulates all error-related information:
- `Err`: The underlying error
- `StatusCode`: HTTP status code
- `Message`: User-friendly error message
- `Field`: For validation errors, identifies the problematic field
- `Internal`: Whether the error details should be hidden from clients

### 3. Error Factories

Helper functions for creating consistent errors:
- `BadRequest(message)`: Creates a 400 Bad Request error
- `ValidationError(message, field)`: Creates a 400 with field identification
- `Unauthorized(message)`: Creates a 401 Unauthorized error
- `Forbidden(message)`: Creates a 403 Forbidden error
- `NotFound(message)`: Creates a 404 Not Found error
- `InternalServerError(message)`: Creates a 500 Internal Server error
- `DatabaseError(message)`: Creates a 500 for database errors
- `AlreadyExists(message)`: Creates a 409 Conflict error

### 4. Middleware

The `ErrorHandler` middleware provides consistent error handling across the application:
- Converts errors to appropriate HTTP responses
- Sanitizes error details based on environment
- Maintains consistent error response format

### 5. Response Utilities

Helper functions for consistent response formatting:
- `HandleError(c, err)`: Handles an error in route handlers
- `SuccessResponse(c, data)`: Formats successful responses
- `ErrorResponse(c, statusCode, message)`: Creates custom error responses
- `ValidateFields(fields)`: Validates required fields
- `ValidateRequest(c, data)`: Parses and validates request body

## Usage Examples

### Basic Error Handling

```go
func MyHandler(c *fiber.Ctx) error {
    if err := doSomething(); err != nil {
        return errors.InternalServerError("Failed to do something")
    }
    return errors.SuccessResponse(c, data)
}
```

### Field Validation

```go
func CreateUser(c *fiber.Ctx) error {
    var payload struct {
        Username string `json:"username"`
        Email    string `json:"email"`
    }
    
    if err := errors.ValidateRequest(c, &payload); err != nil {
        return err
    }
    
    if err := errors.ValidateFields(map[string]string{
        "username": payload.Username,
        "email":    payload.Email,
    }); err != nil {
        return err
    }
    
    // Continue with valid data...
} 