package query

import (
	"database/sql"
	"dbaTools/internal/utils"
	"fmt"
)

// GetServicesByServerID retrieves all services for a given server
func GetServicesByServerID(db *sql.DB, serverID int) ([]utils.ServiceInfo, error) {
	rows, err := db.Query(`
		SELECT id, server_id, service_name, service_type, status,
		       port, version, install_path, config_file, 
		       auto_start, last_checked, last_status_change,
		       created_at, updated_at, notes
		FROM server_services
		WHERE server_id = ?
		ORDER BY service_name
	`, serverID)
	if err != nil {
		return nil, fmt.Errorf("failed to query services: %w", err)
	}
	defer rows.Close()

	var services []utils.ServiceInfo
	for rows.Next() {
		var service utils.ServiceInfo

		err := rows.Scan(
			&service.ID,
			&service.ServerID,
			&service.ServiceName,
			&service.ServiceType,
			&service.Status,
			&service.Port,
			&service.Version,
			&service.InstallPath,
			&service.ConfigFile,
			&service.AutoStart,
			&service.LastChecked,
			&service.LastStatusChange,
			&service.CreatedAt,
			&service.UpdatedAt,
			&service.Notes,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan service row: %w", err)
		}

		services = append(services, service)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return services, nil
}

// GetServiceByID retrieves a specific service by its ID
func GetServiceByID(db *sql.DB, serviceID int) (*utils.ServiceInfo, error) {
	var service utils.ServiceInfo

	err := db.QueryRow(`
		SELECT id, server_id, service_name, service_type, status,
		       port, version, install_path, config_file, 
		       auto_start, last_checked, last_status_change,
		       created_at, updated_at, notes
		FROM server_services
		WHERE id = ?
	`, serviceID).Scan(
		&service.ID,
		&service.ServerID,
		&service.ServiceName,
		&service.ServiceType,
		&service.Status,
		&service.Port,
		&service.Version,
		&service.InstallPath,
		&service.ConfigFile,
		&service.AutoStart,
		&service.LastChecked,
		&service.LastStatusChange,
		&service.CreatedAt,
		&service.UpdatedAt,
		&service.Notes,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No service found with the given ID
	} else if err != nil {
		return nil, fmt.Errorf("failed to query service: %w", err)
	}

	return &service, nil
}
