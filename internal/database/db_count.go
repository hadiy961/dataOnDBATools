package database

import (
	"context"
	"fmt"
	"time"
)

type StatementTimeManager struct {
	conn          *DBConnection
	originalValue int
}

func NewStatementTimeManager(conn *DBConnection) *StatementTimeManager {
	return &StatementTimeManager{conn: conn}
}

func (stm *StatementTimeManager) Begin() error {
	// Get original value
	err := stm.conn.db.QueryRow("SELECT @@max_statement_time").Scan(&stm.originalValue)
	if err != nil {
		return fmt.Errorf("error getting original max_statement_time: %w", err)
	}

	// Set to unlimited
	_, err = stm.conn.db.Exec("SET MAX_STATEMENT_TIME = 0")
	if err != nil {
		return fmt.Errorf("error setting max_statement_time to 0: %w", err)
	}

	return nil
}

func (stm *StatementTimeManager) Restore() error {
	_, err := stm.conn.db.Exec(fmt.Sprintf("SET MAX_STATEMENT_TIME = %d", stm.originalValue))
	if err != nil {
		return fmt.Errorf("error restoring max_statement_time: %w", err)
	}
	return nil
}

// GetDatabaseCharset returns charset and collation
func (d *DBConnection) GetDatabaseCharset() (string, string, error) {
	var charset, collation string
	query := `SELECT default_character_set_name, default_collation_name 
		FROM information_schema.schemata WHERE schema_name = ?`
	err := d.db.QueryRow(query, d.config.Database).Scan(&charset, &collation)
	if err != nil {
		return "", "", fmt.Errorf("error getting charset info: %w", err)
	}
	return charset, collation, nil
}

func (d *DBConnection) GetDatabaseSize() (int64, error) {
	stm := NewStatementTimeManager(d)
	if err := stm.Begin(); err != nil {
		return 0, err
	}
	defer stm.Restore()

	ctx, cancel := context.WithTimeout(context.Background(), 1300*time.Second)
	defer cancel()

	var size int64
	query := `SELECT COALESCE(SUM(data_length + index_length), 0)
			  FROM information_schema.tables 
			  WHERE table_schema = ?
			  AND table_type = 'BASE TABLE'`

	err := d.db.QueryRowContext(ctx, query, d.config.Database).Scan(&size)
	if err != nil {
		return 0, fmt.Errorf("error getting size: %w", err)
	}

	return size, nil
}

func (d *DBConnection) GetTotalDatabaseSize() (int64, error) {
	query := `
		SELECT SUM(data_length + index_length) as total_size
		FROM information_schema.TABLES`

	var totalSize int64
	err := d.db.QueryRow(query).Scan(&totalSize)
	if err != nil {
		return 0, fmt.Errorf("error getting total database size: %w", err)
	}

	return totalSize, nil
}

// GetDatabaseCreationTime returns database creation timestamp
func (d *DBConnection) GetDatabaseCreationTime() (time.Time, error) {
	var creationTime time.Time
	query := `SELECT created FROM information_schema.schemata 
		WHERE schema_name = ?`
	err := d.db.QueryRow(query, d.config.Database).Scan(&creationTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("error getting creation time: %w", err)
	}
	return creationTime, nil
}

// GetTriggerCount returns number of triggers
func (d *DBConnection) GetTriggerCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.triggers 
		WHERE trigger_schema = ?`
	err := d.db.QueryRow(query, d.config.Database).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting trigger count: %w", err)
	}
	return count, nil
}

// GetEventCount returns number of events
func (d *DBConnection) GetEventCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.events 
		WHERE event_schema = ?`
	err := d.db.QueryRow(query, d.config.Database).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting event count: %w", err)
	}
	return count, nil
}

// GetFunctionCount returns number of functions
func (d *DBConnection) GetFunctionCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.routines 
		WHERE routine_schema = ? AND routine_type = 'FUNCTION'`
	err := d.db.QueryRow(query, d.config.Database).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting function count: %w", err)
	}
	return count, nil
}

// GetStoredProcedureCount returns number of stored procedures
func (d *DBConnection) GetStoredProcedureCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.routines 
		WHERE routine_schema = ? AND routine_type = 'PROCEDURE'`
	err := d.db.QueryRow(query, d.config.Database).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting stored procedure count: %w", err)
	}
	return count, nil
}

// GetTableCount returns number of tables in database
func (d *DBConnection) GetTableCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = ? AND table_type = 'BASE TABLE'`
	err := d.db.QueryRow(query, d.config.Database).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting table count: %w", err)
	}
	return count, nil
}

// GetViewCount returns number of views in database
func (d *DBConnection) GetViewCount() (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM information_schema.tables 
		WHERE table_schema = ? AND table_type = 'VIEW'`
	err := d.db.QueryRow(query, d.config.Database).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("error getting view count: %w", err)
	}
	return count, nil
}
