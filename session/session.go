package session

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateSessionID creates a cryptographically-secure random 128-byte string; referentially
// transparent
func GenerateSessionID() (string, error) {
	b := make([]byte, 128) // 128 bytes should be more than ample, which yields a 172 character string

	_, err := rand.Read(b)

	if err != nil {
		return "", err
	}

	return base64.URLEncoding.EncodeToString(b), nil
}
