//go:build integration

package integration

import (
	"context"
	"testing"
	"time"

	"github.com/dbplatform/api/internal/auth"
	"github.com/dbplatform/api/internal/config"
	"github.com/dbplatform/api/internal/user"
	"github.com/dbplatform/api/tests/integration/testhelper"
)

func TestIntegrationAuth_RegisterAndLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	pg, cleanup := testhelper.SetupPostgres(ctx, t, testhelper.MigrationsPath())
	defer cleanup()

	userRepo := user.NewRepository(pg.Pool)
	sessionRepo := user.NewSessionRepository(pg.Pool)
	jwtSvc := auth.NewJWTService(config.JWTConfig{
		Secret:          "test-secret-for-integration-tests-minimum-32-chars",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	})
	authSvc := auth.NewService(userRepo, sessionRepo, jwtSvc, 7*24*time.Hour)

	t.Run("Register", func(t *testing.T) {
		u, err := authSvc.Register(ctx, auth.RegisterInput{
			Email:    "test@example.com",
			Password: "SecureP@ss123!",
			FullName: "Integration Test User",
		})
		if err != nil {
			t.Fatalf("Register failed: %v", err)
		}
		if u.ID.String() == "" {
			t.Fatal("expected non-empty user ID")
		}
		if u.Email != "test@example.com" {
			t.Fatalf("expected email 'test@example.com', got %q", u.Email)
		}
		if u.Role != user.RoleViewer {
			t.Fatalf("expected role %q, got %q", user.RoleViewer, u.Role)
		}
		if u.PasswordHash == "SecureP@ss123!" {
			t.Fatal("password should be hashed, not plaintext")
		}
	})

	t.Run("Register_Duplicate", func(t *testing.T) {
		_, err := authSvc.Register(ctx, auth.RegisterInput{
			Email:    "test@example.com",
			Password: "AnotherPass123!",
			FullName: "Duplicate User",
		})
		if err == nil {
			t.Fatal("expected error for duplicate email registration")
		}
	})

	t.Run("Login_Valid", func(t *testing.T) {
		u, tokens, err := authSvc.Login(ctx, auth.LoginInput{
			Email:     "test@example.com",
			Password:  "SecureP@ss123!",
			UserAgent: "test-agent",
			IPAddress: "127.0.0.1",
		})
		if err != nil {
			t.Fatalf("Login failed: %v", err)
		}
		if tokens.AccessToken == "" {
			t.Fatal("expected non-empty access token")
		}
		if tokens.RefreshToken == "" {
			t.Fatal("expected non-empty refresh token")
		}
		if u.Email != "test@example.com" {
			t.Fatalf("expected email 'test@example.com', got %q", u.Email)
		}
	})

	t.Run("Login_WrongPassword", func(t *testing.T) {
		_, _, err := authSvc.Login(ctx, auth.LoginInput{
			Email:    "test@example.com",
			Password: "wrong-password",
		})
		if err == nil {
			t.Fatal("expected error for wrong password")
		}
	})

	t.Run("Login_UnknownEmail", func(t *testing.T) {
		_, _, err := authSvc.Login(ctx, auth.LoginInput{
			Email:    "nonexistent@example.com",
			Password: "irrelevant",
		})
		if err == nil {
			t.Fatal("expected error for non-existent email")
		}
	})

	t.Run("Refresh_TokenRotation", func(t *testing.T) {
		_, tokens, err := authSvc.Login(ctx, auth.LoginInput{
			Email:     "test@example.com",
			Password:  "SecureP@ss123!",
			UserAgent: "test-agent",
			IPAddress: "127.0.0.1",
		})
		if err != nil {
			t.Fatalf("Login for refresh test failed: %v", err)
		}

		_, newTokens, err := authSvc.Refresh(ctx, tokens.RefreshToken, "test-agent", "127.0.0.1")
		if err != nil {
			t.Fatalf("Refresh failed: %v", err)
		}
		if newTokens.AccessToken == "" {
			t.Fatal("expected new access token")
		}
		_, _, err = authSvc.Refresh(ctx, tokens.RefreshToken, "test-agent", "127.0.0.1")
		if err == nil {
			t.Fatal("expected error when reusing revoked refresh token")
		}
	})
}
