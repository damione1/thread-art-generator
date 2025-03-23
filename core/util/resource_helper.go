package util

import (
	"fmt"
	"regexp"
)

// ResourceNamePattern is a regexp pattern for validating resource names
var (
	userResourcePattern = regexp.MustCompile(`^users/([^/]+)$`)
	artResourcePattern  = regexp.MustCompile(`^users/([^/]+)/arts/([^/]+)$`)
)

// ExtractUserID extracts the user ID from a user resource name
func ExtractUserID(resourceName string) string {
	matches := userResourcePattern.FindStringSubmatch(resourceName)
	if len(matches) == 2 {
		return matches[1]
	}

	// Try to extract from art resource name
	matches = artResourcePattern.FindStringSubmatch(resourceName)
	if len(matches) == 3 {
		return matches[1]
	}

	return resourceName // Return as-is if not a valid pattern
}

// ExtractArtID extracts the art ID from an art resource name
func ExtractArtID(resourceName string) (string, error) {
	matches := artResourcePattern.FindStringSubmatch(resourceName)
	if len(matches) != 3 {
		return "", fmt.Errorf("invalid art resource name: %s", resourceName)
	}
	return matches[2], nil
}

// CreateUserResourceName creates a user resource name from a user ID
func CreateUserResourceName(userID string) string {
	return fmt.Sprintf("users/%s", userID)
}

// CreateArtResourceName creates an art resource name from a user ID and art ID
func CreateArtResourceName(userID, artID string) string {
	return fmt.Sprintf("users/%s/arts/%s", userID, artID)
}
