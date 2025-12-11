package database

import (
	"cosign/internal/service"
	"fmt"
)

// DBLocationOptionStore implements service.LocationOptionStore
type DBLocationOptionStore struct{}

// NewLocationOptionStore returns a new LocationOptionStore implementation
func NewLocationOptionStore() DBLocationOptionStore {
	return DBLocationOptionStore{}
}

func (DBLocationOptionStore) GetOptions(campaignID string) ([]*service.LocationOption, error) {
	rows, err := DB().Query(
		`SELECT id, value, display_order FROM location_options WHERE campaign_id = ?1 ORDER BY display_order ASC`,
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

func (DBLocationOptionStore) AddOption(campaignID, value string, displayOrder int) (int64, error) {
	result, err := DB().Exec(
		`INSERT INTO location_options (campaign_id, value, display_order) VALUES (?1, ?2, ?3)`,
		campaignID, value, displayOrder,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to add location option: %w", err)
	}
	return result.LastInsertId()
}

func (DBLocationOptionStore) UpdateOption(campaignID string, id int64, value string, displayOrder int) error {
	result, err := DB().Exec(
		`UPDATE location_options SET value = ?1, display_order = ?2 WHERE campaign_id = ?3 AND id = ?4`,
		value, displayOrder, campaignID, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update location option: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return service.ErrLocationOptionNotFound
	}

	return nil
}

func (DBLocationOptionStore) DeleteOption(campaignID string, id int64) error {
	result, err := DB().Exec(`DELETE FROM location_options WHERE campaign_id = ?1 AND id = ?2`, campaignID, id)
	if err != nil {
		return fmt.Errorf("failed to delete location option: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rows == 0 {
		return service.ErrLocationOptionNotFound
	}

	return nil
}
