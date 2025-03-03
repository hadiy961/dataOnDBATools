package zone

import (
	"database/sql"
	"dbaTools/internal/logger"
	zoneForm "dbaTools/internal/package/manage/zone/form"
	zoneService "dbaTools/internal/package/manage/zone/service"
	"dbaTools/internal/utils"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// UpdateZone menjalankan proses update zona
func UpdateZone(logger *logger.Logger, db *sql.DB, zoneId string) error {
	logger.Info(fmt.Sprintf("Starting Zone Update for ID: %s", zoneId))
	utils.ClearScreen()

	// Create service
	service := zoneService.NewZoneService(db, logger)

	// Get zone data
	zoneData, err := service.GetZoneById(zoneId)
	if err != nil {
		return fmt.Errorf("failed to get zone data: %w", err)
	}

	// Show form and get updated values
	values, err := zoneForm.ShowZoneUpdateForm(logger, zoneData)
	if err != nil {
		if err.Error() == "canceled" {
			logger.Info("Zone update canceled by user")
			return nil
		}
		return fmt.Errorf("form error: %w", err)
	}

	// Update zone in database
	err = service.UpdateZone(zoneId, values["zone_name"], values["zone_desc"])
	if err == nil {
		// Success case
		return nil
	}

	// Any other error, return to caller
	return err
}

// AddZone handles the zone addition workflow
func DeleteZone(log *logger.Logger, dbConn *sql.DB, id string) error {
	log.Info("Starting Zone Delete Action")
	utils.ClearScreen()

	service := zoneService.NewZoneService(dbConn, log)

	// Initialize spinner for initial UI
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Prefix = "[ Zone Delete ] "
	s.Color("green")
	s.Start()

	s.Suffix = " Initialize and run delete zone..."
	time.Sleep(200 * time.Millisecond)
	s.Stop()

	// Try to insert the zone
	err := service.DeleteZone(id)
	if err == nil {
		// Success case
		return nil
	}
	// Any other error, return to caller
	return err
}

// AddZone handles the zone addition workflow
func AddZone(log *logger.Logger, dbConn *sql.DB) error {
	log.Info("Starting Zone Add Form")
	utils.ClearScreen()

	service := zoneService.NewZoneService(dbConn, log)

	// Initialize spinner for initial UI
	s := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	s.Prefix = "[ Zone Add ] "
	s.Color("green")
	s.Start()

	s.Suffix = " Initialize and run add zone form..."
	time.Sleep(200 * time.Millisecond)
	s.Stop()

	formValues, err := zoneForm.ShowZoneAddForm(log)
	if err != nil {
		return err
	}

	// Extract zone details from form values
	zoneName := formValues["zone_name"]
	zoneDesc := formValues["zone_desc"]

	// Try to insert the zone
	err = service.InsertZone(zoneName, zoneDesc)
	if err == nil {
		// Success case
		return nil
	}
	// Any other error, return to caller
	return err
}
