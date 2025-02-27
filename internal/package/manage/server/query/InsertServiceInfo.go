package query

import (
	"database/sql"
	"dbaTools/internal/utils"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

// InsertServiceInfo inserts or updates service information in the database
func InsertServiceInfo(db *sql.DB, services []utils.ServiceInfo) error {
	if len(services) == 0 {
		return fmt.Errorf("no services to insert")
	}

	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Service DB Update ] "
	spin.Suffix = fmt.Sprintf(" Analyzing service information for server_id %d...", services[0].ServerID)
	spin.Color("green")
	spin.Start()
	time.Sleep(101 * time.Millisecond)
	defer spin.Stop()

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		spin.Stop()
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback in case of error
	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// First, delete any existing records that aren't in the current set
	// This prevents orphaned records if service detection changes
	serviceNames := make([]interface{}, len(services))
	for i, service := range services {
		serviceNames[i] = service.ServiceName
	}

	// Build placeholders for IN clause
	placeholders := make([]string, len(services))
	for i := range services {
		placeholders[i] = "?"
	}

	// Only proceed with deletion if we have services to check against
	if len(services) > 0 {
		deleteQuery := fmt.Sprintf(`
			DELETE FROM server_services 
			WHERE server_id = ? AND service_name NOT IN (%s)
		`, strings.Join(placeholders, ","))

		// Create args array with server_id as first parameter
		deleteArgs := make([]interface{}, len(serviceNames)+1)
		deleteArgs[0] = services[0].ServerID
		for i, name := range serviceNames {
			deleteArgs[i+1] = name
		}

		// Execute the delete
		_, err = tx.Exec(deleteQuery, deleteArgs...)
		if err != nil {
			spin.Stop()
			return fmt.Errorf("failed to clean up old service records: %w", err)
		}
	}

	// Insert or update each service
	var updated, inserted int
	for _, service := range services {
		// Check if service record already exists and get current values
		var exists bool
		var existingID int
		var existingService utils.ServiceInfo

		checkQuery := `
			SELECT id, status, port, version, install_path, config_file, auto_start, 
                   IFNULL(last_status_change, ''), IFNULL(last_checked, ''),
                   IFNULL(created_at, ''), IFNULL(updated_at, '')
			FROM server_services 
			WHERE server_id = ? AND service_name = ?
		`

		var lastStatusChangeStr, lastCheckedStr, createdAtStr, updatedAtStr string

		err = db.QueryRow(checkQuery, service.ServerID, service.ServiceName).Scan(
			&existingID,
			&existingService.Status,
			&existingService.Port,
			&existingService.Version,
			&existingService.InstallPath,
			&existingService.ConfigFile,
			&existingService.AutoStart,
			&lastStatusChangeStr,
			&lastCheckedStr,
			&createdAtStr,
			&updatedAtStr,
		)

		exists = (err == nil)

		// If scanning resulted in an error other than "not found", return it
		if err != nil && err != sql.ErrNoRows {
			spin.Stop()
			return fmt.Errorf("failed to check if service exists: %w", err)
		}

		// Parse existing timestamps
		formats := []string{
			"2006-01-02 15:04:05",
			time.RFC3339,
			"2006-01-02T15:04:05Z",
		}

		if lastStatusChangeStr != "" {
			for _, format := range formats {
				if t, err := time.Parse(format, lastStatusChangeStr); err == nil {
					existingService.LastStatusChange = t
					break
				}
			}
		}

		if lastCheckedStr != "" {
			for _, format := range formats {
				if t, err := time.Parse(format, lastCheckedStr); err == nil {
					existingService.LastChecked = t
					break
				}
			}
		}

		if createdAtStr != "" {
			for _, format := range formats {
				if t, err := time.Parse(format, createdAtStr); err == nil {
					existingService.CreatedAt = t
					break
				}
			}
		}

		if updatedAtStr != "" {
			for _, format := range formats {
				if t, err := time.Parse(format, updatedAtStr); err == nil {
					existingService.UpdatedAt = t
					break
				}
			}
		}

		if exists {
			// Check if any data has changed
			hasChanges := service.Status != existingService.Status ||
				service.Port != existingService.Port ||
				service.Version != existingService.Version ||
				service.InstallPath != existingService.InstallPath ||
				service.ConfigFile != existingService.ConfigFile ||
				service.AutoStart != existingService.AutoStart

			if !hasChanges {
				// Only update last_checked if no other changes
				_, err = tx.Exec(`
					UPDATE server_services
					SET last_checked = ?
					WHERE id = ?
				`, service.LastChecked, existingID)

				if err != nil {
					spin.Stop()
					return fmt.Errorf("failed to update last_checked: %w", err)
				}

				spin.Suffix = fmt.Sprintf(" Service unchanged: %s", service.ServiceName)
				continue
			}

			// Set last_status_change if status has changed
			lastStatusChange := existingService.LastStatusChange
			if service.Status != existingService.Status {
				lastStatusChange = time.Now()
			}

			// Update the record with changes
			_, err = tx.Exec(`
				UPDATE server_services
				SET status = ?,
					port = ?,
					version = ?,
					install_path = ?,
					config_file = ?,
					auto_start = ?,
					last_checked = ?,
					last_status_change = ?,
					updated_at = ?
				WHERE id = ?
			`,
				service.Status,
				service.Port,
				service.Version,
				service.InstallPath,
				service.ConfigFile,
				service.AutoStart,
				service.LastChecked,
				lastStatusChange,
				time.Now(),
				existingID,
			)

			if err != nil {
				spin.Stop()
				return fmt.Errorf("failed to update service: %w", err)
			}

			updated++
			spin.Suffix = fmt.Sprintf(" Updated service: %s (%s)...", service.ServiceName, service.Status)
		} else {
			// Insert new record
			_, err = tx.Exec(`
				INSERT INTO server_services (
					server_id, service_name, service_type, status, 
					port, version, install_path, config_file,
					auto_start, last_checked, last_status_change, 
					created_at, updated_at, notes
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`,
				service.ServerID,
				service.ServiceName,
				service.ServiceType,
				service.Status,
				service.Port,
				service.Version,
				service.InstallPath,
				service.ConfigFile,
				service.AutoStart,
				service.LastChecked,
				time.Now(), // new service, set last_status_change to now
				time.Now(), // created_at
				time.Now(), // updated_at
				service.Notes,
			)

			if err != nil {
				spin.Stop()
				return fmt.Errorf("failed to insert service: %w", err)
			}

			inserted++
			spin.Suffix = fmt.Sprintf(" Added new service: %s (%s)...", service.ServiceName, service.Status)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		spin.Stop()
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	if updated == 0 && inserted == 0 {
		spin.Suffix = " All services are up to date. No changes needed."
	} else {
		spin.Suffix = fmt.Sprintf(" Service info updated: %d added, %d modified", inserted, updated)
	}
	time.Sleep(101 * time.Millisecond)

	return nil
}
