package lib

import (
	"regexp"
	"strings"
)

func SanitizeFilename(filename string) string {
	replacer := strings.NewReplacer(
		" ", "_",
		"/", "_",
		"\\", "_",
		":", "_",
		"*", "_",
		"?", "_",
		"\"", "_",
		"<", "_",
		">", "_",
		"|", "_",
	)
	return replacer.Replace(filename)
}

func SanitizeInput(input string) string {
	// Trim whitespace
	sanitized := strings.TrimSpace(input)

	// Replace multiple spaces with single space
	multipleSpaces := regexp.MustCompile(`\s+`)
	sanitized = multipleSpaces.ReplaceAllString(sanitized, " ")

	return sanitized
}
