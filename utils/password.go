package utils

import (
	"regexp"

	"golang.org/x/crypto/bcrypt"
)

// Regex patterns (defined once)
var (
	upper = regexp.MustCompile(`[A-Z]`)
	lower = regexp.MustCompile(`[a-z]`)
	digit = regexp.MustCompile(`[0-9]`)
	// Escaping special characters for Go strings (the original had double-escaping issues)
	special = regexp.MustCompile(`[!@#$%^&*]`)
)

// HashPassword hashes a password using bcrypt.
func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPasswordHash compares a plaintext password with a hashed one.
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// IsStrongPassword checks if a password meets complexity rules.
func IsStrongPassword(p string) bool {
	// Minimum length is often checked via Gin's 'min=8' binding tag, but we re-check here
	// for comprehensive strength verification.
	return len(p) >= 8 &&
		upper.MatchString(p) &&
		lower.MatchString(p) &&
		digit.MatchString(p) &&
		special.MatchString(p)
}
