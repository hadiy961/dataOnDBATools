package StatusService

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// GetStatusByIdDetail retrieves Status data from the database
func (s *StatusService) GetStatusByIdDetail(StatusId string) (StatusData, error) {
	var Status StatusData
	Status.Id = StatusId

	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Status Detail ] "
	spin.Color("green")
	spin.Suffix = " Retrieving Status data... \n"
	spin.Start()
	time.Sleep(500 * time.Millisecond)

	defer spin.Stop()

	query := "SELECT name, description, created_at, updated_at, updated_by, created_by FROM Status WHERE id = ?"
	var createdAtBytes, updatedAtBytes []uint8
	err := s.DB.QueryRow(query, StatusId).Scan(&Status.Name, &Status.Description, &createdAtBytes, &updatedAtBytes, &Status.UpdatedBy, &Status.CreatedBy)
	if err != nil {
		if err == sql.ErrNoRows {
			spin.Stop()
			s.Logger.Warning(fmt.Sprintf("Status with ID %s not found", StatusId))
			return Status, fmt.Errorf("Status with ID %s not found", StatusId)
		} else {
			spin.Stop()
			s.Logger.Error(fmt.Sprintf("Failed to retrieve Status with ID %s", StatusId), err)
			return Status, fmt.Errorf("failed to retrieve Status with ID %s: %w", StatusId, err)
		}
	}

	// Convert byte slices to time.Time
	Status.CreatedAt, err = time.Parse("2006-01-02 15:04:05", string(createdAtBytes))
	if err != nil {
		spin.Stop()
		s.Logger.Error(fmt.Sprintf("Failed to parse created_at for Status with ID %s", StatusId), err)
		return Status, fmt.Errorf("failed to parse created_at for Status with ID %s: %w", StatusId, err)
	}

	Status.UpdatedAt, err = time.Parse("2006-01-02 15:04:05", string(updatedAtBytes))
	if err != nil {
		spin.Stop()
		s.Logger.Error(fmt.Sprintf("Failed to parse updated_at for Status with ID %s", StatusId), err)
		return Status, fmt.Errorf("failed to parse updated_at for Status with ID %s: %w", StatusId, err)
	}

	spin.Stop()
	s.Logger.Info(fmt.Sprintf("Successfully retrieved Status with ID %s", StatusId))
	return Status, nil
}
