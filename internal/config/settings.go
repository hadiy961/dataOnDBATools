// File: internal/config/settings.go
package config

import (
	"database/sql"
	"fmt"
)

func loadSettings(db *sql.DB) (map[string]interface{}, error) {
	query := `SELECT setting_key, setting_value, setting_type FROM apps_settings`
	rows, err := db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query settings: %w", err)
	}
	defer rows.Close()

	settings := make(map[string]interface{})
	for rows.Next() {
		var key, value, settingType string
		if err := rows.Scan(&key, &value, &settingType); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}

		switch settingType {
		case "boolean":
			settings[key] = value == "true"
		case "string":
			settings[key] = value
		default:
			settings[key] = value
		}
	}

	return settings, nil
}
