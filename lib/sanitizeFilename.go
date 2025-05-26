package lib

import (
	"html"
	"regexp"
	"strings"
	"unicode"
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

	// Replace multiple spaces with a single space
	multipleSpaces := regexp.MustCompile(`\s+`)
	sanitized = multipleSpaces.ReplaceAllString(sanitized, " ")

	// Remove non-printable/control characters (e.g., \n, \t, etc.)
	sanitized = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\n' && r != '\r' {
			return -1
		}
		return r
	}, sanitized)

	// Escape HTML to prevent XSS
	sanitized = html.EscapeString(sanitized)

	return sanitized
}
