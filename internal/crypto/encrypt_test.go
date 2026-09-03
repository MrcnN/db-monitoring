package crypto_test

import (
	"strings"
	"testing"

	"github.com/dbplatform/api/internal/crypto"
)

const testKey = "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"

func TestNewEncryptor_ValidKey(t *testing.T) {
	enc, err := crypto.NewEncryptor(testKey)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if enc == nil {
		t.Fatal("expected non-nil encryptor")
	}
}

func TestNewEncryptor_InvalidHex(t *testing.T) {
	_, err := crypto.NewEncryptor("not-valid-hex")
	if err == nil {
		t.Fatal("expected error for invalid hex key")
	}
}

func TestNewEncryptor_WrongKeyLength(t *testing.T) {
	_, err := crypto.NewEncryptor("0102030405060708090a0b0c0d0e0f10")
	if err == nil {
		t.Fatal("expected error for wrong key length")
	}
}

func TestEncryptDecrypt_Roundtrip(t *testing.T) {
	enc, err := crypto.NewEncryptor(testKey)
	if err != nil {
		t.Fatalf("setup error: %v", err)
	}

	plaintext := "super-secret-password-123!"

	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt error: %v", err)
	}
	if ciphertext == "" {
		t.Fatal("expected non-empty ciphertext")
	}
	if ciphertext == plaintext {
		t.Fatal("ciphertext should differ from plaintext")
	}

	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt error: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("expected %q, got %q", plaintext, decrypted)
	}
}

func TestEncrypt_NonDeterministic(t *testing.T) {
	enc, _ := crypto.NewEncryptor(testKey)
	c1, _ := enc.Encrypt("test")
	c2, _ := enc.Encrypt("test")
	if c1 == c2 {
		t.Fatal("expected different ciphertexts for same plaintext (nonce must be random)")
	}
}

func TestEncrypt_EmptyString(t *testing.T) {
	enc, _ := crypto.NewEncryptor(testKey)
	ciphertext, err := enc.Encrypt("")
	if err != nil {
		t.Fatalf("Encrypt empty string error: %v", err)
	}
	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt empty string error: %v", err)
	}
	if decrypted != "" {
		t.Fatalf("expected empty string, got %q", decrypted)
	}
}

func TestDecrypt_InvalidBase64(t *testing.T) {
	enc, _ := crypto.NewEncryptor(testKey)
	_, err := enc.Decrypt("not-valid-base64!!!")
	if err == nil {
		t.Fatal("expected error for invalid base64")
	}
}

func TestDecrypt_TruncatedCiphertext(t *testing.T) {
	enc, _ := crypto.NewEncryptor(testKey)
	_, err := enc.Decrypt("aGVsbG8=") // base64("hello") — too short
	if err == nil {
		t.Fatal("expected error for ciphertext too short")
	}
}

func TestEncrypt_LongString(t *testing.T) {
	enc, _ := crypto.NewEncryptor(testKey)
	plaintext := strings.Repeat("a", 10000)
	ciphertext, err := enc.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encrypt long string error: %v", err)
	}
	decrypted, err := enc.Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Decrypt long string error: %v", err)
	}
	if decrypted != plaintext {
		t.Fatalf("roundtrip mismatch for long string")
	}
}
