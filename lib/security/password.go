package security

import (
	"golang.org/x/crypto/bcrypt"
)

// PasswordHasher provides methods for hashing and verifying passwords
type PasswordHasher struct {
	Cost int
}

// NewPasswordHasher creates a new hasher with default settings
func NewPasswordHasher() *PasswordHasher {
	return &PasswordHasher{
		Cost: bcrypt.DefaultCost,
	}
}

// HashPassword hashes a password using bcrypt
func (h *PasswordHasher) HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), h.Cost)
	return string(bytes), err
}

// VerifyPassword checks if a password matches a hash
func (h *PasswordHasher) VerifyPassword(hashedPassword, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}

// HashPassword is a convenience function that uses the default hasher
func HashPassword(password string) (string, error) {
	hasher := NewPasswordHasher()
	return hasher.HashPassword(password)
}

// VerifyPassword is a convenience function that uses the default hasher
func VerifyPassword(hashedPassword, password string) bool {
	hasher := NewPasswordHasher()
	return hasher.VerifyPassword(hashedPassword, password)
}
