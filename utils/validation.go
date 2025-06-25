package utils

import (
	"fmt"
	"strings"
)

// ValidateString validates string length and content
func ValidateString(s string, minLen, maxLen int) error {
	s = strings.TrimSpace(s)
	
	if len(s) < minLen {
		return fmt.Errorf("string too short, minimum length is %d", minLen)
	}
	
	if len(s) > maxLen {
		return fmt.Errorf("string too long, maximum length is %d", maxLen)
	}
	
	return nil
}

// ValidateRequired checks if a string is not empty after trimming
func ValidateRequired(s string, fieldName string) error {
	if strings.TrimSpace(s) == "" {
		return fmt.Errorf("%s is required", fieldName)
	}
	return nil
}

// SanitizeString removes potentially dangerous characters from strings
func SanitizeString(s string) string {
	// Remove null bytes and control characters
	cleaned := strings.ReplaceAll(s, "\x00", "")
	
	// Trim whitespace
	cleaned = strings.TrimSpace(cleaned)
	
	return cleaned
}

// IsValidUUID checks if a string is a valid UUID format
func IsValidUUID(uuid string) bool {
	// Basic UUID format validation (8-4-4-4-12 hex digits)
	if len(uuid) != 36 {
		return false
	}
	
	expectedHyphens := []int{8, 13, 18, 23}
	for _, pos := range expectedHyphens {
		if uuid[pos] != '-' {
			return false
		}
	}
	
	// Check if remaining characters are hex digits
	hexChars := strings.Replace(uuid, "-", "", -1)
	if len(hexChars) != 32 {
		return false
	}
	
	for _, char := range hexChars {
		if !((char >= '0' && char <= '9') || (char >= 'a' && char <= 'f') || (char >= 'A' && char <= 'F')) {
			return false
		}
	}
	
	return true
}

// ValidateJSONSize validates JSON payload size
func ValidateJSONSize(data []byte, maxSizeBytes int) error {
	if len(data) > maxSizeBytes {
		return fmt.Errorf("JSON payload too large, maximum size is %d bytes", maxSizeBytes)
	}
	return nil
}
