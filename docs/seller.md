# Seller Package

The Seller package manages seller information in the GreenTrade API, providing functionality to retrieve and display seller profiles.

## Core Components

### Seller Retrieval

The package provides the core function:

1. **GetSeller**: Retrieves detailed seller information by user ID
2. **Profile Data**: Includes seller name, avatar, and statistics

### Seller Metrics

The package handles seller performance metrics:

1. **Listing Count**: Number of active listings from the seller
2. **Rating Average**: Average rating from buyer reviews
3. **Review Count**: Total number of reviews received
4. **Join Date**: When the seller account was created

### Database Integration

Seller data is managed through:

1. **User Table**: Core user information shared with the Auth package
2. **Seller View**: A database view that aggregates seller metrics
3. **Query Optimization**: Efficient retrieval of seller profile data

## Implementation Details

The Seller package implements these features:

1. **Public Access**: Seller profiles are publicly accessible without authentication
2. **Data Aggregation**: Combining user data with metrics from listings and reviews
3. **Response Formatting**: Structured API responses for seller profiles

## Data Models

The package uses several data structures:

1. **Seller**: Core seller profile information
2. **SellerMetrics**: Performance statistics for the seller
3. **SellerResponse**: Formatted response structure for API clients

## API Endpoints

The Seller package exposes:

1. **GET /seller/:user_id**: Public endpoint to fetch seller information

## Integration Points

The Seller package integrates with:

1. **Auth Package**: Shares user profile information
2. **Listings Package**: Retrieves seller's active listings
3. **Reviews Package**: Fetches seller ratings and reviews
