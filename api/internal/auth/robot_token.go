package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"strings"
)

func GenerateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func HashRobotToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func ConstantTimeEqual(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func IsRobotUsername(username string) bool {
	return strings.HasPrefix(username, "robot$")
}

func ParseRobotUsername(username string) (projectName, tokenName string, err error) {
	if !IsRobotUsername(username) {
		return "", "", fmt.Errorf("not a robot username")
	}
	rest := strings.TrimPrefix(username, "robot$")
	parts := strings.SplitN(rest, "-", 2)
	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid robot username format")
	}
	return parts[0], parts[1], nil
}
