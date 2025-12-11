package database

import (
	"database/sql"
	"fmt"

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
			CREATE TABLE IF NOT EXISTS signons (
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
			CREATE TABLE IF NOT EXISTS location_config (
				id INTEGER PRIMARY KEY CHECK (id = 1),
				allow_custom_text INTEGER NOT NULL DEFAULT 1
			);
			INSERT INTO location_config (id, allow_custom_text) VALUES (1, 1);
		`,
	},
	{
		version: 3,
		sql: `
			CREATE TABLE IF NOT EXISTS location_options (
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
			CREATE TABLE IF NOT EXISTS api_key (
				id TEXT NOT NULL PRIMARY KEY,
				salt TEXT NOT NULL,
				hash TEXT NOT NULL,
				created INTEGER,
				last_used INTEGER
			);
		`,
	},
	{
		version: 5,
		sql: `
			CREATE TABLE IF NOT EXISTS allowed_origin (
				url TEXT NOT NULL PRIMARY KEY
			);
		`,
	},
	{
		version: 6,
		sql: `
			CREATE TABLE IF NOT EXISTS campaigns (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				allow_custom_text INTEGER NOT NULL DEFAULT 1,
				created_at INTEGER NOT NULL
			);

			ALTER TABLE signons ADD COLUMN campaign_id TEXT;
			ALTER TABLE location_options ADD COLUMN campaign_id TEXT;

			INSERT INTO campaigns (id, name, allow_custom_text, created_at)
			SELECT
				'00000000-0000-0000-0000-000000000000',
				'Default Campaign',
				COALESCE((SELECT allow_custom_text FROM location_config WHERE id = 1), 1),
				COALESCE((SELECT MIN(created_at) FROM signons), strftime('%s', 'now'));

			UPDATE signons SET campaign_id = '00000000-0000-0000-0000-000000000000' WHERE campaign_id IS NULL;
			UPDATE location_options SET campaign_id = '00000000-0000-0000-0000-000000000000' WHERE campaign_id IS NULL;
		`,
	},
	{
		version: 7,
		sql: `
			CREATE INDEX IF NOT EXISTS idx_signons_campaign_id ON signons(campaign_id);
			CREATE INDEX IF NOT EXISTS idx_signons_campaign_email ON signons(campaign_id, email);
			CREATE INDEX IF NOT EXISTS idx_location_options_campaign ON location_options(campaign_id);
			DROP TABLE IF EXISTS location_config;
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

	// log.Printf("Database initialized at %s", dbPath)
	return nil
}

func runMigrations() error {
	current := getSchemaVersion()
	for _, m := range migrations {
		if m.version > current {
			// log.Printf("Running migration %d...", m.version)
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
