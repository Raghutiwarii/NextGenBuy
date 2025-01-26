package utils

import (
	"math/rand"
)

const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// GenerateNanoID generates a unique ID with the specified length and prefix.
func GenerateNanoID(length int, prefix string) (string, error) {
	if length <= 0 {
		return "", ErrInvalidLength
	}

	id := make([]byte, length)
	for i := range id {
		id[i] = charset[rand.Intn(len(charset))]
	}

	return prefix + string(id), nil
}

// ErrInvalidLength is an error returned when the length is invalid.
var ErrInvalidLength = &GenerateIDError{"Invalid length specified, must be greater than 0"}

// GenerateIDError represents an error during ID generation.
type GenerateIDError struct {
	Message string
}

func (e *GenerateIDError) Error() string {
	return e.Message
}
