package auth

import (
	"crypto/rand"
	"encoding/hex"
)

func MakeRefreshToken() (string, error) {
	ret := make([]byte, 32)
	_, err := rand.Read(ret)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(ret), nil
}
