package security_test

import (
	"greenvue/lib/security"
	"testing"
)

func TestPasswordHashAndVerify(t *testing.T) {
	// Test cases with different password strengths
	passwords := []string{
		"simple123",
		"Complex!Password123",
		"VeryLongAndSecurePasswordWithSpecialChars!@#$%^&*()",
		"短い", // Unicode characters
	}

	// Test the convenience functions
	for i, password := range passwords {
		// Hash the password
		hashedPassword, err := security.HashPassword(password)
		if err != nil {
			t.Errorf("Failed to hash password #%d: %v", i, err)
			continue
		}

		// Verify the password is not stored in plain text
		if hashedPassword == password {
			t.Errorf("Password #%d is stored in plain text", i)
		}

		// Verify the correct password
		if !security.VerifyPassword(hashedPassword, password) {
			t.Errorf("Failed to verify password #%d", i)
		}

		// Verify an incorrect password fails
		if security.VerifyPassword(hashedPassword, password+"wrong") {
			t.Errorf("Incorrectly verified wrong password #%d", i)
		}
	}
}

func TestPasswordHasher(t *testing.T) {
	// Create a default hasher
	hasher := security.NewPasswordHasher()

	// Test with a simple password
	password := "test123"

	// Hash the password
	hashedPassword, err := hasher.HashPassword(password)
	if err != nil {
		t.Errorf("Failed to hash password: %v", err)
		return
	}

	// Verify the correct password
	if !hasher.VerifyPassword(hashedPassword, password) {
		t.Errorf("Failed to verify password")
	}

	// Verify an incorrect password fails
	if hasher.VerifyPassword(hashedPassword, "wrongpassword") {
		t.Errorf("Incorrectly verified wrong password")
	}

	// Test with custom cost
	customHasher := security.NewPasswordHasher()
	customHasher.Cost = 5 // Lower cost for testing purposes

	// Hash with custom cost
	customHashed, err := customHasher.HashPassword(password)
	if err != nil {
		t.Errorf("Failed to hash password with custom cost: %v", err)
		return
	}

	// Verify the password still works with custom cost
	if !customHasher.VerifyPassword(customHashed, password) {
		t.Errorf("Failed to verify password with custom cost")
	}

	// Different hashers should still work with each other's hashed passwords
	if !hasher.VerifyPassword(customHashed, password) {
		t.Errorf("Failed to verify password between different hashers")
	}

	if !customHasher.VerifyPassword(hashedPassword, password) {
		t.Errorf("Failed to verify password between different hashers (reverse)")
	}
}
