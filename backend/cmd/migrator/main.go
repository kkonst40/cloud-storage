package main

import (
	"log"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/kkonst40/cloud-storage/backend/internal/config"
	"github.com/kkonst40/cloud-storage/backend/internal/storage"
	"github.com/pressly/goose/v3"
)

func main() {
	cfg := config.MustNew("")
	dbClient, err := storage.NewClient(cfg)
	if err != nil {
		log.Fatalf("failed to connect to DB: %v", err)
	}

	db := dbClient.DB()
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("failed to set dialect: %v", err)
	}

	log.Println("Applying migrations...")
	if err := goose.Up(db, "db/migrations"); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Println("Migrations applied successfully!")
}
