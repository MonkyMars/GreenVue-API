package lib

import (
	"fmt"
	"math"
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

// EcoAttributeWeights maps eco attributes to their weight values
var EcoAttributeWeights = map[string]float32{
	"Second-hand":      1.2,
	"Refurbished":      1.2,
	"Upcycled":         1.2,
	"Locally Made":     0.8,
	"Organic Material": 1.2,
	"Biodegradable":    1.2,
	"Energy Efficient": 1.2,
	"Plastic-free":     1.2,
	"Vegan":            0.8,
	"Handmade":         0.8,
	"Repaired":         1.2,
}

// CalculateEcoScore calculates an eco score based on provided attributes
func CalculateEcoScore(attributes []string) float32 {
	maxScore := float32(0)
	for _, weight := range EcoAttributeWeights {
		maxScore += weight
	}

	totalScore := float32(0)
	for _, attribute := range attributes {
		if weight, exists := EcoAttributeWeights[attribute]; exists {
			totalScore += weight
		}
	}

	ecoScore := float32(math.Pow(float64(totalScore/maxScore), 0.25) * 5)

	// Round to 1 decimal place
	return float32(math.Round(float64(ecoScore*10)) / 10)
}

func Min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
