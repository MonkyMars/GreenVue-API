# Database Package

The Database package provides a unified interface for database operations in the GreenTrade API, with a specific implementation for Supabase.

## Core Components

### Supabase Implementation

The package includes a Supabase-specific implementation of the Repository interface:

1. **SupabaseClient**: Handles HTTP communication with Supabase
2. **Query Builder**: Constructs PostgreSQL-compatible queries

### Connection Management

Database connections are managed through:

1. **Global Client**: A singleton client instance for application-wide use
2. **Client Configuration**: Settings for connecting to Supabase
3. **Connection Validation**: Health checks to verify database connectivity

### Query Parameters

The package supports flexible querying through a structured params object:

1. **Filtering**: Applying WHERE conditions
2. **Pagination**: Limiting and offsetting results
3. **Sorting**: Ordering results by specified fields
4. **Selection**: Choosing which columns to return

## Error Handling

Database errors are handled through:

1. **Error Classification**: Categorizing errors (connection, query, etc.)
2. **Error Wrapping**: Adding context to database errors
3. **Retry Logic**: Attempting to recover from transient errors

## Performance Considerations

The package implements several performance optimizations:

1. **Connection Pooling**: Reusing connections for better performance
2. **Query Caching**: Avoiding redundant database calls
3. **Batch Operations**: Grouping operations when possible
