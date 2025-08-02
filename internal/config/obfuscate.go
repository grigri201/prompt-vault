package config

import (
	"crypto/sha256"
	"encoding/base64"
	"os"
	"runtime"
)

// generateXorKey creates a machine-specific XOR key for obfuscation
// This combines multiple machine-specific identifiers to create a unique key
func generateXorKey() byte {
	// Collect machine-specific data
	hostname, _ := os.Hostname()
	username := os.Getenv("USER")
	if username == "" {
		username = os.Getenv("USERNAME") // Windows fallback
	}

	// Combine with runtime info
	data := hostname + username + runtime.GOOS + runtime.GOARCH

	// If no machine-specific data available, use a default
	if data == "" {
		return byte(0xAB)
	}

	// Hash and return first byte
	hash := sha256.Sum256([]byte(data))
	return hash[0]
}

// getXorKey returns the XOR key, generating it once and caching
var cachedXorKey *byte

func getXorKey() byte {
	if cachedXorKey == nil {
		key := generateXorKey()
		cachedXorKey = &key
	}
	return *cachedXorKey
}

// obfuscate applies simple XOR obfuscation to the token
// This is not encryption - it's just to avoid storing tokens in plain text
func obfuscate(token string) string {
	if token == "" {
		return ""
	}

	bytes := []byte(token)
	result := make([]byte, len(bytes))

	for i, b := range bytes {
		result[i] = b ^ getXorKey()
	}

	// Base64 encode to ensure the result is safely storable
	return base64.StdEncoding.EncodeToString(result)
}

// deobfuscate reverses the obfuscation to retrieve the original token
func deobfuscate(obfuscatedToken string) string {
	if obfuscatedToken == "" {
		return ""
	}

	// Base64 decode
	bytes, err := base64.StdEncoding.DecodeString(obfuscatedToken)
	if err != nil {
		// If decoding fails, assume it's not obfuscated
		return obfuscatedToken
	}

	result := make([]byte, len(bytes))
	for i, b := range bytes {
		result[i] = b ^ getXorKey()
	}

	return string(result)
}
