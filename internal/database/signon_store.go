package database

import (
	"cosign/internal/service"
	"database/sql"
	"fmt"
)

// DBSignonStore implements service.SignonStore
type DBSignonStore struct{}

// NewSignonStore returns a new SignonStore implementation
func NewSignonStore() DBSignonStore {
	return DBSignonStore{}
}

func (DBSignonStore) Insert(name, email, location string, createdAt int64) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO signons (name, email, location, created_at) VALUES (?1, ?2, ?3, ?4)`,
		name, email, location, createdAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert signon: %w", err)
	}
	return result.LastInsertId()
}

func (DBSignonStore) GetByID(id int64) (*service.Signon, error) {
	var s service.Signon
	err := db.QueryRow(
		`SELECT id, name, email, location, created_at FROM signons WHERE id = ?1`,
		id,
	).Scan(&s.ID, &s.Name, &s.Email, &s.Location, &s.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, service.ErrSignonNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get signon: %w", err)
	}
	return &s, nil
}

func (DBSignonStore) List(limit, offset int) ([]*service.Signon, error) {
	rows, err := db.Query(
		`SELECT id, name, email, location, created_at FROM signons ORDER BY created_at DESC LIMIT ?1 OFFSET ?2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to list signons: %w", err)
	}
	defer rows.Close()

	var signons []*service.Signon
	for rows.Next() {
		var s service.Signon
		if err := rows.Scan(&s.ID, &s.Name, &s.Email, &s.Location, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan signon: %w", err)
		}
		signons = append(signons, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return signons, nil
}

func (DBSignonStore) Count() (int, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM signons`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count signons: %w", err)
	}
	return count, nil
}

func (DBSignonStore) Delete(id int64) error {
	result, err := db.Exec(`DELETE FROM signons WHERE id = ?1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete signon: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return service.ErrSignonNotFound
	}

	return nil
}

func (DBSignonStore) EmailExists(email string) (bool, error) {
	var count int
	err := db.QueryRow(`SELECT COUNT(*) FROM signons WHERE email = ?1`, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email: %w", err)
	}
	return count > 0, nil
}
