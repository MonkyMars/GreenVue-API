package lib

import (
	"math"
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
	// Step 1: Enforce maximum length to prevent DoS attacks
	maxLength := 10000 // Adjust based on your needs
	if len(input) > maxLength {
		input = input[:maxLength]
	}

	// Step 2: Remove unsafe control characters but keep printable, \n, \r, \t
	cleaned := strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || r == '\n' || r == '\r' || r == '\t' {
			return r
		}
		return -1
	}, input)

	// Step 3: Remove HTML tags but keep inner text
	cleaned = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(cleaned, "")

	// Step 4: Remove potentially dangerous sequences
	// Remove javascript: protocol
	cleaned = regexp.MustCompile(`(?i)javascript:`).ReplaceAllString(cleaned, "")
	// Remove data: protocol (can be used for XSS)
	cleaned = regexp.MustCompile(`(?i)data:`).ReplaceAllString(cleaned, "")
	// Remove on* event handlers (onclick, onload, etc.)
	cleaned = regexp.MustCompile(`(?i)on\w+\s*=`).ReplaceAllString(cleaned, "")

	// Step 5: Normalize multiple spaces and tabs (but not newlines)
	cleaned = regexp.MustCompile(` {2,}`).ReplaceAllString(cleaned, " ")
	cleaned = regexp.MustCompile(`\t{2,}`).ReplaceAllString(cleaned, "\t")

	// Step 6: Trim only leading/trailing spaces (preserve \n, \t)
	cleaned = strings.Trim(cleaned, " ")

	return cleaned
}

// SanitizeInputStrict provides stricter sanitization for sensitive contexts
func SanitizeInputStrict(input string) string {
	// Use basic sanitization first
	cleaned := SanitizeInput(input)

	// Additional strict measures
	// Remove all angle brackets to prevent any HTML/XML
	cleaned = strings.ReplaceAll(cleaned, "<", "")
	cleaned = strings.ReplaceAll(cleaned, ">", "")

	// Remove quotes that could break out of attributes
	cleaned = strings.ReplaceAll(cleaned, "\"", "")
	cleaned = strings.ReplaceAll(cleaned, "'", "")

	// Remove backslashes that could be used for escaping
	cleaned = strings.ReplaceAll(cleaned, "\\", "")

	return cleaned
}

// SanitizeInputForDisplay prepares user input for safe HTML display
// This preserves formatting while preventing XSS
func SanitizeInputForDisplay(input string) string {
	// Use basic sanitization
	cleaned := SanitizeInput(input)

	// Escape HTML entities for safe display
	cleaned = strings.ReplaceAll(cleaned, "&", "&amp;")
	cleaned = strings.ReplaceAll(cleaned, "<", "&lt;")
	cleaned = strings.ReplaceAll(cleaned, ">", "&gt;")
	cleaned = strings.ReplaceAll(cleaned, "\"", "&quot;")
	cleaned = strings.ReplaceAll(cleaned, "'", "&#x27;")

	return cleaned
}

func SanitizePrice(price float64) float64 {
	if price < 0 {
		return 0
	}
	return math.Round(price*100) / 100 // Round to 2 decimal places
}
