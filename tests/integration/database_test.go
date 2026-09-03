//go:build integration

package integration

import (
	"context"
	"testing"

	"github.com/dbplatform/api/internal/crypto"
	"github.com/dbplatform/api/internal/database"
	"github.com/dbplatform/api/internal/user"
	"github.com/dbplatform/api/tests/integration/testhelper"
	"github.com/google/uuid"
)

const testEncKey = "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"

func TestIntegrationDatabase_CRUD(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	ctx := context.Background()
	pg, cleanup := testhelper.SetupPostgres(ctx, t, testhelper.MigrationsPath())
	defer cleanup()

	userRepo := user.NewRepository(pg.Pool)
	testUser := &user.User{
		Email:        "dbtest@example.com",
		PasswordHash: "$2a$12$hash",
		FullName:     "DB Test User",
		Role:         user.RoleAdmin,
		IsActive:     true,
	}
	if err := userRepo.Create(ctx, testUser); err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}

	encryptor, err := crypto.NewEncryptor(testEncKey)
	if err != nil {
		t.Fatalf("encryptor setup failed: %v", err)
	}
	dbRepo := database.NewRepository(pg.Pool)
	dbSvc := database.NewService(dbRepo, encryptor)

	const testPassword = "test-db-password-123"

	var createdID uuid.UUID

	t.Run("Create", func(t *testing.T) {
		encPwd, err := encryptor.Encrypt(testPassword)
		if err != nil {
			t.Fatalf("encrypt password failed: %v", err)
		}

		db := &database.MonitoredDatabase{
			Name:                "Test PostgreSQL",
			Description:         "Integration test database",
			Type:                database.TypePostgreSQL,
			Host:                "localhost",
			Port:                5432,
			DatabaseName:        "testdb",
			Username:            "testuser",
			SSLMode:             database.SSLModeDisable,
			MonitoringInterval:  15,
			IsMonitoringEnabled: true,
			Status:              database.StatusActive,
			CreatedBy:           testUser.ID,
		}

		if err := dbSvc.Create(ctx, db, encPwd); err != nil {
			t.Fatalf("Create database failed: %v", err)
		}

		if db.ID == uuid.Nil {
			t.Fatal("expected non-nil database ID after create")
		}
		createdID = db.ID
	})

	t.Run("Get", func(t *testing.T) {
		db, err := dbSvc.GetByID(ctx, createdID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		if db.Name != "Test PostgreSQL" {
			t.Fatalf("expected name 'Test PostgreSQL', got %q", db.Name)
		}
	})

	t.Run("GetWithCredentials_PasswordDecryptable", func(t *testing.T) {
		_, encPwd, err := dbSvc.GetWithCredentials(ctx, createdID)
		if err != nil {
			t.Fatalf("GetWithCredentials failed: %v", err)
		}
		decrypted, err := encryptor.Decrypt(encPwd)
		if err != nil {
			t.Fatalf("Decrypt failed: %v", err)
		}
		if decrypted != testPassword {
			t.Fatalf("expected password %q, got %q", testPassword, decrypted)
		}
	})

	t.Run("List", func(t *testing.T) {
		dbs, err := dbSvc.List(ctx)
		if err != nil {
			t.Fatalf("List failed: %v", err)
		}
		if len(dbs) == 0 {
			t.Fatal("expected at least one database in list")
		}
	})

	t.Run("Update", func(t *testing.T) {
		db, err := dbSvc.GetByID(ctx, createdID)
		if err != nil {
			t.Fatalf("GetByID failed: %v", err)
		}
		db.Name = "Updated Database Name"
		db.MonitoringInterval = 30

		if err := dbSvc.Update(ctx, db, nil); err != nil {
			t.Fatalf("Update failed: %v", err)
		}

		updated, err := dbSvc.GetByID(ctx, createdID)
		if err != nil {
			t.Fatalf("GetByID after update failed: %v", err)
		}
		if updated.Name != "Updated Database Name" {
			t.Fatalf("expected name 'Updated Database Name', got %q", updated.Name)
		}
		if updated.MonitoringInterval != 30 {
			t.Fatalf("expected interval 30, got %d", updated.MonitoringInterval)
		}
	})

	t.Run("Delete", func(t *testing.T) {
		if err := dbSvc.Delete(ctx, createdID); err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		_, err := dbSvc.GetByID(ctx, createdID)
		if err == nil {
			t.Fatal("expected error when getting soft-deleted database")
		}
	})

	t.Run("Get_NotFound", func(t *testing.T) {
		_, err := dbSvc.GetByID(ctx, uuid.New())
		if err == nil {
			t.Fatal("expected not-found error for non-existent database ID")
		}
	})
}
