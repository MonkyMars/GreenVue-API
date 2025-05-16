# Favorites Package

The Favorites package manages user favorites functionality in the GreenTrade API, allowing users to bookmark and manage listings of interest.

## Core Components

### Favorites Management

The package provides several key functions:

1. **GetFavorites**: Retrieves a user's favorite listings
2. **AddFavorite**: Adds a listing to a user's favorites
3. **DeleteFavorite**: Removes a listing from a user's favorites
4. **IsFavorite**: Checks if a listing is in a user's favorites

### Database Integration

The package uses a specialized database view for efficient data retrieval:

1. **User Favorites View**: A PostgreSQL view that joins favorites with listing and seller information
2. **Query Optimization**: Structured queries to efficiently fetch favorites with related data
3. **Response Formatting**: Converts database results into structured API responses

### Data Models

The favorites functionality uses several data structures:

1. **FetchedFavorite**: Represents a favorite with its associated listing and seller data
2. **FavoritePayload**: Contains data required to add a new favorite

### Error Handling

The package implements comprehensive error handling for favorites operations:

1. **Input Validation**: Ensures required parameters are present
2. **Database Error Handling**: Manages database connection and query errors
3. **Response Formatting**: Provides consistent error responses for client applications

## Implementation Details

The Favorites package follows a RESTful API approach with:

1. **GET /api/favorites/:user_id**: Retrieves all favorites for a user
2. **GET /api/favorites/check/:listing_id/:user_id**: Checks if a listing is favorited
3. **POST /api/favorites**: Adds a new favorite
4. **DELETE /api/favorites/:listing_id/:user_id**: Removes a favorite
