package database

import (
	"cosign/internal/service"
	"database/sql"
	"fmt"
)

// DBLocationConfigStore implements service.LocationConfigStore
type DBLocationConfigStore struct{}

// NewLocationConfigStore returns a new LocationConfigStore implementation
func NewLocationConfigStore() DBLocationConfigStore {
	return DBLocationConfigStore{}
}

func (DBLocationConfigStore) GetConfig() (*service.LocationConfig, error) {
	var cfg service.LocationConfig
	var allowCustomInt int
	err := db.QueryRow(
		`SELECT allow_custom_text FROM location_config WHERE id = 1`,
	).Scan(&allowCustomInt)

	if err == sql.ErrNoRows {
		return nil, service.ErrLocationConfigNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get location config: %w", err)
	}

	cfg.AllowCustomText = allowCustomInt != 0
	return &cfg, nil
}

func (DBLocationConfigStore) SetAllowCustomText(allow bool) error {
	allowInt := 0
	if allow {
		allowInt = 1
	}

	_, err := db.Exec(
		`UPDATE location_config SET allow_custom_text = ?1 WHERE id = 1`,
		allowInt,
	)
	if err != nil {
		return fmt.Errorf("failed to update location config: %w", err)
	}
	return nil
}

func (DBLocationConfigStore) GetOptions() ([]*service.LocationOption, error) {
	rows, err := db.Query(
		`SELECT id, value, display_order FROM location_options ORDER BY display_order ASC`,
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

func (DBLocationConfigStore) AddOption(value string, displayOrder int) (int64, error) {
	result, err := db.Exec(
		`INSERT INTO location_options (value, display_order) VALUES (?1, ?2)`,
		value, displayOrder,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to add location option: %w", err)
	}
	return result.LastInsertId()
}

func (DBLocationConfigStore) UpdateOption(id int64, value string, displayOrder int) error {
	result, err := db.Exec(
		`UPDATE location_options SET value = ?1, display_order = ?2 WHERE id = ?3`,
		value, displayOrder, id,
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

func (DBLocationConfigStore) DeleteOption(id int64) error {
	result, err := db.Exec(`DELETE FROM location_options WHERE id = ?1`, id)
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
