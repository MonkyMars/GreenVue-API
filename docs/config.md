# Config Package

The Config package manages application configuration for the GreenTrade API, providing a centralized way to access and manage settings across the application.

## Core Components

### Configuration Structure

The package defines a hierarchical configuration structure with these main sections:

1. **Server Configuration**:

   - Port settings
   - Request timeouts (read, write, idle)

2. **Database Configuration**:

   - Supabase URL
   - Supabase API key

3. **JWT Configuration**:

   - Secret keys for access and refresh tokens
   - Token expiration durations

4. **Environment Settings**:
   - Environment identifier (development, production)

### Configuration Loading

The package provides a `LoadConfig` function that:

1. Reads configuration from environment variables
2. Applies sensible defaults for missing values
3. Parses duration strings into time.Duration objects
4. Returns a complete configuration object

### Helper Functions

The package includes utilities for:

1. Retrieving environment variables with defaults
2. Parsing duration values from string environment variables
3. Converting configuration values to appropriate types

## Usage Pattern

The Config package is typically used at application startup to load configuration, which is then passed to various components. This allows different parts of the application to access only the configuration they need without global variables.

## Security Considerations

The package handles sensitive configuration like database credentials and JWT secrets. In production, these values should be provided through secure environment variables rather than using the default development values.
