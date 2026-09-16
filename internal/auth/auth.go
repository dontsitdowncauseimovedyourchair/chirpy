package auth

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
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

func MakeJWT(userID uuid.UUID, tokenSecret string, expiresIn time.Duration) (string, error) {
	now := time.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		Subject:   userID.String(),
		ExpiresAt: jwt.NewNumericDate(now.Add(expiresIn).UTC()),
		NotBefore: jwt.NewNumericDate(now.UTC()),
		IssuedAt:  jwt.NewNumericDate(now.UTC()),
	})

	signedString, err := token.SignedString([]byte(tokenSecret))
	if err != nil {
		return "", err
	}

	return signedString, nil
}

func ValidateJWT(tokenString, tokenSecret string) (uuid.UUID, error) {
	claims, err := jwt.ParseWithClaims(tokenString, &jwt.RegisteredClaims{}, func(token *jwt.Token) (any, error) {
		return []byte(tokenSecret), nil
	})
	if err != nil {
		return uuid.UUID{}, err
	}

	claimUUID, err := claims.Claims.GetSubject()
	if err != nil {
		return uuid.UUID{}, err
	}
	return uuid.Parse(claimUUID)
}

func GetBearerToken(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if len(auth) == 0 {
		return "", fmt.Errorf("auth headers missing")
	}
	token := strings.ReplaceAll(auth, "Bearer ", "")
	return token, nil
}

func GetAPIKey(headers http.Header) (string, error) {
	auth := headers.Get("Authorization")
	if len(auth) == 0 {
		return "", fmt.Errorf("auth headers missing")
	}
	token := strings.ReplaceAll(auth, "ApiKey ", "")
	return token, nil
}
