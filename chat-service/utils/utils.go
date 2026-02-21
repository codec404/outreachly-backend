package utils

import (
	"slices"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	return emailRegex.MatchString(email)
}

// HasUppercase reports whether s contains at least one ASCII uppercase letter.
func HasUppercase(s string) bool {
	for _, c := range s {
		if c >= 'A' && c <= 'Z' {
			return true
		}
	}
	return false
}

// HasDigit reports whether s contains at least one ASCII digit.
func HasDigit(s string) bool {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			return true
		}
	}
	return false
}

// GenerateToken returns a cryptographically secure random 32-byte token as a hex string.
func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

// HashToken returns the SHA-256 hex digest of the given token string.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// strconv wrappers — all return an error so callers can detect invalid input.

func FormatInt(s string) (int, error) {
	return strconv.Atoi(s)
}

func FormatInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

func FormatUint(s string) (uint, error) {
	n, err := strconv.ParseUint(s, 10, 0)
	return uint(n), err
}

func FormatUint64(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}

// Contains checks if slice contains the given string.
func Contains(haystack []string, needle string, exactMatch bool) bool {
	if exactMatch {
		for _, s := range haystack {
			if s == needle {
				return true
			}
		}
		return false
	}

	for _, s := range haystack {
		if strings.EqualFold(s, needle) {
			return true
		}
	}
	return false
}

func ContainsGeneric[T comparable](haystack []T, needle T) bool {
	return slices.Contains(haystack, needle)
}

// ConvertToStruct converts a map[string]interface{} to a struct using json tags.
func ConvertToStruct(input any, output any) error {
	jsonData, err := json.Marshal(input)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, output)
}
