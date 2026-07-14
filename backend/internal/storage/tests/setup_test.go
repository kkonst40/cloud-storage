//go:build integration

package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"testing"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB

func TestMain(m *testing.M) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase("test_db"),
		postgres.WithUsername("test_user"),
		postgres.WithPassword("test_pass"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		fmt.Printf("Failed to start Testcontainers: %v\n", err)
		os.Exit(1)
	}

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		fmt.Printf("Failed to get connection string: %v\n", err)
		os.Exit(1)
	}

	testDB, err = sql.Open("pgx", connStr)
	if err != nil {
		fmt.Printf("Failed to connect to database: %v\n", err)
		os.Exit(1)
	}

	if err := testDB.PingContext(ctx); err != nil {
		fmt.Printf("Ping failed: %v\n", err)
		os.Exit(1)
	}

	if err := migrateSchema(testDB); err != nil {
		fmt.Printf("Failed to apply DDL schema: %v\n", err)
		os.Exit(1)
	}

	if err := migrateSchema(testDB); err != nil {
		fmt.Printf("Failed to apply DDL schema: %v\n", err)
		os.Exit(1)
	}

	code := m.Run()

	testDB.Close()
	pgContainer.Terminate(ctx)

	os.Exit(code)
}

func migrateSchema(db *sql.DB) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return err
	}

	return goose.Up(db, "../../../db/migrations")
}
