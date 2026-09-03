package auth_test

import (
	"testing"
	"time"

	"github.com/dbplatform/api/internal/auth"
	"github.com/dbplatform/api/internal/config"
	"github.com/dbplatform/api/internal/user"
	"github.com/google/uuid"
)

func testJWTService() *auth.JWTService {
	return auth.NewJWTService(config.JWTConfig{
		Secret:          "test-secret-for-testing-only-must-be-at-least-32-chars",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	})
}

func testUser() *user.User {
	return &user.User{
		ID:       uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		Email:    "test@example.com",
		FullName: "Test User",
		Role:     user.RoleViewer,
		IsActive: true,
	}
}

func TestGenerateAccessToken_Success(t *testing.T) {
	svc := testJWTService()
	u := testUser()

	token, err := svc.GenerateAccessToken(u)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestValidateAccessToken_Valid(t *testing.T) {
	svc := testJWTService()
	u := testUser()

	token, _ := svc.GenerateAccessToken(u)
	claims, err := svc.ValidateAccessToken(token)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}
	if claims.UserID != u.ID.String() {
		t.Fatalf("expected user_id %q, got %q", u.ID, claims.UserID)
	}
	if claims.Email != u.Email {
		t.Fatalf("expected email %q, got %q", u.Email, claims.Email)
	}
	if claims.Role != string(u.Role) {
		t.Fatalf("expected role %q, got %q", u.Role, claims.Role)
	}
	if claims.TokenType != "access" {
		t.Fatalf("expected token_type 'access', got %q", claims.TokenType)
	}
}

func TestValidateAccessToken_WrongSecret(t *testing.T) {
	svc1 := testJWTService()
	svc2 := auth.NewJWTService(config.JWTConfig{
		Secret:         "completely-different-secret-that-is-at-least-32-chars",
		AccessTokenTTL: 15 * time.Minute,
	})

	token, _ := svc1.GenerateAccessToken(testUser())
	_, err := svc2.ValidateAccessToken(token)
	if err == nil {
		t.Fatal("expected error for token signed with different secret")
	}
}

func TestValidateAccessToken_Tampered(t *testing.T) {
	svc := testJWTService()
	token, _ := svc.GenerateAccessToken(testUser())

	tampered := token + "x"
	_, err := svc.ValidateAccessToken(tampered)
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestValidateAccessToken_Invalid(t *testing.T) {
	svc := testJWTService()
	_, err := svc.ValidateAccessToken("not.a.valid.jwt.token")
	if err == nil {
		t.Fatal("expected error for invalid JWT")
	}
}

func TestGenerateRefreshToken_UniqueTokens(t *testing.T) {
	svc := testJWTService()

	raw1, hash1, err1 := svc.GenerateRefreshToken()
	raw2, hash2, err2 := svc.GenerateRefreshToken()

	if err1 != nil || err2 != nil {
		t.Fatalf("unexpected errors: %v, %v", err1, err2)
	}
	if raw1 == raw2 {
		t.Fatal("expected unique raw refresh tokens")
	}
	if hash1 == hash2 {
		t.Fatal("expected unique refresh token hashes")
	}
}

func TestGenerateRefreshToken_HashDiffersFromRaw(t *testing.T) {
	svc := testJWTService()
	raw, hash, err := svc.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if raw == hash {
		t.Fatal("hash should differ from raw token")
	}
}
