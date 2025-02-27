package zoneService

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// GetZoneByIdDetail retrieves zone data from the database
func (s *ZoneService) GetZoneByIdDetail(zoneId string) (ZoneData, error) {
	var zone ZoneData
	zone.Id = zoneId

	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Zone Detail ] "
	spin.Color("green")
	spin.Suffix = " Retrieving zone data... \n"
	spin.Start()
	time.Sleep(500 * time.Millisecond)

	defer spin.Stop()

	query := "SELECT name, description, created_at, updated_at, updated_by, created_by FROM zone WHERE id = ?"
	var createdAtBytes, updatedAtBytes []uint8
	err := s.DB.QueryRow(query, zoneId).Scan(&zone.Name, &zone.Description, &createdAtBytes, &updatedAtBytes, &zone.UpdatedBy, &zone.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			spin.Stop()
			s.Logger.Warning(fmt.Sprintf("Zone with ID %s not found", zoneId))
			return zone, fmt.Errorf("zone with ID %s not found", zoneId)
		} else {
			spin.Stop()
			s.Logger.Error(fmt.Sprintf("Failed to retrieve zone with ID %s", zoneId), err)
			return zone, fmt.Errorf("failed to retrieve zone with ID %s: %w", zoneId, err)
		}
	}

	// Convert byte slices to time.Time
	zone.CreatedAt, err = time.Parse("2006-01-02 15:04:05", string(createdAtBytes))
	if err != nil {
		spin.Stop()
		s.Logger.Error(fmt.Sprintf("Failed to parse created_at for zone with ID %s", zoneId), err)
		return zone, fmt.Errorf("failed to parse created_at for zone with ID %s: %w", zoneId, err)
	}

	zone.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", string(updatedAtBytes))
	if err != nil {
		spin.Stop()
		s.Logger.Error(fmt.Sprintf("Failed to parse updated_at for zone with ID %s", zoneId), err)
		return zone, fmt.Errorf("failed to parse updated_at for zone with ID %s: %w", zoneId, err)
	}

	spin.Stop()
	s.Logger.Info(fmt.Sprintf("Successfully retrieved zone with ID %s", zoneId))
	return zone, nil
}
