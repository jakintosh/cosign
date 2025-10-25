package database

import (
	"cosign/internal/service"
	"database/sql"
	"fmt"
)

// DBKeyStore implements service.KeyStore
type DBKeyStore struct{}

// NewKeyStore returns a new KeyStore implementation
func NewKeyStore() DBKeyStore {
	return DBKeyStore{}
}

func (DBKeyStore) Insert(id string, hash, salt []byte, createdAt int64) error {
	_, err := db.Exec(
		`INSERT INTO api_keys (id, hash, salt, created_at) VALUES (?1, ?2, ?3, ?4)`,
		id, hash, salt, createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert api key: %w", err)
	}
	return nil
}

func (DBKeyStore) GetByID(id string) (*service.APIKey, error) {
	var key service.APIKey
	err := db.QueryRow(
		`SELECT id, hash, salt, created_at FROM api_keys WHERE id = ?1`,
		id,
	).Scan(&key.ID, &key.Hash, &key.Salt, &key.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, service.ErrAPIKeyNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get api key: %w", err)
	}
	return &key, nil
}

func (DBKeyStore) List() ([]*service.APIKey, error) {
	rows, err := db.Query(
		`SELECT id, hash, salt, created_at FROM api_keys ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list api keys: %w", err)
	}
	defer rows.Close()

	var keys []*service.APIKey
	for rows.Next() {
		var key service.APIKey
		if err := rows.Scan(&key.ID, &key.Hash, &key.Salt, &key.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan api key: %w", err)
		}
		keys = append(keys, &key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return keys, nil
}

func (DBKeyStore) Delete(id string) error {
	result, err := db.Exec(`DELETE FROM api_keys WHERE id = ?1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete api key: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return service.ErrAPIKeyNotFound
	}

	return nil
}
