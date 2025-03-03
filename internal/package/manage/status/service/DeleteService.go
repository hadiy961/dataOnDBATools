package StatusService

import (
	"dbaTools/internal/utils"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// DeleteStatus deletes a Status without retry logic
func (s *StatusService) DeleteStatus(StatusID string) error {
	utils.ClearScreen()
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Status Delete ] "
	spin.Color("red")
	spin.Start()
	defer spin.Stop()

	spin.Suffix = " Deleting Status data from database... \n"
	time.Sleep(200 * time.Millisecond)

	// Execute database delete directly without retry
	query := `DELETE FROM Status WHERE id = ?`
	result, err := s.DB.Exec(query, StatusID)

	if err != nil {
		// For other database errors
		spin.Stop()
		s.Logger.Error("Failed to delete Status", err)
	} else {
		// Delete successful
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			spin.Stop()
			s.Logger.Success(fmt.Sprintf("Status with ID %s successfully deleted", StatusID))
		} else {
			spin.Stop()
			s.Logger.Warning(fmt.Sprintf("Status with ID %s not found", StatusID))
		}
	}

	return err
}
