package database

import (
	"database/sql"
	"dbaTools/internal/logger"
	"fmt"
	"time"
)

// DatabaseOperation provides retryable database operations
type DatabaseOperation struct {
	DB         *sql.DB
	Logger     *logger.Logger
	MaxRetries int
	RetryDelay time.Duration
	SilentMode bool // Tambahkan flag ini
}

// NewDatabaseOperation creates a new database operation handler
func NewDatabaseOperation(db *sql.DB, logger *logger.Logger, maxRetries int, retryDelay time.Duration) *DatabaseOperation {
	return &DatabaseOperation{
		DB:         db,
		Logger:     logger,
		MaxRetries: maxRetries,
		RetryDelay: retryDelay,
	}
}

// ExecuteWithRetry executes a database operation with retry logic
func (o *DatabaseOperation) ExecuteWithRetry(operation func() error) error {
	var lastErr error

	// Log the start of the operation
	if !o.SilentMode {
		o.Logger.Info(fmt.Sprintf("Starting database operation with max %d retries", o.MaxRetries))
	}

	for attempt := 1; attempt <= o.MaxRetries; attempt++ {
		// Log each attempt
		o.Logger.Info(fmt.Sprintf("Executing database operation (attempt %d/%d)",
			attempt, o.MaxRetries))

		err := operation()
		if err == nil {
			// Log successful operation
			o.Logger.Success(fmt.Sprintf("Database operation successful on attempt %d", attempt))
			return nil
		}

		lastErr = err

		// Log warning for failed attempt with detailed error
		o.Logger.Warning(fmt.Sprintf("Database operation attempt %d/%d failed: %v",
			attempt, o.MaxRetries, err))

		// If this is not the last attempt, log and wait before retrying
		if attempt < o.MaxRetries {
			retrySeconds := o.RetryDelay.Seconds()
			o.Logger.Info(fmt.Sprintf("Waiting %.1f seconds before retry %d/%d",
				retrySeconds, attempt+1, o.MaxRetries))
			time.Sleep(o.RetryDelay)
		}
	}

	// Log failure after all retry attempts
	o.Logger.Error(fmt.Sprintf("Database operation failed after %d attempts", o.MaxRetries), lastErr)

	return lastErr
}
