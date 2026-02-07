package database

import (
	"database/sql"
	"fmt"

	"git.sr.ht/~jakintosh/command-go/pkg/cors"
	"git.sr.ht/~jakintosh/command-go/pkg/keys"
	_ "modernc.org/sqlite"
)

type Options struct {
	Path string
	WAL  bool
}

type DB struct {
	Conn      *sql.DB
	KeysStore *keys.SQLStore
	CORSStore *cors.SQLStore
}

type migration struct {
	version int
	sql     string
}

var migrations = []migration{
	{
		version: 1,
		sql: `
			CREATE TABLE IF NOT EXISTS campaigns (
				id TEXT PRIMARY KEY,
				name TEXT NOT NULL,
				allow_custom_text INTEGER NOT NULL DEFAULT 1,
				created_at INTEGER NOT NULL
			);

			CREATE TABLE IF NOT EXISTS locations (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				campaign_id TEXT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
				value TEXT NOT NULL,
				display_order INTEGER NOT NULL,
				UNIQUE(campaign_id, value)
			);
			CREATE INDEX IF NOT EXISTS idx_locations_campaign_order ON locations(campaign_id, display_order);

			CREATE TABLE IF NOT EXISTS signons (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				campaign_id TEXT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
				name TEXT NOT NULL,
				email TEXT NOT NULL,
				location TEXT NOT NULL,
				created_at INTEGER NOT NULL,
				UNIQUE(campaign_id, email)
			);
			CREATE INDEX IF NOT EXISTS idx_signons_campaign ON signons(campaign_id);
			CREATE INDEX IF NOT EXISTS idx_signons_campaign_created ON signons(campaign_id, created_at);
		`,
	},
}

func Open(
	opts Options,
) (
	*DB,
	error,
) {
	conn, err := sql.Open("sqlite", opts.Path)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if err := configure(conn, opts.WAL); err != nil {
		conn.Close()
		return nil, err
	}

	if err := runMigrations(conn); err != nil {
		conn.Close()
		return nil, err
	}

	keyStore, err := keys.NewSQL(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create keys store: %w", err)
	}

	corsStore, err := cors.NewSQL(conn)
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("create cors store: %w", err)
	}

	return &DB{
		Conn:      conn,
		KeysStore: keyStore,
		CORSStore: corsStore,
	}, nil
}

func (db *DB) Close() error {
	if db == nil || db.Conn == nil {
		return nil
	}
	return db.Conn.Close()
}

func configure(
	conn *sql.DB,
	useWAL bool,
) error {
	conn.SetMaxOpenConns(1)

	if _, err := conn.Exec("PRAGMA busy_timeout=5000"); err != nil {
		return fmt.Errorf("set busy_timeout: %w", err)
	}

	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		return fmt.Errorf("enable foreign keys: %w", err)
	}

	if useWAL {
		if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
			return fmt.Errorf("enable wal: %w", err)
		}
	}

	return nil
}

func runMigrations(
	conn *sql.DB,
) error {
	current, err := getSchemaVersion(conn)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if m.version <= current {
			continue
		}

		if _, err := conn.Exec(m.sql); err != nil {
			return fmt.Errorf("migration %d failed: %w", m.version, err)
		}

		if err := setSchemaVersion(conn, m.version); err != nil {
			return fmt.Errorf("set schema version %d: %w", m.version, err)
		}
	}

	return nil
}

func getSchemaVersion(
	conn *sql.DB,
) (
	int,
	error,
) {
	var version int
	if err := conn.QueryRow("PRAGMA user_version").Scan(&version); err != nil {
		return 0, fmt.Errorf("read schema version: %w", err)
	}
	return version, nil
}

func setSchemaVersion(
	conn *sql.DB,
	version int,
) error {
	_, err := conn.Exec(fmt.Sprintf("PRAGMA user_version = %d", version))
	return err
}

func (db *DB) HealthCheck() error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec("CREATE TEMP TABLE health_check (id INTEGER)"); err != nil {
		return fmt.Errorf("create temp table: %w", err)
	}

	if _, err := tx.Exec("INSERT INTO health_check (id) VALUES (1)"); err != nil {
		return fmt.Errorf("insert into temp table: %w", err)
	}

	var id int
	if err := tx.QueryRow("SELECT id FROM health_check WHERE id = 1").Scan(&id); err != nil {
		return fmt.Errorf("select from temp table: %w", err)
	}

	if _, err := tx.Exec("DROP TABLE health_check"); err != nil {
		return fmt.Errorf("drop temp table: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit health check transaction: %w", err)
	}

	return nil
}
