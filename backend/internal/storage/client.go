package storage

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kkonst40/cloud-storage/backend/internal/config"
)

type Client struct {
	db *sql.DB
}

func NewClient(cfg *config.Config) (*Client, error) {
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.PostgresUser,
		cfg.PostgresPassword,
		cfg.PostgresHost,
		cfg.PostgresPort,
		cfg.PostgresDBName,
	)

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return nil, fmt.Errorf("database connection opening error: %w", err)
	}

	db.SetMaxOpenConns(cfg.DBMaxOpenConns)
	db.SetConnMaxLifetime(time.Minute * time.Duration(cfg.DBConnMaxLifetimeMinutes))
	db.SetMaxIdleConns(cfg.DBMaxIdleConns)

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("database is not responding: %w", err)
	}

	return &Client{db: db}, nil
}

func (cl *Client) DB() *sql.DB {
	return cl.db
}

func (cl *Client) Shutdown() error {
	if err := cl.db.Close(); err != nil {
		return fmt.Errorf("storage.Client: Shutdown error: %w", err)
	}
	slog.Info("DB shutdown")

	return nil
}
