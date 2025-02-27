package StatusService

import (
	"database/sql"
	"dbaTools/internal/logger"
	"fmt"
	"time"

	"github.com/briandowns/spinner"
)

// StatusService provides operations for Status management
type StatusService struct {
	DB     *sql.DB
	Logger *logger.Logger
}

// NewStatusService creates a new Status service
func NewStatusService(db *sql.DB, log *logger.Logger) *StatusService {
	return &StatusService{
		DB:     db,
		Logger: log,
	}
}

// InsertStatus inserts a new Status without retry logic
func (s *StatusService) InsertStatus(StatusName, StatusDesc string) error {
	// Initialize spinner
	spin := spinner.New(spinner.CharSets[11], 100*time.Millisecond)
	spin.Prefix = "[ Status Add ] "
	spin.Color("green")
	spin.Start()
	defer spin.Stop()

	spin.Suffix = " Inserting Status data into database... \n"
	time.Sleep(500 * time.Millisecond)

	// Execute database insert directly without retry
	query := `INSERT INTO Status (name, description, created_at, created_by, updated_by) VALUES (?, ?, NOW(), 'DBA','DBA')`
	_, err := s.DB.Exec(query, StatusName, StatusDesc)

	if err != nil {
		// For other database errors
		spin.Stop()
		s.Logger.Error("Failed to insert Status", err)
	} else {
		// Insert successful
		spin.Stop()
		s.Logger.Success(fmt.Sprintf("Status '%s' successfully added", StatusName))
	}

	return err
}
