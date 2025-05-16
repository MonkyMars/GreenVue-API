# Reviews Package

The Reviews package manages user reviews and ratings in the GreenTrade API, allowing buyers to provide feedback on sellers.

## Core Components

### Review Retrieval

The package provides functions for fetching reviews:

1. **GetReviews**: Retrieves all reviews for a specific seller
2. **Query Parameters**: Supports filtering and sorting reviews

### Review Management

For creating reviews, the package includes:

1. **PostReview**: Creates a new review for a seller
2. **Validation**: Ensures reviews contain required information
3. **Authentication**: Verifies that the reviewer is authenticated

### Review Metrics

The package handles review statistics:

1. **Rating Calculation**: Computes average seller ratings
2. **Review Counts**: Tracks the number of reviews per seller

### Database Integration

Review data is managed through:

1. **Reviews Table**: Stores review content, ratings, and metadata
2. **Seller Association**: Links reviews to the appropriate seller
3. **User Association**: Associates reviews with the users who created them

## Implementation Details

The Reviews package implements these features:

1. **Star Ratings**: Numeric ratings system (typically 1-5 stars)
2. **Text Feedback**: Written reviews providing detailed seller feedback
3. **Timestamps**: Recording when reviews were created
4. **Response Formatting**: Structured API responses for review data

## Data Models

The package uses several key data structures:

1. **Review**: Core review data including rating and text
2. **FetchedReview**: Enhanced review data with reviewer information
3. **ReviewPayload**: Input format for creating new reviews

## API Endpoints

The Reviews package exposes:

1. **GET /reviews/:seller_id**: Public endpoint to fetch seller reviews
2. **POST /api/reviews**: Protected endpoint for creating new reviews
