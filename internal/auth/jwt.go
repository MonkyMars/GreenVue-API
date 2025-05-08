package auth

import (
	"errors"
	response "greenvue-eu/lib/errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

type Claims struct {
	UserID string `json:"sub"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
	ExpiresIn    int64
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
func GenerateTokenPair(userID, email string) (*TokenPair, error) {
	// Generate access token
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		Email:  email,
		Role:   "authenticated", // required for Supabase RLS
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  []string{"authenticated"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	})

	// Generate refresh token
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, Claims{
		UserID: userID,
		Email:  email,
		Role:   "authenticated", // required for Supabase RLS
		RegisteredClaims: jwt.RegisteredClaims{
			Audience:  []string{"authenticated"},
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(7 * 24 * time.Hour)),
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
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}, nil
}

// ValidateToken validates a JWT token and returns the claims
func ValidateToken(tokenString string, isRefresh bool) (*Claims, error) {
	accessSecret, refreshSecret := getJWTSecrets()
	secret := accessSecret
	if isRefresh {
		secret = refreshSecret
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (any, error) {
		return secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// AuthMiddleware is a middleware that validates JWT tokens
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Get token from Authorization header
		authHeader := c.Get("Authorization")
		if !strings.HasPrefix(authHeader, "Bearer ") {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid or missing token")
		}
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		path := c.Path()
		if strings.HasPrefix(path, "/api/health") {
			healthAccessToken := os.Getenv("HEALTH_ACCESS_TOKEN")
			if tokenString != healthAccessToken {
				return fiber.NewError(fiber.StatusUnauthorized, "invalid health access token")
			}
			return c.Next()
		}

		// Validate token
		claims, err := ValidateToken(tokenString, false)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, err.Error())
		}

		// Store claims in context for later use
		c.Locals("user", claims)
		return c.Next()
	}
}

func GetAccessToken(c *fiber.Ctx) (*Claims, error) {
	// Get token from Authorization header
	authHeader := c.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "invalid or missing token")
	}
	tokenString := strings.TrimPrefix(authHeader, "Bearer ")

	// Validate token
	claims, err := ValidateToken(tokenString, false)
	if err != nil {
		return nil, err
	}

	return claims, nil
}

// RefreshTokenHandler handles token refresh requests
func RefreshTokenHandler(c *fiber.Ctx) error {
	var payload struct {
		RefreshToken string `json:"refreshToken"`
	}

	if err := c.BodyParser(&payload); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request format")
	}

	// Validate refresh token
	claims, err := ValidateToken(payload.RefreshToken, true)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid refresh token")
	}

	// Generate new token pair
	tokens, err := GenerateTokenPair(claims.UserID, claims.Email)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate tokens")
	}

	return response.SuccessResponse(c, fiber.Map{
		"accessToken":  tokens.AccessToken,
		"refreshToken": tokens.RefreshToken,
		"expiresIn":    tokens.ExpiresIn,
	})
}
