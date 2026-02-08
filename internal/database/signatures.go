package database

import (
	"cosign/internal/service"
	"database/sql"
	"fmt"
)

func (db *DB) InsertSignature(
	campaignID string,
	name string,
	email string,
	location string,
	createdAt int64,
) (int64, error) {
	result, err := db.Conn.Exec(`
		INSERT INTO signatures (campaign_id, name, email, location, created_at)
		VALUES (?1, ?2, ?3, ?4, ?5)`,
		campaignID,
		name,
		email,
		location,
		createdAt,
	)
	if err != nil {
		return 0, fmt.Errorf("insert signature: %w", err)
	}
	return result.LastInsertId()
}

func (db *DB) ListSignatures(
	campaignID string,
	limit int,
	offset int,
) ([]*service.Signature, error) {
	rows, err := db.Conn.Query(`
		SELECT id, name, email, location, created_at
		FROM signatures
		WHERE campaign_id = ?1
		ORDER BY created_at DESC
		LIMIT ?2 OFFSET ?3`,
		campaignID,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list signatures: %w", err)
	}
	defer rows.Close()

	var signatures []*service.Signature
	for rows.Next() {
		var s service.Signature
		if err := rows.Scan(
			&s.ID,
			&s.Name,
			&s.Email,
			&s.Location,
			&s.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan signature: %w", err)
		}
		signatures = append(signatures, &s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate signatures: %w", err)
	}

	return signatures, nil
}

func (db *DB) CountSignatures(
	campaignID string,
) (
	int,
	error,
) {
	row := db.Conn.QueryRow(`
		SELECT COUNT(*)
		FROM signatures
		WHERE campaign_id = ?1`,
		campaignID,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count signatures: %w", err)
	}
	return count, nil
}

func (db *DB) DeleteSignature(
	campaignID string,
	id int64,
) error {
	result, err := db.Conn.Exec(`
		DELETE FROM signatures
		WHERE campaign_id = ?1 AND id = ?2`,
		campaignID,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete signature: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for signature delete: %w", err)
	}

	if rows == 0 {
		return service.ErrSignatureNotFound
	}

	return nil
}

func (db *DB) SignatureEmailExists(
	campaignID string,
	email string,
) (
	bool,
	error,
) {
	row := db.Conn.QueryRow(`
		SELECT COUNT(*)
		FROM signatures
		WHERE campaign_id = ?1 AND email = ?2`,
		campaignID,
		email,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return false, fmt.Errorf("check signature email: %w", err)
	}
	return count > 0, nil
}

func (db *DB) GetSignature(
	campaignID string,
	id int64,
) (
	*service.Signature,
	error,
) {
	row := db.Conn.QueryRow(`
		SELECT id, name, email, location, created_at
		FROM signatures
		WHERE campaign_id = ?1 AND id = ?2`,
		campaignID,
		id,
	)

	var s service.Signature
	if err := row.Scan(
		&s.ID,
		&s.Name,
		&s.Email,
		&s.Location,
		&s.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrSignatureNotFound
		}
		return nil, fmt.Errorf("get signature: %w", err)
	}

	return &s, nil
}
