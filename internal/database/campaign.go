package database

import (
	"cosign/internal/service"
	"database/sql"
	"fmt"
)

func (db *DB) InsertCampaign(
	id string,
	name string,
	allowCustomText bool,
	createdAt int64,
) error {
	allowInt := 0
	if allowCustomText {
		allowInt = 1
	}

	_, err := db.Conn.Exec(`
		INSERT INTO campaigns (id, name, allow_custom_text, created_at)
		VALUES (?1, ?2, ?3, ?4)`,
		id,
		name,
		allowInt,
		createdAt,
	)
	if err != nil {
		return fmt.Errorf("insert campaign: %w", err)
	}
	return nil
}

func (db *DB) GetCampaign(
	id string,
) (
	*service.Campaign,
	error,
) {
	row := db.Conn.QueryRow(`
		SELECT id, name, allow_custom_text, created_at
		FROM campaigns
		WHERE id = ?1`,
		id,
	)

	var campaign service.Campaign
	var allowInt int
	if err := row.Scan(
		&campaign.ID,
		&campaign.Name,
		&allowInt,
		&campaign.CreatedAt,
	); err != nil {
		if err == sql.ErrNoRows {
			return nil, service.ErrCampaignNotFound
		}
		return nil, fmt.Errorf("get campaign: %w", err)
	}

	campaign.AllowCustomText = allowInt == 1
	return &campaign, nil
}

func (db *DB) ListCampaigns(
	limit int,
	offset int,
) (
	[]*service.Campaign,
	error,
) {
	rows, err := db.Conn.Query(`
		SELECT id, name, allow_custom_text, created_at
		FROM campaigns
		ORDER BY created_at DESC
		LIMIT ?1 OFFSET ?2`,
		limit,
		offset,
	)
	if err != nil {
		return nil, fmt.Errorf("list campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*service.Campaign
	for rows.Next() {
		var campaign service.Campaign
		var allowInt int
		if err := rows.Scan(
			&campaign.ID,
			&campaign.Name,
			&allowInt,
			&campaign.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan campaign: %w", err)
		}
		campaign.AllowCustomText = allowInt == 1
		campaigns = append(campaigns, &campaign)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate campaigns: %w", err)
	}

	return campaigns, nil
}

func (db *DB) CountCampaigns() (
	int,
	error,
) {
	row := db.Conn.QueryRow(`
		SELECT COUNT(*)
		FROM campaigns`,
	)

	var count int
	if err := row.Scan(&count); err != nil {
		return 0, fmt.Errorf("count campaigns: %w", err)
	}
	return count, nil
}

func (db *DB) UpdateCampaign(
	id string,
	name string,
	allowCustomText bool,
) error {
	allowInt := 0
	if allowCustomText {
		allowInt = 1
	}

	result, err := db.Conn.Exec(`
		UPDATE campaigns
		SET name = ?1,
			allow_custom_text = ?2
		WHERE id = ?3`,
		name,
		allowInt,
		id,
	)
	if err != nil {
		return fmt.Errorf("update campaign: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for campaign update: %w", err)
	}
	if rowsAffected == 0 {
		return service.ErrCampaignNotFound
	}

	return nil
}

func (db *DB) DeleteCampaign(
	id string,
) error {
	result, err := db.Conn.Exec(`
		DELETE FROM campaigns
		WHERE id = ?1`,
		id,
	)
	if err != nil {
		return fmt.Errorf("delete campaign: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected for campaign delete: %w", err)
	}
	if rowsAffected == 0 {
		return service.ErrCampaignNotFound
	}

	return nil
}

func (db *DB) GetCampaignLocations(
	campaignID string,
) (
	[]*service.LocationOption,
	error,
) {
	row := db.Conn.QueryRow(`
		SELECT COUNT(*)
		FROM campaigns
		WHERE id = ?1`,
		campaignID,
	)

	var exists int
	if err := row.Scan(&exists); err != nil {
		return nil, fmt.Errorf("verify campaign exists: %w", err)
	}
	if exists == 0 {
		return nil, service.ErrCampaignNotFound
	}

	rows, err := db.Conn.Query(`
		SELECT id, value, display_order
		FROM locations
		WHERE campaign_id = ?1
		ORDER BY display_order ASC`,
		campaignID,
	)
	if err != nil {
		return nil, fmt.Errorf("get campaign locations: %w", err)
	}
	defer rows.Close()

	var options []*service.LocationOption
	for rows.Next() {
		var opt service.LocationOption
		if err := rows.Scan(
			&opt.ID,
			&opt.Value,
			&opt.DisplayOrder,
		); err != nil {
			return nil, fmt.Errorf("scan campaign location: %w", err)
		}
		options = append(options, &opt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate campaign locations: %w", err)
	}

	return options, nil
}

func (db *DB) ReplaceCampaignLocations(
	campaignID string,
	options []service.LocationOption,
) error {
	tx, err := db.Conn.Begin()
	if err != nil {
		return fmt.Errorf("begin replace locations transaction: %w", err)
	}
	defer tx.Rollback()

	row := tx.QueryRow(`
		SELECT COUNT(*)
		FROM campaigns
		WHERE id = ?1`,
		campaignID,
	)

	var exists int
	if err := row.Scan(&exists); err != nil {
		return fmt.Errorf("verify campaign exists: %w", err)
	}
	if exists == 0 {
		return service.ErrCampaignNotFound
	}

	if _, err := tx.Exec(`
		DELETE FROM locations
		WHERE campaign_id = ?1`,
		campaignID,
	); err != nil {
		return fmt.Errorf("clear campaign locations: %w", err)
	}

	for _, opt := range options {
		if _, err := tx.Exec(`
			INSERT INTO locations (campaign_id, value, display_order)
			VALUES (?1, ?2, ?3)`,
			campaignID,
			opt.Value,
			opt.DisplayOrder,
		); err != nil {
			return fmt.Errorf("insert campaign location: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit replace locations: %w", err)
	}

	return nil
}
