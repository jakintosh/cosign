package database

import (
	"cosign/internal/service"
	"fmt"
)

// DBCampaignStore implements CampaignStore interface
type DBCampaignStore struct{}

// NewCampaignStore creates a new campaign store
func NewCampaignStore() DBCampaignStore {
	return DBCampaignStore{}
}

// Insert creates a new campaign
func (DBCampaignStore) Insert(id, name string, allowCustomText bool, createdAt int64) error {
	allowInt := 0
	if allowCustomText {
		allowInt = 1
	}

	_, err := DB().Exec(
		`INSERT INTO campaigns (id, name, allow_custom_text, created_at) VALUES (?1, ?2, ?3, ?4)`,
		id, name, allowInt, createdAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert campaign: %w", err)
	}
	return nil
}

// GetByID retrieves a campaign by ID
func (DBCampaignStore) GetByID(id string) (*service.Campaign, error) {
	row := DB().QueryRow(
		`SELECT id, name, allow_custom_text, created_at FROM campaigns WHERE id = ?1`,
		id,
	)

	var campaign service.Campaign
	var allowInt int
	err := row.Scan(&campaign.ID, &campaign.Name, &allowInt, &campaign.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to get campaign: %w", err)
	}

	campaign.AllowCustomText = allowInt == 1
	return &campaign, nil
}

// List retrieves campaigns with pagination
func (DBCampaignStore) List(limit, offset int) ([]*service.Campaign, error) {
	rows, err := DB().Query(
		`SELECT id, name, allow_custom_text, created_at FROM campaigns ORDER BY created_at DESC LIMIT ?1 OFFSET ?2`,
		limit, offset,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query campaigns: %w", err)
	}
	defer rows.Close()

	var campaigns []*service.Campaign
	for rows.Next() {
		var campaign service.Campaign
		var allowInt int
		if err := rows.Scan(&campaign.ID, &campaign.Name, &allowInt, &campaign.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan campaign: %w", err)
		}
		campaign.AllowCustomText = allowInt == 1
		campaigns = append(campaigns, &campaign)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating campaigns: %w", err)
	}

	return campaigns, nil
}

// Count returns the total number of campaigns
func (DBCampaignStore) Count() (int, error) {
	var count int
	err := DB().QueryRow(`SELECT COUNT(*) FROM campaigns`).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count campaigns: %w", err)
	}
	return count, nil
}

// Update modifies a campaign
func (DBCampaignStore) Update(id, name string, allowCustomText bool) error {
	allowInt := 0
	if allowCustomText {
		allowInt = 1
	}

	result, err := DB().Exec(
		`UPDATE campaigns SET name = ?1, allow_custom_text = ?2 WHERE id = ?3`,
		name, allowInt, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update campaign: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return service.ErrCampaignNotFound
	}

	return nil
}

// Delete removes a campaign
func (DBCampaignStore) Delete(id string) error {
	result, err := DB().Exec(`DELETE FROM campaigns WHERE id = ?1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete campaign: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return service.ErrCampaignNotFound
	}

	return nil
}

// GetLocationOptions returns location options for a campaign.
func (DBCampaignStore) GetLocationOptions(campaignID string) ([]*service.LocationOption, error) {
	var exists int
	if err := DB().QueryRow(`SELECT COUNT(*) FROM campaigns WHERE id = ?1`, campaignID).Scan(&exists); err != nil {
		return nil, fmt.Errorf("failed to verify campaign: %w", err)
	}
	if exists == 0 {
		return nil, service.ErrCampaignNotFound
	}

	rows, err := DB().Query(
		`SELECT id, value, display_order FROM locations WHERE campaign_id = ?1 ORDER BY display_order ASC`,
		campaignID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get location options: %w", err)
	}
	defer rows.Close()

	var options []*service.LocationOption
	for rows.Next() {
		var opt service.LocationOption
		if err := rows.Scan(&opt.ID, &opt.Value, &opt.DisplayOrder); err != nil {
			return nil, fmt.Errorf("failed to scan location option: %w", err)
		}
		options = append(options, &opt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return options, nil
}

// ReplaceLocationOptions replaces location options for a campaign transactionally.
func (DBCampaignStore) ReplaceLocationOptions(campaignID string, options []service.LocationOption) error {
	tx, err := DB().Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var exists int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM campaigns WHERE id = ?1`, campaignID).Scan(&exists); err != nil {
		return fmt.Errorf("failed to verify campaign: %w", err)
	}
	if exists == 0 {
		return service.ErrCampaignNotFound
	}

	if _, err := tx.Exec(`DELETE FROM locations WHERE campaign_id = ?1`, campaignID); err != nil {
		return fmt.Errorf("failed to clear location options: %w", err)
	}

	for _, opt := range options {
		if _, err := tx.Exec(
			`INSERT INTO locations (campaign_id, value, display_order) VALUES (?1, ?2, ?3)`,
			campaignID, opt.Value, opt.DisplayOrder,
		); err != nil {
			return fmt.Errorf("failed to insert location option: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit config update: %w", err)
	}

	return nil
}
