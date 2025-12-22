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

func (DBSignonStore) Insert(campaignID, name, email, location string, createdAt int64) (int64, error) {
	result, err := DB().Exec(
		`INSERT INTO signons (campaign_id, name, email, location, created_at) VALUES (?1, ?2, ?3, ?4, ?5)`,
		campaignID, name, email, location, createdAt,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to insert signon: %w", err)
	}
	return result.LastInsertId()
}

func (DBSignonStore) GetByID(campaignID string, id int64) (*service.Signon, error) {
	var s service.Signon
	err := DB().QueryRow(
		`SELECT id, name, email, location, created_at FROM signons WHERE campaign_id = ?1 AND id = ?2`,
		campaignID, id,
	).Scan(&s.ID, &s.Name, &s.Email, &s.Location, &s.CreatedAt)

	if err == sql.ErrNoRows {
		return nil, service.ErrSignonNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get signon: %w", err)
	}
	return &s, nil
}

func (DBSignonStore) List(campaignID string, limit, offset int) ([]*service.Signon, error) {
	rows, err := DB().Query(
		`SELECT id, name, email, location, created_at FROM signons WHERE campaign_id = ?1 ORDER BY created_at DESC LIMIT ?2 OFFSET ?3`,
		campaignID, limit, offset,
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

func (DBSignonStore) Count(campaignID string) (int, error) {
	var count int
	err := DB().QueryRow(`SELECT COUNT(*) FROM signons WHERE campaign_id = ?1`, campaignID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count signons: %w", err)
	}
	return count, nil
}

func (DBSignonStore) Delete(campaignID string, id int64) error {
	result, err := DB().Exec(`DELETE FROM signons WHERE campaign_id = ?1 AND id = ?2`, campaignID, id)
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

func (DBSignonStore) EmailExists(campaignID, email string) (bool, error) {
	var count int
	err := DB().QueryRow(`SELECT COUNT(*) FROM signons WHERE campaign_id = ?1 AND email = ?2`, campaignID, email).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check email: %w", err)
	}
	return count > 0, nil
}
