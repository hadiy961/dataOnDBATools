package zoneService

import (
	TUIutils "dbaTools/internal/ui/component/utils"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// DeleteZone deletes a zone without retry logic
func (s *ZoneService) DeleteZone(zoneID string) error {
	TUIutils.ClearScreen()
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Zone Delete ] "
	spin.Color("red")
	spin.Start()
	defer spin.Stop()

	spin.Suffix = " Deleting zone data from database... \n"
	time.Sleep(500 * time.Millisecond)

	// Execute database delete directly without retry
	query := `DELETE FROM zone WHERE id = ?`
	result, err := s.DB.Exec(query, zoneID)

	if err != nil {
		// For other database errors
		spin.Stop()
		s.Logger.Error("Failed to delete zone", err)
	} else {
		// Delete successful
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			spin.Stop()
			s.Logger.Success(fmt.Sprintf("Zone with ID %s successfully deleted", zoneID))
		} else {
			spin.Stop()
			s.Logger.Warning(fmt.Sprintf("Zone with ID %s not found", zoneID))
		}
	}

	return err
}
