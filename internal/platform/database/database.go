package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
	"github.com/salambhai/vecdim-backend/ent"
)

// Config holds database connection parameters.
type Config struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
	SSLMode  string
}

// DSN returns the PostgreSQL connection string.
func (c Config) DSN() string {
	if c.Host == "" {
		c.Host = "localhost"
	}
	if c.Port == 0 {
		c.Port = 5432
	}
	if c.User == "" {
		c.User = "vecdim_admin"
	}
	if c.Password == "" {
		c.Password = "vecdim_secret_2026"
	}
	if c.DBName == "" {
		c.DBName = "vecdim"
	}
	if c.SSLMode == "" {
		c.SSLMode = "disable"
	}
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.DBName, c.SSLMode)
}

// Connect initializes PostgreSQL connection and runs automatic Ent schema migrations.
func Connect(ctx context.Context, cfg Config) (*ent.Client, *sql.DB, error) {
	dsn := cfg.DSN()

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open raw sql connection: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.PingContext(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	// Initialize Ent client with the established driver
	client, err := ent.Open("postgres", dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to open ent client: %w", err)
	}

	// Run automatic schema migration
	if err := client.Schema.Create(ctx); err != nil {
		log.Printf("Warning: Ent schema create: %v", err)
	}

	return client, db, nil
}
