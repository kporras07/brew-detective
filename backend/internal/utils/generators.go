package utils

import (
	"crypto/rand"
	"strings"
)

// GenerateOrderID generates a random 6-character order ID
func GenerateOrderID() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 6
	
	b := make([]byte, length)
	for i := range b {
		// Generate random index
		randomBytes := make([]byte, 1)
		_, err := rand.Read(randomBytes)
		if err != nil {
			// Fallback if crypto/rand fails
			randomBytes[0] = byte(i * 7 % len(charset))
		}
		b[i] = charset[randomBytes[0]%byte(len(charset))]
	}
	
	return strings.ToUpper(string(b))
}