package lib

import (
	"regexp"
)

// IsNumeric checks if a string contains only numeric characters (0-9)
func IsNumeric(s string) bool {
	re := regexp.MustCompile(`^[0-9]+$`)
	return re.MatchString(s)
}
