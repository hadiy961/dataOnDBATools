package query

import (
	"database/sql"
	"dbaTools/internal/utils"
	"fmt"
	"strings"
	"time"

	"github.com/briandowns/spinner"
)

// InsertServerStorage inserts or updates server storage information in the database
func InsertServerStorage(db *sql.DB, serverID int, storageList []utils.StorageInfo) error {
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Storage DB Update ] "
	spin.Suffix = fmt.Sprintf(" Saving Storage information for server_id %d...", serverID)

	spin.Color("green")
	spin.Start()
	time.Sleep(101 * time.Millisecond)
	defer spin.Stop()

	// Current timestamp for created_at/updated_at
	currentTime := time.Now().Format("2006-01-02 15:04:05")

	// Begin transaction
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Defer rollback in case of error
	defer tx.Rollback()

	// Get list of current devices and their data for this server
	existingDevicesData := make(map[string]utils.StorageInfo)
	currentDevicesQuery := `
		SELECT device_name, mount_point, filesystem_type, total_space, used_space, free_space, use_percent 
		FROM server_storage 
		WHERE server_id = ?
	`
	rows, err := db.Query(currentDevicesQuery, serverID)
	if err != nil {
		return fmt.Errorf("failed to query existing devices: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var info utils.StorageInfo
		if err := rows.Scan(
			&info.DeviceName,
			&info.MountPoint,
			&info.FilesystemType,
			&info.TotalSpace,
			&info.UsedSpace,
			&info.FreeSpace,
			&info.UsePercent,
		); err != nil {
			return fmt.Errorf("failed to scan device data: %w", err)
		}
		// Use combined key of device name and mount point
		existingDevicesData[info.DeviceName+"|"+info.MountPoint] = info
	}

	// Keep track of devices we're updating
	updatedDevices := make(map[string]bool)

	// Insert/update storage records
	for _, storage := range storageList {
		deviceKey := storage.DeviceName + "|" + storage.MountPoint

		// Check if this device exists and if data has changed
		existingInfo, exists := existingDevicesData[deviceKey]
		updateNeeded := true // Default to true for new devices

		if exists {
			// Compare fields to see if actual changes are needed
			updateNeeded =
				existingInfo.FilesystemType != storage.FilesystemType ||
					existingInfo.TotalSpace != storage.TotalSpace ||
					existingInfo.UsedSpace != storage.UsedSpace ||
					existingInfo.FreeSpace != storage.FreeSpace ||
					existingInfo.UsePercent != storage.UsePercent
		}

		if exists {
			// Device exists, update with different queries based on if data changed
			if updateNeeded {
				// Data changed, update all fields including last_update
				updateQuery := `
					UPDATE server_storage SET
						filesystem_type = ?,
						total_space = ?,
						used_space = ?,
						free_space = ?,
						use_percent = ?,
						updated_at = ?,
						updated_by = 'system',
						last_check = ?,
						last_update = ?
					WHERE server_id = ? AND device_name = ? AND mount_point = ?
				`
				_, err = tx.Exec(
					updateQuery,
					storage.FilesystemType,
					storage.TotalSpace,
					storage.UsedSpace,
					storage.FreeSpace,
					storage.UsePercent,
					currentTime,
					currentTime,
					currentTime, // Update last_update since data changed
					serverID,
					storage.DeviceName,
					storage.MountPoint,
				)
			} else {
				// No changes, only update last_check field
				updateQuery := `
					UPDATE server_storage SET
						last_check = ?
					WHERE server_id = ? AND device_name = ? AND mount_point = ?
				`
				_, err = tx.Exec(
					updateQuery,
					currentTime,
					serverID,
					storage.DeviceName,
					storage.MountPoint,
				)
			}
		} else {
			// New device, insert all data
			insertQuery := `
				INSERT INTO server_storage (
					server_id, device_name, mount_point, filesystem_type,
					total_space, used_space, free_space, use_percent,
					created_at, created_by, updated_at, updated_by,
					last_check, last_update
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, 'system', ?, 'system', ?, ?)
			`
			_, err = tx.Exec(
				insertQuery,
				serverID,
				storage.DeviceName,
				storage.MountPoint,
				storage.FilesystemType,
				storage.TotalSpace,
				storage.UsedSpace,
				storage.FreeSpace,
				storage.UsePercent,
				currentTime,
				currentTime,
				currentTime, // Initial last_check
				currentTime, // Initial last_update
			)
		}

		if err != nil {
			return fmt.Errorf("failed to upsert storage record for %s: %w", storage.DeviceName, err)
		}

		// Mark this device as updated
		updatedDevices[deviceKey] = true
	}

	// Delete any devices that existed before but weren't in this update
	// This handles the case where devices have been removed
	for existingDevice := range existingDevicesData {
		if !updatedDevices[existingDevice] {
			// Split the combined key back into device name and mount point
			parts := strings.Split(existingDevice, "|")
			if len(parts) != 2 {
				continue
			}
			deviceName := parts[0]
			mountPoint := parts[1]

			deleteQuery := "DELETE FROM server_storage WHERE server_id = ? AND device_name = ? AND mount_point = ?"
			_, err = tx.Exec(deleteQuery, serverID, deviceName, mountPoint)
			if err != nil {
				return fmt.Errorf("failed to remove obsolete device %s: %w", deviceName, err)
			}
		}
	}

	// Commit transaction
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Add a small pause to ensure spinner is visible
	time.Sleep(101 * time.Millisecond)

	return nil
}
