# Listings Package

The Listings package manages product listings in the GreenTrade API, providing functionality for creating, retrieving, and deleting marketplace listings.

## Core Components

### Listing Retrieval

The package includes several listing retrieval functions:

1. **GetListings**: Fetches all listings with pagination support
2. **GetListingById**: Retrieves a specific listing by its ID
3. **GetListingByCategory**: Filters listings by category
4. **GetListingBySeller**: Gets all listings from a specific seller

### Listing Management

For listing owners, the package provides:

1. **PostListing**: Creates a new product listing
2. **DeleteListingById**: Removes a listing from the marketplace
3. **QueuedUploadHandler**: Processes image uploads asynchronously with background jobs

### Database Integration

The package uses a specialized database view for efficient data retrieval:

1. **Listings with Seller View**: A PostgreSQL view that joins listings with seller information
2. **Query Optimization**: Structured queries for efficient data retrieval
3. **Response Formatting**: Converts database results into structured API responses

### Image Handling

The listing image functionality includes:

1. **Image Upload**: Processing and storing listing images
2. **Image Validation**: Verifying image formats and sizes
3. **Image URL Generation**: Creating accessible URLs for uploaded images

## Implementation Details

The Listings package implements several important features:

1. **Data Validation**: Ensures listings contain required information
2. **Pagination**: Supports limiting and offsetting results for performance
3. **Sorting**: Orders listings by creation date (newest first)
4. **Owner Authentication**: Verifies that only owners can modify or delete listings

## Data Models

The package uses several data structures:

1. **Listing**: Core listing data with product information
2. **FetchedListing**: Enhanced listing data with seller information
3. **ListingPayload**: Input format for creating new listings
