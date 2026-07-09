package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestHashPassword(t *testing.T) {
	password := "correcthorsebatterystaple"

	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("HashPassword returned error: %v", err)
	}
	if hash == "" {
		t.Fatal("HashPassword returned an empty hash")
	}
	if hash == password {
		t.Fatal("hash should not equal the plaintext password")
	}
}

func TestHashPasswordIsRandomized(t *testing.T) {
	password := "correcthorsebatterystaple"

	hash1, err := HashPassword(password)
	if err != nil {
		t.Fatalf("first HashPassword call returned error: %v", err)
	}
	hash2, err := HashPassword(password)
	if err != nil {
		t.Fatalf("second HashPassword call returned error: %v", err)
	}
	if hash1 == hash2 {
		t.Fatal("expected different hashes for the same password due to random salt")
	}
}

func TestCheckPasswordHash(t *testing.T) {
	password := "correcthorsebatterystaple"
	hash, err := HashPassword(password)
	if err != nil {
		t.Fatalf("could not hash password for test setup: %v", err)
	}

	tests := []struct {
		name     string
		password string
		hash     string
		want     bool
		wantErr  bool
	}{
		{name: "correct password", password: password, hash: hash, want: true},
		{name: "incorrect password", password: "wrongpassword", hash: hash, want: false},
		{name: "malformed hash", password: password, hash: "not-a-real-hash", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CheckPasswordHash(tt.password, tt.hash)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected an error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMakeAndValidateJWT(t *testing.T) {
	userID := uuid.New()
	secret := "test-secret"

	token, err := MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("MakeJWT returned error: %v", err)
	}
	if token == "" {
		t.Fatal("MakeJWT returned an empty token")
	}

	gotID, err := ValidateJWT(token, secret)
	if err != nil {
		t.Fatalf("ValidateJWT returned error: %v", err)
	}
	if gotID != userID {
		t.Errorf("got user ID %v, want %v", gotID, userID)
	}
}

func TestValidateJWTErrors(t *testing.T) {
	userID := uuid.New()

	validToken, err := MakeJWT(userID, "correct-secret", time.Hour)
	if err != nil {
		t.Fatalf("could not create token for test setup: %v", err)
	}

	expiredToken, err := MakeJWT(userID, "correct-secret", -time.Hour)
	if err != nil {
		t.Fatalf("could not create expired token for test setup: %v", err)
	}

	tests := []struct {
		name   string
		token  string
		secret string
	}{
		{name: "expired token", token: expiredToken, secret: "correct-secret"},
		{name: "wrong secret", token: validToken, secret: "wrong-secret"},
		{name: "malformed token", token: "not-a-real-token", secret: "correct-secret"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ValidateJWT(tt.token, tt.secret)
			if err == nil {
				t.Fatal("expected an error, got nil")
			}
		})
	}
}
