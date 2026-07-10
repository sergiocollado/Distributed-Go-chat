package token

import (
	"crypto/rand"
	"encoding/hex"
)

/*
Tokens here are transmitted in plaintext over TCP. For production use,
you should use TLS encryption to protect tokens in transit, hash tokens
before storage so a database breach doesn't expose them, and implement
rate limiting on reconnection attempts to prevent brute force attacks.
*/

// GenerateToken returns a secure random 16-byte hex token
func GenerateToken() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
