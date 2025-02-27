package StatusService

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// StatusData menyimpan data zona
type StatusData struct {
	Id          string
	Name        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	UpdatedBy   string
	CreatedBy   string
}

func (s *StatusService) GetStatusById(StatusId string) (StatusData, error) {
	var Status StatusData
	Status.Id = StatusId

	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Status Update ] "
	spin.Color("green")
	spin.Suffix = " Retrieving Status data... \n"
	spin.Start()
	time.Sleep(101 * time.Millisecond)

	defer spin.Stop()

	// Use a temporary variable for timestamp scanning
	var createdAt, updatedAt sql.NullString

	query := "SELECT name, description, created_at, created_by, updated_at, updated_by FROM Status WHERE id = ?"
	err := s.DB.QueryRow(query, StatusId).Scan(
		&Status.Name,
		&Status.Description,
		&createdAt,
		&Status.CreatedBy,
		&updatedAt,
		&Status.UpdatedBy,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			spin.Stop()
			s.Logger.Warning(fmt.Sprintf("Status with ID %s not found", StatusId))
			return Status, fmt.Errorf("Status with ID %s not found", StatusId)
		}
		spin.Stop()
		s.Logger.Error(fmt.Sprintf("Failed to retrieve Status with ID %s", StatusId), err)
		return Status, fmt.Errorf("failed to retrieve Status with ID %s: %w", StatusId, err)
	}

	// Parse the timestamps if they are valid
	if createdAt.Valid {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", createdAt.String)
		if err == nil {
			Status.CreatedAt = parsedTime
		}
	}

	if updatedAt.Valid {
		parsedTime, err := time.Parse("2006-01-02 15:04:05", updatedAt.String)
		if err == nil {
			Status.UpdatedAt = parsedTime
		}
	}

	spin.Stop()
	s.Logger.Info(fmt.Sprintf("Successfully retrieved Status with ID %s", StatusId))
	return Status, nil
}

// UpdateStatus memperbarui data Status di database
func (s *StatusService) UpdateStatus(StatusId string, name, desc string) error {
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Status Update ] "
	spin.Color("green")
	spin.Suffix = " Updating Status data...\n"
	spin.Start()
	defer spin.Stop()
	time.Sleep(101 * time.Millisecond)

	// Execute database update directly
	query := "UPDATE Status SET name = ?, description = ?, updated_by = 'DBA' WHERE id = ?"
	_, err := s.DB.Exec(query, name, desc, StatusId)

	if err != nil {
		spin.Stop()
		// For other database errors
		s.Logger.Error(fmt.Sprintf("Failed to update Status #%s", StatusId), err)
	}
	spin.Stop()

	// Update successful
	s.Logger.Success(fmt.Sprintf("Status #%s successfully updated", StatusId))
	return nil
}
