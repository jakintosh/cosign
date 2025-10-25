package database

import (
	"cosign/internal/service"
	"fmt"
)

// DBCORSStore implements service.CORSStore
type DBCORSStore struct{}

// NewCORSStore returns a new CORSStore implementation
func NewCORSStore() DBCORSStore {
	return DBCORSStore{}
}

func (DBCORSStore) Add(origin string, createdAt int64) error {
	_, err := db.Exec(
		`INSERT INTO cors_origins (origin, created_at) VALUES (?1, ?2)
		 ON CONFLICT(origin) DO UPDATE SET created_at = ?2`,
		origin, createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to add cors origin: %w", err)
	}
	return nil
}

func (DBCORSStore) List() ([]string, error) {
	rows, err := db.Query(`SELECT origin FROM cors_origins ORDER BY origin ASC`)
	if err != nil {
		return nil, fmt.Errorf("failed to list cors origins: %w", err)
	}
	defer rows.Close()

	var origins []string
	for rows.Next() {
		var origin string
		if err := rows.Scan(&origin); err != nil {
			return nil, fmt.Errorf("failed to scan cors origin: %w", err)
		}
		origins = append(origins, origin)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return origins, nil
}

func (DBCORSStore) Delete(origin string) error {
	result, err := db.Exec(`DELETE FROM cors_origins WHERE origin = ?1`, origin)
	if err != nil {
		return fmt.Errorf("failed to delete cors origin: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return service.ErrCORSOriginNotFound
	}

	return nil
}

func (DBCORSStore) IsAllowed(origin string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM cors_origins WHERE origin = ?1`, origin).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check cors origin: %w", err)
	}
	return count > 0, nil
}
