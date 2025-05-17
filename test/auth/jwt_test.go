package auth_test

import (
	"greenvue/internal/auth"
	"os"
	"testing"
)

func TestGenerateAndValidateTokens(t *testing.T) {
	// Set up test environment variables with the same values
	testAccessSecret := "test-access-secret"
	testRefreshSecret := "test-refresh-secret"

	os.Setenv("JWT_ACCESS_SECRET", testAccessSecret)
	os.Setenv("JWT_REFRESH_SECRET", testRefreshSecret)
	defer func() {
		os.Unsetenv("JWT_ACCESS_SECRET")
		os.Unsetenv("JWT_REFRESH_SECRET")
	}()

	// Test data
	userID := "test-user-123"
	role := "authenticated"

	// Generate token pair
	tokenPair, err := auth.GenerateTokenPair(userID, role)
	if err != nil {
		t.Fatalf("Failed to generate token pair: %v", err)
	}

	// Verify token pair structure
	if tokenPair.AccessToken == "" {
		t.Error("Access token is empty")
	}
	if tokenPair.RefreshToken == "" {
		t.Error("Refresh token is empty")
	}
	if tokenPair.ExpiresIn <= 0 {
		t.Error("ExpiresIn should be positive")
	}

	// Validate access token
	claims, err := auth.ValidateToken(tokenPair.AccessToken, "access_token")
	if err != nil {
		t.Errorf("Failed to validate access token: %v", err)
	} else {
		// Verify claims
		if claims.UserId != userID {
			t.Errorf("Expected user ID %s, got %s", userID, claims.UserId)
		}
		if claims.Role != role {
			t.Errorf("Expected email %s, got %s", role, claims.Role)
		}
	}

	// Validate refresh token with the proper isRefresh flag
	refreshClaims, refreshErr := auth.ValidateToken(tokenPair.RefreshToken, "refresh_token")
	if refreshErr != nil {
		t.Errorf("Failed to validate refresh token: %v", refreshErr)
	} else {
		// Verify refresh token claims
		if refreshClaims.UserId != userID {
			t.Errorf("Expected user ID %s in refresh token, got %s", userID, refreshClaims.UserId)
		}
		if refreshClaims.Role != role {
			t.Errorf("Expected email %s in refresh token, got %s", role, refreshClaims.Role)
		}
	}
}

func TestTokenSecurity(t *testing.T) {
	// Set up test environment variables with different secrets
	firstAccessSecret := "first-access-secret"
	secondAccessSecret := "second-access-secret"

	// First set of tokens
	os.Setenv("JWT_ACCESS_SECRET", firstAccessSecret)
	os.Setenv("JWT_REFRESH_SECRET", "test-refresh-secret")

	firstTokenPair, err := auth.GenerateTokenPair("user1", "user1@example.com")
	if err != nil {
		t.Fatalf("Failed to generate first token pair: %v", err)
	}

	// Second set of tokens with different secret
	os.Setenv("JWT_ACCESS_SECRET", secondAccessSecret)

	secondTokenPair, err := auth.GenerateTokenPair("user2", "user2@example.com")
	if err != nil {
		t.Fatalf("Failed to generate second token pair: %v", err)
	}

	// Tokens should be different
	if firstTokenPair.AccessToken == secondTokenPair.AccessToken {
		t.Error("Access tokens should be different with different secrets")
	}

	// First token should not validate with second secret
	_, err = auth.ValidateToken(firstTokenPair.AccessToken, "access_token")
	if err == nil {
		t.Error("First token should not validate with second secret")
	}

	// Restore first secret and validate first token
	os.Setenv("JWT_ACCESS_SECRET", firstAccessSecret)
	_, err = auth.ValidateToken(firstTokenPair.AccessToken, "access_token")
	if err != nil {
		t.Errorf("First token should validate with first secret: %v", err)
	}

	// Clean up
	os.Unsetenv("JWT_ACCESS_SECRET")
	os.Unsetenv("JWT_REFRESH_SECRET")
}
