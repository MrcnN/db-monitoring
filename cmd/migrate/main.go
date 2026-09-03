package main

import (
	"fmt"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		fatalf("DATABASE_URL is required")
	}

	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}

	command := "up"
	if len(os.Args) > 1 {
		command = os.Args[1]
	}

	m, err := migrate.New(fmt.Sprintf("file://%s", migrationsPath), dbURL)
	if err != nil {
		fatalf("could not create migrate instance: %v", err)
	}
	defer m.Close()

	switch command {
	case "up":
		if err := m.Up(); err != nil && err != migrate.ErrNoChange {
			fatalf("migration up failed: %v", err)
		}
		fmt.Println("Migrations applied successfully.")

	case "down":
		steps := 1
		if len(os.Args) > 2 {
			steps, _ = strconv.Atoi(os.Args[2])
		}
		if err := m.Steps(-steps); err != nil && err != migrate.ErrNoChange {
			fatalf("migration down failed: %v", err)
		}
		fmt.Printf("Rolled back %d migration(s).\n", steps)

	case "version":
		v, dirty, err := m.Version()
		if err != nil {
			fatalf("could not get version: %v", err)
		}
		fmt.Printf("Current version: %d, dirty: %v\n", v, dirty)

	case "drop":
		if err := m.Drop(); err != nil {
			fatalf("drop failed: %v", err)
		}
		fmt.Println("Database dropped.")

	default:
		fatalf("unknown command: %s (use up, down [n], version, drop)", command)
	}
}

func fatalf(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "ERROR: "+format+"\n", args...)
	os.Exit(1)
}
