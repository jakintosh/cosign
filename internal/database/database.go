package database

import (
	"database/sql"
	"fmt"
	"log"

	_ "modernc.org/sqlite"
)

var db *sql.DB

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		sql: `
			CREATE TABLE signons (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				email TEXT NOT NULL,
				location TEXT NOT NULL,
				created_at INTEGER NOT NULL
			);
			CREATE INDEX idx_signons_email ON signons(email);
			CREATE INDEX idx_signons_created_at ON signons(created_at);
		`,
	},
	{
		version: 2,
		sql: `
			CREATE TABLE location_config (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				allow_custom_text INTEGER NOT NULL DEFAULT 1
			);
			INSERT INTO location_config (id, allow_custom_text) VALUES (1, 1);
		`,
	},
	{
		version: 3,
		sql: `
			CREATE TABLE location_options (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				value TEXT NOT NULL UNIQUE,
				display_order INTEGER NOT NULL
			);
			CREATE INDEX idx_location_options_order ON location_options(display_order);
		`,
	},
	{
		version: 4,
		sql: `
			CREATE TABLE api_keys (
				id TEXT PRIMARY KEY,
				hash BLOB NOT NULL,
				salt BLOB NOT NULL,
				created_at INTEGER NOT NULL
			);
		`,
	},
	{
		version: 5,
		sql: `
			CREATE TABLE cors_origins (
				origin TEXT PRIMARY KEY,
				created_at INTEGER NOT NULL
			);
		`,
	},
}

// Init initializes the database connection and runs migrations
func Init(dbPath string, useWAL bool) error {
	var err error
	db, err = sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	// Configure SQLite
	db.SetMaxOpenConns(1) // Serial writes

	if _, err := db.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return fmt.Errorf("failed to set busy_timeout: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("failed to enable foreign keys: %w", err)
	}

	if useWAL {
		if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
			return fmt.Errorf("failed to enable WAL: %w", err)
		}
	}

	// Run migrations
	if err := runMigrations(); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	log.Printf("Database initialized at %s", dbPath)
	return nil
}

func runMigrations() error {
	current := getSchemaVersion()
	for _, m := range migrations {
		if m.version > current {
			log.Printf("Running migration %d...", m.version)
			if _, err := db.Exec(m.sql); err != nil {
				return fmt.Errorf("migration %d failed: %w", m.version, err)
			}
			if err := setSchemaVersion(m.version); err != nil {
				return fmt.Errorf("failed to update schema version: %w", err)
			}
		}
	}
	return nil
}

func getSchemaVersion() int {
	var version int
	err := db.QueryRow("PRAGMA user_version").Scan(&version)
	if err != nil {
		return 0
	}
	return version
}

func setSchemaVersion(version int) error {
	_, err := db.Exec(fmt.Sprintf("PRAGMA user_version = %d", version))
	return err
}

// HealthCheck performs a transactional health probe
func HealthCheck() error {
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create temp table
	if _, err := tx.Exec("CREATE TEMP TABLE health_check (id INTEGER)"); err != nil {
		return fmt.Errorf("failed to create temp table: %w", err)
	}

	// Insert row
	if _, err := tx.Exec("INSERT INTO health_check (id) VALUES (1)"); err != nil {
		return fmt.Errorf("failed to insert: %w", err)
	}

	// Read row
	var id int
	if err := tx.QueryRow("SELECT id FROM health_check WHERE id = 1").Scan(&id); err != nil {
		return fmt.Errorf("failed to read: %w", err)
	}

	// Drop table
	if _, err := tx.Exec("DROP TABLE health_check"); err != nil {
		return fmt.Errorf("failed to drop temp table: %w", err)
	}

	return tx.Commit()
}

// DB returns the database connection for use by store implementations
func DB() *sql.DB {
	return db
}
