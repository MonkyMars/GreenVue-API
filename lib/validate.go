package lib

import (
	"regexp"
	"strings"
)

// UsernameValidation checks if a username is valid based on security and spam rules.
func UsernameValidation(username string) (bool, string) {
	// Length checks
	if len(username) < 3 || len(username) > 30 {
		return false, "Username must be between 3 and 30 characters"
	}

	// Basic inappropriate content check
	blockedWords := []string{"admin", "moderator", "support", "staff"}
	lowerUsername := strings.ToLower(username)

	for _, word := range blockedWords {
		if strings.Contains(lowerUsername, word) {
			return false, "Username contains reserved words"
		}
	}

	// Block excessive repeating characters (potential spam)
	if hasExcessiveRepeatingChars(username, 5) {
		return false, "Username contains too many repeating characters"
	}

	// Allow only alphanumeric characters and underscores
	validChars := regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
	if !validChars.MatchString(username) {
		return false, "Username can only contain letters, numbers, and underscores"
	}

	return true, ""
}

// hasExcessiveRepeatingChars detects if a character repeats `limit` or more times consecutively.
func hasExcessiveRepeatingChars(s string, limit int) bool {
	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= limit {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}
