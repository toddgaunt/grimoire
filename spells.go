package main

import (
	"fmt"
	"os"
	"strings"
)

// EnsurePathExists will create the spellPath if it doesn't
// yet already exists.
func EnsurePathExists(spellPath string) error {
	// Check if the spells directory exists, create if it doesn't
	if _, err := os.Stat(spellPath); os.IsNotExist(err) {
		err = os.MkdirAll(spellPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create spellpath: %w", err)
		}
	}

	return nil
}

// SanitizeFilename strips leading and trailing whitespace, and removes all
// characters that are not alphanumeric, underscores, or hyphens.
func SanitizeFilename(name string) string {
	sanitized := strings.TrimSpace(name)
	sanitized = strings.ReplaceAll(name, " ", "_")
	sanitized = strings.ToLower(sanitized)

	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}
