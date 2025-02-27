package query

import (
	"database/sql"
	"dbaTools/internal/utils"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// InsertServerDetail inserts or updates server system information into the database
func InsertServerDetail(db *sql.DB, serverID int, sysInfo *utils.SystemInfo) error {
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Database Update ] "
	spin.Color("green")
	spin.Suffix = fmt.Sprintf(" Saving system information for server_id %d...", serverID)
	spin.Start()
	time.Sleep(101 * time.Millisecond)
	defer spin.Stop()

	// Current timestamp for created_at/updated_at/last_check
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Begin transaction to ensure atomicity
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback in case of error - it will be a no-op if the transaction is committed
	defer tx.Rollback()

	// Check if record exists and get current values to determine if update is needed
	existingQuery := `
		SELECT 
			hostname, total_ram, cpu_core, cpu_model, os_type, os_version
		FROM server_detail 
		WHERE server_id = ? OR hostname = ?
	`

	var (
		existingHostname, existingTotalRAM, existingCPUCore string
		existingCPUModel, existingOSType, existingOSVersion string
		recordExists, updateNeeded                          bool
	)

	err = tx.QueryRow(existingQuery, serverID, sysInfo.Hostname).Scan(
		&existingHostname, &existingTotalRAM, &existingCPUCore,
		&existingCPUModel, &existingOSType, &existingOSVersion,
	)

	if err == nil {
		// Record exists, check if there are changes
		recordExists = true
		updateNeeded = existingHostname != sysInfo.Hostname ||
			existingTotalRAM != sysInfo.TotalRAM ||
			existingCPUCore != sysInfo.CPUCore ||
			existingCPUModel != sysInfo.CPUModel ||
			existingOSType != sysInfo.OSType ||
			existingOSVersion != sysInfo.OSVersion
	} else if err == sql.ErrNoRows {
		// Record doesn't exist
		recordExists = false
		updateNeeded = true // New record, so update is needed
	} else {
		// An actual error occurred
		return fmt.Errorf("error checking existing record: %w", err)
	}

	if recordExists {
		// Check if there's a conflict between server_id and hostname
		conflictQuery := `
			SELECT server_id 
			FROM server_detail 
			WHERE hostname = ? AND server_id != ?
		`
		var conflictID int
		err = tx.QueryRow(conflictQuery, sysInfo.Hostname, serverID).Scan(&conflictID)

		if err != nil && err != sql.ErrNoRows {
			return fmt.Errorf("error checking for conflicts: %w", err)
		}

		if err == nil {
			// We found a record with the same hostname but different server_id
			return fmt.Errorf("hostname '%s' already assigned to server_id %d", sysInfo.Hostname, conflictID)
		}

		if updateNeeded {
			// Data has changed, update all fields including last_update
			query := `
				UPDATE server_detail
				SET 
					hostname = ?,
					total_ram = ?,
					cpu_core = ?,
					cpu_model = ?,
					os_type = ?,
					os_version = ?,
					updated_at = ?,
					updated_by = 'DBA',
					last_check = ?,
					last_update = ?
				WHERE server_id = ? OR hostname = ?
			`
			_, err = tx.Exec(query,
				sysInfo.Hostname,
				sysInfo.TotalRAM,
				sysInfo.CPUCore,
				sysInfo.CPUModel,
				sysInfo.OSType,
				sysInfo.OSVersion,
				currentTime,
				currentTime,
				currentTime, // Update last_update since data changed
				serverID,
				existingHostname,
			)
		} else {
			// No changes, only update last_check field
			query := `
				UPDATE server_detail
				SET 
					last_check = ?
				WHERE server_id = ? OR hostname = ?
			`
			_, err = tx.Exec(query,
				currentTime,
				serverID,
				existingHostname,
			)
		}
	} else {
		// Insert new record
		query := `
			INSERT INTO server_detail (
				id, server_id, hostname, total_ram, cpu_core, cpu_model, 
				os_type, os_version, created_at, created_by, updated_at, updated_by,
				last_check, last_update
			) VALUES (null, ?, ?, ?, ?, ?, ?, ?, ?, 'DBA', ?, 'DBA', ?, ?)
		`
		_, err = tx.Exec(query,
			serverID,
			sysInfo.Hostname,
			sysInfo.TotalRAM,
			sysInfo.CPUCore,
			sysInfo.CPUModel,
			sysInfo.OSType,
			sysInfo.OSVersion,
			currentTime,
			currentTime,
			currentTime, // Initial last_check
			currentTime, // Initial last_update
		)
	}

	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	// Commit the transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Add a small pause to ensure spinner is visible
	time.Sleep(101 * time.Millisecond)

	return nil
}
