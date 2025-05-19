package auth

import (
	"crypto/subtle"
	"errors"
	"greenvue/internal/config"
	response "greenvue/lib/errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken      = errors.New("invalid token")
	ErrExpiredToken      = errors.New("token has expired")
	ErrTokenTypeMismatch = errors.New("token type mismatch")
	ErrTokenTampering    = errors.New("token has been tampered with")
)

var (
	cfg      = config.LoadConfig()
	Secure   bool
	SameSite string
)

const (
	TokenTypeAccess  = "access_token"
	TokenTypeRefresh = "refresh_token"
)

type Claims struct {
	UserId uuid.UUID `json:"user_id"`
	Role   string    `json:"role"`
	Type   string    `json:"type"` // New field to identify token type
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
}

// Cookie configuration constants
const (
	AccessTokenCookieName  = "access_token"
	RefreshTokenCookieName = "refresh_token"
	AccessCookieMaxAge     = 3600          // 1 hour
	RefreshCookieMaxAge    = 7 * 24 * 3600 // 7 days
)

// InitEnvironmentConfig initializes the Secure and SameSite variables based on the environment
func InitEnvironmentConfig() {
	if cfg.Environment == "production" {
		Secure = true
		SameSite = "Strict"
	} else {
		// For development environment
		// When testing with cross-origin requests (e.g., front-end on localhost:3000
		// and backend on a different origin), we need special cookie settings
		Secure = false   // Initially false, may be set to true by ConfigureCrossOriginCookies
		SameSite = "Lax" // Initially Lax, may be set to None by ConfigureCrossOriginCookies
	}
}

// ClearAuthCookies clears all authentication-related cookies
func ClearAuthCookies(c *fiber.Ctx) {
	// Initialize environment config
	// This is necessary to ensure Secure and SameSite are set correctly
	InitEnvironmentConfig()

	// Get the hostname and extract an appropriate domain
	host := c.Hostname()
	domain := getDomainFromHost(host)

	// Clear access token cookie
	c.Cookie(&fiber.Cookie{
		Name:     AccessTokenCookieName,
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
		Secure:   Secure,
		HTTPOnly: true,
		SameSite: SameSite,
	})

	// Clear refresh token cookie
	c.Cookie(&fiber.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    "",
		Path:     "/",
		Domain:   domain,
		MaxAge:   -1,
		Expires:  time.Now().Add(-time.Hour),
		Secure:   Secure,
		HTTPOnly: true,
		SameSite: SameSite,
	})
}

// SetTokenCookie sets the JWT token as a secure HTTP-only cookie
func SetTokenCookie(c *fiber.Ctx, token string) {
	// For cross-origin development setup, we need to omit the domain completely
	// to allow the browser to handle cookie domains naturally
	var domain string
	if cfg.Environment != "production" && SameSite == "None" {
		// For cross-origin development, don't set domain at all
		domain = ""
	} else {
		// Get the hostname and extract an appropriate domain
		host := c.Hostname()
		domain = getDomainFromHost(host)
	}

	cookie := &fiber.Cookie{
		Name:     AccessTokenCookieName,
		Value:    token,
		Path:     "/",
		Domain:   domain, // May be empty for cross-origin development
		MaxAge:   AccessCookieMaxAge,
		Expires:  time.Now().Add(time.Duration(AccessCookieMaxAge) * time.Second),
		Secure:   Secure,
		HTTPOnly: true,
		SameSite: SameSite,
	}

	// Setting the cookie
	c.Cookie(cookie)
}

// SetRefreshTokenCookie sets the JWT refresh token as a secure HTTP-only cookie
func SetRefreshTokenCookie(c *fiber.Ctx, token string) {
	// For cross-origin development setup, we need to omit the domain completely
	// to allow the browser to handle cookie domains naturally
	var domain string
	if cfg.Environment != "production" && SameSite == "None" {
		// For cross-origin development, don't set domain at all
		domain = ""
	} else {
		// Get the hostname and extract an appropriate domain
		host := c.Hostname()
		domain = getDomainFromHost(host)
	}

	cookie := &fiber.Cookie{
		Name:     RefreshTokenCookieName,
		Value:    token,
		Path:     "/",
		Domain:   domain, // May be empty for cross-origin development
		Expires:  time.Now().Add(time.Duration(RefreshCookieMaxAge) * time.Second),
		MaxAge:   RefreshCookieMaxAge,
		Secure:   Secure,
		HTTPOnly: true,
		SameSite: SameSite,
	}

	// Setting the cookie
	c.Cookie(cookie)
}

func SetAuthCookies(c *fiber.Ctx, tokens *TokenPair) {
	// Set the access token cookie
	SetTokenCookie(c, tokens.AccessToken)

	// Set the refresh token cookie
	SetRefreshTokenCookie(c, tokens.RefreshToken)
}

func getJWTSecrets() (access []byte, refresh []byte) {
	accessSecret := os.Getenv("JWT_ACCESS_SECRET")
	refreshSecret := os.Getenv("JWT_REFRESH_SECRET")

	if accessSecret == "" || refreshSecret == "" {
		log.Fatal("JWT secrets are not set. Please set JWT_ACCESS_SECRET and JWT_REFRESH_SECRET in your environment.")
	}

	return []byte(accessSecret), []byte(refreshSecret)
}

// GenerateTokenPair creates a new access token and refresh token
func GenerateTokenPair(userID uuid.UUID, email string) (*TokenPair, error) {
	accessExpiration := time.Now().Add(time.Duration(AccessCookieMaxAge) * time.Second)
	refreshExpiration := time.Now().Add(time.Duration(RefreshCookieMaxAge) * time.Second)

	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserId: userID,
		Role:   "authenticated",
		Type:   TokenTypeAccess, // Specify token type
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  []string{"greenvue-client"},          // Audience of the token
			Issuer:    "greenvue",                           // Issuer of the token
			Subject:   userID.String(),                      // Subject of the token
			ExpiresAt: jwt.NewNumericDate(accessExpiration), // Expiration time (1 hour)
			IssuedAt:  jwt.NewNumericDate(time.Now()),       // Time when the token was issued
		},
	})

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserId: userID,
		Role:   "authenticated",
		Type:   TokenTypeRefresh, // Specify token type
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  []string{"greenvue-client"},
			Issuer:    "greenvue",
			Subject:   userID.String(),
			ExpiresAt: jwt.NewNumericDate(refreshExpiration),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	// Sign the tokens
	accessSecret, refreshSecret := getJWTSecrets()
	accessTokenString, err := accessToken.SignedString(accessSecret)
	if err != nil {
		return nil, err
	}

	refreshTokenString, err := refreshToken.SignedString(refreshSecret)
	if err != nil {
		return nil, err
	}
	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(AccessCookieMaxAge),
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, expectedType string) (*Claims, error) {
	accessSecret, refreshSecret := getJWTSecrets()

	// Determine which secret to use based on expected token type
	var secret []byte
	if expectedType == TokenTypeAccess {
		secret = accessSecret
	} else if expectedType == TokenTypeRefresh {
		secret = refreshSecret
	} else {
		return nil, errors.New("invalid token type specified")
	}

	// Split the token into its parts to check for tampering first
	parts := strings.Split(tokenString, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}

	// Use ParseWithClaims with strict validation options
	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (any, error) {
			// Ensure that the signing method is HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		},
		jwt.WithValidMethods([]string{"HS256"}),
		jwt.WithStrictDecoding(),
		jwt.WithExpirationRequired(),
	)

	// Handle parsing errors
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) ||
			errors.Is(err, jwt.ErrTokenMalformed) ||
			errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, ErrTokenTampering
		}
		return nil, ErrInvalidToken
	}

	// Extract and validate claims
	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	// Verify token type matches expected type
	if claims.Type != expectedType {
		return nil, ErrTokenTypeMismatch
	}

	// For the most rigorous verification, we need to regenerate the token
	// with the exact same claims and ensure it matches the original

	// Create a new token with the same signing method
	newToken := jwt.New(jwt.SigningMethodHS256)

	// Copy all claims exactly as they were in the original token
	newToken.Claims = claims

	// Sign the token with the same secret
	verifiedTokenString, err := newToken.SignedString(secret)
	if err != nil {
		return nil, err
	}

	// Compare the original token with our regenerated token - they must match exactly
	// Use constant-time comparison to prevent timing attacks
	if subtle.ConstantTimeCompare([]byte(tokenString), []byte(verifiedTokenString)) != 1 {
		log.Printf("Token tampering detected: token strings don't match exactly")
		return nil, ErrTokenTampering
	}

	return claims, nil
}

// AuthMiddleware is a middleware that validates JWT tokens from either a bearer token or a cookie.
// It prefers to use cookies over bearer tokens when both are available.
// It also checks for a health access token for specific routes.
// It gets the user ID from the request body and compares it with the token claims.
// If the user ID in the request body does not match the token claims, it returns an unauthorized error.
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var tokenString string
		var err error

		// Check for token in cookie first (cookies take precedence)
		tokenCookie := c.Cookies(AccessTokenCookieName)

		if tokenCookie != "" {
			tokenString = tokenCookie
		} else {
			// If no cookie, check for token from Authorization header
			authHeader := c.Get("Authorization")
			if strings.HasPrefix(authHeader, "Bearer ") {
				tokenString = strings.TrimPrefix(authHeader, "Bearer ")
			}
		}

		// If no token found in either place
		if tokenString == "" {
			return response.Unauthorized("invalid or missing token")
		}

		path := c.Path()
		if strings.HasPrefix(path, "/api/health") {
			healthAccessToken := os.Getenv("HEALTH_ACCESS_TOKEN")
			if tokenString != healthAccessToken {
				return response.Unauthorized("invalid health access token")
			}
			return c.Next()
		}

		// Validate token as an access token specifically
		claims, err := ValidateToken(tokenString, TokenTypeAccess)
		if err != nil {
			// If the token is invalid or expired, clear the cookies
			if err == ErrInvalidToken || err == ErrExpiredToken || err == ErrTokenTypeMismatch || err == ErrTokenTampering {
				ClearAuthCookies(c)
			}
			return response.Unauthorized("authentication failed: " + err.Error())
		}

		switch c.Method() {
		case fiber.MethodPost:
			var payload struct {
				UserId uuid.UUID `json:"user_id"`
			}

			if err := c.BodyParser(&payload); err != nil {
				return response.BadRequest("invalid request format")
			}

			// Only validate if user_id is included in the body
			if payload.UserId != uuid.Nil && payload.UserId != claims.UserId {
				return response.Unauthorized("user ID in request body does not match token claims")
			}

		case fiber.MethodGet:
			userId := c.Query("user_id")
			// Validate user ID in query string
			if userId != "" {
				parsedID, err := uuid.Parse(userId)
				if err != nil {
					return response.BadRequest("invalid user ID format")
				}
				if parsedID != claims.UserId {
					return response.Unauthorized("user ID in query does not match token claims")
				}
			}
		case fiber.MethodDelete:
			userId := c.Query("user_id")

			if userId != "" && userId != claims.UserId.String() {
				return response.Unauthorized("user ID in query does not match token claims")
			}
		case fiber.MethodPatch:
			var payload struct {
				UserId string `json:"user_id"`
			}

			if err := c.BodyParser(&payload); err != nil {
				return response.BadRequest("invalid request format")
			}

			if payload.UserId != "" && payload.UserId != claims.UserId.String() {
				return response.Unauthorized("user ID in request body does not match token claims")
			}
		}

		// Store claims in context for later use
		c.Locals("user", claims)
		return c.Next()
	}
}

func GetAccessToken(c *fiber.Ctx) (*Claims, error) {
	var tokenString string

	// Try to get token from cookie first (cookies take precedence)
	tokenCookie := c.Cookies(AccessTokenCookieName)
	if tokenCookie != "" {
		tokenString = tokenCookie
	} else {
		// If no cookie, check for token from Authorization header
		authHeader := c.Get("Authorization")
		if strings.HasPrefix(authHeader, "Bearer ") {
			tokenString = strings.TrimPrefix(authHeader, "Bearer ")
		}
	}

	// If no token found in either place
	if tokenString == "" {
		return nil, response.Unauthorized("invalid or missing token")
	}

	// Validate token as an access token specifically
	claims, err := ValidateToken(tokenString, TokenTypeAccess)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// RefreshTokenHandler handles token refresh requests
func RefreshTokenHandler(c *fiber.Ctx) error {
	var refreshToken string

	// Check for refresh token in cookie first (cookies take precedence)
	refreshCookie := c.Cookies(RefreshTokenCookieName)
	log.Println("Refresh token from cookie:", refreshCookie)

	if refreshCookie != "" {
		refreshToken = refreshCookie
	} else {
		// If no cookie, try to get refresh token from request body
		var payload struct {
			RefreshToken string `json:"refreshToken"`
		}

		if err := c.BodyParser(&payload); err == nil && payload.RefreshToken != "" {
			refreshToken = payload.RefreshToken
		}
	}

	// Return error if no refresh token found
	if refreshToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "missing refresh token")
	}

	// Validate refresh token - specifically checking it's a refresh token type
	claims, err := ValidateToken(refreshToken, TokenTypeRefresh)
	if err != nil {
		// If refresh token is invalid or expired, clear all cookies
		ClearAuthCookies(c)
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token: "+err.Error())
	}

	// Generate new token pair
	tokens, err := GenerateTokenPair(claims.UserId, claims.Role)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate tokens")
	}

	// Set the tokens as secure cookies for web clients
	SetAuthCookies(c, tokens)

	return response.SuccessResponse(c, fiber.Map{
		"userId":       claims.UserId,
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
	})
}

// getDomainFromHost extracts an appropriate cookie domain from the hostname
func getDomainFromHost(host string) string {
	// Don't set domain for localhost or direct IP addresses to avoid issues
	if strings.Contains(host, "localhost") ||
		strings.Contains(host, "127.0.0.1") ||
		strings.HasPrefix(host, "192.168.") {
		log.Println("Using default domain for development host:", host)
		return "" // Empty domain works better for local development
	}

	// Extract domain without port for production hosts
	var domain string
	if idx := strings.Index(host, ":"); idx != -1 {
		domain = host[:idx]
	} else {
		domain = host
	}

	return domain
}
