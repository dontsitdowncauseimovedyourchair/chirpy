package auth

import (
	"fmt"

	"github.com/alexedwards/argon2id"
)

func HashPassword(password string) (string, error) {
	hashed, err := argon2id.CreateHash(password, argon2id.DefaultParams)
	if err != nil {
		return "", fmt.Errorf("flop hashing password: %w", err)
	}
	return hashed, nil
}

func CheckPassword(password string, hash string) (bool, error) {
	equality, err := argon2id.ComparePasswordAndHash(password, hash)
	if err != nil {
		return false, fmt.Errorf("flop checking password: %w", err)
	}
	return equality, nil
}
