package database

import (
	"cosign/internal/service"
	"database/sql"
	"fmt"
)

func (db *DB) InsertSignon(
	campaignID string,
	name string,
	email string,
	location string,
	createdAt int64,
) (int64, error) {
	result, err := db.Conn.Exec(`
		INSERT INTO signons (campaign_id, name, email, location, created_at)
		VALUES (?1, ?2, ?3, ?4, ?5)`,
		campaignID,
		name,
		email,
		location,
		createdAt,
	)
	if err != nil {
		return 0, fmt.Errorf("insert signon: %w", err)
	}
	return result.LastInsertId()
}

func (db *DB) ListSignons(
	campaignID string,
	limit int,
	offset int,
) ([]*service.Signon, error) {
	rows, err := db.Conn.Query(`
		SELECT id, name, email, location, created_at
		FROM signons
		WHERE campaign_id = ?1
		ORDER BY created_at DESC
		LIMIT ?2 OFFSET ?3`,
		campaignID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list signons: %w", err)
	}
	defer rows.Close()

	var signons []*service.Signon
	for rows.Next() {
		var s service.Signon
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Email,
			&s.Location,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan signon: %w", err)
		}
		signons = append(signons, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signons: %w", err)
	}

	return signons, nil
}

func (db *DB) CountSignons(
	campaignID string,
) (
	int,
	error,
) {
	row := db.Conn.QueryRow(`
		SELECT COUNT(*)
		FROM signons
		WHERE campaign_id = ?1`,
		campaignID,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count signons: %w", err)
	}
	return count, nil
}

func (db *DB) DeleteSignon(
	campaignID string,
	id int64,
) error {
	result, err := db.Conn.Exec(`
		DELETE FROM signons
		WHERE campaign_id = ?1 AND id = ?2`,
		campaignID,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete signon: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for signon delete: %w", err)
	}

	if rows == 0 {
		return service.ErrSignonNotFound
	}

	return nil
}

func (db *DB) SignonEmailExists(
	campaignID string,
	email string,
) (
	bool,
	error,
) {
	row := db.Conn.QueryRow(`
		SELECT COUNT(*)
		FROM signons
		WHERE campaign_id = ?1 AND email = ?2`,
		campaignID,
		email,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check signon email: %w", err)
	}
	return count > 0, nil
}

func (db *DB) GetSignon(
	campaignID string,
	id int64,
) (
	*service.Signon,
	error,
) {
	row := db.Conn.QueryRow(`
		SELECT id, name, email, location, created_at
		FROM signons
		WHERE campaign_id = ?1 AND id = ?2`,
		campaignID,
		id,
	)

	var s service.Signon
	if err := row.Scan(
		&s.ID,
		&s.Name,
		&s.Email,
		&s.Location,
		&s.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrSignonNotFound
		}
		return nil, fmt.Errorf("get signon: %w", err)
	}

	return &s, nil
}
