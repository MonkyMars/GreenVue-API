# Auth Package

The Auth package handles all aspects of user authentication and authorization in the GreenTrade API.

## Core Components

### JWT Management

The JWT (JSON Web Token) functionality includes:

1. **Token Generation**: Creates access and refresh tokens
2. **Token Validation**: Verifies token integrity and expiration
3. **Token Refreshing**: Allows users to obtain new access tokens
4. **Cookie Management**: Securely handles token storage in cookies

### User Authentication

The package implements several authentication methods:

1. **Username/Password Authentication**: Traditional login system
2. **Social Authentication**: Integration with Google OAuth
3. **Token Authentication**: Validates user sessions via JWT

### User Management

User-related functionality includes:

1. **User Registration**: Creates new user accounts
2. **Email Verification**: Confirms user email addresses
3. **User Profile**: Retrieves and updates user information
4. **User Lookup**: Finds users by ID or token

### Middleware

The `AuthMiddleware` function protects routes that require authentication by:

1. Extracting the access token from request cookies or headers
2. Validating the token's signature and expiration
3. Attaching the user's identity to the request context
4. Rejecting requests with invalid or missing tokens

### Email Management

Email-related functions handle:

1. Sending confirmation emails to new users
2. Processing email verification links
3. Resending verification emails when requested

## Security Features

The Auth package implements several security best practices:

1. Separate secrets for access and refresh tokens
2. Token expiration (15 minutes for access, 7 days for refresh)
3. Token type validation to prevent token misuse
4. Secure cookie settings with proper flags for HTTPS environments
