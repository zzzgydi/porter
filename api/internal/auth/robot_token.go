package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

const minTokenBytes = 16

func GenerateRandomToken(n int) (string, error) {
	if n < minTokenBytes {
		return "", fmt.Errorf("token length must be at least %d bytes", minTokenBytes)
	}
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashRobotToken(token string) string {
	hash, err := bcrypt.GenerateFromPassword([]byte(token), bcrypt.DefaultCost)
	if err != nil {
		// Fallback: this should never happen with valid input
		return ""
	}
	return string(hash)
}

func CheckRobotToken(token, hash string) bool {
	// New format: bcrypt
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(token)); err == nil {
		return true
	}
	// Legacy fallback: SHA256 (hex-encoded, 64 chars)
	if len(hash) == 64 {
		sum := sha256.Sum256([]byte(token))
		return subtle.ConstantTimeCompare([]byte(hash), []byte(hex.EncodeToString(sum[:]))) == 1
	}
	return false
}

func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func IsRobotUsername(username string) bool {
	return strings.HasPrefix(username, "robot$")
}
