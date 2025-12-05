package main

import (
	"fmt"
	"os"
	"strings"
)

func EnsurePathExists(spellPath string) error {
	// Check if the spells directory exists, create if it doesn't
	if _, err := os.Stat(spellPath); os.IsNotExist(err) {
		fmt.Printf("Creating spellpath: %s\n", spellPath)
		err = os.MkdirAll(spellPath, 0755)
		if err != nil {
			return fmt.Errorf("failed to create spellpath: %w", err)
		}
	}

	return nil
}

func SanitizeFilename(name string) string {
	// Replace spaces with underscores and remove invalid characters
	sanitized := strings.ReplaceAll(name, " ", "_")
	sanitized = strings.ToLower(sanitized)

	// Remove characters that are not alphanumeric, underscore, or hyphen
	var result strings.Builder
	for _, r := range sanitized {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			result.WriteRune(r)
		}
	}

	return result.String()
}
