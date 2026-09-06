package auth_test

import (
	"testing"

	"github.com/dbplatform/api/internal/auth"
)

func TestHashPassword_Success(t *testing.T) {
	hash, err := auth.HashPassword("MySecret123!")
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if hash == "" {
		t.Fatal("expected non-empty hash")
	}
	if hash == "MySecret123!" {
		t.Fatal("hash should not equal plaintext")
	}
}

func TestCheckPassword_Correct(t *testing.T) {
	password := "Correct-Horse-Battery-Staple"
	hash, _ := auth.HashPassword(password)
	if err := auth.CheckPassword(password, hash); err != nil {
		t.Fatalf("expected password to match, got: %v", err)
	}
}

func TestCheckPassword_Wrong(t *testing.T) {
	hash, _ := auth.HashPassword("correct-password")
	if err := auth.CheckPassword("wrong-password", hash); err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestCheckPassword_EmptyPassword(t *testing.T) {
	hash, _ := auth.HashPassword("some-password")
	if err := auth.CheckPassword("", hash); err == nil {
		t.Fatal("expected error for empty password")
	}
}

func TestHashPassword_DifferentHashes(t *testing.T) {
	hash1, _ := auth.HashPassword("same-password")
	hash2, _ := auth.HashPassword("same-password")
	if hash1 == hash2 {
		t.Fatal("expected different bcrypt hashes for same password (bcrypt should be salted)")
	}
}

func TestHashPassword_LongPassword(t *testing.T) {
	long := "a"
	for i := 0; i < 100; i++ {
		long += "a"
	}
	_, err := auth.HashPassword(long)
	if err == nil {
		t.Fatalf("expected error for long password exceeding bcrypt 72 byte limit, got nil")
	}
}
