package zoneService

import (
	"database/sql"
	"dbaTools/internal/logger"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// ZoneService provides operations for zone management
type ZoneService struct {
	DB     *sql.DB
	Logger *logger.Logger
}

// NewZoneService creates a new zone service
func NewZoneService(db *sql.DB, log *logger.Logger) *ZoneService {
	return &ZoneService{
		DB:     db,
		Logger: log,
	}
}

// InsertZone inserts a new zone without retry logic
func (s *ZoneService) InsertZone(zoneName, zoneDesc string) error {
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Zone Add ] "
	spin.Color("green")
	spin.Start()
	defer spin.Stop()

	spin.Suffix = " Inserting zone data into database... \n"
	time.Sleep(200 * time.Millisecond)

	// Execute database insert directly without retry
	query := `INSERT INTO zone (name, description, created_at, created_by, updated_by) VALUES (?, ?, NOW(), 'DBA','DBA')`
	_, err := s.DB.Exec(query, zoneName, zoneDesc)

	if err != nil {
		// For other database errors
		spin.Stop()
		s.Logger.Error("Failed to insert zone", err)
	} else {
		// Insert successful
		spin.Stop()
		s.Logger.Success(fmt.Sprintf("Zone '%s' successfully added", zoneName))
	}

	return err
}
