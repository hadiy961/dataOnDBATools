package status

import (
	"database/sql"
	"dbaTools/internal/logger"
	StatusForm "dbaTools/internal/package/manage/status/form"
	StatusService "dbaTools/internal/package/manage/status/service"
	TUIutils "dbaTools/internal/ui/component/utils"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// UpdateStatus menjalankan proses update zona
func UpdateStatus(logger *logger.Logger, db *sql.DB, StatusId string) error {
	logger.Info(fmt.Sprintf("Starting Status Update for ID: %s", StatusId))
	TUIutils.ClearScreen()

	// Create service
	service := StatusService.NewStatusService(db, logger)

	// Get Status data
	StatusData, err := service.GetStatusById(StatusId)
	if err != nil {
		return fmt.Errorf("failed to get Status data: %w", err)
	}

	// Show form and get updated values
	values, err := StatusForm.ShowStatusUpdateForm(logger, StatusData)
	if err != nil {
		if err.Error() == "canceled" {
			logger.Info("Status update canceled by user")
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	// Update Status in database
	err = service.UpdateStatus(StatusId, values["Status_name"], values["Status_desc"])
	if err == nil {
		// Success case
		return nil
	}

	// Any other error, return to caller
	return err
}

// AddStatus handles the Status addition workflow
func DeleteStatus(log *logger.Logger, dbConn *sql.DB, id string) error {
	log.Info("Starting Status Delete Action")
	TUIutils.ClearScreen()

	service := StatusService.NewStatusService(dbConn, log)

	// Initialize spinner for initial UI
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Prefix = "[ Status Delete ] "
	s.Color("green")
	s.Start()

	s.Suffix = " Initialize and run delete Status..."
	time.Sleep(101 * time.Millisecond)
	s.Stop()

	// Try to insert the Status
	err := service.DeleteStatus(id)
	if err == nil {
		// Success case
		return nil
	}
	// Any other error, return to caller
	return err
}

// AddStatus handles the Status addition workflow
func AddStatus(log *logger.Logger, dbConn *sql.DB) error {
	log.Info("Starting Status Add Form")
	TUIutils.ClearScreen()

	service := StatusService.NewStatusService(dbConn, log)

	// Initialize spinner for initial UI
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Prefix = "[ Status Add ] "
	s.Color("green")
	s.Start()

	s.Suffix = " Initialize and run add Status form..."
	time.Sleep(101 * time.Millisecond)
	s.Stop()

	formValues, err := StatusForm.ShowStatusAddForm(log)
	if err != nil {
		return err
	}

	// Extract Status details from form values
	StatusName := formValues["Status_name"]
	StatusDesc := formValues["Status_desc"]

	// Try to insert the Status
	err = service.InsertStatus(StatusName, StatusDesc)
	if err == nil {
		// Success case
		return nil
	}
	// Any other error, return to caller
	return err
}
