package auth

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "super-secure-test-secret"
	expiresIn := 15 * time.Minute

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error creating token, got: %v", err)
	}

	if token == "" {
		t.Fatal("expected token to not be empty")
	}

	// Real-world JWT format: header.payload.signature (3 parts separated by dots)
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("expected token to have 3 parts separated by '.', got %d parts", len(parts))
	}

	validatedID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("expected token to be valid, got: %v", err)
	}

	if validatedID != userID {
		t.Fatalf("expected subject UUID %v, got %v", userID, validatedID)
	}
}

func TestExpiredJWT_LapsedTime(t *testing.T) {
	// In the real world, tokens expire after their TTL has elapsed
	userID := uuid.New()
	secret := "super-secure-test-secret"
	expiresIn := 50 * time.Millisecond

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error creating token, got: %v", err)
	}

	// Allow the token to naturally expire
	time.Sleep(100 * time.Millisecond)

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("expected expired token to be rejected, but validation succeeded")
	}

	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got: %v", err)
	}
}

func TestExpiredJWT_PastExpiration(t *testing.T) {
	userID := uuid.New()
	secret := "super-secure-test-secret"

	// Created with an already-past expiration timestamp
	token, err := MakeJWT(userID, secret, -time.Hour)
	if err != nil {
		t.Fatalf("expected no error creating token, got: %v", err)
	}

	_, err = ValidateJWT(token, secret)
	if err == nil {
		t.Fatal("expected token with past expiration to be rejected, got nil")
	}

	if !errors.Is(err, jwt.ErrTokenExpired) {
		t.Fatalf("expected ErrTokenExpired, got: %v", err)
	}
}

func TestWrongSecretJWT(t *testing.T) {
	userID := uuid.New()
	correctSecret := "correct-secret-key"
	wrongSecret := "attacker-or-different-secret-key"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, correctSecret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error creating token, got: %v", err)
	}

	_, err = ValidateJWT(token, wrongSecret)
	if err == nil {
		t.Fatal("expected token validated with wrong secret to be rejected, got nil")
	}

	if !errors.Is(err, jwt.ErrTokenSignatureInvalid) {
		t.Fatalf("expected ErrTokenSignatureInvalid, got: %v", err)
	}
}

func TestTamperedJWTPayload(t *testing.T) {
	// In the real world, attackers may modify the claims payload in transit
	userID := uuid.New()
	secret := "super-secure-test-secret"
	expiresIn := time.Hour

	token, err := MakeJWT(userID, secret, expiresIn)
	if err != nil {
		t.Fatalf("expected no error creating token, got: %v", err)
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatalf("malformed token generated: %v", token)
	}

	// Tamper with the payload (middle part)
	tamperedPayload := parts[1] + "tampered"
	tamperedToken := parts[0] + "." + tamperedPayload + "." + parts[2]

	_, err = ValidateJWT(tamperedToken, secret)
	if err == nil {
		t.Fatal("expected tampered token to fail validation, but validation succeeded")
	}
}

func TestMalformedJWTs(t *testing.T) {
	secret := "super-secure-test-secret"

	malformedTokens := []string{
		"",
		"not-a-jwt",
		"header.payload",
		"header.payload.signature.extra",
		"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid-payload.signature",
	}

	for _, token := range malformedTokens {
		_, err := ValidateJWT(token, secret)
		if err == nil {
			t.Fatalf("expected validation error for malformed token %q, got nil", token)
		}
	}
}

func TestInvalidSubjectUUID(t *testing.T) {
	// A token signed with the correct secret, but whose subject is not a valid UUID
	secret := "super-secure-test-secret"
	now := time.Now()

	claims := jwt.RegisteredClaims{
		Issuer:    "chirpy-access",
		Subject:   "not-a-uuid-string-like-admin",
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour)),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed to create token with invalid subject: %v", err)
	}

	_, err = ValidateJWT(signedToken, secret)
	if err == nil {
		t.Fatal("expected error when token subject is not a valid UUID, got nil")
	}
}

func TestPasswordHashing(t *testing.T) {
	password := "my-secret-password"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("unexpected error hashing password: %v", err)
	}

	if hash == password {
		t.Fatalf("hash should not equal plaintext password")
	}

	match, err := CheckPassword(password, hash)
	if err != nil {
		t.Fatalf("unexpected error checking password: %v", err)
	}
	if !match {
		t.Fatalf("expected password to match hash")
	}

	wrongMatch, err := CheckPassword("wrong-password", hash)
	if err != nil {
		t.Fatalf("unexpected error checking wrong password: %v", err)
	}
	if wrongMatch {
		t.Fatalf("expected wrong password to not match hash")
	}
}

