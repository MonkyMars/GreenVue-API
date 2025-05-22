package lib

import (
	"fmt"
	"regexp"

	"github.com/google/uuid"
)

// IsNumeric checks if a string contains only numeric characters (0-9)
func IsNumeric(s string) bool {
	re := regexp.MustCompile(`^[0-9]+$`)
	return re.MatchString(s)
}

func ParseUUID(id string) (string, error) {
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !re.MatchString(id) {
		return "", fmt.Errorf("invalid UUID format")
	}
	return id, nil
}

// GenerateUUID generates a random UUID string
func GenerateUUID() string {
	return uuid.New().String()
}

func StringToUUID(id string) (uuid.UUID, error) {
	re := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !re.MatchString(id) {
		return uuid.Nil, fmt.Errorf("invalid UUID format")
	}
	return uuid.Parse(id)
}
