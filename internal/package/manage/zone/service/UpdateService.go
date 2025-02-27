package zoneService

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// ZoneData menyimpan data zona
type ZoneData struct {
	Id          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	UpdatedBy   string
	CreatedBy   string
}

func (s *ZoneService) GetZoneById(zoneId string) (ZoneData, error) {
	var zone ZoneData
	zone.Id = zoneId

	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Zone Update ] "
	spin.Color("green")
	spin.Suffix = " Retrieving zone data... \n"
	spin.Start()
	time.Sleep(101 * time.Millisecond)

	defer spin.Stop()

	// Use a temporary variable for timestamp scanning
	var createdAt, updatedAt sql.NullString

	query := "SELECT name, description, created_at, created_by, updated_at, updated_by FROM zone WHERE id = ?"
	err := s.DB.QueryRow(query, zoneId).Scan(
		&zone.Name,
		&zone.Description,
		&createdAt,
		&zone.CreatedBy,
		&updatedAt,
		&zone.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			spin.Stop()
			s.Logger.Warning(fmt.Sprintf("Zone with ID %s not found", zoneId))
			return zone, fmt.Errorf("zone with ID %s not found", zoneId)
		}
		spin.Stop()
		s.Logger.Error(fmt.Sprintf("Failed to retrieve zone with ID %s", zoneId), err)
		return zone, fmt.Errorf("failed to retrieve zone with ID %s: %w", zoneId, err)
	}

	// Parse the timestamps if they are valid
	if createdAt.Valid {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt.String)
		if err == nil {
			zone.CreatedAt = parsedTime
		}
	}

	if updatedAt.Valid {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", updatedAt.String)
		if err == nil {
			zone.UpdatedAt = parsedTime
		}
	}

	spin.Stop()
	s.Logger.Info(fmt.Sprintf("Successfully retrieved zone with ID %s", zoneId))
	return zone, nil
}

// UpdateZone memperbarui data zone di database
func (s *ZoneService) UpdateZone(zoneId string, name, desc string) error {
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Zone Update ] "
	spin.Color("green")
	spin.Suffix = " Updating zone data...\n"
	spin.Start()
	defer spin.Stop()
	time.Sleep(101 * time.Millisecond)

	// Execute database update directly
	query := "UPDATE zone SET name = ?, description = ?, updated_by = 'DBA' WHERE id = ?"
	_, err := s.DB.Exec(query, name, desc, zoneId)

	if err != nil {
		spin.Stop()
		// For other database errors
		s.Logger.Error(fmt.Sprintf("Failed to update zone #%s", zoneId), err)
	}
	spin.Stop()

	// Update successful
	s.Logger.Success(fmt.Sprintf("Zone #%s successfully updated", zoneId))
	return nil
}
