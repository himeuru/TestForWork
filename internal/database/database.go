package database

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"log"
	"strings"
	"testForWork/internal/config"
	"time"
)

//go:embed migrations/*.sql
var embedMigrations embed.FS

func ConnectAndSetup(cfg config.DBConfig) (*sql.DB, error) {

	if err := createDatabaseIfNotExists(cfg); err != nil {
		return nil, err
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DatabaseName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	goose.SetBaseFS(embedMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("failed to set dialect: %w", err)
	}

	if err := goose.Up(db, "migrations"); err != nil {
		return nil, fmt.Errorf("failed to apply migrations: %w", err)
	}

	log.Println("Database migrations applied successfully")
	return db, nil
}

func createDatabaseIfNotExists(cfg config.DBConfig) error {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("failed to connect to postgres: %w", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx,
		fmt.Sprintf("SELECT 1 FROM pg_database WHERE datname = '%s'", cfg.DatabaseName))

	if err != nil && strings.Contains(err.Error(), "does not exist") {
		_, err = db.ExecContext(ctx, fmt.Sprintf("CREATE DATABASE %s", cfg.DatabaseName))
		if err != nil {
			return fmt.Errorf("failed to create database: %w", err)
		}
		log.Printf("Database %s created", cfg.DatabaseName)
	}
	return nil
}
