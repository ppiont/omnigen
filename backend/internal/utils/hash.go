package utils

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strings"
)

// NormalizePrompt standardizes a prompt string for consistent hashing
func NormalizePrompt(prompt string) string {
	// Remove extra whitespace
	normalized := strings.TrimSpace(prompt)
	normalized = regexp.MustCompile(`\s+`).ReplaceAllString(normalized, " ")

	// Convert to lowercase for consistency
	normalized = strings.ToLower(normalized)

	// Remove punctuation that doesn't affect meaning
	normalized = regexp.MustCompile(`[.,!?;:]`).ReplaceAllString(normalized, "")

	return normalized
}

// HashContent generates a SHA-256 hash of any serializable data structure
func HashContent(data interface{}) string {
	jsonBytes, _ := json.Marshal(data)
	hash := sha256.Sum256(jsonBytes)
	return hex.EncodeToString(hash[:])
}
