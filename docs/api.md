# API Package

The API package is responsible for setting up and configuring the GreenTrade application's HTTP server and routing. It serves as the entry point for all HTTP requests to the system.

## Core Components

### Router

The router is responsible for:

1. Configuring the Fiber web framework
2. Setting up middleware
3. Defining all API routes
4. Organizing routes into logical groups

### Middleware Configuration

The API package configures several middleware components:

- **Request ID**: Assigns a unique ID to each request for tracking
- **Logging**: Records details about each request
- **CORS**: Manages Cross-Origin Resource Sharing policies
- **Rate Limiting**: Prevents abuse by limiting request frequency
- **Compression**: Reduces response size
- **Error Recovery**: Handles panics gracefully
- **Caching**: Improves performance for applicable routes
- **ETag**: Optimizes caching with entity tags

### Route Organization

Routes are organized into logical groups:

1. **Public Routes**:

   - Authentication endpoints
   - Public listing endpoints
   - Seller information
   - Public reviews

2. **Protected Routes** (requiring authentication):
   - User management
   - Listing management (create, delete)
   - Chat functionality
   - Review posting
   - Favorites management
   - Health monitoring

## Implementation Details

The router uses a structured approach to define routes, with separate functions for different route categories. This modular organization makes the codebase easier to maintain and extend.
